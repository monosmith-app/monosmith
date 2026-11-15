package doctorchecks

import "testing"

func TestValidAreaIsTheClosedSet(t *testing.T) {
	for _, a := range Areas {
		if !ValidArea(a) {
			t.Errorf("%s", a)
		}
	}
	if ValidArea("") || ValidArea("Config") || ValidArea("other") {
		t.Error("only the closed set is valid")
	}
}
