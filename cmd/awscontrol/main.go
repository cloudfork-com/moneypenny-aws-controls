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

var tasksInput = flag.String("tasks", "taskplans.json", "description of task schedules")

// strategy:
// get all tasks
// describe all tasks
// find by tag name
// check desired state: stop if needed, start if needed
// for stop, the taskARN is needed
// for start, the taskDefinitionARN is needed (latest?)
func main() {
	flag.Parse()

	data, err := os.ReadFile(*tasksInput)
	if err != nil {
		slog.Error("read fail", "err", err)
		return
	}
	plans := []*mac.TaskPlan{}
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
	services, err := mac.AllServices(client)
	if err != nil {
		slog.Error("all tasks fail", "err", err)
		return
	}
	now := time.Now()
	for _, each := range services {
		sn := mac.NameOfService(each)
		if sn == "" { // never matches
			continue
		}
		for _, other := range plans {
			if other.Task.Name == mac.NameOfService(each) {
				for _, change := range other.StateChanges {
					which, err := mac.TaskForService(client, each)
					if err != nil {
						slog.Error("all tasks fail", "err", err)
						return
					}
					slog.Info("service", "definition", *each.TaskDefinition, "tasks", len(which), "new-state", change.DesiredState, "on", change.Next(now))
				}
			}
		}
	}
}
