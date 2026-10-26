package tui

import (
	"bytes"
	"encoding/base64"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/depsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
)

func TestTheScreenHasNoMenuBarAndTheFooterShowsTheKeysOfTheOpenView(t *testing.T) {
	l := newLab(t, true, 160, 30)
	l.hasNot("File", "Workspace  Run", "Alt+")
	if first := l.lines()[0]; !strings.Contains(first, "acme") || !strings.Contains(first, "main") {
		t.Fatalf("the first row names the workspace and its branch, with no menu bar above: %q", first)
	}
	if tabs := l.lines()[2]; !strings.Contains(tabs, "Home") || !strings.Contains(tabs, "Projects") || !strings.Contains(tabs, "Doctor") {
		t.Fatalf("the views are tabs under the header: %q", tabs)
	}
	footer := l.lines()[len(l.lines())-2]
	for _, want := range []string{"b build", "t test", "l lint", "c typecheck", "w changed", "d doctor", "a add project", ": palette", "? help", "q quit", "m mouse"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the footer lacks %q: %q", want, footer)
		}
	}
	l.keys("4")
	footer = l.lines()[len(l.lines())-2]
	for _, want := range []string{"F fix all", "g group", "r reload", "q quit"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the Doctor footer lacks %q: %q", want, footer)
		}
	}
	if strings.Contains(footer, "w changed") {
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
	l.has("Add a project", "What kind?")
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
	l.keys("5", "a") // Deps: a marks all, it does not open the Add a project form
	if l.sh().tab != "deps" {
		t.Fatalf("a left Deps for %q", l.sh().tab)
	}
}

func TestCtrlPOpensThePaletteListingEveryCommandWithItsShortcut(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p")
	l.has("Command palette", "Build", "Build orders", "smith run build -p orders", "Build all 4")
	l.typeText("typec")
	l.has("Typecheck")
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
	l := newLab(t, true, 100, 28)
	l.next = &scripted{hold: true, release: make(chan struct{})}
	l.keys("b", "q")
	if l.quit {
		t.Fatal("q must not quit while a run goes")
	}
	l.has("Ctrl+C stops it", "Ctrl+C stop")
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

	l.keys("4")
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
	l.keys("5", "y")
	got := string(mustDecode(t, out.String()))
	if !slices.ContainsFunc(sampleDeps.Rows, func(r depsview.Row) bool { return r.Name == got }) {
		t.Fatalf("deps copies the package: %q", got)
	}

	out.Reset()
	l.keys("b", "3", "y")
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
	l.has("Keyboard shortcuts", "Move in a list", "Build all")
	l.keys("pgdown")
	l.has("Stop", "Ctrl+C")
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
	l.hasNot("Depends on")
	l.keys("down", "enter") // web
	if !l.sh().detailOpen {
		t.Fatal("Enter should open the detail")
	}
	l.has("Depends on", "Run build", "Run test")
	l.keys("esc")
	if l.sh().detailOpen {
		t.Fatal("Esc should go back to the list")
	}
	l.has("On: web")
	l.keys("enter", "down", "enter") // the second action: Run test, on web
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestAtTheSupportedFloorTheDetailStillOpensFullWidth(t *testing.T) {
	l := newLab(t, true, 72, 24)
	l.keys("enter")
	l.has("orders", "Run build", "Explain")
	l.keys("esc")
	l.has("On: orders", "Tasks")
	l.hasNot("Run build")
}

func TestNumberedShortcutsSwitchTabs(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("4")
	l.has("2 failed", "2 checks passed")
	l.keys("6")
	l.has("1.0.0 →", "1.1.0", "unchanged: sdk")
	l.keys("5")
	l.has("3 with newer releases")
	l.keys("3")
	l.has("Nothing has run yet")
	l.keys("2")
	l.has("Projects (4)")
	l.keys("1")
	l.has("Next:")
}

func TestTheDoctorTabOpensARemedyAndShowsItsKeysInTheFooter(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("d")
	l.has("✗ go.work lists every module", "F fix all", "r reload")
	l.hasNot("[↻", "Re-check")
	// The list is one line per check; the detail beside it tells the selected one in full.
	l.has("2 failed", "Problem", "libs/x is missing", "How to fix", "go work use ./libs/x", "2 checks passed")
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
	l.keys("6", "x")
	l.has("Packages and versions to publish:", "web 1.0.0 → 1.1.0", "kit 0.2.0 → 0.2.1", "registry", "It cannot")
}

func TestThePreviewStaysOnTheReleaseScreenAndSaysWhatItWaitsFor(t *testing.T) {
	l := newLab(t, true, 120, 34)
	p := &scripted{lines: []string{"web 1.0.0 -> 1.1.0 (dry run)"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("6", "p")
	if l.sh().tab != "release" {
		t.Fatalf("the preview left the Release screen for %q", l.sh().tab)
	}
	l.has("Preview: smith release --dry-run", "registry", "Ctrl+C stops it")
	l.keys("ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("Preview: smith release --dry-run", "full output on the Run screen")
}

func TestReleasePreviewNeedsNoConfirmationButReleaseDoes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("6", "p")
	if l.lastStarted() != "release --dry-run" {
		t.Fatalf("preview started %q", l.lastStarted())
	}
	l.hasNot("The command that will run:")
	l = newLab(t, true, 120, 30)
	l.keys("6", "x")
	l.has("Release", "$ smith release --yes", "Run  Enter", "Cancel  Esc")
	if len(l.started) != 0 {
		t.Fatal("release must wait for the dialog")
	}
	l.keys("enter")
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
		"release": {func(l *lab) { l.keys("6", "x") }},
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
	l.has("Add a project", "What kind?")
	l.keys("down", "down", "enter", "down", "enter").typeText("sdk").keys("enter")
	l.has("The command that will run:", "$ smith add package sdk --lang node")
	l.keys("y")
	if l.lastStarted() != "add package sdk --lang node" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestOutsideAWorkspaceNStartsANewWorkspaceAndCtrlNMeansNothing(t *testing.T) {
	// Issue #472: Ctrl+N was "new workspace" here and "add project" inside.
	l := newLab(t, false, 120, 30)
	l.keys("ctrl+n")
	l.hasNot("Workspace name")
	l.hasNot("a add project")
	l.has("n new workspace")
	l.keys("n")
	l.has("Create a workspace", "Workspace name")
	l.typeText("acme").keys("enter")
	l.has("$ smith new acme")
}

func TestInsideAWorkspaceNDoesNotStartAWorkspaceAndAAddsAProject(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("n")
	l.hasNot("Create a workspace", "Workspace name")
	l.hasNot("n new workspace")
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
