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
	view, _ := Hints(State{View: TabRuns, RunOpen: true, InWorkspace: true, RunFailed: true, Projects: 3})
	if len(view) < 2 || view[0].Label != "run again" || view[1].Label != "copy log" {
		t.Fatalf("%+v", view)
	}
}

func TestTheEscLabelSaysWhatEscDoesInTheRunPanel(t *testing.T) {
	label := func(s State) string {
		view, _ := Hints(s)
		for _, h := range view {
			if h.Glyph == "Esc" {
				return h.Label
			}
		}
		return ""
	}
	if got := label(State{View: TabRuns, RunOpen: true, InWorkspace: true, TaskSelected: true}); got != "all output" {
		t.Errorf("with a task shown Esc is %q", got)
	}
	if got := label(State{View: TabRuns, RunOpen: true, InWorkspace: true}); got != "back" {
		t.Errorf("with all output shown Esc is %q", got)
	}
}

// glyphLabels are the footer hints of a state as "glyph label", the view's own first.
func glyphLabels(s State) []string {
	view, _ := Hints(s)
	var out []string
	for _, h := range view {
		out = append(out, h.Glyph+" "+h.Label)
	}
	return out
}

func hasHint(hints []string, want string) bool {
	for _, h := range hints {
		if h == want {
			return true
		}
	}
	return false
}

func TestHomeNamesTheKeysThatMoveAndScrollIt(t *testing.T) {
	s := State{View: TabHome, InWorkspace: true}
	hints := glyphLabels(s)
	if !hasHint(hints, "↑/↓ move") || !hasHint(hints, "PgUp/PgDn scroll") {
		t.Fatalf("Home footer %v", hints)
	}
	// A page that fits has nothing to scroll: the hint leaves.
	s.Reasons = map[string]string{"home.scroll": "the page fits on the screen"}
	if hints := glyphLabels(s); hasHint(hints, "PgUp/PgDn scroll") || !hasHint(hints, "↑/↓ move") {
		t.Fatalf("Home footer when the page fits %v", hints)
	}
}

func TestALabelOfTheStateNamesWhatEnterDoesOnHome(t *testing.T) {
	s := State{View: TabHome, InWorkspace: true, Labels: map[string]string{"home.next": "see why"}}
	hints := glyphLabels(s)
	if !hasHint(hints, "Enter see why") || hasHint(hints, "Enter next") {
		t.Fatalf("Home footer %v", hints)
	}
	if hints[0] != "↑/↓ move" || hints[1] != "Enter see why" {
		t.Fatalf("the footer starts with the move and Enter hints: %v", hints)
	}
}

func TestTabNamesWhereItGoesNext(t *testing.T) {
	list := State{View: TabProjects, InWorkspace: true, Projects: 3, SideNext: true}
	if hints := glyphLabels(list); !hasHint(hints, "Tab detail") || hasHint(hints, "Tab section") {
		t.Fatalf("from the list, with the detail beside: %v", hints)
	}
	alone := list
	alone.SideNext = false
	if hints := glyphLabels(alone); hasHint(hints, "Tab detail") || hasHint(hints, "Tab section") {
		t.Fatalf("from the list, nothing beside: %v", hints)
	}
	sections := State{View: TabProjects, InWorkspace: true, Projects: 3, SideFocus: true, ProjectDetail: true}
	if hints := glyphLabels(sections); !hasHint(hints, "Tab section") || hasHint(hints, "Tab list") || hasHint(hints, "Tab detail") {
		t.Fatalf("from the sections of the detail: %v", hints)
	}
	doctor := State{View: TabDoctor, InWorkspace: true, SideNext: true}
	if hints := glyphLabels(doctor); !hasHint(hints, "Tab detail") {
		t.Fatalf("Doctor list: %v", hints)
	}
}

func TestWhileTheDetailHasTheKeyboardTheFooterListsItsKeysAndNotTheLists(t *testing.T) {
	for _, view := range []string{TabProjects, TabDoctor} {
		hints := glyphLabels(State{View: view, InWorkspace: true, SideFocus: true, Projects: 3})
		for _, want := range []string{"Tab list", "↑↓ choose", "Enter run"} {
			if !hasHint(hints, want) {
				t.Errorf("%s: missing %q in %v", view, want, hints)
			}
		}
		for _, not := range []string{"Tab detail", "Tab recent runs", "Enter detail"} {
			if hasHint(hints, not) {
				t.Errorf("%s: %q should not show while the detail has the keyboard: %v", view, not, hints)
			}
		}
	}
	// In the Runs view the run's own keys lead; it has no list to give the keyboard back to.
	hints := glyphLabels(State{View: TabRuns, InWorkspace: true, RunOpen: true, Projects: 3})
	if hasHint(hints, "Tab list") || hasHint(hints, "↑↓ choose") || !hasHint(hints, "↑↓ task") {
		t.Errorf("run open: %v", hints)
	}
}

func TestNarrowReleaseNamesTheKeyThatOpensTheChanges(t *testing.T) {
	hints := glyphLabels(State{View: TabRelease, InWorkspace: true, Narrow: true})
	if !hasHint(hints, "Enter changes") {
		t.Fatalf("narrow Release footer %v", hints)
	}
	if hints := glyphLabels(State{View: TabRelease, InWorkspace: true}); hasHint(hints, "Enter changes") {
		t.Fatalf("the wide detail shows the changes, yet the footer offers Enter: %v", hints)
	}
	if hints := glyphLabels(State{View: TabRelease, InWorkspace: true, Narrow: true, Reasons: map[string]string{"release.open": "there is no bump to show"}}); hasHint(hints, "Enter changes") {
		t.Fatalf("nothing to open, yet the footer offers it: %v", hints)
	}
	if hints := glyphLabels(State{View: TabRelease, InWorkspace: true, Narrow: true, DetailOpen: true}); !hasHint(hints, "Esc back") {
		t.Fatalf("the open package has no way back: %v", hints)
	}
}

func TestTheArrowKeysAreNeverACommandOfTheirOwn(t *testing.T) {
	// Left and Right are keys of the view's pane: ByKey leaves them to it.
	for _, tc := range []struct{ key, view string }{
		{"left", TabHome}, {"right", TabHome}, {"left", TabDoctor}, {"right", TabProjects}, {"left", TabRelease}, {"right", TabRelease},
	} {
		if c, _, ok := ByKey(tc.key, State{View: tc.view, InWorkspace: true}); ok {
			t.Errorf("%s in %s is bound to command %s", tc.key, tc.view, c.ID)
		}
	}
	if c, _, ok := ByKey("ctrl+d", State{View: TabHome, InWorkspace: true}); !ok || c.Action != "home-down" {
		t.Errorf("ctrl+d on Home is %+v", c)
	}
	if c, _, ok := ByKey("ctrl+u", State{View: TabHome, InWorkspace: true}); !ok || c.Action != "home-up" {
		t.Errorf("ctrl+u on Home is %+v", c)
	}
}
