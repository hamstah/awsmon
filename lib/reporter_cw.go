package lib

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
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
	region           string
}

type CloudWatchReporterConfig struct {
	Debug     bool
	Namespace string
}

var (
	awsConfig = &aws.Config{}
)

func NewCloudWatchReporter(cfg CloudWatchReporterConfig) (reporter CloudWatchReporter, err error) {
	if cfg.Namespace == "" {
		err = errors.Errorf("A namespace must be provided")
		return
	}

	if cfg.Debug {
		awsConfig.LogLevel =
			aws.LogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create AWS session.")
		return
	}

	err = reporter.fetchInstanceMetadata(sess)
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't fetch instance metadata")
		return
	}

	var instanceNameDimension = cloudwatch.Dimension{
		Name:  aws.String("InstanceName"),
		Value: aws.String(reporter.instanceId),
	}

	var instanceAsgDimension = cloudwatch.Dimension{
		Name:  aws.String("AutoScalingGroup"),
		Value: aws.String(reporter.autoscalingGroup),
	}

	reporter.namespace = cfg.Namespace
	reporter.cw = cloudwatch.New(sess)
	reporter.dimensions = []*cloudwatch.Dimension{
		&instanceNameDimension, &instanceAsgDimension,
	}

	log.Println("cw: reporter created")
	log.Printf("cw: instanceId=%s, region=%s, asg=%s\n",
		reporter.instanceId, reporter.region,
		reporter.autoscalingGroup)

	return
}

func (reporter *CloudWatchReporter) fetchInstanceMetadata(sess *session.Session) (err error) {
	meta := ec2metadata.New(sess)
	asg := autoscaling.New(sess)

	doc, err := meta.GetInstanceIdentityDocument()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't retrieve instance metadata")
		return
	}

	reporter.instanceId = doc.InstanceID
	reporter.region = doc.Region

	resp, err := asg.DescribeAutoScalingInstances(&autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{&doc.InstanceID},
		MaxRecords:  aws.Int64(1),
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't retrieve ASG for instance %s",
			reporter.instanceId)
		return
	}

	if len(resp.AutoScalingInstances) != 0 {
		reporter.autoscalingGroup = *resp.
			AutoScalingInstances[0].AutoScalingGroupName
	} else {
		reporter.autoscalingGroup = "none"
	}

	return
}

func (reporter CloudWatchReporter) SendStat(stat Stat) (err error) {
	log.Printf("cw: sending stat %+v\n", stat)

	var datum = cloudwatch.MetricDatum{
		MetricName: aws.String(stat.Name),
		Timestamp:  aws.Time(stat.When),
		Unit:       aws.String(stat.Unit),
		Dimensions: reporter.dimensions,
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