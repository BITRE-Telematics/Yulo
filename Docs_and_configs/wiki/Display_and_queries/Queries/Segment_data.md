# Segment data
Created Friday 27 April 2018

Note the awkward expression 
	sum(size(filter(x IN o.type WHERE x <> 'imputed')))
Which is necessary because there is no simple way to aggregate a boolean in cypher separate from the match-where functionality. A count of the boolean expression would just give the total number of nodes. As such I have taken a value only when it meets the condition, counted the length of each of these (which will always be 1 or 0) and summed it.
Awkward.





The [Go interface](../../Storage_of_processed_data/neo4j/Go_interfaces.md) seems to be much faster

