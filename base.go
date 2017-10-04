package vt_index


import (
	"github.com/paulmach/go.geojson"
	pc "github.com/murphy214/polyclip"
	m "github.com/murphy214/mercantile"
	"github.com/murphy214/gotile/gotile"

	//"time"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"os"
	_ "net/http/pprof"
)


// structure for creaitng output
type Output_Feature struct {
	Polygon pc.Polygon
	Feature geojson.Feature
	BB m.Extrema
	Area string
}

// structure for a map output
type Map_Output struct {
	Key Output_Feature
	Feats []Output_Feature
}

// structure for a map output
type Index_Output struct {
	TileID m.TileID
	Bytes []byte
}


// given a layer and fields to carry over parses json for that area fields and adds to output feature
func Filter_Fields(layer []*geojson.Feature,fields []string) []*geojson.Feature {
	for ii,i := range layer {
		props := i.Properties
		newmap := map[string]interface{}{}
		for _,field := range fields {
			newmap[field] = props[field]
		}
		b, _ := json.Marshal(newmap)
		i.Properties = map[string]interface{} {"AREA":string(b)}
		layer[ii] = i
	}
	return layer
}



// creates a pc polygon
func Make_Polygon(coords [][][]float64) pc.Polygon {
	thing2 := pc.Contour{}
	things := pc.Polygon{}
	for _, coord := range coords {
		thing2 = pc.Contour{}

		for _, i := range coord {
			if len(i) >= 2 {
				// moving sign in 10 ** -7 pla

				thing2.Add(pc.Point{X: i[0], Y: i[1]})
			}
		}
		things.Add(thing2)
	}

	return things
}

// makes children and returns tilemap of a first intialized tilemap
func Make_Tilemap_Children(tilemap map[m.TileID][]*geojson.Feature) (map[m.TileID][]*geojson.Feature) {

	// iterating through each tileid
	ccc := make(chan map[m.TileID][]*geojson.Feature)
	newmap := map[m.TileID][]*geojson.Feature{}
	count2 := 0
	counter := 0
	sizetilemap := len(tilemap)
	buffer := 100000

	// iterating through each tielmap
	for k, v := range tilemap {
		go func(k m.TileID, v []*geojson.Feature, ccc chan map[m.TileID][]*geojson.Feature) {
			cc := make(chan map[m.TileID][]*geojson.Feature)
			for _, i := range v {
				go func(k m.TileID, i *geojson.Feature, cc chan map[m.TileID][]*geojson.Feature) {
					if i.Geometry.Type == "Polygon" {
						cc <- tile_surge.Children_Polygon(i, k)
					} else if i.Geometry.Type == "LineString" {
						//partmap := Env_Line(i, int(k.Z+1))
						//partmap = Lint_Children_Lines(partmap, k)
						//cc <- partmap
					} else if i.Geometry.Type == "Point" {
						//partmap := map[m.TileID][]*geojson.Feature{}
						//pt := i.Geometry.Point
						//tileid := m.Tile(pt[0], pt[1], int(k.Z+1))
						//partmap[tileid] = append(partmap[tileid], i)
						//cc <- partmap
					}
				}(k, i, cc)
			}

			// collecting all into child map
			childmap := map[m.TileID][]*geojson.Feature{}
			for range v {
				tempmap := <-cc
				for k, v := range tempmap {
					childmap[k] = append(childmap[k], v...)
				}
			}

			ccc <- childmap
		}(k, v, ccc)

		counter += 1
		// collecting shit
		if (counter == buffer) || (sizetilemap-1 == count2) {
			count := 0

			for count < counter {
				tempmap := <-ccc
				for k, v := range tempmap {
					newmap[k] = append(newmap[k], v...)
				}
				count += 1
			}
			counter = 0
			fmt.Printf("\r[%d / %d] Tiles Complete, Size: %d       ", count2, sizetilemap, int(k.Z)+1)

		}
		count2 += 1

	}


	// getting size of total number of features within the tilemap
	totalsize := 0
	for _,v := range newmap {
		totalsize += len(v)
	}



	return newmap
}



// structure for finding overlapping values
func Overlapping_1D(box1min float64,box1max float64,box2min float64,box2max float64) bool {
	if box1max >= box2min && box2max >= box1min {
		return true
	} else {
		return false
	}
	return false
}


// returns a boolval for whether or not the bb intersects
func (feat Output_Feature) Intersect(bds m.Extrema) bool {
	bdsref := feat.BB
	if Overlapping_1D(bdsref.W-.0000001,bdsref.E+.0000001,bds.W-.0000001,bds.E+.0000001) && Overlapping_1D(bdsref.S-.0000001,bdsref.N+.0000001,bds.S-.0000001,bds.N+.0000001) {
		return true
	} else {
		return false
	}

	return false
}

// creates bounding box featues
func Create_BB_Feats(feat Output_Feature,layer2 []Output_Feature) ([]Output_Feature) {
	bb := feat.BB
	feats := []Output_Feature{}
	for _,i := range layer2 {
		boolval := i.Intersect(bb)
		if boolval == true {
			feats = append(feats,i)
		}
	}

	return feats
}

// makes a tilemap and returns
func Make_Tilemap(feats *geojson.FeatureCollection, size int) map[m.TileID][]*geojson.Feature {
	c := make(chan map[m.TileID][]*geojson.Feature)
	for _, i := range feats.Features {
		partmap := map[m.TileID][]*geojson.Feature{}

		go func(i *geojson.Feature, size int, c chan map[m.TileID][]*geojson.Feature) {
			//partmap := map[m.TileID][]*geojson.Feature{}

			if i.Geometry.Type == "Polygon" {
				partmap = Env_Polygon(i, size)
			} else if i.Geometry.Type == "LineString" {
				//partmap = Env_Line(i, size)
			} else if i.Geometry.Type == "Point" {
				//pt := i.Geometry.Point
				///tileid := m.Tile(pt[0], pt[1], size)
				//partmap[tileid] = append(partmap[tileid], i)
			}
			c <- partmap
		}(i, size, c)
	}

	// collecting channel shit
	totalmap := map[m.TileID][]*geojson.Feature{}
	for range feats.Features {
		partmap := <-c
		for k, v := range partmap {
			totalmap[k] = append(totalmap[k], v...)
		}
	}

	// getting size of total number of features within the tilemap
	totalsize := 0
	for _,v := range totalmap {
		totalsize += len(v)
	}


	//filemap := TileMapIO(totalmap)
	return totalmap
}



func Get_Filename(tileid m.TileID) string {
	tilestr := m.Tilestr(tileid)
	tilestr = strings.Replace(tilestr,"/","_",1000)
	tilestr = "temp/" + tilestr
	return tilestr
}

// reads a geobuf
func Read_Geobuf_Filename(filename string) []*geojson.Feature {
	in, _ := ioutil.ReadFile(filename)
	val,_ := geojson.UnmarshalFeatureCollection(in)
	return val.Features
}


// makes a tilemap and returns
func TileMapIO(tilemap map[m.TileID][]*geojson.Feature) map[m.TileID]string {
	os.MkdirAll("temp", os.ModePerm)
	mapfiles := map[m.TileID]string{}
	// iterating throuh eadch value
	for k,v := range tilemap {
		filename := Get_Filename(k)
		mapfiles[k] = filename
		fc := &geojson.FeatureCollection{Features:v}
		bytes,_ := fc.MarshalJSON()
		ioutil.WriteFile(filename, []byte(bytes), 0644)
		fc.Features = []*geojson.Feature{}

	}
	return mapfiles
}

func Small_Tile_Index(tilemap map[m.TileID][]*geojson.Feature) map[m.TileID][]*geojson.Feature {
	count := 0
	newtilemap := map[m.TileID][]*geojson.Feature{}
	for k,v := range tilemap {
		if count < 150 {
			newtilemap[k] = v
		}
		count += 1
	}
	return newtilemap
}




// creats a tile index
func Make_Tile_Index(tilemap map[m.TileID][]*geojson.Feature,zoom int,filename string) {
	//filemap := Make_Tilemap_Index(layer,zoom)
	//TileMapIO(tilemap)

	// creating sqlite db
	db := Create_Database_Meta(filename,zoom)

	// iterating through each tilemap 
	var sema = make(chan struct{},100)
	c := make(chan Index_Output)

	for k,v := range tilemap {
		go func(k m.TileID,v []*geojson.Feature, c chan Index_Output) {

			sema <- struct{}{}        // acquire token
			defer func() { <-sema }() // release token
			// iterating through each polygon in the tileid
			cc := make(chan map[string][]Yrow)
			for _,i := range v {
				go func(i *geojson.Feature,c chan map[string][]Yrow) {
					area := i.Properties["AREA"]
					areastr := area.(string)
					cc <- Make_Xmap_Total(i.Geometry.Polygon,areastr,k)
				}(i,cc)
			}

			// collecting yrow channel
			totalmap := map[string][]Yrow{}
			for range v {
				tempmap := <-cc
				for k,v := range tempmap {
					totalmap[k] = append(totalmap[k],v...)
				}			
			}
			c <- Index_Output{Bytes:Make_Vector_Tile_Index_Byte(totalmap),TileID:k}
		}(k,v,c)
	}

	count := 0
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("insert into tiles(zoom_level, tile_column,tile_row,tile_data) values(?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	for range tilemap {

		output := <-c
		k := output.TileID
		_, err = stmt.Exec(int(k.Z),int(k.X),int(k.Y),output.Bytes)
		if err != nil {
			fmt.Print(err)
		}
		//bytemap[output.TileID] = output.Bytes
		fmt.Printf("[%d/%d]\n",count,len(tilemap))
		count += 1

	}
	tx.Commit()
	os.RemoveAll("temp")
	Make_Index(db)
}

