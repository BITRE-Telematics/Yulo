	#!/bin/bash
set -e
mkdir ../shapefiles
curl https://www.abs.gov.au/statistics/standards/australian-statistical-geography-standard-asgs-edition-3/jul2021-jun2026/access-and-downloads/digital-boundary-files/STE_2021_AUST_SHP_GDA2020.zip -o ../shapefiles/ste_2021.zip
curl https://www.abs.gov.au/statistics/standards/australian-statistical-geography-standard-asgs-edition-3/jul2021-jun2026/access-and-downloads/digital-boundary-files/SA2_2021_AUST_SHP_GDA2020.zip -o ../shapefiles/sa2_2021.zip

unzip ../shapefiles/sa2_2021.zip -d ../shapefiles
unzip ../shapefiles/ste_2021.zip -d ../shapefiles

rm ../shapefiles/*.zip
##geojson files for broken geopandas


echo "If using ABS shapefiles remove null geometries"
ogr2ogr -f Geojson ../shapefiles/SA2_2021_AUST.geojson ../shapefiles/SA2_2021_AUST_GDA2020.shp
ogr2ogr -f Geojson ../shapefiles/STE_2021_AUST.geojson ../shapefiles/STE_2021_AUST_GDA2020.shp

python3 remove_null_geom.py
python3 wkt_convert.py

#mv *wkt.csv ../goyulo/src/yuloserver/shapefiles
