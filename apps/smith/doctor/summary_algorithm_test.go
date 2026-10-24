package doctor

import "testing"

func TestSummaryCountsWhatItLeavesOut(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	if got := Summary(items, "; ", 3); got != "a; b; c; and 2 more" {
		t.Fatalf("%q", got)
	}
	if got := Summary(items[:3], "; ", 3); got != "a; b; c" {
		t.Fatalf("%q", got)
	}
}
