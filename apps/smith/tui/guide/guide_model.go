package guide

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Model is the guided menu and its question steps: pick an action, answer
// its questions one at a time, and it ends as a smith command line.
type Model struct {
	actions []Action
	action  *Action
	step    int
	cursor  int
	marked  map[int]bool
	input   textinput.Model
	answers map[string][]string
	err     string
	// Argv is the chosen command, empty until the questions are answered.
	Argv []string
	// done is set when a command is chosen, or the user backed out of the top.
	done bool
}

// NewModel builds the guide.
func NewModel(actions []Action) Model {
	in := textinput.New()
	in.CharLimit = 50
	return Model{actions: actions, input: in, marked: map[int]bool{}}
}

// Done reports whether the guide finished: with Argv set, a command was
// chosen; with Argv empty, the user left the top level.
func (m Model) Done() bool { return m.done }

// Title is the heading of the chosen action ("Add a project"), or "" while
// the list of actions is showing.
func (m Model) Title() string {
	if m.action == nil {
		return ""
	}
	return m.action.Title
}

// Reset returns the guide to its first screen, to be used again.
func (m Model) Reset() Model { return NewModel(m.actions) }

// Editing reports whether a text field has the keyboard, so letters that are
// shortcuts elsewhere must go to it.
func (m Model) Editing() bool {
	s := m.current()
	return s != nil && s.Options == nil
}

// Pick starts the action with a title, as if chosen from the list; ok is
// false when there is no such action.
func (m Model) Pick(title string) (Model, bool) {
	for i, a := range m.actions {
		if a.Title == title {
			m = m.Reset()
			m.cursor = i
			return m.confirm(), true
		}
	}
	return m, false
}

func (m Model) current() *Step {
	if m.action == nil || m.step >= len(m.action.Steps) {
		return nil
	}
	return &m.action.Steps[m.step]
}

func (m Model) optionCount() int {
	if s := m.current(); s != nil {
		return len(s.Options)
	}
	if m.action != nil {
		return 0
	}
	return len(m.actions)
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model; only key messages matter.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m.Handle(msg) }

// Handle is Update returning the concrete model.
func (m Model) Handle(msg tea.Msg) (Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	step := m.current()
	text := step != nil && step.Options == nil
	if text && key.Type != tea.KeyEnter && key.Type != tea.KeyEsc && key.Type != tea.KeyCtrlC {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	switch key.String() {
	case "ctrl+c":
		m.done = true
	case "esc", "q":
		switch {
		case m.action != nil:
			m.action, m.step, m.cursor, m.err = nil, 0, 0, ""
		case key.String() == "esc":
			m.done = true
		}
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, max(m.optionCount()-1, 0))
	case " ":
		if step != nil && step.Multi {
			m.marked = cloneMarks(m.marked)
			m.marked[m.cursor] = !m.marked[m.cursor]
		}
	case "enter":
		return m.confirm(), nil
	}
	return m, nil
}

// SelectRow moves the cursor to option i (a mouse click); in a multiple
// choice it also toggles the option.
func (m Model) SelectRow(i int) Model {
	if i >= 0 && i < m.optionCount() {
		m.cursor = i
		if s := m.current(); s != nil && s.Multi {
			m.marked = cloneMarks(m.marked)
			m.marked[i] = !m.marked[i]
		}
	}
	return m
}

func cloneMarks(in map[int]bool) map[int]bool {
	out := make(map[int]bool, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (m Model) confirm() Model {
	if m.action == nil {
		if m.cursor < 0 || m.cursor >= len(m.actions) {
			return m
		}
		a := m.actions[m.cursor]
		m.action, m.step, m.answers, m.cursor, m.marked = &a, 0, map[string][]string{}, 0, map[int]bool{}
		return m.advance()
	}
	step := m.current()
	if step == nil {
		return m.advance()
	}
	switch {
	case step.Options == nil:
		value := strings.TrimSpace(m.input.Value())
		if value == "" && !step.Optional {
			m.err = "type a value, or esc to go back"
			return m
		}
		m.answers[step.Key] = []string{value}
		m.input.SetValue("")
	case step.Multi:
		var picked []string
		for i, o := range step.Options {
			if m.marked[i] {
				picked = append(picked, o.Value)
			}
		}
		if len(picked) == 0 {
			m.err = "toggle at least one with space"
			return m
		}
		m.answers[step.Key] = picked
	default:
		m.answers[step.Key] = []string{step.Options[min(m.cursor, len(step.Options)-1)].Value}
	}
	m.step++
	m.cursor, m.marked, m.err = 0, map[int]bool{}, ""
	return m.advance()
}

func (m Model) advance() Model {
	step := m.current()
	if step == nil {
		if m.action != nil && m.action.Argv != nil {
			m.Argv = m.action.Argv(m.answers)
		}
		m.done = true
		return m
	}
	if step.Multi {
		for i, o := range step.Options {
			m.marked[i] = o.Hint == "enabled" || step.Key == "tasks"
		}
	}
	if step.Options == nil {
		m.input.Placeholder = step.Placeholder
		m.input.Focus()
	}
	return m
}

// View implements tea.Model. It can be called in any state, including the
// last frame Bubble Tea draws after the final answer.
func (m Model) View() string {
	lines, _ := m.Body(theme.Default(), 80)
	return strings.Join(lines, "\n")
}

// Body draws the current screen in w cells. A long entry wraps onto further
// lines; optionAt maps every line an option takes to the option, for a click.
func (m Model) Body(th theme.Theme, w int) (lines []string, optionAt map[int]int) {
	optionAt = map[int]int{}
	add := func(i int, row string) {
		for _, l := range pane.Wrap(row, w, 4) {
			optionAt[len(lines)] = i
			lines = append(lines, l)
		}
	}
	step := m.current()
	switch {
	case m.action == nil:
		for i, a := range m.actions {
			add(i, Row(th, i == m.cursor, "", a.Title, a.Hint))
		}
		lines = append(lines, "", th.Muted("up/down choose · enter select · click an entry, double-click to select"))
	case step == nil:
		lines = append(lines, th.Muted("Done: "+m.action.Title))
	case step.Options == nil:
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt, "", m.input.View(), "",
			th.Muted("enter confirm · esc back"))
	default:
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt, "")
		for i, o := range step.Options {
			box := ""
			if step.Multi {
				box = "[ ] "
				if m.marked[i] {
					box = "[x] "
				}
			}
			add(i, Row(th, i == m.cursor, box, o.Label, o.Hint))
		}
		help := "up/down choose · enter select · esc back"
		if step.Multi {
			help = "up/down move · space or click toggles · enter confirm · esc back"
		}
		lines = append(lines, "", th.Muted(help))
	}
	if m.err != "" {
		lines = append(lines, th.Failure(m.err))
	}
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, optionAt
}

// Row draws one selectable line: a pointer, an optional checkbox, the label
// and a dimmed hint.
func Row(th theme.Theme, selected bool, box, label, hint string) string {
	cursor := "  "
	if selected {
		cursor = th.Accent(th.Icons.Pointer + " ")
		label = th.Accent(label)
	}
	if hint != "" {
		hint = "  " + th.Muted(hint)
	}
	return cursor + box + label + hint
}
