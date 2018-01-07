package exp

import (
	m "github.com/murphy214/mercantile"
	"math/rand"
	"fmt"
	geo "github.com/paulmach/go.geo"
	//"io/ioutil"
	h "github.com/mitchellh/hashstructure"
	"github.com/paulmach/go.geojson"

)

func Geohash(lat,long float64,size int) string {
	return geo.NewPoint(long,lat).GeoHash(size)	
}


func Linspace_Delta(min float64,max float64,delta float64) []float64 {
	current := min
	newlist := []float64{}
	for current <= max {
		current += delta
		newlist = append(newlist,current)
	}
	fmt.Println(len(newlist))
	return newlist 
} 

// will always be x
func Zip_1DX(x float64,vals []float64) [][]float64 {
	newlist := make([][]float64,len(vals))

	for i,val := range vals {
		newlist[i] = []float64{x,val}
	}
	return newlist
}


func Create_Grid_Square(bds m.Extrema,size Size) [][]float64 {
	// creating xs and ys
	xs := Linspace_Delta(bds.W,bds.E,size.deltaX)
	ys := Linspace_Delta(bds.S,bds.N,size.deltaY)

	newlist := [][]float64{}
	for _,x := range xs {
		newlist = append(newlist,Zip_1DX(x,ys)...)
	}

	return newlist 
}




func (mb_index *Mb_Index) Get_One_Index() (m.TileID,Tile_Index) {

	value := rand.Intn(len(mb_index.Cache)-1) 
	count := 0
	var kk m.TileID
	var vv Tile_Index

	for k,v := range mb_index.Cache {
		if count == value {
			kk = k
			vv = v
		}
		count += 1
	}

	return kk,vv	
}


func (tile_index Tile_Index) Make_Grid_Pts(tileid m.TileID) [][]float64 {
	bds := m.Bounds(tileid)

	// setting up west-south and east-norht points
	ws := []float64{bds.W,bds.S}
	en := []float64{bds.E,bds.N}

	// geohahsing eachs
	ws_ghash := geo.NewPoint(ws[0], ws[1]).GeoHash(9)
	en_ghash := geo.NewPoint(en[0], en[1]).GeoHash(9)

	// getting middle of each point
	ws_mid := Get_Middle(ws_ghash)
	en_mid := Get_Middle(en_ghash)

	// getitng size of a geohash
	size := get_size_geohash(en_ghash)
	size.deltaX = size.deltaX * 32
	size.deltaY = size.deltaY * 32


	// setting up new bounds 
	ext := m.Extrema{W:ws_mid[0],S:ws_mid[1],E:en_mid[0],N:en_mid[1]}

	// creating create square
	fmt.Printf("%+v\n",ext)
	pts := Create_Grid_Square(ext,size)

	return pts
}


var colorkeys = []string{"#0030E5", "#0042E4", "#0053E4", "#0064E4", "#0075E4", "#0186E4", "#0198E3", "#01A8E3", "#01B9E3", "#01CAE3", "#02DBE3", "#02E2D9", "#02E2C8", "#02E2B7", "#02E2A6", "#03E295", "#03E184", "#03E174", "#03E163", "#03E152", "#04E142", "#04E031", "#04E021", "#04E010", "#09E004", "#19E005", "#2ADF05", "#3BDF05", "#4BDF05", "#5BDF05", "#6CDF06", "#7CDE06", "#8CDE06", "#9DDE06", "#ADDE06", "#BDDE07", "#CDDD07", "#DDDD07", "#DDCD07", "#DDBD07", "#DCAD08", "#DC9D08", "#DC8D08", "#DC7D08", "#DC6D08", "#DB5D09", "#DB4D09", "#DB3D09", "#DB2E09", "#DB1E09", "#DB0F0A"}


// testing a single tile index
// writes out the outputs of geojson grid
func (tile_index Tile_Index) Test(tileid m.TileID) []*geojson.Feature {
	pts := tile_index.Make_Grid_Pts(tileid)
	fmt.Println(len(pts))
	colormap := map[uint64]string{}
	feats := []*geojson.Feature{}
	for _,pt := range pts {
		areamap := tile_index.Pip(pt)
		hash, _ := h.Hash(areamap, nil)
		aream := *areamap
		area := aream["AREA"]

		color,boolval := colormap[hash]
		if boolval == false {
			colormap[hash] = colorkeys[rand.Intn(len(colorkeys)-1)]
			color = colormap[hash]
		}
		feat := &geojson.Feature{
			Geometry:&geojson.Geometry{Type:"Point",Point:pt},
			Properties:map[string]interface{}{"COLORKEY":color,"LAT":pt[1],"LONG":pt[0],"AREA":area},
		}
		feats = append(feats,feat)
	}

	//fc := geojson.FeatureCollection{Features:feats}
	//shit,_ := fc.MarshalJSON()
	//ioutil.WriteFile("a.geojson",[]byte(shit),0666)
	return feats
}


// debugging a cast na mean
func (cast Cast) Debug(x1,x2 float64) *geojson.Geometry {
	pt1 := []float64{x1,cast.Segment1.Interpolate(x1)}
	pt2 := []float64{x2,cast.Segment1.Interpolate(x2)}

	pt4 := []float64{x1,cast.Segment2.Interpolate(x1)}
	pt3 := []float64{x2,cast.Segment2.Interpolate(x2)}

	line := [][]float64{pt1,pt2,pt3,pt4,pt1}
	return &geojson.Geometry{Type:"Polygon",Polygon:[][][]float64{line}}

}

// casting down each cast until a within is found
// if none is found an empty map is returned
func (column Column) Debug(geohash string) []*geojson.Feature {
	size := get_size_geohash(geohash)
	midx := Get_Middle(geohash)[0]
	deltax := size.deltaX / 2.0
	x1,x2 := midx - deltax,midx + deltax
	feats := []*geojson.Feature{}
	colormap := map[uint64]string{}

	for _,cast := range column.Casts {
		hash, _ := h.Hash(*cast.Area, nil)

		color,boolval := colormap[hash]
		if boolval == false {
			colormap[hash] = colorkeys[rand.Intn(len(colorkeys)-1)]
			color = colormap[hash]	
		}
		props := *cast.Area
		props["COLORKEY"] = color
		feats = append(feats,&geojson.Feature{Geometry:cast.Debug(x1,x2),Properties:props})
	}
	return feats
}


// debugging functions
func (tile_index Tile_Index) Debug(geohash string) []*geojson.Feature {
	return tile_index.Index[geohash].Debug(geohash)
}

func Pointize(feats []*geojson.Feature) []*geojson.Feature {
	newfeats := []*geojson.Feature{}
	for _,feat := range feats {	
		//color := colorkeys[rand.Intn(len(colorkeys)-1)]
		for _,line := range feat.Geometry.Polygon {
			for _,pt := range line {
				newfeat := &geojson.Feature{Geometry:&geojson.Geometry{Point:pt,Type:"Point"},Properties:map[string]interface{}{"COLORKEY":"orange"}}
				newfeats = append(newfeats,newfeat)
			}
		}
	}
	return newfeats
}






