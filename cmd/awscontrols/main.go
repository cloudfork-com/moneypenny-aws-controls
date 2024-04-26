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

var isDebug = flag.Bool("debug", false, "if true then more logging")

var profile = flag.String("profile", "default", "run for a specific AWS profile")

var localOnly = flag.Bool("local", false, "if true then only use the local service plans file")

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
	pe.SetLocalPlansOnly(*localOnly)

	if slices.Contains(os.Args, "apply") {
		pe.Apply()
	} else if slices.Contains(os.Args, "report") {
		pe.Report()
	} else if slices.Contains(os.Args, "schedule") {
		pe.Schedule()
	} else {
		pe.Plan()
	}

}

func setupLog() {
	// set global logger with custom options
	lvl := slog.LevelInfo
	if *isDebug {
		lvl = slog.LevelDebug
	}
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      lvl,
			TimeFormat: time.Kitchen,
		}),
	))
}
