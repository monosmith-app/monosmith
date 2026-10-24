package convert

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestSummaryEndsWithTheFilesToKeepAndTheNextSteps(t *testing.T) {
	plan := &convertplan.Plan{Config: workspace.Config{Workspace: workspace.WorkspaceSection{Name: "demo"}}}
	plan.KeepFile("jest.resolver.cjs", "stands in for the resolver @nx/jest provided")
	out := Summary(plan, []string{"ci.yml: build"}, []string{"jest.resolver.cjs", "smith.toml"})
	for _, want := range []string{
		"! pipeline slot ci.yml: build",
		"The full report is in .smith/convert-report.txt",
		"jest.resolver.cjs: stands in for the resolver @nx/jest provided",
		"Staged with git add so they cannot be lost: jest.resolver.cjs, smith.toml",
		"Next:\n  1. git diff",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary lacks %q:\n%s", want, out)
		}
	}
	if !strings.HasSuffix(out, "  4. commit\n") {
		t.Errorf("the next steps are not last:\n%s", out)
	}
}

func TestDryRunNamesTheFilesThatMustStay(t *testing.T) {
	plan := manyProjectsPlan(1)
	plan.KeepFile("jest.resolver.cjs", "the Jest preset loads it")
	if out := Report(plan, ReportOptions{DryRun: true}); !strings.Contains(out, "keep them and commit them") || !strings.Contains(out, "jest.resolver.cjs: the Jest preset loads it") {
		t.Fatalf("%s", out)
	}
}
