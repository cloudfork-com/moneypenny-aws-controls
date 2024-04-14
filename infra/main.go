package main

import awscdk "github.com/aws/aws-cdk-go/awscdk/v2"

// https://pkg.go.dev/github.com/aws/aws-cdk-go/awscdk/v2
func main() {
	app := awscdk.NewApp()
	stack := awscdk.NewStack(app, jsii.String("TestStack"))
}
