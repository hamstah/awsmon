package main

import (
	"github.com/guillermo/go.procmeminfo"
	"github.com/pkg/errors"
)

type MemorySample struct {
	MemoryUtilization float64
	SwapUtilization   float64
}

var (
	memInfo = &procmeminfo.MemInfo{}
)

// GetMemorySample updates the /proc/meminfo sampler and returns
// a struct with the desired metrics to be consumed.
func GetMemorySample() (sample MemorySample, err error) {
	err = memInfo.Update()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't fetch memory sample from /proc/memInfo")
		return
	}

	used := float64(memInfo.Used())
	total := float64(memInfo.Available())
	swapUsed := float64(memInfo.Available())
	swapTotal := float64(memInfo.Available())

	sample.MemoryUtilization = Round(used / total * 100)
	sample.SwapUtilization = Round(swapUsed / swapTotal * 100)
	return
}
