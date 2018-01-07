# Vt_Index

Point-In-Polygon by tile processing via a interpolation engine. 

# File Structure 
* **create.go** - helper functions for debugging nasty corner cases 
* **mapx** - assembles the columns and tile_indexs for all the features within a tile (most of computation happens here) 
* **mb_index** - a few outer level handler functions for shoving the protobuf compressed structure into an sqlite db currently
* **poly_envelope** - a small clipping script used to create the clipping tile indexes about each tile with the structure map[m.TileID][]*geojson.Feature{}
* **tile_index_io.go** - a set of routines for shoving a tile_index into a smaller protobuffer datastructure for serialization to file (some of the things done include flattening out arrays to only a segment array and then using index references to assemble the casts, and columns and at the top level tile_indexes respectively.
* **tile_ind.go** - base structure for pip in polygon about one tile index 

# Benchmarks 

Assuming your dataset can fit into memory it should be around ~500-1000k points a second, however memory is an issue, with denser polygon sets but I will eventually construct something to reads sequentially as needed against an organized point set. 

# Mimimal Example 


