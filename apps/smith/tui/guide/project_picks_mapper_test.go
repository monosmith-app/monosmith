package guide

import (
	"strings"
	"testing"
)

func removeForm() Action {
	for _, a := range Actions([]string{"go"}, nil, nil) {
		if a.Title == "Remove a project" {
			return a
		}
	}
	panic("no Remove a project")
}

func TestRemoveAProjectPicksFromTheProjectsAndStartsOnTheSelectedOne(t *testing.T) {
	a := RemoveProjectAction(removeForm(), []string{"api", "web", "kit"})
	m, ok := NewModel([]Action{a}).WithDefault("name", "web").Pick("Remove a project")
	if !ok {
		t.Fatal("no form")
	}
	if view := m.View(); !strings.Contains(view, "api") || !strings.Contains(view, "kit") {
		t.Fatalf("the projects are not listed:\n%s", view)
	}
	m = press(m, "enter", "enter") // the project (web), then Remove it
	if got := strings.Join(m.Argv, " "); got != "remove web --dry-run" {
		t.Fatalf("argv %q: the pick must start on the selected project", got)
	}
}

func TestRemoveAProjectWithNoProjectKeepsTheTextForm(t *testing.T) {
	plain := removeForm()
	if got := RemoveProjectAction(plain, nil); got.Steps[0].Options != nil {
		t.Fatal("with no project there is nothing to pick")
	}
}

func TestAddProjectRefusesAnExistingName(t *testing.T) {
	a := AddProjectAction(plainAdd(), nil, []string{"web", "api"})
	check := a.Steps[2].Check
	if got := check(map[string][]string{"kind": {"app"}}, "web"); !strings.Contains(got, "exists already") {
		t.Fatalf("a taken name is accepted: %q", got)
	}
	if got := check(map[string][]string{"kind": {"app"}}, "shop"); got != "" {
		t.Fatalf("a free name is refused: %q", got)
	}
}
