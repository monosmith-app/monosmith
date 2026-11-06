package toolrun

import "testing"

// TestProblemUnless keeps a passing check free of a Problem.
func TestProblemUnless(t *testing.T) {
	if got := ProblemUnless(true, "x"); got != "" {
		t.Errorf("passing check has Problem %q", got)
	}
	if got := ProblemUnless(false, "x"); got != "x" {
		t.Errorf("failing check Problem = %q, want x", got)
	}
}
