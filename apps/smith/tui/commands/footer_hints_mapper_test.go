package commands

import "testing"

func TestAToggleFooterNamesWhatTheKeyDoesNext(t *testing.T) {
	label := func(id string, on bool) string {
		c, _ := ByID(id)
		key := c.Keys[0]
		s := State{View: TabDoctor, InWorkspace: true, On: map[string]bool{id: on}}
		view, _ := Hints(s)
		for _, h := range view {
			if h.Key == key {
				return h.Label
			}
		}
		return ""
	}
	cases := []struct {
		id   string
		on   bool
		want string
	}{
		{"doctor.group", false, "group"},
		{"doctor.group", true, "ungroup"},
		{"doctor.foldall", false, "expand all"},
		{"doctor.foldall", true, "collapse all"},
		{"doctor.failed", false, "show failed"},
		{"doctor.failed", true, "hide failed"},
		{"doctor.warnings", false, "show warnings"},
		{"doctor.warnings", true, "hide warnings"},
		{"doctor.passed", false, "show passed"},
		{"doctor.passed", true, "hide passed"},
	}
	for _, c := range cases {
		if got := label(c.id, c.on); got != c.want {
			t.Errorf("%s on=%v: label %q, want %q", c.id, c.on, got, c.want)
		}
	}
}

func TestWhileATextBoxIsOpenTheFooterShowsOnlyHowToEndIt(t *testing.T) {
	view, tail := Hints(State{View: TabGuide, Editing: true, Form: true})
	if len(tail) != 0 || len(view) != 2 || view[0].Label != "confirm" || view[0].Glyph != "Enter" || view[1].Label != "back" || view[1].Glyph != "Esc" {
		t.Fatalf("view %+v tail %+v", view, tail)
	}
}

func TestOnAFailedRunTheFooterLeadsWithRunAgainAndCopyLog(t *testing.T) {
	view, _ := Hints(State{View: TabRun, InWorkspace: true, RunFailed: true, Projects: 3})
	if len(view) < 2 || view[0].Label != "run again" || view[1].Label != "copy log" {
		t.Fatalf("%+v", view)
	}
}

func TestTheEscLabelSaysWhatEscDoesOnTheRunTab(t *testing.T) {
	label := func(s State) string {
		view, _ := Hints(s)
		for _, h := range view {
			if h.Glyph == "Esc" {
				return h.Label
			}
		}
		return ""
	}
	if got := label(State{View: TabRun, InWorkspace: true, TaskSelected: true}); got != "all output" {
		t.Errorf("with a task shown Esc is %q", got)
	}
	if got := label(State{View: TabRun, InWorkspace: true}); got != "start screen" {
		t.Errorf("with all output shown Esc is %q", got)
	}
}
