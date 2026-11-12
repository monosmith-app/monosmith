package composer

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/choicerow"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/libs/words/words"
)

// Title is the frame title of the screen.
func (m Model) Title() string {
	switch m.stage {
	case stQuestion:
		return words.Say("composer.title.question")
	case stName:
		return words.Say("composer.title.name")
	case stPlan:
		return words.Say("composer.title.plan")
	}
	return words.Say("composer.title")
}

// Lines draws the screen, framed, at the size Place set.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 20 || r.H < 8 {
		return nil
	}
	inner, height := r.W-4, r.H-2
	var body []string
	switch m.stage {
	case stQuestion:
		body = m.questionBody(th, inner)
	case stName:
		body = m.nameBody(th, inner)
	case stPlan:
		body = m.planBody(th, inner)
	default:
		body = m.browseBody(th, inner, height)
	}
	rows := make([]string, 0, height)
	for _, l := range body {
		if len(rows) >= height {
			break
		}
		rows = append(rows, " "+pane.Fit(l, inner))
	}
	return pane.Frame(th, m.Title(), rows, r.W, r.H, true)
}

// browseBody is the search box, the list (catalogue or stack), the check and
// the keys.
func (m Model) browseBody(th theme.Theme, inner, height int) []string {
	lines := []string{m.search.View(), m.summaryLine(th), ""}
	status := m.statusLines(th, inner)
	keys := m.keyLines(th, inner)
	room := max(height-len(lines)-len(status)-len(keys), 3)
	list, _ := m.listLines(th, inner)
	for i := range list {
		list[i] = pane.Fit(list[i], inner-1) // the scrollbar takes the last cell
	}
	shown, _ := pane.Window(th, list, inner, room, m.cursorLine, m.scroll)
	lines = append(lines, shown...)
	for len(lines) < height-len(status)-len(keys) {
		lines = append(lines, "")
	}
	return append(append(lines, status...), keys...)
}

// keyLines are the keys of the browse screen. They wrap, as the status does, so
// no key is cut off on a narrow screen.
func (m Model) keyLines(th theme.Theme, inner int) []string {
	target := words.Say(map[bool]string{true: "composer.keys.catalogue", false: "composer.keys.stack"}[m.stack])
	return pane.Wrap(th.Muted(words.Say("composer.keys.browse", "target", target)), inner, 2)
}

// summaryLine says what is ticked, by layer.
func (m Model) summaryLine(th theme.Theme) string {
	if len(m.ticked) == 0 && m.lang == nil {
		return th.Muted(words.Say("composer.summary.empty"))
	}
	var parts []string
	for _, it := range m.entries() {
		parts = append(parts, it.Title)
	}
	return th.Muted(words.Say("composer.stack.label")) + strings.Join(parts, " + ")
}

// listLines draws the rows of the catalogue or of the stack and sets where the
// cursor row is, so the window can follow it. items has, for each line, the
// index of the row it draws, or -1 for a heading or a note.
func (m *Model) listLines(th theme.Theme, inner int) (lines []string, items []int) {
	add := func(line string, item int) {
		lines, items = append(lines, line), append(items, item)
	}
	if m.stack {
		entries := m.entries()
		last := ""
		for i, it := range entries {
			if it.Layer != last {
				last = it.Layer
				add(th.Muted(m.layerTitle(it.Layer)), -1)
			}
			if i == m.cursor {
				m.cursorLine = len(lines)
			}
			add(m.row(th, it, i, inner, true), i)
		}
		if len(entries) == 0 {
			add(th.Muted(words.Say("composer.list.empty.stack")), -1)
		}
		return lines, items
	}
	last := ""
	grouped := m.search.Value() == ""
	for i, h := range m.hits {
		if grouped && h.Layer != last {
			last = h.Layer
			add(th.Muted(m.layerTitle(h.Layer)+": "+m.layerAbout(h.Layer)), -1)
		}
		if i == m.cursor {
			m.cursorLine = len(lines)
		}
		add(m.row(th, h.Item, i, inner, false), i)
	}
	if len(m.hits) == 0 {
		add(th.Muted(words.Say("composer.list.empty.search")), -1)
	}
	return lines, items
}

// row is the item at index i of the list: a box, the title, where it is for,
// and what it is, drawn as the one pick-list row (see choicerow.Lines).
func (m Model) row(th theme.Theme, it compose.Item, i, inner int, inStack bool) string {
	label := it.Title
	if it.Language != "" && it.Kind != compose.KindLanguage {
		label += th.Muted(" (" + it.Language + ")")
	}
	switch {
	case it.Default:
		label += th.Muted(words.Say("composer.row.default"))
	case it.Suggested:
		label += th.Muted(words.Say("composer.row.suggested"))
	}
	if it.Problem != "" {
		label += " " + th.Warning(words.Say("composer.row.problem", "problem", it.Problem))
	}
	r := choicerow.Ticks(label, m.isTicked(it), it.Kind == compose.KindPlugin)
	if !inStack {
		r.Hint = it.Description
	}
	r.Selected, r.Hovered = i == m.cursor, i == m.hover && i != m.cursor
	return choicerow.Line(th, r, inner-1)
}

// isTicked reports whether the item is ticked.
func (m Model) isTicked(it compose.Item) bool {
	if it.Kind == compose.KindLanguage {
		return m.lang != nil && m.lang.ID == it.ID
	}
	return slices.ContainsFunc(m.ticked, func(t compose.Item) bool { return t.ID == it.ID })
}

func (m Model) layerTitle(id string) string {
	for _, l := range m.cat.Layers {
		if l.ID == id {
			return l.Title
		}
	}
	return id
}

func (m Model) layerAbout(id string) string {
	for _, l := range m.cat.Layers {
		if l.ID == id {
			return l.About
		}
	}
	return ""
}

// statusLines are the live check: the conflict, what is missing, the open
// question, the defaults that will be added, or that the stack is ready.
func (m Model) statusLines(th theme.Theme, inner int) []string {
	var out []string
	add := func(text string) { out = append(out, pane.Wrap(text, inner, 2)...) }
	switch {
	case m.note != "":
		add(th.Warning(m.note))
	}
	switch {
	case m.chk.Conflict != "":
		add(th.Failure(words.Say("composer.status.conflict")) + m.chk.Conflict)
	case m.chk.Question != nil:
		add(th.Warning(words.Say("composer.status.answer")) + words.Say("composer.status.answer.body", "choice", m.chk.Question.Choice, "axis", m.chk.Question.Axis))
	case m.chk.Missing != "":
		add(th.Muted(words.Say("composer.status.next", "missing", m.chk.Missing)))
	case m.chk.Complete:
		line := th.Success(words.Say("composer.status.ready")) + words.Say("composer.status.ready.body", "selection", strings.Join(m.chk.Selection, ", "), "language", m.chk.Language)
		if len(m.chk.Added) > 0 {
			line += words.Say("composer.status.added", "added", strings.Join(m.chk.Added, ", "))
		}
		add(line + words.Say("composer.status.plan"))
	}
	return out
}

// questionBody asks the open axis, with what each answer means.
func (m Model) questionBody(th theme.Theme, inner int) []string {
	q := m.chk.Question
	if q == nil {
		return nil
	}
	lines := []string{th.Bold(words.Say("composer.question.title", "axis", q.Axis, "choice", q.Choice))}
	if q.About != "" {
		lines = append(lines, pane.Wrap(th.Muted(words.Say("composer.question.writes", "choice", q.Choice, "about", q.About)), inner, 2)...)
	}
	lines = append(lines, th.Muted(words.Say("composer.question.hint")), "")
	for i, o := range q.Options {
		label := o.Title
		if o.Suggested {
			label += th.Muted(words.Say("composer.row.suggested"))
		}
		r := choicerow.Row{Label: label, Hint: o.Description, Selected: i == m.qcursor, Hovered: i == m.hover && i != m.qcursor}
		lines = append(lines, choicerow.Line(th, r, inner))
	}
	return append(lines, "", th.Muted(words.Say("composer.question.keys")))
}

// nameBody asks the name and the kind.
func (m Model) nameBody(th theme.Theme, inner int) []string {
	lines := []string{
		th.Bold(words.Say("composer.name.title")),
		th.Muted(words.Say("composer.name.stack", "selection", strings.Join(m.chk.Selection, ", "), "language", m.chk.Language)),
		"",
		words.Say("composer.name.field", "field", m.name.View()),
		words.Say("composer.name.kind") + th.Bold(kinds[m.kind]) + th.Muted("   "+kindMeaning(kinds[m.kind])),
	}
	if m.note != "" {
		lines = append(lines, "", th.Warning(m.note))
	}
	return append(lines, "", th.Muted(words.Say("composer.name.keys")))
}

// kindMeaning says what a kind of project is.
func kindMeaning(kind string) string {
	switch kind {
	case "app":
		return words.Say("composer.kind.app")
	case "lib":
		return words.Say("composer.kind.lib")
	}
	return words.Say("composer.kind.package")
}

// planBody shows the plan: the command, the files, the tools, the notes, and
// what can be done with it.
func (m Model) planBody(th theme.Theme, inner int) []string {
	if m.planErr != "" {
		return append(pane.Wrap(th.Failure(words.Say("composer.plan.failed"))+m.planErr, inner, 2), "", th.Muted(words.Say("composer.plan.back")))
	}
	if m.plan == nil {
		return []string{th.Muted(words.Say("composer.plan.reading"))}
	}
	p := m.plan
	// Every line is wrapped to the dialog, with a hanging indent, never cut.
	var lines []string
	add := func(l ...string) {
		for _, one := range l {
			lines = append(lines, pane.Wrap(one, inner, 4)...)
		}
	}
	add(th.Bold(words.Say("composer.plan.run")), "  "+th.Accent(p.Command), "")
	if len(p.Files) > 0 {
		add(th.Bold(words.Say("composer.plan.writes")))
		for _, f := range p.Files {
			text := "  " + f.Action + " " + f.Path
			if f.Note != "" {
				text += th.Muted("  " + f.Note)
			}
			add(text)
		}
		add("")
	}
	if len(p.Tools) > 0 {
		add(th.Bold(words.Say("composer.plan.tools")))
		for _, t := range p.Tools {
			text := "  " + t.Title + th.Muted(words.Say("composer.plan.tool", "state", t.State, "choice", t.Choice))
			if t.State != "installed" {
				text += th.Muted(": " + t.Command)
			}
			add(text)
		}
		add("")
	}
	for _, n := range p.Notes {
		add(th.Muted(n))
	}
	if p.RecipeNote != "" {
		add(th.Muted(words.Say("composer.plan.recipe", "note", p.RecipeNote)))
	}
	add("", th.Muted(words.Say("composer.plan.keys")))
	return lines
}
