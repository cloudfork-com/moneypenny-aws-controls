package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewMoneypennyStack(app, "MoneypennyStack", &MoneypennyStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

type MoneypennyStackProps struct {
	awscdk.StackProps
}

func NewMoneypennyStack(scope constructs.Construct, id string, props *MoneypennyStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	role := awsiam.NewRole(stack, jsii.String("moneypenny-aws-controls-role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
	})
	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: jsii.Strings(
			"ecs:ListServices",
			"ecs:UpdateService",
			"ecs:ListTagsForResource",
			"ecs:ListTasks",
			"ecs:StopTask",
			"ecs:DescribeServices",
			"ecs:DescribeTaskSets",
			"ecs:DescribeTasks",
			"ecs:ListTaskDefinitions",
			"ecs:ListClusters",
		),
		Resources: jsii.Strings("*"),
	}))

	// Add a managed policy to a role you can use
	// role.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AmazonECS_FullAccess")))
	// role.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")))

	// https://aws.amazon.com/blogs/compute/migrating-aws-controlsLambda-functions-from-the-go1-x-runtime-to-the-custom-runtime-on-amazon-linux-2/
	controlsLambda := awslambda.NewFunction(stack, jsii.String("moneypenny-aws-controls"), &awslambda.FunctionProps{
		Code:         awslambda.Code_FromAsset(jsii.String("../../controls-lambda"), &awss3assets.AssetOptions{}), // folder where bootstrap executable is located
		Runtime:      awslambda.Runtime_PROVIDED_AL2023(),
		Handler:      jsii.String("bootstrap"),
		Architecture: awslambda.Architecture_ARM_64(),
		Role:         role,
		Description:  jsii.String("Moneypenny AWS Controls - Lambda function to control the desired count of ECS services"),
		Environment: &map[string]*string{
			"Variables": jsii.String("{TIME_ZONE=Europe/Amsterdam,BASIC_USER=nil,BASIC_PASSWORD=zork}"),
		},
		MemorySize: jsii.Number(128),
		Timeout:    awscdk.Duration_Seconds(jsii.Number(10)),
		CurrentVersionOptions: &awslambda.VersionOptions{
			RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
			RetryAttempts: jsii.Number(1),
		},
		LogRetention: awslogs.RetentionDays_FIVE_DAYS,
	})

	// Add protected URL
	// TODO needed when using API Gateway ???
	controlsLambda.AddFunctionUrl(&awslambda.FunctionUrlOptions{
		AuthType: awslambda.FunctionUrlAuthType_AWS_IAM,
	})

	// Create an API Gateway
	httpApi := awsapigatewayv2.NewHttpApi(stack, jsii.String("moneypenny-aws-controls-api"), &awsapigatewayv2.HttpApiProps{
		ApiName: jsii.String("Moneypenny AWS Controls API"),
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowHeaders: jsii.Strings("*"),
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{"GET", "POST"},
			AllowOrigins: jsii.Strings("*"),
		},
		CreateDefaultStage: jsii.Bool(true),
	})

	// Create a Lambda integration
	lambdaIntegration := awsapigatewayv2integrations.NewHttpLambdaIntegration(
		jsii.String("moneypenny-aws-controls-integration"),
		controlsLambda,
		&awsapigatewayv2integrations.HttpLambdaIntegrationProps{})

	awsapigatewayv2.NewHttpRoute(stack, jsii.String("moneypenny-aws-controls-get"), &awsapigatewayv2.HttpRouteProps{
		HttpApi:     httpApi,
		Integration: lambdaIntegration,
		RouteKey:    awsapigatewayv2.HttpRouteKey_With(jsii.String("/"), awsapigatewayv2.HttpMethod_GET),
	})
	awsapigatewayv2.NewHttpRoute(stack, jsii.String("moneypenny-aws-controls-post"), &awsapigatewayv2.HttpRouteProps{
		HttpApi:     httpApi,
		Integration: lambdaIntegration,
		RouteKey:    awsapigatewayv2.HttpRouteKey_With(jsii.String("/"), awsapigatewayv2.HttpMethod_POST),
	})
	return stack
}
