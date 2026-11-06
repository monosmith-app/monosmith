package dashboard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Lines is Page without the buttons: the lines, the row of the next action
// and the command a click on it runs.
func Lines(th theme.Theme, f Facts, w int) (lines []string, actionRow int, commandID string) {
	lines, actionRow, commandID, _ = Page(th, f, w)
	return lines, actionRow, commandID
}

// Page draws the page in w cells: the headline, the next action, then the
// facts and the places to go, with a button each to add a project and to
// compose a stack. actionRow is the first line of the next action and
// commandID the command a click on it runs ("" when there is none); buttons
// maps the line of each button to its command.
func Page(th theme.Theme, f Facts, w int) (lines []string, actionRow int, commandID string, buttons map[int]string) {
	next := NextAction(f)
	lines = append(lines, "", " "+headline(th, f), "")
	actionRow = len(lines)
	arrow := th.Accent(th.Icons.Pointer)
	if next.CommandID == "" {
		arrow = th.Muted(th.Icons.Queued)
	}
	lines = append(lines, pane.Wrap(" "+arrow+" "+th.Bold(words.Say("dashboard.next.label"))+next.Text, w, 4)...)
	lines = append(lines, "")

	row := func(name, text string) {
		lines = append(lines, pane.Wrap(" "+th.Muted(pane.Cell(name, 10))+text, w, 11)...)
	}
	row(words.Say("dashboard.row.health"), healthText(th, f))
	row(words.Say("dashboard.row.affected"), affectedText(th, f))
	row(words.Say("dashboard.row.last"), lastRunText(th, f))
	row(words.Say("dashboard.row.updates"), updatesText(th, f))
	if f.ProjectsLoaded && f.Projects > 0 {
		row(words.Say("dashboard.row.projects"), projectsText(f))
	}
	if f.SettingsTip {
		row(words.Say("dashboard.row.settings"), words.Say("dashboard.settings.tip", "new", th.Accent(words.Say("dashboard.settings.new")), "shortcut", commands.ShortcutOf("view.settings")))
	}
	where, at := whereToGo(th, w)
	buttons = map[int]string{}
	for i, id := range at {
		buttons[len(lines)+i] = id
	}
	lines = append(lines, where...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, actionRow, next.CommandID, buttons
}

// headline is the whole workspace's state in a few words.
func headline(th theme.Theme, f Facts) string {
	switch {
	case f.CheckErr != "":
		return th.Failure(th.Icons.Failed + " " + words.Say("dashboard.headline.checkerr"))
	case !f.Checked:
		return th.Muted(progress.Waiting(words.Say("dashboard.headline.waiting", "projects", progress.OfProjects(f.Projects)), f.Waited))
	case f.Failed > 0:
		return th.Failure(th.Icons.Failed+" "+words.Count("dashboard.check.fails", f.Failed)) + warnTail(th, f)
	case f.Warnings > 0:
		return th.Warning(th.Icons.Warning + " " + words.Count("dashboard.warning", f.Warnings))
	}
	return th.Success(th.Icons.Passed + " " + words.Say("dashboard.headline.healthy"))
}

func warnTail(th theme.Theme, f Facts) string {
	if f.Warnings == 0 {
		return ""
	}
	return th.Muted(" · ") + th.Warning(words.Count("dashboard.warning", f.Warnings))
}

func healthText(th theme.Theme, f Facts) string {
	switch {
	case f.CheckErr != "":
		return th.Failure(f.CheckErr)
	case !f.Checked:
		return th.Muted(words.Say("dashboard.health.running"))
	}
	parts := []string{th.Success(words.Say("dashboard.health.passed", "count", f.Passed))}
	if f.Failed > 0 {
		parts = append(parts, th.Failure(words.Say("dashboard.health.failed", "count", f.Failed)))
	}
	if f.Warnings > 0 {
		parts = append(parts, th.Warning(words.Count("dashboard.warning", f.Warnings)))
	}
	return strings.Join(parts, th.Muted(" · ")) + th.Muted(words.Say("dashboard.health.hint", "shortcut", commands.ShortcutOf("ws.doctor")))
}

func affectedText(th theme.Theme, f Facts) string {
	switch {
	case !f.ProjectsLoaded:
		return th.Muted(words.Say("dashboard.affected.reading"))
	case f.Affected == 0:
		return words.Say("dashboard.affected.none", "projects", words.Count("dashboard.project", f.Projects))
	}
	return words.Say("dashboard.affected.some", "affected", th.Warning(fmt.Sprintf("%d", f.Affected)), "projects", words.Count("dashboard.project", f.Projects))
}

func lastRunText(th theme.Theme, f Facts) string {
	r := f.Last
	switch {
	case r == nil:
		return th.Muted(words.Say("dashboard.last.none"))
	case r.Cancelled:
		return th.Warning(th.Icons.Skipped+" "+words.Say("dashboard.last.stopped")) + "  " + r.Command
	case r.OK:
		return th.Success(th.Icons.Passed+" "+words.Say("dashboard.last.passed")) + "  " + r.Command + th.Muted("  "+r.Took)
	}
	return th.Failure(th.Icons.Failed+" "+words.Say("dashboard.last.failed")) + "  " + r.Command + th.Muted("  "+r.Took)
}

func updatesText(th theme.Theme, f Facts) string {
	switch {
	case !f.DepsChecked:
		return th.Muted(words.Say("dashboard.updates.unchecked", "shortcut", commands.ShortcutOf("view.deps")))
	case f.Outdated == 0:
		return words.Say("dashboard.updates.current")
	}
	return words.Say("dashboard.updates.outdated", "count", th.Warning(fmt.Sprintf("%d", f.Outdated)))
}
