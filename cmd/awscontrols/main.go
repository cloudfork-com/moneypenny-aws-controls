package main

import (
	"flag"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
	"github.com/lmittmann/tint"
)

var servicesInput = flag.String("services", "", "description of service schedules")

func main() {
	flag.Parse()
	setupLog()

	for _, each := range mac.GetLocalAwsProfiles() {
		pe, err := mac.NewPlanExecutor(*servicesInput, each)
		if err != nil {
			return
		}
		if slices.Contains(os.Args, "apply") {
			pe.Apply()
		} else if slices.Contains(os.Args, "report") {
			pe.Report()
		} else {
			pe.Plan()

		}
	}
}

func setupLog() {
	// set global logger with custom options
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))
}
