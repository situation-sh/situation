// Package perf manages the agent performance. It
// is mainly a monitoring tool to check memory and
// cpu usage
package perf

import (
	"runtime"

	"github.com/situation-sh/situation/models"
)

var memStats runtime.MemStats

func Collect() models.Performance {
	runtime.GC()
	runtime.ReadMemStats(&memStats)

	return models.Performance{
		HeapAlloc: memStats.HeapAlloc,
		HeapSys:   memStats.HeapSys,
	}
}
