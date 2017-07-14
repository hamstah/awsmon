package main

import (
	"time"
)

type Stat struct {
	Name  string
	Unit  string
	Value float64
	When  time.Time
}

// NewMemoryUtilizationStat generates a generic Stat
// structure prefilled with information about memory
// utilization
func NewMemoryUtilizationStat(sample *MemorySample) Stat {
	return Stat{
		Name:  "MemoryUtilization",
		Unit:  "Percent",
		When:  sample.When,
		Value: sample.MemoryUtilization,
	}
}

// NewSwapUtilizationStat generates a generic Stat
// structure prefilled with information about memory
// utilization
func NewSwapUtilizationStat(sample *MemorySample) Stat {
	return Stat{
		Name:  "SwapUtilization",
		Unit:  "Percent",
		When:  sample.When,
		Value: sample.SwapUtilization,
	}
}

// NewDiskUtilizationStat generates a generic Stat
// structure prefilled with information about disk
// utilization.
func NewDiskUtilizationStat(sample *DiskSample) Stat {
	return Stat{
		Name:  "DiskUtilization",
		Unit:  "Percent",
		When:  sample.When,
		Value: sample.DiskUtilization,
	}
}

// NewInodesUtilizationStat generates a generic Stat
// structure prefilled with information about inodes
// utilization.
func NewInodesUtilizationStat(sample *DiskSample) Stat {
	return Stat{
		Name:  "InodesUtilization",
		Unit:  "Percent",
		When:  sample.When,
		Value: sample.DiskUtilization,
	}
}
