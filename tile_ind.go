package exp

/*
Experimental data structure where:
	* Segment represents interpolation structure
	* Cast represents the checking of a point between two lines
	* Column represents a single geohash column of values
	* Tile_Index Represents a single tile

NOTES:
This structure will probably use its own buffer and utilize a set of arrays for segments still need to think how to utilize memory structure best

Advanantages:
	* Hopefully a much smaller memory footprint


*/

import (
	geo "github.com/paulmach/go.geo"
)


// gets the slope of two points along a line
// if statement logic accounts for undefined corner case
func Get_Slope(pt1 []float64, pt2 []float64) float64 {
	if pt1[0] == pt2[0] {
		return 1000000.0
	}
	return (pt2[1] - pt1[1]) / (pt2[0] - pt1[0])
}

// a segment representing the space between two points
type Segment struct {
	B float64
	X float64
	Slope float64
}

// iterpolates a point
func (segment Segment) Interpolate(x float64) float64 {
	return (x - segment.X) * segment.Slope + segment.B
}

// creates a new segment
func New_Segment(pt1 []float64,pt2 []float64) Segment {
 	slope := Get_Slope(pt1,pt2)
	return Segment{B:pt1[1],X:pt1[0],Slope:slope}
}

// structure that is casting upon two lines
// the segments represent the upper line 1 and lower line2
type Cast struct {
	Segment1 *Segment
	Segment2 *Segment
	Area *map[string]interface{}
}

// creates a new segment
func New_Cast(segment1 *Segment,segment2 *Segment,area map[string]interface{},x float64) *Cast {
	if segment1.Interpolate(x) > segment2.Interpolate(x) {
		dummy := segment2
		segment2 = segment1
		segment1 = dummy
	}
	return &Cast{Segment1:segment1,Segment2:segment2,Area:&area}
}

// checks to see if a given cast structure contains the given point
func (cast Cast) Within(point []float64) bool {
	a := cast.Segment1.Interpolate(point[0])
	b := cast.Segment2.Interpolate(point[0])
	return ((a >= point[1]) && ( point[1] >= b) || (a <= point[1]) && ( point[1] <= b))
}


// structure representing one column of data
// ideally this should be sorted by largest delta or something
type Column struct {
	Casts []*Cast
}

// casting down each cast until a within is found
// if none is found an empty map is returned
func (column Column) Within(point []float64) *map[string]interface{} {
	for _,cast := range column.Casts {
		if cast.Within(point) == true {
			return cast.Area
		}
	}
	return &map[string]interface{}{}
}

// a tile index structure for which points can be queried
type Tile_Index struct {
	Index map[string]*Column
	Lat float64
}

// the outer api point in polygon function query for a single tile_index structure
func (tile_index Tile_Index) Pip(point []float64) *map[string]interface{} {
	val,boolval := tile_index.Index[geo.NewPoint(point[0], tile_index.Lat).GeoHash(9)]
	if boolval == true {
		return val.Within(point)
	} else {
		return &map[string]interface{}{}
	}
}



 







