import urllib.request
import pandas as pd
import numpy as np
import zipfile
import neo4j
from multiprocessing import Pool
from functools import partial
import yaml
import argparse
import os
from shutil import rmtree
from collections import deque

'''The upload is really slow for reasons I don't understand so this will be turned into a scripts taht creats a CSV for READ CSV
	Having done this I suspect it is because I merged without asserting unique ids'''


# def write_addr(row, creds, state):
# 	db = 'bolt://%s' % creds['ipport']
# 	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
# 	ra_upload = "MERGE(a:Address{id:$ADDRESS_SITE_PID})\
# 				 SET a.lat = toFloat($LATITUDE)\
# 				 SET a.lon = toFloat(LONGITUDE)\
# 				 SET a.number = toInteger($NUMBER_FIRST)\
# 				 SET a.street = $STREET_NAME + ' ' + $STREET_TYPE_CODE\
# 				 SET a.locality = $LOCALITY_NAME\
# 				 SET a.state = '%s'" % state
# 	with g.session() as session:
# 		session.run(ra_upload, row)


def process_state(state, path):
	print("Processing addresses from " + state)
	site_gc = pd.read_csv(path + state + '_ADDRESS_SITE_GEOCODE_psv.psv', sep = '|', low_memory = False, dtype = str)[['LONGITUDE', 'LATITUDE', 'ADDRESS_SITE_PID', 'GEOCODE_TYPE_CODE']]
	add_det = pd.read_csv(path + state + '_ADDRESS_DETAIL_psv.psv', sep = '|', low_memory = False, dtype = str)[['NUMBER_FIRST', 'STREET_LOCALITY_PID', 'LOCALITY_PID', 'POSTCODE', 'ADDRESS_SITE_PID']].drop_duplicates()
	str_loc = pd.read_csv(path + state + '_STREET_LOCALITY_psv.psv', sep = '|', low_memory = False, dtype = str)[['STREET_LOCALITY_PID', 'STREET_NAME', 'STREET_TYPE_CODE']]
	loc_loc = pd.read_csv(path + state + '_LOCALITY_psv.psv', sep = '|', low_memory = False, dtype = str)[['LOCALITY_PID', 'LOCALITY_NAME']]

	##Filter site_gc by geocode type to get unique row - or take average, or just take frontage?


	addr = pd.merge(site_gc, add_det, on = 'ADDRESS_SITE_PID', how = 'left')
	addr = pd.merge(addr, str_loc, on = 'STREET_LOCALITY_PID', how = 'left')
	addr = pd.merge(addr, loc_loc, on = 'LOCALITY_PID', how = 'left')
	addr['state'] = state
	#addr.NUMBER_FIRST = addr.NUMBER_FIRST.apply(lambda x: int(x))
	addr = addr.dropna(subset = ['ADDRESS_SITE_PID'])


	if not os.path.exists('addr/csvs'): os.makedirs('addr/csvs')

	addr.to_csv('addr/csvs/%s.csv' % state)

	# addr = [dict(r[1]) for r in addr.iterrows()]

	# write_addr1 = partial(write_addr, creds = creds, state = state)
	
	# with Pool() as p:
	# 	p.map(write_addr1, addr)


if __name__ == "__main__":

	
	parser = argparse.ArgumentParser()
	# parser.add_argument("-c", "--creds", type = str, default = "../Graphupload/neo4jcredsWIN.yaml",
	# 	help="credential yaml for database")
	parser.add_argument("-u", "--url", type = str,
	 default = 'https://data.gov.au/data/dataset/19432f89-dc3a-4ef3-b943-5326ef1dbecc/resource/4b084096-65e4-4c8e-abbe-5e54ff85f42f/download/may20_gnaf_pipeseparatedvalue.zip',
		help="credential yaml for database")
	
	args = parser.parse_args()

	# with open(args.creds, 'r') as credsfile:
	# 	creds = yaml.load(credsfile)

	
	rmtree('addr')
	os.mkdir('addr')
	
	urllib.request.urlretrieve(args.url, 'addr/PSMA.zip')
	zipfile.ZipFile('addr/PSMA.zip', 'r').extractall('addr')

	dir1 = [x for x in os.listdir('addr') if x != 'PSMA.zip'][0]
	dir2 = [x for x in os.listdir('addr/%s/G-NAF/' % dir1) if x.startswith('G-NAF')][0]

	path ='addr/%s/G-NAF/%s/Standard/' % (dir1, dir2)
	
	states = [f.split('_')[0] for f in os.listdir(path)]
	states = list(set(states))
	#states = [s for s in states if s not in ['VIC', 'SA', 'OT', 'ACT']]
	process_state1 = partial(process_state, path = path)
	
	deque(map(process_state1, states), maxlen = 0)