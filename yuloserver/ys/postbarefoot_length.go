package ys

import (
	//"encoding/json"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	//"strconv"
)

//query_route_length sends a query to the osrm routing server where a distance was not returned by barefoot
func query_route_length(p1 orb.Point, p2 orb.Point) float64 {
	url := fmt.Sprintf("http://127.0.0.1:%s/route/v1/driving/%g,%g;%g,%g?steps=true", Params.Router_port, p1.Lon(), p1.Lat(), p2.Lon(), p2.Lat())

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Route error")
		return (0.0)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	bodystr := string(body)

	var distance float64
	if gjson.Get(bodystr, "code").String() == "Ok" {
		distance = gjson.Get(bodystr, "routes.0.distance").Float()
		//fmt.Printf("Returned distance of %g\n", distance)
	}

	return (distance)
}
