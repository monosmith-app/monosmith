package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func names(ps []*Project) string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Name)
	}
	return strings.Join(out, ",")
}

func TestADeclaredLanguageListCoversOnlyThoseLanguages(t *testing.T) {
	covered, skipped := SplitByLanguage(projects(), []string{"go", "python"})
	if names(covered) != "tool,old,kit" || names(skipped) != "web" {
		t.Fatalf("covered %s, skipped %s", names(covered), names(skipped))
	}
}

func TestNoDeclaredLanguagesMeansEveryLanguage(t *testing.T) {
	covered, skipped := SplitByLanguage(projects(), nil)
	if names(covered) != "tool,old,kit,web" || len(skipped) != 0 {
		t.Fatalf("covered %s, skipped %s", names(covered), names(skipped))
	}
}

func TestAConventionThatDeclaresNoLanguagesTargetsEveryProject(t *testing.T) {
	got, err := ConventionTargets("anything", projects(), nil, workspace.ConventionConfig{Exclude: []string{"old"}})
	if err != nil || names(got) != "tool,kit,web" {
		t.Fatalf("%s %v", names(got), err)
	}
	got, err = ConventionTargets("anything", projects(), nil, workspace.ConventionConfig{Languages: []string{"node"}})
	if err != nil || names(got) != "web" {
		t.Fatalf("narrowing still works: %s %v", names(got), err)
	}
}

func TestCoverageLeavesOutWhatThePersonChoseToLeaveOut(t *testing.T) {
	cov := ConventionCoverage(projects(), []string{"go", "python"}, workspace.ConventionConfig{Exclude: []string{"kit"}})
	if names(cov.Checked) != "tool,old" || names(cov.Skipped) != "web" {
		t.Fatalf("checked %s, skipped %s", names(cov.Checked), names(cov.Skipped))
	}
	cov = ConventionCoverage(projects(), []string{"go", "python"}, workspace.ConventionConfig{Languages: []string{"go"}})
	if names(cov.Checked) != "tool,old" || len(cov.Skipped) != 0 {
		t.Fatalf("narrowing to go is a choice, not a gap: checked %s, skipped %s", names(cov.Checked), names(cov.Skipped))
	}
}
