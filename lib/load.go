package lib

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type LoadSample struct {
	One      float64
	Five     float64
	Fithteen float64
	When     time.Time
}

var (
	loadavgFileName string  = "/proc/loadavg"
	cpuCount        float64 = float64(runtime.NumCPU())
)

// getLoad retrieves the a slice of 'float64' values from
// '/proc/loadavg' file.
func getLoad() (loads []float64, err error) {
	data, err := ioutil.ReadFile(loadavgFileName)
	if err != nil {
		err = errors.Wrapf(err, "couldn't read loadavg file")
		return
	}

	loads, err = parseLoad(string(data))
	if err != nil {
		err = errors.Wrapf(err, "couldn't parse loadavg retrieved")
		return
	}

	return
}

// parseLoad takes the slice of float64 values and returns 1m, 5m and 15m.
func parseLoad(data string) (loads []float64, err error) {
	loads = make([]float64, 3)

	parts := strings.Fields(data)
	if len(parts) < 3 {
		err = errors.Errorf("unexpected content in loadavg file")
		return
	}

	for i, load := range parts[0:3] {
		loads[i], err = strconv.ParseFloat(load, 64)
		if err != nil {
			err = errors.Errorf("could not parse load '%s': %s", load, err)
			return
		}
	}

	return
}

// TakeLoadSample updates the /proc/loadavg and returns
// a struct with the desired metrics to be consumed.
func TakeLoadSample(relativize bool) (sample LoadSample, err error) {
	sample.When = time.Now()
	loads, err := getLoad()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't extract loads from load %s", loadavgFileName)
		return
	}

	if relativize {
		sample.One = loads[0] / cpuCount
		sample.Five = loads[1] / cpuCount
		sample.Fithteen = loads[2] / cpuCount
	} else {
		sample.One = loads[0]
		sample.Five = loads[1]
		sample.Fithteen = loads[2]
	}

	return
}
