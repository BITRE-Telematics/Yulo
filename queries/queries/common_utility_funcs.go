package queries

import (
	"encoding/csv"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
)

var Fabric string
var Seg_db string
var Activityfile string
var Year int64
var MinDur int64
var Month int64
var Act_type string
var err_c chan []string
var Db neo4j.Driver
var Sesh_config neo4j.SessionConfig
var Max_routines int64

type Cred_struct struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Ipport       string `yaml:"ipport"`
	Ipporthttp   string `yaml:"importhttp"`
	Bolt         string `yaml:"bolt"`
	Db_name      string `yaml:"db_name"`
	Transfer_key string `yaml:"transferkey"`
	Fabric       string `yaml:"fabric_db"`
	Segs_db      string `yaml:"segs_db"`
	Locs_db      string `yaml:"locs_db"`
}

func Read_creds(credsfile string) Cred_struct {
	file, err := ioutil.ReadFile(credsfile)
	if err != nil {
		fmt.Println(err)
	}
	var creds Cred_struct
	err = yaml.Unmarshal(file, &creds)
	return creds
}

//convenience funcs for accessing neo4j records
func floatGet(record *neo4j.Record, key string) float64 {
	var v float64
	if value, ok := record.Get(key); ok {
		if value != nil {
			v = value.(float64)
		}
	}
	return v

}

func intGet(record *neo4j.Record, key string) int64 {
	var v int64
	if value, ok := record.Get(key); ok {
		if value != nil {
			v = value.(int64)
		}
	}
	return v

}

func stringGet(record *neo4j.Record, key string) string {
	var v string
	if value, ok := record.Get(key); ok {
		if value != nil {
			v = value.(string)
		}
	}
	return v

}

//convenience functions for resuming segments
func seg_writer(w *csv.Writer, c chan []string) {
	for l := range c {
		if len(l) > 0 {
			w.Write(l)
			w.Flush()
		}

	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	//fmt.Println(e)
	return false
}

func ext_undone(done []string, osm_id string, c chan string, onExit func()) {
	go func() {
		defer onExit()
		if !contains(done, osm_id) {
			c <- osm_id
			//fmt.Println(osm_id)
		}
	}()
}

//formatting funcs
func azi_to_string(azi float64) string {
	var d string
	switch {
	case azi > 337.5 || azi <= 22.5:
		d = "N"
	case azi > 22.5 && azi <= 67.5:
		d = "NE"
	case azi > 67.5 && azi <= 112.5:
		d = "E"
	case azi > 112.5 && azi <= 157.5:
		d = "SE"
	case azi > 157.5 && azi <= 202.5:
		d = "S"
	case azi > 202.5 && azi <= 247.5:
		d = "SW"
	case azi > 247.5 && azi <= 292.5:
		d = "W"
	case azi > 292.5 && azi <= 337.5:
		d = "NW"
	}
	return d
}

func floatFrmt(record *neo4j.Record, key string) string {
	str := "0"
	if value, ok := record.Get(key); ok {
		if value != nil {
			str = strconv.FormatFloat(value.(float64), 'f', 2, 64)
		}
	}
	return str

}

func unspecifiedNumFrmt(record *neo4j.Record, key string) string {
	var length_str string
	l, _ := record.Get(key)
	switch l.(type) {
	case float64:
		length_str = floatFrmt(record, key)

	case int64:
		length_str = strconv.FormatInt(intGet(record, key), 10)
	}
	return length_str
}

//for filling vectors of precomputed values
func get_min_max_bd(Bd_type string) (int64, int64) {
	var min_bd int64
	var max_bd int64
	switch Bd_type {
	case "hour":
		min_bd, max_bd = 0, 23
	case "dayOfWeek":
		min_bd, max_bd = 1, 7
	case "month":
		min_bd, max_bd = 1, 12
	}
	return min_bd, max_bd
}

func fill_negs_int(bd []int64) []int64 {
	for i := range bd {
		bd[i] = -1
	}
	return bd
}

func fill_negs_float(bd []float64) []float64 {
	for i := range bd {
		bd[i] = -1
	}
	return bd
}
