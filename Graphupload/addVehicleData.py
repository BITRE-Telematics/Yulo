import neo4j
import yaml
import pandas as pd 
import argparse
from multiprocessing import Pool
from functools import partial
import numpy as np
import sys
import os
from re import search

def writer(row, creds):
		db = 'bolt://%s' % creds['ipport']
		print('adding data for %s' % row['Vehicle'])
		veh_upload = "MATCH(v:Asset{id:$Vehicle}) SET v.firm = $Firm, v.type = $asset_type"
		g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
		with g.session() as session:
			session.run(veh_upload, row)

if __name__ == '__main__':

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--creds", type = str, default = "neo4jcredsWIN.yaml",
                        help="credential yaml for database")
	parser.add_argument("-f", "--file", type = str, default = None,
                        help="File with vehicle data. Accepts regex to process all with a pattern")
	parser.add_argument("-d", "--datadir", type = str, default = '../data/',
                        help="directory of data")
	args = parser.parse_args()
	
	if args.file is None:
		"Specify data input please"
		sys.exit()
	
	files = [f for f in os.listdir(args.datadir) if search(args.file, f)]

	for f in files:
		print("Uploading vehicles from file %s" % f)
		data = pd.read_csv(args.datadir + f).drop_duplicates(['Vehicle'])
		if 'asset_type' not in data : data['asset_type'] = 'Unknown'
		if 'Firm' not in data : data['Firm'] = 'NA'
		
		data = data.iterrows()
		data = [dict(v[1]) for v in data]
		


		with open(args.creds, 'r') as credsfile:
			creds = yaml.safe_load(credsfile)

		writer1 = partial(writer, creds = creds)
		with Pool() as p:
			p.map(writer1, data)

		g = neo4j.GraphDatabase.driver('bolt://%s' % creds['ipport'], auth = (creds['username'], creds['password']))
		## This changes 'Vehicles" that are actually trailers to Trailers'
		with g.session() as session:
			session.run("MATCH(a:Asset) WHERE a.type  = 'Trailer' SET a:Trailer")
			session.run("MATCH(a:Asset) WHERE a.type <> 'Trailer' SET a:Vehicle")





