package lintfix

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

const crashOutput = `
Oops! Something went wrong! :(

ESLint: 9.39.3

Error: Key "rules": Key "unicorn/logical-assignment-operators": Value [null] should match exactly one schema in oneOf.
    at RuleValidator.validate (file:///x/node_modules/eslint/lib/config/rule-validator.js:1:1)
`

// fakeESLint answers per project folder: the report before the fixes, the
// report after them, or a crash.
type fakeESLint struct {
	before, after map[string]string
	crash         map[string]bool
	calls         []string
}

func (f *fakeESLint) run(dir string, argv []string) (string, string, int, error) {
	fixing := slices.Contains(argv, "--fix")
	f.calls = append(f.calls, dir+map[bool]string{true: " fix", false: " count"}[fixing])
	if f.crash[dir] {
		return "", crashOutput, 2, nil
	}
	if dir == "missing-node" {
		return "", "", 0, errors.New(`exec: "node": executable file not found in %PATH%`)
	}
	report := f.before[dir]
	if fixing {
		report = f.after[dir]
	}
	code := 0
	if report != "[]" {
		code = 1
	}
	return report, "", code, nil
}

func TestFixRunsEachProjectInItsOwnFolderAndACrashDoesNotStopTheOthers(t *testing.T) {
	f := &fakeESLint{
		before: map[string]string{
			"api":    `[{"errorCount": 10, "warningCount": 2}, {"errorCount": 3, "warningCount": 0}]`,
			"models": `[{"errorCount": 4, "warningCount": 0}]`,
		},
		after: map[string]string{
			"api":    `[{"errorCount": 5, "warningCount": 1}]`,
			"models": `[]`,
		},
		crash: map[string]bool{"app": true},
	}
	got := Fix(Options{Root: "/ws", Run: f.run, ESLintVersion: "10.12.0", Targets: []Target{{"api", "api"}, {"app", "app"}, {"models", "models"}}})
	if strings.Join(f.calls, ",") != "api count,api fix,app count,models count,models fix" {
		t.Fatalf("calls = %v", f.calls)
	}
	want := []Outcome{
		{Project: "api", Fixed: 9, Errors: 5, Warnings: 1},
		{Project: "app", Crash: `Error: Key "rules": Key "unicorn/logical-assignment-operators": Value [null] should match exactly one schema in oneOf.`, Remedy: "run smith run lint -p app to see the whole error"},
		{Project: "models", Fixed: 4},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("outcomes:\n got %+v\nwant %+v", got, want)
	}
}

func TestFixNamesTheRemedyForAnESLintOlderThanTheConfigNeeds(t *testing.T) {
	f := &fakeESLint{crash: map[string]bool{"app": true}}
	got := Fix(Options{Root: "/ws", Run: f.run, ESLintVersion: "9.39.3", Targets: []Target{{"app", "app"}}})
	if !strings.Contains(got[0].Remedy, `set "eslint": "`+removalplan.LintESLintRange+`" in package.json, then run npm install`) || !strings.Contains(got[0].Remedy, "has 9.39.3") {
		t.Fatalf("remedy = %q", got[0].Remedy)
	}
}

func TestFixWithoutAnInstalledESLintSaysToInstall(t *testing.T) {
	f := &fakeESLint{}
	got := Fix(Options{Root: "/ws", Run: f.run, Targets: []Target{{"app", "app"}, {"api", "api"}}})
	if len(f.calls) != 0 || len(got) != 2 || got[1].Crash != "ESLint is not installed in the workspace" || !strings.Contains(got[1].Remedy, "npm install") {
		t.Fatalf("calls %v, outcomes %+v", f.calls, got)
	}
}

func TestFixReportsAProgramThatCannotStart(t *testing.T) {
	got := Fix(Options{Root: "/ws", Run: (&fakeESLint{}).run, ESLintVersion: "10.12.0", Targets: []Target{{"web", "missing-node"}}})
	if !strings.Contains(got[0].Crash, "executable file not found") {
		t.Fatalf("outcome %+v", got[0])
	}
}
