package main

import (
	"log"
	"time"

	"github.com/alexflint/go-arg"
)

var args struct {
	Interval  time.Duration `arg:"-i,help:interval between samples"`
	Memory    bool          `arg:"-m,help:retrieve memory samples"`
	Disk      []string      `arg:"-d,separate,help:retrieve disk samples from disk locations"`
	Namespace string        `arg:"-n,help:cloudwatch metric namespace"`
	Aws       bool          `arg:"-a,help:whether or not the instance is running in aws"`
}

func main() {
	args.Memory = true
	args.Disk = []string{"/"}
	args.Interval = 1 * time.Minute
	args.Namespace = "System/Linux"
	args.Aws = true

	arg.MustParse(&args)

	log.Printf("starting sampling - %+v", args)
}
