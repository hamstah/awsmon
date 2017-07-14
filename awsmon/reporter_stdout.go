package main

import (
	"log"
)

type StdoutReporter struct{}

func NewStdoutReporter() (reporter StdoutReporter, err error) {
	return
}

func (reporter StdoutReporter) SendStat(stat Stat) (err error) {
	log.Printf("%+v", stat)
	return
}
