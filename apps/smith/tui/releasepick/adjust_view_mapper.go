package releasepick

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// intro is the text above the list.
func intro(inner int) []string {
	return pane.Wrap(words.Say("releasepick.intro"), inner, 0)
}

// tailLines are the command that will run and the blank line before it.
func (m Model) tailLines(inner int) []string {
	tail := []string{"", words.Say("releasepick.command.label")}
	for _, l := range pane.Wrap(commands.Line(m.Argv()), inner-4, 2) {
		tail = append(tail, "  $ "+l)
	}
	return tail
}

// rowText is one package: the hold box, its versions and its bump.
func (m Model) rowText(it release.Item, now map[string]release.Item, nameW int) string {
	box := choicerow.Box(m.held[it.Project], false)
	from := it.Current
	if from == "" {
		from = words.Say("releasepick.version.new")
	}
	name := pane.Fit(it.Project, nameW)
	switch {
	case m.held[it.Project]:
		return words.Say("releasepick.row.held", "box", box, "name", name, "from", from)
	case !m.canBump(it):
		return words.Say("releasepick.row.fixed", "box", box, "name", name, "from", from, "to", it.Next, "because", m.fixedBecause(it))
	}
	next, ok := now[it.Project]
	if !ok {
		return words.Say("releasepick.row.dependency", "box", box, "name", name, "from", from)
	}
	choice := words.Say("releasepick.bump.commits")
	if level := bumpLevels[m.bump[it.Project]]; level != "" {
		choice = level
	}
	return words.Say("releasepick.row.bump", "box", box, "name", name, "from", from, "to", next.Next, "choice", choice)
}

// fixedBecause says why a package has no bump to choose.
func (m Model) fixedBecause(it release.Item) string {
	switch {
	case m.plan.Channel != "":
		return words.Say("releasepick.fixed.channel")
	case it.Forced:
		return words.Say("releasepick.fixed.forced")
	}
	return words.Say("releasepick.fixed.first")
}

// Lines draws the dialog: Rect().H lines of Rect().W cells.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 6 || r.H < 3 {
		return nil
	}
	inner := r.W - 4
	row := func(s string) string { return " " + pane.Fit(s, inner) }
	var rows []string
	for _, l := range intro(inner) {
		rows = append(rows, row(l))
	}
	rows = append(rows, row(""))
	nameW := 8
	for _, it := range m.plan.Items {
		nameW = min(max(nameW, ansi.StringWidth(it.Project)), 24)
	}
	now := m.adjusted()
	end := min(m.top+m.room, len(m.plan.Items))
	for i := m.top; i < end; i++ {
		it := m.plan.Items[i]
		cursor := choicerow.Cursor(th, false)
		line := m.rowText(it, now, nameW)
		if i == m.cursor && m.focus < 0 {
			cursor, line = th.Accent(choicerow.Cursor(th, true)), th.Selected(pane.Fit(line, inner-2))
		}
		rows = append(rows, row(cursor+line))
	}
	for len(rows) < m.head+m.room {
		rows = append(rows, row(""))
	}
	tail := m.tailLines(inner)
	for i, l := range tail {
		if i >= 2 {
			l = th.Bold(l)
		}
		rows = append(rows, row(l))
	}
	rows = append(rows, row(""))
	labels, keys := buttonLabels()
	line := strings.Repeat(" ", button.RowSpans(th, labels, keys, inner)[0][0]-2)
	for i, label := range labels {
		state := button.Rest
		switch {
		case m.hover == i:
			state = button.Hover
		case m.focus == i:
			state = button.Active
		}
		line += button.Render(th, label, keys[i], state)
		if i == 0 {
			line += " "
		}
	}
	rows = append(rows, row(line))
	if len(m.plan.Items) > m.room {
		// The list is longer than the dialog: say so on the line above the buttons.
		rows[len(rows)-2] = row(th.Muted(words.Say("releasepick.shown", "shown", m.room, "total", len(m.plan.Items))))
	}
	return pane.Frame(th, m.Title(), rows, r.W, r.H, true)
}
