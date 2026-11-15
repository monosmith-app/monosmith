package graphview

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	markedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	headingStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	faintStyle    = lipgloss.NewStyle().Faint(true)
	panelStyle    = lipgloss.NewStyle().PaddingRight(4)
)

// Browser is `smith graph --format tui`: the projects on the left; on the
// right the selected one, what it depends on and what depends on it. The
// same nodes the other formats render.
type Browser struct {
	nodes      []Node
	index      map[string]int
	dependents map[string][]string
	cursor     int
	// back is where → came from, so ← returns.
	back []int
}

// NewBrowser builds the browser over the nodes, sorted by name.
func NewBrowser(nodes []Node) Browser {
	b := Browser{nodes: nodes, index: map[string]int{}, dependents: map[string][]string{}}
	for i, n := range nodes {
		b.index[n.Name] = i
	}
	for _, n := range nodes {
		for _, d := range n.Deps {
			b.dependents[d] = append(b.dependents[d], n.Name)
		}
	}
	for _, list := range b.dependents {
		sort.Strings(list)
	}
	return b
}

// Init implements tea.Model.
func (b Browser) Init() tea.Cmd { return nil }

// Selected is the project under the cursor.
func (b Browser) Selected() string {
	if len(b.nodes) == 0 {
		return ""
	}
	return b.nodes[b.cursor].Name
}

// Update implements tea.Model: ↑/↓ (k/j) move, → (l, enter) goes to the first
// project the selected one depends on, ← (h) comes back, d goes to the first
// project that depends on it, q quits.
func (b Browser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok || len(b.nodes) == 0 {
		if ok && (key.String() == "q" || key.String() == "ctrl+c" || key.String() == "esc") {
			return b, tea.Quit
		}
		return b, nil
	}
	jump := func(name string) {
		if i, ok := b.index[name]; ok {
			b.back = append(b.back, b.cursor)
			b.cursor = i
		}
	}
	switch key.String() {
	case "q", "ctrl+c", "esc":
		return b, tea.Quit
	case "up", "k":
		b.cursor = (b.cursor - 1 + len(b.nodes)) % len(b.nodes)
	case "down", "j":
		b.cursor = (b.cursor + 1) % len(b.nodes)
	case "right", "l", "enter":
		if deps := b.nodes[b.cursor].Deps; len(deps) > 0 {
			jump(deps[0])
		}
	case "d":
		if users := b.dependents[b.Selected()]; len(users) > 0 {
			jump(users[0])
		}
	case "left", "h":
		if n := len(b.back); n > 0 {
			b.cursor, b.back = b.back[n-1], b.back[:n-1]
		}
	}
	return b, nil
}

// View implements tea.Model.
func (b Browser) View() string {
	if len(b.nodes) == 0 {
		return "No projects yet. Add one: smith add app web --lang node\n"
	}
	var left strings.Builder
	left.WriteString(headingStyle.Render("Projects") + "\n")
	for i, n := range b.nodes {
		line := "  " + n.Name
		switch {
		case i == b.cursor:
			line = selectedStyle.Render("› " + n.Name)
		case n.Marked:
			line = markedStyle.Render(line)
		}
		left.WriteString(line + "\n")
	}
	n := b.nodes[b.cursor]
	var right strings.Builder
	fmt.Fprintf(&right, "%s\n%s · %s · %s\n", headingStyle.Render(n.Name), n.Kind, n.Language, n.Root)
	if n.Stack != "" {
		fmt.Fprintf(&right, "stack %s\n", n.Stack)
	}
	fmt.Fprintf(&right, "tasks %s\n\n", strings.Join(n.Tasks, ", "))
	list := func(title string, names []string) {
		right.WriteString(headingStyle.Render(title) + "\n")
		if len(names) == 0 {
			right.WriteString(faintStyle.Render("  none") + "\n")
		}
		for _, name := range names {
			right.WriteString("  " + name + "\n")
		}
		right.WriteString("\n")
	}
	list("Projects it needs", n.Deps)
	list("Projects that use it", b.dependents[n.Name])
	help := faintStyle.Render("↑↓ move · → projects it needs · d projects that use it · ← back · q quit")
	return lipgloss.JoinHorizontal(lipgloss.Top, panelStyle.Render(left.String()), right.String()) + "\n" + help + "\n"
}
