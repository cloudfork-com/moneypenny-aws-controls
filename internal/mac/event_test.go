package mac

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestDayPlan(t *testing.T) {
	task := Task{
		Name: "task",
	}
	tpOn := TimePlan{
		Task:         task,
		DesiredState: Running,
		Hour:         9,
		Minute:       1,
	}
	tpOff := TimePlan{
		Task:         task,
		DesiredState: Stopped,
		Hour:         18,
		Minute:       2,
	}
	dp := DayPlan{
		Weekday: 2,
		Plans:   []TimePlan{tpOn, tpOff},
	}
	wp := WeekPlan{
		Plans: []DayPlan{dp},
	}
	for _, each := range wp.ScheduledEventsOn(time.Now()) {
		t.Logf("event: %v", each)
	}
	e, ok := wp.LastScheduledEventAt(task, time.Now().Add(8*time.Hour))
	t.Log(e, ok)

	json.NewEncoder(os.Stdout).Encode(wp)
}

//go:embed taskplan.json
var taskplanspec []byte

func TestTaskPlan(t *testing.T) {
	tp := new(TaskPlan)
	json.Unmarshal(taskplanspec, &tp)
	t.Log(tp)
	t.Log(tp.Validate())
	for _, each := range tp.StateChanges {
		t.Log(each.schedule.Next(time.Now()))
	}
}
