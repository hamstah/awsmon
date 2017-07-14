package main

import (
	"github.com/pkg/errors"
)

func NewReporter(reporterType string) (reporter Reporter, err error) {
	switch reporterType {
	case "cw":
		reporter, err = NewCloudWatchReporter()
	case "stdout":
		reporter, err = NewStdoutReporter()
	default:
		err = errors.Errorf("Unknown reporter type %s",
			reporterType)
		return
	}

	return
}
