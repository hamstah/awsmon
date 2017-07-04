type AWSManager interface {
	GetAutoscalingGroup()
	GetDimensions()
	AddMetric()
	PutMetric()
	GetInstanceMetadata()
}
