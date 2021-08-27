import neo4j
import yaml
from functools import partial
from collections import deque
import argparse
import os
from multiprocessing import Pool
'''
This assumes addresses.py has been run and the relevant files moved to the neo4j upload folder
'''

def query_serve(file, db, creds):
	print(file)

	q = "USING PERIODIC COMMIT 1000\
		 LOAD CSV WITH HEADERS FROM 'file:///%s' AS row\
		 MERGE(a:Address{id:row.ADDRESS_SITE_PID})\
		 SET a.lat = toFloat(row.LATITUDE)\
		 SET a.lon = toFloat(row.LONGITUDE)\
		 SET a.number = toInteger(row.NUMBER_FIRST)\
		 SET a.street = row.STREET_NAME + ' ' + row.STREET_TYPE_CODE\
		 SET a.locality = row.LOCALITY_NAME\
		 SET a.state = row.state\
		 "
	q = q % file

	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	with g.session() as session:
		session.run("CREATE CONSTRAINT ON (a:Address) ASSERT a.id IS UNIQUE;")
		session.run(q)
	print('Finished ' + file)



if __name__ == '__main__':

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--credspath", type = str, default = '../Graphupload/neo4jcredsWIN.yaml',
                        help="database credentials and address")
	args = parser.parse_args()

	files = os.listdir('addr/csvs/')

	
	with open(args.credspath, 'r') as credsfile:
		creds = yaml.safe_load(credsfile)

	db = 'bolt://%s' % creds['ipport']



	query_serve1 = partial(query_serve, db = db, creds = creds)
	with Pool() as p:
		p.map(query_serve1, files)
