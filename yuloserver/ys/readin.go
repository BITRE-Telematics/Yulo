package ys

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"github.com/paulmach/orb"
	"io"
	"net/http"
	"strconv"
	"strings"
)

//opts collects the instructions included in the http request
type opts struct {
	gen_resids_only bool
	prune_dupes     bool
	drop_first_stop bool
	max_prune       int64
}

//readCsvRequest reads csv data from an http request and returns it along with options included in the http headers
func readCsvRequest(r http.Request, w http.ResponseWriter) (map[string][]obv, opts) {

	r.ParseMultipartForm(10 << 30)
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Fprintln(w, "Error in form file")
	}
	fmt.Println(handler.Header)

	fn := handler.Filename
	fmt.Printf("file is %s\n", fn)

	gen_resids_only := r.Header.Get("gen_resids_only") == "true"
	prune_dupes := r.Header.Get("prune_dupes") == "true"
	drop_first_stop := r.Header.Get("drop_first_stop") == "true"
	max_prune_str := r.Header.Get("max_prune")
	max_prune, _ := strconv.ParseInt(max_prune_str, 10, 64)

	if err != nil {
		fmt.Println("Error retrieving the file")
	}

	compressed := strings.Contains(fn, ".gz")

	obvs := readCsv(file, false, compressed)
	opts := opts{
		gen_resids_only: gen_resids_only,
		prune_dupes:     prune_dupes,
		drop_first_stop: drop_first_stop,
		max_prune:       max_prune,
	}
	fmt.Println(opts)
	return obvs, opts
}

//readCsv reads data from the request and also uploads vehicle data to the database
func readCsv(file io.Reader, upload_info bool, compressed bool) map[string][]obv {
	//fmt.Println("Starting readCsv")
	m := make(map[string][]obv)
	var reader *csv.Reader
	if compressed {
		fmt.Println("Making gz reader")
		gz, err := gzip.NewReader(file)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("made gz reader")
		defer gz.Close()

		reader = csv.NewReader(gz)

		//content, _ := io.ReadAll(gz)
		//reader = csv_NewReader(content)

	} else {
		reader = csv.NewReader(bufio.NewReader(file))
	}
	headers, err := reader.Read()
	if err != nil {
		fmt.Println("csv error")
		return m
	}
	headermap := make(map[string]int)
	veh_type_present := false
	firm_present := false
	for i, h := range headers {
		headermap[h] = i
		veh_type_present = (h == "asset_type") || veh_type_present
		firm_present = (h == "Firm") || firm_present
	}

	//var obvs []obv
	var o obv

	in_map_already := true
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
		} else {
			o = obv{}

			o.datetime, _ = strconv.ParseInt(line[headermap["datetime"]], 10, 64)
			lon, _ := strconv.ParseFloat(line[headermap["lon"]], 64)
			lat, _ := strconv.ParseFloat(line[headermap["lat"]], 64)
			o.point = orb.Point{lon, lat}

			o.id = line[headermap["Vehicle"]]

			if azi, ok := headermap["azimuth"]; ok {
				o.azimuth, _ = strconv.ParseFloat(line[azi], 64)
			} else {
				o.azimuth = float64(-1)
			}
			if sp, ok := headermap["Speed"]; ok {
				o.speed, _ = strconv.ParseFloat(line[sp], 64)
			} else {
				o.speed = float64(-1)
			}
		}
		//obvs = append(obvs, o)

		//untested alt to append directly - if used get rid of split by veh call and make this func return a map
		//fmt.Println(line)
		//fmt.Printf("Firm Present:%s\n", firm_present)
		m, in_map_already = append_veh_map(m, o)

		if !in_map_already && upload_info {

			firm := "Unknown"
			if firm_present {
				firm = line[headermap["Firm"]]
			}

			veh_type := "Unknown"
			if veh_type_present {
				veh_type = line[headermap["asset_type"]]
			}
			upload_veh_data(o.id, veh_type, firm)
		}

	}
	//return obvs

	return m
}

/*func sep_by_veh(table []obv) map[string][]obv {
	m := make(map[string][]obv)
	for _, o := range table {
		_, ok := m[o.id]
		if ok {
			m[o.id] = append(m[o.id], o)
		} else {
			var n []obv
			m[o.id] = append(n, o)
		}
	}
	return m
}*/

//append_veh_map appends the observation to the map and return whether it is already there
func append_veh_map(m map[string][]obv, o obv) (map[string][]obv, bool) {
	_, ok := m[o.id]
	if ok {
		m[o.id] = append(m[o.id], o)
	} else {
		var n []obv
		//fmt.Println(m[o.id])

		m[o.id] = append(n, o)
		// fmt.Println("Print loop")
		// for k := range m {
		// 	fmt.Println(k)
		// }
	}
	//fmt.Println(ok)
	return m, ok
}
