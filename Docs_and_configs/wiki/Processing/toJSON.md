# toJSON
Created Friday 27 April 2018

This script is now redundant, which it's important features built into [TripGrouping](./Tripgrouping.txt)

This script prepares the trip GPS points for input to [Processing:Barefoot](./Barefoot.md) by creating a JSON format, including GPS in WKT format and azimuth where available. These are saved in the directory “JSONInput”.
Notably it also turns the datetime into milliseconds as used by Barefoot.
It also has a command line argument -r --[rerun](./reruns.md) to indicate whether it should only process those trips which [have been previously matched to problematic off ramps with associated routing errors.](./Matching_issues_and_errors/GPS_error.md) This will recreate the JSON with the problematic observation removed. 

