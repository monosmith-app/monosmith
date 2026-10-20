package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// doctorView lists doctor's findings, problems first; enter opens the
// selected one's detail and remedy.
type doctorView struct {
	report *doctor.Report
	err    error
	cursor int
	open   bool
}

type doctorLoaded struct {
	report doctor.Report
}

func (v doctorView) update(key tea.KeyMsg) doctorView {
	if v.report == nil || len(v.report.Findings) == 0 {
		return v
	}
	switch key.String() {
	case "up":
		v.cursor = (v.cursor - 1 + len(v.report.Findings)) % len(v.report.Findings)
		v.open = false
	case "down":
		v.cursor = (v.cursor + 1) % len(v.report.Findings)
		v.open = false
	case "enter", " ":
		v.open = !v.open
	}
	return v
}

// view draws the findings in w cells; rowOf maps a line to the finding drawn on it.
func (v doctorView) view(th theme.Theme, w int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if v.err != nil {
		return []string{th.Failure(v.err.Error())}, rowOf
	}
	if v.report == nil {
		return []string{th.Muted("Checking the workspace…")}, rowOf
	}
	lines = append(lines, fmt.Sprintf("%s · %s · %s", th.Success(fmt.Sprintf("%d passed", v.report.Passed)),
		th.Failure(fmt.Sprintf("%d failed", v.report.Failed)), th.Warning(fmt.Sprintf("%d warning(s)", v.report.Warnings))), "")
	for i, f := range v.report.Findings {
		mark := th.Success(th.Icons.Passed)
		switch {
		case !f.OK && f.Warning:
			mark = th.Warning(th.Icons.Warning)
		case !f.OK:
			mark = th.Failure(th.Icons.Failed)
		}
		rowOf[len(lines)] = i
		lines = append(lines, guide.Row(th, i == v.cursor, mark+" ", f.Check, f.Source))
		if i == v.cursor && v.open && !f.OK {
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
func (v doctorView) selectRow(i int) doctorView {
	if v.report == nil || i < 0 || i >= len(v.report.Findings) {
		return v
	}
	if v.cursor == i {
		v.open = !v.open
	} else {
		v.cursor, v.open = i, false
	}
	return v
}
