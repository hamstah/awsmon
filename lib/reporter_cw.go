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
}

// CloudWatchReporterConfig represents all the configuration
// needed for initializing the cloudwatch reporter.
// Note.: AutoScalingGroup is optional.
type CloudWatchReporterConfig struct {
	Debug bool

	Namespace    string
	InstanceId   string
	InstanceType string

	AutoScalingGroup string
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

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create AWS session.")
		return
	}

	var instanceTypeDimension = cloudwatch.Dimension{
		Name:  aws.String("InstanceType"),
		Value: aws.String(reporter.instanceType),
	}

	var instanceIdDimension = cloudwatch.Dimension{
		Name:  aws.String("InstanceId"),
		Value: aws.String(reporter.instanceId),
	}

	var instanceAsgDimension = cloudwatch.Dimension{
		Name:  aws.String("AutoScalingGroupName"),
		Value: aws.String(reporter.autoscalingGroup),
	}

	reporter.cw = cloudwatch.New(sess)
	reporter.dimensions = []*cloudwatch.Dimension{
		&instanceIdDimension,
		&instanceTypeDimension,
	}

	if reporter.autoscalingGroup != "" {
		reporter.dimensions = append(
			reporter.dimensions, &instanceAsgDimension)
	}

	log.Println("cw: reporter created")
	log.Printf("cw: instanceId=%s, instanceType=%s, asg=%s\n",
		reporter.instanceId, reporter.instanceType,
		reporter.autoscalingGroup)

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
