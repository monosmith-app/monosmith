package dashboard

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// NextAction chooses the one thing to do next: what is broken first, then the
// last run that failed, then the work the changes left, then what is only
// worth a look, and, when there is nothing, a build.
func NextAction(f Facts) Next {
	press := func(id string) string { return words.Say("dashboard.press", "shortcut", commands.ShortcutOf(id)) }
	switch {
	case f.ProjectsLoaded && f.Projects == 0:
		return Next{Text: words.Say("dashboard.next.empty", "add", press("file.add"), "compose", press("file.compose")), CommandID: "file.add"}
	case f.CheckErr != "":
		return Next{Text: words.Say("dashboard.next.checkerr", "retry", press("ws.doctor")), CommandID: "ws.doctor"}
	case !f.Checked:
		return Next{Text: words.Say("dashboard.next.checking")}
	case f.Failed > 0 && f.Fixable > 0:
		return Next{Text: words.Say("dashboard.next.fixable", "failed", words.Count("dashboard.check.fails", f.Failed), "fixable", words.Count("dashboard.fix.ready", f.Fixable)), CommandID: "ws.fix"}
	case f.Failed > 0:
		return Next{Text: words.Say("dashboard.next.failed", "failed", words.Count("dashboard.check.fails", f.Failed), "doctor", press("ws.doctor")), CommandID: "ws.doctor"}
	case f.Last != nil && !f.Last.OK && !f.Last.Cancelled:
		return Next{Text: words.Say("dashboard.next.lastfailed", "run", press("view.run")), CommandID: "view.run"}
	case f.Affected > 0:
		return Next{Text: words.Say("dashboard.next.affected", "affected", words.Count("dashboard.project.affected", f.Affected), "run", press("run.affected")), CommandID: "run.affected"}
	case f.Warnings > 0:
		return Next{Text: words.Say("dashboard.next.warnings", "warnings", words.Count("dashboard.warning", f.Warnings), "doctor", press("ws.doctor")), CommandID: "ws.doctor"}
	case f.DepsChecked && f.Outdated > 0:
		return Next{Text: words.Say("dashboard.next.outdated", "outdated", words.Count("dashboard.outdated", f.Outdated), "deps", press("view.deps")), CommandID: "view.deps"}
	}
	return Next{Text: words.Say("dashboard.next.good", "build", press("run.build")), CommandID: "run.build"}
}
