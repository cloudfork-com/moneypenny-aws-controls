package mac

import (
	"fmt"
	"time"

	cron "github.com/robfig/cron/v3"
)

const Running = "RUNNING"
const Stopped = "STOPPED"

// ECS task
type Task struct {
	Name string `json:"tags-name"`
}

// Single occurrence in time
type ScheduledEvent struct {
	Task
	DesiredState string    `json:"desired-state"`
	At           time.Time `json:"at"`
}

func (s ScheduledEvent) String() string {
	return fmt.Sprintf("on [%s] [%s] changes state to [%s]", s.At.Format(time.DateTime), s.Name, s.DesiredState)
}

type ProcessedEvent struct {
	ScheduledEvent
	TaskDefinitionARN string    // last known task definition when stopped or started
	ActualState       string    `json:"actual-state"`
	At                time.Time `json:"at"`
}

// Weekly plan

type WeekPlan struct {
	Plans []DayPlan `json:"plans"`
}

type DayPlan struct {
	Weekday int        `json:"weekday"` // 0=Sunday, 1=Monday, ...
	Plans   []TimePlan `json:"plans"`
}

type TimePlan struct {
	Task
	DesiredState string `json:"desired-state"`
	Hour         int    `json:"hour"` // 24
	Minute       int    `json:"minute"`
}

type TaskPlan struct {
	Task
	StateChanges []*StateChange `json:"state-changes"`
}
type StateChange struct {
	DesiredState string `json:"desired-state"`
	Cron         string `json:"cron"`
	schedule     cron.Schedule
}

func (t *TaskPlan) Validate() error {
	p := cron.NewParser(cron.Minute | cron.Hour | cron.Dow)
	for _, each := range t.StateChanges {
		sched, err := p.Parse(each.Cron)
		if err != nil {
			return err
		}
		each.schedule = sched
	}
	return nil
}

// Next returns the next activation time, later than the given time.
// Next is invoked initially, and then each time the job is run.
func (s *StateChange) Next(w time.Time) time.Time {
	return s.schedule.Next(w)
}
