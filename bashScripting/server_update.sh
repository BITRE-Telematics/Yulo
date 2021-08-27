#!/bin/bash
##must be run with sudo
#set -e
# #fuser -k 1234/tcp
docker stop barefoot-australia
docker rm barefoot-australia
#yes | docker system prune
cd barefoot



mvn clean




curl http://download.geofabrik.de/australia-oceania/australia-latest.osm.pbf -o map/osm/australia.osm.pbf


##update shapefile derived from http://wiki.openstreetmap.org/wiki/User:Bgirardot/How_To_Convert_osm_.pbf_files_to_Esri_Shapefiles

ogr2ogr -overwrite -f "ESRI Shapefile" ../shapefiles/AustraliaRoads.shp map/osm/australia.osm.pbf -progress -sql "select osm_id,name,highway,other_tags from lines where highway is not null or other_tags like '%ferry%'" OSM_CONFIG_FILE = ../Docs_and_configs/osmconf.ini
#ogr2ogr -overwrite -f "geojson" ../shapefiles/AustraliaRoads.geojson ../shapefiles/AustraliaRoads.shp -progress
ogr2ogr  -f Geojson ../shapefiles/AustraliaRoads.geojson map/osm/australia.osm.pbf -progress -sql "select osm_id,name,highway,other_tags from lines where highway is not null or other_tags like '%ferry%'" OSM_CONFIG_FILE = ../Docs_and_configs/osmconf.ini
##if changing to web mercator
#ogr2ogr -overwrite -f "ESRI Shapefile" ../shapefiles/AustraliaRoads.shp map/osm/australia.osm.pbf -progress -sql 'select osm_id,name,highway,other_tags from lines where highway is not null' OSM_CONFIG_FILE = ../Docs_and_configs/osmconf.ini -s_srs EPSG:4326 -t_srs EPSG:3857
##if simplifying with the commandline mapshaper interface
#node --max-old-space-size=20000 `which mapshaper` ../shapefiles/AustraliaRoads.shp -simplify 5% -o ../shapefiles/out.shp




cd ../Graphupload
python3 addSegments.py -g=False

cd ../barefoot

mvn package -DskipTests
docker build -t barefoot-map ./map
docker run -it -d -p 5432:5432 --name="barefoot-australia" -v ${PWD}/map/:/mnt/map barefoot-map

sleep 30s ##otherwise the psql isn't running properly
##This could be added as CMD in the dockerfile as CMD ["bash", "/mnt/map/osm/import.sh"]
docker exec barefoot-australia bash /mnt/map/osm/import.sh


##update shapefile derived from http://wiki.openstreetmap.org/wiki/User:Bgirardot/How_To_Convert_osm_.pbf_files_to_Esri_Shapefiles

cd ../osrm
## hash out these if using Docker
./osrm-extract -p car.lua ../barefoot/map/osm/australia.osm.pbf 
./osrm-contract "../barefoot/map/osm/australia.osrm"
## and unhash these
# docker build -t osrm-server ./
# docker run --rm -t -v ${PWD}/../barefoot/map/osm/:/data osrm-server osrm-extract -p /opt/car.lua /data/australia.osm.pbf
# docker run --rm -t -v ${PWD}/../barefoot/map/osm/:/data osrm-server osrm-contract /data/australia.osrm