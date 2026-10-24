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

func TestF10OpensTheMenuBarOverThePanesWithoutPushingThemDown(t *testing.T) {
	l := newLab(t, true, 120, 30)
	before := len(l.lines())
	l.keys("f10")
	l.has("Add project…", "Ctrl+N", "Command palette…", "Ctrl+P", "Quit", "Q")
	if got := len(l.lines()); got != before {
		t.Fatalf("the menu changed the screen height: %d -> %d", before, got)
	}
	// The row of an entry still shows the pane underneath, right of the box.
	_, y := l.find("Command palette…")
	if row := l.lines()[y]; !strings.Contains(row, "Inspector") && !strings.Contains(row, "│") {
		t.Fatalf("the pane under the menu is gone: %q", row)
	}
	l.keys("esc")
	l.hasNot("Command palette…")
}

func TestAltAndALetterOpenAMenuAndEveryItemShowsItsShortcut(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("alt+r")
	l.has("Run selected task", "Ctrl+R", "Build", "Ctrl+B", "Test", "Ctrl+T", "Lint", "Ctrl+L", "Typecheck", "Ctrl+Y", "Stop", "Follow log", "Ctrl+F")
	l.keys("alt+e") // Release has the mnemonic E
	l.has("Preview next release", "F2", "Release…", "F3")
	l.keys("esc", "alt+h")
	l.has("Keyboard shortcuts", "F1")
}

func TestArrowsWalkTheMenuAndEnterRunsTheEntry(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("alt+w", "enter") // Doctor is the first entry of Workspace
	l.has("2 passed")
	l.hasNot("Fix with doctor…")
}

func TestADisabledMenuItemSaysWhyInTheStatusBar(t *testing.T) {
	l := newLab(t, false, 120, 30)
	l.keys("alt+r")
	l.has("Run selected task")
	l.has("unavailable: not in a workspace")
	l.keys("enter")
	l.has("not in a workspace")
	if len(l.started) != 0 {
		t.Fatalf("a disabled item must not run: %v", l.started)
	}
}

func TestShortcutsExplainWhenTheyCannotRun(t *testing.T) {
	l := newLab(t, false, 100, 28)
	l.keys("ctrl+b")
	l.has("Build is unavailable: not in a workspace")
	l = newLab(t, true, 100, 28)
	l.keys("ctrl+r")
	l.has("Run selected task is unavailable: select a task in the navigator first")
}

func TestCtrlPOpensThePaletteListingEveryCommandWithItsShortcut(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+p")
	l.has("Command palette", "Build", "Ctrl+B", "Doctor", "Ctrl+D", "Quit", "Ctrl+C")
	l.typeText("typec")
	l.has("Typecheck", "Ctrl+Y")
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
	l.keys("alt+6", "esc")
	if l.quit {
		t.Fatal("esc in the guide must not quit")
	}
}

func TestQDoesNotQuitWhileARunGoesNorWhileTyping(t *testing.T) {
	l := newLab(t, true, 100, 28)
	l.next = &scripted{hold: true, release: make(chan struct{})}
	l.keys("ctrl+b", "q")
	if l.quit {
		t.Fatal("q must not quit while a run goes")
	}
	l.has("Ctrl+C stops it", "ctrl+c stop")
	l.keys("ctrl+c")

	l = newLab(t, true, 100, 28)
	l.keys("ctrl+p").typeText("quit")
	if l.quit {
		t.Fatal("q typed in the palette must not quit")
	}
}

func TestQInsideAGuideFormGoesBackInsteadOfQuitting(t *testing.T) {
	l := newLab(t, true, 100, 28)
	l.keys("alt+6", "down", "enter")
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

	l.keys("alt+3")
	l.keys("y")
	want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte("go work use ./libs/x")) + "\a"
	if out.String() != want {
		t.Fatalf("doctor copies the remedy: %q", out.String())
	}
	l.has("copied: go work use ./libs/x")

	out.Reset()
	l.keys("alt+1", "y")
	if out.Len() == 0 || !strings.Contains(string(mustDecode(t, out.String())), "/") {
		t.Fatalf("projects copies the project root: %q", out.String())
	}

	out.Reset()
	l.keys("alt+8", "y")
	got := string(mustDecode(t, out.String()))
	if !slices.ContainsFunc(sampleDeps.Rows, func(r depsview.Row) bool { return r.Name == got }) {
		t.Fatalf("deps copies the package: %q", got)
	}

	out.Reset()
	l.keys("ctrl+b", "alt+2", "y")
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
	l.has("Keyboard shortcuts", "Ctrl+P", "Ctrl+C")
	l.keys("end")
	l.has("F10 / Alt+letter", "Esc", "Mouse", "Double-click", "Right-click")
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
	l.keys("alt+3")
	l.has("2 passed", "1 failed")
	l.keys("alt+4")
	l.has("1.0.0 →", "1.1.0", "unchanged: sdk")
	l.keys("alt+5")
	l.has("ci.devcontainer", "default: main")
	l.keys("alt+2")
	l.has("Nothing has run yet")
	l.keys("alt+1")
	l.has("Projects (4)")
}

func TestTheDoctorTabOpensARemedyAndHasItsButtons(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+d")
	l.has("✖ go.work lists every module", "Re-check", "Ctrl+D", "Fix this…", "Fix all…", "Ctrl+X")
	// The problem and its remedy are inline, without opening the row.
	l.has("→ go work use ./libs/x", "libs/x is missing", "2 checks passed")
	l.hasNot("git is installed")
}

func TestTheDoctorTabShowsThePassedChecksOnRequest(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+d")
	l.hasNot("git is installed")
	l.keys("p")
	l.has("git is installed", "orders has a lockfile")
}

func TestTheFixDialogsListTheFilesOrThePackagesTheyTouch(t *testing.T) {
	l := newLab(t, true, 120, 34)
	l.keys("alt+3", "down", "f")
	l.has("Files that will change:", "writes apps/web/Dockerfile")
	l.keys("esc", "alt+4", "f3")
	l.has("Packages and versions to publish:", "web 1.0.0 → 1.1.0", "kit 0.2.0 → 0.2.1", "registry", "It cannot")
}

func TestThePreviewStaysOnTheReleaseScreenAndSaysWhatItWaitsFor(t *testing.T) {
	l := newLab(t, true, 120, 34)
	p := &scripted{lines: []string{"web 1.0.0 -> 1.1.0 (dry run)"}, hold: true, release: make(chan struct{})}
	l.next = p
	l.keys("alt+4", "f2")
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
	l.keys("ctrl+x")
	l.has("Fix with doctor", "The command that will run:", "$ smith doctor --fix --yes")
	if len(l.started) != 0 {
		t.Fatal("nothing runs before the answer")
	}
	l.keys("esc")
	l.hasNot("The command that will run:")
	l.has("was not run")
	l.keys("ctrl+x", "y")
	if l.lastStarted() != "doctor --fix --yes" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestReleasePreviewNeedsNoConfirmationButReleaseDoes(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("f2")
	if l.lastStarted() != "release --dry-run" {
		t.Fatalf("preview started %q", l.lastStarted())
	}
	l.hasNot("The command that will run:")
	l = newLab(t, true, 120, 30)
	l.keys("f3")
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
	for _, keys := range [][]string{{"ctrl+g"}, {"f4"}, {"ctrl+x"}} {
		l := newLab(t, true, 120, 30)
		for _, k := range keys {
			l.keys(k)
		}
		l.has("The command that will run:")
		if len(l.started) != 0 {
			t.Fatalf("%v ran without a confirmation", keys)
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
	l.keys("alt+5", "enter") // a yes/no setting flips
	l.has("$ smith config set ci.devcontainer false")
	l.keys("y")
	if l.lastStarted() != "config set ci.devcontainer false" {
		t.Fatalf("started %q", l.lastStarted())
	}
	l = newLab(t, true, 120, 30)
	l.keys("alt+5", "down", "enter")
	l.has("accepts  any of github, azure")
	// While typing, Ctrl+A and Ctrl+F belong to the text field, not to the shell.
	l.keys("ctrl+a", "ctrl+f", "ctrl+e")
	l.hasNot("The command that will run:")
	l.typeText(" azure").keys("enter")
	l.has("$ smith config set ci.providers github azure")
}

func TestTheGuideAddsAProjectThroughTheConfirmationDialog(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.keys("ctrl+n") // Add project…
	l.has("Add a project", "What kind?")
	l.keys("down", "down", "enter", "down", "enter").typeText("sdk").keys("enter")
	l.has("The command that will run:", "$ smith add package sdk --lang node")
	l.keys("y")
	if l.lastStarted() != "add package sdk --lang node" {
		t.Fatalf("started %q", l.lastStarted())
	}
}

func TestOutsideAWorkspaceCtrlNStartsANewWorkspace(t *testing.T) {
	l := newLab(t, false, 120, 30)
	l.keys("ctrl+n")
	l.has("Create a workspace", "Workspace name")
	l.typeText("acme").keys("enter")
	l.has("$ smith new acme")
}

func TestTheRunButtonsAndMenusShareOneCatalog(t *testing.T) {
	// Each Run menu command, run from its shortcut with a project selected,
	// builds the same command line the menu would.
	cases := map[string]string{"ctrl+b": "run build -p orders", "ctrl+t": "run test -p orders", "ctrl+l": "run lint -p orders", "ctrl+a": "run lint test --affected"}
	for key, want := range cases {
		l := newLab(t, true, 120, 30)
		l.keys(key)
		if l.lastStarted() != want {
			t.Errorf("%s started %q, want %q", key, l.lastStarted(), want)
		}
	}
	l := newLab(t, true, 120, 30)
	l.keys("alt+r", "down", "enter") // Build, through the menu
	if l.lastStarted() != "run build -p orders" {
		t.Fatalf("menu started %q", l.lastStarted())
	}
}

func TestStartingASecondCommandWhileOneRunsIsRefusedWithAReason(t *testing.T) {
	l := newLab(t, true, 120, 30)
	l.next = &scripted{lines: []string{"kit:build ▶ go build"}, hold: true, release: make(chan struct{})}
	l.keys("ctrl+b")
	l.keys("ctrl+t")
	l.has("Test is unavailable: a run is in progress")
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
