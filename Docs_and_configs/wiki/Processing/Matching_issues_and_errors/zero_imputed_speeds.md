# zero imputed speeds
Created Tuesday 24 July 2018

The data available in 2018 seems to return the last GPS co-ordinates when it is unable to record a location. This differs from previous data which returned either obviously errneous co-ordinates or NAs.
This poses a problem because [Barefoot](../Barefoot.md) will match to the same segment at the same point and impute a zero speed. This means the imputed speed for tunnels will tend to be zero.
There is an optipon to filter out replicated lan lons in [Tripgrouping](../Tripgrouping.md) in the 2nd pass algorithm (that is after stops are detected). This shouldn't sacrifice data as the recorded speeds on the replicating pings are the same as well which suggests the speed variable is GPS based here as well. Potentially this could remove some valid observations where the vehicle genuinely is stationary (for less than the stop threshold time) but this seems unlikely. The data sources where this is happening also report lat lons of 5 and 4 decimal places respectively meaning a genuinely stationary vehicle will likely report different lat lons just from measurement error. I think we can afford any leakage that happens in the rare instances where a vehicle manages to replicate a lat long exactly. If it does occur more often than we'd wish we can use the speed variable as well.

Some firms provide a boolean when the GPS is invalid and so using this to filter it out beforehand is a better option.

