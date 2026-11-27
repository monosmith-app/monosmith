package runpage

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// The two panes of a narrow Runs view, as the ids of the parts of the switch.
const (
	TasksID = "x:run-tasks"
	LogID   = "x:run-log"
)

// Span is a part of a line that answers a click: the cells [From, To) and its id.
type Span struct {
	From, To int
	ID       string
}

// Switch draws the line that chooses between the tasks and the log on a narrow
// screen: " Tasks   Log: web:test", the one shown as the active tab, and the
// hint "tab switches" at the right when it fits.
//
// Parameters: th is the theme; task is the task whose log is shown ("" for all
// output); onLog is true while the log is shown; w is the width.
// Returns: the line and the spans of its two parts.
func Switch(th theme.Theme, task string, onLog bool, w int) (string, []Span) {
	tasks := " " + words.Say("runpage.switch.tasks") + " "
	log := " " + words.Say("runpage.switch.log") + " "
	if task != "" {
		log = " " + words.Say("runpage.switch.log_of", "task", task) + " "
	}
	style := func(s string, on bool) string {
		if on {
			return th.ActiveTab(s)
		}
		return th.Muted(s)
	}
	left := " " + style(tasks, !onLog) + "  " + style(log, onLog)
	tw := pane.Width(tasks)
	spans := []Span{{From: 1, To: 1 + tw, ID: TasksID}, {From: 3 + tw, To: 3 + tw + pane.Width(log), ID: LogID}}
	hint := th.Faint(words.Say("runpage.switch.hint"))
	if gap := w - pane.Width(left) - pane.Width(hint) - 1; gap >= 2 {
		return pane.Fit(left+strings.Repeat(" ", gap)+hint, w), spans
	}
	return pane.Fit(left, w), spans
}
