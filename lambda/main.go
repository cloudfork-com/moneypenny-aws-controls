package main

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/emicklei/htmlslog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
)

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	buf := new(bytes.Buffer)
	resp := events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "text/html; charset=UTF-8"},
		StatusCode: 200}

	logHandler := htmlslog.New(htmlslog.Options{Level: slog.LevelDebug, Title: "moneypenny-aws-controls"})
	slog.SetDefault(slog.New(logHandler))

	pe, err := mac.NewPlanExecutor([]*mac.ServicePlan{}, "") // default profile
	if err != nil {
		resp.StatusCode = 500
		resp.Body = logHandler.Close()
		return resp, err
	}
	action := req.QueryStringParameters["do"]
	switch action {
	case "apply":
		pe.Apply()
	case "report":
		if err := pe.ReportHTMLOn(buf); err != nil {
			resp.StatusCode = 500
			resp.Body = logHandler.Close()
			return resp, err
		}
		resp.Body = buf.String()
		return resp, nil
	default:
		pe.Plan()
	}
	resp.Body = logHandler.Close()
	return resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}
