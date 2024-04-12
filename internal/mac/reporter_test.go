package mac

import "testing"

func TestReportSchedule(t *testing.T) {
	e, _ := NewPlanExecutor([]*ServicePlan{}, "default")
	r := NewReporter(e)
	r.Report()
}
