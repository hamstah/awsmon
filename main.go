package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alexflint/go-arg"

	. "github.com/cirocosta/awsmon/lib"
)

// CliArguments groups all the arguments that are
// passed by the user to `awsmon`.
type CliArguments struct {
	Interval  time.Duration `arg:"-i,help:interval between samples"`
	Memory    bool          `arg:"-m,help:retrieve memory samples"`
	Disk      []string      `arg:"-d,separate,help:retrieve disk samples from disk locations"`
	Namespace string        `arg:"-n,help:cloudwatch metric namespace"`

	Debug bool `arg:"env,help:toggles debugging mode"`

	Aws           bool   `arg:"-a,help:whether or not to enable AWS support"`
	AwsAsg        string `arg:"help:autoscaling group that the instance is in"`
	AwsInstanceId string `arg:"help:id of the instance (required if wanting AWS support)"`
	AwsRegion     string `arg:"help:region of the instance (required if wanting AWS support)"`
}

var (
	args = CliArguments{
		Memory:    true,
		Disk:      []string{"/"},
		Interval:  30 * time.Second,
		Namespace: "System/Linux",
		Aws:       true,
		Debug:     false,
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
		reporter, err = NewReporter("cw", CloudWatchReporterConfig{
			Debug:     args.Debug,
			Namespace: args.Namespace,
		})
	} else {
		reporter, err = NewReporter("stdout", struct{}{})
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
				} else {
					reporter.SendStat(NewDiskUtilizationStat(&diskSample))
				}

				memorySample, err := TakeMemorySample()
				if err != nil {
					log.Println(err)
					continue
				} else {
					reporter.SendStat(NewMemoryUtilizationStat(&memorySample))
				}
			}
		}
	}()

	log.Printf("starting sampling - %+v", args)
	<-closeChan
	ticker.Stop()
}
