package guide

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func keys(m tea.Model, ks ...string) tea.Model {
	for _, k := range ks {
		m, _ = m.Update(keyMsg(k))
	}
	return m
}

func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func menu() Model {
	return NewModel(Actions([]string{"go", "node"}, []string{"go", "node", "python"}, []string{"go", "node"})[1:])
}

func TestAddAProject(t *testing.T) {
	// Add a project (2nd entry) → package (3rd) → node (2nd) → name.
	m := keys(menu(), "down", "enter", "down", "down", "enter", "down", "enter", "s", "d", "k", "enter")
	got := strings.Join(m.(Model).Argv, " ")
	if got != "add package sdk --lang node" {
		t.Fatalf("argv %q", got)
	}
}

func TestRunTasksDefaultsToEveryTaskAndToggles(t *testing.T) {
	// Run tasks → untoggle lint → confirm → affected.
	m := keys(menu(), "enter", "space", "enter", "down", "enter")
	if got := strings.Join(m.(Model).Argv, " "); got != "run typecheck test build --affected" {
		t.Fatalf("argv %q", got)
	}
}

func TestPluginsStartFromTheirCurrentState(t *testing.T) {
	// Manage plugins (3rd) → python is off; toggle it on.
	m := keys(menu(), "down", "down", "enter")
	view := m.View()
	if !strings.Contains(view, "[x] go") || !strings.Contains(view, "[ ] python") {
		t.Fatalf("view:\n%s", view)
	}
	m = keys(m, "down", "down", "space", "enter")
	if got := strings.Join(m.(Model).Argv, " "); got != "plugin set go node python" {
		t.Fatalf("argv %q", got)
	}
}

func TestValidationAndBack(t *testing.T) {
	m := keys(menu(), "down", "enter", "enter", "enter", "enter")
	if !strings.Contains(m.View(), "type a value") {
		t.Fatalf("an empty name must be refused:\n%s", m.View())
	}
	m = keys(m, "esc")
	if m.(Model).action != nil || len(m.(Model).Argv) != 0 {
		t.Fatal("esc returns to the menu")
	}
	m = keys(m, "q")
	if len(m.(Model).Argv) != 0 {
		t.Fatal("quitting runs nothing")
	}
}

// Bubble Tea draws one last frame after the final answer, when no step is
// left; that frame must render (it once dereferenced a nil step).
func TestTheFrameAfterTheFinalAnswerRenders(t *testing.T) {
	m := keys(menu(), "down", "enter", "down", "down", "enter", "down", "enter", "s", "d", "k", "enter")
	if !m.(Model).Done() {
		t.Fatal("the guide should be done")
	}
	if view := m.View(); view == "" {
		t.Fatal("empty final frame")
	}
	for _, w := range []int{0, 1, 5, 80} {
		lines, _ := m.(Model).Body(theme.Default(), w)
		_ = lines
	}
}

func TestPickStartsTheNamedAction(t *testing.T) {
	m, ok := menu().Pick("Manage plugins")
	if !ok || !strings.Contains(m.View(), "Enabled plugins") {
		t.Fatalf("%v\n%s", ok, m.View())
	}
	if _, ok := menu().Pick("No such action"); ok {
		t.Fatal("unknown")
	}
}

func TestClickingAnOptionMovesTheCursorAndTogglesAMultipleChoice(t *testing.T) {
	m, _ := menu().Pick("Run tasks")
	m = m.SelectRow(1)
	if view := m.View(); !strings.Contains(view, "[ ] typecheck") {
		t.Fatalf("a click toggles:\n%s", view)
	}
}

func TestALongEntryWrapsAndEveryLineOfItIsClickable(t *testing.T) {
	long := []Action{{Title: "Adopt this repository", Hint: "smith adopt: bring an existing repository under smith without changing how it builds"}, {Title: "Other"}}
	lines, optionAt := NewModel(long).Body(theme.Default(), 40)
	if len(lines) < 4 {
		t.Fatalf("the hint was cut, not wrapped:\n%s", strings.Join(lines, "\n"))
	}
	text := strings.Join(lines, " ")
	if !strings.Contains(text, "builds") {
		t.Fatalf("the end of the hint is missing:\n%s", strings.Join(lines, "\n"))
	}
	if optionAt[0] != 0 || optionAt[1] != 0 || optionAt[len(lines)-3] != 1 {
		t.Errorf("rows do not map to their option: %v", optionAt)
	}
}
