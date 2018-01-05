<<<<<<< HEAD
# vt_index_experiment

Distributed point in polygon processing, hopefully will be pretty fast. This is like the third version of a project like this I've written each time getting quite a bit more clean. The issues before were memory footprint and while that still might be an issue I'm inclined to believe it probably wont be as big of an issue to be honest but will find out. Currently writing this while batch processing a US county set (it does take a signicant amount of time to make ~10-20 minutes with counties with something a little more dense it may take longer. Have yet to do any performance metrics or any shit like that. 

It does have a much cleanier api in every aspect reading, writing, base structs, and the usage of a specific protobuf file instead of just just hacking together something with vector tile's protobuf spec. The main performance advantage this project uses compared to the one previously is it points to a structure easily interpolatble and gets the y values from ray casting via interpolation therefore, we don't need to store every unique Y. However complexity structure wise is quite a bit mroe crazy.
=======
# vt_index
An api for dealing with distributed geospatial indexes about vector tiles.

# What is it? 
This is a hacky api for taking combined layers of polygons and combined them using [layersplit](https:/github.com/murphy214/layersplit) from their I take an api I wrote a while back that uses vector tiles at a defined zoom level and mapbox vector_tile pbf structure (although this may change) to provide a super fast api to perform point in polygon operations insanely fast, in that specific vector tile. (millions of points / second) This API attepts to apstract on that idea idea by wrapping the byte data of each vector tile index in a sqllite db pretty similiar to the mbtiles spec. By doing this you can build out multiplexing websockets or everything completely in memory loading only a certain amount into cache at a time. 

So this api creates a sqllite data structure that is used in memory to load specific vector tile indexs into a map by mapping a point to a specific tile then mapping the point in that tile to a polygon that exists in it. This single polygon will could represent 100s of layers of varying layers consisting of multiple fields each. Essentially this library in conjunction with layersplit is like a super direct implementation of two pretty hard problems being:
  * layersplit - takes a two layers (although each could be combined from two layers themselves) and combines them not just combining and splitting all intersecting polygons but IMPORTANTLY performing the difference on each layer as well so if say a state has an area where no zip code actually exists its still makes it into the combined layer
  * vt_index - point in polygon on a distributed scale using for as many layers / fields within each layer as you want which will be pretty cool 
 
**Currently works but its super hacky I've rewritten both these modules like 3 times in the past 2 days.**
>>>>>>> a66414ac28fea63a51fd8c495a380856aa6bef04
