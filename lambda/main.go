package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"

	"github.com/emicklei/htmlslog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
)

var Version string = "dev"

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "text/html; charset=UTF-8"},
		StatusCode: 200}

	// setup logging
	logBuffer := new(bytes.Buffer)
	stdoutHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: removeTimeAndLevel})
	logHandler := htmlslog.New(logBuffer, htmlslog.Options{Level: slog.LevelDebug, Title: "moneypenny-aws-controls", PassthroughHandler: stdoutHandler})
	slog.SetDefault(slog.New(logHandler))

	pe, err := mac.NewPlanExecutor([]*mac.ServicePlan{}, "") // default profile
	if err != nil {
		logHandler.Close()
		resp.StatusCode = 500
		resp.Body = logBuffer.String()
		return resp, err
	}
	action := req.QueryStringParameters["do"]
	switch action {
	case "apply":
		pe.Apply()
	case "report":
		html := new(bytes.Buffer)
		if err := pe.ReportHTMLOn(html); err != nil {
			logHandler.Close()
			resp.StatusCode = 500
			resp.Body = logBuffer.String()
			return resp, err
		}
		resp.Body = html.String()
		return resp, nil
	default:
		pe.Plan()
	}
	logHandler.Close()
	resp.Body = logBuffer.String()
	return resp, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func removeTimeAndLevel(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" || a.Key == "level" {
		return slog.Attr{}
	}
	return a
}
