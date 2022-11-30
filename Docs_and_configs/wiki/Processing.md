# Processing
Created Friday 01 February 2019

These scripts and processes form the basic processing.

For a given vehicle it must go through

[TripGrouping](./Processing/Tripgrouping.txt) - Where trips and stops will be defined and JSON produced for mapmatching

[Barefoot](./Processing/Barefoot.txt) - matching observations to roads and intuiting further data

[Postbarefoot](./Processing/postbarefoot.txt) - merging matched data and geocoding overvations and stops for uploading

[Graphuploading](./Storage_of_processed_data/neo4j/uploadscripts) - Adding observations and stops to the database


In addition the following scripts must be run periodically

[+stopclustering](./Processing/stopclustering.md) - defining stops that are clustered spatially and temporally that identify important locations and those that are anonymous

[addvehicledata](./Storage_of_processed_data/neo4j/uploadscripts) and addsegments - incorporating data on vehicles and segments into the database

