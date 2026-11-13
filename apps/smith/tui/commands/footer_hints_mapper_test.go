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
	view, _ := Hints(State{View: TabProjects, RunOpen: true, InWorkspace: true, RunFailed: true, Projects: 3})
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
	if got := label(State{View: TabProjects, RunOpen: true, InWorkspace: true, TaskSelected: true}); got != "all output" {
		t.Errorf("with a task shown Esc is %q", got)
	}
	if got := label(State{View: TabProjects, RunOpen: true, InWorkspace: true}); got != "back" {
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
	if !hasHint(hints, "←↑↓→ move") || !hasHint(hints, "PgUp/PgDn scroll") {
		t.Fatalf("Home footer %v", hints)
	}
	// A page that fits has nothing to scroll: the hint leaves.
	s.Reasons = map[string]string{"home.scroll": "the page fits on the screen"}
	if hints := glyphLabels(s); hasHint(hints, "PgUp/PgDn scroll") || !hasHint(hints, "←↑↓→ move") {
		t.Fatalf("Home footer when the page fits %v", hints)
	}
}

func TestTabNamesWhereItGoesNext(t *testing.T) {
	list := State{View: TabProjects, InWorkspace: true, Projects: 3}
	if hints := glyphLabels(list); !hasHint(hints, "Tab recent runs") || hasHint(hints, "Tab detail") {
		t.Fatalf("from the list with runs: %v", hints)
	}
	noRuns := list
	noRuns.SideNext, noRuns.Reasons = true, map[string]string{"run.recent-focus": "nothing has run yet"}
	if hints := glyphLabels(noRuns); !hasHint(hints, "Tab detail") || hasHint(hints, "Tab recent runs") {
		t.Fatalf("from the list without runs: %v", hints)
	}
	recent := list
	recent.RecentFocus, recent.SideNext = true, true
	if hints := glyphLabels(recent); !hasHint(hints, "Tab detail") || hasHint(hints, "Tab projects") || !hasHint(hints, "↑↓ choose run") {
		t.Fatalf("from the recent runs, with the detail beside: %v", hints)
	}
	recent.SideNext, recent.On = false, map[string]bool{"run.recent-focus": true}
	if hints := glyphLabels(recent); !hasHint(hints, "Tab projects") || hasHint(hints, "Tab detail") {
		t.Fatalf("from the recent runs, nothing beside: %v", hints)
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
	// With a run open the run's own keys lead, and Tab gives the keyboard back.
	hints := glyphLabels(State{View: TabProjects, InWorkspace: true, SideFocus: true, RunOpen: true, Projects: 3})
	if !hasHint(hints, "Tab list") || hasHint(hints, "↑↓ choose") || !hasHint(hints, "↑↓ task") {
		t.Errorf("run open: %v", hints)
	}
}

func TestReleaseNamesTheKeysThatShowAndHideTheCommits(t *testing.T) {
	hints := glyphLabels(State{View: TabRelease, InWorkspace: true})
	if !hasHint(hints, "←→ expand/collapse") {
		t.Fatalf("Release footer %v", hints)
	}
	if hints := glyphLabels(State{View: TabRelease, InWorkspace: true, Reasons: map[string]string{"release.expand": "there is no bump to show"}}); hasHint(hints, "←→ expand/collapse") {
		t.Fatalf("nothing to expand, yet the footer offers it: %v", hints)
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
