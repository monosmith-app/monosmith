package dashboard

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/progress"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// maxProblems is how many failing checks the detail line names.
const maxProblems = 2

// Queue lists what needs attention, the most urgent first, and only what
// applies: a workspace without projects, the first doctor run that is still
// going, what is broken, the last run that failed, the work the changes left,
// a release that is ready and the updates. When nothing asks for attention it
// says so and offers a build.
//
// Parameters: f is what the shell knows now.
// Returns: at least one item. The first button of the first item is the one
// Enter runs, and carries the Enter key.
func Queue(f Facts) []Item {
	switch {
	case f.ProjectsLoaded && f.Projects == 0:
		return []Item{emptyItem()}
	case f.CheckErr == "" && !f.Checked:
		return []Item{checkingItem(f)}
	}
	var q []Item
	if f.CheckErr != "" {
		q = append(q, Item{Kind: KindDoctor, Severity: SeverityFailure,
			Title: words.Say("dashboard.checkerr.title"), Detail: f.CheckErr,
			Buttons: []Button{queueButton("dashboard.retry.label", "ws.doctor")}})
	}
	if r := f.Last; r != nil && !r.OK && !r.Cancelled {
		q = append(q, failedRunItem(*r))
	}
	if f.Failed > 0 {
		q = append(q, setupItem(f))
	}
	if f.Warnings > 0 {
		q = append(q, warningsItem(f))
	}
	if f.Affected > 0 {
		q = append(q, Item{Kind: KindChanges, Severity: SeverityWarning,
			Title:   words.Count("dashboard.changes.title", f.Affected),
			Buttons: []Button{queueButton("dashboard.changes.test", "run.affected")}})
	}
	if r := f.Release; r != nil {
		q = append(q, Item{Kind: KindRelease, Severity: SeverityInfo, Title: releaseTitle(*r),
			Buttons: []Button{queueButton("dashboard.release.open", "view.release")}})
	}
	switch {
	case f.DepsChecked && f.Outdated > 0:
		q = append(q, Item{Kind: KindUpdates, Severity: SeverityInfo, Title: words.Count("dashboard.outdated", f.Outdated),
			Buttons: []Button{queueButton("dashboard.updates.review", "view.deps")}})
	case !f.DepsChecked:
		q = append(q, Item{Kind: KindUpdates, Severity: SeverityMuted, Title: words.Say("dashboard.updates.unlooked"),
			Buttons: []Button{queueButton("dashboard.updates.review", "view.deps")}})
	}
	if !asksForAttention(q) {
		good := Item{Kind: KindGood, Severity: SeveritySuccess, Title: words.Say("dashboard.good.title"),
			Buttons: []Button{queueButton("dashboard.good.build", "run.build-all")}}
		if r := f.Last; r != nil && r.OK && r.Took != "" {
			good.Detail = words.Say("dashboard.good.last", "command", shortCommand(r.Command), "took", r.Took)
		}
		q = append([]Item{good}, q...)
	}
	q[0].Buttons[0].Key = commands.ShortcutOf("home.next")
	return q
}

// Primary is the id of the command Enter runs on Home: the first button of the
// first item. It is "" while nothing can be pressed yet.
func Primary(q []Item) string {
	if len(q) == 0 || len(q[0].Buttons) == 0 {
		return ""
	}
	return q[0].Buttons[0].CommandID
}

// asksForAttention is true when an item of q is more than something that waits.
func asksForAttention(q []Item) bool {
	for _, it := range q {
		if it.Severity != SeverityMuted {
			return true
		}
	}
	return false
}

// queueButton is a button for a command: its label comes from the catalogue and its
// key is the shortcut of the command.
func queueButton(labelKey, id string) Button {
	return Button{Label: words.Say(labelKey), CommandID: id, Key: commands.ShortcutOf(id)}
}

// emptyItem is the whole queue of a workspace without projects. Its buttons
// keep their own keys: Enter runs the first one all the same.
func emptyItem() Item {
	return Item{Kind: KindEmpty, Severity: SeverityInfo,
		Title: words.Say("dashboard.empty.title"), Detail: words.Say("dashboard.empty.detail"),
		Buttons: []Button{queueButton("dashboard.add.label", "file.add"), queueButton("dashboard.compose.label", "file.compose")}}
}

// checkingItem is the whole queue while the first doctor run goes on.
func checkingItem(f Facts) Item {
	return Item{Kind: KindChecking, Severity: SeverityMuted,
		Title:  progress.Waiting(words.Say("dashboard.headline.waiting", "projects", progress.OfProjects(f.Projects)), f.Waited),
		Detail: words.Say("dashboard.checking.detail")}
}

// setupItem says that checks fail, names the first two and offers the fixes.
func setupItem(f Facts) Item {
	it := Item{Kind: KindSetup, Severity: SeverityFailure, Title: words.Count("dashboard.check.fails", f.Failed)}
	if f.Fixable > 0 {
		it.Title = words.Say("dashboard.setup.title", "failed", it.Title, "fixable", words.Count("dashboard.setup.fixable", f.Fixable))
		it.Buttons = append(it.Buttons, queueButton("dashboard.fix.label", "ws.fix"))
	}
	it.Buttons = append(it.Buttons, queueButton("dashboard.doctor.label", "view.doctor"))
	it.Detail = strings.Join(f.Problems[:min(len(f.Problems), maxProblems)], " · ")
	return it
}

// warningsItem says how many checks warn and names the first one. The fix
// button comes first when doctor can fix something, as in setupItem.
func warningsItem(f Facts) Item {
	first := ""
	if len(f.WarningNotes) > 0 {
		first = f.WarningNotes[0]
	}
	it := Item{Kind: KindWarnings, Severity: SeverityWarning, Title: words.Count("dashboard.warn.title", f.Warnings, "first", first)}
	if f.Fixable > 0 {
		it.Buttons = append(it.Buttons, queueButton("dashboard.fix.label", "ws.fix"))
	}
	it.Buttons = append(it.Buttons, queueButton("dashboard.doctor.label", "view.doctor"))
	return it
}

// failedRunItem names the task (or, without one, the command) that failed and
// offers to read its log or repeat the run.
func failedRunItem(r Run) Item {
	what := r.Task
	if what == "" {
		what = shortCommand(r.Command)
	}
	it := Item{Kind: KindRun, Severity: SeverityFailure, Title: words.Say("dashboard.run.failed", "what", what),
		Buttons: []Button{{Label: words.Say("dashboard.run.open"), CommandID: "x:why-failed"}}}
	if r.Repeatable {
		// On Home the key of "run again" reloads the page, so the button shows none.
		it.Buttons = append(it.Buttons, Button{Label: words.Say("dashboard.run.again"), CommandID: "run.again"})
	}
	return it
}

// releaseTitle says how many packages leave and names one.
func releaseTitle(r Ready) string {
	bump := words.Say("dashboard.release.bump", "name", r.Name, "from", r.From, "to", r.To)
	if r.From == "" {
		bump = words.Say("dashboard.release.first", "name", r.Name, "to", r.To)
	}
	packages := words.Count("dashboard.release.packages", r.Packages)
	if r.Commits == 0 {
		return words.Say("dashboard.release.title", "packages", packages, "bump", bump)
	}
	return words.Say("dashboard.release.title_commits", "packages", packages, "bump", bump, "commits", words.Count("dashboard.release.commits", r.Commits))
}

// shortCommand is a command line without the leading "run", which every task
// run has.
func shortCommand(command string) string {
	return strings.TrimPrefix(command, "run ") // words:literal
}
