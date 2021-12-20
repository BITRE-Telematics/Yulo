package ys

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var Creds Cred_struct

type Para struct {
	Max_routines         int64   `yaml:"max_routines"`
	Max_memory           float64 `yaml:"max_memory"`
	Max_cpu              float64 `yaml:"max_cpu"`
	Residual_dir         string  `yaml:"residualDir"`
	Yuloport             string  `yaml:"yuloport"`
	StopDuration         int64   `yaml:"stopDuration"`
	StopDistance         float64 `yaml:"stopDistance"`
	StopCollateDuration  int64   `yaml:"stopCollateDuration"`
	MaxSpeed             float64 `yaml:"maxSpeed"`
	MaxTime              int64   `yaml:"maxTime"`
	TimeThresh           int64   `yaml:"timeThresh"`
	DistThresh           int64   `yaml:"distThresh"`
	MaxDist              float64 `yaml:"maxDist"`
	SkipDupes            bool    `yaml:"skipDupes"`
	MaxResidsGap         int64   `yaml:"maxResidsGap"`
	Max_stop_gap         int64   `yaml:"maxStopGap"`
	Router_port          string  `yaml:"router_port"`
	Host                 string  `yaml:"host"`
	Port                 string  `yaml:"port"`
	Format               string  `yaml:"format"`
	Creds                string  `yaml:"creds"`
	STE                  string  `yaml:"STE"`
	SA2                  string  `yaml:"SA2"`
	File_transfer_server string  `yaml:"file_transfer_server"`
	Error_log            string  `yaml:"error_log"`
	Match_locs           bool    `yaml:"match_locs"`
	Max_loc_dist         int     `yaml:"max_loc_dist"`
}

type Cred_struct struct {
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Ipport       string `yaml:"ipport"`
	Ipporthttp   string `yaml:"importhttp"`
	Bolt         string `yaml:"bolt"`
	Db_name      string `yaml:"db_name"`
	Transfer_key string `yaml:"transferkey"`
}

func Set_parameters() Para {
	configfile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
	}
	var Params Para
	err = yaml.Unmarshal(configfile, &Params)
	return Params
}

func Read_creds(credsfile string) Cred_struct {
	file, err := ioutil.ReadFile(credsfile)
	if err != nil {
		fmt.Println(err)
	}
	var creds Cred_struct
	err = yaml.Unmarshal(file, &creds)
	return creds
}
