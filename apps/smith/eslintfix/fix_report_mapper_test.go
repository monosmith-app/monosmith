package eslintfix

import (
	"testing"
)

func TestReportListsEachProjectCrashesFirstThenTheTotalsAndTheRemedies(t *testing.T) {
	text, failed := Report([]Outcome{
		{Project: "models", Fixed: 231, Errors: 200},
		{Project: "lib", Fixed: 3},
		{Project: "app", Crash: "Error: Key \"rules\": broken", Remedy: "run smith run lint -p app to see the whole error"},
		{Project: "api", Fixed: 544, Errors: 870, Warnings: 6},
	})
	want := `Project   Fixed    Left  Result
app           -       -  × ESLint could not run: Error: Key "rules": broken
api         544     876  ▲ 870 errors, 6 warnings
models      231     200  ▲ 200 errors, 0 warnings
lib           3       0  ✓ no problems left
× app: ESLint could not run: Error: Key "rules": broken
→ run smith run lint -p app to see the whole error
▲ ESLint fixed 778 problems; 1076 are left in 2 of 3 projects
→ see them: smith run lint -p api
→ see them: smith run lint -p models
`
	if !failed || text != want {
		t.Fatalf("failed=%v, report:\n%s\nwant:\n%s", failed, text, want)
	}
}

func TestReportOfACleanWorkspaceIsNotAFailure(t *testing.T) {
	text, failed := Report([]Outcome{{Project: "api", Fixed: 1}})
	want := "Project   Fixed    Left  Result\napi           1       0  ✓ no problems left\n✓ ESLint fixed 1 problem; none are left in 1 project\n"
	if failed || text != want {
		t.Fatalf("failed=%v, report:\n%s", failed, text)
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
