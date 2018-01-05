# vt_index
An api for dealing with distributed geospatial indexes about vector tiles.

# What is it? 
This is a hacky api for taking combined layers of polygons and combined them using [layersplit](https:/github.com/murphy214/layersplit) from their I take an api I wrote a while back that uses vector tiles at a defined zoom level and mapbox vector_tile pbf structure (although this may change) to provide a super fast api to perform point in polygon operations insanely fast, in that specific vector tile. (millions of points / second) This API attepts to apstract on that idea idea by wrapping the byte data of each vector tile index in a sqllite db pretty similiar to the mbtiles spec. By doing this you can build out multiplexing websockets or everything completely in memory loading only a certain amount into cache at a time. 

So this api creates a sqllite data structure that is used in memory to load specific vector tile indexs into a map by mapping a point to a specific tile then mapping the point in that tile to a polygon that exists in it. This single polygon will could represent 100s of layers of varying layers consisting of multiple fields each. Essentially this library in conjunction with layersplit is like a super direct implementation of two pretty hard problems being:
  * layersplit - takes a two layers (although each could be combined from two layers themselves) and combines them not just combining and splitting all intersecting polygons but IMPORTANTLY performing the difference on each layer as well so if say a state has an area where no zip code actually exists its still makes it into the combined layer
  * vt_index - point in polygon on a distributed scale using for as many layers / fields within each layer as you want which will be pretty cool 
 
**Currently works but its super hacky I've rewritten both these modules like 3 times in the past 2 days.**
>>>>>>> a66414ac28fea63a51fd8c495a380856aa6bef04
