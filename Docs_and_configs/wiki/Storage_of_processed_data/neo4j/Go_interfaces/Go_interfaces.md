# Go interfaces
Created Monday 20 August 2018

As of August 2018 I am experimenting with go interfaces for bulk transactions on Neo4j.

I am using the neoism interface. An important thing to note is the results structs must have labels that begin with capitals, else no results will be returned. No error will occur, just no results.

It also use the http connector, not bolt, which the newly released driver does.

2020-07-20 - The Go code now uses the official driver which uses interfaces not structs. 

