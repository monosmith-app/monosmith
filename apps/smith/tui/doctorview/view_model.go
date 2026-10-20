package doctorview

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model lists doctor's findings, problems first; enter opens the
// selected one's detail and remedy.
type Model struct {
	Report *doctor.Report
	Err    error
	Cursor int
	Open   bool
}

// Result is what running doctor gave: the shell receives it as a message.
type Result struct {
	Report doctor.Report
	Err    error
}

// FromResult is the model for a finished check.
func FromResult(r Result) Model {
	if r.Err != nil {
		return Model{Err: r.Err}
	}
	return Model{Report: &r.Report}
}

// Update handles a key.
func (v Model) Update(key tea.KeyMsg) Model {
	if v.Report == nil || len(v.Report.Findings) == 0 {
		return v
	}
	switch key.String() {
	case "up":
		v.Cursor = (v.Cursor - 1 + len(v.Report.Findings)) % len(v.Report.Findings)
		v.Open = false
	case "down":
		v.Cursor = (v.Cursor + 1) % len(v.Report.Findings)
		v.Open = false
	case "enter", " ":
		v.Open = !v.Open
	}
	return v
}

// view draws the findings in w cells; rowOf maps a line to the finding drawn on it.
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
	for i, f := range v.Report.Findings {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !f.OK && f.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !f.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		rowOf[len(lines)] = i
		lines = append(lines, guide.Row(th, i == v.Cursor, mark+" ", f.Check, f.Source))
		if i == v.Cursor && v.Open && !f.OK {
			if f.Detail != "" {
				lines = append(lines, "      "+f.Detail)
			}
			if f.Remedy != "" {
				lines = append(lines, "      "+th.Accent("remedy: ")+f.Remedy)
			}
		}
	}
	lines = append(lines, "", th.Muted("up/down move · enter details and remedy · ctrl+d re-check · ctrl+x fix (smith doctor --fix)"))
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, rowOf
}

// selectRow moves the cursor to finding i and toggles its details when it was already there.
func (v Model) SelectRow(i int) Model {
	if v.Report == nil || i < 0 || i >= len(v.Report.Findings) {
		return v
	}
	if v.Cursor == i {
		v.Open = !v.Open
	} else {
		v.Cursor, v.Open = i, false
	}
	return v
}
