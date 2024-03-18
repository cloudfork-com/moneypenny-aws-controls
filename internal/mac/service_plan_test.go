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
