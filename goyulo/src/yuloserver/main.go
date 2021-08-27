package main

import (
	"./ys"
	"fmt"
	"github.com/gosexy/to"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	//"github.com/paulmach/orb/geojson"
	//"io/ioutil"
	"./ys/yaml"
	//"github.com/kyroy/kdtree"
	"net/http"
	"runtime"
)

var (
	//db     *neoism.Database
	//STE    *geojson.FeatureCollection
	//SA2    *geojson.FeatureCollection
	params ys.Para
	creds  *yaml.Yaml
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
	credsfile := to.String(params.Creds)
	creds, _ := yaml.Open(credsfile)
	ys.Creds = creds

	fmt.Println("Reading in shapefiles")

	SA2 := ys.Wkt_readin(params.SA2, "SA2")
	ys.SA2 = SA2
	STE := ys.Wkt_readin(params.STE, "STE")
	ys.STE = STE

	//Creating upload channel if using ys.transfer_upload
	uploader_chan := make(chan string)
	ys.Uploader_chan = uploader_chan
	go ys.Uploader(ys.Uploader_chan)

	//Creating error log channel

	error_chan := make(chan ys.Error_line)
	ys.Error_chan = error_chan
	go ys.Error_logger(ys.Error_chan)

	//add creds start db
	fmt.Println("Connecting to database")
	user := to.String(creds.Get("username"))
	pass := to.String(creds.Get("password"))
	boltaddr := to.String(creds.Get("bolt"))

	configForNeo4j35 := func(conf *neo4j.Config) {}
	db, err := neo4j.NewDriver(boltaddr, neo4j.BasicAuth(user, pass, ""), configForNeo4j35)

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
