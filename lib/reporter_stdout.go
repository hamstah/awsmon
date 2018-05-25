package lib

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StdoutReporter struct {
	logger zerolog.Logger
}

func NewStdoutReporter() (reporter *StdoutReporter, err error) {
	reporter = &StdoutReporter{
		logger: log.With().Str("from", "reporter_stdout").Logger(),
	}
	return
}

func (r *StdoutReporter) SendStat(stat Stat) (err error) {
	r.logger.Info().
		Interface("stat", stat).
		Msg("sending stat")
	return
}
