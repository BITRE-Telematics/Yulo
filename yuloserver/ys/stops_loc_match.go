package ys

//functions for making and using rtree and matchign stops to locations
import (
	"fmt"
	//"github.com/xiam/to"
	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/points"
	//"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"strconv"
	"strings"
)

var Addr_tree *kdtree.KDTree
var Ra_tree *kdtree.KDTree
var Loc_ra []loc
var Loc_addr []loc

type loc struct {
	lon, lat float64
	id       string
}

type Data struct {
	value int
}

func get_locs(loc_type string) []loc {
	//fmt.Println("Creating session")
	session := Db.NewSession(Sesh_config)

	defer session.Close()
	loc_cypher := `
	 MATCH (l:%s)
	 WHERE l.lat IS NOT NULL AND l.lat <0
	 RETURN l.lat as lat, l.lon as lon, l.id as loc_id`
	loc_cypher = fmt.Sprintf(loc_cypher, loc_type)
	//fmt.Println("Running query")
	locquery, err := session.Run(loc_cypher,
		map[string]interface{}{})

	if err != nil {
		fmt.Println(err)
	}

	if locquery.Err() != nil {
		fmt.Println(locquery.Err())
	}
	fmt.Printf("Building %s list\n", loc_type)
	var locs []loc
	var loc loc
	for locquery.Next() {
		loc.lat = locquery.Record().GetByIndex(0).(float64)
		loc.lon = locquery.Record().GetByIndex(1).(float64)
		loc.id = locquery.Record().GetByIndex(2).(string)
		locs = append(locs, loc)

	}
	fmt.Printf("returning %v %s\n", len(locs), loc_type)
	return locs
}

func Make_tree(loc_type string) (*kdtree.KDTree, []loc) {
	locs := get_locs(loc_type)
	fmt.Printf("Making %s tree\n", loc_type)
	t := kdtree.New([]kdtree.Point{})
	for i, l := range locs {
		t.Insert(points.NewPoint([]float64{l.lon, l.lat}, Data{value: i}))

	}

	return t, locs
}

func check_nearest_tree(stop processedStop, tree *kdtree.KDTree) int {
	nearest := tree.KNN(&points.Point{Coordinates: []float64{stop.Lon, stop.Lat}}, 1)
	s := fmt.Sprintf("%v", nearest[0])
	s = strings.SplitAfter(s[:len(s)-2], "{")[2]
	s_int, _ := strconv.Atoi(s)
	return s_int
}

func check_dist(stop processedStop, loc loc, max int) string {
	orb_point := orb.Point{stop.Lon, stop.Lat}
	orb_loc := orb.Point{loc.lon, loc.lat}
	dist := geo.DistanceHaversine(orb_point, orb_loc)
	//fmt.Printf("%v %v\n", dist, orb_loc)
	//fmt.Println(loc.id)
	if dist <= float64(max) {
		//fmt.Println(loc.id)
		return loc.id
	} else {
		return ""
	}
}

func match_locs(stop processedStop) processedStop {
	if Ra_tree != nil {
		//fmt.Println("Matching to locations")
		i_ra := check_nearest_tree(stop, Ra_tree)
		id_ra := check_dist(stop, Loc_ra[i_ra], Params.Max_loc_dist)
		stop.Loc = id_ra
	}
	if Addr_tree != nil {
		//fmt.Println("Matching to Addresses")
		i_addr := check_nearest_tree(stop, Addr_tree)
		id_addr := check_dist(stop, Loc_addr[i_addr], Params.Max_loc_dist)
		stop.Addr = id_addr
	}
	//fmt.Println("Finished Matching")
	return stop
}
