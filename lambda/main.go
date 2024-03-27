package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
)

type Command struct {
	Name string `json:"command"` // plan,apply,report
}

func HandleRequest(ctx context.Context, cmd *Command) (*string, error) {
	if cmd == nil {
		return nil, fmt.Errorf("received nil command")
	}
	pe, err := mac.NewPlanExecutor([]*mac.ServicePlan{}, "") // default
	if err != nil {
		return nil, fmt.Errorf("failed to create executor:%w", err)
	}
	switch cmd.Name {
	case "apply":
		pe.Apply()
	case "report":
		pe.Report()
	default:
		pe.Plan()
	}
	return nil, nil
}

func main() {
	lambda.Start(HandleRequest)
}
