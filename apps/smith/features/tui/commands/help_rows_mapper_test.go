package commands

import "testing"

func TestHelpGreysAKeyThatCannotRunAndGivesTheReason(t *testing.T) {
	s := State{View: TabProjects, InWorkspace: true, Running: true, Project: "web", Projects: 3}
	var found bool
	for _, r := range HelpRows(s, "Projects") {
		if r.Title == "Build" {
			found = true
			if r.Disabled == "" {
				t.Errorf("a run is in progress, so build is disabled: %+v", r)
			}
		}
	}
	if !found {
		t.Fatal("help lists a disabled key, with its reason")
	}
}

func TestHelpInAnEmptyFolderListsOnlyWhatCanRun(t *testing.T) {
	for _, r := range HelpRows(State{View: TabGuide}, "Start") {
		if r.Disabled != "" {
			t.Errorf("%q cannot run here, so help must not list it: %q", r.Title, r.Disabled)
		}
		if r.Title == "Build" {
			t.Errorf("build needs a workspace: %+v", r)
		}
	}
}
