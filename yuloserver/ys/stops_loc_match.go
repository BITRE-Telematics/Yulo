package ys

//functions for making and using rtree and matchign stops to locations
import (
	"fmt"
	//"github.com/xiam/to"
	"database/sql"
	"github.com/kyroy/kdtree"
	"github.com/kyroy/kdtree/points"
	_ "github.com/mattn/go-sqlite3"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/xitongsys/parquet-go-source/local"
	//"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/writer"
	"os"
	"strconv"
	"strings"
)

// Addr_tree stores an r tree of addresses to be matched to
var Addr_tree *kdtree.KDTree

// Ra_tree stores an r tree of rest area locations to be matched to
var Ra_tree *kdtree.KDTree

// Loc_ra contains rest area locations
var Loc_ra []loc

// Loc_addr contains address locations
var Loc_addr []loc

// loc contains a location stores as a point for inserting into rtrees
type loc struct {
	lon, lat float64
	id       string
}

// Data is a convenience type for inserting locs into rtrees
type Data struct {
	value int
}

// get_locs retrives either addess or rest areas from the database
func get_locs(loc_type string) []loc {
	var locs []loc
	//first check cache
	locs = read_cache_pq(loc_type)
	if len(locs) > 0 {
		fmt.Printf("returning %v %s\n", len(locs), loc_type)
		return locs
	}
	//else query db
	fmt.Println("No cached locs, querying main db")
	locs = query_db_locs(loc_type)
	fmt.Printf("returning %v %s\n", len(locs), loc_type)
	return locs
}

func query_db_locs(loc_type string) []loc {
	var locs []loc
	sesh_config_locs := neo4j.SessionConfig{
		DatabaseName: Creds.Locs_db,
	}
	session := Db.NewSession(sesh_config_locs)

	defer session.Close()
	loc_cypher := `
	 MATCH (l:%s)
	 WHERE l.lat IS NOT NULL AND l.lat <0
	 RETURN l.lat as lat, l.lon as lon, toString(l.id) as loc_id`
	loc_cypher = fmt.Sprintf(loc_cypher, loc_type)
	//fmt.Println("Running query")
	locquery, err := session.Run(loc_cypher,
		map[string]interface{}{})

	if err != nil {
		fmt.Println(err)
	}
	if locquery == nil {
		fmt.Println("nil result")
		return locs
	}

	if locquery.Err() != nil {
		fmt.Println(locquery.Err())
	}
	fmt.Printf("Building %s list\n", loc_type)

	var loc loc
	for locquery.Next() {
		//fmt.Println(locquery.Record().GetByIndex(2))
		loc.lat = locquery.Record().GetByIndex(0).(float64)
		loc.lon = locquery.Record().GetByIndex(1).(float64)
		loc.id = locquery.Record().GetByIndex(2).(string)
		locs = append(locs, loc)

	}

	//write cache
	write_cache_pq(loc_type, locs)
	return locs

}

// Make_tree retrives locations from the database and builds an rtree for quick location matching
func Make_tree(loc_type string) (*kdtree.KDTree, []loc) {
	locs := get_locs(loc_type)
	fmt.Printf("Making %s tree\n", loc_type)
	t := kdtree.New([]kdtree.Point{})
	for i, l := range locs {
		t.Insert(points.NewPoint([]float64{l.lon, l.lat}, Data{value: i}))

	}

	return t, locs
}

// check_nearest_tree matches a processedStop to the nearest location in a given rtree
func check_nearest_tree(stop processedStop, tree *kdtree.KDTree) int {
	nearest := tree.KNN(&points.Point{Coordinates: []float64{stop.Lon, stop.Lat}}, 1)
	s := fmt.Sprintf("%v", nearest[0])
	s = strings.SplitAfter(s[:len(s)-2], "{")[2]
	s_int, _ := strconv.Atoi(s)
	return s_int
}

// check_dist checks if the distance between a stop and a given location is less than a given value
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

// match_locs adds the nearest rest_area and address location, if below the threshold distance, to a processed stop
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

func read_cache(loc_type string) []loc {
	db_file := "loc_cache.db"
	db, err := sql.Open("sqlite3", db_file)
	if err != nil {
		fmt.Println(err)
		return []loc{}
	}
	rows, err := db.Query("SELECT * FROM " + loc_type)
	if err != nil {
		return []loc{}
	}
	defer rows.Close()
	var locs []loc
	for rows.Next() {
		var loc_id string
		var lon float64
		var lat float64

		err := rows.Scan(&loc_id, &lon, &lat)
		if err != nil {
			return []loc{}
		}
		locs = append(locs, loc{id: loc_id, lon: lon, lat: lat})
	}
	return locs
}

func write_cache(loc_type string, locs []loc) {
	db_file := "loc_cache.db"
	db, err := sql.Open("sqlite3", db_file)
	if err != nil {
		fmt.Println(err)
		return
	}

	sql := `CREATE TABLE ` + loc_type + ` (
          loc_id STRING PRIMARY KEY,  
          lon REAL NOT NULL,
          lat REAL NOT NULL
       );`

	_, err = db.Exec(sql)
	if err != nil {
		fmt.Println(err)
		os.Remove(db_file)
		return
	}

	stmt, err := db.Prepare("INSERT OR IGNORE INTO " + loc_type + "(loc_id, lon, lat) values(?, ?, ?)")
	defer stmt.Close()
	for _, l := range locs {
		//fmt.Println(l)
		_, err = stmt.Exec(l.id, l.lon, l.lat)
		if err != nil {
			fmt.Println(err)
			os.Remove(db_file)
			return
		}
	}
	db.Close()
}

type Pq_loc struct {
	Lat float64 `parquet:"name=lat, type=DOUBLE"`
	Lon float64 `parquet:"name=lon, type=DOUBLE"`
	Id  string  `parquet:"name=loc_id, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
}

func read_cache_pq(loc_type string) []loc {
	cache_file := loc_type + "_cache.parquet"
	var locs []loc
	fr, err := local.NewLocalFileReader(cache_file)
	if err != nil {
		fmt.Println("Cannot open cache")
		return locs
	}
	defer fr.Close()
	pr, err := reader.NewParquetReader(fr, new(Pq_loc), 4)
	if err != nil {
		fmt.Println("Cannot open reader")
		return locs
	}
	defer pr.ReadStop()

	var pq_locs []Pq_loc

	for i := 0; i < int(pr.GetNumRows()); i++ {
		var loc Pq_loc
		if err = pr.Read(&loc); err != nil {
			fmt.Println("Read cache error")
			return locs
		}
		pq_locs = append(pq_locs, loc)
	}
	for _, l := range pq_locs {
		locs = append(locs, loc{id: l.Id, lon: l.Lon, lat: l.Lat})
	}
	return locs
}

func write_cache_pq(loc_type string, locs []loc) {
	cache_file := loc_type + "_cache.parquet"
	fw, err := local.NewLocalFileWriter(cache_file)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Cannot write cache")
		return
	}
	defer fw.Close()

	pw, err := writer.NewParquetWriter(fw, new(Pq_loc), 4)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Cannot open reader")
		return
	}
	defer pw.WriteStop()
	for _, l := range locs {
		l_pq := Pq_loc{Lat: l.lat, Lon: l.lon, Id: l.id}
		if err = pw.Write(l_pq); err != nil {
			fmt.Println(err)
			fmt.Println("Cache write error")
			return
		}
	}

}
