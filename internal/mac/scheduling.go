package mac

import (
	"fmt"
	"path"
	"strings"
	"time"
)

const Running = "RUNNING"
const Stopped = "STOPPED"
const Unknown = "UNKNOWN"

// ECS service
type Service struct {
	ARN               string `json:"service-arn"`
	DesiredTasksCount int    `json:"desired-tasks-count"`
}

func (s Service) Name() string {
	return path.Base(s.ARN)
}
func (s Service) ClusterARN() string {
	return strings.Replace(path.Dir(s.ARN), "service", "cluster", -1)
}
func (s Service) ClusterName() string {
	return path.Base(s.ClusterARN())
}

// https://eu-central-1.console.aws.amazon.com/ecs/v2/clusters/CICD/services/cockpit-cockpit-dev/tags?region=eu-central-1
func (s Service) TagsURL() string {
	region := "eu-central-1" // from ENV
	return fmt.Sprintf("https://%s.console.aws.amazon.com/ecs/v2/clusters/%s/services/%s/tags?region=%s",
		region, s.ClusterName(), s.Name(), region)
}

// Single occurrence in time
type ScheduledEvent struct {
	Service
	DesiredState string    `json:"desired-state"`
	At           time.Time `json:"at"`
}

func (s ScheduledEvent) String() string {
	return fmt.Sprintf("on [%s] the desired state of service [%s] is [%s]", s.At.Format(time.DateTime), s.Name(), s.DesiredState)
}

// Weekly plan

type DayPlan struct {
	Weekday time.Weekday `json:"weekday"` // 0=Sunday, 1=Monday, ...
	Plans   []*TimePlan  `json:"plans"`
}

func (d *DayPlan) AddStateChange(service Service, change *StateChange) {
	// deduplicate
	for _, each := range d.Plans {
		if each.ARN == service.ARN && each.DesiredState == change.DesiredState && each.Hour == change.CronSpec.Hour && each.Minute == change.CronSpec.Minute {
			return
		}
	}
	d.Plans = append(d.Plans, &TimePlan{
		Service:      service,
		Hour:         change.CronSpec.Hour,
		Minute:       change.CronSpec.Minute,
		DesiredState: change.DesiredState,
		cron:         change.Cron,
	})
}

type TimePlan struct {
	Service
	DesiredState string `json:"desired-state"`
	Hour         int    `json:"hour"` // 24
	Minute       int    `json:"minute"`
	cron         string // what was used to create this
	doesNotExist bool   // verified with AWS, for reporting
}

func (t TimePlan) String() string {
	return fmt.Sprintf("on [%dH:%dM] the state of service [%s] is changed to [%s]", t.Hour, t.Minute, t.Service.Name(), t.DesiredState)
}
