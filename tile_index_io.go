package exp 

import (
	"encoding/json"
	h "github.com/mitchellh/hashstructure"
	"fmt"
	"github.com/murphy214/vt_index/tile_index"
	"github.com/golang/protobuf/proto"
	geo "github.com/paulmach/go.geo"
	util "github.com/murphy214/mbtiles-util"
	m "github.com/murphy214/mercantile"
)


type Struct_Map struct {
	Cast_Map map[Cast]int32
	Casts []*tile_index.Cast
	Segment_Map map[Segment]int32
	Segments []*tile_index.Segment
	Area_Map map[uint64]int32 
	Areas []string
	Column_Map map[uint64]int32
	Columns []*tile_index.Column
}

// adds a single segment to the segment map and segment list
func (structmap *Struct_Map) Add_Segment(segment Segment) int32 {
	val,ok := structmap.Segment_Map[segment]
	if ok == true {
		return val
	} else {
		structmap.Segments = append(structmap.Segments,&tile_index.Segment{Segment:[]float64{segment.B,segment.X,segment.Slope}})
		val := int32(len(structmap.Segments) - 1)
		structmap.Segment_Map[segment] = val
		return val
	}

	return int32(0)
}

// adding an area
func (structmap *Struct_Map) Add_Area(area map[string]interface{}) int32 {
	hash, _ := h.Hash(area, nil)
	val,ok := structmap.Area_Map[hash]
	if ok == true {
		return val
	} else {
	    bytevals, _ := json.Marshal(area)
	    structmap.Areas = append(structmap.Areas,string(bytevals))
		val := int32(len(structmap.Areas) - 1)
		structmap.Area_Map[hash] = val
		return val
	}
	return int32(0)
}

// adding a single cast
func (structmap *Struct_Map) Add_Cast(cast Cast) int32 {
	val,ok := structmap.Cast_Map[cast]
	if ok == true {
		return val
	} else {
		seg1 := structmap.Add_Segment(*cast.Segment1)
		seg2 := structmap.Add_Segment(*cast.Segment2)
		area := structmap.Add_Area(*cast.Area)
		structmap.Casts = append(structmap.Casts,&tile_index.Cast{Cast:[]int32{seg1,seg2,area}})
		val := int32(len(structmap.Casts) - 1)
		structmap.Cast_Map[cast] = val
		return val

	}

	return int32(0)
}

// writes an entire column of data
func (structmap *Struct_Map) Add_Column(column *Column) int32 {
	hash, _ := h.Hash(column, nil)
	val,ok := structmap.Column_Map[hash]
	if ok == true {
		return val
	} else {
		castints := []int32{}
		for _,cast := range column.Casts {
			castints = append(castints,structmap.Add_Cast(*cast))
		}
		structmap.Columns = append(structmap.Columns,&tile_index.Column{Column:castints})
		val := int32(len(structmap.Columns) - 1)
		structmap.Column_Map[hash] = val
		return val		
	}
	return int32(0)
}

// delineating tilemap 
func Get_Inds(totalmap map[string]int32,tileid m.TileID) []int32 {
	// g 
	bds := m.Bounds(tileid) 
	ghash1 := geo.NewPoint(bds.W,bds.N).GeoHash(9)
	ghash2 := geo.NewPoint(bds.E, bds.N).GeoHash(9)
	gpt1 := Get_Middle(ghash1)
	//gpt2 := Get_Middle(ghash2)
	size := get_size_geohash(ghash1)
	//ghashmap := map[string]int{ghash1:0}
	currentghash := ghash1
	count := 0
	currentpt := gpt1
	newlist := []int32{}
	for currentghash != ghash2 {
		currentpt = []float64{currentpt[0] + size.deltaX,bds.N}
		count += 1
		currentghash =  geo.NewPoint(currentpt[0],currentpt[1]).GeoHash(9)

		val,boolval := totalmap[currentghash] 
		if boolval == true {
			newlist = append(newlist,val)
		} else {
			newlist = append(newlist,-1)
		}

	}

	return newlist

}


func Write_Tile_Index(tindex Tile_Index,tileid m.TileID,mbtile util.Mbtiles) {
	segment_map := map[Segment]int32{}
	cast_map := map[Cast]int32{}
	area_map := map[uint64]int32{}
	column_map := map[uint64]int32{}

	structmap := Struct_Map{Segment_Map:segment_map,Cast_Map:cast_map,Area_Map:area_map,Column_Map:column_map}
	totalmap := map[string]int32{}
	for k,v := range tindex.Index {
		val := structmap.Add_Column(v)
		totalmap[k] = val
	}
	inds := Get_Inds(totalmap,tileid)

	total_index := tile_index.Tile_Index{Lat:tindex.Lat,Areas:structmap.Areas,Segments:structmap.Segments,Casts:structmap.Casts,Columns:structmap.Columns,Properties:inds}

	bytevals,_ := proto.Marshal(&total_index)
	//fmt.Println(totalmap)
	mbtile.Add_Tile(tileid,bytevals)
	
}

// gets all the segments
func Get_Segments(segments []*tile_index.Segment) []*Segment {
	new_segs := []*Segment{}
	for _,seg := range segments {
		new_segs = append(new_segs,&Segment{B:seg.Segment[0],X:seg.Segment[1],Slope:seg.Segment[2]})
	}
	return new_segs
}

// gets all the areas
func Get_Areas(areas []string) []map[string]interface{} {
	new_areas := []map[string]interface{}{}
	for _,area := range areas {
		var m map[string]interface{}
		json.Unmarshal([]byte(area), &m)
		new_areas = append(new_areas,m)
	}
	return new_areas
}

// gets all the casts
func Get_Casts(casts []*tile_index.Cast,segments []*Segment,areas []map[string]interface{}) []*Cast {
	new_casts := []*Cast{}
	for _,cast := range casts {
		seg1ind := cast.Cast[0]
		seg2ind := cast.Cast[1]
		areaind := cast.Cast[2]
		new_casts = append(new_casts,&Cast{Segment1:segments[seg1ind],Segment2:segments[seg2ind],Area:&areas[areaind]})
	}
	return new_casts
}

// gets all columns
func Get_Columns(columns []*tile_index.Column,casts []*Cast) []*Column {
	new_columns := []*Column{}
	for _,column := range columns {
		temp_casts := []*Cast{}
		for _,ind := range column.Column {
			temp_casts = append(temp_casts,casts[ind])
		}
		new_columns = append(new_columns,&Column{Casts:temp_casts})
	}
	return new_columns
}


// 
func Read_Tile_Index(bytevals []byte,tileid m.TileID) Tile_Index {
	tindex := &tile_index.Tile_Index{}
	if err := proto.Unmarshal(bytevals, tindex); err != nil {
		fmt.Printf("Failed to parse address book: %s\n", err)
	}
	//fmt.Printf("%+v\n",tindex)
	tileindex := map[string]*Column{}
	segments := Get_Segments(tindex.Segments)
	areas := Get_Areas(tindex.Areas)
	casts := Get_Casts(tindex.Casts,segments,areas)
	columns := Get_Columns(tindex.Columns,casts)
	
	bds := m.Bounds(tileid) 
	ghash1 := geo.NewPoint(bds.W,bds.N).GeoHash(9)
	ghash2 := geo.NewPoint(bds.E, bds.N).GeoHash(9)
	gpt1 := Get_Middle(ghash1)
	//gpt2 := Get_Middle(ghash2)
	size := get_size_geohash(ghash1)
	//ghashmap := map[string]int{ghash1:0}
	currentghash := ghash1
	count := 0
	currentpt := gpt1
	//fmt.Println(sizeval)
	sizeval := len(tindex.Properties)
	for currentghash != ghash2  {
		var val int32
		currentpt = []float64{currentpt[0] + size.deltaX,bds.N}
		count += 1
		if sizeval > count {
			val = tindex.Properties[count]
		}
		currentghash =  geo.NewPoint(currentpt[0],currentpt[1]).GeoHash(9)
		if val != -1 {
			tileindex[currentghash] = columns[val]
		}



	}

	return Tile_Index{Index:tileindex,Lat:tindex.Lat}
}