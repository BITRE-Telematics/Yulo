# static maps
Created Wednesday 23 May 2018

Static maps can be created using a number of tools including ggmap in R, using ggplot2 combined with raster images from servers including google etc.
The road data can be plotted with ggmap one of two ways
The first is reading in the data from the shapefile, setting the CRS and then fortifying it for display with ggmap using geom_path(). This takes some time
The other is using sf::st_read to read it in as a simple features object which can be rendered with geom_sf(). If used with ggmap this requries two considerations
1) set "inherit.aes = F" so geom_sf() does not try to apply geom_map aesthetics to ggplot
2) The sf objects must have it's projection set as web mercator [(EPSG:3857](./EPSG/3857.md)) so they align with the rasters pulled by get_map
I can't actually fix number 2. get_map projects at 3857, but ggmap takes 4326. If the SF object is in 3857 the coordinates are outside the bounds of the raster layer defined in 4326 terms, however if the SF object is in 4326 the lines do not render above their equivalent roads because they are projected differently.
