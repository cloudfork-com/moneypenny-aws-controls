package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"

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
		slog.Info("apply state changes")
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
	// now := time.Now()
	client := ecs.NewFromConfig(cfg)
	// for _, each := range plans {
	// 	event, ok := wp.LastScheduledEventAt(each.Service, now)
	// 	if ok {
	// 		isRunning := mac.IsServiceRunning(client, each.Service)
	// 		if event.DesiredState != mac.Running && isRunning {
	// 			slog.Info("service is running but must be stopped", "name", each.Service.Name())
	// 			if dryRun {
	// 				continue
	// 			}
	// 			if err := mac.StopService(client, each.Service); err != nil {
	// 				slog.Error("failed to stop service", "err", err, "name", each.Service.Name())
	// 			}
	// 		}
	// 		if event.DesiredState == mac.Running && !isRunning {
	// 			slog.Info("service has stopped but must be running", "name", each.Service.Name())
	// 			if dryRun {
	// 				continue
	// 			}
	// 			if err := mac.StartService(client, each.Service); err != nil {
	// 				slog.Error("failed to start service", "err", err, "name", each.Service.Name())
	// 			}
	// 		}
	// 	}
	// }
	list, err := mac.AllServices(client, "moneypenny")
	if err != nil {
		slog.Error("AllServices fail", "err", err)
		return
	}
	for _, each := range list {
		slog.Info("moneypenny controlled", "name", *each.ServiceName)
	}
}

func Star[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}
