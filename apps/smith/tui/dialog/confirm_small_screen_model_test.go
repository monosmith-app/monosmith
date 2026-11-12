package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestOnASmallScreenAWaitingConfirmKeepsItsSpinnerAndScrollsItsHead(t *testing.T) {
	why := strings.Repeat("The form asks which build outputs to cache, so smith reads the plan first. ", 3)
	c := NewWaiting("Convert", why, []string{"convert", "--dry-run"}, "Reading the plan").Within(3, 2).Place(100, 16)
	r := c.Rect()
	if r.Y < 3 || r.Y+r.H > 14 {
		t.Fatalf("the dialog covers the tabs or the footer: %+v", r)
	}
	text := pane.Strip(strings.Join(c.Lines(theme.Default()), "\n"))
	if !strings.Contains(text, "Reading the plan…") || !strings.Contains(text, "↓ more") {
		t.Fatalf("the spinner must stay in view under a head that scrolls:\n%s", text)
	}
	c, out := c.Key("pgdown")
	if out.Done || !strings.Contains(pane.Strip(strings.Join(c.Lines(theme.Default()), "\n")), "↑ more") {
		t.Fatalf("PgDn scrolls the head while the plan is read:\n%s", pane.Strip(strings.Join(c.Lines(theme.Default()), "\n")))
	}
}
