package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var Version string = "dev"

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers:    map[string]string{"Content-Type": "text/html; charset=UTF-8"},
		StatusCode: http.StatusOK}

	// https://stackoverflow.com/questions/58037317/getting-x-amzn-remapped-www-authenticate-instead-of-www-authenticate-and-jetty
	// auth check
	if len(req.Headers) == 0 {
		slog.Info("function is not invoked from public APIGateway so no user credentials check needed")
	} else {
		httpreq, _ := http.NewRequest(http.MethodGet, "/", nil)
		httpreq.Header.Add("Authorization", req.Headers["authorization"]) // must be lowercase
		user, pass, ok := httpreq.BasicAuth()
		if !ok || user != os.Getenv("BASIC_USER") || pass != os.Getenv("BASIC_PASSWORD") {
			slog.Warn("invalid credentials", "user-length", len(user), "pass-length", len(pass), "request-headers", req.Headers)
			resp.Headers["WWW-Authenticate"] = `Basic realm="Restricted"`
			resp.StatusCode = http.StatusUnauthorized
			return resp, nil
		}
	}
	return resp, nil
}
