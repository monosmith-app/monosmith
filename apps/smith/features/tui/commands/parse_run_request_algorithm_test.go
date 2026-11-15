package commands

import (
	"reflect"
	"testing"
)

func TestParseRunReadsTasksProjectsAndFlags(t *testing.T) {
	cases := []struct {
		argv []string
		want RunRequest
		ok   bool
	}{
		{[]string{"run", "lint", "test", "--affected"}, RunRequest{Tasks: []string{"lint", "test"}, Affected: true}, true},
		{[]string{"run", "build", "-p", "kit,web", "--no-cache"}, RunRequest{Tasks: []string{"build"}, Projects: []string{"kit", "web"}}, true},
		{[]string{"test", "web", "--parallel", "2"}, RunRequest{Tasks: []string{"test"}, Projects: []string{"web"}}, true},
		{[]string{"run", "build", "--projects=kind:app", "--base", "origin/main"}, RunRequest{Tasks: []string{"build"}, Projects: []string{"kind:app"}, Base: "origin/main"}, true},
		{[]string{"doctor"}, RunRequest{}, false},
		{[]string{"run"}, RunRequest{}, false},
		{nil, RunRequest{}, false},
	}
	for _, c := range cases {
		got, ok := ParseRun(c.argv)
		if ok != c.ok || (ok && !reflect.DeepEqual(got, c.want)) {
			t.Errorf("%v = %+v %v, want %+v %v", c.argv, got, ok, c.want, c.ok)
		}
	}
}

func TestRunArgv(t *testing.T) {
	if got := RunArgv("test", "kit"); !reflect.DeepEqual(got, []string{"run", "test", "-p", "kit"}) {
		t.Fatal(got)
	}
	if got := RunArgv("test", ""); !reflect.DeepEqual(got, []string{"run", "test"}) {
		t.Fatal(got)
	}
}

func TestTheReleaseBuildRunsTheBuildWithTheReleaseFlag(t *testing.T) {
	c, ok := ByID("run.build-release")
	if !ok {
		t.Fatal("no run.build-release command")
	}
	if got := c.TaskArgv("api"); !reflect.DeepEqual(got, []string{"run", "build", "-p", "api", "--release"}) {
		t.Errorf("one project: %v", got)
	}
	if got := c.TaskArgv(""); !reflect.DeepEqual(got, []string{"run", "build", "--release"}) {
		t.Errorf("every project: %v", got)
	}
	plain, _ := ByID("run.build")
	if got := plain.TaskArgv("api"); !reflect.DeepEqual(got, []string{"run", "build", "-p", "api"}) {
		t.Errorf("the plain build takes no flag: %v", got)
	}
}

func TestThePaletteOffersTheReleaseBuild(t *testing.T) {
	hints := func(s State) map[string]string {
		out := map[string]string{}
		for _, e := range PaletteEntries(s) {
			out[e.ID] = e.Hint
		}
		return out
	}
	home := hints(State{InWorkspace: true, View: TabHome, Projects: 3})
	if home["run.build-release-all"] != "smith run build --release" {
		t.Errorf("every project: %q", home["run.build-release-all"])
	}
	projects := hints(State{InWorkspace: true, View: TabProjects, Project: "web", Projects: 3})
	if projects["run.build-release"] != "smith run build -p web --release" {
		t.Errorf("the selected project: %q", projects["run.build-release"])
	}
}
