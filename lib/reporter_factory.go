package lib

import (
	"github.com/pkg/errors"
)

func NewReporter(reporterType string, cfg interface{}) (reporter Reporter, err error) {
	switch reporterType {
	case "cw":
		reporter, err = NewCloudWatchReporter(cfg.(CloudWatchReporterConfig))
	case "stdout":
		reporter, err = NewStdoutReporter()
	default:
		err = errors.Errorf("Unknown reporter type %s",
			reporterType)
		return
	}

	return
}
