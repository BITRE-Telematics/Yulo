package main

import (
	"flag"
	"fmt"
	"github.com/bitre-telematics/queries/queries"
	"github.com/bitre-telematics/queries/yaml"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/xiam/to"
	"strings"
)

func main() {

	creds_file := flag.String("creds", "creds_parameters/creds.yaml", "database credentials")
	params_file := flag.String("params", "creds_parameters/go_query.yaml", "parameter file")
	resume := flag.Bool("resume", false, "whether to resume an interupted query, skipping segs in outfile")
	breakdown := flag.Bool("breakdown", false, "whether to compute breakdowns of segments")
	bd_type := flag.String("bd_type", "hour", "time by which to break down (one of hour, month, dayOfWeek")
	direction := flag.Bool("direction", false, "whether to break down by direction")
	byfirm := flag.Bool("byfirm", false, "whether to cross tab by firm")
	update_db := flag.Bool("update_db", true, "whether to upload precomputes directly - automatically does week and hour bd and direction")
	flag.Parse()
	if *resume {
		fmt.Println("The resume flag is true")
	}

	creds := queries.Read_creds(*creds_file)
	//fmt.Println(creds)
	queries.Fabric = creds.Fabric
	queries.Seg_db = creds.Segs_db
	fmt.Println("Fabric db : ", creds.Fabric)
	fmt.Println("Segs db : ", creds.Segs_db)

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

	params, errparams := yaml.Open(*params_file)
	if errparams != nil {
		fmt.Printf("Could not open YAML file: %s", errparams.Error())
	}
	start := to.Int64(params.Get("start"))
	queries.Start = start
	finish := to.Int64(params.Get("finish"))
	queries.Finish = finish
	queries.Breakdown = *breakdown
	queries.Byfirm = *byfirm
	queries.Bd_type = *bd_type
	queries.Direction = *direction
	queries.Db = db
	speedfile := to.String(params.Get("speedsfile"))
	max_routines := to.Int64(params.Get("max_routines"))
	fmt.Println(params)
	queries.Max_routines = max_routines
	//volfile := to.String(params.Get("volfile"))
	if *update_db {
		fmt.Println("Updating database")
		queries.Seg_write_db("output/database_updated_segs.csv", *resume, speedfile)
	} else {

		if *breakdown {
			speedfile = strings.Replace(speedfile, ".csv", fmt.Sprintf("_by%s.csv", *bd_type), -1)
			fmt.Printf("The breakdown type is %s\n", *bd_type)
		}

		if *direction {
			speedfile = strings.Replace(speedfile, ".csv", "_dir.csv", -1)
			fmt.Println("The direction flag is true")
		}

		if *byfirm {
			speedfile = strings.Replace(speedfile, ".csv", "_byfirm.csv", -1)
			fmt.Println("The byfirm flag is true")
		}

		//var osm_ids []string

		//booleans acting weird

		fmt.Println(speedfile)
		queries.Seg_write(speedfile, *resume)
	}
}
