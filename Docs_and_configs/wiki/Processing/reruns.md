# reruns
Created Friday 25 May 2018

Rerunning is an indevelopment process (2018-05-25) to take trips which barefoot has run on a small number of problematic off ramp segments where [GPS error](./Matching_issues_and_errors/GPS_error.md) is leading [Barefoot](./Barefoot.md) to incorrectly match observations to off ramps, and subsequently later routes the trip to side roads which, as a result, have erroneous (and very high) imputed speeds and recorded speeds where subsequent pings are mismatched because they cannot be routed back on to the motoway.

The code now looks for matches onto recognised problematic segments (in probbo-ramps.txt) due to [GPS error](./Matching_issues_and_errors/GPS_error.md) and records the observations whilst neglecting to process the vehicle further. Then after being fed through [toJSON.r](./toJSON.md) and barefoot again without the observation matched to an offramp. Ideally this means barefoot will correctly route to the next observation whether it is . This process is lumped into postbarebash.sh.

As are 2018-06-05 there are approximately 60 identified offramps, however I have included an option in postbarefootmerge.r to do this to all segments identified as "motorway_link".

UPDATE - Most of these errors are attributable to a single firms which had prematched it's data using nearest neighbour. For problematics segments, where data is plentiful as they are usually motorways, we can simply exclude this firm. I have not incoporated rerun code into postbarefoot.py

