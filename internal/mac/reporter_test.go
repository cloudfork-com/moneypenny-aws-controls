package mac

import (
	"testing"
)

func TestReportSchedule(t *testing.T) {
	c, _ := NewECSClient("default")
	e := NewPlanExecutor(c, []*ServicePlan{}, "default")
	r := NewReporter(e)
	r.Report()
}
