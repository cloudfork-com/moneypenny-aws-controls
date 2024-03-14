package mac

import "testing"

func TestCronSpecDOWRange(t *testing.T) {
	s, err := ParseCronSpec("0 18 1-5")
	t.Log(s, err)
}
func TestCronSpecDOWCommas(t *testing.T) {
	s, err := ParseCronSpec("0 18 1,2,4,5")
	t.Log(s, err)
}
