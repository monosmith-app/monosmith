package dashboard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Lines draws the page in w cells: the headline, the next action, then the
// facts. actionRow is the first line of the next action and commandID the
// command a click on it runs ("" when there is none).
func Lines(th theme.Theme, f Facts, w int) (lines []string, actionRow int, commandID string) {
	next := NextAction(f)
	lines = append(lines, "", " "+headline(th, f), "")
	actionRow = len(lines)
	arrow := th.Accent(th.Icons.Pointer)
	if next.CommandID == "" {
		arrow = th.Muted(th.Icons.Queued)
	}
	lines = append(lines, pane.Wrap(" "+arrow+" "+th.Bold("Next: ")+next.Text, w, 4)...)
	lines = append(lines, "")

	row := func(name, text string) {
		lines = append(lines, pane.Wrap(" "+th.Muted(pane.Cell(name, 10))+text, w, 11)...)
	}
	row("Health", healthText(th, f))
	row("Affected", affectedText(th, f))
	row("Last run", lastRunText(th, f))
	row("Updates", updatesText(th, f))
	if f.ProjectsLoaded && f.Projects > 0 {
		row("Projects", projectsText(f))
	}
	if f.SettingsTip {
		row("Settings", th.Accent("New here?")+" Plugins, the colour theme and the settings of smith.toml are in the Settings tab: press "+commands.ShortcutOf("view.settings"))
	}
	lines = append(lines, whereToGo(th, w)...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, actionRow, next.CommandID
}

// headline is the whole workspace's state in a few words.
func headline(th theme.Theme, f Facts) string {
	switch {
	case f.CheckErr != "":
		return th.Failure(th.Icons.Failed + " doctor could not run")
	case !f.Checked:
		return th.Muted(progress.Waiting("Running the health checks on every project…", f.Waited))
	case f.Failed > 0:
		return th.Failure(th.Icons.Failed+" "+count(f.Failed, "check fails", "checks fail")) + warnTail(th, f)
	case f.Warnings > 0:
		return th.Warning(th.Icons.Warning + " " + count(f.Warnings, "warning", "warnings"))
	}
	return th.Success(th.Icons.Passed + " The workspace is healthy")
}

func warnTail(th theme.Theme, f Facts) string {
	if f.Warnings == 0 {
		return ""
	}
	return th.Muted(" · ") + th.Warning(count(f.Warnings, "warning", "warnings"))
}

func healthText(th theme.Theme, f Facts) string {
	switch {
	case f.CheckErr != "":
		return th.Failure(f.CheckErr)
	case !f.Checked:
		return th.Muted("running doctor in the background")
	}
	parts := []string{th.Success(fmt.Sprintf("%d passed", f.Passed))}
	if f.Failed > 0 {
		parts = append(parts, th.Failure(fmt.Sprintf("%d failed", f.Failed)))
	}
	if f.Warnings > 0 {
		parts = append(parts, th.Warning(count(f.Warnings, "warning", "warnings")))
	}
	return strings.Join(parts, th.Muted(" · ")) + th.Muted("   ("+commands.ShortcutOf("ws.doctor")+" opens Doctor)")
}

func affectedText(th theme.Theme, f Facts) string {
	switch {
	case !f.ProjectsLoaded:
		return th.Muted("reading the workspace…")
	case f.Affected == 0:
		return "nothing changed, of " + count(f.Projects, "project", "projects")
	}
	return th.Warning(fmt.Sprintf("%d", f.Affected)) + " of " + count(f.Projects, "project", "projects") + " changed"
}

func lastRunText(th theme.Theme, f Facts) string {
	r := f.Last
	switch {
	case r == nil:
		return th.Muted("nothing has run yet")
	case r.Cancelled:
		return th.Warning(th.Icons.Skipped+" stopped") + "  " + r.Command
	case r.OK:
		return th.Success(th.Icons.Passed+" passed") + "  " + r.Command + th.Muted("  "+r.Took)
	}
	return th.Failure(th.Icons.Failed+" failed") + "  " + r.Command + th.Muted("  "+r.Took)
}

func updatesText(th theme.Theme, f Facts) string {
	switch {
	case !f.DepsChecked:
		return th.Muted("not checked: " + commands.ShortcutOf("view.deps") + " opens Deps")
	case f.Outdated == 0:
		return "every dependency is on its latest release"
	}
	return th.Warning(fmt.Sprintf("%d", f.Outdated)) + " with newer releases"
}
