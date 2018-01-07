package exp

/*
This portion of code creates the tile_index from a set of polygons.

*/

import (
	m "github.com/murphy214/mercantile"
	geo "github.com/paulmach/go.geo"
	"math"
	"strings"
	"sort"
	//"fmt"
	"github.com/murphy214/rdp"
	//"fmt"
	"github.com/paulmach/go.geojson"
	"math/rand"

	//"io/ioutil"
)


// Point represents a point in space.
type Size struct {
	deltaX float64
	deltaY float64
	linear float64
}


// returns a data structure contain the size of your deltaX deltaY and linear distance (accross the box)
// given a geohash return the linear distance from one corner to another
func get_size_geohash(ghash string) Size {
	ex := get_extrema_ghash(ghash)
	size := math.Sqrt(math.Pow(ex.N-ex.S, 2) + math.Pow(ex.E-ex.W, 2))
	return Size{ex.E - ex.W, ex.N - ex.S, size}
}

// returns a data structure contain the size of your deltaX deltaY and linear distance (accross the box)
// given a geohash return the linear distance from one corner to another
func Get_size_geohash(ghash string) float64 {
	ex := get_extrema_ghash(ghash)
	return ex.E - ex.W
}

func geoHash2ranges2(hash string) (float64, float64, float64, float64) {
	latMin, latMax := -90.0, 90.0
	lngMin, lngMax := -180.0, 180.0
	even := true

	for _, r := range hash {
		// TODO: index step could probably be done better
		i := strings.Index("0123456789bcdefghjkmnpqrstuvwxyz", string(r))
		for j := 0x10; j != 0; j >>= 1 {
			if even {
				mid := (lngMin + lngMax) / 2.0
				if i&j == 0 {
					lngMax = mid
				} else {
					lngMin = mid
				}
			} else {
				mid := (latMin + latMax) / 2.0
				if i&j == 0 {
					latMax = mid
				} else {
					latMin = mid
				}
			}
			even = !even
		}
	}
	return lngMin, lngMax, latMin, latMax
}

// gets the extrema object of  a given geohash
func get_extrema_ghash(ghash string) m.Extrema {
	w, e, s, n := geoHash2ranges2(ghash)
	return m.Extrema{S:s, W:w, N:n, E:e}
}

// gets the extrema object of  a given geohash
func Get_Middle(ghash string) []float64 {
	w, e, s, n := geoHash2ranges2(ghash)
	return []float64{(w + e) / 2.0, (s + n) / 2.0}
}


// fills in the x values from a point on the polygon line representation to
// the next point on the line and fills the appopriate geohashs between them
// this section can be though of as the raycasting solver for geohash level we desire
// currently defaults to size 9
func fill_x_values(pt1 []float64, pt2 []float64, sizes Size, latconst float64) map[string]Segment {
	// creating temporary map
	tempmap := map[string]Segment{}
	seg := New_Segment(pt1,pt2)
	if seg.Slope != 1000000.0 {

		// getting the geohashs for the relevant points
		ghash1 := geo.NewPoint(pt1[0], pt1[1]).GeoHash(9)
		ghash2 := geo.NewPoint(pt2[0], pt2[1]).GeoHash(9)

		// decoding each geohash to a point in space
		long1 := Get_Middle(ghash1)[0]
		long2 := Get_Middle(ghash2)[0]

		//long1 := geo.NewPointFromGeoHash(ghash1).Lng()
		//long2 := geo.NewPointFromGeoHash(ghash2).Lng()

		// sorting are actual longs
		longs := []float64{pt1[0], pt2[0]}
		sort.Float64s(longs)

		// getting potential longs
		// this is mainly just to check the first one
		potlongs := []float64{long1, long2}
		sort.Float64s(potlongs)
		potlongs[1] = potlongs[1] + sizes.deltaX
		potlongs[0] = potlongs[0] - sizes.deltaX

		//sizes.deltaX = sizes.deltaX 
		xcurrent := potlongs[0]

		var ghash string
		//total := [][]float64{}
		//tempmap[]
		for xcurrent <= potlongs[1] {
			ghash = geo.NewPoint(xcurrent, latconst).GeoHash(9)

			if (xcurrent >= longs[0]) && (xcurrent <= longs[1]) {
				//total = append(total, []float64{xcurrent, pt.Y})
				tempmap[ghash] = seg
			}
			xcurrent += sizes.deltaX 
		}
		//if (xcurrent >= longs[0]) && (xcurrent <= longs[1]) {
		//	ghash = geo.NewPoint(xcurrent, latconst).GeoHash(9)

			//total = append(total, []float64{xcurrent, pt.Y})
		//	tempmap[ghash] = seg
		//}
		//tempmap[ghash] = &seg


	} else {
	}
	return tempmap

}

func Reverse(vals []float64) []float64 {
	count := len(vals) -1
	newvals := make([]float64,len(vals))
	count2 := 0
	for count != -1 {
		newvals[count2] = vals[count]
		count2 += 1
		count -= 1 
	}
	return newvals
}


func Remove_Doubles(vals []float64) []float64 {
	oldi := vals[0]
	newvals := []float64{}
	for _,i := range vals[1:] {
		if oldi == i {
			if len(newvals) > 0 {
				newvals = newvals[:len(newvals)-1]
			}
		} else {
			newvals = append(newvals,oldi)			
		}
	}
	if newvals[len(newvals)-1] == oldi {
		newvals = newvals[:len(newvals)-2]
	} else {
		newvals = append(newvals,oldi)
	}	
	return newvals
}

// sorts the segments out from a raw list of segments
func Sort_Segments(geohash string,segments []Segment) []Segment {


	// getting the middle piont
	point := Get_Middle(geohash)
	
	// getting the float list and map
	floatmap := map[float64]Segment{}
	floatlist := []float64{}
	for _,v := range segments {

		floatval := v.Interpolate(point[0])
		//fmt.Println(floatval,v,geohash)
		floatlist = append(floatlist,floatval)
		floatmap[floatval] = v
		//fmt.Println(v,floatval,geohash)
	}
	


	floatlist = Unique_Floats(floatlist)

	// sorting the floatlist
	sort.Float64s(floatlist)
	//fmt.Println(floatlist,geohash)
	//floatlist = Remove_Doubles(floatlist)
	//fmt.Println(floatlist,geohash)


    for i, j := 0, len(floatlist)-1; i < j; i, j = i+1, j-1 {
        floatlist[i], floatlist[j] = floatlist[j], floatlist[i]
    }	//fmt.Println(floatlifloatlistt,geohafloatlisth)

	// iterating through sorted floatlist
	// iterating through sorted floatlist
	newsegs := []Segment{}
	for _,k := range floatlist {
		newsegs = append(newsegs,floatmap[k])
	}


	return newsegs
}




// given a set of x,y coordinates returns a map string that will be used as the base
// string for constructing our geohash tables
// this is essentially the most important data structure for the algorithm
func Make_Xmap(coords [][][]float64, areaval map[string]interface{}, bds m.Extrema) map[string]Column {
	// quick lint
	N := bds.N

	// sizing a single geohash like this for now
	ghash := geo.NewPoint(coords[0][0][0], coords[0][0][1]).GeoHash(9)

	sizes := get_size_geohash(ghash)

	// linting coord values
	//coords = lint_coords(coords)
	boolval := false 
	if boolval == true {
		for ii := range coords {
			//pre := len(coords[ii])
			coords[ii] = rdp.RDPSimplify(coords[ii],sizes.linear)
			//fmt.Println(len(coords[ii]),pre)
		}
	}

	// getting coords extrema

	// intialization variables
	latconst := N - .0000000001
	//coords = Easy_Lint(coords)
	//coords = Lint_Polygon(coords,latconst)

	coords = append(coords, coords[0])
 
	oldpt := []float64{0.0,0.0}
	count := 0
	topmap := map[string][]Segment{}
	for _,coord := range coords {
		// iterating through each coordinate collecting each fill_x_values output
		count = 0
		lastcoord := coord[len(coord)-1]
		firstcoord := coord[0]
		if (firstcoord[0] == lastcoord[0]) && (firstcoord[1] == lastcoord[1]) {

		} else {
			coord = append(coord,coord[0])
		
		}	

		for _,pt := range coord {
			if count == 0 {
				count = 1
			} else {
				//go func(oldpt Point,pt Point,sizes Size,latconst float64,ccc chan<- []float64) {
				tempmap := fill_x_values(oldpt, pt, sizes, latconst)
				for k, v := range tempmap {
					topmap[k] = append(topmap[k], v)

				}

			}
			oldpt = pt

		}
	}
	
	// creating outer level map but sorting
	newmap := map[string]Column{}
	for k, v := range topmap {
		v = Sort_Segments(k,v)
		column := Column{}
		newlist := []Segment{}
		//newlist2 := [][]float64{}
		boolval := true
		//fmt.Println(v)
		//oldi := v[0]
		for _,i := range v {
			newlist = append(newlist,i)
			
			if len(newlist) == 2 {

				if boolval == true {
					column.Casts = append(column.Casts,New_Cast(&newlist[0],&newlist[1],areaval,Get_Middle(k)[0]))
					//column.Casts = append(column.Casts,&Cast{Segment1:newlist[0],Segment2:newlist[1],Area:&areaval})
					//boolval = false
					newlist = []Segment{}
				
				} else {
					newlist = []Segment{}
					//boolval = true

				}
			}
			
		}
		newmap[k] = column

	}
	

	//fmt.Print(len(topmap), "\n\n")
	return newmap

}

// adds all missing geohash columns
func Lint_Total_Index(bds m.Extrema,total_index map[string]Column) map[string]Column {
	point := []float64{bds.W,bds.N}
	ghash := geo.NewPoint(point[0], point[1]).GeoHash(9)
	point = Get_Middle(ghash)
	sizes := get_size_geohash(ghash)
	delta := sizes.deltaX 

	latconst := point[1]
	currentx := point[0]
	for currentx < bds.E {
		currentx += delta
		ghash = geo.NewPoint(currentx, latconst).GeoHash(9)
		_,boolval := total_index[ghash]
		if boolval == false {
			total_index[ghash] = Column{}
		}
	}

	return total_index
}

// creates a tile_index from a given series of polygons and a tileid
func Make_Xmap_Polygons(feats []*geojson.Feature,tileid m.TileID) Tile_Index {
	bds := m.Bounds(tileid)
	c := make(chan map[string]Column)
	for _,feat := range feats {
		go func(feat *geojson.Feature,c chan map[string]Column) {
			c <- Make_Xmap(feat.Geometry.Polygon,feat.Properties,bds)
		}(feat,c)
	}

	total_index := map[string]*Column{}
	for range feats {
		temp_index := <-c
		//fmt.Println(temp_index,"here")
		for k,v := range temp_index {
			val,boolval := total_index[k]
			if boolval == false {
				total_index[k] = &Column{}
				val = total_index[k]
			}

			//fmt.Println(v.Casts,val,"casts")
			val.Casts = append(val.Casts,v.Casts...)
			total_index[k] = val
		}
	}
	return Tile_Index{Index:total_index,Lat:bds.N-.000000001}
}


func Shit(index Tile_Index,bds m.Extrema) []*geojson.Feature {
	point := []float64{bds.W,bds.N}
	ghash := geo.NewPoint(point[0], point[1]).GeoHash(9)
	point = Get_Middle(ghash)
	sizes := get_size_geohash(ghash)
	delta := sizes.deltaX 

	latconst := point[1]
	currentx := point[0]
	newlist := []*geojson.Feature{}
	for currentx < bds.E {
		currentx += delta
		ghash = geo.NewPoint(currentx, latconst).GeoHash(9)
		mid := Get_Middle(ghash)
		midx := mid[0]
		column := index.Index[ghash]
		for _,cast := range column.Casts {
			pt1 := []float64{midx,cast.Segment1.Interpolate(midx)}
			pt2 := []float64{midx,cast.Segment2.Interpolate(midx)}
			area := *cast.Area
			area["GHASH"] =  ghash
			newfeat1 := &geojson.Feature{Geometry:&geojson.Geometry{Point:pt1,Type:"Point"},Properties:area}
			newfeat2 := &geojson.Feature{Geometry:&geojson.Geometry{Point:pt2,Type:"Point"},Properties:area}
			newlist = append(newlist,newfeat1)
			newlist = append(newlist,newfeat2)

		}

	}
	return newlist
}

func Unique_Floats(input []float64) []float64 {
	u := make([]float64, 0, len(input))
	m := make(map[float64]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}


func Easy_Lint(coords [][][]float64) [][][]float64 {
	newcoords := [][][]float64{}
	for _,coord := range coords {
		newline := [][]float64{}
		var oldpt []float64 
		for i,pt := range coord {
			if i != 0 {
				if ((oldpt[0] == pt[0]) && (oldpt[1] == pt[1])) == false {
					newline = append(newline,oldpt)
				}
			}
			oldpt = pt
		}
		newline = append(newline,oldpt)
		newcoords = append(newcoords,newline)
	}
	return newcoords
}

// 
func Lint_Coords(coords [][]float64,latconst float64) [][]float64 {
	var oldghash,ghash string
	var oldpt []float64
	newcoords := [][]float64{}
	for i,pt := range coords {
		ghash = geo.NewPoint(pt[0], pt[1]).GeoHash(9)
		if i != 0 {
			if oldghash != ghash {
				newcoords = append(newcoords,oldpt)
			}
		}
		oldghash = ghash
		oldpt = pt
	}
	newcoords = append(newcoords,oldpt)
	return newcoords
}

func Lint_Polygon(coords [][][]float64,latconst float64) [][][]float64 {
	for i,coord := range coords {
		coords[i] = Lint_Coords(coord,latconst)
	}
	return coords
}



// random point x
func RandomPt_X(bds m.Extrema,X float64) []float64 {
	deltay := math.Abs(bds.N - bds.S)
	return []float64{X, (rand.Float64() * deltay) + bds.S}
}

// get a random point
func RandomPt(bds m.Extrema) []float64 {
	deltax := math.Abs(bds.W - bds.E)
	deltay := math.Abs(bds.N - bds.S)
	return []float64{(rand.Float64() * deltax) + bds.W, (rand.Float64() * deltay) + bds.S}
}

func Rand_Each_Geohash(index Tile_Index,bds m.Extrema) []*geojson.Feature {
	point := []float64{bds.W,bds.N}
	ghash := geo.NewPoint(point[0], point[1]).GeoHash(9)
	point = Get_Middle(ghash)
	sizes := get_size_geohash(ghash)
	delta := sizes.deltaX 

	latconst := point[1]
	currentx := point[0]
	feats := []*geojson.Feature{}
	point = RandomPt_X(bds,0.0)
	randy := point[1]

	for currentx < bds.E {
		currentx += delta
		ghash = geo.NewPoint(currentx, latconst).GeoHash(9)
		mid := Get_Middle(ghash)
		midx := mid[0]
		point = []float64{midx,randy}

		if len(*index.Pip(point)) > 0 {
			feats = append(feats,&geojson.Feature{Geometry:&geojson.Geometry{Point:point,Type:"Point"},Properties:*index.Pip(point)})
		}
	}
	return feats
}
