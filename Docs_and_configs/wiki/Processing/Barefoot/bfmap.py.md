# bfmap.py
Created Tuesday 01 May 2018

bfmap.py is a script within the barefoot library which builds a map database from OSM data using a geofabrik extract and using parameters from [road-types.json.](./road-types.json.md)
I have edited it to include the function is_noaccess() to remove service roads and bus ways that were being improperly matched to, but which vehicles could never use.

I have edited it further to account for particular [GPS map matching errors](../Matching_issues_and_errors/GPS_error.md) where changing properties are insufficient. Since this is a small number of prominent roads have edited segment() to take a third argument with troublesome osm_ids whree a different weight will be attributed than that from [road-types.json](./road-types.json.md). This also required changing the function bfmap2ways() as well, to accept and then apply the extra argument, ways2bfmap.py to accept an argument of a filename and a function to read in a file and finally import.sh to add a default argument and argument to australia.properties with the relevant file.

Note the original code in Barefoot has been considerably rewritten by a contributor. I haven't yet adopted this with my own modifications.

