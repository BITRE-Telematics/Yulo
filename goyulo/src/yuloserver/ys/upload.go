package ys

// ////DO NOT USE - untested - make sure to check res.Err() for syntax errors

/*import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strconv"
	"time"
	//"yaml"
	//"github.com/paulmach/orb"
)



func stopwrite(stop processedStop, id string, i int) {

	session, err := Db.Session(neo4j.AccessModeRead)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()
	stopquery := `
			MERGE (stop:Stop{
				  id: $STOPID})

				 SET
				  stop.start_time = toInteger($START),
				  stop.start_time_utc = toInteger($START_UTC),
				  stop.start_time_utcdt = datetime($START_UTCDT),
				  stop.start_timedt = datetime($STARTDT),
				  stop.end_time = toInteger($END),
				  stop.end_time_utc = toInteger($END_UTC),
				  stop.end_time_utcdt = datetime($END_UTCDT),
				  stop.end_timedt = datetime($ENDDT),
				  stop.lat = toFloat($LAT),
				  stop.lon = toFloat($LON),
				  stop.sa2 = $SA2,
				  stop.gcc = $GCC,
				  stop.added = timestamp()/1000

				MERGE (vehicle:Asset{
				  id: $VEHICLE
				})

				MERGE (vehicle)-[:STOPPED_AT]->(stop)
				MERGE (stop)-[:USED]->($LOC)
				MERGE (stop)-[:AT]->($ADDR)
				;
			`
	parameters := map[string]interface{}{
		"STOPID":      stop.Stopid,
		"START":       stop.Start,
		"START_UTC":   stop.Start_utc,
		"START_UTCDT": stop.Start_utcdt,
		"STARTDT":     stop.Startdt,
		"END":         stop.End,
		"END_UTC":     stop.End_utc,
		"END_UTCDT":   stop.End_utcdt,
		"ENDDT":       stop.Enddt,
		"LAT":         stop.Lat,
		"LON":         stop.Lon,
		"SA2":         stop.Sa2,
		"GCC":         stop.Gcc,
		"VEHICLE":     id,
		"LOC":         stop.loc,
		"ADDR":        stop.addr,
	}
	//fmt.Println(parameters)
	_, err = session.Run(stopquery, parameters)
	if err != nil && i < 400 {
		if i > 5 {
			fmt.Println(err, stop.Stopid, i)
		}
		stopwrite(stop, id, i+1)
	}

}
func stopswrite(stops []processedStop, id string) {
	fmt.Printf("writing stops for %s\n", id)
	for _, stop := range stops {
		stopwrite(stop, id, 1)
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
}

func obvswrite(obv processedObv, id string, trip string) {
	var o_type string
	if len(obv.Imp_Obvs) > 1 {
		o_type = "matched path"
	} else if len(obv.Imp_Obvs) == 1 {
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
	}

	var imp_obv_str []string
	for _, o := range obv.Imp_Obvs {
		imp_obv_str = append(imp_obv_str, o.Osm_id+"$"+strconv.FormatBool(o.Forward))
	}
	writeObv.IMP_OBVS = imp_obv_str

	obvwritesingle(writeObv, id, 0)

}

func to_string(f float64, v int) string {
	vf := float64(v)
	if f < vf {
		return "NA"
	} else {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
}

//This doesn't account for the missing values values in the observations, use to_string() equivalent. Perhaps convert to string then back in database
func obvwritesingle(obv writeObv, id string, i int) {
	session, err := Db.Session(neo4j.AccessModeRead)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()
	obvquery := `
		MATCH (vehicle:Asset{
		  id: $VEHICLE})

		MATCH (segment:Segment{
		  osm_id: $OSM_ID
		  })

		MERGE (trip:Trip{
		  id: $TRIP})

		CREATE (observation:Observation{
		  speed: toFloat($SPEED),
		  datetime: toInteger($DATETIME),
		  datetime_utc: toInteger($DATETIME_UTC),
		  datetime_utcdt: datetime($DATETIME_UTCDT),
		  datetimedt: datetime($DATETIMEDT),
		  lat: toFloat($LAT),
		  lon: toFloat($LON),
		  imputed_speed: toFloat($IMPUTED_SPEED),
		  azimuth: toInteger($AZIMUTH),
		  length: toFloat($LENGTH),
		  type: $TYPE,
		  add_date: timestamp()/1000,
		  forward: toBoolean($FORWARD)})

		MERGE (vehicle)-[:EMBARKED_ON]->(trip)

		CREATE (trip)-[:OBSERVED_AT]->(observation)

		CREATE (observation)-[on:ON]->(segment)
		SET on.type = $TYPE

		WITH observation
		UNWIND $IMP_OBVS as impobvs
		MATCH(impseg:Segment{osm_id: split(impobvs, '$')[0] })
		CREATE (observation)-[imp_on:ON]->(impseg)
		SET imp_on.type = 'imputed', imp_on.forward = toBoolean(split(impobvs, '$')[1]);
			`
	parameters := map[string]interface{}{
		"VEHICLE":        id,
		"OSM_ID":         obv.OSM_ID,
		"TRIP":           obv.TRIP,
		"DATETIME":       obv.DATETIME,
		"DATETIME_UTC":   obv.DATETIME_UTC,
		"DATETIME_UTCDT": obv.DATETIME_UTCDT,
		"DATETIMEDT":     obv.DATETIME_DT,
		"LAT":            obv.LAT,
		"SPEED":          obv.SPEED,
		"IMPUTED_SPEED":  obv.IMPUTED_SPEED,
		"AZIMUTH":        obv.AZIMUTH,
		"LENGTH":         obv.LENGTH,
		"TYPE":           obv.TYPE,
		"IMP_OBVS":       obv.IMP_OBVS,
	}

	_, err = session.Run(obvquery, parameters)
	if err != nil && i < 400 {
		if i > 10 {
			fmt.Println(err, id, i)
			//fmt.Println(obv)
		}
		time.Sleep(time.Millisecond * 200)
		obvwritesingle(obv, id, i+1)
	}
}

func tripwrite(trip processedTrip, i int) {
	session, err := Db.Session(neo4j.AccessModeWrite)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	tripquery := `
			MATCH (trip:Trip{
			  id: $TRIP})
			//assumes stops already uploaded
			MATCH (prior_stop:Stop{
			  id: $PRIOR_STOP})

			MATCH (following_stop:Stop{
			  id: $FOLLOWING_STOP})

			CREATE (trip)-[:PRECEDED_BY]->(prior_stop)

			CREATE (trip)-[:FOLLOWED_BY]->(following_stop)

			WITH trip

			MATCH (last_stop:Stop)<-[:PRECEDED_BY]-(trip)-[:FOLLOWED_BY]->(next_stop:Stop)

			CREATE (last_stop)-[:NEXT_STOP]->(next_stop)

			WITH trip

			CREATE (last_trip)-[:NEXT_TRIP]->(trip)

			`
	parameters := map[string]interface{}{
		"TRIP":           trip.Trip,
		"PRIOR_STOP":     trip.Prior_stop,
		"FOLLOWING_STOP": trip.Following_stop,
	}

	_, err = session.Run(tripquery, parameters)
	if err != nil && i < 400 {
		if i > 50 {
			fmt.Println(err, trip.Trip, i)
		}
		time.Sleep(time.Millisecond * 200)
		tripwrite(trip, i+1)
	}

}

func tripswrite(trips []processedTrip, id string) {
	fmt.Printf("writing observations and trips for %s\n", id)
	for _, t := range trips {
		for _, o := range t.Obvs {
			obvswrite(o, id, t.Trip)
			//fmt.Print(o)
		}
		tripwrite(t, 1)

	}
}
*/
