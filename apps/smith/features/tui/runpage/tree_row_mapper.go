package runpage

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/choicerow"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// TaskLine is what a task row of the tree tells: its mark (already styled), its
// time and its cache note ("hit", "remote", "" for none).
type TaskLine struct {
	Mark, Took, Cache string
	// Selected is true for the task whose log is shown.
	Selected bool
}

// HeadingLine draws the row that names a project in w cells: the name in bold
// after the cells of the cursor mark. A task without a project is under the
// heading "Other tasks".
//
// Parameters: th is the theme; project is the name; w is the width.
// Returns: the row, w cells wide.
func HeadingLine(th theme.Theme, project string, w int) string {
	if project == "" {
		project = words.Say("runpage.tree.other")
	}
	return pane.Fit(choicerow.Cursor(th, false)+th.Bold(project), w)
}

// TaskRow draws a task row of the tree in w cells: the cursor mark, the branch
// (├─, or └─ on the last task of a project), the mark, the task's name, then
// its time and its cache note at the right. The selected row has the cursor
// mark and an accent name, as every pick list.
//
// Parameters: th is the theme; r is the row; l is what it tells; w is the width.
// Returns: the row, w cells wide.
func TaskRow(th theme.Theme, r Row, l TaskLine, w int) string {
	branch := "├─ "
	if r.Last {
		branch = "└─ "
	}
	cursor := choicerow.Cursor(th, l.Selected)
	name := TaskOf(r.TaskID)
	if l.Selected {
		cursor, name = th.Accent(cursor), th.Accent(name)
	}
	left := cursor + th.Faint(branch) + l.Mark + " " + name
	right := strings.TrimSpace(l.Took + "  " + l.Cache)
	gap := w - pane.Width(left) - pane.Width(right) - 1
	if right == "" || gap < 1 {
		return pane.Fit(left, w)
	}
	return pane.Fit(left+strings.Repeat(" ", gap)+th.Muted(right), w)
}

// SessionLine is a run of the session as its row tells it: the plain mark of
// how it went, its command line and how long ago it ended.
type SessionLine struct {
	Mark, Command, Ago string
	// Selected is true for the run the view shows.
	Selected bool
}

// SessionRow draws a run of the session in w cells: the cursor mark, the mark,
// the command line (cut with an ellipsis when it is too long) and, muted at the
// right, how long ago it ended. The run shown has the accent of every pick list.
//
// Parameters: th is the theme; l is the run; w is the width.
// Returns: the row, w cells wide.
func SessionRow(th theme.Theme, l SessionLine, w int) string {
	ago := " " + l.Ago
	room := max(w-pane.Width(choicerow.Cursor(th, false))-pane.Width(ago), 4)
	text := pane.Fit(pane.CutEllipsis(l.Mark+" "+l.Command, room), room)
	cursor := choicerow.Cursor(th, l.Selected)
	if l.Selected {
		cursor, text = th.Accent(cursor), th.Accent(text)
	}
	return pane.Fit(cursor+text+th.Muted(ago), w)
}

// AllOutputRow draws the first row of the tree, which shows the whole output
// of the run: "All output" and the count of its lines.
//
// Parameters: th is the theme; lines counts the lines of the output; selected
// is true while the whole output is shown; w is the width.
// Returns: the row, w cells wide.
func AllOutputRow(th theme.Theme, lines int, selected bool, w int) string {
	cursor := choicerow.Cursor(th, selected)
	label := words.Say("runpage.tree.all_output")
	if selected {
		cursor, label = th.Accent(cursor), th.Accent(label)
	}
	return pane.Fit(cursor+label+"  "+th.Muted(words.Count("runpage.tree.lines", lines)), w)
}
