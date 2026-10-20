package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// depRow is a dependency with a newer release.
type depRow struct {
	Name     string
	Plugin   string
	Current  string
	Latest   string
	Level    string
	Projects []string
}

// policyLine is one result of the one-version policy.
type policyLine struct {
	Check   string
	OK      bool
	Warning bool
	Detail  string
	Remedy  string
}

type depsLoaded struct {
	rows []depRow
	// policy is empty when smith.toml does not ask for the one-version policy.
	policy []policyLine
	err    error
}

// depsView lists the dependencies that have newer releases, patch before
// minor before major, and the one-version policy's verdict. Space marks a row;
// updating ends as `smith deps up <packages>`, aligning as `smith deps up --align`.
type depsView struct {
	loaded bool
	rows   []depRow
	policy []policyLine
	err    error
	cursor int
	marked map[string]bool
}

func (v depsView) update(key tea.KeyMsg, page int) depsView {
	switch key.String() {
	case " ", "space":
		return v.toggle()
	case "a":
		return v.markAll()
	}
	v.cursor = moveCursor(v.cursor, len(v.rows), key.String(), page)
	return v
}

// toggle marks or unmarks the row under the cursor.
func (v depsView) toggle() depsView {
	if v.cursor < 0 || v.cursor >= len(v.rows) {
		return v
	}
	marked := map[string]bool{}
	for k := range v.marked {
		marked[k] = true
	}
	name := v.rows[v.cursor].Name
	if marked[name] {
		delete(marked, name)
	} else {
		marked[name] = true
	}
	v.marked = marked
	return v
}

// markAll marks every row, or none when every row is already marked.
func (v depsView) markAll() depsView {
	marked := map[string]bool{}
	if len(v.marked) < len(v.rows) {
		for _, r := range v.rows {
			marked[r.Name] = true
		}
	}
	v.marked = marked
	return v
}

// selectRow moves the cursor to row i and marks it when it was already there.
func (v depsView) selectRow(i int) depsView {
	if i < 0 || i >= len(v.rows) {
		return v
	}
	if v.cursor == i {
		return v.toggle()
	}
	v.cursor = i
	return v
}

// chosen are the packages an update would move: the marked rows, or the row
// under the cursor when none is marked.
func (v depsView) chosen() []string {
	var names []string
	for _, r := range v.rows {
		if v.marked[r.Name] {
			names = append(names, r.Name)
		}
	}
	if len(names) == 0 && v.cursor >= 0 && v.cursor < len(v.rows) {
		names = append(names, v.rows[v.cursor].Name)
	}
	return names
}

// upArgv is the command line that updates the chosen packages; nil when none.
func (v depsView) upArgv() []string {
	names := v.chosen()
	if len(names) == 0 {
		return nil
	}
	return append([]string{"deps", "up"}, names...)
}

// alignArgv is the command line that brings the policy's drifting dependencies
// to one version; nil unless the policy found some.
func (v depsView) alignArgv() []string {
	if v.drifting() {
		return []string{"deps", "up", "--align"}
	}
	return nil
}

// drifting is true when the one-version policy found dependencies at different versions.
func (v depsView) drifting() bool {
	return slices.ContainsFunc(v.policy, func(p policyLine) bool { return !p.OK && !p.Warning })
}

// levelCounts counts the rows of each level.
func (v depsView) levelCounts() (patch, minor, major int) {
	for _, r := range v.rows {
		switch r.Level {
		case "patch":
			patch++
		case "minor":
			minor++
		default:
			major++
		}
	}
	return
}

// view draws the view in a w by h area; rowOf maps a line to the row drawn on it.
func (v depsView) view(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.err != nil:
		return []string{th.Failure(v.err.Error())}, rowOf
	case !v.loaded:
		return []string{th.Muted("Looking for newer releases…")}, rowOf
	}
	var tail []string
	tail = append(tail, "", th.Title("One version"))
	if len(v.policy) == 0 {
		tail = append(tail, "  "+th.Muted("not required: set [dependencies] one_version = true in smith.toml"))
	}
	for _, p := range v.policy {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !p.OK && p.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !p.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		tail = append(tail, "  "+mark+" "+p.Check)
		if !p.OK && p.Detail != "" {
			for _, d := range strings.Split(p.Detail, "; ") {
				tail = append(tail, "      "+th.Muted(d))
			}
		}
	}
	tail = append(tail, "", th.Muted("space marks · a marks all · enter or u updates (smith deps up <packages>) · r re-checks"))

	if len(v.rows) == 0 {
		lines = append(lines, th.Success(th.Icons.Passed+" Every dependency is on its latest release."))
	} else {
		patch, minor, major := v.levelCounts()
		lines = append(lines, th.Muted(fmt.Sprintf("%d with newer releases: %d patch · %d minor · %d major   %d marked", len(v.rows), patch, minor, major, len(v.marked))), "")
		nameW := 8
		for _, r := range v.rows {
			nameW = max(nameW, pane.Width(r.Name))
		}
		nameW = min(nameW, max(w/3, 8))
		lines = append(lines, th.Muted(pane.Fit("      "+pad("Package", nameW)+" "+pad("Now", 10)+pad("Latest", 10)+pad("Level", 7)+"Projects", w)))
		avail := max(h-len(lines)-len(tail), 3)
		top := windowTop(v.cursor, len(v.rows), avail)
		for i := top; i < len(v.rows) && i < top+avail; i++ {
			r := v.rows[i]
			box := "[ ] "
			if v.marked[r.Name] {
				box = "[x] "
			}
			level := th.Muted(r.Level)
			switch r.Level {
			case "major":
				level = th.Warning(r.Level)
			case "patch":
				level = th.Success(r.Level)
			}
			plain := "  " + box + pad(r.Name, nameW) + " " + pad(r.Current, 10) + pad(r.Latest, 10) + pad(r.Level, 7) + strings.Join(r.Projects, ", ")
			line := "  " + box + pad(r.Name, nameW) + " " + pad(r.Current, 10) + pad(r.Latest, 10) + pad(level, 7) + strings.Join(r.Projects, ", ")
			if i == v.cursor {
				line = th.Selected(pane.Fit(plain, w))
			}
			rowOf[len(lines)] = i
			lines = append(lines, line)
		}
	}
	lines = append(lines, tail...)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}
