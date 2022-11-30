# cypher
Created Friday 04 May 2018

Cypher is the query language for [Neo4j.](../neo4j.md) There are good resources online so this page only includes a few general notes, especially with its relationship to RCypher and to help understand the [queries](../../Display_and_queries/Queries.md) I have written.

With particularly large queries Neo4j will sometimes return unknown errors, although persisting will bring these through. I have never run a single database long enough to know (sinc ethe project is still in development) but I think the database is reindexing itself in response to queries and helping performance. This may end with better system resources as well.

Cypher can accept a IN argument that takes a list (a vector in R) of arguments, however this has inconsistent effects on speed. As such I have written many queries both by the IN form and using map() create a vectors of queries with each identifier and concatenating them with UNION ALL so they return a single table. Both of these options involve double queries (first to get identifiers and then traverse from there) which perhaps can be optimised.

Note that they keys of a node can be returned with 

	MATCH (n:Node) WITH DISTINCT keys(n) as keys
	UNWIND keys AS keylist 
	WITH DISTINCT keylist AS allfields
	RETURN allfields

