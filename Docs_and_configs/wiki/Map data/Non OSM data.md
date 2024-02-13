`server_update.sh` assumes matching will be performed on OpenStreetMap data. This can be circumvented to use other data by creating an osm formatted xml file. This can be done using the tool `geojson2osm2` javascript tool as such

```
npm i https://github.com/GeoWonk/geojson2osm2
npx geojson2osm2 original.geojson > barefoot/map/osm/australia.osm

```

and altering `server_update.sh` to use execute `import_xml.sh` in place of `import.sh`, as well as directing osrm to extract from the osm file rather than osm.pbf. Note the geojson must include `highway` tags that are accounted for in `road-types.json` which can be edited as needed.

Non OSM sources include proprietary sources and also the Geosience Australia national road sets https://digital.atlas.gov.au/datasets/national-roads