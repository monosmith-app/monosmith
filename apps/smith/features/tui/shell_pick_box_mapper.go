package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// pickRuns are the run keys the box of picked projects offers, in its order.
var pickRuns = []string{"run.build", "run.test", "run.lint", "run.typecheck"}

// pickBox is the box at the foot of the list of Projects while projects are
// picked: which ones, a button for each task at least one of them has, and the
// command line of the first button, dim. A button runs what its key runs:
// smith run <task> with a -p for each picked project. It is empty while nothing
// is picked or the list is too narrow for it.
func (s Shell) pickBox(w int, under string) []buttonRow {
	names := s.pickedNames()
	if len(names) == 0 || w < 24 {
		return nil
	}
	has := map[string]bool{}
	for _, n := range names {
		p, _ := s.list.Project(n)
		for _, t := range p.Tasks {
			has[t.Name] = true
		}
	}
	var buttons []viewButton
	var first commands.Command
	for _, id := range pickRuns {
		c, ok := commands.ByID(id)
		if !ok || !has[c.Task] {
			continue
		}
		if len(buttons) == 0 {
			first = c
		}
		buttons = append(buttons, viewButton{id, c.Title})
	}
	body := []string{" " + pane.CutEllipsis(words.Say("tui.pick.box", "count", len(names), "names", strings.Join(names, ", ")), w-4), ""}
	btns := s.buttonLines(buttons, w-3, under)
	at := len(body)
	if len(buttons) > 0 {
		for _, b := range btns {
			body = append(body, b.line)
		}
		line := commands.Line(first.TaskArgvAcross(names))
		body = append(body, " "+s.th.Faint(pane.Cut(words.Say("tui.pick.command", "command", line), w-4)))
	} else {
		body = append(body, " "+s.th.Muted(pane.CutEllipsis(words.Say("tui.pick.no_task"), w-4)))
	}
	body = append(body, " "+s.th.Muted(pane.CutEllipsis(words.Say("tui.pick.clear", "key", commands.KeyLabel("esc")), w-4)))
	var out []buttonRow
	for i, line := range pane.Frame(s.th, "", body, w, len(body)+2, false) {
		row := buttonRow{line: line}
		if b := i - at - 1; len(buttons) > 0 && b >= 0 && b < len(btns) {
			// The frame's border takes the first cell.
			for _, sp := range btns[b].hit.spans {
				row.hit.spans = append(row.hit.spans, hitSpan{from: sp.from + 1, to: sp.to + 1, id: sp.id})
			}
			row.hit.kind = "action"
		}
		out = append(out, row)
	}
	return out
}
