package composer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/compose"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Planner reads the plan of a request; it is what the host gives the screen
// when the screen runs alone (smith compose in a terminal).
type Planner func(PlanRequest) (compose.Plan, error)

// Result is how a standalone run ended.
type Result struct {
	// Run is the smith command line the person chose to run (without "smith");
	// empty when they only looked or copied.
	Run []string
	// Copied is the command the person copied, shown after the screen closes.
	Copied string
}

// Program is the screen as a program of its own, for `smith compose`.
type Program struct {
	m      Model
	plan   Planner
	result *Result
	th     theme.Theme
}

// NewProgram wraps the screen. result is filled when the program ends.
//
// Parameters: m is the screen; plan reads a plan; result receives the end.
func NewProgram(m Model, plan Planner, result *Result) Program {
	return Program{m: m, plan: plan, result: result, th: theme.Default()}
}

// Init implements tea.Model.
func (p Program) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (p Program) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.m = p.m.Place(msg.Width, msg.Height)
	case tea.KeyMsg:
		m, out := p.m.Key(msg)
		return p.after(m, out)
	case tea.MouseMsg:
		// The same mouse as in the shell: a press ticks or picks, the pointer
		// shows the row hover, the wheel scrolls.
		switch {
		case msg.Action == tea.MouseActionMotion:
			p.m = p.m.Hover(msg.X, msg.Y)
		case msg.Action != tea.MouseActionPress:
		case msg.Button == tea.MouseButtonWheelUp:
			p.m = p.m.Wheel(-3)
		case msg.Button == tea.MouseButtonWheelDown:
			p.m = p.m.Wheel(3)
		case msg.Button == tea.MouseButtonLeft:
			m, out := p.m.Click(msg.X, msg.Y)
			return p.after(m, out)
		}
	}
	return p, nil
}

// after keeps the screen and does what its outcome asks.
func (p Program) after(m Model, out Outcome) (tea.Model, tea.Cmd) {
	p.m = m
	if out.Plan != nil {
		p.m = p.m.WithPlan(p.plan(*out.Plan))
	}
	if out.Copy != "" {
		p.result.Copied = out.Copy
	}
	if out.Run != nil {
		p.result.Run = out.Run
	}
	if out.Done {
		return p, tea.Quit
	}
	return p, nil
}

// View implements tea.Model.
func (p Program) View() string {
	r := p.m.Rect()
	return strings.Repeat("\n", r.Y) + indent(p.m.Lines(p.th), r.X)
}

// indent shifts every line right by n cells.
func indent(lines []string, n int) string {
	pad := strings.Repeat(" ", n)
	for i := range lines {
		lines[i] = pad + lines[i]
	}
	return strings.Join(lines, "\n")
}
