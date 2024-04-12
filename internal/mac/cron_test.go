package mac

import (
	"testing"
	"time"
)

func TestCronSpecDOWRange(t *testing.T) {
	s, err := ParseCronSpec("0 18 1-5")
	t.Log(s, err)
}
func TestCronSpecOneDOW(t *testing.T) {
	s, err := ParseCronSpec("0 18 1")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.DaysOfWeek) != 1 {
		t.Fatal("one day expected")
	}
	if s.DaysOfWeek[0] != time.Weekday(1) {
		t.Fatal("wrong day")
	}
}
func TestCronSpecDOWSlashes(t *testing.T) {
	s, err := ParseCronSpec("0 18 1/2/4/5")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.DaysOfWeek) != 4 {
		t.Fatal("4 days expected")
	}
}

func TestCronSpecDOWSlashesFail(t *testing.T) {
	_, err := ParseCronSpec("0 18 1/2/2/4/5")
	if err == nil {
		t.Fail()
	}
}
