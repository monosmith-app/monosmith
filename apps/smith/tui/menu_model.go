package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	hintStyle   = lipgloss.NewStyle().Faint(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// Model is the menu and its question steps.
type Model struct {
	header  string
	actions []Action
	action  *Action
	step    int
	cursor  int
	marked  map[int]bool
	input   textinput.Model
	answers map[string][]string
	err     string
	// Argv is the chosen command, empty when the user quit.
	Argv []string
	// done is set when the menu has finished: a command chosen, or quit.
	done bool
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	m.done = true
	return m, tea.Quit
}

// NewModel builds the menu.
func NewModel(header string, actions []Action) Model {
	in := textinput.New()
	in.CharLimit = 50
	return Model{header: header, actions: actions, input: in, marked: map[int]bool{}}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

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
	return len(m.actions)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m.quit()
	case "esc":
		if m.action == nil {
			return m.quit()
		}
		m.action, m.step, m.cursor, m.err = nil, 0, 0, ""
		return m, nil
	case "q":
		if m.action == nil {
			return m.quit()
		}
		m.action, m.step, m.cursor = nil, 0, 0
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < m.optionCount()-1 {
			m.cursor++
		}
		return m, nil
	case " ":
		if step != nil && step.Multi {
			m.marked[m.cursor] = !m.marked[m.cursor]
			return m, nil
		}
	case "enter":
		return m.confirm()
	}
	return m, nil
}

func (m Model) confirm() (tea.Model, tea.Cmd) {
	if m.action == nil {
		a := m.actions[m.cursor]
		m.action, m.step, m.answers, m.cursor, m.marked = &a, 0, map[string][]string{}, 0, map[int]bool{}
		return m.advance()
	}
	step := m.current()
	switch {
	case step.Options == nil:
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			m.err = "type a value, or esc to go back"
			return m, nil
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
			return m, nil
		}
		m.answers[step.Key] = picked
	default:
		m.answers[step.Key] = []string{step.Options[m.cursor].Value}
	}
	m.step++
	m.cursor, m.marked, m.err = 0, map[int]bool{}, ""
	return m.advance()
}

func (m Model) advance() (tea.Model, tea.Cmd) {
	step := m.current()
	if step == nil {
		m.Argv = m.action.Argv(m.answers)
		return m.quit()
	}
	if step.Multi {
		for i, o := range step.Options {
			m.marked[i] = o.Hint == "enabled" || step.Key == "tasks"
		}
	}
	if step.Options == nil {
		m.input.Placeholder = step.Placeholder
		m.input.Focus()
		return m, textinput.Blink
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	return titleStyle.Render(m.header) + "\n\n" + m.body()
}

// body is the menu without its header line, for the shell.
func (m Model) body() string {
	var b strings.Builder
	step := m.current()
	if m.action != nil && step == nil {
		// The last answer was given and the program is quitting; Bubble Tea
		// still renders a final frame, with no step left to draw.
		return ""
	}
	switch {
	case m.action == nil:
		for i, a := range m.actions {
			b.WriteString(line(i == m.cursor, "", a.Title, a.Hint))
		}
		b.WriteString(hintStyle.Render("\n↑/↓ choose · enter select · q quit") + "\n")
	case step.Options == nil:
		fmt.Fprintf(&b, "%s › %s\n\n%s\n", m.action.Title, step.Prompt, m.input.View())
		b.WriteString(hintStyle.Render("\nenter confirm · esc back") + "\n")
	default:
		fmt.Fprintf(&b, "%s › %s\n\n", m.action.Title, step.Prompt)
		for i, o := range step.Options {
			box := ""
			if step.Multi {
				box = "[ ] "
				if m.marked[i] {
					box = "[x] "
				}
			}
			b.WriteString(line(i == m.cursor, box, o.Label, o.Hint))
		}
		help := "↑/↓ choose · enter select · esc back"
		if step.Multi {
			help = "↑/↓ move · space toggle · enter confirm · esc back"
		}
		b.WriteString(hintStyle.Render("\n"+help) + "\n")
	}
	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err) + "\n")
	}
	return b.String()
}

func line(selected bool, box, label, hint string) string {
	cursor := "  "
	if selected {
		cursor = cursorStyle.Render("› ")
		label = cursorStyle.Render(label)
	}
	if hint != "" {
		hint = "  " + hintStyle.Render(hint)
	}
	return cursor + box + label + hint + "\n"
}
