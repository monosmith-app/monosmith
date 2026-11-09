package runhistory

import "testing"

func TestStepSelectionStaysInsideTheList(t *testing.T) {
	ids := []string{"", "a:build", "b:build"}
	for _, c := range []struct {
		sel  string
		dir  int
		want string
	}{
		{"", 1, "a:build"},
		{"a:build", 1, "b:build"},
		{"b:build", 1, "b:build"},
		{"a:build", -1, ""},
		{"", -1, ""},
		{"gone", 1, "a:build"},
	} {
		if got := StepSelection(ids, c.sel, c.dir); got != c.want {
			t.Errorf("StepSelection(%q, %d) = %q, want %q", c.sel, c.dir, got, c.want)
		}
	}
}

func TestStepRecentStaysBetweenTheFirstAndTheLastRun(t *testing.T) {
	if StepRecent(0, 3, -1) != 0 || StepRecent(2, 3, 1) != 2 || StepRecent(1, 3, 1) != 2 || StepRecent(0, 0, 1) != 0 {
		t.Fatal("the choice leaves the list")
	}
}
