
To build a dockerised instance try (with creds.yaml in dir)
mkdir sc_dockerised
cp addSegments.py as_dockerised/. & cp creds.yaml as_dockerised/. & cp Dockerfile as_dockerised/. & cp requirements.txt as_dockerised/.

cd as_dockerised
sudo docker build -t addSegments .
sudo docker run -it -v $PWD:/creds/ -v /path/to/barefoot/map/tools:/jsondir/ -v /path/to/shapefiles:/shapefiles/ --rm --name addSegments addSegments -g=False -c /creds/creds.yaml -j /jsondir -f /shapefiles/AustraliaRoads.geojson
mounting whatever dir has a creds file in place of $PWD
  
