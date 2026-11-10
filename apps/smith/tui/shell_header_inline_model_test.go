package tui

import (
	"strings"
	"testing"
)

func TestOnAWideScreenTheWorkspaceAndBranchSitAtTheRightEndOfTheTabRow(t *testing.T) {
	l := newLab(t, true, 160, 30)
	first := l.lines()[0]
	for _, want := range []string{"Home (1)", "Settings (6)", "acme", "main"} {
		if !strings.Contains(first, want) {
			t.Fatalf("the first row lacks %q: %q", want, first)
		}
	}
	if strings.Index(first, "acme") < strings.Index(first, "Settings (6)") {
		t.Fatalf("the workspace must come after the tabs, at the right: %q", first)
	}
	if y := l.sh().lay().Main.Y; y != 1 {
		t.Fatalf("the content starts under the shared row, not %d", y)
	}
}

func TestBelowTheInlineWidthTheHeaderKeepsItsOwnRow(t *testing.T) {
	l := newLab(t, true, 110, 30)
	lines := l.lines()
	if !strings.Contains(lines[0], "acme") || strings.Contains(lines[0], "Home") {
		t.Fatalf("a narrow screen keeps the header alone on the first row: %q", lines[0])
	}
	if !strings.Contains(lines[1], "Home (1)") {
		t.Fatalf("and the tabs on the second: %q", lines[1])
	}
}
