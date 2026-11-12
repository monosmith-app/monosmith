package tui

import (
	"bytes"
	"encoding/base64"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/guide"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

func TestTheScreenHasNoMenuBarAndTheFooterShowsTheKeysOfTheOpenView(t *testing.T) {
	l := newLab(t, true, 260, 30)
	l.hasNot("File", "Workspace  Run", "Alt+")
	if first := l.lines()[0]; !strings.Contains(first, "acme") || !strings.Contains(first, "main") {
		t.Fatalf("the first row names the workspace and its branch, with no menu bar above: %q", first)
	}
	// On a wide screen the header shares the row of the tabs, at its right end.
	if first := l.lines()[0]; !strings.Contains(first, "Home") || !strings.Contains(first, "Projects") || !strings.Contains(first, "Doctor") {
		t.Fatalf("the views are tabs on the first row of a wide screen: %q", first)
	}
	footer := l.lines()[len(l.lines())-2]
	for _, want := range []string{"build orders (b)", "test orders (t)", "lint orders (l)", "typecheck orders (c)", "changed (w)", "doctor (d)", "add project (a)", "palette (:)", "help (?)", "quit (q)", "mouse (m)"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the footer lacks %q: %q", want, footer)
		}
	}
	l.keys("3")
	footer = l.lines()[len(l.lines())-2]
	for _, want := range []string{"fix all (F)", "group (g)", "reload view (r)", "quit (q)"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the Doctor footer lacks %q: %q", want, footer)
		}
	}
	if strings.Contains(footer, "changed (w)") {
		t.Fatalf("the Doctor footer keeps the Projects keys: %q", footer)
	}
}

func TestNoAltChordOrFunctionKeyOpensAnythingAnyMore(t *testing.T) {
	l := newLab(t, true, 120, 30)
	before := l.view()
	for _, k := range []string{"alt+1", "alt+3", "alt+r", "alt+f", "f10", "f2", "f3", "f4", "ctrl+o"} {
		l.keys(k)
	}
	if l.view() != before || len(l.started) != 0 {
		t.Fatalf("a retired chord did something: %v %s", l.started, l.view())
	}
}

func TestSingleKeysRunTheViewsCommands(t *testing.T) {
	cases := map[string]string{"b": "run build -p orders", "t": "run test -p orders", "l": "run lint -p orders", "c": "run typecheck -p orders", "w": "run lint test --affected"}
	for key, want := range cases {
		l := newLab(t, true, 120, 30)
		l.keys(key)
		if l.lastStarted() != want {
			t.Errorf("%s started %q, want %q", key, l.lastStarted(), want)
		}
	}
}

func TestDRechecksTheWorkspace(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("d")
	l.has("2 failed", "2 checks passed")
}

func TestEnterBesideTheDetailChangesNothing(t *testing.T) {
	l := newLab(t, true, 140, 34)
	before := l.view()
	l.keys("enter")
	if l.view() != before || len(l.started) != 0 || l.sh().detailOpen {
		t.Fatalf("the detail is already beside the list: %v", l.started)
	}
}

func TestJAndKMoveLikeTheArrows(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("j")
	if got := l.sh().list.SelectedName(); got != "web" {
		t.Fatalf("j moved to %q", got)
	}
	l.keys("k")
	if got := l.sh().list.SelectedName(); got != "orders" {
		t.Fatalf("k moved to %q", got)
	}
}

func TestKeysDoNotFireWhileATextFieldIsEditing(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("a", "down", "down", "enter", "down", "enter") // Add project, a package in node: the name is a text field
	l.typeText("bq?:1/ mw")
	if len(l.started) != 0 || l.quit {
		t.Fatalf("typing ran something: %v quit %v", l.started, l.quit)
	}
	if l.sh().tab != "guide" || l.sh().palOpen || l.sh().help != nil {
		t.Fatalf("typing changed the screen: tab %q", l.sh().tab)
	}
	l.keys("esc", "ctrl+p")
	if !l.sh().palOpen {
		t.Fatal("ctrl+p still opens the palette while typing")
	}
}

func TestKeysDoNotFireInsideAGuideForm(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("a")
	l.has("Add a project", "What are you building?")
	l.keys("b", "t", "3")
	if len(l.started) != 0 || l.sh().tab != "guide" {
		t.Fatalf("a form option list ran %v in %q", l.started, l.sh().tab)
	}
}

func TestShortcutsExplainWhenTheyCannotRun(t *testing.T) {
	l := newLab(t, false, 100, 28)
	l.keys("b")
	l.has("Build is unavailable: not in a workspace (press n to create one)")
}

func TestAViewsLettersAreItsOwnNotTheGlobalOnes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("4", "p") // Deps: p shows the patch releases, a letter of the view
	if l.sh().tab != "deps" {
		t.Fatalf("p left Deps for %q", l.sh().tab)
	}
}

func TestCtrlPOpensThePaletteListingEveryCommandWithItsShortcut(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p")
	l.has("Command palette", "Build", "Build orders", "smith run build -p orders", "Build all 3")
	l.typeText("typec")
	l.has("Check types")
	l.hasNot("Quit")
	l.keys("esc", "ctrl+p").typeText("affect") // one entry per intent, in plain words
	l.has("Show what changed", "smith affected")
	l.hasNot("List projects a change touches")
	l.keys("esc")
	l.hasNot("Command palette")
}

func TestThePaletteRunsAShellCommandAndACommandLine(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p").typeText("doctor").keys("enter") // the shell's Doctor comes first
	l.has("2 failed")
	l.keys("ctrl+p").typeText("show what").keys("enter") // a command line: it runs
	if l.lastStarted() != "affected" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestQAndCtrlCQuitWhenNothingRuns(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c"} {
		l := newLab(t, true, 100, 28)
		l.keys(k)
		if !l.quit {
			t.Fatalf("%s with nothing running should quit", k)
		}
	}
}

func TestCtrlQDoesNothing(t *testing.T) {
	// Ctrl+Q closes the integrated terminal of some editors, so smith never uses it.
	l := newLab(t, true, 100, 28)
	l.keys("ctrl+q")
	if l.quit {
		t.Fatal("ctrl+q must stay unused")
	}
}

func TestEscNeverQuitsFromTheTopLevel(t *testing.T) {
	l := newLab(t, true, 100, 28)
	l.keys("esc", "esc")
	if l.quit {
		t.Fatal("esc must not quit")
	}
	l.keys("a", "esc")
	if l.quit {
		t.Fatal("esc in the guide must not quit")
	}
}

func TestQDoesNotQuitWhileARunGoesNorWhileTyping(t *testing.T) {
	l := newLab(t, true, 170, 28)
	l.next = &scripted{hold: true, release: make(chan struct{})}
	l.keys("b", "q")
	if l.quit {
		t.Fatal("q must not quit while a run goes")
	}
	l.has("Ctrl+C stops it", "stop (Ctrl+C)")
	l.keys("ctrl+c")

	l = newLab(t, true, 100, 28)
	l.keys("ctrl+p").typeText("quit")
	if l.quit {
		t.Fatal("q typed in the palette must not quit")
	}
}

func TestQInsideAGuideFormGoesBackInsteadOfQuitting(t *testing.T) {
	l := newLab(t, true, 100, 28)
	l.keys("a", "down", "enter")
	l.keys("q")
	if l.quit {
		t.Fatal("q inside a form goes back")
	}
}

func TestYCopiesTheFocusedItemWithOSC52AndSaysSo(t *testing.T) {
	l := newLab(t, true, 100, 28)
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh

	l.keys("3")
	l.keys("y")
	want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte("go work use ./libs/x")) + "\a"
	if out.String() != want {
		t.Fatalf("doctor copies the remedy: %q", out.String())
	}
	l.has("copied: go work use ./libs/x")

	out.Reset()
	l.keys("2", "y")
	if out.Len() == 0 || !strings.Contains(string(mustDecode(t, out.String())), "/") {
		t.Fatalf("projects copies the project root: %q", out.String())
	}

	out.Reset()
	l.keys("4", "y")
	got := string(mustDecode(t, out.String()))
	if !slices.ContainsFunc(sampleDeps.Rows, func(r depsview.Row) bool { return r.Name == got }) {
		t.Fatalf("deps copies the package: %q", got)
	}

	out.Reset()
	l.keys("b", "y")
	if got := string(mustDecode(t, out.String())); !strings.HasPrefix(got, "smith ") {
		t.Fatalf("run copies the command line: %q", got)
	}
}

func TestYWhileTypingIsTheFieldsAndCopiesNothing(t *testing.T) {
	l := newLab(t, true, 100, 28)
	var out bytes.Buffer
	sh := l.sh()
	sh.clip = &out
	l.m = sh
	l.keys("ctrl+p").typeText("y")
	if out.Len() != 0 {
		t.Fatal("y typed in the palette must not copy")
	}
}

func TestMTurnsTheMouseOffAndOnAndTheFooterSaysSo(t *testing.T) {
	l := newLab(t, true, 140, 28)
	l.keys("m")
	if !l.sh().mouseOff {
		t.Fatal("m turns the mouse off")
	}
	l.has("mouse off")
	l.keys("m")
	if l.sh().mouseOff {
		t.Fatal("m turns the mouse on again")
	}
	l.hasNot("mouse off · m")
}

func TestHelpTellsHowToSelectTextAndHasNoStaleKeys(t *testing.T) {
	l := newLab(t, true, 100, 50)
	l.keys("f1", "end")
	l.has("Shift+drag selects text")
	l.hasNot("Ctrl+Q", "Hover")
}

func TestF1ListsEveryShortcutAndTheMouse(t *testing.T) {
	l := newLab(t, true, 100, 44)
	l.keys("f1")
	l.has("Keyboard shortcuts", "Move in a list")
	l.keys("pgdown")
	l.has("Build all", "Stop", "Ctrl+C")
	l.keys("pgup")
	l.keys("end")
	l.has(": / Ctrl+P", "? / F1", "Quit", "Esc", "Mouse", "Double-click", "footer")
	l.hasNot("Right-click", "Alt+", "F10", "menu")
	l.keys("esc")
	l.hasNot("Keyboard shortcuts")
	l.keys("?")
	l.has("Keyboard shortcuts")
}

func TestBelow100ColumnsEnterOpensTheDetailAndEscGoesBack(t *testing.T) {
	l := newLab(t, true, 90, 30)
	l.hasNot("Build app")   // the detail is not open
	l.keys("down", "enter") // web
	if !l.sh().detailOpen {
		t.Fatal("Enter should open the detail")
	}
	l.has("dependencies", "Build app", "Test app")
	l.keys("esc")
	if l.sh().detailOpen {
		t.Fatal("Esc should go back to the list")
	}
	l.has("4 projects")
	l.keys("enter", "down", "enter", "enter") // the second action: Test app, on web; Enter moves to its run button, then runs it
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestAtTheSupportedFloorTheDetailStillOpensFullWidth(t *testing.T) {
	l := newLab(t, true, 72, 24)
	l.keys("enter")
	l.has("orders", "Build app", "How was this resolved?")
	l.keys("esc")
	l.has("4 projects", "Project", "Language")
	l.hasNot("Tasks")
	l.hasNot("Build app")
}

func TestNumberedShortcutsSwitchTabs(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("3")
	l.has("2 failed", "2 checks passed")
	l.keys("5")
	l.has("1.0.0 →", "1.1.0", "unchanged: sdk")
	l.keys("4")
	l.has("3 with newer releases")
	l.keys("2")
	l.has("4 projects")
	l.keys("1")
	l.has("Fix (Enter)")
}

func TestTheDoctorTabOpensARemedyAndShowsItsKeysInTheFooter(t *testing.T) {
	l := newLab(t, true, 170, 30)
	l.keys("d")
	l.has("× go.work lists every module", "fix all (F)", "reload view (r)")
	l.hasNot("[↻", "Re-check")
	// The list is one line per check; the detail beside it tells the selected one in full.
	l.has("2 failed", "What is wrong", "libs/x is missing", "How to fix", "go work use ./libs/x", "2 checks passed")
	l.hasNot("git is installed")
}

func TestTheDoctorTabShowsThePassedChecksOnRequest(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("d")
	l.hasNot("git is installed")
	l.keys("p")
	l.has("git is installed", "orders has a lockfile")
}

func TestTheReleaseDialogListsThePackagesItPublishes(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("5", "x", "enter")
	l.has("Packages and versions to publish:", "web 1.0.0 → 1.1.0", "kit 0.2.0 → 0.2.1", "registry", "It cannot")
}

func TestThePreviewStaysOnTheReleaseScreenAndSaysWhatItWaitsFor(t *testing.T) {
	l := newLab(t, true, 120, 34)
	p := &scripted{lines: []string{"web 1.0.0 -> 1.1.0 (dry run)"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("5", "p")
	if l.sh().tab != "release" {
		t.Fatalf("the preview left the Release screen for %q", l.sh().tab)
	}
	l.has("Preview: smith release --dry-run", "registry", "Ctrl+C stops it")
	l.keys("ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("Preview: smith release --dry-run", "full output in the run panel of Projects")
}

func TestReleasePreviewNeedsNoConfirmationButReleaseDoes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("5", "p")
	if l.lastStarted() != "release --dry-run" {
		t.Fatalf("preview started %q", l.lastStarted())
	}
	l.hasNot("The command that will run:")
	l = newLab(t, true, 120, 30)
	l.keys("5", "x")
	l.has("Release: hold back or bump", "Continue (Enter)")
	l.keys("enter")
	l.has("Release", "$ smith release --yes", "Release anyway (y)", "Cancel (Esc)")
	if len(l.started) != 0 {
		t.Fatal("release must wait for the dialog")
	}
	l.keys("y")
	if l.lastStarted() != "release --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestEveryWorkspaceChangingCommandAsksFirst(t *testing.T) {
	fromPalette := func(name string) []func(*lab) {
		return []func(*lab){func(l *lab) { l.keys("ctrl+p").typeText(name).keys("enter") }}
	}
	for name, steps := range map[string][]func(*lab){
		"install": {func(l *lab) { l.keys("i") }},
		"ci":      fromPalette("Regenerate CI"),
		"upgrade": fromPalette("Upgrade workspace"),
		"release": {func(l *lab) { l.keys("5", "x", "enter") }},
	} {
		l := newLab(t, true, 120, 30)
		for _, step := range steps {
			step(l)
		}
		l.has("The command that will run:")
		if len(l.started) != 0 {
			t.Fatalf("%s ran without a confirmation", name)
		}
	}
}

func TestTheGuideAddsAProjectThroughTheConfirmationDialog(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("a") // Add project…
	l.has("Add a project", "What are you building?")
	// Enter on the language keeps the default: what most projects here use (Go).
	l.keys("down", "down", "enter", "enter").typeText("toolkit").keys("enter")
	l.has("How do you want to set it up?", "Start with the defaults", "Customise")
	l.keys("down", "enter") // Customise
	// Four projects are enough to follow: the form asks, and Enter keeps the default (match them).
	l.has("Should it do what the other projects here do?")
	l.keys("enter")
	// The plan comes first; nothing is written until the person says yes.
	l.has("Add toolkit?", "Nothing is written yet", "The command that will run:", "$ smith add package toolkit --lang go")
	if len(l.started) != 0 {
		t.Fatalf("started before the answer: %v", l.started)
	}
	l.keys("y")
	if l.lastStarted() != "add package toolkit --lang go --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestOutsideAWorkspaceNStartsANewWorkspaceAndCtrlNMeansNothing(t *testing.T) {
	// Issue #472: Ctrl+N was "new workspace" here and "add project" inside.
	l := newLab(t, false, 120, 30)
	l.keys("ctrl+n")
	l.hasNot("What should the new workspace be called?")
	l.hasNot("add project (a)")
	l.has("new workspace (n)")
	l.keys("n")
	l.has("Create a workspace", "What should the new workspace be called?")
	l.typeText("acme").keys("enter")
	// The first question: the defaults (Enter), or the questions (Customise).
	l.has("How do you want to set it up?", "Start with the defaults", "CI = where your tests run on every push")
	l.keys("down", "enter")
	// The CI, the main language, the code layout, the agent, the dev container and "more options": Enter keeps each default.
	l.has("Which CI should smith write pipelines for?")
	l.keys("enter", "enter", "enter", "enter", "enter", "enter")
	l.has("$ smith new acme --yes")
}

func TestStartingWithTheDefaultsNeedsOnlyTheName(t *testing.T) {
	l := newLab(t, false, 120, 30)
	l.keys("n").typeText("acme").keys("enter", "enter")
	l.has("$ smith new acme --yes")
}

func TestInsideAWorkspaceNDoesNotStartAWorkspaceAndAAddsAProject(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("n")
	l.hasNot("Create a workspace", "What should the new workspace be called?")
	l.hasNot("new workspace (n)")
	l.keys("a")
	l.has("Add a project")
}

func TestTheKeysAndThePaletteShareOneKeymap(t *testing.T) {
	// A key and its palette entry build the same command line, and the palette
	// shows the key the keymap gives.
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p").typeText("Build")
	l.has("Build", "b")
	l.keys("enter")
	if l.lastStarted() != "run build -p orders" {
		t.Fatalf("palette started %q", l.lastStarted())
	}
}

func TestStartingASecondCommandWhileOneRunsIsRefusedWithAReason(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.next = &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.keys("b")
	l.keys("t")
	l.has("Test is unavailable: a run is in progress (Ctrl+C stops it)")
	if len(l.started) != 1 {
		t.Fatalf("started %v", l.started)
	}
}

// mustDecode returns the text an OSC 52 sequence copies.
func mustDecode(t *testing.T, seq string) []byte {
	t.Helper()
	body, ok := strings.CutPrefix(seq, "\x1b]52;c;")
	body, ok2 := strings.CutSuffix(body, "\a")
	if !ok || !ok2 {
		t.Fatalf("not an OSC 52 sequence: %q", seq)
	}
	text, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		t.Fatal(err)
	}
	return text
}

// U and S on the Deps tab open the guided forms of smith deps up and smith
// sync, the same forms the Start tab lists, and Esc returns to the tab.
func TestUAndSOnTheDepsTabOpenTheFormsOfDepsUpAndSync(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	sh := l.sh()
	for _, a := range []guide.Action{
		{Title: "Update the dependencies", Presents: "deps up", Steps: []guide.Step{{Key: "arg:names", Prompt: "names", Optional: true}}},
		{Title: "Sync the versions", Presents: "sync", Steps: []guide.Step{{Key: "arg:names", Prompt: "names", Optional: true}}},
	} {
		sh.guide = sh.guide.WithAction(a)
	}
	l.m = sh
	for _, key := range []string{"U", "S"} {
		title := map[string]string{"U": "Update the dependencies", "S": "Sync the versions"}[key]
		l.openTab(commands.TabDeps)
		l.keys(key)
		if got := l.sh(); got.tab != commands.TabGuide || got.guide.Title() != title {
			t.Fatalf("%s opens %q on %q", key, got.guide.Title(), got.tab)
		}
		l.keys("esc")
		if got := l.sh().tab; got != commands.TabDeps {
			t.Fatalf("Esc returns to the Deps tab, not %q", got)
		}
	}
}

// Without the form (a smith that has no such command) the key says so.
func TestSOnTheDepsTabSaysWhenThereIsNoSyncForm(t *testing.T) {
	l := newHomeLab(t, 140, 40)
	l.openTab(commands.TabDeps)
	l.keys("S")
	l.has("There is no form for smith sync here")
	if got := l.sh().tab; got != commands.TabDeps {
		t.Fatalf("the tab stays: %q", got)
	}
}
