package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/runner"
)

// searchableRun is a run that ended, with a log that says "panic" twice.
func searchableRun(t *testing.T, w, h int) *lab {
	t.Helper()
	l := newLab(t, true, w, h)
	l.next = &scripted{lines: []string{
		"kit:build ▶ go build",
		"kit:build │ compiling",
		"kit:build │ panic: nil map",
		"kit:build │ goroutine 1",
		"kit:build │ PANIC again",
		"kit:build × exited with code 2",
	}, exit: runner.Exit{Code: 2}}
	l.keys("B")
	return l
}

func TestSlashSearchesTheLogAndNGoesFromMatchToMatch(t *testing.T) {
	l := searchableRun(t, 150, 44)
	if l.sh().tab != commands.TabRuns {
		t.Fatalf("the run is in the Runs view: %q", l.sh().tab)
	}
	l.keys("/")
	if !l.sh().logTyping {
		t.Fatal("/ gives the keyboard to the search")
	}
	l.typeText("panic")
	if len(l.started) != 1 {
		t.Fatalf("typing ran a key: %v", l.started)
	}
	l.has("Search: panic▏", "1 of 2")
	l.keys("enter")
	l.has("Search: panic  1 of 2")
	l.keys("n")
	l.has("2 of 2")
	l.keys("n")
	l.has("1 of 2")
	l.keys("N")
	l.has("2 of 2")
	if foot := l.lines()[l.sh().lay().Footer.Y]; !strings.Contains(foot, "Esc clear search") {
		t.Fatalf("Esc clears the search: %q", foot)
	}
	view, _ := commands.Hints(l.sh().state())
	found := false
	for _, h := range view {
		found = found || (h.Glyph == "n/N" && h.Label == "next match")
	}
	if !found {
		t.Fatalf("n/N is a hint of the view while the log has a search: %v", view)
	}
	l.keys("esc")
	if l.sh().logQuery != "" || l.sh().tab != commands.TabRuns {
		t.Fatalf("Esc clears the search and stays: %q, %q", l.sh().logQuery, l.sh().tab)
	}
	l.hasNot("Search:")
}

func TestASearchThatFindsNothingSaysSo(t *testing.T) {
	l := searchableRun(t, 150, 44)
	l.keys("/")
	l.typeText("zebra")
	l.keys("enter")
	l.has("Search: zebra  no match")
	l.keys("n")
	if !strings.Contains(l.sh().statusInfo().Hint, "Nothing in this log matches zebra") {
		t.Fatalf("n says there is nothing to go to: %q", l.sh().statusInfo().Hint)
	}
}

func TestTheSearchFitsEveryScreen(t *testing.T) {
	for _, size := range [][2]int{{80, 30}, {120, 40}, {150, 44}} {
		l := searchableRun(t, size[0], size[1])
		l.keys("/")
		l.typeText("panic")
		lines := l.lines()
		if len(lines) != size[1] {
			t.Fatalf("%v: %d rows", size, len(lines))
		}
		for y, line := range lines {
			if pane.Width(line) != size[0] {
				t.Fatalf("%v: row %d is %d cells: %q", size, y, pane.Width(line), line)
			}
		}
		l.has("panic▏")
	}
}
