	#!/bin/bash
set -e
mkdir ../shapefiles
curl http://www.ausstats.abs.gov.au/ausstats/subscriber.nsf/0/A09309ACB3FA50B8CA257FED0013D420/\$File/1270055001_sa2_2016_aust_shape.zip -o ../shapefiles/sa_2016.zip
curl http://www.ausstats.abs.gov.au/ausstats/subscriber.nsf/0/65819049BE2EB089CA257FED0013E865/\$File/1270055001_ste_2016_aust_shape.zip -o ../shapefiles/ste_2016.zip

unzip ../shapefiles/sa2_2016.zip
unzip ../shapefiles/ste_2016.zip
rm ../shapefiles/*.zip
##geojson files for broken geopandas
python3 remove_null_geom.py

echo "If using ABS shapefiles remove null geometries"
ogr2ogr -f Geojson ../shapefiles/SA2_2016_AUST.geojson ../shapefiles/SA2_2016_AUST.shp
ogr2ogr -f Geojson ../shapefiles/STE_2016_AUST.geojson ../shapefiles/STE_2016_AUST.shp
python3 wkt_convert.py
#mv *wkt.csv ../goyulo/src/yuloserver/shapefiles
