# shapefile
Created Friday 27 April 2018

This script merely converts outputs from the above into shapefiles for potential use in leaflet or qgis.

Note that due to the limitations of the shp format variable names can be truncated.

I had a problem with some shapefiles where road segment osm-id == 386098582, which at the link (as of 2017-06-21) is shown to be a short segment, was erroneously identified as the entire in-construction bypass. This meant that, when linked to the shapefile, it appeared as if vehicles were being matched to the in-construction road. This was because the shapefile was out of date. When the database is updated, the shapefile should be reconstructed using ogr2ogr as in process.sh.

