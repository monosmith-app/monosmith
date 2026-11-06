package guide

import (
	"strings"
	"testing"
)

func TestAddAProject(t *testing.T) {
	// Add a project (2nd entry) → package (3rd) → node (1st: the popular languages lead) → name.
	m := keys(menu(), "down", "enter", "down", "down", "enter", "enter", "s", "d", "k", "enter")
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
	// Esc goes back one question (docs/manual/tui.md), not out of the form.
	m = keys(m, "esc")
	if m.(Model).action == nil {
		t.Fatal("esc left the whole form instead of going back one question")
	}
	for i := 0; i < 10 && m.(Model).action != nil; i++ {
		m = keys(m, "esc")
	}
	if m.(Model).action != nil || len(m.(Model).Argv) != 0 {
		t.Fatal("esc on the first question returns to the menu")
	}
	m = keys(m, "q")
	if len(m.(Model).Argv) != 0 {
		t.Fatal("quitting runs nothing")
	}
}

// TestEscGoesBackOneQuestionAndKeepsTheAnswers keeps what was typed.
func TestEscGoesBackOneQuestionAndKeepsTheAnswers(t *testing.T) {
	// Add a project, app, node, name "sdk"; Esc returns to the name with it filled in.
	m := keys(menu(), "down", "enter", "enter", "enter", "s", "d", "k", "enter")
	if !m.(Model).Done() {
		t.Fatalf("the form should have ended:\n%s", m.View())
	}
	m = keys(menu(), "down", "enter", "enter", "enter", "s", "d", "k")
	m = keys(m, "esc")
	if v := m.View(); strings.Contains(v, "Name it") {
		t.Fatalf("esc stayed on the name:\n%s", v)
	}
	m = keys(m, "enter")
	if v := m.View(); !strings.Contains(v, "sdk") {
		t.Fatalf("the name typed before is lost:\n%s", v)
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
