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

// Model lists doctor's findings grouped by project (the workspace's first),
// problems first in each group; enter opens the selected one's detail and
// remedy, and f applies its fix when it is a safe one, as
// `smith doctor --fix --only <check>`.
type Model struct {
	Report   *doctor.Report
	Projects []ProjectRef
	Err      error
	// Cursor is a position in the order the findings are drawn.
	Cursor int
	Open   bool
	// Flat lists the findings in one run instead of grouped by project.
	Flat bool
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
	if v.Flat {
		idx := make([]int, len(v.Report.Findings))
		for i := range idx {
			idx[i] = i
		}
		return []Group{{Indexes: idx}}
	}
	return GroupByProject(v.Report.Findings, v.Projects)
}

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
	n := len(v.Report.Findings)
	switch key.String() {
	case "up":
		v.Cursor = (v.Cursor - 1 + n) % n
		v.Open = false
	case "down":
		v.Cursor = (v.Cursor + 1) % n
		v.Open = false
	case "enter", " ":
		v.Open = !v.Open
	case "g":
		v.Flat, v.Cursor, v.Open = !v.Flat, 0, false
	}
	return v
}

// View draws the findings in w cells; rowOf maps a line to the finding drawn on it.
func (v Model) View(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.Err != nil {
		return []string{th.Failure(v.Err.Error())}, rowOf
	}
	if v.Report == nil {
		return []string{th.Muted("Checking the workspace…")}, rowOf
	}
	lines = append(lines, fmt.Sprintf("%s · %s · %s", th.Success(fmt.Sprintf("%d passed", v.Report.Passed)),
		th.Failure(fmt.Sprintf("%d failed", v.Report.Failed)), th.Warning(fmt.Sprintf("%d warning(s)", v.Report.Warnings))), "")
	position := 0
	for _, g := range v.groups() {
		if !v.Flat {
			lines = append(lines, v.header(th, g))
		}
		for _, i := range g.Indexes {
			f := v.Report.Findings[i]
			mark := th.Success(th.Icons.Passed)
			switch {
			case !f.OK && f.Warning:
				mark = th.Warning(th.Icons.Warning)
			case !f.OK:
				mark = th.Failure(th.Icons.Failed)
			}
			selected := position == v.Cursor
			rowOf[len(lines)] = i
			lines = append(lines, guide.Row(th, selected, mark+" ", f.Check, f.Source))
			if selected && v.Open && !f.OK {
				lines = append(lines, v.detail(th, f)...)
			}
			position++
		}
	}
	lines = append(lines, "", th.Muted("up/down move · enter details and remedy · g group by project or not"),
		th.Muted("f fix this one (smith doctor --fix --only) · ctrl+x fix all (smith doctor --fix) · ctrl+d re-check"))
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
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

// detail is what the open finding says and what could fix it.
func (v Model) detail(th theme.Theme, f doctor.Finding) []string {
	var out []string
	if f.Detail != "" {
		out = append(out, "      "+f.Detail)
	}
	if f.Remedy != "" {
		out = append(out, "      "+th.Accent("remedy: ")+f.Remedy)
	}
	switch {
	case f.Fix != nil && f.Fix.Safe:
		out = append(out, "      "+th.Accent("fix: ")+f.Fix.Description+th.Muted(" (safe: f applies it)"))
	case f.Fix != nil:
		out = append(out, "      "+th.Accent("fix: ")+f.Fix.Description+th.Muted(" (shown by smith doctor --fix, never applied: it could lose something of yours)"))
	}
	if f.Fix != nil {
		for _, e := range f.Fix.Files {
			verb := "writes "
			if e.Delete {
				verb = "removes "
			}
			out = append(out, "        "+th.Muted(verb+e.Path))
		}
	}
	return out
}

// SelectRow moves the cursor to finding i and toggles its details when it was already there.
func (v Model) SelectRow(i int) Model {
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
