package main

// Reporter is responsible for reporting
// stats to somewhere.
type Reporter interface {

	// SendStat sends the stat to a metrics collector.
	SendStat(stat Stat) (err error)
}
