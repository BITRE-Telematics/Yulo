package ys

func cypherScript() map[string]string {
	script := make(map[string]string)

	script["stops"] = `
		USING PERIODIC COMMIT 1000
		LOAD CSV WITH HEADERS FROM 'file:///UploadData/StopsOut%s.csv' AS row

		

		MERGE (stop:Stop{
		  id: row.Stop
		})
		 SET
		  stop.start_time = toInteger(row.start_time),
		  stop.start_time_utc = toInteger(row.start_time_utc),
		  stop.start_time_utcdt = datetime(row.start_time_utcdt),
		  stop.start_timedt = datetime(row.start_timedt),
		  stop.end_time = toInteger(row.end_time),
		  stop.end_time_utc = toInteger(row.end_time_utc),
		  stop.end_time_utcdt = datetime(row.end_time_utcdt),
		  stop.end_timedt = datetime(row.end_timedt),
		  stop.lat = toFloat(row.lat),
		  stop.lon = toFloat(row.lon),
		  stop.sa2 = row.sa2,
		  stop.ste = toInteger(left(row.sa2, 1)),
		  stop.gcc = row.gcc, 
		  stop.added = timestamp()/1000
		
		

		MERGE (vehicle:Asset{
		  id: row.Vehicle})

		



		MERGE (vehicle)-[:STOPPED_AT]->(stop)

		FOREACH (ignoreMe IN CASE WHEN exists(row.addr) THEN [1] ELSE [] END | 
        MERGE (addr:Address{
		  id: row.addr
		})
		MERGE (stop)-[:AT]->(addr)
	    )

	    FOREACH (ignoreMe IN CASE WHEN exists(row.loc) THEN [1] ELSE [] END | 
        MERGE (addr:Location{
		  id: row.loc
		})
		MERGE (stop)-[:USED]->(addr)
	    )
		

		`

	script["obvs"] = `
		USING PERIODIC COMMIT 10000
		LOAD CSV WITH HEADERS FROM 'file:///UploadData/mergedPBF%s.csv' AS row

		MATCH (vehicle:Asset{
		  id: row.Vehicle})

		MERGE (segment:Segment{
		  osm_id: row.osm_id
		  })

		MERGE (trip:Trip{
		  id: row.Trip})
		

		CREATE (observation:Observation{
		  speed: toFloat(row.speed),
		  datetime: toInteger(row.datetime),
		  datetime_utc: toInteger(row.datetime_utc),
		  datetime_utcdt: datetime(row.datetime_utcdt),
		  datetimedt: datetime(row.datetimedt),
		  lat: toFloat(row.lat),
		  lon: toFloat(row.lon),
		  imputed_speed: toFloat(row.imputed_speed),
		  azimuth: toInteger(row.azimuth),
		  length: toFloat(row.length),
		  type: row.type,
		  add_date: timestamp()/1000,
		  target: row.target,
		  imputed_azimuth: toInteger(row.imputed_azimuth),
		  forward: toBoolean(row.forward)})


		MERGE (vehicle)-[:EMBARKED_ON]->(trip)

		CREATE (trip)-[:OBSERVED_AT]->(observation)

		CREATE (observation)-[on:ON]->(segment)
		SET on.type = row.type


		WITH observation, row
		UNWIND split(row.imp_obvs, '|') AS impobvs
		MERGE(impseg:Segment{osm_id: split(impobvs, '$')[0] })
		CREATE (observation)-[imp_on:ON]->(impseg)
		SET imp_on.type = 'imputed', imp_on.forward = toBoolean(split(impobvs, '$')[1]) 

		`

	// 	`
	//Adding other trip stop edges

	script["trips"] = `
		USING PERIODIC COMMIT 1000
		LOAD CSV WITH HEADERS FROM 'file:///UploadData/Tripframe%s.csv' AS row
		MATCH (trip:Trip{
		  id: row.Trip})
		//assumes stops already uploaded
		MATCH (prior_stop:Stop{
		  id: row.Prior_stop})

		MATCH (following_stop:Stop{
		  id: row.Following_stop})


		MERGE (trip)-[:PRECEDED_BY]->(prior_stop)

		MERGE (trip)-[:FOLLOWED_BY]->(following_stop)

		WITH trip

		MATCH (last_stop:Stop)<-[:PRECEDED_BY]-(trip)-[:FOLLOWED_BY]->(next_stop:Stop)

		MERGE (last_stop)-[:NEXT_STOP]->(next_stop)


		`

	return script
}
