package lib

type AWSManagerMock struct { }

func (aws AWSManagerMock) getInstanceMetadata() (metadata map[string]string, err error) {
	metadata = map[string]string{
		"devpayProductCodes": null,
		"privateIp": "10.0.5.89",
		"availabilityZone": "us-west-1a",
		"version" : "2010-08-31",
		"region" : "us-west-1",
		"instanceId" : "i-e0iag2b",
		"billingProducts" : null,
		"accountId" : "208372078340",
		"instanceType" : "m3.xlarge",
		"imageId" : "ami-43f91b07",
		"kernelId" : null,
		"ramdiskId" : null,
		"pendingTime" : "2015-06-30T08:28:48Z",
		"architecture" : "x86_64"
	}

	return
}
