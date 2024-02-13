# schema
Created Friday 08 June 2018

This page describes the fields of the nodes and the types of relation ships in the Neo4j database as os 2019-04-18. The structure is not disimilar to the graph [in this page](./graph_databases.md), which I will update in time. Attribute type is in ()

The schema follows a style. 
Relationship types (edges) are IN_CAPITALS
Node labels have Initial_capitals
Attributes are all_lowercase
Neo4j does not allow one to asset type of attributes so care should be taken to make sure only these types are uploaded.


**NODES**
Asset (Vehicle, Trailer):
 id (string) - A unique vehicle identifier
type (string) - the type of asset
firm (string) - the firm that provided the data
In addition each Asset node has an additional label Vehicle, or Trailer.

	


Stop:
id (string) - A unique stop identifier constructed from the vehicle id and a hex of the start datetime
start_time (int) - the start time in local time in unix epoch format
end_time (int) - the end time in local time in unix epoch format
start_time_utc (int) - the start time in UTC in unix epoch format
   end_time_utc (int) - the end time in UTC in unix epoch format
start_timedt (datetime) - the start time in local time
end_timedt (datetime) - the end time in local time
start_time_utcdt (datetime) - the start time in UTC
end_time_utcdt (datetime) - the end time in UTC
lat (float) - latitude
lon (float) - longitude
sa2 (int) - sa2 code
gcc (string) - ASGS Greater Capital City Statistical Area name
added: time uploaded in unix epoch (to aid removal in debugging)

Cluster:
id (string): A unqiue cluster identifier
lat (float): The mean latitude of stops in the cluster
lon (float): The mean longitude of stops in the cluster
year (int): The year of the cluster
month: The month of the clister
anon (bool): Indicating whether the cluster comprises at least x firms and y vehicles

Trip:
id (string) - A unique trip identifier constructed from the vehicle id and a hex of the start datetime

Note, as of 2020-07-22 observations of type 'imputed' have been elimated and replaced with multple edges from the same node.
Observation:
speed (Float) - recorded speed where available (only from actual pings)
datetime (int) - the start time in local time in unix epoch format (due to python behaviour these are in error in the database prior to 2019-05-01)
datetime_utc (int) - the start time in UTC in unix epoch format
datetimedt (datetime) - the start time in local time
datetime_utcdt (datetime) - the start time in UTC
lat (float) - latitude
lon (float) - longitude
imputed_speed (float) - the imptued average speed between two matched points attributed to all points on the route (calculated in [MatcherKState.java](../Processing/Barefoot/MatcherKState.java.md).)
date (string) - date in YYYY-MM-DD in local time
azimuth (integer) - where available azimuth expressed as decimal degrees (0-359)
type (string) - either "matched no path" a recorded observation with a route (path) leading to it, "matched path", or "imputed", and observation imputed from the matching of recorded points 
add_date (int) - date the observation was added to the database in unix epoch, local Canberra time, in case it needs to be reconciled with a state of OSM data.
olson (string): The olson timezone
target (string): The osm_id of the node towards which the vehicle is heading, as determined by barefoot
forward (bool): Whether the vehicle is heading "forward" as determined by barefoot's process
imputed_azimuth (float): where barefoot has matched to a point on a segment the azimuth of the segment at that point (given 'forward') and where the point is imputed, the azimuth at the mid point of that segment
length(float): the length leading up to a matched segment in metres as calculated to barefoot. This is geodesic length. From September 2019 segments follwing a subtrip break (ie some of type "matched no path" will have a length derived from the [OSRM](Map%20data/OSRM.md) engine.

Segment: - note I am applying the term "segment" to what is better known in OSM parlance as a "way". Properly speaking a segment is some fraction of a way between nodes
osm_id (string) - the osm id, as described [here](osm_id.md) 
name (string) - the common name for the road
highway (strong) - the classification given by osm, one of those in [road-types.json](../Processing/Barefoot/road-types.json.md)
data_date(int) - date the segment attributes and geometry were last updated in unix epoch local Canberra time
length(float) - the geodesic length of the segment
wkt (string) - the geometry in well known text format
*variable_breakdown *(list of float) - precalculated breaks downs of <variable> by time unit <type>. The list is ordered but not labelled.
   *dir*_*variable_breakdown* (list of float) - precalculated breaks downs by direction. The suffix is either "fw" - forward - or "bw" - "backward - 
forward (string) - the median ordinal direction of the segment going forward
   backward (string) - the median ordinal direction of the segment going backwards
updated (string) - when the breakdown variables were last uploaded
	
Other attributes that are attributed to a segment in OSM data are also added

Request: (for when a segment is queried via the app - for analyzing what users are interested in)
time (string) - the datetime (I haven't bothered coverting into unix epoch etc yet)
ir (string) - request for either imputed or recorded data
slice (string) - the time unit by which the data was requested (ie broken down by hour, day etc)
graf_x (string) - the data to be graphed (speed or volumes)
ip (string) - the ip address of the request, if available
city (string) - the city of the request from the ip information
loc (string) - approximate location of the request from the ip information
	
Location (Rest_area, Loading_zone):
lat (float)
lon (float)
id (string)
name(string)
And a bunch of booleans and misc attributes
	
**RELATIONSHIPS (expressed in **[cypher](./neo4j/cypher.md)** syntax)**

(Vehicle)- [:STOPPED_AT]->(Stop)
(Vehicle)-[:EMBARKED_ON]->(Trip)
(Trailer)- [:STOPPED_AT]->(Stop)
(Trailer)-[:EMBARKED_ON]->(Trip)
(Stop)-[:PRECEDED_BY]->(Trip)
(Stop)-[:FOLLOWED_BY]->(Trip)
(Stop)-[:LAST_STOP]->(Stop)
(Stop)-[:NEXT_STOP]->(Stop) (these last two are the same but with directionality reversed. It could easily be implemented as a single relationship but I have chosen to improve readability of queries at the expense of doubling the number of such relationships, which is trivial in full context)
(Stop)-[:PART_OF]->(Cluster)
(Stop)-[:USED]->(Location)
(Cluster)-[:USED]->(Location)
(Stop)-[:AT]->(Address)
(Cluster)-[:AT]->(Address)
(Trip)-[:OBSERVED_AT]->(Observation)
(Observation)-[:ON]->(Segment) - as of 2020-07-22 I have eliminated observation nodes with the imputerd attribute and repaced them with multiple edges from the matched observation. These edges have attributes "type" (string) and "forward" (bool)
(Segment)-[:REQUESTED_AT]->(Request)



