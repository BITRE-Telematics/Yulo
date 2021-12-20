package ys

import (
  //"fmt"
  "github.com/shirou/gopsutil/cpu"
  "runtime"
  "time"
)

func ReturnMemUse() float64 {
  var m runtime.MemStats
  runtime.ReadMemStats(&m)
  usage := float64(m.Alloc) / float64(m.Sys)
  return usage
}

func check_resources(c chan struct{}, memlimit float64, cpulimit float64) bool {
  var memusage float64
  var cpuusage []float64
  for {
    memusage = ReturnMemUse()
    cpuusage, _ = cpu.Percent(time.Second/4, false)
    cpuusage_ := cpuusage[0]
    if memusage < memlimit && cpuusage_ < cpulimit {
      return true
    }
  }
}
