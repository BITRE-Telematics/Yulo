package main

import (
	"./queries"
	"flag"
	"fmt"
	"github.com/gosexy/to"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"strings"
	"yaml"
)

func main() {

	creds_file := flag.String("creds", "../../../Graphupload/neo4jcredsWIN.yaml", "database credentials")
	params_file := flag.String("params", "../../../Graphupload/precompute_segs/go_query.yaml", "paramenter file")
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

	creds, errcreds := yaml.Open(*creds_file)
	if errcreds != nil {
		fmt.Printf("Could not open YAML file: %s", errcreds.Error())
	}
	user := to.String(creds.Get("username"))
	pass := to.String(creds.Get("password"))
	boltaddr := to.String(creds.Get("bolt"))

	configForNeo4j35 := func(conf *neo4j.Config) {}
	db, err := neo4j.NewDriver(boltaddr, neo4j.BasicAuth(user, pass, ""), configForNeo4j35)

	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer db.Close()

	if err != nil {
		fmt.Printf("Error %v", err)
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
	//volfile := to.String(params.Get("volfile"))
	if *update_db {
		fmt.Println("Updating database")
		queries.Seg_write_db("database_updated_segs.csv", *resume, speedfile)
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
