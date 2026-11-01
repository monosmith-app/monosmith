package guide

import (
	"slices"
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
	// filter is what was typed to narrow a Filter step's options.
	filter string
	// defaults are the options a step starts on, by step key.
	defaults map[string]string
	// defaultWhy is the line that says why a step starts on its default.
	defaultWhy map[string]string
	// Argv is the chosen command, empty until the questions are answered.
	Argv []string
	// Wizard names the wizard an action handed itself to (Action.Wizard).
	Wizard string
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

// WithAction puts an action in the menu, in the place of the one with the same
// title (or at the end): the shell uses it for a form whose questions come from
// what it just read, such as the choices of the plugins.
func (m Model) WithAction(a Action) Model {
	actions := slices.Clone(m.actions)
	for i := range actions {
		if actions[i].Title == a.Title {
			actions[i] = a
			m.actions = actions
			return m
		}
	}
	m.actions = append(actions, a)
	return m
}

// ActionTitled returns the action with a title; ok is false when there is none.
func (m Model) ActionTitled(title string) (Action, bool) {
	for _, a := range m.actions {
		if a.Title == title {
			return a, true
		}
	}
	return Action{}, false
}

// Reset returns the guide to its first screen, to be used again.
func (m Model) Reset() Model {
	r := NewModel(m.actions)
	r.defaults = m.defaults
	r.defaultWhy = m.defaultWhy
	return r
}

// WithDefault makes the step with a key start on the option with a value, for
// the questions whose best answer depends on the workspace (the language the
// workspace mostly uses).
func (m Model) WithDefault(key, value string) Model {
	d := make(map[string]string, len(m.defaults)+1)
	for k, v := range m.defaults {
		d[k] = v
	}
	d[key] = value
	m.defaults = d
	return m
}

// WithDefaultWhy sets the line a step shows under its question to say why it
// starts on its default ("Preselected: ..."); the line is empty for no reason.
func (m Model) WithDefaultWhy(key, why string) Model {
	d := make(map[string]string, len(m.defaultWhy)+1)
	for k, v := range m.defaultWhy {
		d[k] = v
	}
	d[key] = why
	m.defaultWhy = d
	return m
}

// Editing reports whether a text field has the keyboard, so letters that are
// shortcuts elsewhere must go to it.
func (m Model) Editing() bool {
	s := m.current()
	return s != nil && (s.Options == nil || s.Filter)
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

// options are the answers on offer: all of them, or for a Filter step those
// that match what was typed.
func (m Model) options() []Option {
	s := m.current()
	if s == nil {
		return nil
	}
	if !s.Filter || m.filter == "" {
		return s.Options
	}
	var out []Option
	for _, o := range s.Options {
		if MatchesFilter(o, m.filter) {
			out = append(out, o)
		}
	}
	return out
}

func (m Model) optionCount() int {
	if s := m.current(); s != nil {
		return len(m.options())
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
	if step != nil && step.Filter && step.Options != nil {
		switch key.Type {
		case tea.KeyRunes, tea.KeySpace:
			m.filter += string(key.Runes)
			m.cursor = 0
			return m, nil
		case tea.KeyBackspace:
			if r := []rune(m.filter); len(r) > 0 {
				m.filter, m.cursor = string(r[:len(r)-1]), 0
			}
			return m, nil
		case tea.KeyEsc:
			if m.filter != "" {
				m.filter, m.cursor = "", 0
				return m, nil
			}
		}
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
		if a.Wizard != "" {
			m.Wizard, m.done = a.Wizard, true
			return m
		}
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
		if problem := m.problem(step, value); problem != "" {
			m.err = problem
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
		if len(picked) == 0 && !step.Optional {
			m.err = "toggle at least one with space"
			return m
		}
		m.answers[step.Key] = picked
	default:
		opts := m.options()
		if len(opts) == 0 {
			m.err = "nothing matches \"" + m.filter + "\": backspace to widen the list"
			return m
		}
		m.answers[step.Key] = []string{opts[min(m.cursor, len(opts)-1)].Value}
	}
	m.step++
	m.cursor, m.marked, m.err, m.filter = 0, map[int]bool{}, "", ""
	return m.advance()
}

func (m Model) advance() Model {
	for m.action != nil {
		if s := m.current(); s == nil || s.Skip == nil || !s.Skip(m.answers) {
			break
		}
		m.step++
	}
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
		if step.PlaceholderFor != nil {
			m.input.Placeholder = step.PlaceholderFor(m.answers)
		}
		m.input.Focus()
	} else if want := m.defaults[step.Key]; want != "" {
		for i, o := range step.Options {
			if o.Value == want {
				m.cursor = i
			}
		}
	}
	return m
}

// problem is what is wrong with a typed answer, or "" when it is fine.
func (m Model) problem(step *Step, value string) string {
	if step.Check == nil {
		return ""
	}
	return step.Check(m.answers, value)
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
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt, "", m.input.View(), "")
		if v := strings.TrimSpace(m.input.Value()); v == "" && m.input.Placeholder != "" && !step.Optional {
			// The grey text is an example, not a default Enter would accept.
			lines = append(lines, th.Muted("The grey text is an example: type your own, enter does not take it."))
		} else if v != "" {
			if problem := m.problem(step, v); problem != "" {
				lines = append(lines, th.Failure(problem))
			} else if step.Check != nil {
				lines = append(lines, th.Success(th.Icons.Passed+" that name works"))
			}
		}
		lines = append(lines, th.Muted("enter confirm · esc back"))
	default:
		lines = append(lines, th.Title(m.action.Title)+" › "+step.Prompt)
		if why := m.defaultWhy[step.Key]; why != "" && m.defaults[step.Key] != "" {
			lines = append(lines, th.Muted(why))
		}
		lines = append(lines, "")
		if step.Filter {
			if m.filter == "" {
				lines = append(lines, th.Muted("Type to narrow the list"), "")
			} else {
				lines = append(lines, th.Muted("Showing matches for ")+th.Bold(m.filter)+th.Muted(" (esc clears it)"), "")
			}
		}
		for i, o := range m.options() {
			box := ""
			if step.Multi {
				box = "[ ] "
				if m.marked[i] {
					box = "[x] "
				}
			}
			add(i, Row(th, i == m.cursor, box, o.Label, o.Hint))
		}
		if step.Filter && len(m.options()) == 0 {
			lines = append(lines, th.Muted("  nothing matches"))
		}
		help := "up/down choose · enter select · esc back"
		if step.Filter {
			help = "type to filter · up/down choose · enter select · esc back"
		}
		if step.Multi {
			help = "up/down move · space or click toggles · enter confirm · esc back"
			if step.Optional {
				help = "up/down move · space or click toggles · enter confirms (nothing ticked skips this) · esc back"
			}
		}
		lines = append(lines, "", th.Muted(help))
	}
	if m.err != "" && !shown(lines, m.err) {
		lines = append(lines, th.Failure(m.err))
	}
	for i := range lines {
		lines[i] = pane.Fit(lines[i], w)
	}
	return lines, optionAt
}

// shown reports whether a line of the screen already says text, so the same
// hint is not printed twice.
func shown(lines []string, text string) bool {
	for _, l := range lines {
		if strings.Contains(pane.Strip(l), text) {
			return true
		}
	}
	return false
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
