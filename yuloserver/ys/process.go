package ys

import (
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	"os"
	"sort"
	"time"
)

var Db neo4j.Driver
var Params Para
var Guard chan struct{}

func ProcessFile(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.Header.Get("filename"))
	//fmt.Print(Params)
	start := time.Now()
	obvs_map, opts := readCsvRequest(*r, w)
	if opts.gen_resids_only {
		fmt.Println("Generating residuals only")
	}

	if opts.prune_dupes {
		fmt.Println("Pruning observations to avoid duplication")
	}

	if opts.drop_first_stop {
		fmt.Println("Dropping first stop pair to residuals to be captured when prior data is processed")
	}

	for _, v := range obvs_map {

		Guard <- struct{}{} // should block if channel full, comment out if using resource limits and comment below

		go func(v []obv) {
			//_ := check_resources(Params.Max_memory, Params.Max_cpu) //should block until resources are free
			ProcessVehicle(v, opts)
			<-Guard

		}(v)
	}

	fmt.Fprintf(w, "File completely entered into server in %s at %s\n", time.Since(start).String(), time.Now().String())

}

func ProcessVehicle(obvs []obv, opts opts) {
	sort.SliceStable(obvs, func(i, j int) bool { return obvs[i].datetime < obvs[j].datetime })
	id := obvs[0].id
	fmt.Println(Yellow+"Starting ", id+Reset)
	if inc_zero_dt(obvs) {
		fmt.Printf("Asset %s has invalid datetimes, potentially due to malformed csv, not processing \n", id)
		Error_chan <- Error_line{
			id:    id,
			err:   errors.New("Includes 0 datetimes"),
			stage: "CSV read in",
		}
		return
	}
	resids := readResiduals(id)
	max := max_datetime(obvs) + Params.MaxResidsGap
	min := min_datetime(obvs) - Params.MaxResidsGap
	var reserved []obv
	var write_resids []obv
	resids, reserved = split_resids(resids, min, max)

	//This is an option to only generate residuals in case they need to be fixed for whatever reason
	if opts.gen_resids_only {
		fmt.Println("Generating residuals only")
		start := time.Now()
		vehpack := CichCluster(obvs, id, opts.drop_first_stop)
		write_resids = append(vehpack.residuals, reserved...)
		fmt.Printf("Writing %d total residuals for %s \n", len(write_resids), id)
		writeResiduals(write_resids)
		//the following is usually in transfer_upload to make sure residuals aren't lost if upload fails
		//it perhaps is more ideally done with a channel and function rather than this duplicated code
		resids_tmp := Params.Residual_dir + id + "TEMP"
		os.Remove(Params.Residual_dir + id)
		os.Rename(resids_tmp, Params.Residual_dir+id)
		fmt.Printf("Residuals for %s generated in %s\n", id, time.Since(start).String())
		return
	}

	//check database for duplicates
	//fmt.Println("Checking for dupes for %s", id)
	if opts.prune_dupes {
		dupes, max_db_dt := checkDatabaseDupe(obvs, opts.max_prune)
		if dupes {
			fmt.Printf(Yellow+"Possible duplicate data for Asset %s with %s < %s\n"+Reset, id, min, max_db_dt)
			obvs = prune_dupes(obvs, max_db_dt)

		}
	}
	obvs = append(resids, obvs...)
	sort.SliceStable(obvs, func(i, j int) bool { return obvs[i].datetime < obvs[j].datetime })
	if len(obvs) > 0 {

		start := time.Now()
		vehpack := CichCluster(obvs, id, opts.drop_first_stop)

		fmt.Printf(Grey+"CichCluster for %s completed in %s\n"+Reset, id, time.Since(start).String())

		start = time.Now()
		stops := sum_stops(vehpack.stops)
		// fmt.Println(len(stops))
		fmt.Printf(Grey+"SummaryStops for %s completed in %s\n"+Reset, id, time.Since(start).String())

		start = time.Now()
		tripsout, err := bffeeder(vehpack.trips)
		if err != nil {
			fmt.Printf("Error in Barefoot for %s completed in because of : \n", id, err)
		}

		fmt.Printf(Grey+"Barefoot for %s completed in %s\n"+Reset, id, time.Since(start).String())

		start = time.Now()
		tripspbf := postbarefoot(tripsout)
		fmt.Printf(Grey+"Postbarefoot for %s completed in %s\n"+Reset, id, time.Since(start).String())

		//transfer_upload(tripspbf, stops, id)

		// alternate db write
		start = time.Now()
		stopswrite(stops, id, 1)
		tripswrite(tripspbf, id)
		fmt.Printf(White+"Upload for %s completed in %s\n"+Reset, id, time.Since(start).String())
		//fmt.Printf("writing residuals for %s\n", id)

		//add in retained too early/late residuals
		write_resids = append(vehpack.residuals, reserved...)

		writeResiduals(write_resids)
		fmt.Printf(Red+"%s done at %s\n"+Reset, id, time.Now().String())
	}
}
