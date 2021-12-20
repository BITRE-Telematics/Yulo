package ys

import (
	"fmt"
	//"github.com/paulmach/orb"
	//"github.com/paulmach/orb/geojson"
	//"github.com/paulsmith/gogeos/geos"
)

var SA2 *[]Geog

type processedStop struct {
	Stopid      string
	Vehicle     string
	Start       int64
	Start_utc   int64
	Start_utcdt string
	Startdt     string
	End         int64
	End_utc     int64
	End_utcdt   string
	Enddt       string
	Lat         float64
	Lon         float64
	Sa2         string
	Gcc         string
	Loc         string
	Addr        string
}

func add_times_stop(stop processedStop) processedStop {

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

	tz := olson[string(stop.Sa2[0])]

	stop.Start_utc, stop.Start_utcdt = tsConvert(stop.Start, "UTC")
	stop.Start, stop.Startdt = tsConvert(stop.Start, tz)

	stop.End_utc, stop.End_utcdt = tsConvert(stop.End, "UTC")
	stop.End, stop.Enddt = tsConvert(stop.End, tz)
	return stop
}

func geocode_stop(stop stop, last_index int) (processedStop, int) {
	stopout := processedStop{
		Stopid:  stop.stopid,
		Vehicle: stop.id,
		Start:   stop.start,
		End:     stop.end,
		Lat:     stop.point.Lat(),
		Lon:     stop.point.Lon(),
	}

	//pt := orb.Point{stop.point.Lon(), stop.point.Lat()}
	//var last_index_out int
	// stopout.Sa2, stopout.Gcc, last_index_out = pointInPoly(pt, SA2, "SA2_MAIN16", last_index)
	// //fmt.Println(stopout.Gcc)
	// if stopout.Sa2 == "NA" {
	// 	stopout.Sa2, stopout.Gcc = nearestPoly(pt, SA2, "SA2_MAIN16")
	// }
	last_index_out := match_point(stopout.Lon, stopout.Lat, SA2, last_index)
	stopout.Sa2 = (*SA2)[last_index_out].SA2
	stopout.Gcc = (*SA2)[last_index_out].GCC

	return stopout, last_index_out
}

func sum_stops(stops []stop) []processedStop {
	fmt.Println("Summarising stops")
	var stopsout []processedStop
	last_index := int(0)
	s_out := processedStop{}
	for _, s := range stops {
		s_out, last_index = geocode_stop(s, last_index)
		s_out = add_times_stop(s_out)
		if Params.Match_locs {
			s_out = match_locs(s_out)
		}
		stopsout = append(stopsout, s_out)
	}
	return stopsout
}
