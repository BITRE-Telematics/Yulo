Yulo server is an attempt to port the entire process encompassing Tripgrouping, batch mod.py, postbarefoot and the uploadscripts into a single server applciation written in Go. It takes a csv fed in by an http protocol and processes the file until upload, interacting with the neo4j database, the barefoot server and an OSRM server using their respective interfaces. This makes the system more robust, reduce disk read and write times and increases speed particularly in TripGrouping.

One can also, in the yuloserver folder build a docker image of the go process
```
sudo docker build -t yuloserver ./
sudo docker run  -p 6969:6969 --network="host" --name="yuloserver" -v ${PWD}/../shapefiles/:shapefiles -v ${PWD}/../data/:/data yuloserver 
```

It assumes there is a neo4j server running with the credentials and address stipulated in creds.yaml and config.yaml. If using ys/transfer_upload.go and a neo4j installation on another machine it also requires a receiving server such as github.com/GeoWonk/receivingserver.

It accepts files in the same format as Tripgrouping via an curl interface from the command line or a scripting language as needed.. An example query is
```curl -X POST -F 'myFile=@<DATA>CSV>' 0.0.0.0:<PORT>/process ```



These files can be any number of vehicles but the whole file will be read into memory. I may adjust the read in script to avoid this. The parameter max_routines in config.yaml limits the number of vehicles being processed concurrrently but does not control memory usage.

As on 2019-11-25 it processes vehicles in batches for a given time period (for instance a month) passing all the data sequentially through the stages of processing. In future I intend to use channels more so that a trip can be passed directly from the tripgrouping process to barefoot and to barefoot without waiting for the whole batch for that vehicle to be completed. 

Currently resources are controlled solely by the size of the input files and limits on how many processes the server can run. I am implementing direct checks on system resources as an alternative.

There is an inexplicable bug on some vehicles when feeding trips to Barefoot (in bffeed.go). Yuloserver will generate a valid json string but this will receive an empty reply from the barefoot server. However if this JSON is dumpted to disk and submitted via `netcat` it works. It is always the same trips but there are no identifiable special characters or encoding issues. For the rare number of trips where this happens the error handling will dump to disk and submit from there. There is alternate code for Windows machines that assumes the instalation of `ncat` but this is not tested yet.

There are two alternative processes for uploading. One uploads from within Go and is not tested. The other creates CSVs and submits them to a transfer server running on the machine running the database and then submits READ_CSV calls via cypher. These upload calls are run sequentially rather than concurrently to avoid locking issues.

--Address Matching -- 
By default Yuloserver will try to match stops to points drawn from the database with the labels "Address" and "Location" (rest areas and loading zones). Theoretically if there are none in the database it will skip this part but just in case one can change the boolean Match_locs in config.yaml. Note that if the addresses and locations are updated one will need to restart yuloserver

--Geocoding--
The geocoding in yulo server was initially slower than in Python. This, I assume, is because the Geopandas package is a direct port of Geos (in C) and the vectorisation is also in a low level language. Raw Golang implementations did not compare well. I have changed to golang implementations of GEOS but this has required the generation of wkt versions of the geometries since the geos libraries available have not been matched with geojson read in capacities.




Yaml code is taken from github.com/xiam/yaml which fails to install normally because of broken repository links