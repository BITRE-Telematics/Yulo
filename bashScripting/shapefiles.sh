	#!/bin/bash
set -e
mkdir ../shapefiles
curl https://www.abs.gov.au/AUSSTATS/subscriber.nsf/log?openagent\&1270055001_sa2_2016_aust_shape.zip\&1270.0.55.001\&Data%20Cubes\&A09309ACB3FA50B8CA257FED0013D420\&0\&July%202016\&12.07.2016\&Latest -o ../shapefiles/sa2_2016.zip
curl https://www.abs.gov.au/AUSSTATS/subscriber.nsf/log?openagent\&1270055001_ste_2016_aust_shape.zip\&1270.0.55.001\&Data%20Cubes\&65819049BE2EB089CA257FED0013E865\&0\&July%202016\&12.07.2016\&Latest -o ../shapefiles/ste_2016.zip

unzip ../shapefiles/sa2_2026.zip
unzip ../shapefiles/ste_2026.zip
rm ../shapefiles/*.zip
##geojson files for broken geopandas


echo "If using ABS shapefiles remove null geometries"
ogr2ogr -f Geojson ../shapefiles/SA2_2026_AUST.geojson ../shapefiles/SA2_2026_AUST_SHP_GDA2020.shp
ogr2ogr -f Geojson ../shapefiles/STE_2026_AUST.geojson ../shapefiles/STE_2026_AUST_SHP_GDA2020.shp
python3 wkt_convert.py
python3 remove_null_geom.py
#mv *wkt.csv ../goyulo/src/yuloserver/shapefiles
