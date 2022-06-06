package ys

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"strconv"
)

//second pass stuff should be to the barefoot feeder

//fix these struct fields so they're exported
//var Params Para

type obv struct {
	datetime    int64
	point       orb.Point
	speed       float64
	id          string
	new_subtrip bool
	azimuth     float64
}

type aggdict struct {
	clustercentre  orb.Point
	lastobv        obv
	pottripobvs    []obv
	potstopobvs    []obv
	stops          []stop
	trips          []trip
	id             string
	dump_to_resids []obv
	passed_first   bool
}

type stop struct {
	start  int64
	end    int64
	stopid string
	id     string
	point  orb.Point
}

type trip struct {
	obvs           []obv
	tripid         string
	id             string
	prior_stop     string
	following_stop string
}

type vehpack struct {
	trips     []trip
	stops     []stop
	residuals []obv
	assetid   string
}

///consider adding speed in again
func checkdupe(obv1 obv, obv2 obv) bool {
	return (obv1.point.Lat() == obv2.point.Lat() &&
		obv1.point.Lon() == obv2.point.Lon() &&
		(obv1.speed == obv2.speed))

}

//add lastobv for all

func inpotstop(obv obv, agg aggdict) aggdict {
	agg.potstopobvs = append(agg.potstopobvs, obv)
	l := float64(len(agg.potstopobvs))
	cent_lat := (l*agg.clustercentre.Lat() + agg.lastobv.point.Lat()) / (l + 1)
	cent_lon := (l*agg.clustercentre.Lon() + agg.lastobv.point.Lon()) / (l + 1)
	//fmt.Println(cent_lat)
	agg.clustercentre = orb.Point{cent_lon, cent_lat}
	agg.lastobv = obv
	return agg
}

//second pass stuff here?

func nostop(obv obv, agg aggdict) aggdict {

	agg.clustercentre = obv.point
	agg.pottripobvs = append(agg.pottripobvs, agg.potstopobvs...)
	agg.pottripobvs = append(agg.pottripobvs, obv)
	//fmt.Println(len(agg.pottripobvs))
	agg.potstopobvs = nil
	agg.lastobv = obv
	return agg
}

func isstop(obv obv, agg aggdict, drop_first_stop bool) aggdict {
	//fmt.Println("is stop")
	l := len(agg.stops)
	agg.potstopobvs = append(agg.potstopobvs, obv)
	if !agg.passed_first && drop_first_stop {
		fmt.Println("Dropping first stop")
		agg.dump_to_resids = append(agg.potstopobvs, agg.pottripobvs...)
		agg.passed_first = true
	} else {
		interstopdur := Params.StopCollateDuration + 1
		interstopdist := Params.StopDistance + 1
		if l > 0 {

			interstopdur = agg.potstopobvs[0].datetime - agg.stops[l-1].end
			interstopdist = geo.DistanceHaversine(agg.clustercentre, agg.stops[l-1].point)
		}
		if interstopdist < Params.StopDistance && interstopdur < Params.StopCollateDuration {
			/// adjust centroid
			agg.stops[l-1].end = agg.potstopobvs[len(agg.potstopobvs)-1].datetime

		} else {
			stopid := agg.id + "_" + strconv.FormatInt(agg.potstopobvs[0].datetime, 16)
			stop := stop{
				start:  agg.potstopobvs[0].datetime,
				end:    agg.potstopobvs[len(agg.potstopobvs)-1].datetime,
				stopid: stopid,
				id:     obv.id,
				point:  agg.clustercentre,
			}
			agg.stops = append(agg.stops, stop)
			//fmt.Println(agg.stops)

			if len(agg.pottripobvs) > 0 {
				agg.pottripobvs = append(agg.pottripobvs, agg.potstopobvs[0])

				tripid := agg.id + "_" + strconv.FormatInt(agg.pottripobvs[0].datetime, 16)

				prior_stop := "NA"
				if len(agg.stops) > 1 {
					prior_stop = agg.stops[len(agg.stops)-2].stopid
				} else if !drop_first_stop {
					prior_stop = checkPriorStop(obv.id, agg.pottripobvs[0].datetime)
				}
				//add prior following stop ids
				trip := trip{
					obvs:           agg.pottripobvs,
					tripid:         tripid,
					id:             obv.id,
					prior_stop:     prior_stop,
					following_stop: agg.stops[len(agg.stops)-1].stopid,
				}
				agg.trips = append(agg.trips, trip)
			}

		}
	}

	agg.clustercentre = obv.point

	agg.potstopobvs = nil
	agg.pottripobvs = nil
	agg.pottripobvs = append(agg.pottripobvs, agg.lastobv)
	agg.lastobv = obv
	agg.pottripobvs = append(agg.pottripobvs, obv)
	return agg
}

//changing parameters into seconds
func cichiter(agg aggdict, obv obv, drop_first_stop bool) aggdict {
	if obv.datetime == agg.lastobv.datetime {
		//fmt.Println("double time stamp")
		return agg
	} else if Params.SkipDupes {
		if checkdupe(agg.lastobv, obv) {
			//fmt.Println("Dupe")
			return agg
		}
	} else {

		dist := geo.DistanceHaversine(obv.point, agg.lastobv.point)
		//fmt.Println(agg.lastobv.datetime)
		//fmt.Println(obv.datetime)
		diff_time := obv.datetime - agg.lastobv.datetime

		imp_speed := (dist / 1000) / (float64(diff_time) / 3600)

		if imp_speed > Params.MaxSpeed {
			return agg //erroneous jump, discard
		}

		if dist < Params.StopDistance {
			//fmt.Println("In pot stop")
			agg = inpotstop(obv, agg)
		} else if len(agg.potstopobvs) > 0 {
			time := agg.lastobv.datetime - agg.potstopobvs[0].datetime

			if dist > Params.MaxDist || (obv.datetime-agg.lastobv.datetime) > Params.MaxTime {
				obv.new_subtrip = true
				//fmt.Printf("subtrip md %g mt %d d %g t %d\n", Params.MaxDist, Params.MaxTime, dist, time)
			}

			if time > Params.StopDuration {
				//fmt.Println("is stop")
				agg = isstop(obv, agg, drop_first_stop)
			} else {
				//fmt.Println("no stop")
				agg = nostop(obv, agg)
			}

		} else {
			//fmt.Println("no stop")
			agg = nostop(obv, agg)
		}
	}
	//fmt.Println("iter")
	return agg

}

func CichCluster(obvs []obv, id string, drop_first_stop bool) vehpack {
	fmt.Printf(Cyan+"Tripgrouping %s with %d observations \n"+Reset, id, len(obvs))
	//fmt.Println(len(obvs))
	agg := aggdict{clustercentre: obvs[0].point,
		lastobv: obvs[0],
		id:      id,
	}
	//this should be appending to stops and trips
	for _, obv := range obvs[1:] {
		agg = cichiter(agg, obv, drop_first_stop)

	}
	residuals := append(agg.potstopobvs, agg.pottripobvs...)
	residuals = append(residuals, agg.dump_to_resids...)
	return vehpack{stops: agg.stops, trips: agg.trips, residuals: residuals, assetid: id}

}
