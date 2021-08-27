package ys

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
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

func readResiduals(id string) []obv {
	resid_fn := Params.Residual_dir + id
	file, err := os.Open(resid_fn)
	defer file.Close()
	var o []obv
	if err != nil {
		fmt.Printf("No residuals for %s\n", id)
		return o
	}
	//readCsv returns a map[string] which has one member here
	o = readCsv(file, false, false)[id]

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
	fmt.Printf("Using %d of %d recorded residuals\n", len(include), len(include)+len(exclude))
	return include, exclude
}
