import zipfile
import argparse
import requests
import os
import subprocess



if __name__ == '__main__':

	


	parser = argparse.ArgumentParser()
	parser.add_argument("-r", "--roadfile", type = str, default = "dataroads",
		help="filename of geojson in shapefile directory")
	parser.add_argument("-u", "--username", type = str, default = "geowonk",
		help="mapbox username")
	parser.add_argument("-t", "--token", type = str, default = 'default',
		help="mapbox secret token")
	parser.add_argument("-ts", "--tileset", type = str, default = 'roadsdata-ak289c',
		help="tileset name")

	args = parser.parse_args()


	files = os.listdir('../shapefiles/')
	files = ['../shapefiles/' + f for f in files if f.startswith(args.roadfile) and "geojson" not in f]
	ul_zip = '../shapefiles/dataroads.zip'
	if os.path.isfile(ul_zip):
		os.remove(ul_zip)

	with zipfile.ZipFile(ul_zip, 'w', zipfile.ZIP_DEFLATED) as zfile:
		for file in files:
			zfile.write(file)


	if args.token != 'default':
		token = args.token 
	else:
		with open('../Shinyleaflet/mapboxtokensecret.txt', 'r') as f:
			token = f.read().splitlines()[0] 

	credsurl = 'https://api.mapbox.com/uploads/v1/%s/credentials?access_token=%s' % (args.username, token)
	creds = requests.get(credsurl).json()
	#print(creds)

	# os.system('export AWS_ACCESS_KEY_ID=' + creds['accessKeyId'])
	# os.system('export AWS_SECRET_ACCESS_KEY=' + creds['secretAccessKey'])
	# os.system('export AWS_SESSION_TOKEN=' + creds['sessionToken'])

	os.environ['AWS_ACCESS_KEY_ID'] = creds['accessKeyId']
	os.environ['AWS_SECRET_ACCESS_KEY'] = creds['secretAccessKey']
	os.environ['AWS_SESSION_TOKEN'] = creds['sessionToken']

	os.system('aws s3 cp %s s3://%s/%s' % ('../shapefiles/dataroads.zip' , creds['bucket'], creds['key']))

	upload_command = 'curl -X POST -H "Content-Type: application/json" -H "Cache-Control: no-cache" -d\'{ "url": "http://%s.s3.amazonaws.com/%s", "tileset": "%s.%s"}\' "https://api.mapbox.com/uploads/v1/%s?access_token=%s"' % (creds['bucket'], creds['key'], args.username, args.tileset, args.username, token)

	#print(upload_command)

	os.system(upload_command)
