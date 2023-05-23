package queries

import (
	//"encoding/csv"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	//"io"
	//"math"
	//"os"
	//"sort"
	//"strconv"
	//"strings"
	//"sync"
	//"time"
)

func batch_query_write_nodir(seg_data []*neo4j.Record, Bd_type string) {

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
	for _, rec := range seg_data {
		var r int64
		if v_bd, ok := rec.Get(Bd_type); ok {
			r = v_bd.(int64)
		}
		i := r - min_bd
		//fmt.Println(rec.Values())
		osm_id = rec.GetByIndex(0).(string) //osm_id
		lq_imp[i] = floatGet(rec, "LQ_imp")
		median_imp[i] = floatGet(rec, "Median_imp")
		uq_imp[i] = floatGet(rec, "UQ_imp")

		n_trips[i] = intGet(rec, "n_trips") //trips

		n_vehicles[i] = intGet(rec, "n_vehicles") //vehicles

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
	//fmt.Println(data)
	write_segs_db(data, Bd_type, false)
}

func batch_query_write_dir(seg_data []*neo4j.Record, Bd_type string) {

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
	for _, rec := range seg_data {
		var r int64
		var forward bool
		if v_bd, ok := rec.Get(Bd_type); ok {
			r = v_bd.(int64)
		}
		i := r - min_bd
		if v_for, ok := rec.Get("forward"); ok {
			if v_for == nil {
				continue
			}
			forward = v_for.(bool)
		}
		//fmt.Println(rec.Values())
		osm_id = rec.GetByIndex(0).(string) //osm_id

		if forward {
			lq_imp_fw[i] = floatGet(rec, "LQ_imp")
			median_imp_fw[i] = floatGet(rec, "Median_imp")
			uq_imp_fw[i] = floatGet(rec, "UQ_imp")
			n_trips_fw[i] = intGet(rec, "n_trips")       //trips
			n_vehicles_fw[i] = intGet(rec, "n_vehicles") //vehicles
		} else {
			lq_imp_bw[i] = floatGet(rec, "LQ_imp")
			median_imp_bw[i] = floatGet(rec, "Median_imp")
			uq_imp_bw[i] = floatGet(rec, "UQ_imp")
			n_trips_bw[i] = intGet(rec, "n_trips")       //trips
			n_vehicles_bw[i] = intGet(rec, "n_vehicles") //vehicles
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

}

func write_segs_db(data map[string]interface{}, Bd_type string, direction bool) {

	session := Db.NewSession(Sesh_config_segs)

	defer session.Close()
	var statement string
	if !direction {
		statement = `
		USE %[2]s
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
		statement = fmt.Sprintf(statement, Bd_type, Seg_db)
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
		USE %[2]s
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
		statement = fmt.Sprintf(statement, Bd_type, Seg_db)
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
