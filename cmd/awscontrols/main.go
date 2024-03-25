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

	pe, err := mac.NewPlanExecutor(*servicesInput)
	if err != nil {
		return
	}
	if slices.Contains(os.Args, "apply") {
		pe.Apply()
		return
	} else if slices.Contains(os.Args, "report") {
		pe.Report()
		return
	} else {
		pe.Plan()
		return
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
