package ys

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

//checks database to see if data already present
var Sesh_config neo4j.SessionConfig

func get_min_obv_time(obvs []obv) int64 {
	min := obvs[0].datetime
	//fmt.Printf("Min [0] is %d\n", min)
	for _, o := range obvs {
		if o.datetime < min {
			min = o.datetime
		}
	}
	return min

}

func checkDatabaseDupe(obvs []obv, max_prune int64) (bool, int64) {
	id := obvs[0].id
	//fmt.Printf("I have %s as an id\n", id)
	min_dt := get_min_obv_time(obvs)
	//fmt.Printf("I have %s as a dt\n", min_dt)
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	statement := ""
	if max_prune == 0 {
		statement = `
		MATCH(a:Asset{id: $ID})-[:STOPPED_AT]->(s:Stop)
		RETURN max(s.end_time_utc) as max
	 `
	} else {
		statement = `
		MATCH(a:Asset{id: $ID})-[:STOPPED_AT]->(s:Stop)
		WHERE s.end_time_utc <= $MAX_PRUNE
		RETURN max(s.end_time_utc) as max
	 `
	}
	parameters := map[string]interface{}{"ID": id, "MAX_PRUNE": max_prune}

	max_result, err := session.Run(statement, parameters)
	if err != nil {
		fmt.Print("Check database error")
	}
	if max_result.Err() != nil {
		fmt.Print("Check database error")
		fmt.Println(max_result.Err())
	}

	if max_result.Next() {
		max := max_result.Record().GetByIndex(0)
		if max != nil {
			max_db_dt := max.(int64)
			return (min_dt <= max_db_dt), max_db_dt
		} else {
			fmt.Printf("No max time in db for %s \n", id)
			return false, 0
		}
	} else {
		fmt.Printf("No max time in db for %s \n", id)
		return false, 0
	}

}

func prune_dupes(obvs []obv, max_db_time int64) []obv {
	fmt.Printf("Pruning observations below %s\n", max_db_time)
	var obvs_out []obv
	for _, o := range obvs {
		if o.datetime >= max_db_time {
			obvs_out = append(obvs_out, o)
		}
	}
	return obvs_out

}

func inc_zero_dt(obvs []obv) bool {
	for _, o := range obvs {
		if o.datetime == 0 {
			return true
		}
	}
	return false

}

func checkPriorStop(id string, end_time int64) string {

	session := Db.NewSession(Sesh_config)

	defer session.Close()

	statement := "MATCH(a:Asset{id: $ID})-[:STOPPED_AT]->(s:Stop{end_time_utc:$END_TIME}) RETURN s.id as id"
	parameters := map[string]interface{}{"ID": id, "END_TIME": end_time}
	ps_res, err := session.Run(statement, parameters)
	if err != nil {
		fmt.Print("Check database error")
	}
	if ps_res.Next() {
		return ps_res.Record().GetByIndex(0).(string)
	} else {

		fmt.Printf("Resorting to most recent stop for %s at %d\n", id, end_time)
		return checkMostRecentStop(id, end_time)
	}
}

func checkMostRecentStop(id string, end_time int64) string {

	//this assumes you aren't adding data that dates prior to the data already in the database
	//A WHERE max < $END_TIME could account for this but the query does not work properly
	//In this instance use a script to correct the connections
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	statement := `
                  MATCH(a:Asset{id: $ID})-[:STOPPED_AT]->(s:Stop)
                  WHERE s.end_time_utc < $END_TIME
                  WITH max(s.end_time_utc) as max, a 
                  MATCH(a)-[:STOPPED_AT]->(ss:Stop{end_time_utc:max})
                  RETURN ss.id as id, ss.end_time_utc as end_time
                 `
	parameters := map[string]interface{}{"ID": id, "END_TIME": end_time}
	ps_res, err := session.Run(statement, parameters)
	if err != nil {
		fmt.Print("Check database error")
	}
	if ps_res.Next() {

		gap := end_time - ps_res.Record().GetByIndex(1).(int64)
		if (gap <= Params.Max_stop_gap) && (gap > 0) {
			fmt.Printf("Found most recent stop for %s with gap %d\n", id, gap)
			return ps_res.Record().GetByIndex(0).(string)
		} else {
			fmt.Printf("No recent stop found for %s at %d with smallest gap %d > %d\n", id, end_time, gap, Params.Max_stop_gap)
			return ""
		}
	} else {
		fmt.Printf("No recent stop found for %s at %d \n", id, end_time)
		return ""
	}
}
