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

var localOnly = flag.Bool("local", false, "if true then only use the local service plans file")

func main() {
	flag.Parse()
	setupLog()

	slog.Info("awscontrols - scheduling ECS services")
	loader := mac.NewPlanLoader(*plansInput)
	if err := loader.LoadServicePlans(); err != nil {
		return
	}
	client, err := mac.NewECSClient()
	if err != nil {
		return
	}
	fetcher := mac.NewPlanFetcher(client)
	if err := fetcher.CheckServicePlans(loader.Plans); err != nil {
		return
	}
	executor := mac.NewPlanExecutor(client, loader.Plans)

	if slices.Contains(os.Args, "apply") {
		executor.Apply()
	} else if slices.Contains(os.Args, "report") {
		executor.Report()
	} else if slices.Contains(os.Args, "schedule") {
		executor.Schedule()
	} else {
		executor.Plan()
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
