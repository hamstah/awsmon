package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"
)

type CloudWatchReporter struct {
	cw *cloudwatch.CloudWatch

	autoscalingGroup string
	instanceId       string
	region           string
}

func NewCloudWatchReporter() (reporter CloudWatchReporter, err error) {
	sess, err := session.NewSession()
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't create AWS session.")
		return
	}

	err = reporter.fetchInstanceMetadata(sess)
	if err != nil {
		return
	}

	reporter.cw = cloudwatch.New(sess)
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

	if len(resp.AutoScalingInstances) == 0 {
		err = errors.New(
			"No autoscaling group found")
		return
	}

	reporter.autoscalingGroup = *resp.AutoScalingInstances[0].AutoScalingGroupName
	return
}

func (reporter *CloudWatchReporter) SendStat(stat Stat) (err error) {
	var instanceNameDimension = cloudwatch.Dimension{
		Name:  aws.String("InstanceName"),
		Value: aws.String("blabla"),
	}

	var datum = cloudwatch.MetricDatum{
		MetricName: aws.String(stat.Name),
		Timestamp:  aws.Time(stat.When),
		Unit:       aws.String(stat.Unit),
		Dimensions: []*cloudwatch.Dimension{
			&instanceNameDimension,
		},
		Value: aws.Float64(stat.Value),
	}

	var input = cloudwatch.PutMetricDataInput{
		Namespace: aws.String("System/Linux"),
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
