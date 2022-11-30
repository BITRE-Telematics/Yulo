# orientdb
Created Friday 27 April 2018

Gremlin and pyorient notes.
Gremlin is specialised language for accessing graph databases, which can be interfaced through the pyorient library of python. Gremlin is based on the idea of traversing a graph from a particular point, for instance a segment node, trip node etc. Here is a brief example of pyorient and gremlin code that extracts summary information about the observations of a road segment at a given time.

	import pyorient
	from statistics import median
	client = pyoritent.OrientDB(“localhost”, 2424)
	client.db_open(“FPMP”, “username”, “psswd”)
	road_id = “234562562”
	weekday = 7
	
	##concatenates a gremlin command that finds the road ID and extracts all observations connected to it that are on a Sunday
	gremobvs = “g.V(‘osm_id’, ‘%s’).inV.filter{weekday == %d}” %(road_id, weekday)
	
	#extracts the recspeed from these vertices and finds the mean
	print(“The mean recorded speed for %s on weekday %d is %d kph) %(road_id, weekday, client.gremlin(gremobvs + “recspeed.mean()”.)
	#orientdb is not set up for vector operations easily, and several summary statistics, including median, are not built in, hence code like the following is necessary which will iterate through the values and add them to a list for processing.
	
	speeds = [obv.recspeed for obv in client.gremlin(recobvs)]
	print(“The median recorded speed for %s on weekday %d is %d kph” (road_id, weekday, median(speeds))

Assuming we are writing to the database using a python script from the postbarefootmerged.r files (or replacing it with a pure python script)…. Note word messes up indentation here. Assuming one trip per file and no schema

	import numpy as np
	import pandas as pd
	data = pd.read_csv(file).sort_values(by = [‘datetime’]).reset_index(drop = True)
	
	##assuming one trip per file
	
	vehicle = data.vehicle[0]
	if not client.Gremlin(“g.V(‘VehicleID’, %s).hasNext()” vehicle):
		client.Gremlin(“g.addVertex(‘VehicleID, %s)” vehicle)
	trip = data.trip[0]
	#the following command both adds the vertex and retains the unique database id
	#Functions including adding vertices return pyorient objects which identify the vertex etc. These are returned in lists, even if there is only one element, hence the [0].
	tripid = client.Gremlin(“g.addVertex(‘tripID’, %s)” %trip)[0]._rid
	client.Gremlin(“tripvert = g.v(‘%s’); vehvert = g.V.has(‘vehicle’, ‘%s’).next();g.addEdge(vehvert, tripvert, ‘embarked on’” %(tripid, vehicle)
	##function to add just speed. Lat lons and other speeds can be done whenever. 
	def addObv(tripid, time, speed, osm_id):
		#add obv vert
		if not client.Gremlin(“g.V(‘osm_id’, ‘%s’).hasNext()” %osm_id):
			client.Gremlin(“g.addVertex(‘osm_id’, %s)” %osm_id)
	client.Gremlin(“tripvert = g.v(‘%(tripid)s’);obvsvert = g.addVertex(‘time’, %(time)d,    ‘speed’, %(speed)d); roadvert = g.V.has(‘osm_id’, ‘(osm_id)s’)”.next();g.addEdge = (obsvert, roadvert, ‘on’); g.addEdge(tripvert, obsvert, “observed at”)” % \
	{‘time’: time, ‘osm_id’: osm_id, ‘speed’: speed, ‘tripid’: tripid})
	
	for i in data.index:
		addObv(tripid, data.datetime[i], data.speed[i], data.osm_id[i])
	
	Other notes on syntax
	Adding a vertex of class
	g.addVertex(‘class:osm_id’, ‘property1’, ‘property1value’);
	retrieving vertices of class
	g.V(‘@class’, ‘osm_id’)
	updating vertex properties (surprisingly hard to find a method that worked). Sideeffect is so called because they traverse the object by change it (a side effect) in doing so. Some sources say use property() but I can’t get that to work.
	g.V(‘identifyingproperty’, ‘propertyvalue’).next().sideEffect{it.changingproperty = newvalue}


It is entirely possible to add to this database using user input data from the map widget so we can see what segments are of most interest to users.

The above is entirely based on on the client interface, which may be unsuitable to write large batch amounts, although Graphupload.py does work, albeit slowly.

And alternative is to use the OGM – Object-Graph Mapper. This includes batch operations.
Update – the OGM is poorly documented but appears to be working.
Update again

The OGM interface works but does not scale and stalls out without any timeouts or error messages. This appears to be a problem with pyorient not accepting concurrency properly, so I am considering changing the upload script to Scala.

GraphuploadOGM.py uploads data by vehicle. It uses dictionaries of vertices to reduce contact with the server but these cannot be parcelled out to concurrent processes as the graph objects therein cannot be pickled. Querying the database to create the osm dictionary in particular gets very slow and may be the reason the whole thing stalls. The dictionary is necessary in particular to avoid recreating vertices.

Orientdb also has a spatial indexing capacity.

I am now (2017-01-16) considering shifting to neo4j, especially as it has a CSV upload facility that seems well suited to the task for which I am failing.

