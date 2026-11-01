package dependencies

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	groupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	faintStyle  = lipgloss.NewStyle().Faint(true)
)

// picker lets the user choose updates: patch and minor ones start selected,
// major ones (which may break) do not.
type picker struct {
	updates  []Update
	selected []bool
	cursor   int
	done     bool
}

func newPicker(updates []Update) picker {
	p := picker{updates: updates, selected: make([]bool, len(updates))}
	for i, u := range updates {
		p.selected[i] = u.Level != "major"
	}
	return p
}

// Chosen are the selected updates once confirmed, nil when cancelled.
func (p picker) Chosen() []Update {
	if !p.done {
		return nil
	}
	var out []Update
	for i, u := range p.updates {
		if p.selected[i] {
			out = append(out, u)
		}
	}
	return out
}

// Init implements tea.Model.
func (p picker) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (p picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return p, nil
	}
	switch key.String() {
	case "ctrl+c", "esc", "q":
		return p, tea.Quit
	case "up", "k":
		p.cursor = (p.cursor - 1 + len(p.updates)) % len(p.updates)
	case "down", "j":
		p.cursor = (p.cursor + 1) % len(p.updates)
	case " ":
		p.selected[p.cursor] = !p.selected[p.cursor]
	case "a":
		all := true
		for _, s := range p.selected {
			all = all && s
		}
		for i := range p.selected {
			p.selected[i] = !all
		}
	case "enter":
		p.done = true
		return p, tea.Quit
	}
	return p, nil
}

// View implements tea.Model.
func (p picker) View() string {
	var b strings.Builder
	current := ""
	for i, u := range p.updates {
		if u.Group() != current {
			current = u.Group()
			b.WriteString(groupStyle.Render(current) + "\n")
		}
		box := "[ ]"
		if p.selected[i] {
			box = "[x]"
		}
		row := fmt.Sprintf("%s %-28s %s → %s", box, u.Name, u.Current, u.Latest)
		if i == p.cursor {
			row = cursorStyle.Render("› " + row)
		} else {
			row = "  " + row
		}
		b.WriteString(row + "  " + faintStyle.Render(strings.Join(u.Projects(), ", ")) + "\n")
	}
	b.WriteString(faintStyle.Render("\n↑/↓ move · space toggle · a all/none · enter update · esc cancel") + "\n")
	return b.String()
}
