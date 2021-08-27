library(sf); library(tidyverse)

datasegs = read_csv("segspeeds.csv", col_names = c(
  "osm_id",
  "n_obvs",
  "n_trips",
  "n_vehicles",
  "LQ_imp",
  "median_imp",
  "UQ_imp",
  "stDev_imp"
)) %>%
  
  mutate(osm_id = as.character(osm_id))



names(datasegs) = stringr::str_to_lower(names(datasegs))

query_params = yaml::yaml.load_file("go_query.yaml")


dataroads = st_read("../../shapefiles/AustraliaRoads.shp") %>% 
  mutate(osm_id = as.character(osm_id)) %>% 
  left_join(datasegs, by = "osm_id") %>% 
  filter(!is.na(n_trips)) %>% 
  mutate(name = enc2utf8(as.character(name)),
         other_tags  = enc2utf8(as.character(other_tags)),
         speed_lim = gsub(".*maxspeed\\\"\\=>\\\"([0-9]+).*", "\\1",other_tags) %>% as.numeric(),
         under_lim = speed_lim - median_imp,
         under_lim = ifelse(under_lim < 0, 0, under_lim),
         #impratio = (n_obvs - n_recobvs)/n_recobvs, impratio = ifelse(is.infinite(impratio), NA, impratio),
         n_trips = ifelse( 
          ( str_detect(highway, "motorway") | str_detect(highway, "primary")) & str_detect(other_tags, "oneway\\\"\\=>\\\"[yt1]"),
           n_trips*2,
           n_trips)
         ) %>% 
  st_sf() %>%
  filter(n_trips >= query_params$min_vol) %>%
  select(-other_tags, -name, -n_obvs, -lq_imp, -uq_imp, -speed_lim) 

st_write(dataroads, "../../shapefiles/dataroads.geojson", delete_dsn = TRUE)

dataroads%>% 
  as("Spatial") %>% 
  rgdal::writeOGR("../../shapefiles/dataroads.shp", "roads", "ESRI Shapefile", overwrite_layer = T)

