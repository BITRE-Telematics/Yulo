package queries

import (
	"encoding/csv"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	//"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Seg_speedquery_write(osm_id string, i int, Bd_type string, direction bool, onExit func()) {
	defer onExit()
	//on database restucture change to "count(a) as n_obvs and sum(size([x IN oa.type WHERE x <> 'imputed'])) as rec_obvs"
	statement := `
                 MATCH (s:Segment{osm_id: $OSM_ID})<-[oa:ON]-(o:Observation)<-[:OBSERVED_AT]-(t:Trip)<-[:EMBARKED_ON]-(v:Vehicle)
	              WHERE o.datetime > $START AND o.datetime < $FINISH AND o.datetimedt IS NOT NULL
	              WITH s, o, t, v, oa
                  RETURN  
                  s.osm_id as osm_id,
                  percentileCont(o.imputed_speed, 0.25) as LQ_imp,
                  percentileCont(o.imputed_speed, 0.5) as Median_imp,
                  percentileCont(o.imputed_speed, 0.75) as UQ_imp,
                  stDev(o.imputed_speed) as stDev_imp,
		          count(o) as n_obvs,
                  count(distinct(t)) as n_trips,
                  count(distinct(v)) as n_vehicles
                 `
	if Bd_type != "" {
		statement = statement + fmt.Sprintf(", o.datetimedt.%[1]s as %[1]s", Bd_type)
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
	}

	if Bd_type != "" {
		statement = statement + fmt.Sprintf(" ORDER BY o.datetimedt.%[1]s", Bd_type)
	}

	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	parameters := map[string]interface{}{"OSM_ID": osm_id, "START": Start, "FINISH": Finish}

	results, err := session.Run(statement, parameters)
	if err != nil && i < 1000 {
		if i%50 == 0 {
			fmt.Println(err, osm_id, i)
		}
		time.Sleep(time.Millisecond * 1000 * time.Duration(i))
		results = Seg_speedquery(osm_id, i+1)
	}
	if results.Err() != nil {
		fmt.Println(results.Err())
	}
	//fmt.Print(osm_id, ":", res)
	//fmt.Print(res.N)
	if direction {
		batch_query_write_dir(results, Bd_type)
	} else {
		batch_query_write_nodir(results, Bd_type)
	}
}

func batch_query_write_nodir(seg_data neo4j.Result, Bd_type string) {

	//defer onExit()
	min_bd, max_bd := get_min_max_bd(Bd_type)
	len_bd := max_bd - min_bd + 1
	lq_imp := fill_negs_float(make([]float64, len_bd))
	median_imp := fill_negs_float(make([]float64, len_bd))
	uq_imp := fill_negs_float(make([]float64, len_bd))
	n_trips := fill_negs_int(make([]int64, len_bd))
	n_vehicles := fill_negs_int(make([]int64, len_bd))
	osm_id := ""
	//index in slive with i - min_db
	for seg_data.Next() {
		var r int64
		if v_bd, ok := seg_data.Record().Get(Bd_type); ok {
			r = v_bd.(int64)
		}
		i := r - min_bd
		//fmt.Println(seg_data.Record().Values())
		osm_id = seg_data.Record().GetByIndex(0).(string) //osm_id
		lq_imp[i] = floatGet(seg_data.Record(), "LQ_imp")
		median_imp[i] = floatGet(seg_data.Record(), "Median_imp")
		uq_imp[i] = floatGet(seg_data.Record(), "UQ_imp")

		n_trips[i] = intGet(seg_data.Record(), "n_trips") //trips

		n_vehicles[i] = intGet(seg_data.Record(), "n_vehicles") //vehicles

	}

	//write to database
	data := map[string]interface{}{
		"OSM_ID":                             osm_id,
		fmt.Sprintf("LQ_IMP%s", Bd_type):     lq_imp,
		fmt.Sprintf("UQ_IMP%s", Bd_type):     uq_imp,
		fmt.Sprintf("MED_IMP%s", Bd_type):    median_imp,
		fmt.Sprintf("N_TRIPS%s", Bd_type):    n_trips,
		fmt.Sprintf("N_VEHICLES%s", Bd_type): n_vehicles,
	}
	write_segs_db(data, Bd_type, false)
}

func batch_query_write_dir(seg_data neo4j.Result, Bd_type string) {

	//defer onExit()
	min_bd, max_bd := get_min_max_bd(Bd_type)
	len_bd := max_bd - min_bd + 1

	lq_imp_fw := fill_negs_float(make([]float64, len_bd))
	median_imp_fw := fill_negs_float(make([]float64, len_bd))
	uq_imp_fw := fill_negs_float(make([]float64, len_bd))
	n_trips_fw := fill_negs_int(make([]int64, len_bd))
	n_vehicles_fw := fill_negs_int(make([]int64, len_bd))

	lq_imp_bw := fill_negs_float(make([]float64, len_bd))
	median_imp_bw := fill_negs_float(make([]float64, len_bd))
	uq_imp_bw := fill_negs_float(make([]float64, len_bd))
	n_trips_bw := fill_negs_int(make([]int64, len_bd))
	n_vehicles_bw := fill_negs_int(make([]int64, len_bd))

	osm_id := ""
	//index in slive with i - min_db
	for seg_data.Next() {
		var r int64
		var forward bool
		if v_bd, ok := seg_data.Record().Get(Bd_type); ok {
			r = v_bd.(int64)
		}
		i := r - min_bd
		if v_for, ok := seg_data.Record().Get("forward"); ok {
			if v_for == nil {
				continue
			}
			forward = v_for.(bool)
		}
		//fmt.Println(seg_data.Record().Values())
		osm_id = seg_data.Record().GetByIndex(0).(string) //osm_id

		if forward {
			lq_imp_fw[i] = floatGet(seg_data.Record(), "LQ_imp")
			median_imp_fw[i] = floatGet(seg_data.Record(), "Median_imp")
			uq_imp_fw[i] = floatGet(seg_data.Record(), "UQ_imp")
			n_trips_fw[i] = intGet(seg_data.Record(), "n_trips")       //trips
			n_vehicles_fw[i] = intGet(seg_data.Record(), "n_vehicles") //vehicles
		} else {
			lq_imp_bw[i] = floatGet(seg_data.Record(), "LQ_imp")
			median_imp_bw[i] = floatGet(seg_data.Record(), "Median_imp")
			uq_imp_bw[i] = floatGet(seg_data.Record(), "UQ_imp")
			n_trips_bw[i] = intGet(seg_data.Record(), "n_trips")       //trips
			n_vehicles_bw[i] = intGet(seg_data.Record(), "n_vehicles") //vehicles
		}

	}
	data := map[string]interface{}{
		"OSM_ID":                                osm_id,
		fmt.Sprintf("LQ_IMP%s_fw", Bd_type):     lq_imp_fw,
		fmt.Sprintf("UQ_IMP%s_fw", Bd_type):     uq_imp_fw,
		fmt.Sprintf("MED_IMP%s_fw", Bd_type):    median_imp_fw,
		fmt.Sprintf("N_TRIPS%s_fw", Bd_type):    n_trips_fw,
		fmt.Sprintf("N_VEHICLES%s_fw", Bd_type): n_vehicles_fw,
		fmt.Sprintf("LQ_IMP%s_bw", Bd_type):     lq_imp_bw,
		fmt.Sprintf("UQ_IMP%s_bw", Bd_type):     uq_imp_bw,
		fmt.Sprintf("MED_IMP%s_bw", Bd_type):    median_imp_bw,
		fmt.Sprintf("N_TRIPS%s_bw", Bd_type):    n_trips_bw,
		fmt.Sprintf("N_VEHICLES%s_bw", Bd_type): n_vehicles_bw,
	}
	//fmt.Println(data)
	write_segs_db(data, Bd_type, true)
	fmt.Printf("%s at %s\n", osm_id, time.Now())
}

func write_segs_db(data map[string]interface{}, Bd_type string, direction bool) {
	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()
	var statement string
	if !direction {
		statement = `
		MATCH(s:Segment{osm_id: $OSM_ID})
		SET
		s.lq_imp%[1]s = $LQ_IMP%[1]s,
		s.uq_imp%[1]s = $UQ_IMP%[1]s,
		s.med_imp%[1]s = $MED_IMP%[1]s,
		s.n_trips%[1]s = $N_TRIPS%[1]s,
		s.n_vehicles%[1]s = $N_VEHICLES%[1]s,
		s.updated = date()
		return s.osm_id
		`
		statement = fmt.Sprintf(statement, Bd_type)
		//fmt.Println(statement)
		//fmt.Println(data)
		res, err := session.Run(statement, data)
		//fmt.Println(res.Err())
		if err != nil {
			fmt.Println(err)
		}
		if res.Err() != nil {
			fmt.Println(err)
		}
	} else {
		statement = `
		MATCH(s:Segment{osm_id: $OSM_ID})
		SET
		s.lq_imp%[1]s_fw = $LQ_IMP%[1]s_fw,
		s.uq_imp%[1]s_fw = $UQ_IMP%[1]s_fw,
		s.med_imp%[1]s_fw = $MED_IMP%[1]s_fw,
		s.n_trips%[1]s_fw = $N_TRIPS%[1]s_fw,
		s.n_vehicles%[1]s_fw = $N_VEHICLES%[1]s_fw,
		s.lq_imp%[1]s_bw = $LQ_IMP%[1]s_bw,
		s.uq_imp%[1]s_bw = $UQ_IMP%[1]s_bw,
		s.med_imp%[1]s_bw = $MED_IMP%[1]s_bw,
		s.n_trips%[1]s_bw = $N_TRIPS%[1]s_bw,
		s.n_vehicles%[1]s_bw = $N_VEHICLES%[1]s_bw,
		s.updated = date()
		return s.osm_id
		`
		statement = fmt.Sprintf(statement, Bd_type)
		//fmt.Println(statement)
		//fmt.Println(data)
		res, err := session.Run(statement, data)
		//fmt.Println(res.Err())
		if err != nil {
			fmt.Println(err)
		}
		if res.Err() != nil {
			fmt.Println(res.Err())
		}
	}

}

func full_query_csv(batch_id string, c chan []string, onExit func()) {

	defer onExit()
	seg_data := Seg_speedquery_full(batch_id)
	for seg_data.Next() {
		//replace using keys, account for nil values
		//fmt.Println(seg_data.Record().Values())
		lq_imp := floatFrmt(seg_data.Record(), "LQ_imp")
		med_imp := floatFrmt(seg_data.Record(), "Median_imp")
		uq_imp := floatFrmt(seg_data.Record(), "UQ_imp")
		stDev_imp := floatFrmt(seg_data.Record(), "stDev_imp")

		n_trips := intGet(seg_data.Record(), "n_trips")
		n_veh := intGet(seg_data.Record(), "n_vehicles")
		n_obvs := intGet(seg_data.Record(), "n_obvs")

		l := []string{
			seg_data.Record().GetByIndex(0).(string), //osm_id
			strconv.FormatInt(n_obvs, 10),            //obvs
			strconv.FormatInt(n_trips, 10),           //trips
			strconv.FormatInt(n_veh, 10),             //vehicles
			lq_imp,                                   //lq_imp
			med_imp,                                  //med_imp
			uq_imp,                                   //uq_imp
			stDev_imp,                                //st_dev_imp
		}

		//fmt.Println(l)
		c <- l

	}

}

func Seg_speedquery_full(osm_id string) neo4j.Result {
	//on database restucture change to "count(a) as n_obvs and sum(size([x IN oa.type WHERE x <> 'imputed'])) as rec_obvs"
	statement := `
                 MATCH (s:Segment{osm_id: $OSM_ID})<-[oa:ON]-(o:Observation)<-[OBSERVED_AT]-(t:Trip)<-[:EMBARKED_ON]-(v:Vehicle)
	              WHERE o.datetime > $START AND o.datetime < $FINISH AND o.datetimedt IS NOT NULL
	              WITH s, o, t, v, oa
                  RETURN  
                  s.osm_id as osm_id,
                  percentileCont(o.imputed_speed, 0.25) as LQ_imp,
                  percentileCont(o.imputed_speed, 0.5) as Median_imp,
                  percentileCont(o.imputed_speed, 0.75) as UQ_imp,
                  stDev(o.imputed_speed) as stDev_imp,
		          count(o) as n_obvs,
                  count(distinct(t)) as n_trips,
                  count(distinct(v)) as n_vehicles
                 `

	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	parameters := map[string]interface{}{"OSM_ID": osm_id, "START": Start, "FINISH": Finish}

	results, err := session.Run(statement, parameters)
	if err != nil {

		fmt.Println(err, osm_id)

	}
	//fmt.Print(osm_id, ":", res)
	if results.Err() != nil {
		fmt.Println(results.Err())
	}
	return results
}

func Seg_write_db(filename string, resume bool, speedfile_nm string) {

	var osm_ids []string
	var process_osm_ids []string

	fmt.Println("Getting segment list")

	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	results, err := session.Run("MATCH(s:Segment) WHERE s.osm_id <> 'nan' RETURN s.osm_id as osm_id ", map[string]interface{}{})

	if err != nil {
		fmt.Printf("Id query error %s", err)
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
		fmt.Printf("File error: %s", fileerr)
	}

	speedfile, fileerr := os.OpenFile(speedfile_nm, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	//file, fileerr := os.Create(filename)
	if fileerr != nil {
		fmt.Printf("File error: %s", fileerr)
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
		w.Write(headers_speeds_csv)
	}
	w_speeds_csv.Flush()

	c_speeds_csv := make(chan []string)
	//defer close(c)
	var wg sync.WaitGroup
	go seg_writer(w_speeds_csv, c_speeds_csv)

	length := len(process_osm_ids)
	fmt.Printf("Processing %d segments\n", length)
	guard := make(chan struct{}, 10)
	//wg.Add(length)
	last := ""

	for _, id := range process_osm_ids {
		if id != last {

			wg.Add(7)
			guard <- struct{}{} //blocks until space in the channel opens
			//fmt.Println(id)
			go func(batch_id string, c chan []string, c_csv chan []string, guard chan struct{}) {
				//writes to file for tiles
				full_query_csv(batch_id, c_csv, func() { wg.Done() })
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
			}(id, c, c_speeds_csv, guard)
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
