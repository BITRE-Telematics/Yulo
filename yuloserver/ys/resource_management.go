package ys

import (
  //"fmt"
  "github.com/shirou/gopsutil/cpu"
  "runtime"
  "time"
)

//ReturnMemUse returns the proportion of system memory being used as a percentage
func ReturnMemUse() float64 {
  var m runtime.MemStats
  runtime.ReadMemStats(&m)
  usage := float64(m.Alloc) / float64(m.Sys)
  return usage
}

//check_resources determines if system resource use for memory and CPU is below given thresholds
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
