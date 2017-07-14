package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alexflint/go-arg"
)

type CliArguments struct {
	Interval  time.Duration `arg:"-i,help:interval between samples"`
	Memory    bool          `arg:"-m,help:retrieve memory samples"`
	Disk      []string      `arg:"-d,separate,help:retrieve disk samples from disk locations"`
	Namespace string        `arg:"-n,help:cloudwatch metric namespace"`
	Aws       bool          `arg:"-a,help:whether or not the instance is running in aws"`
}

var (
	args = CliArguments{
		Memory:    true,
		Disk:      []string{"/"},
		Interval:  30 * time.Second,
		Namespace: "System/Linux",
		Aws:       true,
	}
)

func main() {
	var reporter Reporter
	var err error

	arg.MustParse(&args)
	ticker := time.NewTicker(args.Interval)
	defer ticker.Stop()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)

	if args.Aws {
		reporter, err = NewReporter("cw")
	} else {
		reporter, err = NewReporter("stdout")
	}
	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				diskSample, err := TakeDiskSample("/")
				if err != nil {
					log.Println(err)
					continue
				}

				reporter.SendStat(NewDiskUtilizationStat(&diskSample))

				memorySample, err := TakeMemorySample()
				if err != nil {
					log.Println(err)
					continue
				}

				reporter.SendStat(NewMemoryUtilizationStat(&memorySample))
			}
		}
	}()

	log.Printf("starting sampling - %+v", args)
	<-closeChan
	ticker.Stop()
}
