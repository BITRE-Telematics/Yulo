# Yuloserver
Created Friday 25 October 2019

Yulo server is an attempt to port the entire process encompassing [Tripgrouping](./Tripgrouping.md), [batch mod.py](./Barefoot/batch_mod.py.md), [postbarefoot](./postbarefoot.md) and the up[loadscripts](../Storage_of_processed_data/uploadscripts.md) into a single server applciation written in Go. It takes a csv fed in by an http protocol and processes the file until upload, interacting with the [neo4j](../Storage_of_processed_data/neo4j.md) database, the barefoot server and an [OSRM](./postbarefoot/OSRM.md) server using their respective interfaces. This makes the system more robust, reduce disk read and write times and increases speed particularly in TripGrouping.

It accepts files in the same format as Tripgrouping via an curl interface from the command line or a scripting language as needed.. An example query is
	curl -X POST -F 'myFile=@<DATA>CSV>' 0.0.0.0:<PORT>/process 
	


These files can be any number of vehicles but the whole file will be read into memory. I may adjust the read in script to avoid this. The parameter max_routines in config.yaml limits the number of vehicles being processed concurrrently from a given file.

As on 2019-11-25 it processes vehicles in batches for a given time period (for instance a month) passing all the data sequentially through the stages of processing. In future I intend to use channels more so that a trip can be passed directly from the tripgrouping process to barefoot and to barefoot without waiting for the whole batch for that vehicle to be completed. 

There is an inexplicable bug on some vehicles when feeding trips to Barefoot (in bffeed.go). Yuloserver will generate a valid json string but this will receive an empty reply from the barefoot server. However if this JSON is dumpted to disc and submitted via netcat it works. It is always the same trips but there are no identifiable special characters or coding issues. For the rare number of trips where this happens the error handling will dump to disk and submit from there.

There are two alternative processes for uploading. One uploads from within Go and is not tested. The other creates CSVs and submits them to a transfer server running on the machine running the database and then submits READ_CSV calls via cypher. These upload calls are run sequentially rather than concurrently to avoid locking issues.

In time this can be adapted to produce real time processing with direct insertion of data from providers

--Geocoding--
The geocoding in yulo server was initially slower than in Python. This, I assume, is because the Geopandas package is a direct port of Geos (in C) and the vectorisation is also in a low level language. Raw Golang implementations did not compare well. I have changed to golang implementations of GEOS but this has required the generation of wkt versions of the geometries since the geos libraries available have not been matched with geojson read in capacities.

