import neo4j
import yaml
import pandas as pd 
import argparse
from multiprocessing import Pool
from functools import partial
import time


def add_missing(df, osm_id, bd):
	ranges = {
		'hour': list(range(0, 24)),
		'dayOfWeek': list(range(1, 8)),
		'month': list(range(1, 13))
	}
	present = list(df[bd])
	missing = [x for x in ranges[bd] if x not in present]
	miss_dict = pd.DataFrame({bd: missing, 'osm_id': osm_id})
	df = df.append(miss_dict).fillna(-1)
	if len(df.index)> len(ranges[bd]):
		df = df[(df[bd] != 0) | (df['UQ_imp'] !=0) ] #This is a failsafe because incorrect datetimes were returning nulls which led to a default 0 value in the go code
	return df

def create_ul_dict(df, osm_id, bd):
	df = df.sort_values(by = bd)
	ul_dict = {
		"osm_id"           : osm_id,
		"lq_imp%s" %bd     : list(df.LQ_imp),
		"median_imp%s" %bd : list(df.median_imp),
		"uq_imp%s" %bd     : list(df.UQ_imp),
		"n_trips%s" %bd    : list(df.n_trips),
		"n_vehicles%s" %bd : list(df.n_vehicles)
	}
	return ul_dict

##This is messy because the curly brackets don't work well with the better python string formatting options. The new cypher syntax should fix this
def upload(ul_dict):
	db = 'bolt://%s' % creds['ipport']
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	keys = [x for x in ul_dict.keys() if x != "osm_id"]
	q = "MATCH(s:Segment{osm_id:$osm_id}) SET"
	for i in keys:
		q = q + " s.%s = {%s}, " % (i, i)
	q = q + " s.updated = date()"
	with g.session() as session:
		session.run(q, ul_dict)



def writer(df, creds, bd):
	if isinstance(df, tuple):
		df = df[1]
	direction = "direction" in df.columns
	if direction:
		ul_dict = direction_writer(df, creds, bd)
	else:
		df = add_missing(df, df.osm_id.iloc[0], bd)
		ul_dict = create_ul_dict(df, df.osm_id.iloc[0], bd)
	print("Uploading %s" % ul_dict['osm_id'])
	upload(ul_dict)




def direction_writer(df, creds, bd):
	osm_id = df.osm_id.iloc[0]
	forward_df = add_missing(df[df.forward], osm_id, bd)
	backward_df = add_missing(df[~df.forward], osm_id, bd)
	forward_dict = create_ul_dict(forward_df, osm_id, bd)
	backward_dict = create_ul_dict(backward_df, osm_id, bd)

	forward  =  df.direction[df.forward].iloc[0]  if len(df.direction[df.forward])  > 0 else None
	backward =  df.direction[~df.forward].iloc[0] if len(df.direction[~df.forward]) > 0 else None

	lq_imp  = "lq_imp%s" % bd    
	med_imp = "median_imp%s" % bd 
	uq_imp  = "uq_imp%s"  %bd     
	n_trips = "n_trips%s" % bd    
	n_veh   = "n_vehicles%s" % bd 

	ul_dict = {
		"osm_id"           : df.osm_id.iloc[0],
		lq_imp + '_fw'     : list(forward_dict[lq_imp]),
		med_imp + '_fw'    : list(forward_dict[med_imp]),
		uq_imp + '_fw'     : list(forward_dict[uq_imp]),
		n_trips + '_fw'    : list(forward_dict[n_trips]),
		n_veh + '_fw'      : list(forward_dict[n_veh]),
		lq_imp + '_bw'     : list(backward_dict[lq_imp]),
		med_imp + '_bw'    : list(backward_dict[med_imp]),
		uq_imp + '_bw'     : list(backward_dict[uq_imp]),
		n_trips + '_bw'    : list(backward_dict[n_trips]),
		n_veh + '_bw'      : list(backward_dict[n_veh]),
		"forward"		   : forward,
		"backward"		   : backward
	}
	return ul_dict






if __name__ == "__main__":

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--creds", type = str, default = "../neo4jcredsWIN.yaml",
                        help="credential yaml for database")
	parser.add_argument("-f", "--file", type = str, default = "segspeeds_byhour.csv",
                        help="File with vehicle data. Accepts regex to process all with a pattern")
	
	args = parser.parse_args()

	with open(args.creds, 'r') as credsfile:
		creds = yaml.safe_load(credsfile)

	data = pd.read_csv(args.file, dtype = {"osm_id" : "str"}).drop_duplicates()


	if "hour" in data.columns:
		bd = "hour"
	elif "dayOfWeek" in data.columns:
		bd = "dayOfWeek"
	elif "month" in data.columns:
		bd = "month"

	data = data.groupby('osm_id')

	
	writer1 = partial(writer, creds = creds, bd = bd)


	with Pool() as p:
		p.map(writer1, data)