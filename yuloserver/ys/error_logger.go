package ys

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

var Error_chan chan Error_line

type Error_line struct {
	id    string
	err   error
	stage string
}

func Error_logger(c chan Error_line) {
	file, fileerr := os.OpenFile(Params.Error_log, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)

	if fileerr != nil {
		fmt.Printf("Log file error: %s", fileerr)
	}
	w := csv.NewWriter(file)

	for e := range c {
		l := []string{
			e.id,
			e.err.Error(),
			e.stage,
			time.Now().String(),
		}
		w.Write(l)
		w.Flush()
	}
}
