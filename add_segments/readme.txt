
To build a dockerised instance try (with creds.yaml in dir)

sudo docker build -t pyneo .
sudo docker run -it -v $PWD:/script -v path/to/creds:/creds/ -v /path/to/barefoot/map/tools:/jsondir/ -v /path/to/shapefiles:/shapefiles/ --rm --name pyneo addSegments python3 addSegments.py -g=False -c /creds/creds.yaml -j /jsondir/sa2_segs.json -f /shapefiles/AustraliaRoads.geojson

  
