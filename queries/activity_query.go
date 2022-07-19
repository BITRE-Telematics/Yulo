package main

import (
	"flag"
	"fmt"
	"github.com/bitre-telematics/queries/queries"
	"github.com/xiam/to"
	//"github.com/jmcvetta/neoism"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	//"strings"
	"github.com/bitre-telematics/queries/yaml"
)

func main() {

	creds_file := flag.String("creds", "creds_parameters/creds.yaml", "database credentials")
	resume := flag.Bool("resume", false, "whether to resume an interupted query, skipping segs in outfile")
	act_type := flag.String("type", "usage", "activity type: either 'usage' or 'length'")
	year := flag.Int64("year", 2020, "year to query")
	mindur := flag.Int64("mindur", 1800, "minimum duration of stops for activity query")
	month := flag.Int64("month", 0, "month to query, if 0 whole year will be queried")

	flag.Parse()

	if *resume {
		fmt.Println("The resume flag is true")
	}

	ccreds := queries.Read_creds(*creds_file)

	queries.Fabric = creds.Fabric
	queries.Seg_db = creds.Segs_db

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
	//naming database in neo4j4
	sesh_config := neo4j.SessionConfig{
		DatabaseName: creds.Fabric,
	}
	queries.Sesh_config = sesh_config

	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer db.Close()
	queries.Db = db
	if err != nil {
		fmt.Println("Database connection error")
	}

	//volfile := to.String(params.Get("volfile"))
	queries.Max_routines = 15
	activityfile := fmt.Sprintf("output/activity_%s_%d_%d.csv", *act_type, *year, *month)

	queries.Activityfile = activityfile
	queries.Year = *year
	queries.Month = *month
	queries.Act_type = *act_type
	queries.MinDur = *mindur

	queries.Fabric = fabric

	fmt.Println(activityfile)
	//fmt.Println(*month)
	queries.Act_write(activityfile, *resume)

}
