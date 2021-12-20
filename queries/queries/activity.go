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

var Activityfile string
var Year int64
var MinDur int64
var Month int64
var Act_type string
var err_c chan []string
var Db neo4j.Driver
var Sesh_config neo4j.SessionConfig

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
	for act_data.Next() {
		//wg.Add(1)

		l := []string{
			stringGet(act_data.Record(), "vehicle"), //vehicle
			unspecifiedNumFrmt(act_data.Record(), "sum_length"),
			strconv.FormatInt(intGet(act_data.Record(), "month"), 10), //month
			strconv.FormatInt(intGet(act_data.Record(), "day"), 10),   //day
			stringGet(act_data.Record(), "state"),                     //state
		}

		c <- l

	}
	if ind%100 == 0 {
		fmt.Println(ind)
	}

}

func batch_query_usage(batch_id string, c chan []string, onExit func(), ind int) {

	defer onExit()
	act_data := Act_usage_query(batch_id, 1)
	for act_data.Next() {
		//wg.Add(1)
		l := []string{
			stringGet(act_data.Record(), "vehicle"),                        //vehicle
			strconv.FormatInt(intGet(act_data.Record(), "start_time"), 10), //start_time
			strconv.FormatInt(intGet(act_data.Record(), "end_time"), 10),   //end_time
			strconv.FormatInt(intGet(act_data.Record(), "month"), 10),      //month
			strconv.FormatInt(intGet(act_data.Record(), "day"), 10),        //day
			stringGet(act_data.Record(), "sa2"),                            //sa2
		}

		c <- l

	}
	if ind%100 == 0 {
		fmt.Println(ind)
	}

}

func Act_length_query(id string, i int) neo4j.Result {
	//res := ActLength{}

	statement := `MATCH(v:Vehicle{id:$VEH})-[:EMBARKED_ON]->(Trip)-[:OBSERVED_AT]->(o:Observation)-[on:ON]->(s:Segment)
       			  WHERE (on.type <> 'imputed' OR on.type IS NULL) AND o.datetimedt.year = $YEAR %s 
      			  RETURN
      			   v.id as vehicle,
      			   sum(o.length) as sum_length,
      			   o.datetimedt.month as month,
      			   o.datetimedt.day as day,
      			   left(s.sa2, 1) as state
                 `
	var parameters map[string]interface{}
	if (Month > 0) && (Month < 13) {
		//str_month := strconv.FormatInt(Month, 10)
		statement = fmt.Sprintf(statement, "AND o.datetimedt.month = $MONTH")

		parameters = map[string]interface{}{"VEH": id, "YEAR": Year, "MONTH": Month}

	} else {
		statement = fmt.Sprintf(statement, "")
		parameters = map[string]interface{}{"VEH": id, "YEAR": Year, "MONTH": Month}

	}
	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}

	defer session.Close()
	result, err := session.Run(statement, parameters)
	if err != nil && i < 5 {
		if (i%5 == 0) || (i == 0) {
			fmt.Println(err, id, i)
			err_c <- []string{
				id,
				err.Error(),
			}
		}
		time.Sleep(time.Millisecond * 1000 * time.Duration(i))
		result = Act_length_query(id, i+1)
	}
	if result.Err() != nil {
		fmt.Println(result.Err())
		return nil
	}
	//fmt.Print(id, ":", res)
	//fmt.Print(res.N)
	return result
}

func Act_usage_query(id string, i int) neo4j.Result {

	statement := `MATCH(v:Vehicle{id: $veh })-[:STOPPED_AT]->(s:Stop)
			      			   WHERE s.end_timedt.year = $year
			      			   AND s.end_time - s.start_time > $mindur
			      			   RETURN v.id as vehicle,
			      			    s.start_time as start_time,
			      			    s.end_time as end_time,
			      			    s.start_timedt.month as month,
			      			    s.start_timedt.day as day,
			      			    s.sa2 as sa2

			      			   `

	//fmt.Println(id)
	//fmt.Println(Year)
	//fmt.Println(MinDur)
	parameters := map[string]interface{}{
		"veh":    id,
		"year":   Year,
		"mindur": MinDur,
	}
	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}

	defer session.Close()

	result, err := session.Run(statement, parameters)

	if err != nil && i < 5 {
		if (i%5 == 0) || (i == 0) {
			fmt.Println(err, id, i)
			err_c <- []string{
				id,
				err.Error(),
			}
		}
		time.Sleep(time.Millisecond * 1000 * time.Duration(i))
		result = Act_usage_query(id, i+1)
	}
	if result.Err() != nil {
		fmt.Println(result.Err())
		return nil
	}
	return result
}

func Act_write(filename string, resume bool) {

	var ids []string
	var process_ids []string

	//idstruct := VehicleIds{}

	fmt.Println("Getting vehicle list")
	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	idquery, err := session.Run("MATCH(v:Vehicle) RETURN v.id as id", map[string]interface{}{})

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

		//file, _ := os.OpenFile(filename, os.O_RDWR, 0755)
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
	guard := make(chan struct{}, 10)
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
