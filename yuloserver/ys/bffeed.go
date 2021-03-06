package ys

import (
	"bufio"
	"encoding/json"
	"fmt"
	//"io.Copy"
	//"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sort"
	"time"
	//"unicode/utf8"
	"strings"
)

var Current_OS string

type json_in struct {
	Point    string  `json:"point"`
	Datetime int64   `json:"time"`
	Azimuth  float64 `json:"azimuth"`
	Id       string  `json:"id"`
}

//try reshaping the barefoot output again

type imp_obv struct {
	Osm_id          string  `json:"osm_id"`
	Imputed_azimuth float64 `json:"imputed_azimuth"`
	Obv_type        string  `json:"type"`
	Target          string  `json:"target"`
	Forward         bool    `json:"forward"`
}

type Json_out struct {
	Datetime        int64   `json:"datetime"`
	Osm_id          string  `json:"osm_id"`
	Imputed_azimuth float64 `json:"imputed_azimuth"`
	Forward         bool    `json:"forward"`
	Obv_type        string  `json:"type"`
	//Target          string    `json:"target"`
	Roads         []imp_obv `json:"roads"`
	Imputed_speed float64   `json:"imputed_speed"`
	Length        float64   `json:"length"`
	SA2           string    `json:"sa2"`
	GCC           string    `json:"gcc"`
	Newsubtrip    bool
	Lat           float64
	Lon           float64
	Speed         float64
}

type trip_bf_out struct {
	obvs           []Json_out
	prior_stop     string
	following_stop string
	id             string
}

//add error handling
func bffeed(trip trip) (trip_bf_out, error) {
	//fmt.Println(trip.tripid)
	addr := Params.Host + ":" + Params.Port
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		fmt.Println("Barefoot connection error")
		return trip_bf_out{}, err

	}
	cmds := to_json(trip)

	var matched_trip []Json_out
	r := bufio.NewReader(conn)

	for _, cmd := range cmds {
		//matched_subtrip = nil
		body, err := io_func(r, conn, cmd)
		if err != nil {
			fmt.Println(err)
		}

		var obvs []Json_out

		err = json.Unmarshal([]byte(body), &obvs)
		if err != nil {
			//fmt.Printf("%s error from barefoot output json with %s with %d points\n", err, trip.tripid, len(trip.obvs))

		}

		//fmt.Println(obvs)

		matched_trip = append(matched_trip, obvs...)

	}
	matched_trip_out := merge_orig(trip, matched_trip)
	return matched_trip_out, nil

}

/*
There is a strange error on a small subset of the json (cmd below) that return empty responses from the
Barefoot server. It is always the same JSON. The same JSON works when submitted via the command line.
This may be an encoding isue but in the mean time the error handling will write the JSON to disk and
submit it via the system call.
*/
func io_func(r *bufio.Reader, conn net.Conn, cmd string) (string, error) {
	fmt.Fprintf(conn, cmd+"\n")
	header, err := r.ReadString('\n')
	if err != nil {
		//fmt.Printf("\n\n\nNew error at %s\n", time.Now().String())
		//fmt.Println(header)
		//fmt.Println(err)
		//fmt.Println(cmd)
		fn := fmt.Sprintf("%d-temp.json", time.Now().UnixNano())
		file, fileerr := os.OpenFile(fn, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

		//file, fileerr := os.Create(filename)
		if fileerr != nil {
			fmt.Printf("File error: %s", fileerr)
		}
		defer os.Remove(fn)

		fmt.Fprintf(file, "%s\n", cmd)
		file.Close()
		var out []byte
		if Current_OS == "windows" {
			bash_command := fmt.Sprintf("type %s | ncat %s %s", fn, Params.Host, Params.Port)
			out, err = exec.Command("cmd", "/C", bash_command).Output()
		} else {
			bash_command := fmt.Sprintf("cat %s | netcat %s %s", fn, Params.Host, Params.Port)
			out, err = exec.Command("bash", "-c", bash_command).Output()
		}
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Printf("Exec output is \n %s \n", string(out))
		split_out := strings.Split(string(out), "\n")
		if strings.TrimSpace(split_out[0]) == "SUCCESS" {
			return split_out[1], nil
		} else {
			fmt.Println(strings.TrimSpace(split_out[0]))
			return "", err
		}

	}
	if header == "SUCCESS\n" {
		return r.ReadString('\n')
	}
	return "", nil
}

func bffeeder(trips []trip) ([]trip_bf_out, error) {
	var tripsout []trip_bf_out

	for _, t := range trips {
		bf_out, err := bffeed(t)
		if err != nil {
			return tripsout, err
		}
		tripsout = append(tripsout, trip_bf_out{
			obvs:           bf_out.obvs,
			prior_stop:     t.prior_stop,
			following_stop: t.following_stop,
			id:             t.tripid,
		})
	}
	return tripsout, nil

}

func tojsonobv(obvs []obv, tripid string) []json_in {
	var jsonobv []json_in

	for _, o := range obvs {

		js := json_in{
			Point:    fmt.Sprintf("POINT(%g %g)", o.point.Lon(), o.point.Lat()),
			Datetime: o.datetime * 1000,
			Azimuth:  o.azimuth,
			Id:       tripid,
		}
		//fmt.Println(js)
		jsonobv = append(jsonobv, js)
	}

	return jsonobv
}

func to_json(t trip) []string {
	var cmds []string
	var subtrip []obv
	var subtrip_json []json_in
	l := len(t.obvs) - 1
	//fmt.Println(l)
	for i, o := range t.obvs {
		if (i != 0) && (o.new_subtrip || i == l) {
			subtrip_id := fmt.Sprintf("%s.%d", t.tripid, i)
			subtrip_json = tojsonobv(subtrip, subtrip_id)
			cmd_byte, _ := json.Marshal(subtrip_json)
			cmd := string(cmd_byte)
			cmd = fmt.Sprintf("{\"format\": \"%s\", \"request\": %s}", Params.Format, cmd)
			cmds = append(cmds, cmd)
			subtrip = nil
			subtrip = append(subtrip, o)

		} else {
			subtrip = append(subtrip, o)
		}
	}
	return cmds
}

func merge_orig(trip trip, matched_trip []Json_out) trip_bf_out {
	var left_overs []Json_out
	var matched_trip_out trip_bf_out
	var add_obv Json_out
	var add_obvs []Json_out
	for _, o := range trip.obvs {
		not_matched := true
		for _, m := range matched_trip {
			//fmt.Printf("%g %g\n", o.datetime, m.Datetime)
			if o.datetime == m.Datetime {
				add_obv = m
				add_obv.Lon = o.point.Lon()
				add_obv.Lat = o.point.Lat()
				add_obv.Newsubtrip = o.new_subtrip
				add_obv.Speed = o.speed
				add_obvs = append(add_obvs, add_obv)
				//fmt.Println(add_obvs)
				not_matched = false

				break
			}
		}

		if not_matched {
			//fmt.Printf("notmatched %g\n", o.datetime)
			left_overs = append(left_overs, Json_out{
				Datetime:   o.datetime,
				Obv_type:   "not matched",
				Newsubtrip: o.new_subtrip,
				Lat:        o.point.Lat(),
				Lon:        o.point.Lon(),
				Speed:      o.speed,
			})
			//fmt.Printf("Not matched Lon %g, Lat %g\n", o.point.Lon(), o.point.Lat())
		}
	}
	matched_trip_out.obvs = append(add_obvs, left_overs...)

	//sorting by datetime
	sort.SliceStable(matched_trip, func(i, j int) bool { return matched_trip[i].Datetime < matched_trip[j].Datetime })

	matched_trip_out.prior_stop = trip.prior_stop
	matched_trip_out.following_stop = trip.following_stop
	return matched_trip_out

}
