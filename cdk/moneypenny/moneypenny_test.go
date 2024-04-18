package main

import (
	"testing"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
)

func TestMoneypennyStack(t *testing.T) {
	// GIVEN
	app := awscdk.NewApp(nil)

	// WHEN
	stack := NewMoneypennyStack(app, "MyStack", nil)

	// THEN
	template := assertions.Template_FromStack(stack, nil)

	// template.HasResourceProperties(jsii.String("AWS::LAMBDA::Function"), map[string]interface{}{
	// 	"VisibilityTimeout": 300,
	// })
	t.Log(template.ToJSON())
}
