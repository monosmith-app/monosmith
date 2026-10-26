package releaseview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// maxCommits bounds the commits listed under one bump so a long history does
// not push the other projects off the screen.
const maxCommits = 12

// Model shows what the next release would do, project by project. Enter
// drills down to the commits behind the selected bump; releasing itself runs
// `smith release` in the run view, after a confirmation that shows the command.
type Model struct {
	Loaded   bool
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
	// Cursor is the project selected in the plan, Open is true while its
	// commits are shown.
	Cursor int
	Open   bool
}

// Result is what planning the release found: the shell receives it as a message.
type Result struct {
	Plan     release.Plan
	Blockers []string
	Warnings []string
	Err      error
}

// FromResult is the model for a planned release.
func FromResult(r Result) Model {
	return Model{Loaded: true, Plan: r.Plan, Blockers: r.Blockers, Warnings: r.Warnings, Err: r.Err}
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	n := len(v.Plan.Items)
	if !v.Loaded || v.Err != nil || len(v.Blockers) > 0 || n == 0 {
		return v
	}
	switch key.String() {
	case "up", "k":
		v.Cursor, v.Open = (v.Cursor-1+n)%n, false
	case "down", "j":
		v.Cursor, v.Open = (v.Cursor+1)%n, false
	case "home":
		v.Cursor, v.Open = 0, false
	case "end":
		v.Cursor, v.Open = n-1, false
	case "enter", " ":
		v.Open = !v.Open
	}
	return v
}

// SelectRow moves the cursor to item i and opens or closes its commits when it was already there.
func (v Model) SelectRow(i int) Model {
	if i < 0 || i >= len(v.Plan.Items) {
		return v
	}
	if v.Cursor == i {
		v.Open = !v.Open
	} else {
		v.Cursor, v.Open = i, false
	}
	return v
}

// Selected is the project under the cursor.
func (v Model) Selected() (release.Item, bool) {
	if v.Cursor < 0 || v.Cursor >= len(v.Plan.Items) {
		return release.Item{}, false
	}
	return v.Plan.Items[v.Cursor], true
}

// View draws the plan in a w by h area; rowOf maps a line to the plan item drawn on it.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		lines = append(lines, th.Failure(v.Err.Error()))
	case !v.Loaded:
		lines = append(lines, pane.Wrap(th.Muted("Reading the commits since the last release to work out each version. This takes a few seconds, up to half a minute on a large workspace; you can use the other tabs meanwhile."), w, 0)...)
	case len(v.Blockers) > 0:
		lines = append(lines, th.Failure(th.Icons.Failed+" The release is blocked: its tags would give wrong versions."), "")
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		for _, m := range v.Blockers {
			lines = append(lines, "  • "+m)
		}
		lines = append(lines, "", th.Muted("smith release --dry-run shows the same"))
	default:
		lines = append(lines, v.headline(th), "")
		for _, m := range v.Warnings {
			lines = append(lines, th.Warning("warning: "+m))
		}
		if len(v.Plan.Items) == 0 {
			lines = append(lines, strings.Split(release.NothingToRelease(), "\n")...)
		}
		nameW := 8
		for _, item := range v.Plan.Items {
			nameW = max(nameW, pane.Width(item.Project))
		}
		nameW = min(nameW, max(w/3, 8))
		for i, item := range v.Plan.Items {
			from := item.Current
			if from == "" {
				from = "new"
			}
			selected := i == v.Cursor
			cursor := "  "
			if selected {
				cursor = th.Accent(th.Icons.Pointer + " ")
			}
			rowOf[len(lines)] = i
			lines = append(lines, fmt.Sprintf("%s%s%s → %s  %s  %s", cursor, pane.MiddleCell(item.Project, nameW+1), from, th.Bold(item.Next), levelMark(th, item.Level), th.Muted(item.Reason)))
			if selected && v.Open {
				lines = append(lines, v.commits(th, item)...)
			}
		}
		if len(v.Plan.Unchanged) > 0 {
			lines = append(lines, "", th.Muted("unchanged: "+strings.Join(v.Plan.Unchanged, ", ")))
		}
	}
	// A long line wraps (the unchanged list, the help) instead of being cut.
	lines, rowOf = pane.WrapRows(lines, rowOf, w, 2)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// headline says what the release is in a few words, before the detail.
func (v Model) headline(th theme.Theme) string {
	n := len(v.Plan.Items)
	if n == 0 {
		return th.Muted("Nothing to release")
	}
	levels := map[string]int{}
	for _, item := range v.Plan.Items {
		levels[item.Level]++
	}
	var parts []string
	for _, l := range []string{"major", "minor", "patch"} {
		if levels[l] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", levels[l], l))
		}
	}
	noun := "packages"
	if n == 1 {
		noun = "package"
	}
	return th.Bold(fmt.Sprintf("%d %s to release", n, noun)) + th.Muted(" ("+strings.Join(parts, " · ")+")")
}

// levelMark names a bump's level with a symbol, so it reads without colour.
func levelMark(th theme.Theme, level string) string {
	switch level {
	case "major":
		return th.Failure("major")
	case "minor":
		return th.Warning("minor")
	case "patch":
		return th.Success("patch")
	}
	return th.Muted(level)
}

// commits is the drill-down: the tag and the commits that make the project release.
func (v Model) commits(th theme.Theme, item release.Item) []string {
	out := []string{"      " + th.Muted("tag "+item.Tag+" · "+item.Level+" bump · plugin "+item.Plugin)}
	if len(item.Changes) == 0 {
		return append(out, "      "+th.Muted(item.Reason+" (no commit of its own releases it)"))
	}
	for i, c := range item.Changes {
		if i == maxCommits {
			out = append(out, "      "+th.Muted(fmt.Sprintf("… and %d more (smith release --dry-run --notes lists them all)", len(item.Changes)-maxCommits)))
			break
		}
		kind := c.Type
		if len(c.Scopes) > 0 {
			kind += "(" + strings.Join(c.Scopes, ",") + ")"
		}
		if c.Breaking {
			kind += "!"
		}
		line := "      " + th.Muted(short(c.Hash)) + " " + th.Bold(kind) + ": " + c.Description
		if c.Author != "" {
			line += th.Muted("  " + c.Author)
		}
		out = append(out, line)
		if c.BreakingNote != "" {
			out = append(out, "        "+th.Failure("BREAKING: "+c.BreakingNote))
		}
	}
	return out
}

// short is a commit hash as git abbreviates it.
func short(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// CanRelease reports whether a release can be started from the plan.
func (v Model) CanRelease() bool {
	return v.Loaded && v.Err == nil && len(v.Blockers) == 0 && len(v.Plan.Items) > 0
}
