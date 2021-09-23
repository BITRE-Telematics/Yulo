import geopandas as gpd


if __name__ == "__main__":
	STE = gpd.read_file('../shapefiles/STE_2021_AUST.geojson')
	STE = STE[[g is not None for g in STE.geometry]]
	STE.to_file("../shapefiles/STE_2021_AUST.geojson", driver="GeoJSON")

	SA2 = gpd.read_file('../shapefiles/SA2_2021_AUST.geojson')
	SA2 = SA2[[g is not None for g in SA2.geometry]]
	SA2.to_file("../shapefiles/SA2_2021_AUST.geojson", driver="GeoJSON")
