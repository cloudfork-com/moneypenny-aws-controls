package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Command struct {
	Name    string `json:"name"`    // plan,apply,report
	Profile string `json:"profile"` // optionally for one AWS profile
	Debug   bool   `json:"debug"`   // default is false
}

func HandleRequest(ctx context.Context, cmd *Command) (*string, error) {
	if cmd == nil {
		return nil, fmt.Errorf("received nil command")
	}
	message := fmt.Sprintf("Performing %s", cmd.Name)
	return &message, nil
}

func main() {
	lambda.Start(HandleRequest)
}
