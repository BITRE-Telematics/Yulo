# fabric
Created Tuesday 12 July 2022

Neo4j provides a system of sharding databases into multiple databases that can be stored on one of multiple machines. In cases where there is little cross over between areas this can improve scale.

In the case of Yulo I have sharded by time, in particular by year. Because indices increase write time but reduce query time I process a years worth of data on each shard, and then add an index on the datetime attributes of observations at the end. This takes about a day but dramatically improves query times. Yuloserver assumes a fabric set up for attaching stops to each other.

See official Neo4j documentation for more detail.

