package ys

import (
	//"fmt"
	//"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/paulmach/orb"
	//"github.com/paulsmith/gogeos/geos"
	"fmt"
	"strconv"
	"strings"
)

var STE *[]Geog

type processedImpObv struct {
	Osm_id  string
	Azimuth float64
	//Target string
	Forward bool
}

type processedObv struct {
	Datetime       int64
	Datetime_utc   int64
	Datetime_dt    string
	Datetime_utcdt string
	Imp_Obvs       []processedImpObv
	Speed          float64
	Imputed_speed  float64
	STE            string
	Azimuth        float64
	Point          orb.Point
	Osm_id         string
	Length         float64
	Lat            float64
	Lon            float64
	Forward        bool
}

type processedTrip struct {
	Trip           string
	Prior_stop     string
	Following_stop string
	Obvs           []processedObv
}

func update_seg_sa2(sa2 string, gcc string, osm_id string) {

	session := Db.NewSession(Sesh_config)

	defer session.Close()

	statement := `MATCH(s:Segment{osm_id:$OSM_ID})
				  SET s.sa2 =  $SA2, s.gcc = $GCC
				  WITH s
				  MERGE (s)-[:IN]->(sa2:SA2{sa2_code:$SA2})
				  WITH s
				  RETURN s.sa2
					`
	parameters := map[string]interface{}{
		"OSM_ID": osm_id,
		"SA2":    sa2,
		"GCC":    gcc,
	}

	res, err := session.Run(statement, parameters)

	if err != nil || res.Err() != nil {
		fmt.Printf("Error updating segment %s sa2\n", osm_id)
		fmt.Print(err)
		fmt.Print(res.Err())
	}
}

func pbfObv(obv Json_out, last_point orb.Point, last_index int, i int) (processedObv, int) {
	o_out := processedObv{
		Datetime:      obv.Datetime,
		Speed:         obv.Speed,
		Imputed_speed: obv.Imputed_speed,
		Azimuth:       obv.Imputed_azimuth,
		Point:         orb.Point{obv.Lon, obv.Lat},
		Osm_id:        obv.Osm_id,
		Length:        obv.Length,
		Lat:           obv.Lat,
		Lon:           obv.Lon,
		Forward:       obv.Forward,
	}

	if o_out.Osm_id == "" {
		o_out.Osm_id = "unknown"
	}
	var last_index_out int
	//fmt.Println(obv.SA2)
	if strings.HasPrefix(obv.SA2, "N") || obv.SA2 == "" || obv.Osm_id == "unknown" {
		//accepting lower accuracy for non matched observations
		if obv.Osm_id != "unknown" || (i == 0) {

			//fmt.Printf("resorting to matching for %s \n", obv.Osm_id)
			last_index_out = match_point(obv.Lon, obv.Lat, SA2, last_index)
			o_out.STE = (*SA2)[last_index_out].STE
			if o_out.STE == "" {
				fmt.Println("SA2 without STE is ", obv.SA2)
			}
			if obv.Osm_id != "unknown" {
				update_seg_sa2((*SA2)[last_index_out].SA2, (*SA2)[last_index_out].GCC, obv.Osm_id)
			}

		}

	} else {
		//fmt.Printf("Prematched with %s\n", obv.SA2)
		o_out.STE = string(obv.SA2[0])
		if o_out.STE == "N" {
			fmt.Println(obv.SA2)
		}
		//fmt.Println(o_out)
		last_index_out, _ = strconv.Atoi(o_out.STE)
		last_index_out = last_index_out - 1
	}

	o_out = add_times_obv(o_out)
	//fmt.Printf("Added times %s\n", o_out.Datetime_utcdt)
	var imp_obvs []processedImpObv
	for _, o := range obv.Roads {
		if o.Osm_id == "" {
			o.Osm_id = "unknown"
		}
		imp_obvs = append(imp_obvs, processedImpObv{
			Osm_id:  o.Osm_id,
			Azimuth: o.Imputed_azimuth,
			Forward: o.Forward,
		})

	}
	o_out.Imp_Obvs = imp_obvs

	if obv.Newsubtrip {
		o_out.Length = query_route_length(last_point, o_out.Point)
	}

	return o_out, last_index_out
}

func pbTrip(trip []Json_out, ps string, fs string, tripid string) processedTrip {

	tripout := processedTrip{
		Prior_stop:     ps,
		Following_stop: fs,
		Trip:           tripid,
	}

	var obvs []processedObv
	last_point := orb.Point{trip[0].Lon, trip[0].Lat}
	last_index := int(0)
	o_out := processedObv{}
	for i, o := range trip {
		o_out, last_index = pbfObv(o, last_point, last_index, i)
		obvs = append(obvs, o_out)
		last_point = o_out.Point
	}
	tripout.Obvs = obvs
	return (tripout)
}

func postbarefoot(trips []trip_bf_out) []processedTrip {
	var tripsout []processedTrip
	for _, t := range trips {

		tripout := pbTrip(t.obvs, t.prior_stop, t.following_stop, t.id)
		tripsout = append(tripsout, tripout)
		//fmt.Println(tripout)
	}
	return tripsout
}
