package projectpage

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// TaskRow is one task of the Tasks table.
type TaskRow struct {
	Name string
	// Mark is how the task last ran, and Result says it in words ("passed",
	// "failed"); Time is how long it took ("" when it did not run).
	Mark   Mark
	Result string
	Time   string
	// Cache is "hit", "miss" or "miss · why" once it ran; "" when smith does
	// not know, and Cached false for a task smith never caches.
	Cache  string
	Cached bool
	// Key runs the task; "" when it has none.
	Key string
	// OwnTerminal is true for a task that needs a terminal of its own (start).
	OwnTerminal bool
}

// TaskTable draws the table of the Tasks section in the look of
// bubbles/table: a muted header, a rule, then one row per task with Task,
// Last result (its mark and its words), Time, Cache and Key. The row at the
// cursor has the pointer and is on the Highlight background while the detail
// has the keyboard. A
// table with a task that needs its own terminal ends with a line that says so.
//
// Parameters: th is the theme; rows are the tasks; w is the width; cursor is
// the row of the keyboard; focused is true while the detail has the keyboard;
// tick moves the spinner of a running task.
// Returns: the lines, and the row of the table on each line of a task.
func TaskTable(th theme.Theme, rows []TaskRow, w, cursor int, focused bool, tick int) (lines []string, rowOf map[int]int) {
	rowOf = map[int]int{}
	if len(rows) == 0 {
		return []string{pane.Fit(" "+th.Muted(words.Say("projectpage.tasks.none")), w)}, rowOf
	}
	nameW := 6
	for _, r := range rows {
		nameW = max(nameW, pane.Width(r.Name)+2)
	}
	nameW = min(nameW, 18)
	resultW, timeW, keyW := 16, 9, 5
	cacheW := min(w-3-nameW-resultW-timeW-keyW, 26)
	if cacheW < 6 {
		cacheW = 0
	}
	if w-3-nameW-resultW-cacheW-keyW < timeW {
		timeW = 0
	}
	resultHead := words.Say("projectpage.tasks.result")
	if w < 64 {
		resultHead = words.Say("projectpage.tasks.result.short")
	}
	head := "   " + pane.Cell(words.Say("projectpage.tasks.task"), nameW) + pane.Cell(resultHead, resultW)
	if timeW > 0 {
		head += pane.Cell(words.Say("projectpage.tasks.time"), timeW)
	}
	if cacheW > 0 {
		head += pane.Cell(words.Say("projectpage.tasks.cache"), cacheW)
	}
	head += pane.Cell(words.Say("projectpage.tasks.key"), keyW)
	lines = append(lines, pane.Fit(th.Muted(head), w), " "+th.Border(strings.Repeat("─", max(w-2, 1)), false))
	own := ""
	for i, r := range rows {
		result := Glyph(th, r.Mark, tick) + " " + r.Result
		if r.OwnTerminal && r.Mark == NotRun {
			result = th.Muted(words.Say("projectpage.tasks.own_terminal"))
			own = r.Name
		}
		chosen := i == cursor && focused
		name, lead := pane.Cell(r.Name, nameW), "   "
		if chosen {
			name, lead = th.Accent(th.Bold(name)), " "+th.Accent(Pointer)
		}
		line := lead + name + pane.Cell(result, resultW)
		if timeW > 0 {
			line += pane.Cell(orDash(r.Time), timeW)
		}
		if cacheW > 0 {
			line += th.Muted(pane.Cell(cacheText(r), cacheW))
		}
		line += th.Accent(pane.Cell(r.Key, keyW))
		line = pane.Fit(line, w)
		if chosen {
			line = th.Selected(line)
		}
		rowOf[len(lines)] = i
		lines = append(lines, line)
	}
	if own != "" {
		lines = append(lines, "")
		for _, part := range pane.Wrap(th.Muted(words.Say("projectpage.tasks.own_terminal.note", "task", own)), max(w-2, 8), 0) {
			lines = append(lines, " "+part)
		}
	}
	return lines, rowOf
}

// cacheText is what the Cache column says of a task.
func cacheText(r TaskRow) string {
	switch {
	case !r.Cached:
		return words.Say("projectpage.tasks.never_cached")
	case r.Cache == "":
		return "–" // words:literal
	}
	return r.Cache
}

// orDash is the text, or a dash when it is empty.
func orDash(s string) string {
	if s == "" {
		return "–" // words:literal
	}
	return s
}
