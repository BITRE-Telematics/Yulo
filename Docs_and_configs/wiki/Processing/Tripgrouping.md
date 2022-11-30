# Tripgrouping
Created Friday 27 April 2018

Cich style trip grouping/
TripGrouping/ TripgroupingFunctional.py

This script iterates though points by a vehicle defining stop and trip events. It uses the approach
described in Cich et al with some modification. Because a trip and stop is only defined at the end of the subsequent stop, the trip/stop pair in process at the end of the month will not be attributed a trip or stop id. This data is residual and can be used the next period.
I have recently (July 2017) added a further element whereby, at the end of a trip stop pair, the last point of the previous stop and the first point of the next stop are attributed to the trip as well. This is to make sure trips start and end near the stop location. The need for this became apparent for firms whose depots were close to an SA4 boundary where trips would not be registered until after the vehicle crossed the boundary, and so the trip start would be in another SA4.
It would be more robust to take the centroids of the stops and merge it into the trip data after processing, particularly because the last stop observation may record the vehicle already moving away from the stop, but within the stop threshold distance of the centroid. This can still be done (in SummaryTrips.r) if the results of the implemented method are unsatisfactory (as might occur when a trip is caused by a jump and the tripgrouping function doesn’t like it), but the latter was easier to implement.

UPDATE — Queries to the database aggregate trips based on stops so this is immaterial


This procedure, especially in the second pass, also has the effect of removing erroneous GPS readings, including jumps into the ocean, up to the equator, or to locations measured some time previously.
The parameters for the trip grouping are input using config.yaml. These have not been calibrated yet. 
There are three main functions. CichCluster() defines stop events and the intervening periods which are defined as trips. Cich2ndPass() processes the trips defined in CichCluster() to remove erroneous GPS points and define subtrips where there is a substantial gap (by default 15 minutes) between GPS points. This reduces the strain on the current version of Barefoot, but was developed as a debugging measure when I was cleaning out the jumps. CichProcess() combines these two. Data read in and configuration are handled by the main method.
NOTE that the object func_args is a list and not a dictionary. AS such the arguments must be in the same order they are listed in the definition of CichProcess(). It is not reading the names of variables.
The processed data is exported into three directories.
Trips: A CSV for each vehicle with all trip GPS points plus their assigned trip and subtrip.
Stops: A CSV for each vehicle with each stop as one observation with centroid and start/end times.
Residuals: All points at the end of the period not attributed to a trip or stop, to be used next 	period.
Stops are concatenated when they are near in time and place. This means that there will usually be more trips defined than stops. The trips between concatenated stops should be discarded in later analysis for being too short (by default they will be less than 10 minutes). Although these points are formally part of the stop event, they will add no data other than slightly changing the centroid of the stop.

Some vehicles will produce stops but no trips. This is because they have a single stop that ends before the end of the period, but do not produce another trip stop pair before the end of the period. Note this cannot happen the other way round since trips are only recorded at the end of a stop.

This code would ideally be implemented in C or another faster language, however I had difficulty using classes in C as effectively as in Python. With many cores the speed slow down should be reduced dramatically.

In December 2017 I discovered an error in the trip grouping where two logical tests had been nested incorrectly. This was leading to discrepancy between trips and stops.

In order to reduce the risk of things like this I have a strong urge to implement the algorithm with more functional principles, but I am aware this will likely reduce readability for my successors.  UPDATE: Tested and working but also apparently 60 to 70 per cent faster. My speculative explanation for this is that numba’s ability to semi compile functions is limited by its capacity to infer type, and it was struggling with the state variables compared to the dictionary types, but I don’t know how much credence this speculation has.

The functional script also takes a command line argument which can override the input file specified in config yaml.

I have added an option to add filtering to the second pass algorithm to pull out replicated GPS points that are [causing problems](./Matching_issues_and_errors/zero_imputed_speeds.md). We should consider moving this to the 1st pass algorithm as these sequences of ostensibly stationary pings could be interpreteted as a stop.
This would be achieved by putting the following code at the beginning of CichIter() along with adding the skipDupes argument to the necessary functions and removing the equivalent code from Cich2ndPassIter(). It will also require expanding the dictionary object to take the last GPS ping. I have done this but not checked the results yet
I have belatedly discovered some of the data specifies a boolean as to whether the GPS is valid before hand, letting one filter it out. Oops. Still, this flag is available for other cases.

In AUgust 2018 I discovered an error that was attributing the wrong point to residuals meaning in rare cases a stop id would be duplicated. This is fixed.


Note that the second pass algorithm implicitly assumes the first observation in a trip is valid

