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

var debugging = flag.Bool("debug", false, "if true then more logging")

var profile = flag.String("profile", "", "run for a specific AWS profile. If empty then run for all known profiles")

func main() {
	flag.Parse()
	setupLog()

	// either one or all
	profiles := []string{}
	if *profile != "" {
		profiles = append(profiles, *profile)
	} else {
		profiles = mac.GetLocalAwsProfiles()
	}
	for _, each := range profiles {
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
