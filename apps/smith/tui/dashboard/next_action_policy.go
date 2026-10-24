package dashboard

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
)

// NextAction chooses the one thing to do next: what is broken first, then the
// last run that failed, then the work the changes left, then what is only
// worth a look, and, when there is nothing, a build.
func NextAction(f Facts) Next {
	press := func(id string) string { return "press " + commands.ShortcutOf(id) }
	switch {
	case f.CheckErr != "":
		return Next{Text: "doctor could not run: " + press("ws.doctor") + " to try again", CommandID: "ws.doctor"}
	case !f.Checked:
		return Next{Text: "checking the workspace…"}
	case f.Failed > 0:
		return Next{Text: fmt.Sprintf("%s: %s", count(f.Failed, "check fails", "checks fail"), press("ws.doctor")), CommandID: "ws.doctor"}
	case f.Last != nil && !f.Last.OK && !f.Last.Cancelled:
		return Next{Text: "the last run failed: " + press("view.run"), CommandID: "view.run"}
	case f.Affected > 0:
		return Next{Text: fmt.Sprintf("%s: %s to lint and test them", count(f.Affected, "project affected", "projects affected"), press("run.affected")), CommandID: "run.affected"}
	case f.Warnings > 0:
		return Next{Text: fmt.Sprintf("%s: %s", count(f.Warnings, "warning", "warnings"), press("ws.doctor")), CommandID: "ws.doctor"}
	case f.DepsChecked && f.Outdated > 0:
		return Next{Text: fmt.Sprintf("%s: %s", count(f.Outdated, "dependency has a newer release", "dependencies have newer releases"), press("view.deps")), CommandID: "view.deps"}
	}
	return Next{Text: "all good: " + press("run.build") + " to build", CommandID: "run.build"}
}

// count is "1 check fails" or "3 checks fail".
func count(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
