package ys

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	//"os"
	"sort"
	"sync"
	"time"
)

//DB represents a global database connection object
var Db neo4j.Driver

//Params are global parameters
var Params Para

// Guard is a channel to limit the number of concurrent processes
var Guard chan struct{}

//ProcessFile takes an http request and processes the data therein and distributes it beween processes by vehicle
func ProcessFile(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.Header.Get("filename"))
	//fmt.Print(Params)
	start := time.Now()
	obvs_map, opts := readRequest(*r, w)
	if opts.gen_resids_only {
		fmt.Println("Generating residuals only")
	}

	if opts.prune_dupes {
		fmt.Println("Pruning observations to avoid duplication")
	}

	if opts.drop_first_stop {
		fmt.Println("Dropping first stop pair to residuals to be captured when prior data is processed")
	}
	custom, parameters := check_custom_params(opts)
	fmt.Println("Custom params", parameters)
	//potentially go to defaults
	if custom {
		if !opts.raw_output {
			fmt.Fprintf(w, "Custom parameters only available with raw_output, stopping")
			//parameters = Params
			return
		}
	}
	var wg sync.WaitGroup
	raw_chan := make(chan raw_return, len(obvs_map))
	var raw_outputs []raw_return
	for _, v := range obvs_map {

		Guard <- struct{}{} // should block if channel full, comment out if using resource limits and comment below
		wg.Add(1)
		go func(v []obv) {

			//_ := check_resources(Params.Max_memory, Params.Max_cpu) //should block until resources are free
			ProcessVehicle(v, opts, raw_chan, parameters)
			wg.Done()
			//fmt.Println("dropping guard")
			<-Guard

		}(v)
	}
	//fmt.Println(Red + "All submitted" + Reset)
	go func() {
		defer close(raw_chan)
		wg.Wait()

	}()

	if opts.raw_output {
		for r := range raw_chan {
			raw_outputs = append(raw_outputs, r)
		}
		dump, err := json.Marshal(raw_outputs)
		if err != nil {
			fmt.Println("Json marshall error: err")
		}
		fmt.Fprintf(w, string(dump))
		return
	}
	fmt.Fprintf(w, "File completely processed in %s at %s\n", time.Since(start).String(), time.Now().String())
	fmt.Println(Red+"File completely processed in %s at %s\n", time.Since(start).String(), time.Now().String()+Reset)
	//fmt.Fprintf(w, "File completely entered into server in %s at %s\n", time.Since(start).String(), time.Now().String())

}

//ProcessVehicle processes a given vehicle
func ProcessVehicle(obvs []obv, opts opts, raw_chan chan raw_return, parameters Para) {
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
		fmt.Println(Red + "pushing to channel: " + Reset)
		fmt.Println(len(raw_chan))
		raw_chan <- raw_return{}
		return
	}
	resids := readResidualsDb(id)
	if len(resids) == 0 {
		resids = readResiduals(id)
	}
	max := max_datetime(obvs) + Params.MaxResidsGap
	min := min_datetime(obvs) - Params.MaxResidsGap
	var reserved []obv
	var write_resids []obv
	resids, reserved = split_resids(resids, min, max)

	//This is an option to only generate residuals in case they need to be fixed for whatever reason
	if opts.gen_resids_only {
		fmt.Println("Generating residuals only")
		//start := time.Now()
		vehpack := CichCluster(obvs, id, opts.drop_first_stop, parameters)
		write_resids = append(vehpack.residuals, reserved...)
		fmt.Printf("Writing %d total residuals for %s \n", len(write_resids), id)

		if opts.raw_output {
			// fmt.Println(Red + "pushing to channel: " + Reset)
			// fmt.Println(len(raw_chan))
			raw_chan <- raw_return{
				Residuals: toResidualReturn(write_resids),
			}
			return

		} else {
			writeResidualsDb(write_resids, 1)

			// fmt.Println(Red + "pushing to channel: " + Reset)
			// fmt.Println(len(raw_chan))
			raw_chan <- raw_return{}
			return
		}
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
		vehpack := CichCluster(obvs, id, opts.drop_first_stop, parameters)

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
		tripspbf := postbarefoot(tripsout) // pbf is post bare foot
		fmt.Printf(Grey+"Postbarefoot for %s completed in %s\n"+Reset, id, time.Since(start).String())

		if opts.raw_output {
			residuals := toResidualReturn(vehpack.residuals)
			out := raw_return{
				Id:        id,
				Stops:     stops,
				Trips:     tripspbf,
				Residuals: residuals,
			}
			// fmt.Println(Red + "pushing to channel: " + Reset)
			// fmt.Println(len(raw_chan))
			raw_chan <- out
			return

		} else {
			//normal upload
			start = time.Now()
			stopswrite(stops, id, 1)
			tripswrite(tripspbf, id)
			fmt.Printf(White+"Upload for %s completed in %s\n"+Reset, id, time.Since(start).String())
			//fmt.Printf("writing residuals for %s\n", id)

			//add in retained too early/late residuals
			write_resids = append(vehpack.residuals, reserved...)

			writeResidualsDb(write_resids, 1)
		}
		fmt.Printf(Red+"%s done at %s\n"+Reset, id, time.Now().String())
		fmt.Println(Red + "pushing to channel: " + Reset)
		fmt.Println(len(raw_chan))
		raw_chan <- raw_return{}
		return

	}
	fmt.Println(Red + "pushing to channel: " + Reset)
	fmt.Println(len(raw_chan))
	raw_chan <- raw_return{}
	return
}

type raw_return struct {
	Id        string
	Stops     []processedStop
	Trips     []processedTrip
	Residuals []residual_return
}
