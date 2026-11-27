package projectpage

import (
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// RunRow is one run of the Runs section.
type RunRow struct {
	Mark Mark
	// Task names the project's tasks in the run; Result says how they ended;
	// Time is how long the run took and Ago how long ago it ended.
	Task, Result, Time, Ago string
}

// RunList draws the Runs section: one line per run, newest first, with its
// mark, the task, the result, the time and how long ago. The row at the
// cursor has the pointer and is on the Highlight background while the detail
// has the keyboard.
//
// Parameters: th is the theme; rows are the runs; w is the width; cursor is
// the row of the keyboard; focused is true while the detail has the keyboard;
// tick moves the spinner of a run that goes on.
// Returns: the lines, and the row on each line of a run.
func RunList(th theme.Theme, rows []RunRow, w, cursor int, focused bool, tick int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if len(rows) == 0 {
		return []string{pane.Fit(" "+th.Muted(words.Say("projectpage.runs.none")), w)}, rowOf
	}
	taskW := 8
	for _, r := range rows {
		taskW = max(taskW, pane.Width(r.Task)+2)
	}
	taskW = min(taskW, max(w/3, 8))
	for i, r := range rows {
		chosen := i == cursor && focused
		task, lead := pane.Cell(r.Task, taskW), "   "
		if chosen {
			task, lead = th.Accent(th.Bold(task)), " "+th.Accent(Pointer)
		}
		line := lead + Glyph(th, r.Mark, tick) + " " + task + pane.Cell(r.Result, 12) + pane.Cell(r.Time, 10) + th.Muted(r.Ago)
		line = pane.Fit(line, w)
		if chosen {
			line = th.Selected(line)
		}
		rowOf[len(lines)] = i
		lines = append(lines, line)
	}
	return lines, rowOf
}
