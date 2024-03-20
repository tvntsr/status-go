// +build go1.5,!js

package metrics

import "runtime"

func gcCPUFraction(memStats *runtime.MemStats) float64 {
	return memStats.GCCPUFraction
}
