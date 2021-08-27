package main

import (
	"./queries"
	"flag"
	"fmt"
	"github.com/gosexy/to"
	//"github.com/jmcvetta/neoism"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	//"strings"
	"yaml"
)

func main() {

	creds_file := flag.String("creds", "../../../Graphupload/neo4jcredsWIN.yaml", "database credentials")
	resume := flag.Bool("resume", false, "whether to resume an interupted query, skipping segs in outfile")
	act_type := flag.String("type", "usage", "activity type: either 'usage' or 'length'")
	year := flag.Int64("year", 2020, "year to query")
	mindur := flag.Int64("mindur", 1800, "minimum duration of stops for activity query")
	month := flag.Int64("month", 0, "month to query, if 0 whole year will be queried")

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
	//ipport := to.String(creds.Get("ipporthttp"))
	boltaddr := to.String(creds.Get("bolt"))

	//connect := fmt.Sprintf("http://%s:%s@%s/db/data", user, pass, ipport)
	//fmt.Printf("%s\n", connect)

	//db, err := neoism.Connect(connect)
	configForNeo4j35 := func(conf *neo4j.Config) {}
	db, err := neo4j.NewDriver(boltaddr, neo4j.BasicAuth(user, pass, ""), configForNeo4j35)

	if err != nil {
		fmt.Printf("Error %v", err)
	}
	defer db.Close()

	//volfile := to.String(params.Get("volfile"))
	activityfile := fmt.Sprintf("activity_%s_%d_%d.csv", *act_type, *year, *month)

	queries.Activityfile = activityfile
	queries.Year = *year
	queries.Month = *month
	queries.Act_type = *act_type
	queries.MinDur = *mindur
	queries.Db = db

	fmt.Println(activityfile)
	//fmt.Println(*month)
	queries.Act_write(activityfile, *resume)

}
