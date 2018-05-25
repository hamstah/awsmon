package lib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// CloudWatchReporter implements the Reporter interface
// to provide the connection between samples generated
// by the machine and CloudWatch.
type CloudWatchReporter struct {
	logger     zerolog.Logger
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

	AccessKey        string
	SecretKey        string
	AutoScalingGroup string
	InstanceId       string
	InstanceType     string
	Namespace        string
	Region           string
	AggregatedOnly   bool
}

func NewCloudWatchReporter(cfg CloudWatchReporterConfig) (reporter *CloudWatchReporter, err error) {
	var awsConfig = &aws.Config{}

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

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			cfg.AccessKey, cfg.SecretKey, "")
	}

	if cfg.Region != "" {
		awsConfig.Region = aws.String(cfg.Region)
	}

	if cfg.Debug {
		awsConfig.LogLevel =
			aws.LogLevel(aws.LogDebug | aws.LogDebugWithRequestErrors)
	}

	reporter = &CloudWatchReporter{
		instanceId:       cfg.InstanceId,
		instanceType:     cfg.InstanceType,
		autoscalingGroup: cfg.AutoScalingGroup,
		namespace:        cfg.Namespace,
		aggregatedOnly:   cfg.AggregatedOnly,
		logger:           log.With().Str("from", "reporter_cw").Logger(),
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

	reporter.logger.Debug().
		Interface("reporter", reporter).
		Msg("reporter created")

	return
}

func (reporter *CloudWatchReporter) SendStat(stat Stat) (err error) {
	reporter.logger.Debug().
		Interface("stat", stat).
		Msg("sending stat")

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
