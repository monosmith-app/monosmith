package tui

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/detail"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// buttonRow is a line of buttons and what a click on each of them does.
type buttonRow struct {
	line string
	hit  hit
}

// viewButton is a button of a view: the command it stands for and its label.
type viewButton struct{ id, label string }

// buttonLines draws buttons from the second cell of w-cell lines, moving a
// button that does not fit to the next line, with the span of each for a click.
func (s Shell) buttonLines(buttons []viewButton, w int, under string) []buttonRow {
	var out []buttonRow
	cur := buttonRow{line: " ", hit: hit{kind: "action"}}
	x := 1
	for _, b := range buttons {
		key := s.shortcutHere(b.id)
		bw := button.Width(s.th, b.label, key)
		if x+bw > w && len(cur.hit.spans) > 0 {
			out = append(out, cur)
			cur, x = buttonRow{line: " ", hit: hit{kind: "action"}}, 1
		}
		state := s.buttonState(b.id, under)
		if c, ok := commands.ByID(b.id); ok && c.Unavailable(s.state()) != "" {
			state = button.Disabled
		}
		cur.hit.spans = append(cur.hit.spans, hitSpan{from: x, to: x + bw, id: b.id})
		cur.line += button.Render(s.th, b.label, key, state) + "  "
		x += bw + 2
	}
	return append(out, cur)
}

// depsBox is the box of actions at the foot of the Dependencies view: what is
// marked (or what u updates when nothing is), the buttons that update it, take
// every safe release or repair the installed dependencies, and the command
// lines the update and its dry run run, dim. Each button asks before it changes anything. under
// is the id of the button under the pointer. It is empty until the view loaded.
func (s Shell) depsBox(w int, under string) []buttonRow {
	if !s.deps.Loaded || s.deps.Err != nil || w < 24 {
		return nil
	}
	var body []string
	if line := s.depsMarkedLine(); line != "" {
		body = append(body, " "+line, "")
	}
	var buttons []viewButton
	if len(s.deps.All) > 0 {
		label := words.Say("depspage.button.update_marked")
		if len(s.deps.Marked) == 0 {
			if r, ok := s.deps.Selected(); ok {
				label = words.Say("depspage.button.update_one", "name", r.Name)
			}
		}
		buttons = append(buttons, viewButton{"deps.update", label}, viewButton{"deps.dry-run", words.Say("depspage.button.dry_run")}, viewButton{"deps.up-all", words.Say("depspage.button.update_safe")})
	}
	buttons = append(buttons, viewButton{"deps.repair", words.Say("tui.deps_repair.button")})
	btns := s.buttonLines(buttons, w-3, under)
	at := len(body)
	for _, b := range btns {
		body = append(body, b.line)
	}
	if argv := s.deps.UpArgv(); argv != nil {
		body = append(body, " "+s.th.Faint(pane.Cut(words.Say("depspage.box.command", "command", strings.Join(argv, " ")), w-4)))
		// The dry run's line, so the button says what it runs before it is pressed.
		body = append(body, " "+s.th.Faint(pane.Cut(words.Say("depspage.box.command", "command", strings.Join(s.deps.DryRunArgv(), " ")), w-4)))
	}
	var out []buttonRow
	for i, line := range pane.Frame(s.th, "", body, w, len(body)+2, false) {
		row := buttonRow{line: line}
		if b := i - at - 1; b >= 0 && b < len(btns) {
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

// depsMarkedLine says what an update would take: the one package marked and
// in how many projects, the names of several, or, with nothing marked, the
// package under the cursor that u updates.
func (s Shell) depsMarkedLine() string {
	var marked []depsview.Row
	for _, r := range s.deps.All {
		if s.deps.Marked[r.Name] {
			marked = append(marked, r)
		}
	}
	switch len(marked) {
	case 0:
		if r, ok := s.deps.Selected(); ok {
			return s.th.Muted(words.Say("depspage.box.cursor", "key", commands.ShortcutOf("deps.update"), "name", r.Name, "version", r.Latest))
		}
		return ""
	case 1:
		r := marked[0]
		return words.Say("depspage.box.one", "name", r.Name, "version", r.Latest, "projects", words.Count("depspage.box.projects", len(r.Projects)))
	}
	var names []string
	for _, r := range marked {
		names = append(names, r.Name)
	}
	return words.Say("depspage.box.many", "count", len(marked), "names", strings.Join(names, ", "))
}

// depsPanel is the detail of the Dependencies view: the package under the
// cursor, with the folder of each project that uses it.
func (s Shell) depsPanel() []detail.Section {
	deps := s.deps
	deps.Links = s.hyperlinks
	roots := map[string]string{}
	for _, p := range s.list.Projects() {
		roots[p.Name] = p.Root
	}
	return deps.Detail(s.th, max(s.detailRect().W-4, 20), func(name string) string { return roots[name] })
}
