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

var servicesInput = flag.String("services", "service-plans.json", "description of service schedules")

// strategy:
// get all tasks
// describe all tasks
// find by tag name
// check desired state: stop if needed, start if needed
// for stop, the taskARN is needed
// for start, the taskDefinitionARN is needed (latest?)
func main() {
	flag.Parse()

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
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("config fail", "err", err)
		return
	}
	client := ecs.NewFromConfig(cfg)
	// services, err := mac.AllServices(client)
	// if err != nil {
	// 	slog.Error("all tasks fail", "err", err)
	// 	return
	// }
	now := time.Now()
	// for _, each := range services {
	// 	sn := mac.NameOfService(each)
	// 	if sn == "" { // never matches
	// 		continue
	// 	}
	// 	for _, other := range plans {
	// 		if other.Service.Name == mac.NameOfService(each) {
	// 			for _, change := range other.StateChanges {
	// 				which, err := mac.TaskForService(client, each)
	// 				if err != nil {
	// 					slog.Error("all tasks fail", "err", err)
	// 					return
	// 				}
	// 				slog.Info("service", "definition", *each.TaskDefinition, "tasks", len(which), "new-state", change.DesiredState, "on", now)
	// 			}
	// 		}
	// 	}
	// }
	wp := new(mac.WeekPlan)
	for _, each := range plans {
		wp.AddServicePlan(*each)
	}
	// tps := wp.ScheduleForDay(now.Weekday())
	// for _, each := range tps {
	// 	tasks, err := mac.TaskForService2(client, each.Service.ClusterARN(), each.Service.Name())
	// 	if err != nil {
	// 		slog.Error("tasks for service fail", "err", err)
	// 		return
	// 	}
	// 	for _, other := range tasks {
	// 		slog.Info("on", "hour", each.Hour,
	// 			"minute", each.Minute,
	// 			"change state of service", each.Service.Name(),
	// 			"current", *other.LastStatus,
	// 			"to", each.DesiredState,
	// 			"using", *other.TaskDefinitionArn)
	// 	}
	// }
	for _, each := range plans {
		event, ok := wp.LastScheduledEventAt(each.Service, now)
		if ok {
			slog.Info("last scheduled", "event", event)
			tasks, _ := mac.TaskForService2(client, event.ClusterARN(), event.Name())
			for _, other := range tasks {
				slog.Info("task", "started", *other.StartedAt, "by", Star(other.StartedBy), "current-state", Star(other.LastStatus))
			}
		}
	}
}

func Star[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}
