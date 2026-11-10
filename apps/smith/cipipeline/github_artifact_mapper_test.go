package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func artifactModel() Model {
	m := model()
	m.Pack = true
	m.Native = true
	return m
}

func TestTheArtifactStepsKeepTheDefaultsUntilTheWorkspaceAsksOtherwise(t *testing.T) {
	out := Rendered(workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}, artifactModel())[GitHubPath]
	if !strings.Contains(out, "name: drop") || !strings.Contains(out, "name: native-${{ matrix.os }}") {
		t.Fatalf("both upload steps are there:\n%s", out)
	}
	if strings.Contains(out, "retention-days") || strings.Contains(out, "continue-on-error") {
		t.Errorf("by default an upload fails the job and uses the repository's retention:\n%s", out)
	}
}

func TestTheArtifactStepsContinueOnErrorAndSetTheRetention(t *testing.T) {
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}, Artifacts: "optional", ArtifactRetentionDays: 7}}
	out := Rendered(cfg, artifactModel())[GitHubPath]
	for _, want := range []string{
		"        continue-on-error: true\n        with:\n          name: drop\n          path: dist/drop\n          if-no-files-found: ignore\n          retention-days: 7\n",
		"        continue-on-error: true\n        with:\n          name: native-${{ matrix.os }}\n          path: dist/drop\n          if-no-files-found: ignore\n          retention-days: 7\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("an upload step lacks continue-on-error and retention-days:\n%s", out)
		}
	}
}
