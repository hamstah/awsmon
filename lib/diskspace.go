package lib

import (
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// DiskSample represents the sample of
type DiskSample struct {
	DiskUtilization   float64
	InodesUtilization float64
	When              time.Time
}

// TakeDiskSample retrieves a sample about disk utilization
// of a mounted filesystem denoted by the first parameter
// by looking at the results from the `statfs` syscall.
func TakeDiskSample(path string) (sample DiskSample, err error) {
	var fsInfo = syscall.Statfs_t{}

	err = syscall.Statfs(path, &fsInfo)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't retrieve FS info for path %s", path)
		return
	}

	diskTotal := int(fsInfo.Bsize) * int(fsInfo.Blocks)
	diskAvail := int(fsInfo.Bsize) * int(fsInfo.Bavail)
	diskUsed := diskTotal - diskAvail

	inodesTotal := int(fsInfo.Files)
	inodesFree := int(fsInfo.Ffree)

	sample.InodesUtilization = Round(100 * (1 - float64(inodesFree)/float64(inodesTotal)))
	sample.DiskUtilization = Round((float64(diskUsed) / float64(diskTotal)) * 100)
	sample.When = time.Now()
	return
}
