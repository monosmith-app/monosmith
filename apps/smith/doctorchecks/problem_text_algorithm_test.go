package doctorchecks

import (
	"testing"
)

func TestLeadProblemNamesTheFirstAndCountsTheRest(t *testing.T) {
	cases := map[string][]string{"": nil, "a": {"a"}, "a (and 2 more)": {"a", "b", "c"}}
	for want, items := range cases {
		if got := LeadProblem(items); got != want {
			t.Errorf("LeadProblem(%v) = %q, want %q", items, got, want)
		}
	}
}

func TestProblemWhenIsEmptyForAPassingCheck(t *testing.T) {
	if ProblemWhen(false, "x") != "" || ProblemWhen(true, "x") != "x" {
		t.Error("ProblemWhen must keep the text only for a failure")
	}
}
