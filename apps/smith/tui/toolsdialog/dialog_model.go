package toolsdialog

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/toolinstall"
	"github.com/monosmith-app/monosmith/apps/smith/tui/dialog"
)

// phase is the step the dialog is on.
type phase int

const (
	// phasePick: how to get the current tool.
	phasePick phase = iota
	// phaseAdmin: the command needs administrator rights; who runs it?
	phaseAdmin
	// phaseRun: the install streams its output.
	phaseRun
	// phaseDone: every tool has a result.
	phaseDone
)

// maxOutput bounds the lines kept from a streaming install.
const maxOutput = 400

// failedLines is how many lines of output a failed install keeps in its result.
const failedLines = 6

// adminAnswers are the three answers of the administrator question, in order.
var adminAnswers = []string{"I will run it myself", "Let smith run it", "Skip it"}

// Model is the dialog. It holds the plans of the missing tools, asks about one
// at a time, and keeps a result for each. It starts nothing: a key that decides
// an install returns an Outcome for the host to carry out.
type Model struct {
	plans   []toolinstall.ToolPlan
	at      int
	cursor  int
	ph      phase
	opt     toolinstall.Option
	out     []string
	results []Result
	rect    dialog.Rect
}

// New builds the dialog on the plans of the missing tools
// (smith tools install --dry-run --json).
//
// Parameters: plans are the tools that are missing or outdated, with their options.
// Returns: the dialog on the first tool, or on the result when there is none.
func New(plans []toolinstall.ToolPlan) Model {
	m := Model{plans: plans}
	if len(plans) == 0 {
		m.ph = phaseDone
		return m
	}
	m.cursor = defaultIndex(plans[0])
	return m
}

// Place sizes the dialog for a screen.
func (m Model) Place(sw, sh int) Model {
	w := min(max(sw-6, 40), 100)
	h := min(max(sh-2, 12), 30)
	m.rect = dialog.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/2, 0), W: w, H: h}
	return m
}

// Rect is where the dialog sits after Place.
func (m Model) Rect() dialog.Rect { return m.rect }

// Running is true while an install streams.
func (m Model) Running() bool { return m.ph == phaseRun }

// Results are the results so far, one per tool that was decided.
func (m Model) Results() []Result { return slices.Clone(m.results) }

// defaultIndex is the position of the marked default option of a plan.
func defaultIndex(p toolinstall.ToolPlan) int {
	for i, o := range p.Options {
		if o.Value == p.Default {
			return i
		}
	}
	return 0
}

// current is the tool being asked about.
func (m Model) current() toolinstall.ToolPlan { return m.plans[m.at] }

// Key handles one key, by the name Bubble Tea gives it ("up", "enter", "esc").
//
// Parameters: name is the key name.
// Returns: the new dialog and what the host must do.
func (m Model) Key(name string) (Model, Outcome) {
	switch m.ph {
	case phasePick:
		return m.pickKey(name)
	case phaseAdmin:
		return m.adminKey(name)
	case phaseRun:
		if name == "esc" || name == "ctrl+c" {
			return m, Outcome{Stop: true}
		}
		return m, Outcome{}
	}
	if name == "enter" || name == "esc" || name == "q" || name == "ctrl+c" {
		return m, Outcome{Done: true}
	}
	return m, Outcome{}
}

// move changes the cursor inside a list of n rows.
func (m Model) move(name string, n int) Model {
	switch name {
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, n-1)
	}
	return m
}

// pickKey answers the question "how do you want to get this tool".
func (m Model) pickKey(name string) (Model, Outcome) {
	plan := m.current()
	switch name {
	case "esc", "ctrl+c", "q":
		// Cancelling asks nothing more and runs nothing.
		return m, Outcome{Done: true}
	case "enter":
		opt := plan.Options[m.cursor]
		switch opt.Kind {
		case toolinstall.OptionInstall:
			if opt.Admin {
				m.ph, m.opt, m.cursor = phaseAdmin, opt, 0
				return m, Outcome{}
			}
			return m.start(opt, false)
		case toolinstall.OptionPage:
			return m.record("Go to " + plan.Download + ", install " + plan.Title + ", then run smith tools in a new terminal."), Outcome{}
		case toolinstall.OptionMyself:
			return m.record("Left to you. " + strings.TrimSpace(strings.ReplaceAll(toolinstall.RenderWays(plan), "\n", " "))), Outcome{}
		default:
			return m.record("Skipped. " + plan.Title + ": " + plan.Consequence + "."), Outcome{}
		}
	}
	return m.move(name, len(plan.Options)), Outcome{}
}

// adminKey answers the second question: the command needs administrator rights.
func (m Model) adminKey(name string) (Model, Outcome) {
	plan := m.current()
	command := strings.Join(m.opt.Command, " ")
	switch name {
	case "esc":
		// Back to the first question: nothing was decided.
		m.ph, m.cursor = phasePick, defaultIndex(plan)
		return m, Outcome{}
	case "ctrl+c":
		return m, Outcome{Done: true}
	case "enter":
		switch m.cursor {
		case 0:
			return m.record("Left to you. Run this, then open a new terminal and run smith tools: " + command), Outcome{}
		case 2:
			return m.record("Skipped. " + plan.Title + ": " + plan.Consequence + "."), Outcome{}
		}
		return m.start(m.opt, true)
	}
	return m.move(name, len(adminAnswers)), Outcome{}
}

// argv is the smith command that installs the tool with the option.
func argv(plan toolinstall.ToolPlan, opt toolinstall.Option, withAdmin bool) []string {
	a := []string{"tools", "install", plan.Name, "--manager", opt.Manager}
	if withAdmin {
		a = append(a, "--with-admin")
	}
	return a
}

// start carries out an install. A sudo command asks for a password on the
// terminal, which a stream cannot give, so it is handed over to the terminal;
// every other command streams here.
func (m Model) start(opt toolinstall.Option, withAdmin bool) (Model, Outcome) {
	plan := m.current()
	cmd := argv(plan, opt, withAdmin)
	if len(opt.Command) > 0 && opt.Command[0] == "sudo" {
		m = m.record("Handed over to the terminal: " + strings.Join(opt.Command, " "))
		return m, Outcome{Handover: cmd, Done: true}
	}
	m.ph, m.opt, m.out = phaseRun, opt, nil
	return m, Outcome{Run: cmd}
}

// record keeps the result of the current tool and goes to the next one.
func (m Model) record(line string) Model {
	plan := m.current()
	m.results = append(slices.Clone(m.results), Result{Name: plan.Name, Title: plan.Title, Line: line})
	return m.next()
}

// next asks about the next tool, or shows the results after the last one.
func (m Model) next() Model {
	m.at++
	m.out = nil
	if m.at >= len(m.plans) {
		m.ph = phaseDone
		return m
	}
	m.ph, m.cursor = phasePick, defaultIndex(m.plans[m.at])
	return m
}

// Output adds lines of the install that streams.
//
// Parameters: lines are the new lines of output.
func (m Model) Output(lines []string) Model {
	m.out = append(slices.Clone(m.out), lines...)
	if len(m.out) > maxOutput {
		m.out = m.out[len(m.out)-maxOutput:]
	}
	return m
}

// Finished ends the install that streamed and goes to the next tool. An exit
// code of 0 means the tool is installed and found; anything else means the
// install did not finish or the tool is not on the PATH of this terminal yet.
//
// Parameters: code is the exit code; cancelled is true when the person stopped it.
func (m Model) Finished(code int, cancelled bool) Model {
	if m.ph != phaseRun {
		return m
	}
	plan := m.current()
	r := Result{Name: plan.Name, Title: plan.Title}
	switch {
	case cancelled:
		r.Line, r.Failed = "Stopped. Check with smith tools whether "+plan.Title+" is there.", true
	case code == 0:
		r.Line = "Installed."
	default:
		r.Line, r.Failed = "Not found yet. The install may have failed, or "+plan.Title+" needs a new terminal (the PATH updates there). Then run smith tools. Official page: "+plan.Download, true
	}
	if r.Failed {
		r.Output = slices.Clone(m.out[max(len(m.out)-failedLines, 0):])
	}
	m.results = append(slices.Clone(m.results), r)
	return m.next()
}
