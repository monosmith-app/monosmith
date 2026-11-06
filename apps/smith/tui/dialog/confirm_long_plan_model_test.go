package dialog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func longPlan() Confirm {
	var plan []string
	for i := 1; i <= 80; i++ {
		plan = append(plan, fmt.Sprintf("change %d: packages/app-%d/package.json", i, i))
	}
	return NewConfirm("Upgrade", "This moves the workspace.", []string{"upgrade", "--defaults"}).WithDetails(plan).Place(110, 32)
}

func TestALongPlanScrollsAndNeverHidesTheCommitBoxOrTheCommand(t *testing.T) {
	c := longPlan()
	for _, step := range []string{"", "pgdown", "down", "end"} {
		if step != "" {
			c, _ = c.Key(step)
		}
		text := ansi.Strip(strings.Join(c.Lines(theme.Default()), "\n"))
		for _, want := range []string{CommitLabel, "The command that will run:", "$ smith upgrade --defaults", "Run (enter)", "more lines"} {
			if !strings.Contains(text, want) {
				t.Fatalf("after %q the dialog lost %q:\n%s", step, want, text)
			}
		}
	}
	if r := c.Rect(); r.H > 32 {
		t.Fatalf("the dialog is %d rows on a 32-row screen", r.H)
	}
}

func TestScrollingShowsTheLaterLinesOfAPlan(t *testing.T) {
	c := longPlan()
	if strings.Contains(strings.Join(c.Lines(theme.Default()), "\n"), "change 70:") {
		t.Fatal("the end of the plan is on screen before any scrolling")
	}
	for range 40 {
		c, _ = c.Key("pgdown")
	}
	text := ansi.Strip(strings.Join(c.Lines(theme.Default()), "\n"))
	if !strings.Contains(text, "change 80:") {
		t.Fatalf("the last line of the plan is out of reach:\n%s", text)
	}
}

func TestTheCommitBoxStaysClickableWhenThePlanIsLong(t *testing.T) {
	c := longPlan()
	r := c.Rect()
	c, _ = c.Click(r.X+3, r.Y+1+c.commitRow())
	if !c.Commit {
		t.Fatal("clicking the box row did not tick it")
	}
}
