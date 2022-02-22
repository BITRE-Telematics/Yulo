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


def stop_query(tx, YEAR, MONTH):
	query = "MATCH (v:Vehicle)-[:STOPPED_AT]->(s:Stop) WHERE s.start_timedt.month = $MONTH AND s.start_timedt.year = $YEAR RETURN s.id as id, s.lat as lat, s.lon as lon, v.firm as firm, v.id as vehicle"
	records = [record for record in tx.run(query,
	 MONTH = MONTH, 
 	 YEAR  = YEAR)]
	if len(records) == 0:
		print("No stops for %s-%s" % (YEAR, MONTH))
		exit()
	df = pd.DataFrame([r.values() for r in records], columns = records[0].keys())
	return df


def pull_stops(creds, year, month, del_stop_edges = False):
	g = neo4j.GraphDatabase.driver("bolt://%s" % creds['ipport'], auth = (creds['username'], creds['password']))
	with g.session() as session:
		ps = session.read_transaction(stop_query, YEAR = year, MONTH = month)
		#in case clustering is being rerun with new data
		session.run("MATCH(c:Cluster) WHERE c.month = $month AND c.year = $year DETACH DELETE c", month = month, year = year)
		if del_stop_edges:
			session.run("MATCH(s:Stop)-[u:USED]->(l:Location) WHERE s.start_timedt.month = $month AND s.start_timedt.year = $year DELETE u", month = month, year = year)
			session.run("MATCH(s:Stop)-[a:AT]->(Address) WHERE s.start_timedt.month = $month AND s.start_timedt.year =  $year DELETE a", month = month, year = year)
	return ps


def cluster_stops(df, year, month, eps = 1000, min_stops = 10):
	##Haversine returns distance in radians but my function takes metres
	df['cluster'] = DBSCAN(eps = eps/6371000, metric = 'haversine', min_samples = min_stops).fit(np.array(df[['lon', 'lat']])).labels_ + 1
	df['inCluster'] = df.apply(lambda x: x.cluster > 0, axis=1)
	df.cluster = df.apply(lambda x: '%s-%s-%s' % (year, month, x.cluster), axis =1)
	return df

## anonymity can also be checked in database

def make_row(tup, n_firms, n_veh, year, month):
	anon = len(set(tup[1].firm)) >= n_firms and len(set(tup[1].vehicle)) >= n_veh
	row =  {
		'cluster': tup[0],
		'anon': anon,
		'lat': tup[1].lat.mean(),
		'lon': tup[1].lon.mean(),
		'year': year,
		'month': month,
		"stops" : tup[1].id.tolist()
		}
	return row


def make_clust_dict(stops, n_firms, n_veh, year, month):
	clusters = [make_row(c, n_firms, n_veh, year, month) for c in stops.groupby('cluster')]
	clusters = dict(CLUSTERS = clusters)
	return clusters

def write_cluster(clusters, creds, year, month, n_firms = 2, n_veh = 3):
	
	g = neo4j.GraphDatabase.driver("bolt://%s" % creds['ipport'], auth = (creds['username'], creds['password']))
	
	query = "UNWIND $CLUSTERS as clust\
       CREATE (c:Cluster{id: clust.cluster}) SET c.anon = toBoolean(clust.anon), c.lat = clust.lat, c.lon = clust.lon, c.year = clust.year, c.month = clust.month\
       	FOREACH (ignoreMe IN CASE WHEN clust.loc_id <> '' THEN [1] ELSE [] END |\
        	MERGE (loc_id:Location{id: clust.loc_id})\
        	MERGE (c)-[:USED]->(loc_id)\
        )\
        FOREACH (ignoreMe IN CASE WHEN clust.addr <> '' THEN [1] ELSE [] END |\
        	MERGE (addr:Address{id: clust.addr})\
			MERGE (c)-[:AT]->(addr)\
       )\
       \
       WITH c, clust\
       UNWIND clust.stops as s_id\
       MATCH (s:Stop{id: s_id})\
       MERGE (s)-[:PART_OF]->(c)\
      "
	with g.session() as session:
		session.run(query, clusters)
	return clusters


##Consider different parameteres for loading zones and rest areas. This can be applied afterwards since lz threhold will likely always be less than ra threshold
'''
There is a bewildering error here I can't work out. Queensland stops will not be matched to Queensland
clusters unless all non-Queensland stuff is removed, even though this should be irrelevant. For now
I have the hacky solution of redoing it with Qld alone, which works but ???????
'''

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




##This duplicates much of the above but I can't merge them easily because of the QLD rest area bug thing

def addr_match(df, creds, locs, max_dist = 50):
	df = ckdnearest(df, locs, 'loc_id', max_dist)
	return df

def get_addr(creds):
	g = neo4j.GraphDatabase.driver("bolt://%s" % creds['ipport'], auth = (creds['username'], creds['password']))
	with g.session() as session:
		locs = session.read_transaction(addr_query)
	return locs

def addr_query(tx):
	query = "MATCH (l:Address) RETURN l.id as loc_id, l.lat as lat, l.lon as lon"
	records = [record for record in tx.run(query)]
	df = pd.DataFrame([r.values() for r in records], columns = records[0].keys())
	return df




def write_lz_rest(stops_loc, creds, node_type):
	db = 'bolt://%s' % creds['ipport']
	col = 'cluster' if node_type == 'Cluster' else 'id'
	query = 		'UNWIND $STOPS as s\
					 MERGE(c:%s{id: s.%s})\
					 MERGE(l:Location{id:s.loc_id})\
					 CREATE(c)-[:USED]->(l)' % (node_type, col)

	stops_loc = [dict(c[1]) for c in stops_loc.iterrows()]
	stops_loc = dict(STOPS = stops_loc)
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	with g.session() as session:
		session.run(query, stops_loc)



def write_addr(stops_addr, creds, node_type):
	db = 'bolt://%s' % creds['ipport']
	col = 'cluster' if node_type == 'Cluster' else 'id'
	query =         'UNWIND $STOPS as s\
					 MERGE(c:%s{id: s.%s})\
					 MERGE(l:Address{id:s.loc_id})\
					 CREATE(c)-[:AT]->(l)' % (node_type, col)
	##I can groupby address for minor performance improvement
	stops_addr = [dict(c[1]) for c in stops_addr.iterrows()]
	stops_addr = dict(STOPS = stops_addr)
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	with g.session() as session:
		session.run(query, stops_addr)



def ckdnearest(gdA, gdB, bcol, max_dist = 100, clust_match = False):   
    nA = np.array(list(zip(gdA.lon, gdA.lat)) )
    nB = np.array(list(zip(gdB.lon, gdB.lat)) )
    btree = cKDTree(nB)
    dist, idx = btree.query(nA, k=1, n_jobs = -1) ##this is fairly fast but can use n_jobs to parallelise. -1 is all cores
    dist = dist * 110250 ## turning euclidean distance into approxiamte metres
    df = pd.DataFrame.from_dict({'distance': dist.astype(int),
                             bcol : gdB.loc[idx, 'loc_id'].values })
    df = pd.concat([gdA, df], axis = 1)
    if clust_match:
        df[bcol] = df.apply(lambda x: x[bcol] if x.distance <= max_dist else "", axis = 1)
    else:
        df = df[df.distance <= max_dist] ##this could be done with distance_upper_bound = 1/110250 *100 in btree.query()
	
    return df

def match_clusters(clusters, locs, addr_locs, max_dist, max_dist_addr):
	df = pd.DataFrame.from_records(clusters["CLUSTERS"])
	df['loc_id'] = ckdnearest(df, locs, 'loc_id', max_dist, clust_match=True).loc_id
	df['addr'] = ckdnearest(df, addr_locs, 'addr', max_dist_addr, clust_match = True).addr
	lst = df.apply(lambda x: dict(x), axis = 1)
	return {"CLUSTERS" : lst.tolist()}






if __name__ == "__main__":

	now = datetime.now()

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--creds", type = str, default = "../Graphupload/neo4jcredsWIN.yaml",
		help="credential yaml for database")
	parser.add_argument("-y", "--year", type = int, default = now.year,
		help="year on which to cluster paired with month")
	parser.add_argument("-m", "--month", type = int, default = now.month - 1,
						help="month on which to cluster, paired with year")
	parser.add_argument("-nf", "--n_firms", type = int, default = 2,
						help="mininum number of firms for anonymity")
	parser.add_argument("-nv", "--n_vehicles", type = int, default = 3,
						help="mininum number of vehicles for anonymity")
	parser.add_argument("-e", "--epsilon", type = int, default = 1000,
						help="Episilon (in metres I think)")
	parser.add_argument("-ms", "--min_stops", type = int, default = 5,
						help="Min cluster size")
	parser.add_argument("-md", "--max_dist", type = float, default = 400,
						help="Max distance to be associated with a site")
	parser.add_argument("-mda", "--max_dist_addr", type = float, default = 400,
						help="Max distance to be associated with an address")
	parser.add_argument("-mis", "--match_ind_stops", type = bool, default = False,
						help="Match individual stops to locations - should be done by yuloserver now")
	args = parser.parse_args()

	if args.month == 0:
		args.year, args.month = args.year - 1, 12





	with open(args.creds, 'r') as credsfile:
		creds = yaml.safe_load(credsfile)

	print("Pulling stops for %d-%d" % (args.year, args.month))
	stops = pull_stops(creds, args.year, args.month, args.match_ind_stops)
	addr_locs = get_addr(creds = creds)
	locs = get_ras_lz(creds = creds)

	if args.match_ind_stops:
		print('matching %s stops to rest areas and loading zones' % len(stops.index))
		
		stops_loc = lz_rest_match(stops, locs = locs, max_dist = args.max_dist, creds = creds)
		print('Writing edges between %s locations and %s stops' % (len(set(stops_loc.loc_id)), len(stops_loc.index)))
		
		write_lz_rest_s =partial(write_lz_rest, creds = creds, node_type =  "Stop")
		write_lz_rest_s(stops_loc)


		print('matching %s stops to addresses' % len(stops.index))
		

		stops_addr = addr_match(stops, locs = addr_locs, max_dist = args.max_dist_addr, creds = creds)
		print('Writing edges between %s addresses and %s stops' % (len(set(stops_addr.loc_id)), len(stops_addr.index)))
		
		write_addr_s =partial(write_addr, creds = creds, node_type =  "Stop")
		
		write_addr_s(stops_addr)

	 

	print("Clustering %s stops" % len(stops.index))
	stops = cluster_stops(stops, args.year, args.month, eps = args.epsilon, min_stops = args.min_stops)
	stops = stops[stops.inCluster]
	clusters = make_clust_dict(stops, args.n_firms, args.n_vehicles, args.year, args.month)
	clusters = match_clusters(clusters, locs, addr_locs, args.max_dist, args.max_dist_addr)
	print("Writing %s stop clusters" % len(set(stops.cluster)))
	write_cluster1 = partial(write_cluster,
							creds = creds,
							year = args.year,
	 						month = args.month,
							n_firms = args.n_firms,
							n_veh = args.n_vehicles
							)
	
	clusters = write_cluster1(clusters)

	

