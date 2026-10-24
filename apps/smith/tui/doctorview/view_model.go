package doctorview

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model lists doctor's failing checks, one line each, failures before
// warnings; g files them by project (the workspace's first); enter expands the
// selected one into its problem and remedy, and f applies its fix when it is a safe one, as
// `smith doctor --fix --only <check>`.
type Model struct {
	Report   *doctor.Report
	Projects []ProjectRef
	Err      error
	// Cursor is a position in the order the findings are drawn.
	Cursor int
	Open   bool
	// Grouped files the findings under their project; by default they are one flat list.
	Grouped bool
	// Loading is true while doctor runs for the first time.
	Loading bool
	// ShowPassed lists the checks that passed; by default they are one
	// collapsed line after the problems.
	ShowPassed bool
}

// FromResult is the model for a finished check.
func FromResult(r Result) Model {
	if r.Err != nil {
		return Model{Err: r.Err}
	}
	return Model{Report: &r.Report, Projects: r.Projects}
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
	if v.ShowPassed {
		return groups
	}
	// Passes are collapsed: only the problems are drawn.
	var shown []Group
	for _, g := range groups {
		var keep []int
		for _, i := range g.Indexes {
			if !v.Report.Findings[i].OK {
				keep = append(keep, i)
			}
		}
		if len(keep) > 0 {
			shown = append(shown, Group{Project: g.Project, Indexes: keep})
		}
	}
	return shown
}

// collapsed is true while passed checks are hidden behind one line.
func (v Model) collapsed() bool {
	return v.Report != nil && !v.ShowPassed && v.Report.Passed > 0
}

// rows is how many rows the cursor moves over: the problems, then the line
// that stands for the collapsed passes.
func (v Model) rows() int {
	n := len(v.order())
	if v.collapsed() {
		n++
	}
	return n
}

// OnPassedLine is true when the cursor is on the collapsed passes' line.
func (v Model) OnPassedLine() bool { return v.collapsed() && v.Cursor == len(v.order()) }

// order is the finding indexes in the order they are drawn.
func (v Model) order() []int {
	var out []int
	for _, g := range v.groups() {
		out = append(out, g.Indexes...)
	}
	return out
}

// Selected is the finding under the cursor.
func (v Model) Selected() (doctor.Finding, bool) {
	order := v.order()
	if v.Report == nil || v.Cursor < 0 || v.Cursor >= len(order) {
		return doctor.Finding{}, false
	}
	return v.Report.Findings[order[v.Cursor]], true
}

// FixArgv is the command line that applies the selected finding's fix; nil
// unless the finding failed and carries a safe fix (others are only shown).
func (v Model) FixArgv() []string {
	f, ok := v.Selected()
	if !ok || f.OK || f.Fix == nil || !f.Fix.Safe {
		return nil
	}
	return []string{"doctor", "--fix", "--only", f.Check}
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	if v.Report == nil || len(v.Report.Findings) == 0 {
		return v
	}
	n := v.rows()
	if n == 0 {
		return v
	}
	switch key.String() {
	case "up":
		v.Cursor = (v.Cursor - 1 + n) % n
		v.Open = false
	case "down":
		v.Cursor = (v.Cursor + 1) % n
		v.Open = false
	case "enter", " ":
		if v.OnPassedLine() {
			v.ShowPassed = true
			return v
		}
		v.Open = !v.Open
	case "p":
		v.ShowPassed, v.Open = !v.ShowPassed, false
		v.Cursor = min(v.Cursor, max(v.rows()-1, 0))
	case "g":
		v.Grouped, v.Cursor, v.Open = !v.Grouped, 0, false
	}
	return v
}

// Headline is the state of the workspace in a few words: "2 failed · 2 warnings".
func (v Model) Headline(th theme.Theme) string {
	r := v.Report
	var parts []string
	if r.Failed > 0 {
		parts = append(parts, th.Failure(fmt.Sprintf("%s %d failed", th.Icons.Failed, r.Failed)))
	}
	if r.Warnings > 0 {
		parts = append(parts, th.Warning(fmt.Sprintf("%s %d warning(s)", th.Icons.Warning, r.Warnings)))
	}
	if len(parts) == 0 {
		parts = append(parts, th.Success(th.Icons.Passed+" Every check passed"))
	}
	return strings.Join(parts, th.Muted(" · "))
}

// View draws the page in w cells: the headline first, then the failing
// checks one line each (the state and the title), the passed ones behind one
// line. With expand the open check shows its problem, remedy and fix under
// its line; without it the detail beside the list says all that, so the page
// does not repeat it. rowOf maps a line to the finding drawn on it.
func (v Model) View(th theme.Theme, w int, expand bool) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.Err != nil {
		return []string{th.Failure(v.Err.Error())}, rowOf
	}
	if v.Report == nil {
		return []string{th.Muted("Checking the workspace…")}, rowOf
	}
	lines = append(lines, " "+v.Headline(th), "")
	position := 0
	for _, g := range v.groups() {
		if v.Grouped {
			lines = append(lines, v.header(th, g))
		}
		for _, i := range g.Indexes {
			f := v.Report.Findings[i]
			selected := position == v.Cursor
			rowOf[len(lines)] = i
			lines = append(lines, guide.Row(th, selected, Mark(th, f)+" ", f.Check, ""))
			if selected && v.Open && expand && !f.OK {
				lines = append(lines, Text(th, f, 5)...)
			}
			position++
		}
	}
	if v.collapsed() {
		if len(lines) > 2 {
			lines = append(lines, "")
		}
		rowOf[len(lines)] = -1
		lines = append(lines, guide.Row(th, v.OnPassedLine(), th.Icons.Closed+" ", fmt.Sprintf("%d checks passed", v.Report.Passed), "enter or p shows them"))
	}
	lines, rowOf = pane.WrapRows(lines, rowOf, w, 5)
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// Mark is the symbol of a finding's state.
func Mark(th theme.Theme, f doctor.Finding) string {
	switch {
	case !f.OK && f.Warning:
		return th.Warning(th.Icons.Warning)
	case !f.OK:
		return th.Failure(th.Icons.Failed)
	}
	return th.Success(th.Icons.Passed)
}

// Text is a check told in full, each part once: the problem, the remedy, the
// fix and the files it touches, indented by indent cells.
func Text(th theme.Theme, f doctor.Finding, indent int) []string {
	pad := strings.Repeat(" ", indent)
	var out []string
	if f.Detail != "" {
		out = append(out, pad+th.Muted(f.Detail))
	}
	if f.Remedy != "" {
		out = append(out, pad+th.Accent("→ ")+f.Remedy)
	}
	if f.OK {
		return out
	}
	switch {
	case f.Fix != nil && f.Fix.Safe:
		out = append(out, pad+th.Accent("fix: ")+f.Fix.Description+th.Muted(" (safe: f applies it)"))
	case f.Fix != nil:
		out = append(out, pad+th.Accent("fix: ")+f.Fix.Description+th.Muted(" (shown by smith doctor --fix, never applied: it could lose something of yours)"))
	}
	if f.Fix != nil {
		for _, e := range f.Fix.Files {
			verb := "writes "
			if e.Delete {
				verb = "removes "
			}
			out = append(out, pad+"  "+th.Muted(verb+e.Path))
		}
	}
	return out
}

// header is a group's title line with what is wrong in it.
func (v Model) header(th theme.Theme, g Group) string {
	name := g.Project
	if name == "" {
		name = "workspace"
	}
	failed, warned := 0, 0
	for _, i := range g.Indexes {
		switch f := v.Report.Findings[i]; {
		case f.OK:
		case f.Warning:
			warned++
		default:
			failed++
		}
	}
	var facts []string
	if failed > 0 {
		facts = append(facts, th.Failure(fmt.Sprintf("%d failed", failed)))
	}
	if warned > 0 {
		facts = append(facts, th.Warning(fmt.Sprintf("%d warning(s)", warned)))
	}
	if len(facts) == 0 {
		facts = append(facts, th.Muted("all passed"))
	}
	return th.Title(name) + "  " + strings.Join(facts, " · ")
}

// SelectRow moves the cursor to finding i and toggles its details when it was already there.
func (v Model) SelectRow(i int) Model {
	if i == -1 && v.collapsed() {
		// The line of collapsed passes: a click opens it.
		v.ShowPassed = true
		return v
	}
	if v.Report == nil || i < 0 || i >= len(v.Report.Findings) {
		return v
	}
	at := -1
	for pos, idx := range v.order() {
		if idx == i {
			at = pos
		}
	}
	if at < 0 {
		return v
	}
	if v.Cursor == at {
		v.Open = !v.Open
	} else {
		v.Cursor, v.Open = at, false
	}
	return v
}
