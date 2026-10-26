package fixwizard

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
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

// elapsedText says how long the step has been running, once it is long enough
// to matter: nothing under three seconds.
func elapsedText(d time.Duration) string {
	if d < 3*time.Second {
		return ""
	}
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf("Running for %ds", s)
	}
	return fmt.Sprintf("Running for %dm %02ds", s/60, s%60)
}

// elapsed is how long the applying phase has lasted.
func (w Wizard) elapsed() time.Duration {
	if w.started.IsZero() {
		return 0
	}
	return time.Since(w.started)
}
