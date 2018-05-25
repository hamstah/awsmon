package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/cirocosta/awsmon/lib"
)

// CliArguments groups all the arguments that are
// passed by the user to `awsmon`.
type CliArguments struct {
	Interval       time.Duration `arg:"help:interval between samples" json:"interval"`
	Memory         bool          `arg:"help:retrieve memory samples" json:"memory"`
	Load1M         bool          `arg:"--load-1m,help:retrieve load 1m avgs" json:"load-1m"`
	Load5M         bool          `arg:"--load-5m,help:retrieve load 5m avgs" json:"load-5m"`
	Load15M        bool          `arg:"--load-15m,help:retrieve load 15m avgs" json:"load-15m"`
	RelativizeLoad bool          `arg:"--relativize-load,help:makes loadavg relative to cpu count" json:"relativize-load"`
	Disk           []string      `arg:"separate,help:retrieve disk samples from disk locations" json:"disk"`

	Config string `arg:"help:path to awsmon configuration file" json:"-"`
	Debug  bool   `arg:"help:toggles debugging mode" json:"debug"`

	Aws                 bool   `arg:"help:whether or not to enable AWS support" json:"aws"`
	AwsAccessKey        string `arg:"--aws-access-key,help:aws access-key with cw putMetric caps" json:"aws-access-key"`
	AwsSecretKey        string `arg:"--aws-secret-key,help:aws secret-key with cw putMetric caps" json:"aws-secret-key"`
	AwsAutoScalingGroup string `arg:"--aws-asg,help:autoscaling group that the instance is in" json:"aws-autoscaling-group"`
	AwsInstanceId       string `arg:"--aws-instance-id,help:id of the instance (required if wanting AWS support)" json:"aws-instance-id"`
	AwsInstanceType     string `arg:"--aws-instance-type,help:type of the instance (required if wanting AWS support)" json:"aws-instance-type"`
	AwsNamespace        string `arg:"--aws-namespace,help:cloudwatch metric namespace" json:"aws-namespace"`
	AwsRegion           string `arg:"--aws-region,help:region for sending cloudwatch metrics to" json:"aws-region"`
	AwsAggregatedOnly   bool   `arg:"--aws-aggregated-only,help:region for sending cloudwatch metrics to" json:"aws-aggregated-only"`
}

var (
	args = CliArguments{
		Config:         "/etc/awsmon/config.json",
		Memory:         true,
		Load1M:         true,
		Disk:           []string{"/"},
		Interval:       30 * time.Second,
		AwsNamespace:   "System/Linux",
		Aws:            false,
		Debug:          false,
		RelativizeLoad: true,
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
	arg.MustParse(&args)
	mustReadConfigFile(&args)

	if args.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().
		Interface("configuration", args).
		Msg("initializing")

	var (
		reporter  Reporter
		err       error
		ticker    = time.NewTicker(args.Interval)
		closeChan = make(chan os.Signal, 1)
	)

	defer ticker.Stop()
	signal.Notify(closeChan, os.Interrupt)

	if args.Aws {
		reporter, err = NewReporter("cw", CloudWatchReporterConfig{
			AccessKey:        args.AwsAccessKey,
			SecretKey:        args.AwsSecretKey,
			Debug:            args.Debug,
			Namespace:        args.AwsNamespace,
			InstanceId:       args.AwsInstanceId,
			InstanceType:     args.AwsInstanceType,
			AutoScalingGroup: args.AwsAutoScalingGroup,
			Region:           args.AwsRegion,
			AggregatedOnly:   args.AwsAggregatedOnly,
		})
	} else {
		reporter, err = NewReporter("stdout", struct{}{})
	}
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("failed to instantiate reporter")
		os.Exit(1)
	}

	var disksToLookupFor = make([]string, 0)
	for _, diskPath := range args.Disk {
		finfo, err := os.Stat(diskPath)
		if err == nil {
			if finfo.IsDir() {
				disksToLookupFor = append(disksToLookupFor, diskPath)
			}
		} else {
			log.Warn().
				Err(err).
				Str("path", diskPath).
				Msg("specified path is not a directory or could not be found")
		}
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				for _, diskPath := range disksToLookupFor {
					diskSample, err := TakeDiskSample(diskPath)
					if err != nil {
						log.Error().
							Err(err).
							Str("path", diskPath).
							Msg("failed to take disk sample")
						continue
					} else {
						err = reporter.SendStat(NewDiskUtilizationStat(&diskSample))
						if err != nil {
							log.Error().
								Err(err).
								Msg("failed to send disk stat")
						}
					}
				}

				loadSample, err := TakeLoadSample(args.RelativizeLoad)
				if err != nil {
					log.Error().
						Err(err).
						Bool("relativize-load", args.RelativizeLoad).
						Msg("failed to take load sample")
					continue
				} else {
					err = reporter.SendStat(NewLoadAvg1Stat(&loadSample))
					if err != nil {
						log.Error().
							Err(err).
							Msg("failed to report load stat")
					}
				}

				memorySample, err := TakeMemorySample()
				if err != nil {
					log.Error().
						Err(err).
						Msg("failed to take memory sample")
					continue
				} else {
					err = reporter.SendStat(NewMemoryUtilizationStat(&memorySample))
					if err != nil {
						log.Error().
							Err(err).
							Msg("failed to report memory sample")
					}
				}
			}
		}
	}()

	log.Info().Msg("starting sampling")
	<-closeChan
	ticker.Stop()
}
