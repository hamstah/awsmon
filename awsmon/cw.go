package main

import (
	_ "github.com/aws/aws-sdk-go/aws/session"
	_ "github.com/aws/aws-sdk-go/service/cloudwatch"
	_ "github.com/aws/aws-sdk-go/aws/ec2metadata"
)

//
// sess := session.Must(session.NewSession())
// svc := cloudwatch.New(sess)
