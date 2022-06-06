package ys

import (
	"fmt"
	//"github.com/paulmach/orb"
	//"github.com/paulmach/orb/geojson"
	//"github.com/paulmach/orb/planar"
	"github.com/paulsmith/gogeos/geos"
)

/*
Two notes.
It would probably be faster to replace the orb libraries with gogeos if possible.
The geojson cannot have ANY null geometries in them
*/

//the last index is a heuristic that lets tit check the last observation's polygon first

func match_point(x float64, y float64, Geogs *[]Geog, ind int) int {
	p, _ := geos.NewPoint(geos.NewCoord(x, y))
	if ind != -1 {

		o, err := p.Intersects((*Geogs)[ind].Geom)
		if err != nil {
			fmt.Println(err)
		}
		//checks the last match first
		if o {
			//fmt.Println(f.SA2)
			return ind
		}
	}
	//fmt.Println(p)
	for i, f := range *Geogs {
		//fmt.Println(f.Geom)
		if f.Geom != nil {
			o, err := p.Intersects(f.Geom)
			if err != nil {
				fmt.Println(err)
			}
			if o {
				//fmt.Println(f.SA2)
				return i
			}
		}
	}

	//nearest neighbour code kicks in if no match yet
	//fmt.Printf(Yellow+"Resorting to nearest neighbour with %s \n"+Reset, p)
	dist_min, err := p.Distance((*Geogs)[0].Geom)
	if err != nil {
		fmt.Println(err)
	}
	for i, f := range (*Geogs)[1:] {
		if f.Geom != nil {
			dist, err := p.Distance(f.Geom)
			if err != nil {
				fmt.Println(err)
			}
			if dist < dist_min {
				dist_min = dist
				ind = i
			}
		}
	}
	return ind
}
