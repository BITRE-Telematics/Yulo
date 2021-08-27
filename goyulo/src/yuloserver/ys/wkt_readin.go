package ys

import (
	"encoding/csv"
	"fmt"
	"github.com/paulsmith/gogeos/geos"
	"io"
	"os"
)

type Geog struct {
	SA2  string
	STE  string
	GCC  string
	Geom *geos.Geometry
}

func Wkt_readin(wktfn string, ASGC_type string) *[]Geog {
	wktfile, err := os.OpenFile(wktfn, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	SA2r := csv.NewReader(wktfile)
	if err != nil {
		fmt.Println(err)
	}
	headers, _ := SA2r.Read()
	headermap := make(map[string]int)
	for i, h := range headers {
		headermap[h] = i
	}

	var sa2s []Geog
	for {
		line, csverr := SA2r.Read()
		if csverr == io.EOF {
			break
		} else {
			geom, _ := geos.FromWKT(line[headermap["wkt"]])
			var pt Geog
			if ASGC_type == "SA2" {
				pt = Geog{
					SA2:  line[headermap["SA2"]],
					GCC:  line[headermap["GCC"]],
					STE:  string(line[headermap["SA2"]][0]),
					Geom: geom,
				}
			} else if ASGC_type == "STE" {
				pt = Geog{
					STE:  line[headermap["SA2"]],
					Geom: geom,
				}

			}
			sa2s = append(sa2s, pt)
		}
	}
	return &sa2s
}
