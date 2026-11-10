package doctorview

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
)

// Model lists doctor's failing checks under a heading per group (the
// workspace's first, then each project), failures before warnings; enter on a
// heading folds or unfolds its group, E all of them, g turns the groups off,
// x, w and p show or hide the failed checks, the warnings and the passed ones;
// enter on a check expands it into its problem and remedy, and f opens the wizard of its fix
// (see package fixwizard); Plan, FixFor and FixAll say which fixes there are.
type Model struct {
	Report   *doctor.Report
	Projects []ProjectRef
	Err      error
	// Cursor is a position in the order the findings are drawn.
	Cursor int
	Open   bool
	// Grouped files the findings under headings, the workspace's first and then
	// each project's; a model that is not grouped is one flat list.
	Grouped bool
	// Folds is what the person chose for single groups, by group (the project's
	// name; "" is the workspace): true for open, false for closed.
	Folds map[string]bool
	// All is what the person chose for every group at once: above 0 opens
	// them, below 0 closes them, 0 leaves each as it is by default.
	All int
	// Loading is true while doctor runs: for the first time, or again after a
	// fix or a click on Check again (the list shown stays until the new one arrives).
	Loading bool
	// Waited is how long that run has taken so far, set by the shell for drawing.
	Waited time.Duration
	// Covered is how many projects the run covers, set by the shell for drawing (0 when unknown).
	Covered int
	// ShowPassed lists the checks that passed; by default they are one
	// collapsed line after the problems.
	ShowPassed bool
	// HideFailed and HideWarnings hide the checks of that state in every group;
	// both are off by default, so every problem is listed (see Filter).
	HideFailed, HideWarnings bool
	// Under and Pressed are the filter button under the pointer and the one held
	// down, set by the shell for drawing.
	Under, Pressed Filter
	// UnderReload and PressedReload are the same for the button that checks again.
	UnderReload, PressedReload bool
	// Extras are the checks doctor does not run itself, listed after the
	// findings as rows that run them now; set by the shell.
	Extras []Extra
}

// FromResult is the model for a finished check.
func FromResult(r Result) Model {
	if r.Err != nil {
		return Model{Err: r.Err}
	}
	m := Model{Report: &r.Report, Projects: r.Projects, Grouped: true}
	m.Cursor = m.firstFinding()
	return m
}

// groups is the findings as drawn: one group, or one per project.
func (v Model) groups() []Group {
	if v.Report == nil {
		return nil
	}
	var groups []Group
	if !v.Grouped {
		idx := make([]int, len(v.Report.Findings))
		for i := range idx {
			idx[i] = i
		}
		groups = []Group{{Indexes: idx}}
	} else {
		groups = GroupByProject(v.Report.Findings, v.Projects)
	}
	for _, g := range groups {
		failuresFirst(v.Report.Findings, g.Indexes)
	}
	// Only the checks the filters let through are drawn; a group with none is
	// left out, and a heading still counts all of its checks.
	var shown []Group
	for _, g := range groups {
		var keep []int
		for _, i := range g.Indexes {
			if v.visible(v.Report.Findings[i]) {
				keep = append(keep, i)
			}
		}
		if len(keep) > 0 {
			shown = append(shown, Group{Project: g.Project, Indexes: keep, Counted: g.Indexes})
		}
	}
	return shown
}

// collapsed is true while passed checks are hidden behind one line.
func (v Model) collapsed() bool {
	return v.Report != nil && !v.ShowPassed && v.Report.Passed > 0
}

// rows is how many rows the cursor moves over: the headings, the problems
// of the open groups, then the line that stands for the collapsed passes.
func (v Model) rows() int { return len(v.layout()) }

// OnPassedLine is true when the cursor is on the collapsed passes' line.
func (v Model) OnPassedLine() bool {
	rows := v.layout()
	return v.Cursor >= 0 && v.Cursor < len(rows) && rows[v.Cursor].kind == passedRow
}

// Selected is the finding under the cursor.
func (v Model) Selected() (doctor.Finding, bool) {
	rows := v.layout()
	if v.Report == nil || v.Cursor < 0 || v.Cursor >= len(rows) || rows[v.Cursor].kind != findingRow {
		return doctor.Finding{}, false
	}
	return v.Report.Findings[rows[v.Cursor].finding], true
}

// SelectedExtra is the check doctor does not run that is under the cursor.
func (v Model) SelectedExtra() (Extra, bool) {
	rows := v.layout()
	if v.Cursor < 0 || v.Cursor >= len(rows) || rows[v.Cursor].kind != extraRow {
		return Extra{}, false
	}
	return v.Extras[rows[v.Cursor].extra], true
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	if v.Report == nil || len(v.Report.Findings)+len(v.Extras) == 0 {
		return v
	}
	if f := filterOfKey(key.String()); f != FilterNone {
		// Before the empty check: with every check filtered out, the key is the way back.
		return v.Toggle(f)
	}
	n := v.rows()
	if n == 0 {
		return v
	}
	switch commands.NavNamed(key.String()) {
	case commands.NavUp:
		v.Cursor = (v.Cursor - 1 + n) % n
		v.Open = false
	case commands.NavDown:
		v.Cursor = (v.Cursor + 1) % n
		v.Open = false
	default:
		return v.keyed(key.String())
	}
	return v
}

// keyed handles the keys that are not movement: fold, expand and the toggles.
func (v Model) keyed(name string) Model {
	switch name {
	case "enter", " ":
		switch {
		case v.OnPassedLine():
			v.ShowPassed = true
			return v
		case v.OnHeader():
			return v.toggleGroup(v.layout()[v.Cursor].group)
		}
		v.Open = !v.Open
	case "right", "left":
		if v.OnHeader() {
			g := v.groups()[v.layout()[v.Cursor].group]
			if v.groupOpen(g) != (name == "right") {
				return v.toggleGroup(v.layout()[v.Cursor].group)
			}
		}
	case "E": // words:literal
		v = v.ToggleAll()
		v.Cursor, v.Open = 0, false
	case "g":
		v.Grouped, v.Open = !v.Grouped, false
		v.Cursor = v.firstFinding()
	}
	return v
}

// SelectRow moves the cursor to the row a click landed on, given the code View
// drew on that line: a finding's index (a second click toggles its details),
// PassedRow, ToggleAllRow or a group's HeaderCode (a click folds or unfolds it).
func (v Model) SelectRow(code int) Model {
	switch {
	case code == PassedRow:
		if v.collapsed() {
			// The line of collapsed passes: a click opens it.
			v.ShowPassed = true
		}
		return v
	case code == ToggleAllRow:
		v = v.ToggleAll()
		v.Cursor, v.Open = 0, false
		return v
	case v.Report == nil:
		return v
	}
	for at, r := range v.layout() {
		if r.kind == extraRow && code == ExtraCode(r.extra) {
			v.Cursor, v.Open = at, false
			return v
		}
		isHeader := r.kind == headerRow && code == HeaderCode(r.group)
		isFinding := r.kind == findingRow && code >= 0 && code == r.finding
		switch {
		case isHeader:
			v.Cursor, v.Open = at, false
			return v.toggleGroup(r.group)
		case isFinding && v.Cursor == at:
			v.Open = !v.Open
			return v
		case isFinding:
			v.Cursor, v.Open = at, false
			return v
		}
	}
	return v
}
