package mac

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestReportSchedule(t *testing.T) {
	c, _ := NewECSClient("default")
	e := NewPlanExecutor(c, []*ServicePlan{}, []types.Service{}, "default")
	r := NewReporter(e)
	r.Report()
}
