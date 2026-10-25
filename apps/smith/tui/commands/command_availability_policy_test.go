package commands

import (
	"slices"
	"strings"
	"testing"
)

var (
	inside  = State{InWorkspace: true, Narrow: true, View: TabProjects}
	outside = State{View: TabGuide}
	busy    = State{InWorkspace: true, Running: true, View: TabProjects}
)

var everyView = []string{TabHome, TabProjects, TabRun, TabDoctor, TabDeps, TabRelease, TabGuide}

func TestNoKeyIsBoundTwiceInTheSameViewAndState(t *testing.T) {
	for _, base := range []State{inside, outside, busy} {
		for _, view := range everyView {
			s := base
			s.View = view
			seen := map[string]string{}
			for _, c := range all {
				if c.Kind != PaneKey && !c.Visible(s) {
					continue
				}
				if len(c.Views) > 0 && !slices.Contains(c.Views, view) {
					continue
				}
				for _, k := range c.Keys {
					// A view's own key may repeat a global one: it wins (see ByKey).
					if prev, dup := seen[k]; dup && len(c.Views) > 0 == (prevScoped(prev)) {
						t.Errorf("%q is bound to %s and %s in %s", k, prev, c.ID, view)
					}
					seen[k] = c.ID
				}
			}
		}
	}
}

// prevScoped says whether the command with that id belongs to one view.
func prevScoped(id string) bool {
	c, _ := ByID(id)
	return len(c.Views) > 0
}

func TestEveryKeyIsASingleKeyWithoutAnAltChord(t *testing.T) {
	for _, c := range all {
		for _, k := range c.Keys {
			if strings.HasPrefix(k, "alt+") {
				t.Errorf("%s is bound to %s: no key needs Alt", c.ID, k)
			}
			if k == "ctrl+q" {
				t.Errorf("%s is bound to ctrl+q, which closes the terminal of some editors", c.ID)
			}
		}
	}
}

func TestEveryCommandWithAKeyShowsItsShortcut(t *testing.T) {
	for _, c := range all {
		if len(c.Keys) > 0 && c.Shortcut == "" {
			t.Errorf("%s has keys but shows no shortcut", c.ID)
		}
	}
	if c, _ := ByID("file.palette"); c.Shortcut != ": / Ctrl+P" {
		t.Errorf("the palette shows %q", c.Shortcut)
	}
	if c, _ := ByID("help.keys"); c.Shortcut != "? / F1" {
		t.Errorf("the help shows %q", c.Shortcut)
	}
}

func TestEveryCommandHasATitleAndExecuteHasAnArgv(t *testing.T) {
	for _, c := range all {
		if c.Title == "" || c.ID == "" {
			t.Errorf("%+v", c)
		}
		if c.Kind == Execute && len(c.Argv) == 0 {
			t.Errorf("%s executes nothing", c.ID)
		}
		if !slices.Contains(GroupOrder, c.Group) {
			t.Errorf("%s is in an unknown group %q", c.ID, c.Group)
		}
		if c.Kind == ViewAction && (c.Action == "" || len(c.Views) == 0) {
			t.Errorf("%s is a view action without an action or a view", c.ID)
		}
		if c.Footer != "" && len(c.Keys) == 0 {
			t.Errorf("%s is in the footer without a key", c.ID)
		}
	}
}

func TestOutsideAWorkspaceTheWorkspaceCommandsSayWhyTheyAreOff(t *testing.T) {
	c, reason, ok := ByKey("d", outside)
	if !ok || c.ID != "ws.doctor" || !strings.Contains(reason, "not in a workspace (press n") {
		t.Fatalf("%v %q %v", c.ID, reason, ok)
	}
	if _, reason, _ := ByKey("d", inside); reason != "" {
		t.Fatalf("available inside: %q", reason)
	}
}

func TestWhileARunIsGoingNothingElseStartsAndStopIsAvailable(t *testing.T) {
	if _, reason, _ := ByKey("b", busy); !strings.Contains(reason, "run is in progress") {
		t.Fatalf("%q", reason)
	}
	if _, reason, _ := ByKey("ctrl+c", busy); reason != "" {
		t.Fatalf("stop should be available: %q", reason)
	}
	if _, reason, _ := ByKey("ctrl+c", inside); reason != "nothing is running" {
		t.Fatalf("%q", reason)
	}
}

func TestNewWorkspaceIsOnlyOutsideAndAddProjectOnlyInside(t *testing.T) {
	// Issue #472: Ctrl+N meant both. Now each has its own key and only shows where it is valid.
	if c, _, _ := ByKey("n", outside); c.ID != "file.new" {
		t.Fatalf("outside: %s", c.ID)
	}
	if c, _, ok := ByKey("n", inside); ok && c.ID == "file.new" {
		t.Fatal("n must not start a workspace inside one")
	}
	if c, _, _ := ByKey("a", inside); c.ID != "file.add" {
		t.Fatalf("inside: %s", c.ID)
	}
	if c, _, ok := ByKey("a", outside); ok && c.ID == "file.add" {
		t.Fatal("a must not add a project outside a workspace")
	}
}

func TestAViewsOwnKeyWinsOverTheSameGlobalKey(t *testing.T) {
	deps := inside
	deps.View = TabDeps
	if _, _, ok := ByKey("a", deps); ok {
		t.Fatal("a marks all in Deps, a pane key: no command runs")
	}
	if c, _, _ := ByKey("c", inside); c.ID != "run.typecheck" {
		t.Fatalf("c in Projects is %s", c.ID)
	}
}

func TestNoPlainKeyWorksWhileATextFieldOrAFormHasTheKeyboard(t *testing.T) {
	for _, s := range []State{{InWorkspace: true, View: TabGuide, Editing: true}, {InWorkspace: true, View: TabGuide, Form: true}} {
		for _, c := range all {
			for _, k := range c.Keys {
				if _, _, ok := ByKey(k, s); ok && plain(k) {
					t.Errorf("%q fires %s while typing", k, c.ID)
				}
			}
		}
	}
	editing := State{InWorkspace: true, View: TabGuide, Editing: true}
	for _, k := range []string{"ctrl+p", "f1"} {
		if c, _, ok := ByKey(k, editing); !ok || !c.WhileEditing {
			t.Errorf("%s should still work while editing", k)
		}
	}
	if _, _, ok := ByKey("q", editing); ok {
		t.Error("q must not quit while a field is editing")
	}
}

func TestTheFooterHintsAreTheKeysThatWork(t *testing.T) {
	for _, view := range everyView {
		for _, base := range []State{inside, outside} {
			s := base
			s.View = view
			v, tail := Hints(s)
			for _, h := range append(v, tail...) {
				if h.Key == "" {
					continue
				}
				c, reason, ok := ByKey(h.Key, s)
				// A pane key answers ok false: the pane handles it.
				if ok && reason != "" {
					t.Errorf("%s shows %q (%s) but it is refused: %s", view, h.Glyph, h.Label, reason)
				}
				if ok && c.Footer != h.Label {
					t.Errorf("%s shows %q as %q but the key runs %s", view, h.Glyph, h.Label, c.ID)
				}
			}
		}
	}
}

func TestEachViewHasItsOwnHintsAndEveryViewEndsWithTheSameTail(t *testing.T) {
	line := func(view string) string {
		s := inside
		s.View = view
		v, tail := Hints(s)
		var out []string
		for _, h := range append(v, tail...) {
			out = append(out, h.Glyph+" "+h.Label)
		}
		return strings.Join(out, " · ")
	}
	for view, want := range map[string]string{
		TabProjects: "Enter detail · b build · t test · l lint · c typecheck · w changed · d doctor · a add project · z zoom · y copy",
		TabDoctor:   "Enter expand · f fix · x fix all · g group · p passed · r reload · b build · t test · z zoom · y copy",
		TabHome:     "Enter next · r reload · b build · t test · w changed · d doctor",
		TabDeps:     "Space mark · s safe · a all · u update · v align · r reload · b build · t test · y copy",
		TabRelease:  "Enter commits · p preview · n notes · x release · r reload · b build · t test",
	} {
		if got := line(view); !strings.HasPrefix(got, want) {
			t.Errorf("%s footer is\n%s\nwant\n%s", view, got, want)
		}
		if got := line(view); !strings.HasSuffix(got, "1-6 views · m mouse · : palette · ? help · q quit") {
			t.Errorf("%s footer ends %q", view, got)
		}
	}
	if got := line(TabRun); !strings.HasPrefix(got, "r again · R failed · Y copy log · , . runs · f follow") {
		t.Errorf("Run footer %q", got)
	}
}

func TestTheDetailKeysFollowTheLayout(t *testing.T) {
	labels := func(s State) []string {
		v, _ := Hints(s)
		var out []string
		for _, h := range v {
			out = append(out, h.Label)
		}
		return out
	}
	wide := State{InWorkspace: true, View: TabProjects}
	if got := labels(wide); slices.Contains(got, "detail") || slices.Contains(got, "back") {
		t.Fatalf("beside the list there is nothing to open or close: %v", got)
	}
	open := State{InWorkspace: true, Narrow: true, DetailOpen: true, View: TabProjects}
	if got := labels(open); !slices.Contains(got, "back") || slices.Contains(got, "detail") {
		t.Fatalf("an open detail offers a way back: %v", got)
	}
}

func TestOutsideAWorkspaceTheFooterOffersNewAndNoAdd(t *testing.T) {
	v, _ := Hints(State{View: TabGuide})
	var labels []string
	for _, h := range v {
		labels = append(labels, h.Label)
	}
	if !slices.Contains(labels, "new workspace") || slices.Contains(labels, "add project") {
		t.Fatalf("outside: %v", labels)
	}
	v, _ = Hints(inside)
	labels = nil
	for _, h := range v {
		labels = append(labels, h.Label)
	}
	if slices.Contains(labels, "new workspace") || !slices.Contains(labels, "add project") {
		t.Fatalf("inside: %v", labels)
	}
}

func TestWhileARunGoesTheFooterOffersStopNotBuild(t *testing.T) {
	v, _ := Hints(busy)
	var labels []string
	for _, h := range v {
		labels = append(labels, h.Label)
	}
	if !slices.Contains(labels, "stop") || slices.Contains(labels, "build") {
		t.Fatalf("running: %v", labels)
	}
}

func TestTheHelpListsTheViewKeysFirstAndNeverAKeyThatDoesNothing(t *testing.T) {
	s := inside
	s.View = TabDoctor
	rows := HelpRows(s, "Doctor")
	if rows[0].Group != "Doctor view" {
		t.Fatalf("first group %q", rows[0].Group)
	}
	var titles []string
	for _, r := range rows {
		titles = append(titles, r.Title+"="+r.Shortcut)
	}
	joined := strings.Join(titles, "\n")
	for _, want := range []string{"Build=b", "Typecheck=c", "Doctor=d", "Command palette…=: / Ctrl+P", "Keyboard shortcuts=? / F1", "Quit=q", "Mouse on or off=m", "Add project…=a"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the help lacks %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "New workspace") {
		t.Error("New workspace is not offered inside a workspace")
	}
}

func TestCtrlQIsNeverBound(t *testing.T) {
	for _, c := range all {
		for _, k := range c.Keys {
			if k == "ctrl+q" {
				t.Errorf("%s is bound to ctrl+q", c.ID)
			}
		}
	}
}

func TestQuitCopyAndMouseAreSingleKeys(t *testing.T) {
	for key, id := range map[string]string{"q": "file.quit", "y": "view.copy", "m": "view.mouse"} {
		if c, reason, ok := ByKey(key, inside); !ok || c.ID != id || reason != "" {
			t.Errorf("%s: %s %q %v", key, c.ID, reason, ok)
		}
	}
}

func TestNumberKeysSwitchViewsInTheOrderOfTheTabs(t *testing.T) {
	want := []string{TabHome, TabProjects, TabRun, TabDoctor, TabDeps, TabRelease}
	for i, tab := range want {
		key := string(rune('1' + i))
		if c, _, ok := ByKey(key, inside); !ok || c.Tab != tab {
			t.Errorf("%s opens %q, want %q", key, c.Tab, tab)
		}
	}
}

func TestAViewLeavesOutTheHintsOfKeysThatCannotDoAnythingNow(t *testing.T) {
	labels := func(s State) []string {
		v, _ := Hints(s)
		var out []string
		for _, h := range v {
			out = append(out, h.Label)
		}
		return out
	}
	deps := inside
	deps.View = TabDeps
	if got := labels(deps); !slices.Contains(got, "update") || !slices.Contains(got, "align") {
		t.Fatalf("with something to do: %v", got)
	}
	deps.Reasons = map[string]string{"deps.update": "no dependency to update", "deps.align": "nothing to align"}
	if got := labels(deps); slices.Contains(got, "update") || slices.Contains(got, "align") {
		t.Fatalf("with nothing to do the hints stay: %v", got)
	}
	if _, reason, ok := ByKey("u", deps); !ok || reason != "no dependency to update" {
		t.Fatalf("the key says why it does nothing: %q %v", reason, ok)
	}
}

func TestTheDoctorExpandsInPlaceOnlyWhenNoDetailIsBesideIt(t *testing.T) {
	has := func(s State, label string) bool {
		v, _ := Hints(s)
		for _, h := range v {
			if h.Label == label {
				return true
			}
		}
		return false
	}
	narrow := State{InWorkspace: true, Narrow: true, View: TabDoctor}
	wide := State{InWorkspace: true, View: TabDoctor}
	if !has(narrow, "expand") || has(wide, "expand") {
		t.Fatal("Enter expands a check in the single pane and does nothing beside the detail")
	}
}

func TestTheRunFooterDoesNotOfferFollowOrTasksBeforeAnythingRan(t *testing.T) {
	s := State{InWorkspace: true, View: TabRun, Reasons: map[string]string{
		"run.follow": "nothing has run yet", "run.scroll": "nothing has run yet",
	}}
	v, _ := Hints(s)
	for _, h := range v {
		if h.Label == "follow" || h.Label == "task" {
			t.Fatalf("the footer offers %q", h.Label)
		}
	}
}

func TestWhatTheRemovedViewsShowedIsStillACommandOfThePalette(t *testing.T) {
	for id, argv := range map[string][]string{
		"cli.plugins": {"plugin", "list"}, "cli.recipes": {"recipe"}, "cli.settings": {"config", "list"}, "cli.guide": {"help"},
	} {
		c, ok := ByID(id)
		if !ok || !slices.Equal(c.Argv, argv) || c.Kind != Execute {
			t.Errorf("%s: %+v", id, c)
		}
	}
}
