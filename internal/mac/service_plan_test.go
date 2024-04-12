package mac

import "testing"

func TestServicePlanMulti(t *testing.T) {
	input := "running=0 0 1-5. running=0 10 1-5. stopped=0 4 2-6. stopped=0 12 3/4."
	sp := new(ServicePlan)
	sp.TagValue = input
	if err := sp.Validate(); err != nil {
		t.Fatal(err)
	}
	t.Log(sp.StateChanges)
	t.Log(sp.PercentageRunning())
}
func TestServicePlanAlways(t *testing.T) {
	sp := new(ServicePlan)
	sp.TagValue = "running=0 0 0-6"
	if err := sp.Validate(); err != nil {
		t.Fatal(err)
	}
	p := sp.PercentageRunning()
	if p != 1 {
		t.Errorf("Expected 1, got %f", p)
	}
}
func TestServicePlanNever(t *testing.T) {
	sp := new(ServicePlan)
	sp.TagValue = "stopped=0 0 0-6"
	if err := sp.Validate(); err != nil {
		t.Fatal(err)
	}
	p := sp.PercentageRunning()
	if p != 0 {
		t.Errorf("Expected 0, got %f", p)
	}
}
func TestServicePlanOneDay(t *testing.T) {
	sp := new(ServicePlan)
	sp.TagValue = "running=0 0 1. stopped=0 0 2."
	if err := sp.Validate(); err != nil {
		t.Fatal(err)
	}
	p := sp.PercentageRunning()
	one7th := float32(1.0 / 7.0)
	if p != one7th {
		t.Errorf("Expected %f, got %f", one7th, p)
	}
}

func TestServicePlanOfficeTime(t *testing.T) {
	sp := new(ServicePlan)
	sp.TagValue = "running=0 9 1-5. stopped=0 17 1-5."
	if err := sp.Validate(); err != nil {
		t.Fatal(err)
	}
	p := int(sp.PercentageRunning() * 100)
	want := 23
	if p != want {
		t.Errorf("Expected %d, got %d", want, p)
	}
}
