package tui

import "testing"

// TestSaidNotCommittedReadsTheLineSmithPrints checks that a run that asked to
// commit but printed "Not committed" is not offered the undo hint.
func TestSaidNotCommittedReadsTheLineSmithPrints(t *testing.T) {
	cases := map[string]struct {
		lines []string
		want  bool
	}{
		"committed":      {[]string{"Done: lib util is ready", "Committed: chore: add lib util"}, false},
		"other changes":  {[]string{"Done: lib util is ready", "Not committed. Your work tree had other changes."}, true},
		"indented":       {[]string{"  Not committed (--no-commit)"}, true},
		"no output":      {nil, false},
		"word in a path": {[]string{"wrote docs/Not committed.md"}, false},
	}
	for name, c := range cases {
		if got := saidNotCommitted(c.lines); got != c.want {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}
