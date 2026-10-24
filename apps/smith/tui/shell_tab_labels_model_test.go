package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

func TestTheTabNamesDoNotChangeWhenABadgeAppears(t *testing.T) {
	for w := 70; w <= 150; w++ {
		l := newLab(t, true, w, 30)
		l.next = &scripted{lines: []string{"kit:build ✖ exited with code 2"}, exit: runner.Exit{Code: 1}}
		names := func() string {
			var out []string
			for _, id := range l.sh().tabs {
				out = append(out, stripBadge(pane.Strip(l.sh().tabLabel(id))))
			}
			return strings.Join(out, "|")
		}
		before := names()
		l.keys("ctrl+b")
		l.advance(time.Second)
		l.keys("alt+1")
		if after := names(); after != before {
			t.Errorf("%d columns: tab names changed from %q to %q", w, before, after)
		}
	}
}

// stripBadge keeps only the name of a tab label.
func stripBadge(label string) string {
	f := strings.Fields(label)
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

func TestTheAddProjectWizardIsTitledAddAProject(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+n")
	for _, line := range l.lines() {
		if strings.Contains(line, "╭") && strings.Contains(line, "Add a project") {
			return
		}
	}
	t.Fatalf("no panel titled Add a project:\n%s", l.view())
}
