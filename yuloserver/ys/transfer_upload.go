package ys

// import (
// 	//"bytes"
// 	"encoding/csv"
// 	"fmt"
// 	//"github.com/neo4j/neo4j-go-driver/v4/neo4j"
// 	//"net/http"
// 	"math/rand"
// 	"os"
// 	"os/exec"
// 	"strconv"
// 	"strings"
// 	//"yaml"
// )

// var Uploader_chan chan string

// //these functions provide a way of using direct transfer to the database server and the LOAD_CSV capacity of Cypher

// func Uploader(c chan string) {
// 	for id := range c {
// 		if len(id) > 59 { //failsafe
// 			fmt.Printf("Uploading %s\n", id[:59]) //id variable includes suffix
// 		}
// 		script := cypherScript()
// 		var err error
// 		err = upload_file(id, script["stops"], 0)
// 		if err != nil {
// 			Error_chan <- Error_line{
// 				id:    id,
// 				stage: "Uploading file",
// 				err:   err,
// 			}
// 			return
// 		}

// 		err = upload_file(id, script["obvs"], 0) //test script["obvs_merge"]
// 		if err != nil {
// 			fmt.Println(err)
// 			Error_chan <- Error_line{
// 				id:    id,
// 				stage: "Uploading file",
// 				err:   err,
// 			}
// 			return
// 		}
// 		err = upload_file(id, script["trips"], 0)
// 		if err != nil {
// 			Error_chan <- Error_line{
// 				id:    id,
// 				stage: "Uploading file",
// 				err:   err,
// 			}
// 			return
// 		}
// 		fmt.Printf("%s uploaded\n", id)
// 		delete_file("StopsOut" + id + ".csv")
// 		delete_file("mergedPBF" + id + ".csv")
// 		delete_file("Tripframe" + id + ".csv")

// 	}
// }

// func upload_file(f string, cmd string, i int64) error {
// 	session := Db.NewSession(Sesh_config)

// 	defer session.Close()
// 	//fmt.Println(fmt.Sprintf(cmd, f))

// 	res, err := session.Run(fmt.Sprintf(cmd, f), map[string]interface{}{})

// 	if err != nil {
// 		fmt.Printf("Error %v\n", err)
// 	}

// 	if res.Err() != nil {
// 		fmt.Printf("Error %v %s\n", res.Err(), f)
// 	}

// 	return err
// }

// //dunno why this function is here
// func cypher_upload(f string) {
// 	script := cypherScript()

// 	for _, cmd := range script {
// 		upload_file(f, cmd, 1)
// 	}

// }

// func create_stops_csv(stops []processedStop, id string, suffix string) {
// 	f, _ := os.Create("temp_transfer/" + "StopsOut" + id + suffix + ".csv")
// 	defer f.Close()
// 	writer := csv.NewWriter(f)
// 	headers := []string{
// 		"Stop",
// 		"Vehicle",
// 		"start_time",
// 		"start_time_utc",
// 		"start_time_utcdt",
// 		"start_timedt",
// 		"end_time",
// 		"end_time_utc",
// 		"end_time_utcdt",
// 		"end_timedt",
// 		"lat",
// 		"lon",
// 		"sa2",
// 		"gcc",
// 		"loc",
// 		"addr",
// 	}
// 	writer.Write(headers)
// 	for _, s := range stops {
// 		defer writer.Flush()
// 		l := []string{
// 			s.Stopid,
// 			id,
// 			strconv.FormatInt(s.Start, 10),
// 			strconv.FormatInt(s.Start_utc, 10),
// 			s.Start_utcdt,
// 			s.Startdt,
// 			strconv.FormatInt(s.End, 10),
// 			strconv.FormatInt(s.End_utc, 10),
// 			s.End_utcdt,
// 			s.Enddt,
// 			strconv.FormatFloat(s.Lat, 'f', -1, 64),
// 			strconv.FormatFloat(s.Lon, 'f', -1, 64),
// 			s.Sa2,
// 			s.Gcc,
// 			s.Loc,
// 			s.Addr,
// 		}
// 		writer.Write(l)
// 	}

// }

// func to_string(f float64, v int) string {
// 	vf := float64(v)
// 	if f < vf {
// 		return "NA"
// 	} else {
// 		return strconv.FormatFloat(f, 'f', -1, 64)
// 	}
// }

// func make_imp_string(imp_obvs []processedImpObv) string {
// 	imp_string := ""
// 	bool_string := "false"
// 	for _, i := range imp_obvs {
// 		if i.Forward {
// 			bool_string = "true"
// 		} else {
// 			bool_string = "false"
// 		}
// 		imp_string = imp_string + "|" + i.Osm_id + "$" + bool_string
// 	}
// 	return imp_string
// }

// //poss add in some maps here
// func unpack_obvs(o processedObv, id string, tripid string, w *csv.Writer) {
// 	var o_type string
// 	var imp_obvs_str string
// 	if len(o.Imp_Obvs) > 1 {
// 		o_type = "matched path"
// 		imp_obvs_str = make_imp_string(o.Imp_Obvs)

// 	} else if len(o.Imp_Obvs) == 1 {
// 		o_type = "matched no path"
// 	} else {
// 		o_type = "not matched"
// 	}

// 	l := []string{
// 		id,
// 		tripid,
// 		o.Osm_id,
// 		to_string(o.Speed, 0),
// 		strconv.FormatInt(o.Datetime, 10),
// 		strconv.FormatInt(o.Datetime_utc, 10),
// 		o.Datetime_utcdt,
// 		o.Datetime_dt,
// 		strconv.FormatFloat(o.Lat, 'f', -1, 64),
// 		strconv.FormatFloat(o.Lon, 'f', -1, 64),
// 		to_string(o.Imputed_speed, 1),
// 		to_string(o.Azimuth, 0),
// 		to_string(o.Length, 1),
// 		o_type,
// 		imp_obvs_str,
// 	}
// 	w.Write(l)
// 	//fmt.Println(l)

// 	return
// }

// func create_obvs_trips_csv(trips []processedTrip, id string, suffix string) {
// 	//f_obvs := bytes.NewBufferString("")
// 	//f_trips := bytes.NewBufferString("")
// 	f_obvs, _ := os.Create("temp_transfer/" + "mergedPBF" + id + suffix + ".csv")
// 	f_trips, _ := os.Create("temp_transfer/" + "Tripframe" + id + suffix + ".csv")
// 	defer f_obvs.Close()
// 	defer f_trips.Close()

// 	writer_obvs := csv.NewWriter(f_obvs)
// 	writer_trips := csv.NewWriter(f_trips)

// 	headers_obvs := []string{
// 		"Vehicle",
// 		"Trip",
// 		"osm_id",
// 		"speed",
// 		"datetime",
// 		"datetime_utc",
// 		"datetime_utcdt",
// 		"datetimedt",
// 		"lat",
// 		"lon",
// 		"imputed_speed",
// 		"azimuth",
// 		"length",
// 		"type",
// 		"imp_obvs",
// 	}
// 	writer_obvs.Write(headers_obvs)
// 	writer_obvs.Flush()

// 	headers_trips := []string{
// 		"Trip",
// 		"Prior_stop",
// 		"Following_stop",
// 	}

// 	writer_trips.Write(headers_trips)
// 	writer_trips.Flush()

// 	for _, t := range trips {
// 		defer writer_trips.Flush()
// 		defer writer_obvs.Flush()
// 		l_trip := []string{
// 			t.Trip,
// 			t.Prior_stop,
// 			t.Following_stop,
// 		}

// 		writer_trips.Write(l_trip)

// 		for _, o := range t.Obvs {
// 			unpack_obvs(o, id, t.Trip, writer_obvs)

// 		}

// 	}

// }

// func transfer_csv(filename string) error {
// 	server := Params.File_transfer_server

// 	key := Creds.Transfer_key

// 	header_other := "-H 'other: false'"
// 	ul_dir := "temp_transfer/"
// 	header_file := fmt.Sprintf("-H 'filename: %s'", filename)
// 	command := fmt.Sprintf("curl -X POST -F 'myFile=@%s' %s -H 'key: %s' %s %s", ul_dir+filename, server, key, header_other, header_file)

// 	//fmt.Printf("Bash command is %s\n", command)
// 	_, err := exec.Command("bash", "-c", command).Output()
// 	if err != nil {
// 		fmt.Println("Transfer error")

// 		return err
// 	}
// 	return nil
// }

// func delete_file(filename string) {
// 	server := Params.File_transfer_server

// 	key := Creds.Transfer_key

// 	header_other := "-H 'other: false'"
// 	header_file := fmt.Sprintf("-H 'filename: %s'", filename)

// 	command := fmt.Sprintf("curl -X POST %s -H 'key: %s' %s %s  -H 'delete: true'", server, key, header_other, header_file)

// 	//fmt.Printf("Bash command is %s\n", command)
// 	_, err := exec.Command("bash", "-c", command).Output()
// 	if err != nil {
// 		fmt.Printf("Deletion error err %s \n", err)
// 	}
// }

// func transfer_upload(trips []processedTrip, stops []processedStop, id string) {
// 	//adding batch specific suffix to files to avoid overwriting whilst they are in the queue to be uploaded.
// 	//checks to see if there is data too
// 	suffix := fmt.Sprintf("%x", rand.Int31())
// 	if len(stops) > 0 {
// 		suffix = strings.Split(stops[0].Stopid, "_")[1]
// 	}
// 	create_stops_csv(stops, id, suffix)

// 	err := transfer_csv("StopsOut" + id + suffix + ".csv")
// 	if err != nil {
// 		Error_chan <- Error_line{
// 			id:    id,
// 			stage: "Transferring file",
// 			err:   err,
// 		}
// 		return
// 	}
// 	os.Remove("temp_transfer/" + "StopsOut" + id + suffix + ".csv")

// 	create_obvs_trips_csv(trips, id, suffix)
// 	err = transfer_csv("mergedPBF" + id + suffix + ".csv")
// 	if err != nil {
// 		Error_chan <- Error_line{
// 			id:    id,
// 			stage: "Transferring file",
// 			err:   err,
// 		}
// 		return
// 	}
// 	os.Remove("temp_transfer/" + "mergedPBF" + id + suffix + ".csv")

// 	err = transfer_csv("Tripframe" + id + suffix + ".csv")
// 	if err != nil {
// 		Error_chan <- Error_line{
// 			id:    id,
// 			stage: "Transferring file",
// 			err:   err,
// 		}
// 		return
// 	}
// 	os.Remove("temp_transfer/" + "Tripframe" + id + suffix + ".csv")

// 	Uploader_chan <- id + suffix

// }
