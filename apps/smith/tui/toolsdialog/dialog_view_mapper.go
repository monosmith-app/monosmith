package toolsdialog

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Title is the frame title of the dialog.
func (m Model) Title() string {
	switch m.ph {
	case phaseDone:
		return words.Say("toolsdialog.title.done")
	case phaseRun:
		return words.Say("toolsdialog.title.running", "current", m.at+1, "total", len(m.plans))
	}
	return words.Say("toolsdialog.title.asking", "current", m.at+1, "total", len(m.plans))
}

// Lines draws the dialog, framed, at the size Place set.
func (m Model) Lines(th theme.Theme) []string {
	r := m.rect
	if r.W < 24 || r.H < 8 {
		return nil
	}
	inner, height := r.W-4, r.H-2
	var body []string
	switch m.ph {
	case phasePick:
		body, _ = m.pickBody(th, inner)
	case phaseAdmin:
		body, _ = m.adminBody(th, inner)
	case phaseRun:
		body = m.runBody(th, inner, height)
	default:
		body = m.doneBody(th, inner)
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

// headLines say what the tool is and what stops working without it.
func headLines(th theme.Theme, p toolinstall.ToolPlan, inner int) []string {
	state := words.Say("toolsdialog.state.missing")
	if p.State == toolinstall.Outdated {
		state = words.Say("toolsdialog.state.old", "tool", strings.TrimSpace(p.Program+" "+p.Version))
		if p.Exact {
			state = p.Problem()
		}
	}
	lines := []string{th.Bold(words.Say("toolsdialog.head", "title", p.Title, "state", state))}
	lines = append(lines, pane.Wrap(th.Muted(words.Say("toolsdialog.why", "why", p.Why, "consequence", p.Consequence)), inner, 3)...)
	return append(lines, "")
}

// pickBody is the first question: the options, the default marked, and the
// exact command or address of the option under the cursor. first is the line
// of the first option, for a click.
func (m Model) pickBody(th theme.Theme, inner int) (lines []string, first int) {
	p := m.current()
	lines = headLines(th, p, inner)
	lines = append(lines, words.Say("toolsdialog.pick.question"))
	first = len(lines)
	for i, o := range p.Options {
		mark, label := "  ", o.Label
		if o.Value == p.Default {
			label += words.Say("toolsdialog.default")
		}
		if i == m.cursor {
			mark, label = th.Accent("> "), th.Bold(label)
		}
		lines = append(lines, mark+label)
	}
	o := p.Options[m.cursor]
	lines = append(lines, "")
	if o.Kind == toolinstall.OptionInstall {
		lines = append(lines, th.Muted(words.Say("toolsdialog.command.label"))+strings.Join(o.Command, " "))
		if o.Admin {
			lines = append(lines, pane.Wrap(th.Warning(words.Say("toolsdialog.admin.warning")), inner, 2)...)
		}
	} else {
		lines = append(lines, pane.Wrap(th.Muted(o.Hint), inner, 2)...)
	}
	if o.Note != "" {
		lines = append(lines, pane.Wrap(th.Muted(words.Say("toolsdialog.note", "note", o.Note)), inner, 2)...)
	}
	lines = append(lines, "")
	return append(lines, th.Muted(words.Say("toolsdialog.pick.keys")), th.Muted(words.Say("toolsdialog.click.hint"))), first
}

// adminBody is the second question: who runs a command that needs
// administrator rights. first is the line of the first answer, for a click.
func (m Model) adminBody(th theme.Theme, inner int) (lines []string, first int) {
	p := m.current()
	lines = headLines(th, p, inner)
	lines = append(lines, words.Say("toolsdialog.admin.needs"), "  "+th.Bold(strings.Join(m.opt.Command, " ")))
	how := words.Say("toolsdialog.admin.system")
	if len(m.opt.Command) > 0 && m.opt.Command[0] == "sudo" {
		how = words.Say("toolsdialog.admin.sudo")
	}
	lines = append(lines, pane.Wrap(th.Muted(how), inner, 2)...)
	lines = append(lines, "")
	first = len(lines)
	for i, a := range adminAnswers() {
		mark, label := "  ", a
		if i == 0 {
			label += words.Say("toolsdialog.default")
		}
		if i == m.cursor {
			mark, label = th.Accent("> "), th.Bold(label)
		}
		lines = append(lines, mark+label)
	}
	lines = append(lines, "")
	return append(lines, th.Muted(words.Say("toolsdialog.admin.keys")), th.Muted(words.Say("toolsdialog.click.hint"))), first
}

// runBody is the command and the end of its output, as it streams.
func (m Model) runBody(th theme.Theme, inner, height int) []string {
	lines := []string{th.Bold(words.Say("toolsdialog.run.label")) + strings.Join(m.opt.Command, " "), ""}
	room := max(height-len(lines)-2, 3)
	out := m.out
	if len(out) > room {
		out = out[len(out)-room:]
	}
	for _, l := range out {
		lines = append(lines, th.Muted(l))
	}
	for len(lines) < height-1 {
		lines = append(lines, "")
	}
	return append(lines, th.Muted(words.Say("toolsdialog.run.keys")))
}

// doneBody is one line per tool, and the way on.
func (m Model) doneBody(th theme.Theme, inner int) []string {
	if len(m.results) == 0 {
		return []string{words.Say("toolsdialog.done.none"), "", th.Muted(words.Say("toolsdialog.done.close"))}
	}
	var lines []string
	for _, r := range m.results {
		head := r.Title
		if r.Failed {
			head = th.Warning(r.Title)
		}
		lines = append(lines, th.Bold(head))
		lines = append(lines, pane.Wrap("  "+r.Line, inner, 2)...)
		for _, o := range r.Output {
			lines = append(lines, th.Muted("  | "+o))
		}
	}
	return append(lines, "", th.Muted(words.Say("toolsdialog.done.keys")))
}

// Click handles a left press on the list of options or answers: one press moves
// the cursor to the entry, a second press on the same cell (double is true)
// answers like Enter. A press elsewhere, or while an install runs, changes
// nothing.
//
// Parameters: th is the theme the dialog is drawn with; x and y are screen
// cells; double says the press is the second of a double click.
// Returns: the new dialog and what the host must do.
func (m Model) Click(th theme.Theme, x, y int, double bool) (Model, Outcome) {
	if !m.rect.Contains(x, y) {
		return m, Outcome{}
	}
	inner := m.rect.W - 4
	var first, n int
	switch m.ph {
	case phasePick:
		_, first = m.pickBody(th, inner)
		n = len(m.current().Options)
	case phaseAdmin:
		_, first = m.adminBody(th, inner)
		n = len(adminAnswers())
	default:
		return m, Outcome{}
	}
	i := y - m.rect.Y - 1 - first
	if i < 0 || i >= n || y-m.rect.Y-1 >= m.rect.H-2 {
		return m, Outcome{}
	}
	m.cursor = i
	if double {
		return m.Key("enter")
	}
	return m, Outcome{}
}
