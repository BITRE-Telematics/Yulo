import neo4j
import yaml
import pandas as pd 
import geopandas as gpd
import re
from multiprocessing import Pool
from functools import partial, reduce
import argparse
from geopy.distance import distance, lonlat
import numba
from re import sub, IGNORECASE
from neobolt.exceptions import CypherSyntaxError, ClientError, SecurityError
import json
import numpy as np

'''
Future options are whether I split the other_tags to add all under the relevant tag, and whether I upload the geometry
'''

def geocode_segs(shp, SA2):
  shp.crs = SA2.crs
  ##add nearest neighbourcode
  shp = gpd.sjoin(shp, SA2, op = 'within', how = 'left')
  shp = shp.drop([ "SA2_NAME21", "CHG_FLAG21", "CHG_LBL21", "SA3_CODE21",
   "SA3_NAME21", "SA4_CODE21", "SA4_NAME21", "GCC_NAME21", "STE_CODE21",
    "STE_NAME21", "AUS_CODE21", "AUS_NAME21", "AREASQKM21", "LOCI_URI21"], axis = 1)
  shp = shp.rename(columns = {'SA2_CODE21': 'SA2', "GCC_CODE21": 'GCC'})
  shp.SA2 = shp.SA2.fillna('0')
  shp.SA2 = shp.SA2.apply(lambda x: '0' if x == '9' else x) ## depending on results maybe just add to the dictionary as NSW time
  ##posisble fix for geo_code errors
  if '0' in set(shp.SA2):
    shp = fix_geocode(shp, SA2)
  else:
    shp = pd.DataFrame(shp)
  shp.geometry = shp.geometry.apply(lambda x: str(x))
  return shp

def fix_geocode(df, SA2):
  df_notmatched = df[df.SA2== '0']
  points = [x.centroid for x in df_notmatched.geometry]
  nearest_poly1 = partial(nearest_poly, SA2 = SA2)
  with Pool() as p:
    ind = p.map(nearest_poly1, points)
 

  df_notmatched.SA2 = [SA2.SA2.iloc[i] for i in ind]
  df_notmatched.GCC = [SA2.GCC.iloc[i] for i in ind]
  
  df = pd.concat([pd.DataFrame(df[df.SA2!='0']), pd.DataFrame(df_notmatched)])
  
  
  return df

def nearest_poly(point, SA2):
  #print("nearest poly matching %s" % point)
  dist = np.array([[point.distance(x) for x in SA2.geometry]])
  ind = dist.argmin()
  return ind
  

def format_tags(row, seg_upload):
  tags_dict = {}
  a = row['other_tags'][1:-1].split('","')
  b = [x.replace('"=>"', '=') for x in a]
  c = [x for x in b if \
       'maxspeed' not in x and \
       '=' in x
   ]

  d = [sub('[:\\.`@]', '', x) for x in c]
  e = [sub('-', '_', x) for x in d]
  f = [sub(r'\\', "'", x) for x in e]
  for t in f:
    tag, value = t.split('=')[0], t.split('=')[1]
    tag = sub( ' ', '_', tag)
    tag = sub('^(\\d)', 'is\\1', tag)
    tags_dict[tag] = value
    ##iterate to add booleans to tag2
    func = ['toBoolean(', ')'] if value in ['yes', 'no'] else ['', '']
    func = ['toInteger(', ')'] if tag == 'lanes' else ['', '']
    ##this is yucky because the % formatting won't work when there's curly brackets which doesn't use but
    ##will accpet it in the str.format() method where they are used ::shrug::
    #seg_upload = seg_upload + ', segment.{}={}{{{}}}{}'.format(tag, func[0], tag, func[1])
    #neo4j 4.0 style
    seg_upload = seg_upload + ', segment.%s=%s $%s %s' % (tag, func[0], tag, func[1])
  ##dict.update() changes the object in place
  row.update(tags_dict)
  return row, seg_upload

'''Barefoot import splits tags like this, left here for reference, where row[1] is other_tags'''
# tags = dict((k.strip(), v.strip()) for k, v in (
#         item.split("\"=>\"") for item in row[1][1:-1].split("\", \"")))


def writer(row, creds, args):
  print(row)
  seg_upload = "MERGE(segment:Segment{\
                  osm_id: $osm_id\
                  })\
                 SET segment.name = $name, segment.highway = $highway, segment.speed_limit = toInteger($speed_lim), segment.wkt = segment.geometry, \
                 segment.data_date = timestamp()/1000"

  
  if args.other_tags == 'True':
    row, seg_upload = format_tags(row, seg_upload)

  if args.geocode == 'True' : 
    seg_upload = seg_upload + ', segment.gcc = $GCC, segment.sa2 = toString($SA2)'
  if args.length == 'True' : 
    seg_upload = seg_upload + ', segment.length = toFloat($length)'             
  db = 'bolt://%s' % creds['ipport']
  g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
  with g.session() as session:
    try :
      session.run(seg_upload, row)
    except (ClientError, SecurityError, CypherSyntaxError) as e :
      print("Error uploading segment : %s" %row['osm_id'])
      print(e)
      #print(seg_upload)
      #print(row)
      print('\n')

# ##The weird slicing is to reverse the tuple because geopy likes lat lons. Once can also use geopy.distance.lonlat
# def calc_length_imp(road):
#   xy_string = road.coords
#   dist = 0
#   for i in range(1, len(xy_string)):
#     dist = dist + distance(xy_string[i-1][::-1], xy_string[i][::-1]).meters
#   return(dist)
  
##The below is quite slow so maybe try geog.distance np.sum(geog.distance(line[:-1, :], line[1:, :]))
 
@numba.jit
def dist_map(aggDict, x):
  dist =  distance(aggDict['point'][::-1], x[::-1]).meters +  aggDict['dist']
  aggDict['dist'], aggDict['point'] = dist, x
  return(aggDict)

@numba.jit
def calc_length(road):
  xy_string = road.coords
  aggDict = {'dist': 0, 'point': (xy_string[0])}
  aggDict = reduce(dist_map, xy_string, aggDict)
  return(aggDict['dist'])

def seg_sa2_pull(tx):
  query = "MATCH(s:Segment) RETURN s.osm_id as osm_id, s.sa2 as SA2, s.gcc as GCC"
  records = [record for record in tx.run(query)]
  if len(records) == 0:
    exit()
  df = pd.DataFrame([r.values() for r in records], columns = records[0].keys())
  return df


if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument("-c", "--creds", type = str, default = "neo4jcredsWIN.yaml",
                        help="credential yaml for database")
  parser.add_argument("-f", "--file", type = str, default = "../shapefiles/AustraliaRoads.geojson",
                        help="file with segment data")
  parser.add_argument("-g", "--geocode", type = str, default = 'True',
                        help="geocode segments")
  parser.add_argument("-o", "--other_tags", type = str, default = 'True',
                        help="add other tags")
  parser.add_argument("-l", "--length", type = str, default = 'False',
                        help="compute segment length")
  parser.add_argument("-j", "--json", type = str, default = '../barefoot/map/tools/seg_sa2s.json',
                        help="location of json output for segment geooding to SA2")
  args = parser.parse_args()



  with open(args.creds, 'r') as credsfile:
    creds = yaml.safe_load(credsfile)
  

  
  print("Reading in file")
  roads = gpd.read_file(args.file)
  print('Processing %s roads' % len(roads.index))
  
  if args.length == 'True':
    with Pool() as p:
      print('Calculating lengths')
      roads['length'] = roads.geometry.apply(lambda x: calc_length(x))
  
  if args.geocode == "True":
    print('Geocoding')
    SA2 = gpd.read_file('../shapefiles/SA2_2021_AUST.geojson')

    SA2 = SA2[[g is not None for g in SA2.geometry]]
    ##parallelise this

    roads = geocode_segs(roads, SA2)

    sa2_dict = dict(zip(roads.osm_id, zip(roads.SA2, roads.GCC)))
    with open(args.json, 'w') as f:
      json.dump(sa2_dict, f)

  else:
    roads = pd.DataFrame(roads)
    roads.geometry = roads.geometry.apply(lambda x: str(x))
  
  print('Processing tags')
  roads.other_tags = roads.other_tags.fillna("")
  roads.highway = roads.apply(lambda x: 'ferry' if 'ferry' in x.other_tags else x.highway, axis = 1)
  roads['speed_lim'] = roads.other_tags.apply(lambda x: re.sub(r".*maxspeed\"=>\"([0-9]+).*" , r"\1", x))
  roads.speed_lim = roads.speed_lim.apply(lambda x: x if "=>" not in x else "NA")


  roads = roads.iterrows()
  roads = [dict(r[1]) for r in roads] ##I could put this in the writer function to speed it up

  db = 'bolt://%s' % creds['ipport']
  g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))


  print('writing roads')
  writer1 = partial(writer, creds = creds, args = args)
  with Pool(1) as p:
    p.map(writer1, roads)



  
  with g.session() as session:
    roads_db = session.read_transaction(seg_sa2_pull)

  sa2_dict = dict(zip(roads_db.osm_id, zip(roads_db.SA2, roads_db.GCC)))
  with open(args.json, 'w') as f:
    json.dump(sa2_dict, f)
  


  
  	


