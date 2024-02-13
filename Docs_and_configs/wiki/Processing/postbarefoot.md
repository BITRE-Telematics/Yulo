# postbarefoot
Created Friday 27 April 2018


This describes a portion of [[Yuloserver]] processing trips after they have been mapmatched by Barefoot. The text dates from when this was a separate script.

Note that sometimes the shapefiles are incorrect. This means a stop or point will not me matched to the local STE because the location is outside the shapefile's polygon. These can be identified in edge cases and corrected manually both for the affected points (which will have an Olson value of UTC, STE and SA2s of 0) and the shapefiles edited to prevent that case in future. The code also searches for these points and matches to the nearest polygon, although this is somewhat computationally intensive. This could be reduced somewhat (since it is not dependent on the ASGS as such) by using OSM's state boundaries that include state waters. Since some points are in other territories I am also matching these to the nearest state. Hopefully this will only be relevant for Jervis Bay, which uses NSW time, and  Norfolk Island, Christmas Island and Cocos Island are not witnessed.

Because of various edge cases the dataframe being written at the end might not always have the same number of fields and subsequently the appended csv may have mismatched columns because of too many or few headers. I should make this safer by checking for columns but I haven't got around to it.

