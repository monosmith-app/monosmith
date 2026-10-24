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
	if first := l.lines()[0]; !strings.Contains(first, "Workspace") || !strings.Contains(first, "Projects") {
		t.Fatalf("the panes start on the first row, with no menu bar above: %q", first)
	}
	footer := l.lines()[len(l.lines())-2]
	for _, want := range []string{"b build", "t test", "l lint", "c typecheck", "w changed", "d doctor", "a add project", "y copy", ": palette", "? help", "q quit", "m mouse"} {
		if !strings.Contains(footer, want) {
			t.Fatalf("the footer lacks %q: %q", want, footer)
		}
	}
	l.keys("3")
	footer = l.lines()[len(l.lines())-2]
	for _, want := range []string{"f fix", "x fix all", "g group", "r reload", "q quit"} {
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

func TestEnterOnATaskRunsItAndDRechecksTheWorkspace(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("shift+tab", "right", "down", "enter") // orders > build
	if l.lastStarted() != "run build -p orders" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l = newLab(t, true, 120, 30)
	l.keys("d")
	l.has("2 passed")
}

func TestJAndKMoveLikeTheArrows(t *testing.T) {
	l := newLab(t, true, 140, 34)
	l.keys("shift+tab")
	l.keys("j")
	if got := l.sh().tree.Selected().Project; got != "web" {
		t.Fatalf("j moved to %q", got)
	}
	l.keys("k")
	if got := l.sh().tree.Selected().Project; got != "orders" {
		t.Fatalf("k moved to %q", got)
	}
}

func TestKeysDoNotFireWhileATextFieldIsEditing(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("8", "down", "enter") // ci.providers: a text field opens
	l.typeText("bq?:1/ mw")
	if len(l.started) != 0 || l.quit {
		t.Fatalf("typing ran something: %v quit %v", l.started, l.quit)
	}
	if l.sh().tab != "settings" || l.sh().palOpen || l.sh().help != nil {
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
	l.keys("6", "a") // Deps: a marks all, it does not open the Add a project form
	if l.sh().tab != "deps" {
		t.Fatalf("a left Deps for %q", l.sh().tab)
	}
	l.keys("7", "c")
	if len(l.started) != 0 || l.sh().tab != "recipes" {
		t.Fatalf("c in Recipes is create, not typecheck: %v", l.started)
	}
}

func TestCtrlPOpensThePaletteListingEveryCommandWithItsShortcut(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p")
	l.has("Command palette", "Build", "Doctor")
	l.typeText("typec")
	l.has("Typecheck")
	l.hasNot("Quit")
	l.keys("esc", "ctrl+p").typeText("affect") // smith's own commands are listed too
	l.has("affected", "List projects a change touches")
	l.keys("esc")
	l.hasNot("Command palette")
}

func TestThePaletteRunsAShellCommandAndASmithCommand(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p").typeText("doctor").keys("enter") // the shell's Doctor comes first
	l.has("2 passed")
	l.keys("ctrl+p").typeText("affected").keys("enter") // a smith command: it asks for arguments, then runs
	l.has("smith affected")
	l.keys("enter")
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
	l.keys("9", "esc")
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
	l.keys("9", "down", "enter")
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
	l.keys("1", "y")
	if out.Len() == 0 || !strings.Contains(string(mustDecode(t, out.String())), "/") {
		t.Fatalf("projects copies the project root: %q", out.String())
	}

	out.Reset()
	l.keys("6", "y")
	got := string(mustDecode(t, out.String()))
	if !slices.ContainsFunc(sampleDeps.Rows, func(r depsview.Row) bool { return r.Name == got }) {
		t.Fatalf("deps copies the package: %q", got)
	}

	out.Reset()
	l.keys("b", "2", "y")
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
	l := newLab(t, true, 100, 40)
	l.keys("f1")
	l.has("Keyboard shortcuts", "Move in a list or tree", "Stop", "Ctrl+C")
	l.keys("end")
	l.has(": / Ctrl+P", "? / F1", "Quit", "Esc", "Mouse", "Double-click", "footer")
	l.hasNot("Right-click", "Alt+", "F10", "menu")
	l.keys("esc")
	l.hasNot("Keyboard shortcuts")
	l.keys("?")
	l.has("Keyboard shortcuts")
}

func TestTabMovesTheFocusBetweenPanesAndEveryPaneHasKeys(t *testing.T) {
	l := newLab(t, true, 140, 34)
	if l.sh().focus.String() != "main" {
		t.Fatalf("focus %v", l.sh().focus)
	}
	l.keys("shift+tab")
	if l.sh().focus.String() != "navigator" {
		t.Fatalf("focus %v", l.sh().focus)
	}
	l.keys("down") // orders → web
	l.has("Depends on")
	l.keys("tab", "tab")
	if l.sh().focus.String() != "inspector" {
		t.Fatalf("focus %v", l.sh().focus)
	}
	l.keys("down", "enter") // the second action: Test, on web
	if l.lastStarted() != "run test -p web" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestNumberedShortcutsSwitchTabs(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("3")
	l.has("2 passed", "1 failed")
	l.keys("4")
	l.has("1.0.0 →", "1.1.0", "unchanged: sdk")
	l.keys("8")
	l.has("ci.devcontainer", "default: main")
	l.keys("2")
	l.has("Nothing has run yet")
	l.keys("1")
	l.has("Projects (4)")
}

func TestTheDoctorTabOpensARemedyAndShowsItsKeysInTheFooter(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("d")
	l.has("✖ go.work lists every module", "f fix", "x fix all", "r reload")
	l.hasNot("[↻", "Re-check")
	// The problem and its remedy are inline, without opening the row.
	l.has("→ go work use ./libs/x", "libs/x is missing", "2 checks passed")
	l.hasNot("git is installed")
}

func TestTheDoctorTabShowsThePassedChecksOnRequest(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("d")
	l.hasNot("git is installed")
	l.keys("p")
	l.has("git is installed", "orders has a lockfile")
}

func TestTheFixDialogsListTheFilesOrThePackagesTheyTouch(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("3", "down", "f")
	l.has("Files that will change:", "writes apps/web/Dockerfile")
	l.keys("esc", "4", "x")
	l.has("Packages and versions to publish:", "web 1.0.0 → 1.1.0", "kit 0.2.0 → 0.2.1", "registry", "It cannot")
}

func TestThePreviewStaysOnTheReleaseScreenAndSaysWhatItWaitsFor(t *testing.T) {
	l := newLab(t, true, 120, 34)
	p := &scripted{lines: []string{"web 1.0.0 -> 1.1.0 (dry run)"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("4", "p")
	if l.sh().tab != "release" {
		t.Fatalf("the preview left the Release screen for %q", l.sh().tab)
	}
	l.has("Preview: smith release --dry-run", "registry", "Ctrl+C stops it")
	l.keys("ctrl+c")
	l.send(outputMsg{proc: p, exit: &runner.Exit{Code: 1, Cancelled: true}})
	l.has("Preview: smith release --dry-run", "full output on the Run screen")
}

func TestFixingAsksFirstWithTheCommandLine(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("3", "x")
	l.has("Fix with doctor", "The command that will run:", "$ smith doctor --fix --yes")
	if len(l.started) != 0 {
		t.Fatal("nothing runs before the answer")
	}
	l.keys("esc")
	l.hasNot("The command that will run:")
	l.has("was not run")
	l.keys("x", "y")
	if l.lastStarted() != "doctor --fix --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestReleasePreviewNeedsNoConfirmationButReleaseDoes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("4", "p")
	if l.lastStarted() != "release --dry-run" {
		t.Fatalf("preview started %q", l.lastStarted())
	}
	l.hasNot("The command that will run:")
	l = newLab(t, true, 120, 30)
	l.keys("4", "x")
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
		"fix all": {func(l *lab) { l.keys("3", "x") }},
		"ci":      fromPalette("Regenerate CI"),
		"upgrade": fromPalette("Upgrade workspace"),
		"release": {func(l *lab) { l.keys("4", "x") }},
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

func TestCommandsThatNeedATerminalHandOverToIt(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p").typeText("mcp").keys("enter", "enter")
	if got := strings.Join(l.sh().Argv, " "); got != "mcp" || !l.quit {
		t.Fatalf("argv %q quit %v", got, l.quit)
	}
	if len(l.started) != 0 {
		t.Fatal("it must not run inside the shell")
	}
}

func TestTheSettingsTabEditsThroughAConfirmedConfigSet(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("8", "enter") // a yes/no setting flips
	l.has("$ smith config set ci.devcontainer false")
	l.keys("y")
	if l.lastStarted() != "config set ci.devcontainer false" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l = newLab(t, true, 120, 30)
	l.keys("8", "down", "enter")
	l.has("accepts  any of github, azure")
	// While typing, the letters belong to the text field, not to the shell.
	l.typeText("bq?:1/").keys("backspace", "backspace", "backspace", "backspace", "backspace", "backspace")
	l.hasNot("The command that will run:")
	l.typeText(" azure").keys("enter")
	l.has("$ smith config set ci.providers github azure")
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
