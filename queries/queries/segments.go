package queries

import (
	"encoding/csv"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Start int64
var Finish int64
var Breakdown bool
var Byfirm bool
var Bd_type string
var Direction bool

func batch_query(batch_id string, c chan []string, onExit func(), ind int) {

	defer onExit()
	seg_data := Seg_speedquery(batch_id, 1)
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

		if Breakdown {
			if v_bd, ok := seg_data.Record().Get(Bd_type); ok {
				l = append(l, strconv.FormatInt(v_bd.(int64), 10))
			}
		}
		var forward bool
		var dir float64
		if Direction {
			if v_for, ok := seg_data.Record().Get("forward"); ok {

				forward = v_for.(bool)
				l = append(l, strconv.FormatBool(forward))

			}
			if v_dir, ok := seg_data.Record().Get("direction"); ok {
				if v_dir != nil {
					dir = v_dir.(float64)
					if !forward {
						dir = dir - 180
						if dir < 1 {
							dir = 360 - math.Abs(dir)
						}
					}
					l = append(l, strconv.FormatFloat(dir, 'f', 2, 64))
				}
			} else {
				l = append(l, "NA")
			}
		}

		if Byfirm {
			if v_firm, ok := seg_data.Record().Get("firm"); ok {
				l = append(l, v_firm.(string))
			}

		}
		//fmt.Println(l)
		c <- l

	}
	if ind%1000 == 0 {
		fmt.Println(ind)
	}

}

func Seg_speedquery(osm_id string, i int) neo4j.Result {
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
	if Breakdown {
		statement = statement + fmt.Sprintf(", o.datetimedt.%[1]s as %[1]s", Bd_type)
	}
	//if restructured change to oa.azimuth
	if Direction {
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

	if Byfirm {
		statement = statement + ", v.firm as firm"

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
	//fmt.Print(osm_id, ":", res)
	if results.Err() != nil {
		fmt.Println(results.Err())
	}
	return results
}

func Seg_write(filename string, resume bool) {

	var osm_ids []string
	var process_osm_ids []string

	if Breakdown {
		fmt.Println("The breakdown flag is true")
	}

	fmt.Println("Getting segment list")

	session, err := session := Db.NewSession(Sesh_config)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer session.Close()

	seg_results, err := session.Run("MATCH(s:Segment) WHERE s.osm_id <> 'nan' RETURN s.osm_id as osm_id", map[string]interface{}{})

	if err != nil {
		fmt.Printf("Id query error %s", err)
	}
	fmt.Println("Building segment list")
	for seg_results.Next() {
		osm_ids = append(osm_ids, seg_results.Record().GetByIndex(0).(string))
	}
	//osm_ids = osm_ids

	file, fileerr := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
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
		process_osm_ids = osm_ids
		sort.Strings(process_osm_ids)
		//defer file.Close()
	}

	defer file.Close()

	headers := []string{
		"osm_id",
		"n_obvs",
		"n_trips",
		"n_vehicles",
		"LQ_imp",
		"median_imp",
		"UQ_imp",
		"stDev_imp",
	}

	if Breakdown {
		headers = append(headers, Bd_type)
	}

	if Direction {
		headers = append(headers, "forward")
		headers = append(headers, "direction")
	}

	if Byfirm {
		headers = append(headers, "firm")
	}

	w := csv.NewWriter(file)
	if !resume {
		w.Write(headers)
	}
	w.Flush()

	c := make(chan []string)
	//defer close(c)
	var wg sync.WaitGroup
	go seg_writer(w, c)
	length := len(process_osm_ids)
	fmt.Printf("Processing %d segments\n", length)
	guard := make(chan struct{}, 10)
	//wg.Add(length)
	last := ""
	for i, id := range process_osm_ids {
		if id != last {

			wg.Add(1)
			guard <- struct{}{} //blocks until space in the channel opens
			//fmt.Println(id)
			go func(batch_id string, c chan []string, guard chan struct{}, ind int) {
				batch_query(id, c, func() { wg.Done() }, ind)
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
