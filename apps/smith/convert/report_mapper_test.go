package convert

import (
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func manyProjectsPlan(n int) *convertplan.Plan {
	plan := &convertplan.Plan{Config: workspace.Config{Workspace: workspace.WorkspaceSection{Name: "demo"}}}
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("pkg%02d", i)
		plan.Projects = append(plan.Projects, &convertplan.ProjectPlan{Name: name, Root: "packages/" + name, Language: "node", Kind: protocol.ProjectKind("package"), Supported: true})
		plan.Add(convertplan.Dropped, name, "watch-deps", "Nx's watch helper")
		plan.Add(convertplan.Mapped, name, "release tags", "continues its tags @acme/"+name+"@<version>")
		plan.Add(convertplan.Mapped, name, "build", "package.json script: rollup -c")
		plan.Add(convertplan.Changed, name, "build", "Rollup is replaced by esbuild")
		plan.Delete = append(plan.Delete, "packages/"+name+"/rollup.config.cjs")
	}
	return plan
}

func TestReportSaysTheSameThingOnceForAllProjects(t *testing.T) {
	out := Report(manyProjectsPlan(23), ReportOptions{DryRun: true})
	for _, want := range []string{
		"Ready to convert: 23 projects (23 packages). Nothing blocks the conversion.",
		"watch-deps: Nx's watch helper\n      23 projects: pkg00, pkg01, pkg02, pkg03, pkg04 and 18 more",
		"release tags: continues its tags @acme/<project>@<version>",
		"rollup.config.cjs in 23 folders",
		"→ smith convert",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report lacks %q:\n%s", want, out)
		}
	}
	if n := strings.Count(out, "Nx's watch helper"); n != 1 {
		t.Errorf("the message is printed %d times:\n%s", n, out)
	}
}

func TestReportStatesABuildOnceNotTwice(t *testing.T) {
	out := Report(manyProjectsPlan(2), ReportOptions{})
	if strings.Contains(out, "package.json script: rollup") {
		t.Fatalf("a build that becomes something else is also reported as kept:\n%s", out)
	}
	if !strings.Contains(out, "build: Rollup is replaced by esbuild") {
		t.Fatalf("%s", out)
	}
}

func TestReportPutsWhatNeedsADecisionFirst(t *testing.T) {
	plan := manyProjectsPlan(1)
	plan.Decide("", "azure-pipelines.yml", "replaced by the pipeline smith generates")
	plan.Decide("", "versions", "before 1.0 a feat: commit releases a minor version")
	plan.Add(convertplan.Manual, "", "script x", "runs Nx in a way smith cannot translate")
	out := Report(plan, ReportOptions{DryRun: true})
	if !strings.Contains(out, "3 things to check before you apply:\n  1. azure-pipelines.yml: replaced") {
		t.Fatalf("%s", out)
	}
	if strings.Count(out, "replaced by the pipeline smith generates") != 1 {
		t.Fatalf("a decision is repeated in the details:\n%s", out)
	}
	if strings.Index(out, "3 things to check") > strings.Index(out, "Behaves differently") {
		t.Fatalf("the decisions come after the details:\n%s", out)
	}
}

func TestReportSaysWhenSomethingBlocks(t *testing.T) {
	plan := manyProjectsPlan(1)
	plan.Add(convertplan.Blocker, "a", "packages/a", "a and b would both be called a")
	out := Report(plan, ReportOptions{DryRun: true})
	if !strings.Contains(out, "Cannot convert yet: 1 problem blocks the conversion.") || !strings.Contains(out, "Fix first:") || strings.Contains(out, "smith convert\n") {
		t.Fatalf("%s", out)
	}
}

func TestVerboseNamesEveryProject(t *testing.T) {
	out := Report(manyProjectsPlan(8), ReportOptions{Verbose: true})
	if !strings.Contains(out, "pkg07") || strings.Contains(out, "and 3 more") {
		t.Fatalf("%s", out)
	}
}
