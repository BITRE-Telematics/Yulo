from geomet import wkt
import json
import pandas as pd

def dump_sa2(f):
	d = pd.DataFrame(
				{'SA2': f['properties']['SA2_MAIN16'],
				 'GCC': f['properties']['GCC_CODE16'],
				 'wkt': wkt.dumps(f['geometry'])
				 }, index = [1]
				 )
	return d


def dump_ste(f):
	d = pd.DataFrame(
				{'STE': f['properties']['STE_CODE16'],
				 'wkt': wkt.dumps(f['geometry'])
				 }, index = [1]
				 )
	return d


if __name__ == "__main__":
	with open('../shapefiles/SA2_2016_AUST.geojson', 'r') as f:
		j = json.load(f)

	out = map(dump_sa2, j['features'])
	out = pd.concat(out)
	out.to_csv('../shapefiles/SA2_wkt.csv', index =False)


	with open('../shapefiles/STE_2016_AUST.geojson', 'r') as f:
		j = json.load(f)

	out = map(dump_ste, j['features'])
	out = pd.concat(out)
	out.to_csv('../shapefiles/STE_wkt.csv', index =False)

