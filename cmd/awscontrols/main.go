package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/cloudfork-com/moneypenny-aws-controls/internal/mac"
)

var servicesInput = flag.String("services", "aws-service-plans.json", "description of service schedules")

var dryRun = true

func main() {
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "apply" {
		slog.SetDefault(slog.With("exec", "APPLY"))
		dryRun = false
	} else {
		slog.SetDefault(slog.With("exec", "PLAN"))
	}

	data, err := os.ReadFile(*servicesInput)
	if err != nil {
		slog.Error("read fail", "err", err)
		return
	}
	plans := []*mac.ServicePlan{}
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
	}
	// build weekplan
	wp := new(mac.WeekPlan)
	for _, each := range plans {
		wp.AddServicePlan(*each)
	}
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("config fail", "err", err)
		return
	}
	client := ecs.NewFromConfig(cfg)

	list, err := mac.AllServices(client, "moneypenny")
	if err != nil {
		slog.Error("AllServices fail", "err", err)
		return
	}
	for _, each := range list {
		sp := new(mac.ServicePlan)
		sp.ARN = *each.ServiceArn
		input := mac.TagValue(each, "moneypenny")
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

	now := time.Now()
	for _, each := range plans {
		if each.Disabled {
			continue
		}
		event, ok := wp.LastScheduledEventAt(each.Service, now)
		if ok {
			lastStatus := mac.ServiceStatus(client, each.Service)
			if lastStatus == mac.Unknown {
				slog.Warn("service has unknown last status, it may not exist", "name", each.Service.Name())
				continue
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
