import urllib
import pandas as pd
import numpy as np
import zipfile
import neo4j
from multiprocessing import Pool
from functools import partial
import yaml
import argparse
import geopandas as gpd
from os import system
import requests
from io import StringIO
#from bs4 import BeautifulSoup

url_nt  = 'https://nt.gov.au/driving/safety/road-rest-stops-in-nt'






def write_ra(row, creds):
	db = 'bolt://%s' % creds['ipport']
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	ra_upload = "MERGE(ra:Location{id:$id})\
					 SET ra:Rest_area\
					 SET ra.name = $rest_area_name\
					 SET ra.number = toInteger($rest_area_number)\
					 SET ra.lon = toFloat($longitude)\
					 SET ra.lat = toFloat($latitude)\
					 SET ra.lga = $lga\
					 SET ra.segregated = toBoolean($separated_parking_areas_light_heavy)\
					 SET ra.h_vehicle_capacity = $heavy_vehicle_parking_capacity\
					 SET ra.location_desc =  $distance_from\
					 SET ra.road_name = $road_name\
					 SET ra.side_of_road = $side_of_road\
					 SET ra.carriageway_configuration = $carriageway_configuration\
					 SET ra.vehicle_type = $vehicle_type\
					 SET ra.power_supply = toBoolean($power_supply)\
					 SET ra.fuel_available = toBoolean($fuel_available)\
					 SET ra.food_available = toBoolean($food_available)\
					 SET ra.shade = toBoolean($shade)\
					 SET ra.shelter = toBoolean($shelter)\
					 SET ra.tree_shading = toBoolean($tree_shading)\
					 SET ra.litter_bins = toBoolean($litter_bins)\
					 SET ra.water_supply = $water_supply\
					 SET ra.picnic_table = toBoolean($picnic_table)\
					 SET ra.toilets = toInteger($toilets)\
					 SET ra.surface_type = $surface_type\
					 SET ra.remarks = $remarks"
	with g.session() as session:
		session.run(ra_upload, row)


def write_lz(row, creds):
	db = 'bolt://%s' % creds['ipport']
	g = neo4j.GraphDatabase.driver(db, auth = (creds['username'], creds['password']))
	with g.session() as session:
		session.run("MERGE(ra:Location{id:$FEATURE_ID})\
					 SET ra:Loading_zone\
					 SET ra.name = $NAME\
					 SET ra.spaces = toInteger($SPACES_AVAILABLE)\
					 SET ra.lon = toFloat($POINT_X)\
					 SET ra.lat = toFloat($POINT_Y)\
					 SET ra.side = $SIDE\
					 SET ra.wd_hrs = $HRS_OPERATION_WEEKDAY\
					 SET ra.type = $TYPE\
					 ",
					  row)


def add_na(df, cols):
	for col in cols:
		if col not in df.columns:
			df[col] = "NA"
			print('Filling column %s with NA' % col)
	return df.fillna('NA')

def sum_roads(tup):
	df = pd.DataFrame({
		'ROAD' : tup[0],
		'NET_DIST' : [tup[1].END_TRUE_DIST.max() - tup[1].START_TRUE_DIST.min()],
		'MAX_DIST' : [tup[1].END_TRUE_DIST.max()],
		'MIN_DIST' : [tup[1].START_TRUE_DIST.min()],
		'NET_SLK' : [tup[1].END_SLK.max() - tup[1].START_SLK.min()],
		'MAX_SLK' : [tup[1].END_SLK.max()],
		'MIN_SLK' : [tup[1].START_SLK.min()]
		})
	return df

def parse_nt_html(url):
	resp_nt = requests.get(url).text
	soup = BeautifulSoup(resp_nt)
	nt_table = soup.find_all('table')
	names = []
	att = []
	for table in nt_table:
		t = table.find_all('td')
		for i in range(0, int(len(t)/3)):
			first = i * 3
			names.append(t[first].get_text(' '))
			att.append(t[first + 2].get_text(' '))
	df = pd.DataFrame({
		'name': names,
		'att': att
		})
	return df


if __name__ == "__main__":
	

	parser = argparse.ArgumentParser()
	parser.add_argument("-c", "--creds", type = str, default = "../Graphupload/neo4jcredsWIN.yaml",
		help="credential yaml for database")
	
	args = parser.parse_args()

	with open(args.creds, 'r') as credsfile:
		creds = yaml.load(credsfile)


	write_ra1 = partial(write_ra, creds = creds)

	##NSW
	with open("nsw/TfNSWapikey.txt", 'r') as f:
		key = f.read()
	req_ra = urllib.request.Request('https://api.transport.nsw.gov.au/v1/roads/spatial?format=csv&q=select%20*%20from%20rest_areas%20')
	req_ra.add_header('Authorization', 'apikey ' + key)
	resp_ra = urllib.request.urlopen(req_ra)
	data_ra = resp_ra.read()

	d = str(data_ra).split('\\n')

	with open('nsw/rest_areas_nsw.csv', 'w') as f:
		for a in d:
			f.writelines(a + '\n')


	ra_nsw = pd.read_csv('nsw/rest_areas_nsw.csv').\
		rename(columns = {
			'water_supply_type': 'water_supply'
			})
	ra_nsw['id'] = ra_nsw.apply(lambda x: 'NSW_ra%s' % int(x.rest_area_number), axis = 1)

	cols = [c for c in ra_nsw.columns] + ['surface_type', 'remarks']

	ra_nsw = add_na(ra_nsw, cols)

	ra_nsw = [dict(dict(r[1])) for r in ra_nsw.iterrows()]
	
	with Pool() as p:
		p.map(write_ra1, ra_nsw)

	



	##WA - hashed out whilst they fix their files


	

	ra_wa = pd.read_csv('http://portal-mainroads.opendata.arcgis.com/datasets/79232357944c4bd6a593d19e0fbdcc77_19.csv').\
		rename(columns={
			'REST_AREA_NAME': 'rest_area_name',
			'LG_NAME': 'lga',
			'CONSTRUCTED_SHELTER' : 'shelter',
			'NATURAL_SHADE': 'tree_shading',
			'COMMON_USAGE_NAME': 'road_name',
			'START_SLK': "SLK",
			'START_TRUE_DIST': 'DIST',
			'NUMBER_OF_TOILETS': 'toilets',
			'X': 'longitude',
			'Y': 'latitude',
			'REST_AREA_TYPE': 'vehicle_type'
			})

	

	ra_wa['id']           = ra_wa.OBJECTID.apply(lambda x: 'WA_ra%s' % x)

	ra_wa.shelter         = ra_wa.shelter.apply(lambda x: x == "Yes")
	ra_wa['tree_shading'] = ra_wa.tree_shading.apply(lambda x: x == "Yes")
	ra_wa['litter_bins']  = ra_wa.NUMBER_OF_BINS.apply(lambda x: x >0 )
	ra_wa['picnic_table'] = ra_wa.NUMBER_OF_TABLES.apply(lambda x: x >0 )

	ra_wa['surface']      = ra_wa.apply(lambda x: x.SURFACE if x.SURFACE == "Unsurfaced" else x.SURFACE_TYPE, axis = 1)


	# #Alt WA

	# ra_wa = pd.read_csv('https://www.mainroads.wa.gov.au/Documents/Rest%20Area%20Guide%20March%202015%20-%20GPS%20Data.RCN-D15%5E23120930.CSV', header = None, names = ['longitude', 'latitude', 'att'])
	# ra_wa['name'] = ra_wa.att.apply(lambda x: x.split(':')[-1].strip())
	# ra_wa['road_name'] = ra_wa.att.apply(lambda x: x.split('-')[0].strip())
	# ra_wa['remarks'] = ra_wa.att.apply(lambda x: x.split('-')[1].strip().split(':')[0])
	# ra_wa['toilets'] = ra_wa.att.apply(lambda x: 1 if 'Toilet - Yes' in x else 0)
	# ra_wa['litter_bins'] = ra_wa.att.apply(lambda x: 'Bin - Yes' in x)
	# ra_wa['shelter'] = ra_wa.att.apply(lambda x: 'Shelter - Yes' in x)
	# ra_wa['picnic_table'] = ra_wa.att.apply(lambda x: 'Table - Yes' in x)

	
	ra_wa = add_na(ra_wa, cols)

	ra_wa = [dict(r[1]) for r in ra_wa.iterrows()]
	
	with Pool() as p:
		p.map(write_ra1, ra_wa)

	##SA
	url = urllib.request.urlretrieve('http://www.dptiapps.com.au/dataportal/StateMaintainedRestAreas_geojson.zip', 'sa/rest_areas_sa.zip')
	zipfile.ZipFile('sa/rest_areas_sa.zip', 'r').extractall('sa/')
	ra_sa = gpd.read_file('sa/StateMaintainedRestAreas_GDA2020.geojson')

	ra_sa = ra_sa.\
		rename(columns={
			'RESTAREASURFACE_DESC': 'surface_type',
			'REST_AREA_NAME': 'rest_area_name'
			})


	ra_sa['latitude']     = ra_sa.geometry.centroid.y
	ra_sa['longitude']    = ra_sa.geometry.centroid.x
	ra_sa['id']           = ra_sa.OBJECTID_1.apply(lambda x: 'SA_ra%s' % int(x))
	#ra_sa['toilets'] = ra_sa.TOILETBLOCKS.apply(lambda x: x >0 )
	ra_sa['shelter']      = ra_sa.SHELTERS.apply(lambda x: x >0 )
	ra_sa['litter_bins']  = ra_sa.BINS.apply(lambda x: x >0 )
	ra_sa['picnic_table'] = ra_sa.TABLES.apply(lambda x: x >0 )
	ra_sa = pd.DataFrame(ra_sa).drop('geometry', axis =1)
	
	ra_sa = add_na(ra_sa, cols)

	ra_sa = [dict(r[1]) for r in ra_sa.iterrows()]
	
	with Pool() as p:
		p.map(write_ra1, ra_sa)

	##QLD
	ra_qld = pd.read_csv('qld/RoadsideAmenities.csv').\
		rename(columns={
			'NAME': 'rest_area_name',
			'LG': 'lga',
			'NAME': 'rest_area_name',
			'LOCATION': 'distance_from',
			'REMARKS': 'remarks',
			'Latitude': 'latitude',
			'Longitude': 'longitude'

			})

	ra_qld['id']           = ra_qld.NUMBER_.apply(lambda x: 'QLD_ra%s' % x)
	ra_qld['picnic_table'] = ra_qld.TABLE_.apply(lambda x: x == 'x')
	ra_qld['water_supply'] = ra_qld.WSUPPLY.apply(lambda x: x == 'x')
	ra_qld['toilets']      = ra_qld.apply(lambda x: 1 if x.WCLOSET == 'x' or x.ECLOSET == 'x' or x.Enviro_Toilet == 'x' or x.Disabled_T == 'x' else 0, axis = 1)
	ra_qld['shelter']      = ra_qld.SSHED.apply(lambda x: x == 'x')

	ra_qld = add_na(ra_qld, cols)

	ra_qld = [dict(r[1]) for r in ra_qld.iterrows()]
	
	with Pool() as p:
		p.map(write_ra1, ra_qld)

	##Vic

	'''This data isn't published properly but the csv underlying the interactive map is accessible through browsers, hence the spoofing here
	'''

	url = "https://www.vicroads.vic.gov.au/~/media/Media/Business and Industry/Rest Areas/restareadata"
	headers = {"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:66.0) Gecko/20100101 Firefox/66.0"}
	req = requests.get(url, headers=headers)
	data = StringIO(req.text)
	ra_vic = pd.read_csv(data).\
		rename(columns = {
    		'RestAreaName': 'rest_area_name',
    		'Latitude': 'latitude',
			'Longitude': 'longitude',
			'Carriageway': 'side_of_road',
			'SurfaceType': 'surface_type',
			'RestAreaType': 'vehicle_type',
			'RoadName': 'road_name',
			'water': 'water_supply'
			})
	ra_vic['id']           = ra_vic.RestAreaID.apply(lambda x: 'VIC_ra%s' % x)
	ra_vic['picnic_table'] = ra_vic.PicnicTables.apply(lambda x: x == 'YES')
	ra_vic['toilets']      = ra_vic.Toilets.apply(lambda x: 1 if x == 'YES' else 0)
	ra_vic['litter_bins']  = ra_vic.RubbishBins.apply(lambda x: x == 'YES')
	

	ra_vic['location_desc'] = ra_vic.apply(lambda x: str(x.DistanceFromInt) + ' km ' + str(x.DirectionFromInt) + ' ' +  str(x.NearestIntersection) + ', ' +  x.Locality ,axis = 1)

	ra_vic = add_na(ra_vic, cols)

	ra_vic = [dict(r[1]) for r in ra_vic.iterrows()]
	with Pool() as p:
		p.map(write_ra1, ra_vic)

	##Tas - file sent by email
	
	#system('ogr2ogr -f Geojson tas/tas_rest_areas.geojson tas/Roadside_Stops_State__Growth.shp')
	ra_tas = gpd.read_file('tas/tas_rest_areas.geojson').\
		rename(columns = {
			"COMMENTS": "surface_type",
			"SITE_NAME": "rest_area_name",
			"TYPE": "vehicle_type",
			"LOCATION": "location_desc",
			"ROAD_NAME": "road_name",
			"DIRECTION": "side_of_road"
			})

	ra_tas['id'] = ra_tas.OBJECTID.apply(lambda x: 'TAS_ra%s' % x)

	##Tasmanian data is in a specifically Australian cartesian projection
	ra_tas.crs = {'init': 'epsg:28355'}
	ra_tas = ra_tas.to_crs({'init': 'epsg:4326'})

	ra_tas['latitude']     = ra_tas.geometry.apply(lambda x: x.y)
	ra_tas['longitude']    = ra_tas.geometry.apply(lambda x: x.x)
	ra_tas['toilets']      = ra_tas.TOILET.apply(lambda x: 1 if x == 'Y' else 0)
	ra_tas['shelter']      = ra_tas.SHELTER.apply(lambda x: x == 'Y')
	ra_tas['tree_shading'] = ra_tas.SHELTER.apply(lambda x: x == 'Shade')
	ra_tas['litter_bins']  = ra_tas.RUBBISH_BI.apply(lambda x: x == 'Y')
	ra_tas['picnic_table'] = ra_tas.PICNIC_ARE.apply(lambda x: x == 'Y')

	ra_tas = pd.DataFrame(ra_tas).drop('geometry', axis =1)
	ra_tas = add_na(ra_tas, cols)

	ra_tas = [dict(r[1]) for r in ra_tas.iterrows()]
	
	with Pool() as p:
		p.map(write_ra1, ra_tas)

	#NT - a combination of scraping and manual work, now just precooked

	

	ra_nt = pd.read_csv('nt/nt.csv')

	ra_nt = add_na(ra_nt, cols)

	ra_nt = [dict(r[1]) for r in ra_nt.iterrows()]
	with Pool() as p:
		p.map(write_ra1, ra_nt)



	ra_act = pd.read_csv('act/act.csv')

	ra_act = add_na(ra_act, cols)

	ra_act = [dict(r[1]) for r in ra_act.iterrows()]
	with Pool() as p:
		p.map(write_ra1, ra_act)


	##loading zones
	##NSW
	req_lz = urllib.request.Request('https://api.transport.nsw.gov.au/v1/roads/static/loadingzones')
	req_lz.add_header('Authorization', 'apikey ' + key)
	resp_lz = urllib.request.urlopen(req_lz)
	data_lz = resp_lz.read()
	with open('nsw/loading_zones.zip', 'wb') as zfile:
		zfile.write(data_lz)
	with zipfile.ZipFile('nsw/loading_zones.zip', 'r') as z:
		z.extractall('nsw/')

	
	lz_nsw  = pd.read_csv('nsw/LoadingzoneData.csv')
	lz_nsw['FEATURE_ID'] = lz_nsw.apply(lambda x: 'NSW_lz%s' % x.Feature_ID, axis = 1)
	lz_nsw  = lz_nsw.drop([c for c in lz_nsw.columns if  c.startswith('User')], axis = 1)
	lz_nsw['NAME'] = lz_nsw.apply(lambda x: x.STREET + ' ' + x.SUFFIX + ' ' +  x.SIDE + ' between ' +  x.BETWEEN_ ,axis = 1)
	cols_lz = [c for c in lz_nsw.columns] + ["TYPE"]
	lz_nsw  = add_na(lz_nsw, cols_lz)

	lz_nsw = [dict(r[1]) for r in lz_nsw.iterrows()]
	write_lz1 = partial(write_lz, creds = creds)
	with Pool() as p:
		p.map(write_lz1, lz_nsw)

	##Qld
	lz_qld = pd.read_csv('https://www.data.brisbane.qld.gov.au/data/dataset/b64bcb00-7d0a-4279-98f2-b96661ff11e7/resource/34a40086-744c-40f6-905c-5a990ed582f8/download/commercial-loading-zones.csv').\
		rename(columns={
			'LONGITUDE': 'POINT_X',
			'LATITUDE': 'POINT_Y',
			'OPERATING TIMES': 'HRS_OPERATION_WEEKDAY',
			'ZONE ID': 'FEATURE_ID'
			})
	lz_qld.FEATURE_ID = lz_qld.apply(lambda x: 'QLD_lz%s' % x.FEATURE_ID, axis = 1)

	lz_qld = add_na(lz_qld, cols_lz)

	lz_qld = [dict(r[1]) for r in lz_qld.iterrows()]

	with Pool() as p:
		p.map(write_lz1, lz_qld)



	##Vic-Melb
	lz_vic_geom = gpd.read_file('https://data.melbourne.vic.gov.au/api/geospatial/crvt-b4kt?method=export&format=GeoJSON')
	lz_vic_geom['BayID'] = lz_vic_geom.bay_id.apply(lambda x: int(x))


	lz_vic = pd.read_csv('https://data.melbourne.vic.gov.au/api/views/ntht-5rk7/rows.csv?accessType=DOWNLOAD').\
		rename(columns = {
			'Description1' : 'HRS_OPERATION_WEEKDAY'
			})
	lz_vic = lz_vic[lz_vic.TypeDesc1.apply(lambda x: "Loading" in x)]
	lz_vic['FEATURE_ID'] = lz_vic.BayID.apply(lambda x: 'VIC_lz' + str(x))

	lz_vic = pd.merge(lz_vic, lz_vic_geom, how = 'left', on = 'BayID')
	lz_vic['cent']    = lz_vic.geometry.apply(lambda x: x.centroid)
	lz_vic['POINT_X'] = lz_vic.cent.apply(lambda x: x.x)
	lz_vic['POINT_Y'] = lz_vic.cent.apply(lambda x: x.y)

	lz_vic = add_na(lz_vic, cols_lz)

	lz_vic = lz_vic.drop(['cent', 'geometry'], axis = 1)

	lz_vic = [dict(r[1]) for r in lz_vic.iterrows()]

	with Pool() as p:
		p.map(write_lz1, lz_vic)


	##WA Perth

	##SA ADL

	url_sa_lz = 'http://opendata.adelaidecitycouncil.com/On_Street_Parking/parkingreport.csv'
	req_lz_sa = urllib.request.Request(url_sa_lz)
	resp_lz_sa = urllib.request.urlopen(req_lz_sa)
	data_lz = resp_lz_sa.readlines()
	headers = list(pd.read_csv(StringIO(str(data_lz[0]).replace(' ', '')[2:-5])).columns)
	ncol = len(headers)

	lz_sa = pd.DataFrame(columns = headers)


	for line in data_lz[2:]:
		l = pd.read_csv(StringIO(str(line).replace('  ', '')[2:-5]), header = None)
		if len(l.columns) == ncol + 1:
			l = l.drop(2, axis = 1) 
		l.columns = headers
		if 'Loading' in str(l['PrimeControl'][0]):
				lz_sa = lz_sa.append(l)

	lz_sa = lz_sa.rename(columns = {
		"Longitude": 'POINT_X',
		"Latitude": 'POINT_Y',
		'PrimeControl': 'HRS_OPERATION_WEEKDAY',
		'Street': 'NAME',
		'NumberofSpaces': 'SPACES_AVAILABLE'
	})

	lz_sa['FEATURE_ID'] = lz_sa.ZONEID.apply(lambda x: 'SA_lz' + str(x))
	
	lz_sa = add_na(lz_sa, cols_lz)

	lz_sa = [dict(r[1]) for r in lz_sa.iterrows()]

	with Pool() as p:
		p.map(write_lz1, lz_sa)


	##Qld Gold Coast 


	##NT scraping code 
	# url_nt  = 'https://nt.gov.au/driving/safety/road-rest-stops-in-nt'

	# ra_nt = parse_nt_html(url_nt)
	# ra_nt['id'] = ra_nt.name.apply(lambda x: 'NT_ra%s' % x)
	# ra_nt['toilets'] = ra_nt.att.apply(lambda x: 1 if 'toilet' in x else 0)
	# ra_nt['water_supply'] = ra_nt.att.apply(lambda x: 'water supply' in x)
	# ra_nt['shelter'] = ra_nt.att.apply(lambda x: 'shelter' in x)
	# ra_nt['picnic_table'] = ra_nt.att.apply(lambda x: 'picnic table' in x)


	# ##geocoding
	# locs_nt = ra_nt.name.apply(lambda x:  geocoder.google(geocoder.google('%s rest Northern Territory' % x))
	# nt = geocoder.google('Northern Territory').latlng
	# locs_nt = locs_nt.apply(lambda x: x.latlng)
	
	# ra_nt['longitude'] = [x[1] for x in locs]
	# ra_nt['latitude'] = [x[0] for x in locs]

	# precooked = pd.read_csv('nt/nt.csv')
	# precooked.index = precooked.name
	# precooked = precooked.to_dict('index')

	# ra_nt.longitude = ra_nt.apply(lambda x: precooked[x['name']]['longitude'] if x['name'] in precooked.keys() else x.longitude, axis = 1)
	# ra_nt.latitude = ra_nt.apply(lambda x: precooked[x['name']]['latitude'] if x['name'] in precooked.keys() else x.latitude, axis = 1)

	##Old WA code

	# ra_wa = pd.read_csv('http://portal-mainroads.opendata.arcgis.com/datasets/79232357944c4bd6a593d19e0fbdcc77_19.csv').\
	# 	rename(columns={
	# 		'REST_AREA_NAME': 'rest_area_name',
	# 		'LG_NAME': 'lga',
	# 		'CONSTRUCTED_SHELTER' : 'shelter',
	# 		'NATURAL_SHADE': 'tree_shading',
	# 		'COMMON_USAGE_NAME': 'road_name',
	# 		'START_SLK': "SLK",
	# 		'START_TRUE_DIST': 'DIST',
	# 		'NUMBER_OF_TOILETS': 'toilets',
	# 		'X': 'longitude',
	# 		'Y': 'latitude'
	# 		}).\
	# 	drop(['END_TRUE_DIST', 'END_SLK'], axis = 1)
	

	# wa_roads = gpd.read_file('http://portal-mainroads.opendata.arcgis.com/datasets/082e88d12c894956945ef5bcee0b39e2_17.geojson')
	# wa_roads = wa_roads.sort_values(['ROAD', 'START_SLK'])
	# wa_roads = wa_roads[wa_roads.ROAD.apply(lambda x: x in list(ra_wa.ROAD))]
	
	# with Pool() as p:
	# 	sum_wa_roads = p.map(sum_roads, wa_roads.groupby('ROAD'))
	# sum_wa_roads = pd.concat(sum_wa_roads)


	# wa_roads = wa_roads.dissolve(by = 'ROAD').drop(['END_SLK', 'START_SLK', 'END_TRUE_DIST', 'START_TRUE_DIST', 'OBJECTID'], axis = 1)
	# wa_roads['ROAD'] = wa_roads.index
	# wa_roads = pd.merge(wa_roads, sum_wa_roads, on = ['ROAD'])
	# ra_wa = pd.merge(wa_roads, ra_wa, on = 'ROAD')
	

	# ra_wa['propDIST'] = ra_wa.apply(lambda x: (x.SLK ) / x.MAX_SLK, axis = 1)

	# ra_wa['longitude'] = ra_wa.apply(lambda x: x.geometry.interpolate(x.propDIST, normalized = True).x, axis =1)
	# ra_wa['latitude'] = ra_wa.apply(lambda x: x.geometry.interpolate(x.propDIST, normalized = True).y, axis =1)

	# ra_wa['id'] = ra_wa.OBJECTID.apply(lambda x: 'WA_ra%s' % x)

	# ra_wa.shelter = ra_wa.shelter.apply(lambda x: x == "Yes")
	# ra_wa['tree_shading'] = ra_wa.tree_shading.apply(lambda x: x == "Yes")
	# ra_wa['litter_bins'] = ra_wa.NUMBER_OF_BINS.apply(lambda x: x >0 )
	# ra_wa['picnic_table'] = ra_wa.NUMBER_OF_TABLES.apply(lambda x: x >0 )

	# ra_wa['surface'] = ra_wa.apply(lambda x: x.SURFACE if x.SURFACE == "Unsurfaced" else x.SURFACE_TYPE, axis = 1)


	# ra_wa = pd.DataFrame(ra_wa).drop('geometry', axis =1)

	##Alt WA

	# ra_wa = pd.read_csv('https://www.mainroads.wa.gov.au/Documents/Rest%20Area%20Guide%20March%202015%20-%20GPS%20Data.RCN-D15%5E23120930.CSV', header = None, names = ['longitude', 'latitude', 'att'])
	# ra_wa['name'] = ra_wa.att.apply(lambda x: x.split(':')[-1].strip())
	# ra_wa['road_name'] = ra_wa.att.apply(lambda x: x.split('-')[0].strip())
	# ra_wa['remarks'] = ra_wa.att.apply(lambda x: x.split('-')[1].strip().split(':')[0])
	# ra_wa['toilets'] = ra_wa.att.apply(lambda x: 1 if 'Toilet - Yes' in x else 0)
	# ra_wa['litter_bins'] = ra_wa.att.apply(lambda x: 'Bin - Yes' in x)
	# ra_wa['shelter'] = ra_wa.att.apply(lambda x: 'Shelter - Yes' in x)
	# ra_wa['picnic_table'] = ra_wa.att.apply(lambda x: 'Table - Yes' in x)

	
	# ra_wa = add_na(ra_wa, cols)

	# ra_wa = [dict(r[1]) for r in ra_wa.iterrows()]
	
	# with Pool() as p:
	# 	p.map(write_ra1, ra_wa)
