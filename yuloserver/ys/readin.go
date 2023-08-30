package ys

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"github.com/paulmach/orb"
	source "github.com/xitongsys/parquet-go-source/http"
	"github.com/xitongsys/parquet-go/reader"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

//opts collects the instructions included in the http request
type opts struct {
	gen_resids_only bool
	prune_dupes     bool
	drop_first_stop bool
	max_prune       int64
	speed_missing   bool
	azimuth_missing bool
	raw_output      bool
}

//a srtucture for processing parquet files
type pq_obv struct {
	Datetime *int32   `parquet:"name=datetime, type=INT64, convertedtype=UINT_64"`
	Lat      *float64 `parquet:"name=lat, type=DOUBLE"`
	Lon      *float64 `parquet:"name=lon, type=DOUBLE"`
	Id       *string  `parquet:"name=Vehicle, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Firm     *string  `parquet:"name=Firm, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Type     *string  `parquet:"name=asset_type, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	Azimuth  *float64 `parquet:"name=azimuth, type=DOUBLE"`
	Speed    *float64 `parquet:"name=Speed, type=DOUBLE"`
}

//readCsvRequest reads csv data from an http request and returns it along with options included in the http headers
func readRequest(r http.Request, w http.ResponseWriter) (map[string][]obv, opts) {

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
	speed_missing := r.Header.Get("speed_missing") == "true"
	azimuth_missing := r.Header.Get("azimuth_missing") == "true"
	raw_output := r.Header.Get("raw_output") == "true"

	opts := opts{
		gen_resids_only: gen_resids_only,
		prune_dupes:     prune_dupes,
		drop_first_stop: drop_first_stop,
		max_prune:       max_prune,
		speed_missing:   speed_missing,
		azimuth_missing: azimuth_missing,
		raw_output:      raw_output,
	}
	fmt.Println("File options:")
	opt_values := reflect.ValueOf(opts)
	opt_types := opt_values.Type()
	for i := 0; i < opt_values.NumField(); i++ {
		fmt.Println(opt_types.Field(i).Name, ":", opt_values.Field(i))
	}

	if err != nil {
		fmt.Println("Error retrieving the file")
	}
	var obvs map[string][]obv

	upload_info := !opts.raw_output
	if strings.Contains(fn, ".csv") {
		obvs = readCsv(file, upload_info, false)
	} else if strings.Contains(fn, ".gz") {
		obvs = readCsv(file, upload_info, true)
	} else if strings.Contains(fn, ".parquet") {
		obvs, err = readParquet(file, handler, upload_info, opts)
		if err != nil {
			fmt.Fprintln(w, "Malformed parquet file")
		}
	} else if strings.Contains(fn, ".pbf") {
		obvs, err = readPbf(file, upload_info, opts)
		if err != nil {
			fmt.Fprintln(w, "Malformed protobuf file")
		}
	} else {
		fmt.Fprintln(w, "File must be one of .csv, .gz, .pbf or .parquet")
	}

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

	return m
}

// to be invoked if there is no speed or azimuth fields as that will result in erroneous zero values
func replace_zeros(obvs []obv, options opts) []obv {
	if options.speed_missing {
		fmt.Println("Missing speed")
		for _, o := range obvs {
			o.speed = float64(-1)
		}

	}
	if options.azimuth_missing {
		fmt.Println("Missing azimuth")
		for _, o := range obvs {
			o.azimuth = float64(-1)
		}
	}
	return obvs
}

func readParquet(file multipart.File, handler *multipart.FileHeader, upload_info bool, options opts) (map[string][]obv, error) {
	m := make(map[string][]obv)

	fr := source.NewMultipartFileWrapper(handler, file)
	pr, err := reader.NewParquetReader(fr, new(pq_obv), 4)
	if err != nil {
		fmt.Println(err)
	}
	num := int(pr.GetNumRows())

	obvs := make([]pq_obv, num)
	//pr.SkipRows(1)
	if err = pr.Read(&obvs); err != nil {
		fmt.Println(err)
		return m, err
	}
	//check missing values on azimuth, speed, check vehcile upload
	in_map_already := true
	speed_all_zeros := true
	azi_all_zeros := true
	for _, o := range obvs {
		o_ := obv{
			datetime: int64(*o.Datetime),
			point:    orb.Point{*o.Lon, *o.Lat},
			speed:    *o.Speed,
			id:       *o.Id,
			azimuth:  *o.Azimuth,
		}
		//will flip to false for any non default value
		speed_all_zeros = o_.speed == float64(0)
		azi_all_zeros = o_.azimuth == float64(0)
		m, in_map_already = append_veh_map(m, o_)

		if !in_map_already && upload_info {

			firm := *o.Firm
			if firm == "" {
				firm = "Unknown"
			}

			veh_type := *o.Type
			if veh_type == "" {
				veh_type = "Unknown"
			}
			upload_veh_data(o_.id, veh_type, firm)
		}
	}
	options.speed_missing = speed_all_zeros
	options.azimuth_missing = azi_all_zeros
	for id, veh := range m {
		m[id] = replace_zeros(veh, options)
	}
	return m, nil
}

func readPbf(file io.Reader, upload_info bool, options opts) (map[string][]obv, error) {
	//fmt.Println("Starting readCsv")
	m := make(map[string][]obv)

	batch := &Batch{}
	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Can't read pbf: ", err)
	}
	if err := proto.Unmarshal(bytes, batch); err != nil {
		fmt.Println("Failed to parse pbf: ", err)
		return m, err
	}

	var obvs []obv
	var o obv
	speed_all_zeros := true
	azi_all_zeros := true
	for _, v := range batch.GetTraces() {
		id := v.GetVehicle().GetId()
		fmt.Println(id)
		for _, line := range v.GetObservations() {
			o = obv{}

			o.datetime = line.GetDatetime()
			lon := line.GetLocation().GetLon()
			lat := line.GetLocation().GetLat()
			o.point = orb.Point{lon, lat}

			o.id = id

			o.azimuth = float64(line.GetAzimuth())

			o.speed = float64(line.GetSpeed())

			obvs = append(obvs, o)
			speed_all_zeros = o.speed == float64(0)
			azi_all_zeros = o.azimuth == float64(0)
		}
		options.speed_missing = speed_all_zeros
		options.azimuth_missing = azi_all_zeros
		obvs = replace_zeros(obvs, options)
		m[id] = obvs
		fmt.Println(m)
		if upload_info {
			veh := v.GetVehicle()
			firm := veh.GetFirm()
			if firm == "" {
				firm = "Unknown"
			}

			veh_type := veh.GetType()
			if veh_type == "" {
				veh_type = "Unknown"
			}
			upload_veh_data(o.id, veh_type, firm)
		}
	}

	return m, nil
}

//append_veh_map appends the observation to the map and return whether it is already there
func append_veh_map(m map[string][]obv, o obv) (map[string][]obv, bool) {
	_, ok := m[o.id]
	if ok {
		m[o.id] = append(m[o.id], o)

	} else {
		var n []obv

		m[o.id] = append(n, o)

	}
	return m, ok
}
