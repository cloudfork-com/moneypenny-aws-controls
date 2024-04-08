package mac

import "testing"

func TestReport(t *testing.T) {
	e, _ := NewPlanExecutor([]*ServicePlan{}, "default")
	r := NewReporter(e)
	r.Report()
}
