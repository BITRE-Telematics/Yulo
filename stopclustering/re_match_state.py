import pandas as pd 
import geopandas as gpd
from multiprocessing import Pool
from shapely.geometry import Point
from functools import partial
import numpy as np
import neo4j
from sklearn.cluster import DBSCAN
from scipy.spatial import cKDTree  
import argparse
from datetime import datetime
from collections import deque
import yaml
'''This script would ideally use py2neo instead of the neo4j bolt driver but I can't get it to connect'''


def stop_query(tx, CODE):
	query = "MATCH (v:Vehicle)-[:STOPPED_AT]->(s:Stop) WHERE left(s.sa2, $LEN) = $CODE RETURN s.id as id, s.lat as lat, s.lon as lon, v.firm as firm, v.id as vehicle"
	records = [record for record in tx.run(query,
	 CODE = CODE, 
 	 LEN  = len(CODE))]
	if len(records) == 0:
		print("No stops for %s" % (CODE))
		exit()
	df = pd.DataFrame([r.values() for r in records], columns = records[0].keys())
	return df


def pull_stops(creds, code):
	g = neo4j.GraphDatabase.driver("bolt://%s" % creds['ipport'], auth = (creds['username'], creds['password']))
	code_len = len(code)
	with g.session() as session:
		ps = session.read_transaction(stop_query, CODE = code)
		#in case clustering is being rerun with new data
		
		session.run("MATCH(s:Stop)-[u:USED]->(l:Location) WHERE left(s.sa2, $LEN) = $CODE DELETE u", LEN = code_len, CODE = code)
		
	return ps




def lz_rest_match(df, creds, locs, max_dist = 150):
	dfrest = ckdnearest(df, locs, 'loc_id', max_dist)

	##Previously code would not match queensland rest areas. This seems to have fixed itself.
	#
	#dfqld = ckdnearest(df, locsqld, 'loc_id', max_dist)
	#df = pd.concat([dfrest, dfqld])
	#df = pd.merge[locs[['type', 'loc_id']]]
	#df = df[df.type == 'Rest_area' or df.distance <= lz_threshold]
	return dfrest

def get_ras_lz(creds):
	g = neo4j.GraphDatabase.driver("bolt://%s" % creds['ipport'], auth = (creds['username'], creds['password']))
	with g.session() as session:
		locs = session.read_transaction(loc_query)
	return locs

def loc_query(tx):
	query = "MATCH (l:Location) RETURN l.id as loc_id, l.lat as lat, l.lon as lon, labels(l)[1] as type"
	records = [record for record in tx.run(query)]
	df = pd.DataFrame([r.values() for r in records], columns = records[0].keys())
	return df



def write_lz_rest(row, creds, node_type):
	db = 'bolt://%s' % creds['ipport']
	col = 'cluster' if node_type == 'Cluster' else 'id'
	query = 'MATCH(c:%s{id: {%s}})\
					 MERGE(l:Location{id:$loc_id})\
					 CREATE(c)-[:USED]->(l)' % (node_type, col)
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	with g.session() as session:
		session.run(query, row)


def ckdnearest(gdA, gdB, bcol, max_dist = 100):   
    nA = np.array(list(zip(gdA.lon, gdA.lat)) )
    nB = np.array(list(zip(gdB.lon, gdB.lat)) )
    btree = cKDTree(nB)
    dist, idx = btree.query(nA, k=1, n_jobs = -1) ##this is fairly fast but can use n_jobs to parallelise. -1 is all cores
    dist = dist * 110250 ## turning euclidean distance into approxiamte metres
    df = pd.DataFrame.from_dict({'distance': dist.astype(int),
                             bcol : gdB.loc[idx, bcol].values })
    df = pd.concat([gdA, df], axis = 1)
    df = df[df.distance <= max_dist] ##this could be done with distance_upper_bound = 1/110250 *100 in btree.query()
    return df


if __name__ == "__main__":

	now = datetime.now()

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--creds", type = str, default = "../Graphupload/neo4jcredsWIN.yaml",
		help="credential yaml for database")
	parser.add_argument("-y", "--year", type = int, default = now.year,
		help="year on which to cluster paired with month")
	parser.add_argument("-ac", "--code", type = str, default = '0',
						help="beginning of ASGS region code")

	parser.add_argument("-md", "--max_dist", type = float, default = 150,
						help="Max distance to be associated with a site")
	args = parser.parse_args()

	





	with open(args.creds, 'r') as credsfile:
		creds = yaml.safe_load(credsfile)

	print("Pulling stops for state %s" % (args.code))
	stops = pull_stops(creds, args.code)

	
	print('matching %s stops to rest areas and loading zones' % len(stops.index))
	locs = get_ras_lz(creds = creds)
	stops_loc = lz_rest_match(stops, locs = locs, max_dist = args.max_dist, creds = creds)
	print('Writing edges between %s locations and %s stops' % (len(set(stops_loc.loc_id)), len(stops_loc.index)))
	stops_loc = [dict(c[1]) for c in stops_loc.iterrows()]
	write_lz_rest_s =partial(write_lz_rest, creds = creds, node_type =  "Stop")
	with Pool() as p:
		p.map(write_lz_rest_s, stops_loc)


	

	 

