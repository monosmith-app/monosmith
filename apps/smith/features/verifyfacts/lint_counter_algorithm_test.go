package verifyfacts

import (
	"testing"
)

func TestLintCounterReadsEslintGolangciAndRuff(t *testing.T) {
	cases := map[string]struct {
		line   string
		issues int
	}{
		"eslint":        {"✖ 3 problems (2 errors, 1 warning)", 3},
		"eslint one":    {"✖ 1 problem (1 error, 0 warnings)", 1},
		"golangci":      {"7 issues:", 7},
		"golangci none": {"0 issues.", 0},
		"ruff":          {"Found 4 errors.", 4},
		"ruff clean":    {"All checks passed!", 0},
	}
	for name, c := range cases {
		var l LintCounter
		l.Add(c.line)
		if issues, known := l.Issues(); !known || issues != c.issues {
			t.Errorf("%s: got %d, known=%v, want %d", name, issues, known, c.issues)
		}
	}
	var l LintCounter
	l.Add("linting...")
	if _, known := l.Issues(); known {
		t.Error("a linter that said nothing readable has a count")
	}
}
