package ys

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"strconv"
)

//obv contains data for a given observation
type obv struct {
	datetime    int64
	point       orb.Point
	speed       float64
	id          string
	new_subtrip bool
	azimuth     float64
}

//aggdict is an aggregation object as part of the tripgrouping process. It contains the current state of the trip grouping
//including determined trips and stops to that point and potential trip and stop data for assessing the next observation
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

//stop objects contain data on determined stops
type stop struct {
	start  int64
	end    int64
	stopid string
	id     string
	point  orb.Point
}

//trip objects include data on determined trips
type trip struct {
	obvs           []obv
	tripid         string
	id             string
	prior_stop     string
	following_stop string
}

//vehpack objects include data for a given vehicle after trip grouping
type vehpack struct {
	trips     []trip
	stops     []stop
	residuals []obv
	assetid   string
}

//checkdupe determines if the two indentical locations have been returned consecutively
//this should not occur except by error
func checkdupe(obv1 obv, obv2 obv) bool {
	return (obv1.point.Lat() == obv2.point.Lat() &&
		obv1.point.Lon() == obv2.point.Lon() &&
		(obv1.speed == obv2.speed))

}

//inpotstop updates the aggdict object when a observation is determined to be in a potential stop
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

//nostop updates the aggdict object when an observation is determined not to be part of
//a stop but the potential stops observations do not constitute a stop
func nostop(obv obv, agg aggdict) aggdict {

	agg.clustercentre = obv.point
	agg.pottripobvs = append(agg.pottripobvs, agg.potstopobvs...)
	agg.pottripobvs = append(agg.pottripobvs, obv)
	//fmt.Println(len(agg.pottripobvs))
	agg.potstopobvs = nil
	agg.lastobv = obv
	return agg
}

//istop updates an aggdict object when an observation is determined to be outside a stop event
//and the potential stop is determined to be a stop
func isstop(obv obv, agg aggdict, drop_first_stop bool, parameters Para) aggdict {
	//fmt.Println("is stop")
	l := len(agg.stops)
	agg.potstopobvs = append(agg.potstopobvs, obv)
	if !agg.passed_first && drop_first_stop {
		fmt.Println("Dropping first stop")
		agg.dump_to_resids = append(agg.potstopobvs, agg.pottripobvs...)
		agg.passed_first = true
	} else {
		interstopdur := parameters.StopCollateDuration + 1
		interstopdist := parameters.StopDistance + 1
		if l > 0 {

			interstopdur = agg.potstopobvs[0].datetime - agg.stops[l-1].end
			interstopdist = geo.DistanceHaversine(agg.clustercentre, agg.stops[l-1].point)
		}
		if interstopdist < parameters.StopDistance && interstopdur < parameters.StopCollateDuration {
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

//cichiter performs an iteration of the clustering process on a given observation in combination
//with the aggregated data up to that point, determining whether each observation is
//part of a trip or stop event
func cichiter(agg aggdict, obv obv, drop_first_stop bool, parameters Para) aggdict {
	if obv.datetime == agg.lastobv.datetime {
		//fmt.Println("double time stamp")
		return agg
	} else if parameters.SkipDupes {
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

		if imp_speed > parameters.MaxSpeed {
			return agg //erroneous jump, discard
		}

		if dist < parameters.StopDistance {
			//fmt.Println("In pot stop")
			agg = inpotstop(obv, agg)
		} else if len(agg.potstopobvs) > 0 {
			time := agg.lastobv.datetime - agg.potstopobvs[0].datetime

			if dist > parameters.MaxDist || (obv.datetime-agg.lastobv.datetime) > parameters.MaxTime {
				obv.new_subtrip = true
				//fmt.Printf("subtrip md %g mt %d d %g t %d\n", parameters.MaxDist, parameters.MaxTime, dist, time)
			}

			if time > parameters.StopDuration {
				//fmt.Println("is stop")
				agg = isstop(obv, agg, drop_first_stop, parameters)
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

//CichCluster takes a slice of observations and groups them into trips, stops and residuals
func CichCluster(obvs []obv, id string, drop_first_stop bool, parameters Para) vehpack {
	fmt.Printf(Cyan+"Tripgrouping %s with %d observations \n"+Reset, id, len(obvs))
	//fmt.Println(len(obvs))
	agg := aggdict{clustercentre: obvs[0].point,
		lastobv: obvs[0],
		id:      id,
	}
	//this should be appending to stops and trips
	for _, obv := range obvs[1:] {
		agg = cichiter(agg, obv, drop_first_stop, parameters)

	}
	residuals := append(agg.potstopobvs, agg.pottripobvs...)
	residuals = append(residuals, agg.dump_to_resids...)
	return vehpack{stops: agg.stops, trips: agg.trips, residuals: residuals, assetid: id}

}
