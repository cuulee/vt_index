package vt_index 

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/murphy214/mercantile"
	"log"
	"fmt"
)

type Mb_Index struct {
	DB *sql.DB
	Zoom int
	Cache map[m.TileID]Polygon_Index
	Cache_Size int
	Tx * sql.Tx
	Tile_Map map[m.TileID]string
}

type Ind_Output struct {
	TileID m.TileID
	Index Polygon_Index
}

// creating mb_index 
func Create_Mb_Index(filename string,cache_size int) Mb_Index {
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
	query = "select tile_column,tile_row,zoom_level from tiles;"
	rows, err := tx.Query(query)
	var X,Y,Z int
	tilemap := map[m.TileID]string{}
	for rows.Next() {
		// creating properties map
		rows.Scan(&X, &Y, &Z)

		tileid := m.TileID{X:int64(X),Y:int64(Y),Z:uint64(Z)}
		tilemap[tileid] = ""

	}
	query = fmt.Sprintf("select tile_column,tile_row,zoom_level,tile_data from tiles LIMIT %d;",cache_size)
	rows, err = tx.Query(query)
	cache := map[m.TileID]Polygon_Index{}
	var data []byte
	fmt.Print(err)
	count := 0
	c := make(chan Ind_Output,cache_size)
	for rows.Next() {
		// creating properties map
		rows.Scan(&X, &Y, &Z, &data)

		tileid := m.TileID{X:int64(X),Y:int64(Y),Z:uint64(Z)}
		bds := m.Bounds(tileid)
		
		// checking to see whether to add the index
		if count < cache_size {
			go func(data []byte,tileid m.TileID,c chan Ind_Output) {
				index := Polygon_Index{Index:Read_Vector_Tile_Index_Byte(data),Latconst:bds.N}
				c <- Ind_Output{TileID:tileid,Index:index}
				//cache[tileid] = index
			}(data,tileid,c)
		}
		count += 1

	}

	counter := count
	count = 0
	for count < counter {
		output := <-c
		cache[output.TileID] = output.Index
		count += 1
		fmt.Printf("\r[%d/%d] Creating Cached Indexes",count,counter)
	}
	fmt.Print("\n")

	return Mb_Index{Cache:cache,DB:db,Zoom:zoom,Cache_Size:cache_size,Tx:tx,Tile_Map:tilemap}

}

//
func (mb_index Mb_Index) Pip(point []float64) string {
	tileid := m.Tile(point[0],point[1],mb_index.Zoom)
	p_index,ok := mb_index.Cache[tileid]
	if ok == true {
		return p_index.Pip_Simple(point)
	} else {
		// checking to see if tileid is in index
		_,ok2 := mb_index.Tile_Map[tileid]
		if ok2 == true {
			query := fmt.Sprintf("select tile_data from tiles where zoom_level = %d and  tile_column = %d and tile_row = %d;",tileid.Z,tileid.X,tileid.Y)
			//tx, err := mb_index.DB.Begin()
			var data []byte
			mb_index.Tx.QueryRow(query).Scan(&data)
			if len(data) > 0 {
				bds := m.Bounds(tileid)
				index := Polygon_Index{Index:Read_Vector_Tile_Index_Byte(data),Latconst:bds.N}
				mb_index.Cache[tileid] = index
				return index.Pip_Simple(point)
			} else {
				return ""
			}
		} else {
			return ""
		}

	}
}

