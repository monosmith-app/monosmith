package commands

import (
	"strings"
	"testing"
)

func TestAllCountsOnlyTheProjectsThatHaveTheTask(t *testing.T) {
	s := State{View: TabProjects, InWorkspace: true, Projects: 4, TaskProjects: map[string]int{"build": 3, "test": 4}}
	titles := map[string]string{}
	for _, e := range PaletteEntries(s) {
		titles[e.ID] = e.Title
	}
	if got := titles["run.build-all"]; got != "Build all 3" {
		t.Errorf("build all: %q", got)
	}
	if got := titles["run.test-all"]; got != "Test all 4" {
		t.Errorf("test all: %q", got)
	}
}

func TestThePaletteListsTheSelectedProjectsOwnTasks(t *testing.T) {
	s := State{View: TabProjects, InWorkspace: true, Projects: 2, Project: "web", OtherTasks: []string{"dev", "storybook"}}
	found := map[string]PaletteEntry{}
	for _, e := range PaletteEntries(s) {
		found[e.ID] = e
	}
	e, ok := found["x:run:dev"]
	if !ok || e.Title != "Run dev on web" || !strings.Contains(e.Hint, "run dev -p web") {
		t.Fatalf("the dev task has no palette entry: %+v", found)
	}
	if _, ok := found["x:run:storybook"]; !ok {
		t.Fatal("the storybook task has no palette entry")
	}
	s.Running = true
	for _, e := range PaletteEntries(s) {
		if e.ID == "x:run:dev" && e.Unavailable == "" {
			t.Fatal("a task cannot start while a run is in progress")
		}
	}
}
