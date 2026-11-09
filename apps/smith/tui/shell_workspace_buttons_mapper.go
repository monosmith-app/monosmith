package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// buttonRow is a line of buttons and what a click on each of them does.
type buttonRow struct {
	line string
	hit  hit
}

// workspaceButtons are the actions on the whole workspace, side by side: the
// button that adds a project, then verify, build, test, lint and typecheck of
// every project. A button that does not fit in w cells starts the next line, so
// none is cut. under is the id of the button under the pointer.
func (s Shell) workspaceButtons(w int, under string) []buttonRow {
	buttons := []struct{ id, label string }{
		{"file.add", words.Say("tui.main.new_project")},
		{"run.verify", words.Say("tui.main.run_verify")},
		{"run.build-all", words.Say("tui.main.run_build")},
		{"run.test-all", words.Say("tui.main.run_test")},
		{"run.lint-all", words.Say("tui.main.run_lint")},
		{"run.typecheck-all", words.Say("tui.main.run_typecheck")},
	}
	var rows []buttonRow
	var line strings.Builder
	row, x := hit{kind: "action"}, 0
	flush := func() {
		rows = append(rows, buttonRow{line: line.String(), hit: row})
		line.Reset()
		row, x = hit{kind: "action"}, 0
	}
	for _, b := range buttons {
		key := s.shortcutHere(b.id)
		bw := button.Width(s.th, b.label, key)
		if x > 0 && x+1+bw > w {
			flush()
		}
		if x > 0 {
			line.WriteString(" ")
			x++
		}
		state := s.buttonState(b.id, under)
		if c, ok := commands.ByID(b.id); ok && c.Unavailable(s.state()) != "" {
			state = button.Disabled
		}
		row.spans = append(row.spans, hitSpan{from: x, to: x + bw, id: b.id})
		line.WriteString(button.Render(s.th, b.label, key, state))
		x += bw
	}
	flush()
	return rows
}
