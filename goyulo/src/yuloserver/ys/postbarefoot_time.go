package ys

import (
	"fmt"
	"time"
)

func add_times_obv(obv processedObv) processedObv {

	olson := make(map[string]string)

	olson["1"] = "Australia/Sydney"
	olson["2"] = "Australia/Melbourne"
	olson["3"] = "Australia/Brisbane"
	olson["4"] = "Australia/Adelaide"
	olson["5"] = "Australia/Perth"
	olson["6"] = "Australia/Hobart"
	olson["7"] = "Australia/Darwin"
	olson["8"] = "Australia/Sydney"
	olson["9"] = "Australia/Sydney"
	olson["0"] = "UTC"

	tz := olson[obv.STE]
	if tz == "" {
		fmt.Println(obv.Osm_id)
		fmt.Println(obv.STE)
		panic("Fucked up tz")
	}

	obv.Datetime_utc, obv.Datetime_utcdt = tsConvert(obv.Datetime, "UTC")
	obv.Datetime, obv.Datetime_dt = tsConvert(obv.Datetime, tz)
	return obv

}

func timeIn(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

func tsConvert(ts int64, tz string) (int64, string) {
	dt := time.Unix(ts, int64(0))
	dt_local, _ := timeIn(dt, tz)
	//there won't be a local unix epoch yet
	dt_local_st := dt_local.Format("2006-01-02T15:04:05") + "[" + tz + "]"
	//fmt.Println(dt_local_st)
	return dt_local.Unix(), dt_local_st

}
