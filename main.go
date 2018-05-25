package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/cirocosta/awsmon/lib"
)

// CliArguments groups all the arguments that are
// passed by the user to `awsmon`.
type CliArguments struct {
	Disk           []string      `arg:"separate,help:retrieve disk samples from disk locations" json:"disk"`
	Interval       time.Duration `arg:"help:interval between samples" json:"interval"`
	Load15M        bool          `arg:"--load-15m,help:retrieve load 15m avgs" json:"load-15m"`
	Load1M         bool          `arg:"--load-1m,help:retrieve load 1m avgs" json:"load-1m"`
	Load5M         bool          `arg:"--load-5m,help:retrieve load 5m avgs" json:"load-5m"`
	Memory         bool          `arg:"help:retrieve memory samples" json:"memory"`
	RelativizeLoad bool          `arg:"--relativize-load,help:makes loadavg relative to cpu count" json:"relativize-load"`

	Config string `arg:"help:path to awsmon configuration file" json:"-"`
	Debug  bool   `arg:"help:toggles debugging mode" json:"debug"`

	Aws                 bool   `arg:"help:whether or not to enable AWS support" json:"aws"`
	AwsAccessKey        string `arg:"--aws-access-key,help:aws access-key with cw putMetric caps" json:"aws-access-key"`
	AwsAggregatedOnly   bool   `arg:"--aws-aggregated-only,help:region for sending cloudwatch metrics to" json:"aws-aggregated-only"`
	AwsAutoScalingGroup string `arg:"--aws-asg,help:autoscaling group that the instance is in" json:"aws-autoscaling-group"`
	AwsInstanceId       string `arg:"--aws-instance-id,help:id of the instance (required if wanting AWS support)" json:"aws-instance-id"`
	AwsInstanceType     string `arg:"--aws-instance-type,help:type of the instance (required if wanting AWS support)" json:"aws-instance-type"`
	AwsNamespace        string `arg:"--aws-namespace,help:cloudwatch metric namespace" json:"aws-namespace"`
	AwsRegion           string `arg:"--aws-region,help:region for sending cloudwatch metrics to" json:"aws-region"`
	AwsSecretKey        string `arg:"--aws-secret-key,help:aws secret-key with cw putMetric caps" json:"aws-secret-key"`
}

var (
	reporter Reporter
	err      error
	disks    = []string{}
	args     = CliArguments{
		Aws:            false,
		AwsNamespace:   "System/Linux",
		Config:         "/etc/awsmon/config.json",
		Debug:          false,
		Disk:           []string{"/"},
		Interval:       30 * time.Second,
		Load1M:         true,
		Memory:         true,
		RelativizeLoad: true,
	}
)

// mustReadConfigFile reads the configuration file as specified
// via `args.Config`, loading the json configuration.
//
// In the case of errors, breaks the whole execution.
func mustReadConfigFile(args *CliArguments) {
	var (
		err    error
		logger = log.With().Str("config", args.Config).Logger()
	)

	_, err = os.Stat(args.Config)
	if os.IsNotExist(err) {
		return
	}

	fd, err := os.Open(args.Config)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("couldn't open configuration file")
		os.Exit(1)
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	err = dec.Decode(args)
	if err != nil {
		logger.Fatal().
			Err(err).
			Msg("couldn't decode json configuration file into args struct")
		os.Exit(1)
	}

	logger.Info().Msg("configuration loaded")
}

func findDisks() {
	for _, diskPath := range args.Disk {
		finfo, err := os.Stat(diskPath)
		if err == nil {
			if finfo.IsDir() {
				disks = append(disks, diskPath)
			}
		} else {
			log.Warn().
				Err(err).
				Str("path", diskPath).
				Msg("specified path is not a directory or could not be found")
		}
	}

	return
}

func collectAndSendDiskMetrics() {
	for _, diskPath := range disks {
		diskSample, err := TakeDiskSample(diskPath)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", diskPath).
				Msg("failed to take disk sample")
			continue
		}

		err = reporter.SendStat(NewDiskUtilizationStat(&diskSample))
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to send disk stat")
		}
	}

	return
}

func collectAndSendLoadMetrics() {
	loadSample, err := TakeLoadSample(args.RelativizeLoad)
	if err != nil {
		log.Error().
			Err(err).
			Bool("relativize-load", args.RelativizeLoad).
			Msg("failed to take load sample")
		return
	}

	// TODO send other loads too
	err = reporter.SendStat(NewLoadAvg1Stat(&loadSample))
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to report load stat")
	}

	return
}

func collectAndSendMemoryMetrics() {
	memorySample, err := TakeMemorySample()
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to take memory sample")
		return
	}

	err = reporter.SendStat(NewMemoryUtilizationStat(&memorySample))
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to report memory sample")
	}

	return
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
		ticker    = time.NewTicker(args.Interval)
		closeChan = make(chan os.Signal, 1)
	)

	defer ticker.Stop()
	signal.Notify(closeChan, syscall.SIGINT, syscall.SIGTERM)

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

	findDisks()

	go func() {
		for {
			select {
			case <-ticker.C:
				collectAndSendDiskMetrics()
				collectAndSendLoadMetrics()
				collectAndSendMemoryMetrics()
			}
		}
	}()

	log.Info().Msg("starting sampling")
	<-closeChan
	ticker.Stop()
}
