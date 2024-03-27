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

var plansInput = flag.String("plans", "", "description of service plans")

var debugging = flag.Bool("debug", false, "if true then more logging")

var profile = flag.String("profile", "default", "run for a specific AWS profile")

func main() {
	flag.Parse()
	setupLog()

	loader := mac.NewPlanLoader(*plansInput)
	if err := loader.LoadServicePlans(); err != nil {
		return
	}
	pe, err := mac.NewPlanExecutor(loader.Plans, *profile)
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

func setupLog() {
	// set global logger with custom options
	lvl := slog.LevelInfo
	if *debugging {
		lvl = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      lvl,
			TimeFormat: time.Kitchen,
		}),
	))
}
