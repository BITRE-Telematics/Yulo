# Shapefiles
Created Friday 04 May 2018

The process requires shapefiles from the ABS

<http://www.abs.gov.au/AUSSTATS/subscriber.nsf/log?openagent&1270055001_sa2_2016_aust_shape.zip&1270.0.55.001&Data%20Cubes&A09309ACB3FA50B8CA257FED0013D420&0&July%202016&12.07.2016&Latest>
<http://www.abs.gov.au/AUSSTATS/subscriber.nsf/log?openagent&1270055001_ste_2016_aust_shape.zip&1270.0.55.001&Data%20Cubes&65819049BE2EB089CA257FED0013E865&0&July%202016&12.07.2016&Latest>

[process.sh](./process.sh.md) Also includes an ogr2ogr command to transform the [geofabrik](./geofabrik.md) extract to a shapefile. Note that the sql command assumes a tabular format for the data, and hence uniform attributes in each table. The later "lines" includes all "ways". There are no universal attributes amonst the tags. As such, the extract creates an attribute "highway" as much ways are actually roads and have this attribute, and all other tags are in a field called "other tags".
