# Ferries
Created Thursday 02 August 2018

In some cases freight vehicles use vehicular ferries. The vehicles may not ping on the route if the engines and electrics are entirely off, or ping at a lower rate. Nonetheless the ferry is part of the freight network and we may have to consider incorporating them somehow.
This could be done by editing [bfmap.py](./Barefoot/bfmap.py.md) to incorporate ways without a highway tag but with route - ferry and motor_vehicle - yes. Alternatively the number of vehicular ferries is low enough that specific analyses can be provided for each that exists.

UPDATE

I have edited import.sh to do this successfully.

