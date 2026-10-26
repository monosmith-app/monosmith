package commands

import "testing"

func TestRunKeysActOnTheSelectionOnlyOnProjects(t *testing.T) {
	home := State{View: TabHome, InWorkspace: true, Project: "web", Projects: 23}
	if got := home.RunProject(); got != "" {
		t.Errorf("Home runs on %q, want the whole workspace", got)
	}
	projects := home
	projects.View = TabProjects
	if got := projects.RunProject(); got != "web" {
		t.Errorf("Projects runs on %q, want web", got)
	}
}

func TestHomeKeysNameTheWholeWorkspace(t *testing.T) {
	view, _ := Hints(State{View: TabHome, InWorkspace: true, Project: "web", Projects: 23})
	for _, h := range view {
		if h.Label == "build web" || h.Label == "test web" {
			t.Errorf("Home hint %q names a hidden selection", h.Label)
		}
	}
	for _, e := range PaletteEntries(State{View: TabHome, InWorkspace: true, Project: "web", Projects: 23}) {
		if e.ID == "run.build" && (e.Title != "Build all 23" || e.Hint != "smith run build") {
			t.Errorf("Home palette build is %+v", e)
		}
	}
}

func TestProjectsKeysNameTheSelection(t *testing.T) {
	s := State{View: TabProjects, InWorkspace: true, Project: "web", Projects: 23}
	found := false
	view, _ := Hints(s)
	for _, h := range view {
		found = found || h.Label == "build web"
	}
	if !found {
		t.Errorf("Projects hints %+v lack build web", view)
	}
}

func TestRunScopeLabel(t *testing.T) {
	for _, c := range []struct {
		argv []string
		n    int
		want string
	}{
		{[]string{"run", "build"}, 23, "Build · all 23 projects"},
		{[]string{"run", "test", "-p", "web"}, 23, "Test · web"},
		{[]string{"run", "lint", "--affected"}, 23, "Lint · affected projects"},
		{[]string{"run", "build"}, 0, "Build · all projects"},
		{[]string{"doctor"}, 23, ""},
	} {
		if got := RunScopeLabel(c.argv, c.n); got != c.want {
			t.Errorf("%v: %q, want %q", c.argv, got, c.want)
		}
	}
}
