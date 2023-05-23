package queries

import (
	"encoding/csv"
	"fmt"
	//"github.com/jmcvetta/neoism"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	"os"
	"sort"
	"strconv"
	//"strings"
	"sync"
	"time"
)

func act_writer(w *csv.Writer, c chan []string) {
	for l := range c {
		w.Write(l)
		w.Flush()

	}
}

func error_writer(w *csv.Writer, c chan []string) {
	for l := range c {
		w.Write(l)
		w.Flush()

	}
}

func batch_query_length(batch_id string, c chan []string, onExit func(), ind int) {

	defer onExit()
	act_data := Act_length_query(batch_id, 1)
	for _, rec := range act_data {
		//wg.Add(1)

		l := []string{
			stringGet(rec, "vehicle"), //vehicle
			unspecifiedNumFrmt(rec, "sum_length"),
			strconv.FormatInt(intGet(rec, "month"), 10), //month
			strconv.FormatInt(intGet(rec, "day"), 10),   //day
			stringGet(rec, "state"),                     //state
			unspecifiedNumFrmt(rec, "n_seg"),
		}
		//fmt.Println(l)
		c <- l

	}
	if ind%1 == 0 {
		fmt.Println(ind)
		fmt.Println(time.Now())
	}

}

func batch_query_usage(batch_id string, c chan []string, onExit func(), ind int) {

	defer onExit()
	act_data := Act_usage_query(batch_id, 1)

	for _, rec := range act_data {
		//wg.Add(1)
		l := []string{
			stringGet(rec, "vehicle"),                        //vehicle
			strconv.FormatInt(intGet(rec, "start_time"), 10), //start_time
			strconv.FormatInt(intGet(rec, "end_time"), 10),   //end_time
			strconv.FormatInt(intGet(rec, "month"), 10),      //month
			strconv.FormatInt(intGet(rec, "day"), 10),        //day
			stringGet(rec, "sa2"),                            //sa2
		}
		//fmt.Println(l)
		c <- l

	}
	if ind%100 == 0 {
		fmt.Println(ind)
		fmt.Println(time.Now())
	}

}

func Act_length_query(id string, i int) []*neo4j.Record {
	//res := ActLength{}

	statement := `MATCH(v:Asset{id:$VEH})-[:EMBARKED_ON]->(Trip)-[:OBSERVED_AT]->(o:Observation)-[on:ON]->(s:Segment)
       			  WHERE (on.type <> 'imputed' or on.type <> 'source' OR on.type IS NULL) AND o.datetimedt.year = $YEAR %s 
      			  with v, o, s, collect(distinct o) as oo
      			  RETURN
      			   v.id as vehicle,
      			   sum(o.length) as sum_length,
      			   o.datetimedt.month as month,
      			   o.datetimedt.day as day,
      			   s.osm_id as osm_id
                 `
	fabric_prefix := "UNWIND graph.names() AS g CALL {USE " + "graph.byName(g) "
	fabric_suffix := `} 
						WITH sum_length, month, day, osm_id, vehicle
						UNWIND osm_id as id 
						CALL{
							USE %s WITH id 
							OPTIONAL MATCH(ss:Segment) WHERE ss.osm_id = id RETURN left(ss.sa2, 1) as state
						}
						RETURN vehicle, sum(sum_length) as sum_length, month, day, state, count(osm_id) as n_seg
						`
	fabric_suffix = fmt.Sprintf(fabric_suffix, Seg_db)
	var parameters map[string]interface{}
	if (Month > 0) && (Month < 13) {
		//str_month := strconv.FormatInt(Month, 10)
		statement = fmt.Sprintf(statement, "AND o.datetimedt.month = $MONTH")

		parameters = map[string]interface{}{"VEH": id, "YEAR": Year, "MONTH": Month}

	} else {
		statement = fmt.Sprintf(statement, "")
		parameters = map[string]interface{}{"VEH": id, "YEAR": Year, "MONTH": Month}

	}
	session := Db.NewSession(Sesh_config)

	statement = fabric_prefix + statement + fabric_suffix
	//fmt.Println(statement)
	defer session.Close()
	result, err := session.Run(statement, parameters)
	//fmt.Println(err)
	result_, _ := result.Collect()
	if err != nil {

		fmt.Println(err, id, i)
		err_c <- []string{
			id,
			err.Error(),
		}
		time.Sleep(time.Millisecond * 1000 * time.Duration(i))
		result_ = Act_length_query(id, i+1)
	}

	return result_
}

func Act_usage_query(id string, i int) []*neo4j.Record {

	statement := `MATCH(v:Asset{id:$VEH})-[:STOPPED_AT]->(s:Stop)
			      			   WHERE s.end_timedt.year = $YEAR
			      			   AND s.end_time - s.start_time > $MINDUR
			      			   RETURN v.id as vehicle,
			      			    s.start_time as start_time,
			      			    s.end_time as end_time,
			      			    s.start_timedt.month as month,
			      			    s.start_timedt.day as day,
			      			    s.sa2 as sa2

			      			   `
	fabric_prefix := "UNWIND graph.names() AS g CALL {USE " + "graph.byName(g)  "
	fabric_suffix := "} RETURN vehicle, start_time, end_time, month, day, sa2"
	//fmt.Println(id)
	//fmt.Println(Year)
	//fmt.Println(MinDur)
	statement = fabric_prefix + statement + fabric_suffix
	//fmt.Println(statement)
	parameters := map[string]interface{}{
		"VEH":    id,
		"YEAR":   Year,
		"MINDUR": MinDur,
	}
	//fmt.Println(parameters)
	session := Db.NewSession(Sesh_config)

	defer session.Close()

	result, err := session.Run(statement, parameters)
	result_, _ := result.Collect()
	if err != nil {
		fmt.Println(err)
		err_c <- []string{
			id,
			err.Error(),
		}
		time.Sleep(time.Millisecond * 1000 * time.Duration(i))
		result_ = Act_usage_query(id, i+1)
	}
	if result.Err() != nil {
		fmt.Println(result.Err())
		return nil
	}

	return result_
}

func Act_write(filename string, resume bool) {

	var ids []string
	var process_ids []string

	//idstruct := VehicleIds{}

	fmt.Println("Getting vehicle list")
	session := Db.NewSession(Sesh_config)

	defer session.Close()

	fabric_prefix := "USE fabric UNWIND graph.names() AS g CALL {USE graph.byName(g) "
	fabric_suffix := "} RETURN distinct(id)"
	id_q := "MATCH(v:Asset) RETURN v.id as id"
	id_q = fabric_prefix + id_q + fabric_suffix
	fmt.Println(id_q)
	idquery, err := session.Run(id_q, map[string]interface{}{})

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Building vehicle list")
	for idquery.Next() {
		ids = append(ids, idquery.Record().GetByIndex(0).(string))
	}

	//ids = ids
	//fmt.Println(ids[0])
	file, fileerr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	//file, fileerr := os.Create(filename)
	if fileerr != nil {
		fmt.Printf("File error: %s", fileerr)
	}

	if resume {

		file, _ = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
		//defer file.Close()
		fmt.Println("Finding undone vehicles")
		csvr := csv.NewReader(file)
		var done []string
		last := ""
		for {
			line, csverr := csvr.Read()
			if csverr == io.EOF {
				break
			} else {
				if line[0] != last {
					done = append(done, line[0])
					last = line[0]
				}
			}
		}

		var wg_contains sync.WaitGroup
		c_undone := make(chan string)

		for _, id := range ids {
			wg_contains.Add(1)
			ext_undone(done, id, c_undone, func() { wg_contains.Done() })
		}

		go func() {
			defer close(c_undone)
			wg_contains.Wait()
		}()
		fmt.Println("Channel closed")
		for j := range c_undone {
			process_ids = append(process_ids, j)
			//fmt.Println(j)
		}

		fmt.Println("Resuming query")

	} else {

		file.Truncate(0)
		file.Seek(0, 0)
		process_ids = ids
		sort.Strings(process_ids)
		//defer file.Close()
	}

	defer file.Close()

	c := make(chan []string)
	err_c = make(chan []string)
	//defer close(c)
	var wg sync.WaitGroup

	length := len(process_ids)
	fmt.Printf("Processing %d vehicles\n", length)
	guard := make(chan struct{}, Max_routines)
	//wg.Add(length)
	last := ""
	var headers []string
	if Act_type == "usage" {
		headers = []string{
			"vehicle",
			"start_time",
			"end_time",
			"month",
			"day",
			"sa2",
		}
	}
	if Act_type == "length" {
		headers = []string{
			"vehicle",
			"sum_length",
			"month",
			"day",
			"state",
			"n_seg",
		}
		//fmt.Println(Month)
	}

	w := csv.NewWriter(file)

	if !resume {
		w.Write(headers)
	}
	w.Flush()

	file_err, _ := os.OpenFile("vehicle errors", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	w_err := csv.NewWriter(file_err)

	go act_writer(w, c)
	go error_writer(w_err, err_c)

	for i, id := range process_ids {
		if id != last {

			wg.Add(1)
			guard <- struct{}{} //blocks until space in the channel opens
			//fmt.Println(id)
			go func(batch_id string, c chan []string, guard chan struct{}, ind int) {
				if Act_type == "usage" {
					//fmt.Println(batch_id)
					batch_query_usage(batch_id, c, func() { wg.Done() }, ind)
				} else {
					batch_query_length(batch_id, c, func() { wg.Done() }, ind)
				}
				<-guard
			}(id, c, guard, i)

			last = id
		} else {
			fmt.Printf("Dupe with %s\n", id)
		}
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()
	wg.Wait()
}
