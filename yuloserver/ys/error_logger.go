package ys

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

//Error_chan accepts errors from goroutines to write to disk
var Error_chan chan Error_line

//Error_line formats errors for writing to disc
type Error_line struct {
	id    string
	err   error
	stage string
}

//Error_logger accepts errors from Error_chan and writes to an error log
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
