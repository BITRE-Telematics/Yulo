
#!/bin/bash
#set -e
ulimit -n 65000
cd osrm
##to use Docker hash this
./osrm-routed ../barefoot/map/osm/australia.osrm &
## andunhash this
#docker run -t -i -p 5000:5000 -v ${PWD}/../barefoot/map/osm/:/data osrm-server osrm-routed /data/australia.osrm --name="osrm-server" &

cd ../barefoot
docker start barefoot-australia


java -jar target/barefoot-0.1.1-matcher-jar-with-dependencies.jar --debug config/server.properties config/australia.properties

