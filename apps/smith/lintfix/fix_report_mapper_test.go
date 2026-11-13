package lintfix

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
)

const crashOutput = `
Oops! Something went wrong! :(

ESLint: 9.39.3

Error: Key "rules": Key "unicorn/logical-assignment-operators": Value [null] should match exactly one schema in oneOf.
    at RuleValidator.validate (file:///x/node_modules/eslint/lib/config/rule-validator.js:1:1)
`

func TestReportListsEachProjectFailuresFirstThenTheRemediesAndTheClosingLine(t *testing.T) {
	text, failed := Report([]Outcome{
		{Project: "models", Fixed: 231, Counted: true, Errors: 200},
		{Project: "lib", Fixed: 3, Counted: true},
		{Project: "app", Crash: "Error: Key \"rules\": broken", Remedy: "run smith run lint -p app to see the whole error"},
		{Project: "api", Fixed: 544, Counted: true, Errors: 870, Warnings: 6},
		{Project: "svc", Fixed: -1, Left: true},
		{Project: "tool", Fixed: -1, NoFix: true},
	}, 3*time.Second)
	want := `Project   Fixed    Left  Result
app           -       -  × could not run: Error: Key "rules": broken
api         544     876  ▲ 870 errors, 6 warnings
models      231     200  ▲ 200 errors, 0 warnings
svc           -    some  ▲ problems left
lib           3       0  ✓ no problems left
tool          -       0  ✓ no problems left
× app: could not run: Error: Key "rules": broken
→ run smith run lint -p app to see the whole error
▲ No fix mode for tool: the plain lint ran, nothing was changed
× Lint fix found problems in 3s: 6 projects, 778 problems fixed, 1077 left in 3 projects
→ see them: smith run lint -p api
→ see them: smith run lint -p models
→ see them: smith run lint -p svc
`
	if !failed || text != want {
		t.Fatalf("failed=%v, report:\n%s\nwant:\n%s", failed, text, want)
	}
}

func TestReportOfACleanWorkspaceIsNotAFailure(t *testing.T) {
	text, failed := Report([]Outcome{{Project: "api", Fixed: 1, Counted: true}}, 2*time.Second)
	want := "Project   Fixed    Left  Result\napi           1       0  ✓ no problems left\n✓ Lint fix passed in 2s: 1 project, 1 problem fixed, none left\n"
	if failed || text != want {
		t.Fatalf("failed=%v, report:\n%s", failed, text)
	}
}

func TestReportSaysAProjectWasNotReached(t *testing.T) {
	text, failed := Report([]Outcome{{Project: "api", Fixed: -1, Skipped: true}}, time.Second)
	if !failed || !strings.Contains(text, "» not run: a task it needs failed") {
		t.Fatalf("failed=%v, report:\n%s", failed, text)
	}
}

func TestReadESLintCountsTheReportAndSeparatesACrash(t *testing.T) {
	errs, warns, crash := ReadESLint(`[{"errorCount": 10, "warningCount": 2}, {"errorCount": 3, "warningCount": 0}]`, "", 1)
	if errs != 13 || warns != 2 || crash != "" {
		t.Fatalf("got %d, %d, %q", errs, warns, crash)
	}
	if _, _, crash := ReadESLint("", crashOutput, 2); crash != `Error: Key "rules": Key "unicorn/logical-assignment-operators": Value [null] should match exactly one schema in oneOf.` {
		t.Fatalf("crash = %q", crash)
	}
	if _, _, crash := ReadESLint("", "", 2); crash != "ESLint exited with code 2" {
		t.Fatalf("crash = %q", crash)
	}
}

func TestFirstErrorLineSkipsESLintsBanner(t *testing.T) {
	if got := firstErrorLine(crashOutput); got != `Error: Key "rules": Key "unicorn/logical-assignment-operators": Value [null] should match exactly one schema in oneOf.` {
		t.Fatalf("got %q", got)
	}
	if got := firstErrorLine("\nCannot find package 'x'\n"); got != "Cannot find package 'x'" {
		t.Fatalf("got %q", got)
	}
}

func TestESLintRemedyNamesTheFixForEachCause(t *testing.T) {
	if got := ESLintRemedy("app", ""); !strings.Contains(got, "npm install") {
		t.Fatalf("not installed: %q", got)
	}
	old := ESLintRemedy("app", "9.39.3")
	if !strings.Contains(old, `set "eslint": "`+eslintpeer.Range+`" in package.json, then run npm install`) || !strings.Contains(old, "has 9.39.3") {
		t.Fatalf("old: %q", old)
	}
	if got := ESLintRemedy("app", "10.12.0"); got != "run smith run lint -p app to see the whole error" {
		t.Fatalf("current: %q", got)
	}
}
