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
  SA2 = SA2.rename(columns = {'SA2_CODE21': 'SA2', "GCC_CODE21": 'GCC'})
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
  df_notmatched = df[df.SA2== '0'].copy()
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
  


def format_tags_dict(rows):
  for row in rows:
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
      
      if value =='yes':
        value = True 
      elif value =='no':
        value = False
      elif (value.count(".") == 1 and value.replace(".", "").isdigit()):
        value = float(value) 
      elif value.isdigit():
        value = int(value) 
      
      tags_dict[tag] = value
    
    row["OTHER_TAGS"] = tags_dict
    _ = row.pop('other_tags')
      
  
  return rows



def writer_batch(rows, creds, args):
  
  seg_upload = "UNWIND $ROWS as r\
                 MERGE(segment:Segment{\
                  osm_id: r.osm_id\
                  })\
                 SET segment.name = r.name, segment.highway = r.highway, segment.speed_limit = toInteger(r.speed_lim), segment.wkt = segment.geometry, \
                 segment.data_date = timestamp()/1000\
                 "
  
  
  if args.length == 'True' : 
    seg_upload = seg_upload + ', segment.length = toFloat(r.length)'   
  if args.geocode == 'True': 
    seg_upload = seg_upload + ', segment.gcc = r.GCC, segment.sa2 = toString(r.SA2)\
    WITH segment, r \
    OPTIONAL MATCH(segment)-[in:IN]->(sa:SA2) DELETE in \
    MERGE(matched_sa2:SA2{sa2_code: toString(r.SA2)})\
    CREATE (segment)-[:IN]->(matched_sa2) \
         '

  if args.other_tags == 'True':
    print('adding other tags')
    rows['ROWS'] = format_tags_dict(rows['ROWS']) 
    seg_upload = seg_upload +  " WITH segment, r UNWIND keys(r.OTHER_TAGS) AS k WITH segment, r, k, properties(r.OTHER_TAGS) as props \
      MERGE(t:Tag{tag_id:k})\
      MERGE(t)<-[ht:HAS_TAG]-(segment)\
      SET ht.value = props[k]\
      "
  
  def seg_batch(tx, batch):
    tx.run(seg_upload, batch)
  
  db = 'bolt://%s' % creds['ipport']
  g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
  with g.session(database = creds['segs_db']) as session:
    try :
      session.write_transaction(seg_batch, rows)
    except (ClientError, SecurityError, CypherSyntaxError) as e :
      print("Error uploading batch " )
      print(e)
      
      #print(row)
      print('\n')


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
  parser.add_argument("-c", "--creds", type = str, default = "/creds/neo4jcredsWIN.yaml",
                        help="credential yaml for database")
  parser.add_argument("-f", "--file", type = str, default = "/shapefiles/AustraliaRoads.geojson",
                        help="file with segment data")
  parser.add_argument("-g", "--geocode", type = str, default = 'True',
                        help="geocode segments")
  parser.add_argument("-o", "--other_tags", type = str, default = 'True',
                        help="add other tags")
  parser.add_argument("-l", "--length", type = str, default = 'False',
                        help="compute segment length")
  parser.add_argument("-j", "--json", type = str, default = '/barefoot/map/tools/seg_sa2s.json',
                        help="location of json output for segment geooding to SA2")
  parser.add_argument("-bs", "--batch_size", type = int, default = 10000,
                        help="No of roads per batch")
  args = parser.parse_args()



  with open(args.creds, 'r') as credsfile:
    creds = yaml.safe_load(credsfile)
  
  writer_batch1 = partial(writer_batch, creds = creds, args = args)
  
  
  if args.geocode == "True":
      print('Geocoding in this session')
      SA2 = gpd.read_file('/shapefiles/SA2_2021_AUST.geojson')

      SA2 = SA2[[g is not None for g in SA2.geometry]]
  else:
    print("Not geocoding in this session")
  running_tot=0
  batch_len = args.batch_size
  


  while batch_len == args.batch_size:
    roads = gpd.read_file(args.file, rows = slice(running_tot,  running_tot + args.batch_size))
    print('Processing %s roads' % len(roads.index))
    
    if args.length == 'True':
    #  with Pool() as p:
      print('Calculating lengths')
    #    roads['length'] = roads.geometry.apply(lambda x: calc_length(x))
      roads.geometry = roads.geometry.to_crs(3112)
      roads['length'] = roads.geometry.length
      roads.geometry = roads.geometry.to_crs(4326)
    
    if args.geocode == "True":
      print('Geocoding')
     
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
    roads['speed_lim'] = roads.maxspeed
    


    roads = roads.iterrows()
    roads = [dict(r[1]) for r in roads] ##I could put this in the writer function to speed it up
    # roads = roads.to_dict('records')
    db = 'bolt://%s' % creds['ipport']
    g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))

    
    print('writing roads')

    
    
    roads = {'ROWS' : roads} 
    batch_len = len(roads['ROWS'])
    
   
    
    writer_batch1(roads)
    
    running_tot= running_tot + batch_len
    print(f'Uploaded {running_tot} segments')
  
  print('Finished writing, making geojson')
  with g.session(database = creds['segs_db']) as session:
    roads_db = session.read_transaction(seg_sa2_pull)

  sa2_dict = dict(zip(roads_db.osm_id, zip(roads_db.SA2, roads_db.GCC)))
  with open(args.json, 'w') as f:
    json.dump(sa2_dict, f)
  


  
    
