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

func TestTheHeaderOnItsOwnRowSitsAtTheRightEnd(t *testing.T) {
	l := newLab(t, true, 110, 30)
	row := l.sh().headerRow()
	if !strings.HasPrefix(row, "   ") || strings.HasSuffix(row, "  ") || !strings.Contains(row, "acme") {
		t.Fatalf("the workspace and the branch belong at the right end: %q", row)
	}
}

func TestTheWorkspaceButtonsOfferToClearTheCache(t *testing.T) {
	l := newLab(t, true, 160, 30)
	var ids []string
	for _, row := range l.sh().workspaceButtons(200, "") {
		for _, sp := range row.hit.spans {
			ids = append(ids, sp.id)
		}
	}
	if len(ids) == 0 || ids[len(ids)-1] != "ws.cache-clear" {
		t.Fatalf("the last workspace button deletes the local cache: %v", ids)
	}
}
