package mac

import (
	"testing"
)

func TestReportSchedule(t *testing.T) {
	c, _ := NewECSClient()
	e := NewPlanExecutor(c, []*ServicePlan{})
	r := NewReporter(e)
	r.Report()
}
