package ys

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

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
		csv_string := b.String()
		session := Db.NewSession(Sesh_config)

		defer session.Close()
		query := "MATCH(a:Asset{id: $ID}) SET a.residuals = $RESIDUALS"
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

func readResidualsDb(id string) []obv {
	session := Db.NewSession(Sesh_config)
	defer session.Close()
	statement := "MATCH(a:Asset{id: $ID}) return c.residuals"
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
	var csv_string string
	if db_resp.Next() {
		csv_string := db_resp.Record().GetByIndex(0)
		if csv_string == nil {
			fmt.Printf("No residuals in db for %s \n", id)
			return o
		}
	} else {
		fmt.Printf("No residuals in db for %s \n", id)
		return o
	}

	csv := strings.NewReader(csv_string)
	o = readCsv(csv, false, false)[id]

	return o
}

func max_datetime(obvs []obv) int64 {
	m := obvs[0].datetime
	for _, o := range obvs {
		if o.datetime > m {
			m = o.datetime
		}
	}
	return m
}

func min_datetime(obvs []obv) int64 {
	m := obvs[0].datetime
	for _, o := range obvs {
		if o.datetime < m {
			m = o.datetime
		}
	}
	return m
}

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
