package mac

import (
	"testing"
	"time"
)

func TestLastScheduledEventAtOnlyStopped(t *testing.T) {
	svc := Service{ARN: "test"}
	sp := ServicePlan{Service: svc, TagValue: "stopped=0 0 0."} // stop on sunday
	sp.Validate()
	wp := new(WeekPlan)
	wp.AddServicePlan(sp)

	monday, _ := time.Parse(time.DateOnly, "2024-04-01")
	ev, ok := wp.LastScheduledEventAt(svc, monday)
	if !ok {
		t.Fail()
	}
	if ev.DesiredState != Stopped {
		t.Fail()
	}
}
