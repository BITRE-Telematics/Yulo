# import.sh
Created Tuesday 01 May 2018

import.sh is a bash script to build the psql database. It has been altered to change arguments referring to the right map file and config. It has a new argument referring to [probbo-roads.json](./probbo-roads.json.md)
It also has been altered to refer directly to the location of the osmosis executable since it was not excuting properly as part of [process.sh.](./process.sh..md)
The osmosis commands have also been altered to accept ways with route=ferry, whereas previously only ways with a highway tag were accepted

