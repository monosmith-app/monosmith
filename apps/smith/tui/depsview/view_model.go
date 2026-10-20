package depsview

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model lists the dependencies that have newer releases, patch before
// minor before major, and the one-version policy's verdict. Space marks a row;
// updating ends as `smith deps up <packages>`, aligning as `smith deps up --align`.
type Model struct {
	Loaded bool
	Rows   []Row
	Policy []PolicyLine
	Err    error
	Cursor int
	Marked map[string]bool
}

// FromResult is the model for a loaded result.
func FromResult(r Result) Model {
	return Model{Loaded: true, Rows: r.Rows, Policy: r.Policy, Err: r.Err}
}

// Update handles a key; page is how many rows a page key moves.
func (v Model) Update(key tea.KeyMsg, page int) Model {
	switch key.String() {
	case " ", "space":
		return v.toggle()
	case "a":
		return v.markAll()
	}
	v.Cursor = pane.MoveCursor(v.Cursor, len(v.Rows), key.String(), page)
	return v
}

// toggle marks or unmarks the row under the cursor.
func (v Model) toggle() Model {
	if v.Cursor < 0 || v.Cursor >= len(v.Rows) {
		return v
	}
	marked := map[string]bool{}
	for k := range v.Marked {
		marked[k] = true
	}
	name := v.Rows[v.Cursor].Name
	if marked[name] {
		delete(marked, name)
	} else {
		marked[name] = true
	}
	v.Marked = marked
	return v
}

// markAll marks every row, or none when every row is already marked.
func (v Model) markAll() Model {
	marked := map[string]bool{}
	if len(v.Marked) < len(v.Rows) {
		for _, r := range v.Rows {
			marked[r.Name] = true
		}
	}
	v.Marked = marked
	return v
}

// SelectRow moves the cursor to row i and marks it when it was already there.
func (v Model) SelectRow(i int) Model {
	if i < 0 || i >= len(v.Rows) {
		return v
	}
	if v.Cursor == i {
		return v.toggle()
	}
	v.Cursor = i
	return v
}

// chosen are the packages an update would move: the marked rows, or the row
// under the cursor when none is marked.
func (v Model) chosen() []string {
	var names []string
	for _, r := range v.Rows {
		if v.Marked[r.Name] {
			names = append(names, r.Name)
		}
	}
	if len(names) == 0 && v.Cursor >= 0 && v.Cursor < len(v.Rows) {
		names = append(names, v.Rows[v.Cursor].Name)
	}
	return names
}

// UpArgv is the command line that updates the chosen packages; nil when none.
func (v Model) UpArgv() []string {
	names := v.chosen()
	if len(names) == 0 {
		return nil
	}
	return append([]string{"deps", "up"}, names...)
}

// AlignArgv is the command line that brings the policy's drifting dependencies
// to one version; nil unless the policy found some.
func (v Model) AlignArgv() []string {
	if v.drifting() {
		return []string{"deps", "up", "--align"}
	}
	return nil
}

// drifting is true when the one-version policy found dependencies at different versions.
func (v Model) drifting() bool {
	return slices.ContainsFunc(v.Policy, func(p PolicyLine) bool { return !p.OK && !p.Warning })
}

// levelCounts counts the rows of each level.
func (v Model) levelCounts() (patch, minor, major int) {
	for _, r := range v.Rows {
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

// View draws the view in a w by h area; rowOf maps a line to the row drawn on it.
func (v Model) View(th theme.Theme, w, h int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	switch {
	case v.Err != nil:
		return []string{th.Failure(v.Err.Error())}, rowOf
	case !v.Loaded:
		return []string{th.Muted("Looking for newer releases…")}, rowOf
	}
	var tail []string
	tail = append(tail, "", th.Title("One version"))
	if len(v.Policy) == 0 {
		tail = append(tail, "  "+th.Muted("not required: set [dependencies] one_version = true in smith.toml"))
	}
	for _, p := range v.Policy {
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

	if len(v.Rows) == 0 {
		lines = append(lines, th.Success(th.Icons.Passed+" Every dependency is on its latest release."))
	} else {
		patch, minor, major := v.levelCounts()
		lines = append(lines, th.Muted(fmt.Sprintf("%d with newer releases: %d patch · %d minor · %d major   %d marked", len(v.Rows), patch, minor, major, len(v.Marked))), "")
		nameW := 8
		for _, r := range v.Rows {
			nameW = max(nameW, pane.Width(r.Name))
		}
		nameW = min(nameW, max(w/3, 8))
		lines = append(lines, th.Muted(pane.Fit("      "+pane.Fit("Package", nameW)+" "+pane.Fit("Now", 10)+pane.Fit("Latest", 10)+pane.Fit("Level", 7)+"Projects", w)))
		avail := max(h-len(lines)-len(tail), 3)
		top := pane.WindowTop(v.Cursor, len(v.Rows), avail)
		for i := top; i < len(v.Rows) && i < top+avail; i++ {
			r := v.Rows[i]
			box := "[ ] "
			if v.Marked[r.Name] {
				box = "[x] "
			}
			level := th.Muted(r.Level)
			switch r.Level {
			case "major":
				level = th.Warning(r.Level)
			case "patch":
				level = th.Success(r.Level)
			}
			plain := "  " + box + pane.Fit(r.Name, nameW) + " " + pane.Fit(r.Current, 10) + pane.Fit(r.Latest, 10) + pane.Fit(r.Level, 7) + strings.Join(r.Projects, ", ")
			line := "  " + box + pane.Fit(r.Name, nameW) + " " + pane.Fit(r.Current, 10) + pane.Fit(r.Latest, 10) + pane.Fit(level, 7) + strings.Join(r.Projects, ", ")
			if i == v.Cursor {
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
