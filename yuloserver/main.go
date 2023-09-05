package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	"os"
	"runtime"
	"yuloserver/ys"
)

var (
	//db     *neoism.Database
	//STE    *geojson.FeatureCollection
	//SA2    *geojson.FeatureCollection
	params ys.Para
)

func setupRoutes(params ys.Para) {
	fmt.Println(params)

	addr := "0.0.0.0:" + params.Yuloport

	guard := make(chan struct{}, params.Max_routines)
	ys.Guard = guard
	//if using resource limits this isn't necessary except for the function call

	fmt.Println("Listening for vehicles on " + addr)

	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {

		ys.ProcessFile(w, r)

	})

	http.Handle("/", http.FileServer(http.Dir("../../../UploadData")))
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}

}

func main() {
	fmt.Println("Setting parameters")
	params = ys.Set_parameters()
	ys.Params = params

	if _, err := os.Stat(params.Residual_dir); os.IsNotExist(err) {
		err := os.MkdirAll(params.Residual_dir, 0755)
		if err != nil {
			fmt.Println(err)
		}
	}

	creds := ys.Read_creds(params.Creds)

	ys.Creds = creds

	fmt.Println("Reading in shapefiles")

	SA2 := ys.Wkt_readin(params.SA2, "SA2")
	ys.SA2 = SA2
	STE := ys.Wkt_readin(params.STE, "STE")
	ys.STE = STE

	// //Creating upload channel if using ys.transfer_upload
	// uploader_chan := make(chan string)
	// ys.Uploader_chan = uploader_chan
	// go ys.Uploader(ys.Uploader_chan)

	//Creating error log channel

	error_chan := make(chan ys.Error_line)
	ys.Error_chan = error_chan
	go ys.Error_logger(ys.Error_chan)

	//add creds start db
	fmt.Println("Connecting to database")

	db, err := neo4j.NewDriver(
		creds.Bolt,
		neo4j.BasicAuth(
			creds.Username,
			creds.Password,
			"",
		),
	)
	defer db.Close()
	//naming database in neo4j5
	sesh_config := neo4j.SessionConfig{
		DatabaseName: creds.Db_name,
	}
	ys.Sesh_config = sesh_config

	sesh_config_fabric := neo4j.SessionConfig{
		DatabaseName: creds.Fabric,
	}
	ys.Sesh_config_fabric = sesh_config_fabric

	resids_config := neo4j.SessionConfig{
		DatabaseName: creds.Resids_db,
	}
	ys.Resids_config = resids_config

	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer db.Close()
	ys.Db = db
	if err != nil {
		fmt.Println("Database connection error")
	}

	//creating rtree
	if params.Match_locs {
		fmt.Println("Creating Address tree")
		ys.Addr_tree, ys.Loc_addr = ys.Make_tree("Address")
		fmt.Println("Creating Location Tree")
		ys.Ra_tree, ys.Loc_ra = ys.Make_tree("Location")
	}
	//to make sure bf_feed workaround still works
	ys.Current_OS = runtime.GOOS
	setupRoutes(params)
}
