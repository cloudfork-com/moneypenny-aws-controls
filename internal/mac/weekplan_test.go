package mac

import (
	"testing"
	"time"
)

func TestLastScheduledEventAtOnlyStopped(t *testing.T) {
	svc := Service{ARN: "test"}
	sp := ServicePlan{Service: svc, TagValue: "stopped=0 0 0-6."}
	sp.Validate()
	wp := new(WeekPlan)
	wp.AddServicePlan(sp)

	ev, ok := wp.LastScheduledEventAt(svc, time.Now())
	t.Log(ev, ok)
}
