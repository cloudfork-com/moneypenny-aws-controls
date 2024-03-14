package mac

import (
	_ "embed"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestDayPlan(t *testing.T) {
	task := Service{
		ARN: "myservice",
	}
	tpOn := TimePlan{
		Service:      task,
		DesiredState: Running,
		Hour:         9,
		Minute:       1,
	}
	tpOff := TimePlan{
		Service:      task,
		DesiredState: Stopped,
		Hour:         18,
		Minute:       2,
	}
	dp := &DayPlan{
		Weekday: 2,
		Plans:   []TimePlan{tpOn, tpOff},
	}
	wp := WeekPlan{
		Plans: []*DayPlan{dp},
	}
	for _, each := range wp.ScheduledEventsOn(time.Now()) {
		t.Logf("event: %v", each)
	}
	e, ok := wp.LastScheduledEventAt(task, time.Now().Add(8*time.Hour))
	t.Log(e, ok)

	json.NewEncoder(os.Stdout).Encode(wp)
}

func TestWeekPlanAddServicePlan(t *testing.T) {
	sp := ServicePlan{}
	json.Unmarshal(serviceplanspec, &sp)
	sp.Validate()
	wp := new(WeekPlan)
	wp.AddServicePlan(sp)
	json.NewEncoder(os.Stdout).Encode(wp)
	for _, each := range wp.ScheduleForDay(1) {
		t.Log(each)
	}
}

//go:embed service-plan.json
var serviceplanspec []byte
