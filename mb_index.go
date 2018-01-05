package exp 

import (
	"fmt"
	"github.com/paulmach/go.geojson"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	util "github.com/murphy214/mbtiles-util"

	"os"
	"sync"
)


type Output_Index struct {
	TileID m.TileID
	Bytevals []byte
}

type Ind_Output struct {
	TileID m.TileID
	Index Tile_Index
}

type Mb_Index struct {
	Zoom int
	Cache map[m.TileID]Tile_Index
}

// outer api for makign sqllite3 idnex object
func Make_Mb_Index(feats []*geojson.Feature,zoom int,filename string) {
	// creating the sqllite database
	//db := Create_Database_Meta(filename,zoom)
	config := util.Config{
		FileName:filename,
		LayerProperties:map[string]interface{}{},
	}
	os.Remove(filename)
	mbtile := util.Create_DB(config)
	// getting tilemap 
	tilemap,_ := Make_Tilemap(&geojson.FeatureCollection{Features:feats},zoom)

	fmt.Println(len(tilemap))
	//count := 0
	var counter = 0

    maxGoroutines := 20
    guard := make(chan struct{}, maxGoroutines)
	
	// iterating throguh each tile
	var wg sync.WaitGroup
	for k,v := range tilemap {
		wg.Add(1)
        guard <- struct{}{} // would block if guard channel is already filled

		go func(k m.TileID,v []*geojson.Feature) {
			temp_tile_index := Make_Xmap_Polygons(v,k)
			//fmt.Println(temp_tile_index)
			Write_Tile_Index(temp_tile_index,k,mbtile)
			<-guard
			counter += 1

			wg.Done()
			fmt.Printf("\r[%d/%d]",counter,len(tilemap))
		}(k,v)
	}
	wg.Wait()

	// inserting data
	//Insert_Data(totalmap,db)
	mbtile.Commit()
	// making index
	//Make_Index(db)

}

// reads the mb index into memory
func Read_Mb_Index(filename string) Mb_Index {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var zoom int
	query := "select value from metadata where name = 'zoom';"
	err = tx.QueryRow(query).Scan(&zoom)
	query = "select tile_column,tile_row,zoom_level,tile_data from tiles;"
	rows, err := tx.Query(query)
	cache := map[m.TileID]Tile_Index{}

	var data []byte
	var X,Y,Z int

	fmt.Print(err)
	counter := 0
	c := make(chan Ind_Output)
	var sema = make(chan struct{}, 10)

	for rows.Next() {
		counter += 1
		// creating properties map
		rows.Scan(&X, &Y, &Z, &data)
		Y = (1 << uint64(Z)) - Y - 1
		//fmt.Println(Y)
		tileid := m.TileID{X:int64(X),Y:int64(Y),Z:uint64(Z)}
		zoom = Z
		//bds := m.Bounds(tileid)
		
		// checking to see whether to add the index
		go func(data []byte,tileid m.TileID,c chan Ind_Output) {
			sema <- struct{}{}        // acquire token
			defer func() { <-sema }() // release token
			//tile_index := Read_Tile_Index(data,tileid)
			//ex := m.Bounds(tileid)
			//fmt.Println(tile_index.Index[geo.NewPoint(ex.W+.00000001, tile_index.Lat).GeoHash(9)])

			c <- Ind_Output{TileID:tileid,Index:Read_Tile_Index(data,tileid)}
			//cache[tileid] = index
		}(data,tileid,c)
	}

	count := 0
	for count < counter {
		output := <-c
		cache[output.TileID] = output.Index
		count += 1
		fmt.Printf("\r[%d/%d] Reading Cached Vector Tile Indexes",count,counter)
	}
	fmt.Println(zoom)
	return Mb_Index{Cache:cache,Zoom:zoom}
}

// outer level abstraction for point in polygon
func (mb_index Mb_Index) Pip(point []float64) map[string]interface{} {
 	return *mb_index.Cache[m.Tile(point[0],point[1],mb_index.Zoom)].Pip(point)
}

