package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"
)

type CloudWatchReporter struct {
	cw *cloudwatch.CloudWatch
	// metadata
}

func NewCloudWatchReporter() (reporter CloudWatchReporter, err error) {
	sess, err := session.NewSession()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create AWS session.")
		return
	}

	reporter.cw = cloudwatch.New(sess)
	return
}

func (reporter *CloudWatchReporter) SendSamples(ms MemorySample, ds DiskSample) (err error) {
	var instanceNameDimension = cloudwatch.Dimension{
		Name:  aws.String("InstanceName"),
		Value: aws.String("blabla"),
	}

	var datum = cloudwatch.MetricDatum{
		MetricName: aws.String("MemoryUtilization"),
		Timestamp:  aws.Time(time.Now()),
		Unit:       aws.String("Percent"),
		Dimensions: []*cloudwatch.Dimension{
			&instanceNameDimension,
		},
		Value: aws.Float64(ms.MemoryUtilization),
	}

	var input = cloudwatch.PutMetricDataInput{
		Namespace:  "System/Linux",
		MetricData: &datum,
	}

	return
}
