package commands

import "testing"

func TestRunKeysActOnTheSelectionOnEveryView(t *testing.T) {
	for _, view := range []string{TabHome, TabProjects, TabRun, TabDoctor, TabDeps, TabRelease} {
		s := State{View: view, InWorkspace: true, Project: "web", Projects: 23}
		if got := s.RunProject(); got != "web" {
			t.Errorf("%s runs on %q, want web", view, got)
		}
	}
	if got := (State{View: TabRun, InWorkspace: true, Projects: 23}).RunProject(); got != "" {
		t.Errorf("with no selection a run covers the workspace, got %q", got)
	}
}

func TestTheCapitalKeysNameEveryProjectOnEveryView(t *testing.T) {
	for _, view := range []string{TabHome, TabRun} {
		s := State{View: view, InWorkspace: true, Project: "web", Projects: 23}
		hints, _ := Hints(s)
		var small, big bool
		for _, h := range hints {
			small = small || h.Label == "build web"
			big = big || h.Label == "build all 23"
		}
		if !small || !big {
			t.Errorf("%s: b is %v and B is %v in %+v", view, small, big, hints)
		}
		for _, e := range PaletteEntries(s) {
			if e.ID == "run.build" && (e.Title != "Build web" || e.Hint != "smith run build -p web") {
				t.Errorf("%s palette build is %+v", view, e)
			}
			if e.ID == "run.build-all" && e.Title != "Build all 23" {
				t.Errorf("%s palette build all is %+v", view, e)
			}
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
		{[]string{"run", "build"}, 1, "Build · the only project"},
		{[]string{"run", "build"}, 0, "Build · all projects"},
		{[]string{"doctor"}, 23, ""},
	} {
		if got := RunScopeLabel(c.argv, c.n); got != c.want {
			t.Errorf("%v: %q, want %q", c.argv, got, c.want)
		}
	}
}
