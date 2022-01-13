package ys



import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/structs"
	//"yaml"
	//"github.com/paulmach/orb"
)

type writeStop struct {
	STOPID      string
	START       int64
	START_UTC   int64
	START_UTCDT string
	STARTDT     string
	END         int64
	END_UTC     int64
	END_UTCDT   string
	ENDDT       string
	LAT         float64
	LON         float64
	SA2         string
	GCC         string
	VEHICLE     string
	LOC         string
	ADDR        string
}

func mapstops(stops []processedStop, id string) map[string]interface{} {
	stops_mapped := make([]map[string]interface{}, len(stops))
	for i, stop := range stops {
		parameters := writeStop{
			STOPID:      stop.Stopid,
			START:       stop.Start,
			START_UTC:   stop.Start_utc,
			START_UTCDT: stop.Start_utcdt,
			STARTDT:     stop.Startdt,
			END:         stop.End,
			END_UTC:     stop.End_utc,
			END_UTCDT:   stop.End_utcdt,
			ENDDT:       stop.Enddt,
			LAT:         stop.Lat,
			LON:         stop.Lon,
			SA2:         stop.Sa2,
			GCC:         stop.Gcc,
			VEHICLE:     id,
			LOC:         stop.Loc,
			ADDR:        stop.Addr,
		}
		stops_mapped[i] = structs.Map(parameters)
	}
	out := map[string]interface{}{
		"VEHICLE": id,
		"STOPS":   stops_mapped,
	}

	return out
}

func stopswrite(stops []processedStop, id string, i int) {
	fmt.Println("writing stops for %s", id)
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	stopquery := `
				MERGE (vehicle:Asset{
					id: $VEHICLE
				}
				)
				WITH vehicle
				UNWIND $STOPS as s
				MERGE (stop:Stop{
				  id: s.STOPID})

				 SET
				  stop.start_time = toInteger(s.START),
				  stop.start_time_utc = toInteger(s.START_UTC),
				  stop.start_time_utcdt = datetime(s.START_UTCDT),
				  stop.start_timedt = datetime(s.STARTDT),
				  stop.end_time = toInteger(s.END),
				  stop.end_time_utc = toInteger(s.END_UTC),
				  stop.end_time_utcdt = datetime(s.END_UTCDT),
				  stop.end_timedt = datetime(s.ENDDT),
				  stop.lat = toFloat(s.LAT),
				  stop.lon = toFloat(s.LON),
				  stop.sa2 = s.SA2,
				  stop.gcc = s.GCC,
				  stop.added = timestamp()/1000

				

				MERGE(sa2:SA2{
				  sa2_code: s.SA2
				})



				MERGE (vehicle)-[:STOPPED_AT]->(stop)
				MERGE (stop)-[:IN]->(sa2)




				FOREACH (ignoreMe IN CASE WHEN s.ADDR <> "" THEN [1] ELSE [] END |
		        MERGE (addr:Address{
				  id: s.ADDR
				})

				MERGE (stop)-[:AT]->(addr)
			    )

			    FOREACH (ignoreMe IN CASE WHEN s.LOC <> "" THEN [1] ELSE [] END |
		        MERGE (addr:Location{
				  id: s.LOC
				})

				MERGE (stop)-[:USED]->(addr)
				)
	`
	parameters := mapstops(stops, id)
	//fmt.Println(parameters)
	res, err := session.Run(stopquery, parameters)

	if err != nil && i < 400 {
		if i > 0 {
			fmt.Println(err, id, i)
		}
		time.Sleep(time.Second * 60)
		stopswrite(stops, id, i+1)
	}

	if res != nil {
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
	}

}

type writeObv struct {
	OSM_ID         string
	AZIMUTH        string
	DATETIME       int64
	DATETIME_UTC   int64
	DATETIME_DT    string
	DATETIME_UTCDT string
	SPEED          string
	IMPUTED_SPEED  string
	STE            string
	LAT            float64
	LON            float64
	LENGTH         string
	TYPE           string
	VEHICLE        string
	TRIP           string
	IMP_OBVS       []string
	FORWARD        bool
}

func to_string(f float64, v int) string {
	vf := float64(v)
	if f < vf {
		return "NA"
	} else {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
}

func tripwrite(trips []processedTrip, i int) {
	session := Db.NewSession(Sesh_config)

	defer session.Close()

	tripquery := `
		UNWIND $TRIPS as t
			MATCH (trip:Trip{
			  id: t.TRIP})
			//assumes stops already uploaded
			MATCH (prior_stop:Stop{
			  id: t.PRIOR_STOP})

			MATCH (following_stop:Stop{
			  id: t.FOLLOWING_STOP})

			CREATE (trip)-[:PRECEDED_BY]->(prior_stop)

			CREATE (trip)-[:FOLLOWED_BY]->(following_stop)

			WITH trip

			MATCH (last_stop:Stop)<-[:PRECEDED_BY]-(trip)-[:FOLLOWED_BY]->(next_stop:Stop)

			CREATE (last_stop)-[:NEXT_STOP]->(next_stop)

			

			`
	trips_mapped := make([]map[string]interface{}, len(trips))
	for i, trip := range trips {
		parameters := map[string]interface{}{
			"TRIP":           trip.Trip,
			"PRIOR_STOP":     trip.Prior_stop,
			"FOLLOWING_STOP": trip.Following_stop,
		}
		trips_mapped[i] = parameters
	}

	trips_out := map[string]interface{}{
		"TRIPS": trips_mapped,
	}
	//fmt.Println(trips_out)
	res, err := session.Run(tripquery, trips_out)
	if err != nil && i < 400 {
		if i > 1 {
			fmt.Println(err, trips_out, i)
		}
		time.Sleep(time.Second * 60)
		go tripwrite(trips, i+1)
	}
	if res != nil {
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
	}

}

func to_writeObv(obv processedObv, id string, trip string) writeObv {
	var o_type string
	if len(obv.Imp_Obvs) > 0 {
		o_type = "matched path"
	} else if obv.Osm_id != "unknown" {
		o_type = "matched no path"
	} else {
		o_type = "not matched"
	}
	writeObv := writeObv{
		VEHICLE:        id,
		OSM_ID:         obv.Osm_id,
		TRIP:           trip,
		DATETIME:       obv.Datetime,
		DATETIME_UTC:   obv.Datetime_utc,
		DATETIME_UTCDT: obv.Datetime_utcdt,
		DATETIME_DT:    obv.Datetime_dt,
		LAT:            obv.Point.Lat(),
		LON:            obv.Point.Lon(),
		SPEED:          to_string(obv.Speed, 1),
		IMPUTED_SPEED:  to_string(obv.Imputed_speed, 1),
		AZIMUTH:        to_string(obv.Azimuth, 0),
		LENGTH:         to_string(obv.Length, 1),
		TYPE:           o_type,
		FORWARD:        obv.Forward,
	}

	var imp_obv_str []string
	for _, o := range obv.Imp_Obvs {
		imp_obv_str = append(imp_obv_str, o.Osm_id+"$"+strconv.FormatBool(o.Forward))
	}
	writeObv.IMP_OBVS = imp_obv_str
	
	return writeObv
}

func maps_obvs(obvs []processedObv, id string, trip string) map[string]interface{} {
	obvs_mapped := make([]map[string]interface{}, len(obvs))
	for i, o := range obvs {
		owo := to_writeObv(o, id, trip)
		obvs_mapped[i] = structs.Map(owo)
	}
	out := map[string]interface{}{
		"VEHICLE": id,
		"TRIP":    trip,
		"OBVS":    obvs_mapped,
	}
	return out
}

func write_obs_batch(tripobvs map[string]interface{}, onExit func(), i int) {
	//fmt.Println(tripobvs)
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	obvquery := `
		MATCH (vehicle:Asset{
		  id: $VEHICLE})

		

		MERGE (trip:Trip{
		  id: $TRIP})
		
		CREATE (vehicle)-[:EMBARKED_ON]->(trip)

		WITH vehicle, trip
		
		UNWIND $OBVS as o
		
		

		CREATE (observation:Observation{
			speed: toFloat(o.SPEED),
			datetime: toInteger(o.DATETIME),
			datetime_utc: toInteger(o.DATETIME_UTC),
			datetime_utcdt: datetime(o.DATETIME_UTCDT),
			datetimedt: datetime(o.DATETIMEDT),
			lat: toFloat(o.LAT),
			lon: toFloat(o.LON),
			imputed_speed: toFloat(o.IMPUTED_SPEED),
			azimuth: toInteger(o.AZIMUTH),
			length: toFloat(o.LENGTH),
			type: o.TYPE,
			add_date: timestamp()/1000,
			forward: toBoolean(o.FORWARD)
		})
		WITH *
		WHERE o.OSM_ID <> "unknown"
		
		MERGE (segment:Segment{
			osm_id: o.OSM_ID
			})

		CREATE (trip)-[:OBSERVED_AT]->(observation)
		
		CREATE (observation)-[on:ON]->(segment) 
		SET on.type = o.TYPE, on.forward = toBoolean(o.FORWARD)
		
		WITH observation, o, segment
		UNWIND o.IMP_OBVS as impobvs
		MERGE(impseg:Segment{osm_id: split(impobvs, '$')[0] })
		CREATE (observation)-[imp_on:ON]->(impseg)
		SET imp_on.type = 'imputed', imp_on.forward = toBoolean(split(impobvs, '$')[1])
		
		
		RETURN segment.osm_id

			`

	res, err := session.Run(obvquery, tripobvs)
	if err != nil && i < 400 {
		if i > 0 {
			fmt.Println(err, tripobvs["VEHICLE"], i)
			//fmt.Println(obv)
		}
		time.Sleep(time.Second * 60)
		write_obvs_batch(tripobvs, onExit, i+1)
	}

	if res != nil {
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
		onExit()
	}

}

func tripswrite(trips []processedTrip, id string) {
	var wg sync.WaitGroup
	fmt.Println("writing observations and trips for %s\n", id)
	for _, t := range trips {

		wg.Add(1)
		
		obvs_mapped := maps_obvs(t.Obvs, id, t.Trip)
		
		//fmt.Println(obvs_mapped)
		go write_obs_batch(obvs_mapped, func() { wg.Done() }, 1)
	}
	

	
	wg.Wait()
	tripwrite(trips, 1)
}
