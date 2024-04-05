package mac

import "testing"

func TestParseStateChanges(t *testing.T) {
	i := "running=0 8 1-5. stopped=0 18 1-5."
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(list)
}
func TestParseStateChangesNotDot(t *testing.T) {
	i := "running=0 8 1-5. stopped=0 18 1-5"
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(list)
}

func TestParseStateChangesWithCount(t *testing.T) {
	i := "running=0 8 1-5. stopped=0 18 1-5. count=7. "
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Fatal()
	}
	one := list[0]
	if one.DesiredCount != 7 {
		t.Fatal("expect 7")
	}
}

func TestParseStateChangesWithDisabled(t *testing.T) {
	i := "// running=0 8 1-5. stopped=0 18 1-5. count=7."
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatal()
	}
}

func TestParseStateChangesNoRunning(t *testing.T) {
	i := "running=0 8 1-5. // stopped=0 18 1-5"
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatal()
	}
}

func TestParseStateChangesWithCountNoRunning(t *testing.T) {
	i := "count=9"
	_, err := ParseStateChanges(i)
	if err != nil {
		t.Error("unexpected error")
	}
}

func TestParseStateChangesOneDayOneChange(t *testing.T) {
	i := "stopped=0 0 0"
	list, err := ParseStateChanges(i)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.FailNow()
	}
	one := list[0]
	if len(one.CronSpec.DaysOfWeek) == 0 {
		t.Fatal("one day expected")
	}
}
func TestParseStateChangesInvalid(t *testing.T) {
	i := "running=0 8 1-5 stopped=0 18 1-5"
	_, err := ParseStateChanges(i)
	if err == nil {
		t.Fail()
	}
	t.Log(err.Error())
}
