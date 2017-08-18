package lib

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"
)

// CloudWatchReporter implements the Reporter interface
// to provide the connection between samples generated
// by the machine and CloudWatch.
type CloudWatchReporter struct {
	cw         *cloudwatch.CloudWatch
	dimensions []*cloudwatch.Dimension

	namespace        string
	autoscalingGroup string
	instanceId       string
	instanceType     string
	aggregatedOnly   bool
}

// CloudWatchReporterConfig represents all the configuration
// needed for initializing the cloudwatch reporter.
// Note.: AutoScalingGroup is optional.
type CloudWatchReporterConfig struct {
	Debug bool

	AutoScalingGroup string
	InstanceId       string
	InstanceType     string
	Namespace        string
	Region           string
	AggregatedOnly   bool
}

var (
	awsConfig = &aws.Config{}
)

func NewCloudWatchReporter(cfg CloudWatchReporterConfig) (reporter CloudWatchReporter, err error) {
	if cfg.Namespace == "" {
		err = errors.Errorf("A namespace must be provided")
		return
	}

	if cfg.InstanceId == "" {
		err = errors.Errorf("An instanceId must be provided")
		return
	}

	if cfg.InstanceType == "" {
		err = errors.Errorf("An instanceType must be provided")
		return
	}

	if cfg.Debug {
		awsConfig.LogLevel =
			aws.LogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors)
	}

	reporter.instanceId = cfg.InstanceId
	reporter.instanceType = cfg.InstanceType
	reporter.autoscalingGroup = cfg.AutoScalingGroup
	reporter.namespace = cfg.Namespace
	reporter.aggregatedOnly = cfg.AggregatedOnly

	if cfg.Region != "" {
		awsConfig.Region = aws.String(cfg.Region)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create AWS session.")
		return
	}

	if cfg.AggregatedOnly {
		if cfg.AutoScalingGroup == "" {
			err = errors.Errorf("aggregatedOnly mode requires autoscaling group.")
			return
		}
	}

	reporter.cw = cloudwatch.New(sess)
	reporter.dimensions = make([]*cloudwatch.Dimension, 0)
	if !cfg.AggregatedOnly {
		reporter.dimensions = append(
			reporter.dimensions, &cloudwatch.Dimension{
				Name:  aws.String("InstanceType"),
				Value: aws.String(reporter.instanceType),
			})

		reporter.dimensions = append(
			reporter.dimensions, &cloudwatch.Dimension{
				Name:  aws.String("InstanceId"),
				Value: aws.String(reporter.instanceId),
			})

		if reporter.autoscalingGroup != "" {
			reporter.dimensions = append(
				reporter.dimensions, &cloudwatch.Dimension{
					Name:  aws.String("AutoScalingGroupName"),
					Value: aws.String(reporter.autoscalingGroup),
				})
		}
	} else {
		reporter.dimensions = append(
			reporter.dimensions, &cloudwatch.Dimension{
				Name:  aws.String("AutoScalingGroupName"),
				Value: aws.String(reporter.autoscalingGroup),
			})
	}

	log.Println("cw: reporter created")
	log.Printf("cw: instanceId=%s, instanceType=%s, asg=%s aggr-only=%v\n",
		reporter.instanceId, reporter.instanceType,
		reporter.autoscalingGroup, reporter.aggregatedOnly)

	return
}

func (reporter CloudWatchReporter) SendStat(stat Stat) (err error) {
	log.Printf("cw: sending stat %+v\n", stat)

	var extraDimensions = make([]*cloudwatch.Dimension, 0)
	for k, v := range stat.ExtraDimensions {
		extraDimensions = append(extraDimensions, &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	var datum = cloudwatch.MetricDatum{
		MetricName: aws.String(stat.Name),
		Timestamp:  aws.Time(stat.When),
		Unit:       aws.String(stat.Unit),
		Dimensions: append(reporter.dimensions, extraDimensions...),
		Value:      aws.Float64(stat.Value),
	}

	var input = cloudwatch.PutMetricDataInput{
		Namespace: aws.String(reporter.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			&datum,
		},
	}

	_, err = reporter.cw.PutMetricData(&input)
	if err != nil {
		err = errors.Wrapf(err,
			"Errored sending metric to cloudwatch.")
		return
	}

	return
}
