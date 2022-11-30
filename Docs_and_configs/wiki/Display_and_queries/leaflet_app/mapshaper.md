# mapshaper
Created Wednesday 30 May 2018

Mapshaper is a tool for simplifying vectors in shapefiles and geojson by removing vertices. This can speed up display in browser at the expense of accurately mirroring the real life geography at close resolution.

It can be invoked with rmapshaper::ms_simplify in R, or with a command line tool which requires installation of a recent version of node.js and exceptions for memory management. At the moment the default algorithm seems to be producing broken shapefiles so use the other option 'dp' (for instance in the leaflet app's Global.r script)

It has now been superseded by mpabox

