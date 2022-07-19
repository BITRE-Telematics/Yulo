package queries

import (
	"encoding/csv"
	"fmt"
	//"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	//"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func bd_query_builder(Bd_type string, direction bool) string {

	statement := `
                 MATCH (s:Segment{osm_id: $OSM_ID})<-[oa:ON]-(o:Observation)<-[:OBSERVED_AT]-(t:Trip)<-[:EMBARKED_ON]-(v:Asset)
	              WHERE o.datetime > $START AND o.datetime < $FINISH AND o.datetimedt IS NOT NULL
	              WITH s, o, t, v, oa
                  RETURN  
                  s.osm_id as osm_id,
                  o.imputed_speed as imputed_speed,
		          count(o) as n_obvs,
                  count(distinct(t)) as n_trips,
                  count(distinct(v)) as n_vehicles
                 `
	fabric_prefix := "UNWIND " + Fabric + ".graphIds() AS graphId CALL {USE " + Fabric + ".graph(graphId) "
	fabric_suffix := `} RETURN osm_id,
							   percentileCont(imputed_speed, 0.25) as LQ_imp,
							   percentileCont(imputed_speed, 0.50) as Median_imp,
							   percentileCont(imputed_speed, 0.75) as UQ_imp,
							   stDev(imputed_speed) as stDev_imp,
							   sum(n_obvs) as n_obvs,
							   sum(n_trips) as n_trips,
							   sum(n_vehicles) as n_vehicles`
	if Bd_type != "" {
		statement = statement + fmt.Sprintf(`, o.datetimedt.%[1]s as %[1]s`, Bd_type)
		fabric_suffix = fabric_suffix + ", " + Bd_type
	}
	//if restructured change to oa.azimuth
	if direction {
		statement = statement +
			`
	                        , s.forward as direction,
                             CASE 
                             WHEN o.forward IS NOT NULL
                             THEN o.forward
                             WHEN oa.forward IS NULL
                             THEN abs(s.forward - o.azimuth) < 90 OR abs(s.forward - o.azimuth) > 270
                             ELSE 
                             	oa.forward 
                             	END 
                             as forward
                             `
		statement = strings.Replace(statement, "RETURN", "WHERE o.azimuth <> 0 RETURN", 1) // otherwise null values wil each get a separate row
		fabric_suffix = fabric_suffix + ", forward"
	}

	if Bd_type != "" {
		statement = statement + fmt.Sprintf(" ORDER BY %[1]s", Bd_type)
	}
	//fmt.Println("starting new session")

	statement = fabric_prefix + statement + fabric_suffix

	return statement
}

func Seg_speedquery_write(osm_id string, i int, Bd_type string, direction bool, onExit func()) {
	fmt.Sprintf("%s for %s\n", Bd_type, osm_id)
	defer onExit()

	session := Db.NewSession(Sesh_config)

	defer session.Close()

	statement := bd_query_builder(Bd_type, direction)
	//fmt.Println(statement)
	parameters := map[string]interface{}{"OSM_ID": osm_id, "START": Start, "FINISH": Finish}

	results, err := session.Run(statement, parameters)
	if err != nil {

		fmt.Println(err, osm_id, i)
		return
	}
	if results.Err() != nil {
		fmt.Println(results.Err())
	}

	results_, _ := results.Collect()

	if direction {
		batch_query_write_dir(results_, Bd_type)
		//fmt.Println("debug filler")
	} else {
		batch_query_write_nodir(results_, Bd_type)

	}

}

func full_query_csv(batch_id string, c chan []string, onExit func()) {

	defer onExit()
	l := Seg_speedquery_full(batch_id)

	c <- l

}

func Seg_speedquery_full(osm_id string) []string {
	//on database restucture change to "count(a) as n_obvs and sum(size([x IN oa.type WHERE x <> 'imputed'])) as rec_obvs"
	statement := `
                 MATCH (s:Segment{osm_id: $OSM_ID})<-[oa:ON]-(o:Observation)<-[OBSERVED_AT]-(t:Trip)<-[:EMBARKED_ON]-(v:Asset)
	              WHERE o.datetime > $START AND o.datetime < $FINISH AND o.datetimedt IS NOT NULL
	              WITH s, o, t, v, oa
                  RETURN  
                  s.osm_id as osm_id,
                  o.imputed_speed as imputed_speed,
                  stDev(o.imputed_speed) as stDev_imp,
		          count(o) as n_obvs,
                  count(distinct(t)) as n_trips,
                  count(distinct(v)) as n_vehicles
                 `
	fabric_prefix := "UNWIND " + Fabric + ".graphIds() AS graphId CALL {USE " + Fabric + ".graph(graphId) "
	fabric_suffix := `
						} RETURN osm_id,
						percentileCont(imputed_speed, 0.25) as LQ_imp,
						percentileCont(imputed_speed, 0.50) as Median_imp,
						percentileCont(imputed_speed, 0.75)as UQ_imp,
						stDev(imputed_speed) as stDev_imp,
						sum(n_obvs) as n_obvs,
						sum(n_trips) as n_trips,
						sum(n_vehicles) as n_vehicles`
	statement = fabric_prefix + statement + fabric_suffix
	session := Db.NewSession(Sesh_config)

	defer session.Close()

	parameters := map[string]interface{}{"OSM_ID": osm_id, "START": Start, "FINISH": Finish}

	results, err := session.Run(statement, parameters)
	l := []string{}
	if err != nil {

		fmt.Println(err, osm_id)

	}
	if results.Err() != nil {
		fmt.Println(results.Err())
	}
	for results.Next() {
		//replace using keys, account for nil values
		//fmt.Println(results.Record().Values())
		lq_imp := floatFrmt(results.Record(), "LQ_imp")
		med_imp := floatFrmt(results.Record(), "Median_imp")
		uq_imp := floatFrmt(results.Record(), "UQ_imp")
		stDev_imp := floatFrmt(results.Record(), "stDev_imp")

		n_trips := intGet(results.Record(), "n_trips")
		n_veh := intGet(results.Record(), "n_vehicles")
		n_obvs := intGet(results.Record(), "n_obvs")

		l = []string{
			results.Record().GetByIndex(0).(string), //osm_id
			strconv.FormatInt(n_obvs, 10),           //obvs
			strconv.FormatInt(n_trips, 10),          //trips
			strconv.FormatInt(n_veh, 10),            //vehicles
			lq_imp,                                  //lq_imp
			med_imp,                                 //med_imp
			uq_imp,                                  //uq_imp
			stDev_imp,                               //st_dev_imp
		}
		//fmt.Println("printing l")
		//fmt.Println(l)
		return l
	}
	return l
}

func Seg_write_db(filename string, resume bool, speedfile_nm string) {

	var osm_ids []string
	var process_osm_ids []string

	fmt.Println("Getting segment list")

	session := Db.NewSession(Sesh_config)

	defer session.Close()

	seg_q := fmt.Sprintf("USE %s MATCH(s:Segment) WHERE s.osm_id <> 'nan' RETURN s.osm_id as osm_id ", Seg_db)

	results, err := session.Run(seg_q, map[string]interface{}{})

	if err != nil {
		fmt.Printf("Id query error %s\n", err)
		return
	}
	if results.Err() != nil {
		fmt.Println(results.Err())
	}
	fmt.Println("Building segment list")
	for results.Next() {
		osm_ids = append(osm_ids, results.Record().GetByIndex(0).(string))
	}
	//osm_ids = osm_ids

	file, fileerr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	//file, fileerr := os.Create(filename)
	if fileerr != nil {
		fmt.Printf("File error: %s\n", fileerr)
	}

	speedfile, fileerr := os.OpenFile(speedfile_nm, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	//file, fileerr := os.Create(filename)
	if fileerr != nil {
		fmt.Printf("File error: %s\n", fileerr)
	}

	if resume {

		//file, _ := os.OpenFile(filename, os.O_RDWR, 0755)
		//defer file.Close()
		fmt.Println("Finding undone osm_ids")
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

		for _, osm_id := range osm_ids {
			wg_contains.Add(1)
			ext_undone(done, osm_id, c_undone, func() { wg_contains.Done() })
		}

		go func() {
			defer close(c_undone)
			wg_contains.Wait()
		}()
		fmt.Println("Channel closed")
		for j := range c_undone {
			process_osm_ids = append(process_osm_ids, j)
			//fmt.Println(j)
		}

		fmt.Println("Resuming query")

	} else {

		file.Truncate(0)
		file.Seek(0, 0)
		speedfile.Truncate(0)
		speedfile.Seek(0, 0)
		process_osm_ids = osm_ids
		sort.Strings(process_osm_ids)
		//defer file.Close()
	}

	defer file.Close()

	headers := []string{
		"osm_id",
	}

	w := csv.NewWriter(file)
	if !resume {
		w.Write(headers)
	}
	w.Flush()

	c := make(chan []string)
	//defer close(c)

	go seg_writer(w, c)

	headers_speeds_csv := []string{
		"osm_id",
		"n_obvs",
		"n_trips",
		"n_vehicles",
		"LQ_imp",
		"median_imp",
		"UQ_imp",
		"stDev_imp",
	}

	w_speeds_csv := csv.NewWriter(speedfile)
	if !resume {
		w_speeds_csv.Write(headers_speeds_csv)
	}
	w_speeds_csv.Flush()

	c_speeds_csv := make(chan []string)
	//defer close(c)
	var wg sync.WaitGroup
	go seg_writer(w_speeds_csv, c_speeds_csv)

	length := len(process_osm_ids)
	fmt.Printf("Processing %d segments\n", length)
	//fmt.Println(Max_routines)
	guard := make(chan struct{}, Max_routines)
	//wg.Add(length)
	last := ""

	for i, id := range process_osm_ids {
		if id != last {
			//fmt.Println(id)
			wg.Add(7)
			guard <- struct{}{} //blocks until space in the channel opens
			fmt.Println(id)

			go func(batch_id string, c chan []string, c_csv chan []string, guard chan struct{}) {
				start := time.Now()
				//writes to file for tiles
				//fmt.Println("starting full query")
				full_query_csv(batch_id, c_csv, func() { wg.Done() })
				//fmt.Println("done full query")
				//writes precomputed values to database
				//direction
				Seg_speedquery_write(batch_id, 1, "dayOfWeek", true, func() { wg.Done() })
				Seg_speedquery_write(batch_id, 1, "hour", true, func() { wg.Done() })
				Seg_speedquery_write(batch_id, 1, "month", true, func() { wg.Done() })
				//nodirection
				Seg_speedquery_write(batch_id, 1, "dayOfWeek", false, func() { wg.Done() })
				Seg_speedquery_write(batch_id, 1, "hour", false, func() { wg.Done() })
				Seg_speedquery_write(batch_id, 1, "month", false, func() { wg.Done() })

				id_slice := make([]string, 1)
				id_slice[0] = batch_id
				c <- id_slice

				<-guard
				fmt.Printf("%s in %s\n", id, time.Since(start).String())
			}(id, c, c_speeds_csv, guard)
			last = id

		} else {
			fmt.Printf("Dupe with %s at %s\n", id, i)
		}
	}
	go func() {
		defer close(c)
		wg.Wait()
	}()
	wg.Wait()

}
