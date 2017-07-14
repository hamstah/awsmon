package main

import (
	"time"

	"github.com/guillermo/go.procmeminfo"
	"github.com/pkg/errors"
)

type MemorySample struct {
	MemoryUtilization float64
	SwapUtilization   float64
	When              time.Time
}

var (
	memInfo = &procmeminfo.MemInfo{}
)

// TakeMemorySample updates the /proc/meminfo sampler and returns
// a struct with the desired metrics to be consumed.
func TakeMemorySample() (sample MemorySample, err error) {
	err = memInfo.Update()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't fetch memory sample from /proc/memInfo")
		return
	}

	used := float64(memInfo.Used())
	total := float64(memInfo.Total())
	swapUsed := float64(memInfo.Swap())
	swapTotal := float64((*memInfo)["SwapTotal"])

	if swapTotal == 0 {
		sample.SwapUtilization = 0
	} else {
		sample.SwapUtilization = Round(swapUsed / swapTotal * 100)
	}

	sample.MemoryUtilization = Round(used / total * 100)
	sample.When = time.Now()
	return
}
