package release

import "testing"

func TestMarkdownIsTheWholeResultSoNoTextPlanComesFirst(t *testing.T) {
	for _, c := range []struct {
		asJSON bool
		format string
		want   bool
	}{{false, "text", true}, {false, "markdown", false}, {true, "text", false}} {
		if got := printsTextPlan(c.asJSON, c.format); got != c.want {
			t.Errorf("json=%v format=%s: %v, want %v", c.asJSON, c.format, got, c.want)
		}
	}
}
