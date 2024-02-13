Valhalla https://github.com/valhalla/valhalla is a routing engine degined for OSM data. It also includes tools for mapmatching and isochrone generation. Yulo does not use the mapmatching because results were not as good as Barefoot but this might change for higher resolution data. BITRE does use the data for isochrones and time of day routing by adding encoded speed values. Notes are below 
# Documentation for Valhalla implementation

## Installation

I had some difficulties with the docker implementations missing necessary outputs. This may be fixed in later versions. I followed this [guide](https://tech.marksblogg.com/valhalla-isochrones.html) to build from source with minor changes (altered data source and installation directory). The executables also were not added to PATH and whilst I could do that manually I just called their location (ie `./build/binary` rather than `binary`).

Seperately I build a version based on [this](https://github.com/alinmindroc/valhalla_traffic_poc). This was solely to extract an extra binary `valhalla_traffic_demo_utils` which I put into the original build folder. The changed makefiles and `mjolnir` functions theoretically could be put straight into the source but this was easier for now.

## Adding traffic

As of 2023-09-29 historical traffic is still very poorly documented and took some time to work out. The basic steps are

1. After data is added obtain a list of edge ids associated with osm_ids using `valhalla_ways_to_edges --config valhalla.json`. This will produce a file `way_edges.txt`

2. An edge may be associated with more than one osm_id as two way segments are duplicated. The resulting file has comma delimited values with one osm_id per line and an even number of values representing direction and edge id. An example of extraction is

```id_df = map(way_edges, \(line){

 osm_id_ = str_split(line, ',', simplify = T)[1]

 if(osm_id_ %in% seg_data$osm_id){

 edges = str_split(line, ',', simplify = T)[-1]

 ##every second value is a valhalla id

 edges_ids = edges[seq(2, length(edges), 2)]

 return(tibble(

 osm_id_,

 edges_ids

 ))

 }

 ```

3. We need to produce four columns csvs. The first column is the edge_id but in a different format (more below), the second and third are single values for free flow and constrained traffic speeds and the last a DCT-II encoded string of 2016 values, one for each estimated speed for a 5 minute period in a week (ie `60*24*7/5`).

4. I implemented the Valhalla `c++` functions in `Rcpp` so they can be called directly in R. This is in this directory as [[`dct.cpp`]]. Because I had only calculated the predicted speeds in hour blocks, ie 24 values, I replicated each 12 times, and then replicated the single day 7 times to produce 2016 values. Because DCT-II is a lossy compressing technique this means the values are interpolated somewhat by Valhalla, as you can see below. This should not be of consquence if we are only estimating on the hour

![A graph showing values on the hour joined by jagged interpolation](decoded_speed.png)

5. The csv requires an id including `/` which also includes the directory structure that mirrors the tile directory structure created by Valhalla. [https://github.com/valhalla/valhalla-docs/blob/master/tiles.md](This is encoded in the edge_id in `way_edges`) but is troublesome to extract. Instead I used the implementation in `valhalla_traffic_demo_utils` to extract them with command line calls. For instance

```

cl = makeCluster(10, type = "FORK")

edges_ids = parSapply(cl,

 encoded_speed$edges_ids,

 \(t){system2('./build/valhalla_traffic_demo_utils', sprintf('--get-tile-id %s',t), stdout = T )})

}) |> unlist()

tile_ids = parSapply(cl,

 encoded_speed$edges_ids,

 \(t){system2('./build/valhalla_traffic_demo_utils', sprintf('--get-traffic-dir %s',t), stdout = T )})

}) |> unlist()

stopCluster(cl)

```

6. Despite the name `--get-traffic-dir` returns a file path for a csv. A given csv file will be associated with multiple edges and is in a directory structure reflecting a road hierarchy etc. Each csv is associated with one Valhalla tile. The data for a edge (id from `--get_tile_id`, free speed, constrained speed and the encoded DCT-II string) will be saved in thaat csv in that directory structure.

7. This data is added using `valhalla_add_predicted_traffic -t traffic --config valhalla.json` where `traffic` is the directory with the csvs.