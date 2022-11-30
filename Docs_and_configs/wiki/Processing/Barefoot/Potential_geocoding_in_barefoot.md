# Potential geocoding in barefoot
Created Wednesday 31 July 2019

Current [postbarefoot](../postbarefoot.md) spends a great deal of time geocoding points to states. Potentially, if improbably, this could be achieved faster by moving the geocoding into Barefoot. Barefoot already imports com.esri.geometry and [Processing:Barefoot:MatcherKState.java](./MatcherKState.java.md) could pull in the Point and Polygon classes and use the contains/within methods of either to check.

However postbarefoot would still need to do the second order matching for objects outside the ASGS polygons. Also the Python functions are an implementation of GEOS which is already written in C and C++ as a port of Java's JTS toolseos, so speed benefits would probably be absent

