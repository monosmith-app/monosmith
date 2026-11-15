package devservers

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	faintStyle  = lipgloss.NewStyle().Faint(true)
)

// chooser picks the projects to start; none start selected.
type chooser struct {
	names    []string
	selected []bool
	cursor   int
	done     bool
}

func newChooser(names []string) chooser {
	return chooser{names: names, selected: make([]bool, len(names))}
}

// Chosen are the picked projects once confirmed, nil when cancelled.
func (c chooser) Chosen() []string {
	if !c.done {
		return nil
	}
	var out []string
	for i, n := range c.names {
		if c.selected[i] {
			out = append(out, n)
		}
	}
	return out
}

// Init implements tea.Model.
func (c chooser) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (c chooser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return c, nil
	}
	switch key.String() {
	case "ctrl+c", "esc", "q":
		return c, tea.Quit
	case "up", "k":
		c.cursor = (c.cursor - 1 + len(c.names)) % len(c.names)
	case "down", "j":
		c.cursor = (c.cursor + 1) % len(c.names)
	case " ":
		c.selected[c.cursor] = !c.selected[c.cursor]
	case "enter":
		c.done = true
		return c, tea.Quit
	}
	return c, nil
}

// View implements tea.Model.
func (c chooser) View() string {
	var b strings.Builder
	b.WriteString("Start which projects?\n\n")
	for i, n := range c.names {
		box := "[ ] "
		if c.selected[i] {
			box = "[x] "
		}
		row := box + n
		if i == c.cursor {
			row = cursorStyle.Render("› " + row)
		} else {
			row = "  " + row
		}
		b.WriteString(row + "\n")
	}
	b.WriteString(faintStyle.Render("\n↑/↓ move · space toggle · enter start · esc cancel") + "\n")
	return b.String()
}
