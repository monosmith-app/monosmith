package convert

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func manyFilesPlan() *convertplan.Plan {
	plan := &convertplan.Plan{Config: workspace.Config{Workspace: workspace.WorkspaceSection{Name: "demo"}}}
	for i := 0; i < 12; i++ {
		plan.Add(convertplan.Manual, "", fmt.Sprintf("libs/a%02d/index.ts", i), "imports Nx (@nx/devkit), which convert removes; rewrite it without Nx")
		plan.Add(convertplan.Manual, "", fmt.Sprintf("libs/a%02d/eslint.config.mjs", i), "uses @nx/eslint-plugin in a way convert cannot rewrite; replace it with typescript-eslint")
	}
	plan.Add(convertplan.Manual, "", "script x", "runs Nx in a way smith cannot translate")
	plan.Add(convertplan.Unsupported, "gradle", "libs/gradle", "JVM (Gradle) project: smith has no plugin for it, so nothing builds it")
	plan.Add(convertplan.Blocker, "ui", "apps/ui", "apps/ui and libs/ui would both be called \"ui\"")
	return plan
}

func TestReportGroupsTheFilesThatImportNxAndLintWithItsPlugin(t *testing.T) {
	out := Report(manyFilesPlan(), ReportOptions{DryRun: true})
	for _, want := range []string{
		"12 files import Nx, which convert removes; rewrite them without Nx",
		"12 ESLint configs use @nx/eslint-plugin in a way convert cannot rewrite",
		"12 files: libs/a00/index.ts, libs/a01/index.ts, libs/a02/index.ts, libs/a03/index.ts, libs/a04/index.ts and 7 more",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if n := strings.Count(out, "imports Nx (@nx/devkit)"); n != 0 {
		t.Errorf("a grouped message is printed once per file (%d):\n%s", n, out)
	}
}

func TestReportListsProjectsThatStopBeingBuiltBeforeManualWork(t *testing.T) {
	out := Report(manyFilesPlan(), ReportOptions{DryRun: true})
	unsupported := strings.Index(out, "smith has no plugin for it")
	manual := strings.Index(out, "runs Nx in a way smith cannot translate")
	if unsupported < 0 || manual < 0 || unsupported > manual {
		t.Errorf("the project smith cannot build must come first (%d, %d):\n%s", unsupported, manual, out)
	}
}

func TestReportEndsWithTheBlockersWhenThereIsMoreToRead(t *testing.T) {
	out := Report(manyFilesPlan(), ReportOptions{DryRun: true})
	tail := out[strings.LastIndex(out, "Fix first, then run this again:"):]
	if !strings.Contains(out, "Fix first, then run this again:") || !strings.Contains(tail, "would both be called") {
		t.Errorf("the blocker is not repeated at the end:\n%s", out)
	}
	if strings.Contains(out, "smith v1") {
		t.Errorf("the report names a smith version that no one knows:\n%s", out)
	}
}
