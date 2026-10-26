package fixwizard

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// stepListRows lists every step by name, so the user sees what is done and
// what is coming: a tick for finished steps, a dot for the ones still to run.
// The running step has its own line above, so it is left out. Steps without
// a description (a job that only counts them) give no list.
func stepListRows(th theme.Theme, steps []protocol.FixStep, current, inner int) []string {
	named := 0
	for _, s := range steps {
		if s.Description != "" {
			named++
		}
	}
	if named < 2 {
		return nil
	}
	rows := []string{""}
	for i, s := range steps {
		n := i + 1
		if n == current || s.Description == "" {
			continue
		}
		mark := th.Muted("·")
		if n < current {
			mark = th.Muted("✓")
		}
		rows = append(rows, pane.Wrap(fmt.Sprintf("%s step %d: %s", mark, n, s.Description), inner, 2)...)
	}
	return rows
}

// capitalFirst starts a step name with a capital, for a line that stands alone.
func capitalFirst(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return strings.ToUpper(string(r[0])) + string(r[1:])
}

// elapsedText says how long the step has been running, once it is long enough
// to matter: nothing under three seconds.
func elapsedText(d time.Duration) string {
	if d < 3*time.Second {
		return ""
	}
	return "Running for " + progress.Elapsed(d)
}

// elapsed is how long the applying phase has lasted.
func (w Wizard) elapsed() time.Duration {
	if w.started.IsZero() {
		return 0
	}
	return time.Since(w.started)
}
