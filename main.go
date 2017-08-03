package main

import (
	"encoding/json"
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
	Interval time.Duration `arg:"-i,help:interval between samples" json:"interval"`
	Memory   bool          `arg:"-m,help:retrieve memory samples" json:"memory"`
	Disk     []string      `arg:"-d,separate,help:retrieve disk samples from disk locations" json:"disk"`

	Config string `arg:env,help:path to awsmon configuration file json:"-"`
	Debug  bool   `arg:"env,help:toggles debugging mode" json:"debug"`

	Aws                 bool   `arg:"-a,help:whether or not to enable AWS support" json:"aws"`
	AwsAutoScalingGroup string `arg:"help:autoscaling group that the instance is in" json:"aws-autoscaling-group"`
	AwsInstanceId       string `arg:"help:id of the instance (required if wanting AWS support)" json:"aws-instance-id"`
	AwsInstanceType     string `arg:"help:type of the instance (required if wanting AWS support)" json:"aws-instance-type"`
	AwsNamespace        string `arg:"-n,help:cloudwatch metric namespace" json:"aws-namespace"`
}

var (
	args = CliArguments{
		Config:       "/etc/awsmon/config.json",
		Memory:       true,
		Disk:         []string{"/"},
		Interval:     30 * time.Second,
		AwsNamespace: "System/Linux",
		Aws:          true,
		Debug:        false,
	}
)

func mustReadConfigFile(args *CliArguments) {
	if _, err := os.Stat(args.Config); os.IsNotExist(err) {
		return
	}

	fd, err := os.Open(args.Config)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	if err = dec.Decode(args); err != nil {
		panic(err)
	}
}

func main() {
	var reporter Reporter
	var err error

	arg.MustParse(&args)
	mustReadConfigFile(&args)
	log.Println("configuration parsed %+v", args)

	ticker := time.NewTicker(args.Interval)
	defer ticker.Stop()

	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)

	if args.Aws {
		reporter, err = NewReporter("cw", CloudWatchReporterConfig{
			Debug:            args.Debug,
			Namespace:        args.AwsNamespace,
			InstanceId:       args.AwsInstanceId,
			InstanceType:     args.AwsInstanceType,
			AutoScalingGroup: args.AwsAutoScalingGroup,
		})
	} else {
		reporter, err = NewReporter("stdout", struct{}{})
	}
	if err != nil {
		log.Println(err)
		return
	}

	var disksToLookupFor = make([]string, 0)
	for _, diskPath := range args.Disk {
		finfo, err := os.Stat(diskPath)
		if err == nil {
			if finfo.IsDir() {
				disksToLookupFor = append(disksToLookupFor, diskPath)
			}
		} else {
			log.Printf(
				"WARNING: specified disk path %s "+
					"is not a directory or couldn't be found - %+v\n",
				diskPath, err)
		}
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, diskPath := range disksToLookupFor {
					diskSample, err := TakeDiskSample(diskPath)
					if err != nil {
						log.Printf("ERROR: errored taking disk sample - %+v\n", err)
						continue
					} else {
						reporter.SendStat(NewDiskUtilizationStat(&diskSample))
					}
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
