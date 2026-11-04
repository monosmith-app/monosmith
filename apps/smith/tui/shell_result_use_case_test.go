package tui

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/runner"
	"github.com/monosmith-app/monosmith/apps/smith/tui/sources"
)

var failedAddOutput = []string{
	"Creating app api-cs in apps/api-cs: a console app",
	"Wrote 12 file(s):",
	"  apps/api-cs/Program.cs",
	"Verifying: smith run build test -p api-cs",
	"dotnet is not installed or not on PATH. Install it from https://dot.net/download, or run smith tools install to choose how, then run the command again.",
	"api-cs was created and its files are kept, but a tool it needs is missing.",
}

// failedAdd runs an add that ends because a tool is missing.
func failedAdd(t *testing.T) *lab {
	t.Helper()
	l := newLab(t, true, 120, 36)
	l.next = &scripted{lines: failedAddOutput, exit: runner.Exit{Code: 1}}
	sh := l.sh()
	cmd := sh.execute([]string{"add", "app", "api-cs", "--lang", "dotnet", "--yes"}, true)
	l.m = sh
	l.drive(cmd)
	return l
}

func TestAnAddThatCouldNotCheckForAMissingToolSaysTheProjectWasMadeAndOffersChoices(t *testing.T) {
	l := failedAdd(t)
	l.has("Add api-cs: it could not be checked", "api-cs was created in apps/api-cs (12 files).",
		"It could not be checked yet: the .NET SDK is not installed.", "Open the install page", "Install it with smith",
		"Check again", "Continue without checking", "Nothing is lost: your project stays on disk.")
}

func TestOpeningTheInstallPageCopiesTheLinkAndOpensIt(t *testing.T) {
	l := failedAdd(t)
	l.keys("enter")
	if len(l.opened) != 1 || l.opened[0] != "https://dot.net/download" {
		t.Fatalf("opened %v", l.opened)
	}
	l.has("Add api-cs: it could not be checked")
}

func TestCheckingAgainRunsTheFirstCheckOfTheProjectAndNotTheAdd(t *testing.T) {
	l := failedAdd(t)
	l.next = &scripted{lines: []string{"api-cs:build ▶ dotnet build"}, exit: runner.Exit{}}
	l.keys("down", "down", "enter")
	if got := l.lastStarted(); got != "run build test -p api-cs" {
		t.Fatalf("started %q", got)
	}
	l.hasNot("it could not be checked")
}

func TestInstallingWithSmithOpensTheToolsDialogAndComesBackToTheChoices(t *testing.T) {
	l := failedAdd(t)
	l.capture = func(argv []string) sources.Captured { return sources.Captured{OK: true, Text: toolPlansJSON(t)} }
	l.keys("down", "enter")
	l.has("Install the missing tools", "Go is missing")
	l.keys("esc")
	l.has("Add api-cs: it could not be checked", "Continue without checking")
}

func TestContinuingWithoutCheckingClosesTheDialogAndSaysDoctorKeepsReminding(t *testing.T) {
	l := failedAdd(t)
	l.keys("down", "down", "down", "enter")
	l.hasNot("Install it with smith")
	l.has("Doctor will keep reminding you")
}

func TestEscClosesTheFailureDialogAndRunsNothing(t *testing.T) {
	l := failedAdd(t)
	before := len(l.started)
	l.keys("esc")
	l.hasNot("Install it with smith")
	if len(l.started) != before {
		t.Fatalf("esc started %v", l.started[before:])
	}
}

func TestAFailedAddThatIsNotAMissingToolKeepsTheRunScreen(t *testing.T) {
	l := newLab(t, true, 120, 36)
	l.next = &scripted{lines: []string{"Wrote 3 file(s):", "the new project web fails its first build"}, exit: runner.Exit{Code: 1}}
	sh := l.sh()
	cmd := sh.execute([]string{"add", "app", "web", "--yes"}, true)
	l.m = sh
	l.drive(cmd)
	l.hasNot("Continue without checking")
}

func TestRunAgainIsNotOfferedForAnAddAndSaysWhy(t *testing.T) {
	l := failedAdd(t)
	l.keys("esc")
	l.hasNot("Run again", "r run again")
	l.keys("r")
	l.has("makes or removes something")
	if got := l.lastStarted(); strings.HasPrefix(got, "add ") && strings.Count(strings.Join(joinAll(l.started), "\n"), "add app api-cs") > 1 {
		t.Fatalf("the add ran twice: %v", l.started)
	}
}

func joinAll(argvs [][]string) []string {
	out := make([]string, len(argvs))
	for i, a := range argvs {
		out[i] = strings.Join(a, " ")
	}
	return out
}

// uncommittedAdd runs an add that finishes and says it did not commit.
func uncommittedAdd(t *testing.T, kind, name string) *lab {
	t.Helper()
	l := newLab(t, true, 120, 36)
	l.next = &scripted{lines: []string{"Creating " + kind + " " + name + " in libs/" + name + ": go", "Wrote 4 file(s):", "Done: " + kind + " " + name + " is ready (4 files written, 1s).",
		"Not committed. Review it with git status; tick \"Commit this change\" next time to commit it as one operation that smith undo can revert."}, exit: runner.Exit{}}
	sh := l.sh()
	cmd := sh.execute([]string{"add", kind, name, "--lang", "go", "--yes"}, true)
	l.m = sh
	l.drive(cmd)
	return l
}

func TestAnUncommittedChangeOffersToCommitItAndStartsOnLeavingItAlone(t *testing.T) {
	l := uncommittedAdd(t, "lib", "core")
	l.has("Added lib core", "core was created in libs/core (4 files).", "Use it from an app: press + on the app", "smith install -p <app>",
		"Leave them uncommitted", "Commit only the files this operation changed", "Commit everything git shows as changed")
	l.keys("enter")
	l.hasNot("Commit these changes now")
	l.has("Not committed: review with git status")
	for _, argv := range l.started {
		if argv[0] == "commit" {
			t.Fatalf("smith committed unasked: %v", argv)
		}
	}
}

func TestChoosingToCommitShowsTheFilesThenRunsOneCommitNamedFromTheOperation(t *testing.T) {
	l := uncommittedAdd(t, "app", "web")
	l.answers(map[string]sources.Captured{"commit add app web --only-operation --dry-run": {OK: true, Text: "Plan: commit add app web\n  create  apps/web/package.json"}})
	l.keys("down", "enter")
	if got := l.lastStarted(); strings.HasPrefix(got, "commit") {
		t.Fatalf("it must ask first: %q", got)
	}
	l.has("Commit these changes", "apps/web/package.json", "smith commit add app web --only-operation")
	l.keys("enter")
	if got := l.lastStarted(); got != "commit add app web --only-operation --yes" {
		t.Fatalf("started %q", got)
	}
}

func TestAnAppResultDoesNotTellYouToUseItFromAnApp(t *testing.T) {
	l := uncommittedAdd(t, "app", "web")
	l.has("Added app web")
	l.hasNot("Use it from an app")
}
