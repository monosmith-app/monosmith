package composer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Title is the frame title of the screen.
func (m Model) Title() string {
	switch m.stage {
	case stQuestion:
		return "Compose a stack: an answer is needed"
	case stName:
		return "Compose a stack: name the project"
	case stPlan:
		return "Compose a stack: the plan"
	}
	return "Compose a stack"
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
	keys := th.Muted("enter tick · tab " + map[bool]string{true: "catalogue", false: "your stack"}[m.stack] + " · ctrl+n next · esc back")
	room := max(height-len(lines)-len(status)-1, 3)
	list := m.listLines(th, inner)
	top := 0
	if m.cursorLine >= room {
		top = m.cursorLine - room + 1
	}
	lines = append(lines, list[min(top, len(list)):min(top+room, len(list))]...)
	for len(lines) < height-len(status)-1 {
		lines = append(lines, "")
	}
	return append(append(lines, status...), keys)
}

// summaryLine says what is ticked, by layer.
func (m Model) summaryLine(th theme.Theme) string {
	if len(m.ticked) == 0 && m.lang == nil {
		return th.Muted("Nothing ticked yet. Search, then press enter on what you want.")
	}
	var parts []string
	for _, it := range m.entries() {
		parts = append(parts, it.Title)
	}
	return th.Muted("Stack: ") + strings.Join(parts, " + ")
}

// listLines draws the rows of the catalogue or of the stack and sets where the
// cursor row is, so the window can follow it.
func (m *Model) listLines(th theme.Theme, inner int) []string {
	var lines []string
	if m.stack {
		entries := m.entries()
		last := ""
		for i, it := range entries {
			if it.Layer != last {
				last = it.Layer
				lines = append(lines, th.Muted(m.layerTitle(it.Layer)))
			}
			if i == m.cursor {
				m.cursorLine = len(lines)
			}
			lines = append(lines, m.row(th, it, i == m.cursor, true))
		}
		if len(entries) == 0 {
			lines = append(lines, th.Muted("Nothing here yet. Press tab to go back to the catalogue."))
		}
		return lines
	}
	last := ""
	grouped := m.search.Value() == ""
	for i, h := range m.hits {
		if grouped && h.Layer != last {
			last = h.Layer
			lines = append(lines, th.Muted(m.layerTitle(h.Layer)+": "+m.layerAbout(h.Layer)))
		}
		if i == m.cursor {
			m.cursorLine = len(lines)
		}
		lines = append(lines, m.row(th, h.Item, i == m.cursor, false))
	}
	if len(m.hits) == 0 {
		lines = append(lines, th.Muted("Nothing matches. Try a word like web, desktop or serverless, or esc to clear."))
	}
	return lines
}

// row is one item: a box, the title, where it is for, and what it is.
func (m Model) row(th theme.Theme, it compose.Item, selected, inStack bool) string {
	box := "[ ]"
	if m.isTicked(it) {
		box = "[x]"
	}
	if it.Kind == compose.KindPlugin {
		box = "[-]"
	}
	text := box + " " + it.Title
	if it.Language != "" && it.Kind != compose.KindLanguage {
		text += th.Muted(" (" + it.Language + ")")
	}
	switch {
	case it.Default:
		text += th.Muted(" (default)")
	case it.Suggested:
		text += th.Muted(" (suggested)")
	}
	if !inStack && it.Description != "" {
		text += th.Muted("  " + it.Description)
	}
	if it.Problem != "" {
		text += " " + th.Warning("(cannot be used: "+it.Problem+")")
	}
	if selected {
		return th.Selected("> " + text)
	}
	return "  " + text
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
		add(th.Failure("Conflict: ") + m.chk.Conflict)
	case m.chk.Question != nil:
		add(th.Warning("An answer is needed: ") + fmt.Sprintf("%s needs a choice on the %s axis. Press ctrl+n to answer; smith does not choose it for you.", m.chk.Question.Choice, m.chk.Question.Axis))
	case m.chk.Missing != "":
		add(th.Muted("Next: " + m.chk.Missing))
	case m.chk.Complete:
		line := th.Success("Ready: ") + strings.Join(m.chk.Selection, ", ") + " for " + m.chk.Language + "."
		if len(m.chk.Added) > 0 {
			line += " Added by default: " + strings.Join(m.chk.Added, ", ") + "."
		}
		add(line + " Press ctrl+n for the plan.")
	}
	return out
}

// questionBody asks the open axis, with what each answer means.
func (m Model) questionBody(th theme.Theme, inner int) []string {
	q := m.chk.Question
	if q == nil {
		return nil
	}
	lines := []string{th.Bold(fmt.Sprintf("Which %s for %s?", q.Axis, q.Choice))}
	if q.About != "" {
		lines = append(lines, pane.Wrap(th.Muted(q.Choice+" writes: "+q.About), inner, 2)...)
	}
	lines = append(lines, th.Muted("smith does not choose this for you. The first line is suggested; enter takes the one under the cursor."), "")
	for i, o := range q.Options {
		text := o.Title
		if o.Suggested {
			text += th.Muted(" (suggested)")
		}
		if o.Description != "" {
			text += th.Muted("  " + o.Description)
		}
		if i == m.qcursor {
			lines = append(lines, th.Selected("> "+text))
		} else {
			lines = append(lines, "  "+text)
		}
	}
	return append(lines, "", th.Muted("up/down choose · enter answer · esc back to the list"))
}

// nameBody asks the name and the kind.
func (m Model) nameBody(th theme.Theme, inner int) []string {
	lines := []string{
		th.Bold("Almost there. Name the project."),
		th.Muted("Stack: " + strings.Join(m.chk.Selection, ", ") + " (" + m.chk.Language + ")"),
		"",
		"Name: " + m.name.View(),
		"Kind: " + th.Bold(kinds[m.kind]) + th.Muted("   "+kindMeaning(kinds[m.kind])),
	}
	if m.note != "" {
		lines = append(lines, "", th.Warning(m.note))
	}
	return append(lines, "", th.Muted("enter shows the plan · tab changes the kind · esc back"))
}

// kindMeaning says what a kind of project is.
func kindMeaning(kind string) string {
	switch kind {
	case "app":
		return "something you run or deploy"
	case "lib":
		return "code other projects here share"
	}
	return "code you publish for others to install"
}

// planBody shows the plan: the command, the files, the tools, the notes, and
// what can be done with it.
func (m Model) planBody(th theme.Theme, inner int) []string {
	if m.planErr != "" {
		return append(pane.Wrap(th.Failure("The plan could not be made: ")+m.planErr, inner, 2), "", th.Muted("esc goes back"))
	}
	if m.plan == nil {
		return []string{th.Muted("Reading the plan…")}
	}
	p := m.plan
	// Every line is wrapped to the dialog, with a hanging indent, never cut.
	var lines []string
	add := func(l ...string) {
		for _, one := range l {
			lines = append(lines, pane.Wrap(one, inner, 4)...)
		}
	}
	add(th.Bold("Run this:"), "  "+th.Accent(p.Command), "")
	if len(p.Files) > 0 {
		add(th.Bold("It writes:"))
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
		add(th.Bold("It needs these tools:"))
		for _, t := range p.Tools {
			text := "  " + t.Title + th.Muted(" ("+t.State+", for "+t.Choice+")")
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
		add(th.Muted("s: " + p.RecipeNote))
	}
	add("", th.Muted("r run it now · c copy the command · s save the plugins as a recipe · esc back"))
	return lines
}
