package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
	"github.com/lmittmann/tint"
)

var servicesInput = flag.String("services", "", "description of service schedules")

var dryRun = true
var reporting = false

var serviceTagName = "moneypenny"

func main() {
	flag.Parse()
	setupLog()

	if slices.Contains(os.Args, "apply") {
		slog.SetDefault(slog.With("exec", "APPLY"))
		dryRun = false
	} else if slices.Contains(os.Args, "report") {
		slog.SetDefault(slog.With("exec", "REPORT"))
		reporting = true
	} else {
		slog.SetDefault(slog.With("exec", "PLAN"))
	}

	// build weekplan
	wp := new(mac.WeekPlan)
	plans := []*mac.ServicePlan{}

	if len(*servicesInput) == 0 {
		slog.Info("no local service plans")
	} else {
		slog.Info("reading service plans", "file", *servicesInput)
		data, err := os.ReadFile(*servicesInput)
		if err != nil {
			slog.Error("read fail", "err", err)
			return
		}
		err = json.Unmarshal(data, &plans)
		if err != nil {
			slog.Error("parse fail", "err", err)
			return
		}
		for _, each := range plans {
			if err := each.Validate(); err != nil {
				slog.Error("validate fail", "err", err)
				return
			}
			slog.Info("adding service plan", "service", each.ARN, "disabled", each.Disabled)
			wp.AddServicePlan(*each)
		}
	}

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("config fail", "err", err)
		return
	}
	client := ecs.NewFromConfig(cfg)

	allServices, err := mac.AllServices(client)
	if err != nil {
		slog.Error("AllServices fail", "err", err)
		return
	}
	for _, each := range allServices {
		input := mac.TagValue(each, serviceTagName)
		if input == "" {
			continue
		}
		sp := new(mac.ServicePlan)
		sp.ARN = *each.ServiceArn
		slog.Info("adding service plan", "service", *each.ServiceArn, "crons", input)
		if input == "" {
			slog.Warn("invalid moneypenny tag value", "value", input, "err", err)
			continue
		}
		chgs, err := mac.ParseStateChanges(input)
		if err != nil {
			slog.Warn("invalid moneypenny tag value", "value", input, "err", err)
			continue
		}
		sp.TagValue = input
		sp.StateChanges = chgs
		plans = append(plans, sp)
		wp.AddServicePlan(*sp)
	}
	if !reporting {
		now := time.Now()
		for _, each := range plans {
			if each.Disabled {
				slog.Warn("disabled plan", "service", each.ARN)
				continue
			}
			event, ok := wp.LastScheduledEventAt(each.Service, now)
			if ok {
				lastStatus := mac.ServiceStatus(client, each.Service)
				if lastStatus == mac.Unknown {
					exitsInCluster := slices.ContainsFunc(allServices, func(existing types.Service) bool {
						return *existing.ServiceArn == each.ARN
					})
					if exitsInCluster {
						slog.Warn("service has unknown last status, assume it is stopped", "name", each.Service.Name())
						lastStatus = mac.Stopped
					} else {
						slog.Warn("service does not exist, update your configuration", "name", each.Service.Name())
						continue
					}
				}
				isRunning := lastStatus == mac.Running
				if event.DesiredState != mac.Running && isRunning {
					slog.Info("service is running but must be stopped", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
					if dryRun {
						continue
					}
					if err := mac.StopService(client, each.Service); err != nil {
						slog.Error("failed to stop service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
					}
				} else if event.DesiredState == mac.Running && !isRunning {
					slog.Info("service must be running", "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
					if dryRun {
						continue
					}
					if err := mac.StartService(client, each.Service); err != nil {
						slog.Error("failed to start service", "err", err, "name", each.Service.Name(), "state", lastStatus, "crons", each.TagValue)
					}
				} else {
					slog.Info("service is in expected state", "name", each.Service.Name(), "state", event.DesiredState, "crons", each.TagValue)
				}
			}
		}
	}
	if !reporting {
		return
	}
	rout, _ := os.Create("schedule.html")
	defer rout.Close()
	slog.Info("write schedule")
	rep := mac.Reporter{}
	if err := rep.WriteOn(wp, rout); err != nil {
		slog.Error("schedule report failed", "err", err)
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
