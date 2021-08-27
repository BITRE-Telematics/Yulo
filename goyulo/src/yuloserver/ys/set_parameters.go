package ys

import (
	"./yaml"
	"github.com/gosexy/to"
)

var Creds *yaml.Yaml

type Para struct {
	Max_routines         int64
	Max_memory           float64
	Max_cpu              float64
	Residual_dir         string
	Yuloport             string
	stopDuration         int64
	stopDistance         float64
	stopCollateDuration  int64
	maxSpeed             float64
	maxTime              int64
	timeThresh           int64
	distThresh           int64
	maxDist              float64
	skipDupes            bool
	maxResidsGap         int64
	Max_stop_gap         int64
	router_port          string
	host                 string
	port                 string
	format               string
	Creds                string
	STE                  string
	SA2                  string
	file_transfer_server string
	Error_log            string
	Match_locs           bool
	max_loc_dist         int
}

func Set_parameters() Para {
	yml, _ := yaml.Open("config.yaml")
	Params := Para{
		Max_routines:         to.Int64(yml.Get("max_routines")),
		Max_memory:           to.Float64(yml.Get("max_memory")),
		Max_cpu:              to.Float64(yml.Get("max_cpu")),
		Residual_dir:         to.String(yml.Get("residualDir")),
		Yuloport:             to.String(yml.Get("yuloport")),
		stopDuration:         to.Int64(yml.Get("stopDuration")),
		stopDistance:         to.Float64(yml.Get("stopDistance")),
		stopCollateDuration:  to.Int64(yml.Get("stopCollateDuration")),
		maxSpeed:             to.Float64(yml.Get("maxSpeed")),
		maxTime:              to.Int64(yml.Get("maxTime")),
		timeThresh:           to.Int64(yml.Get("timeThresh")),
		distThresh:           to.Int64(yml.Get("distThresh")),
		maxDist:              to.Float64(yml.Get("maxDist")),
		skipDupes:            to.Bool(yml.Get("skipDupes")),
		maxResidsGap:         to.Int64(yml.Get("maxResidsGap")),
		Max_stop_gap:         to.Int64(yml.Get("maxStopGap")),
		router_port:          to.String(yml.Get("router_port")),
		host:                 to.String(yml.Get("host")),
		port:                 to.String(yml.Get("port")),
		format:               to.String(yml.Get("format")),
		Creds:                to.String(yml.Get("creds")),
		STE:                  to.String(yml.Get("STE")),
		SA2:                  to.String(yml.Get("SA2")),
		file_transfer_server: to.String(yml.Get("file_transfer_server")),
		Error_log:            to.String(yml.Get("error_log")),
		Match_locs:           to.Bool(yml.Get("match_locs")),
		max_loc_dist:         to.Int(yml.Get("max_loc_dist")),
	}
	return Params
}
