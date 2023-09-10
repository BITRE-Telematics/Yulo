package ys

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"os"
	"strconv"
	"strings"
	"time"
)

var Resids_config neo4j.SessionConfig

type residual_return struct {
	Datetime int64
	Speed    float64
	Id       string
	Azimuth  float64
	Lat      float64
	Lon      float64
}

func toResidualReturn(obvs []obv) []residual_return {
	var out []residual_return
	for _, o := range obvs {
		o_ := residual_return{
			Datetime: o.datetime,
			Speed:    o.speed,
			Id:       o.id,
			Azimuth:  o.azimuth,
			Lat:      o.point.Lat(),
			Lon:      o.point.Lon(),
		}
		out = append(out, o_)
	}
	return out
}

//writeResiduals dumps the unprocessed observations from a vehicle to disk
func writeResiduals(obvs []obv) {

	if len(obvs) > 0 {
		id := obvs[0].id
		file, err := os.Create(Params.Residual_dir + id)
		defer file.Close()
		writer := csv.NewWriter(file)
		if err != nil {
			fmt.Println("Error creating residual file")

		}
		headers := []string{
			"datetime",
			"lon",
			"lat",
			"Vehicle",
			"azimuth",
			"Speed",
		}
		writer.Write(headers)
		for _, o := range obvs {
			l := []string{
				strconv.FormatInt(o.datetime, 10),
				strconv.FormatFloat(o.point.Lon(), 'f', -1, 64),
				strconv.FormatFloat(o.point.Lat(), 'f', -1, 64),
				o.id,
				strconv.FormatFloat(o.azimuth, 'f', 2, 64),
				strconv.FormatFloat(o.speed, 'f', 2, 64),
			}

			writer.Write(l)
			writer.Flush()
		}
	}

}

//writeResidualsDb dumps residuals in csv format to an attribute in the database. It is not currently used
func writeResidualsDb(obvs []obv, i int) {

	if len(obvs) > 0 {
		id := obvs[0].id
		b := new(bytes.Buffer)
		writer := csv.NewWriter(b)
		headers := []string{
			"datetime",
			"lon",
			"lat",
			"Vehicle",
			"azimuth",
			"Speed",
		}
		writer.Write(headers)
		for _, o := range obvs {
			l := []string{
				strconv.FormatInt(o.datetime, 10),
				strconv.FormatFloat(o.point.Lon(), 'f', -1, 64),
				strconv.FormatFloat(o.point.Lat(), 'f', -1, 64),
				o.id,
				strconv.FormatFloat(o.azimuth, 'f', 2, 64),
				strconv.FormatFloat(o.speed, 'f', 2, 64),
			}

			writer.Write(l)
		}
		writer.Flush()
		csv_string := b.String()
		//fmt.Println(b)
		session := Db.NewSession(Resids_config)

		defer session.Close()
		query := "MERGE(a:Asset{id: $ID}) SET a.residuals = $RESIDUALS"
		parameters := map[string]interface{}{"ID": id, "RESIDUALS": csv_string}
		res, err := session.Run(query, parameters)

		if err != nil && i < 400 {
			if i > 0 {
				fmt.Println(err, id, i)
				//fmt.Println(obv)
			}
			time.Sleep(time.Second * 60)
			writeResidualsDb(obvs, i+1)
		}

		if res != nil {
			if res.Err() != nil {
				fmt.Println(res.Err())
			}
		}
	}

}

//readResiduals reads residual data for a vehicle from disc
func readResiduals(id string) []obv {
	resid_fn := Params.Residual_dir + id
	file, err := os.Open(resid_fn)
	defer file.Close()
	var o []obv
	if err != nil {
		fmt.Printf(Yellow+"No residuals for %s\n"+Yellow, id)
		return o
	}
	//readCsv returns a map[string] which has one member here
	o = readCsv(file, false, false)[id]

	return o
}

//readResidualsDb reads residual data for a vehicle from the database
func readResidualsDb(id string) []obv {
	session := Db.NewSession(Resids_config)
	defer session.Close()
	statement := "MATCH(a:Asset{id: $ID}) return a.residuals"
	parameters := map[string]interface{}{"ID": id}
	db_resp, err := session.Run(statement, parameters)

	if db_resp == nil {
		fmt.Println("Nil result")
		return nil
	}

	if err != nil {
		fmt.Print("Check database error")
	}
	if db_resp.Err() != nil {
		fmt.Print("DB residuals error")
		fmt.Println(db_resp.Err())
	}
	var o []obv
	//var csv_string string
	if db_resp.Next() {
		fmt.Println(db_resp)
		csv_string := db_resp.Record().GetByIndex(0)
		if csv_string != nil {
			csv := strings.NewReader(csv_string.(string))
			o = readCsv(csv, false, false)[id]

			//fmt.Println(o)

		} else {
			fmt.Printf("No residuals in db for %s \n", id)
		}
	} else {
		fmt.Printf("No residuals in db for %s \n", id)

	}

	return o
}

//max_datetime returns the maximum datatime from a slice of observations
func max_datetime(obvs []obv) int64 {
	m := obvs[0].datetime
	for _, o := range obvs {
		if o.datetime > m {
			m = o.datetime
		}
	}
	return m
}

//min_datetime returns the minimum datatime from a slice of observations
func min_datetime(obvs []obv) int64 {
	m := obvs[0].datetime
	for _, o := range obvs {
		if o.datetime < m {
			m = o.datetime
		}
	}
	return m
}

//split_resids seperates residuals into data that will be processed for the current time range and that which will not
func split_resids(resids []obv, min int64, max int64) ([]obv, []obv) {
	var include []obv
	var exclude []obv

	for _, o := range resids {
		if o.datetime < min || o.datetime > max {
			exclude = append(exclude, o)
		} else {
			include = append(include, o)
		}
	}
	//fmt.Printf(Purple+"Using %d of %d recorded residuals for %s\n"+Reset, len(include), len(include)+len(exclude))
	return include, exclude
}
