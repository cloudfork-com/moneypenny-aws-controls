package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/emicklei/htmlslog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
)

var Version string = "dev"

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "text/html; charset=UTF-8"},
		StatusCode: 200}

	// setup logging
	logBuffer := new(bytes.Buffer)
	stdoutHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: removeTimeAndLevel})
	logHandler := htmlslog.New(logBuffer, htmlslog.Options{
		Level:              slog.LevelDebug,
		Title:              "moneypenny-aws-controls",
		PassthroughHandler: stdoutHandler,
		TableOnly:          true})
	slog.SetDefault(slog.New(logHandler))

	pe, err := mac.NewPlanExecutor([]*mac.ServicePlan{}, "") // default profile
	if err != nil {
		logHandler.Close()
		resp.StatusCode = 500
		resp.Body = logBuffer.String()
		return resp, err
	}
	rep := mac.NewReporter(pe)
	action := req.QueryStringParameters["do"]
	switch action {
	case "apply":
		pe.Apply()
	case "plan":
		pe.Plan()
	default: // all
		if err := pe.BuildWeekPlan(); err != nil {
			logHandler.Close()
			resp.StatusCode = 500
			resp.Body = logBuffer.String()
			return resp, err
		}
		html := new(bytes.Buffer)
		rep.WriteOpenHTMLOn(html)

		fmt.Fprintln(html, "<h2>Status</h2>")
		if err := rep.WriteStatusOn(html); err != nil {
			logHandler.Close()
			resp.StatusCode = 500
			resp.Body = logBuffer.String()
			return resp, err
		}
		fmt.Fprintln(html, "<h2>Schedule</h2>")
		if err := rep.WriteScheduleOn(html); err != nil {
			logHandler.Close()
			resp.StatusCode = 500
			resp.Body = logBuffer.String()
			return resp, err
		}
		fmt.Fprintln(html, "<h2>Log</h2>")
		logHandler.Close()
		fmt.Fprintln(html, logBuffer.String())

		versionOn(html)
		rep.WriteCloseHTMLOn(html)
		resp.Body = html.String()
		return resp, nil
	}
	// all but report
	logHandler.Close()
	html := new(bytes.Buffer)
	rep.WriteOpenHTMLOn(html)
	html.WriteString(logBuffer.String())
	versionOn(html)
	rep.WriteOpenHTMLOn(html)

	resp.Body = html.String()
	return resp, nil
}

func removeTimeAndLevel(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" || a.Key == "level" {
		return slog.Attr{}
	}
	return a
}

func versionOn(w io.Writer) {
	fmt.Fprintf(w, "<p style='font-size: 10px;'>moneypenny-aws-controls: %s</p>", Version)
}
