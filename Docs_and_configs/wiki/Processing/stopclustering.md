# stopclustering
Created Friday 01 February 2019

stop_cluster.py takes stop data from a given month and defines clusters using the DBSCAN algorithm. It is separated from the rest of the process so individual firms can be processed seperately and the relatively unintiensive work of clustering, which requires all the data for a month to be available, can be done at a later time.

It also matches stops and clusters to known locations, either rest areas, loading zones or street addresses. The former two are done concurrently and the latter by itself. As such, a stop or cluster may be matched to a Location (rest area or loading zone) AND a street address.

The addresses are derived from the Geocoded National Address File (G-NAF) and can be found at the url in [Storage of processed data:addresses.py](../Storage_of_processed_data/addresses.py.md)

The location data is from [Storage of processed data:rest and loading areas.py](../Storage_of_processed_data/rest_and_loading_areas.py.md) and is derived from a large collection of state sources.

For some mystery reason Queensland rest areas do not match unless processed entirely seperately. There is no obvious different between the data to expect this. But this is why lz_rest_match() processes them separately. The distance is measured as euclidean distance because it is faster.

