# Matching issues and errors
Created Friday 27 April 2018

NOTES ON MAP ERRORS

There are roughly four categories of matching error that require different reponses. I am also recording unresolved cases here.

[+Improper map construction](./Matching_issues_and_errors/Improper_map_construction.md)
[+Map connectivity error](./Matching_issues_and_errors/Map_connectivity_error.md)
[+GPS error](./Matching_issues_and_errors/GPS_error.md)
[+Unknown](./Matching_issues_and_errors/Unknown.md)

However before considering these it is extremely important to note where data might be prematched, and matched incorrectly. Just as Barefoot is intended to correct the location of position on in vehicle displays, some of the data from specific providers appears to have been adjusted to nearest neighbour analysis. In many cases this has exacerbated the existing GPS error and, by placing points directly on a segment, made it very hard to correct.

Normal signs of error are imputed and recorded speeds well above the speed limit, and very high ratios of imputed to recorded points even when there is no clear reason for few GPS records (ie the segment isn't very short or in a tunnel)

The are also anomalies which may not be errors

[+Anomalies](./Matching_issues_and_errors/Anomalies.md)
+[Ferries](./Ferries.md)


When such errors are identified well after processing we can request all observations from trips that have used the incorrectly matched segment and pass the observations through Barefoot again before deleting the erroneous reading and uploading replacements using CorrectR.


