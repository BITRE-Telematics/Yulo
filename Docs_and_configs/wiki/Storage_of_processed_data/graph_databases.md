# graph databases
Created Friday 27 April 2018

I have chosen to store data as a graph database, where each node is a segment, a trip or a stop, and edges include things like “observed at <datetime> at <speed>”. A simple schematic is below. T
There are a number of graph frameworks available with open source licences of carrying permissions, including neo4j and orientdb. Rando internet comments seem to indicate that [neo4j](./neo4j.md), whilst requiring payment for commercial uses, is better for heavy read applications, which is what our crucial point will be. Speed is less crucial for write tasks which will be batched anyway. [Orientdb](./orientdb.md) does not require payment, but is more poorly documented and less efficient. Neo4j is the current choice because of its better bulk upload facilities


![](file:///home/rigreen/Documents/graph.png)

