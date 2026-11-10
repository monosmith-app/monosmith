package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"go.yaml.in/yaml/v3"
)

func manualModel() Model {
	m := multiLanguageModel()
	m.ManualRelease = true
	m.AssetPlatforms = []string{"linux-x64"}
	return m
}

func TestGitHubManualReleaseMatchesTheGoldenFile(t *testing.T) {
	got := RenderGitHub(manualModel())
	golden := filepath.Join("testdata", "github_manual_release.golden.yml")
	if os.Getenv("SMITH_UPDATE_GOLDEN") != "" {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.ReplaceAll(string(want), "\r\n", "\n") {
		t.Fatalf("the GitHub workflow changed; review it, then rerun with SMITH_UPDATE_GOLDEN=1:\n%s", got)
	}
}

func TestManualReleaseIsOffByDefault(t *testing.T) {
	for name, out := range map[string]string{
		"github": RenderGitHub(multiLanguageModel()), "azure": RenderAzure(multiLanguageModel()),
		"gitlab": RenderGitLab(multiLanguageModel()), "bitbucket": RenderBitbucket(multiLanguageModel()),
	} {
		if strings.Contains(out, "release_bump") || strings.Contains(out, "RELEASE_BUMP") || strings.Contains(out, "releaseBump") {
			t.Errorf("%s: a workspace that did not opt in has a way to release by hand", name)
		}
	}
}

func TestManualReleaseWorkflowsAreValidYAML(t *testing.T) {
	m := manualModel()
	for name, out := range map[string]string{
		"github": RenderGitHub(m), "azure": RenderAzure(m), "gitlab": RenderGitLab(m), "bitbucket": RenderBitbucket(m),
	} {
		var v map[string]any
		if err := yaml.Unmarshal([]byte(out), &v); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestGitHubManualReleaseInputsAndJobs(t *testing.T) {
	out := RenderGitHub(manualModel())
	for _, want := range []string{
		"release_bump:",
		"RELEASE_BUMP: ${{ inputs.release_bump }}",
		`smith ci release --bump "$RELEASE_BUMP" --force`,
		"if: (github.event_name == 'push' || (github.event_name == 'workflow_dispatch' && inputs.release_bump != '')) && github.ref_name == 'main'",
		"needs.ci.outputs.released == 'true' && (github.event_name == 'push' || (github.event_name == 'workflow_dispatch' && inputs.release_bump != '')) && github.ref_name == 'main'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the workflow lacks %q", want)
		}
	}
	if strings.Contains(out, "run: ${{ inputs.release_bump") {
		t.Error("the input is part of a script")
	}
}

func TestOtherProvidersTakeTheManualValue(t *testing.T) {
	m := manualModel()
	for name, want := range map[string][]string{
		"azure":     {"name: releaseBump", "RELEASE_BUMP: ${{ parameters.releaseBump }}", "eq(variables['Build.Reason'], 'Manual')"},
		"gitlab":    {"RELEASE_BUMP:", `$CI_PIPELINE_SOURCE == "web"`, `smith ci release --bump "$RELEASE_BUMP" --force`},
		"bitbucket": {"custom:", "name: RELEASE_BUMP", `--bump "$RELEASE_BUMP" --force`},
	} {
		out := map[string]string{"azure": RenderAzure(m), "gitlab": RenderGitLab(m), "bitbucket": RenderBitbucket(m)}[name]
		for _, w := range want {
			if !strings.Contains(out, w) {
				t.Errorf("%s lacks %q", name, w)
			}
		}
	}
}

func TestCircleCISaysItCannotTakeAManualRelease(t *testing.T) {
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "circleci"}, ManualRelease: true}}
	got := strings.Join(Limits(cfg, Model{}), "\n")
	if !strings.Contains(got, "CircleCI: [ci] manual_release does not apply") {
		t.Fatal(got)
	}
	cfg.CI.ManualRelease = false
	if strings.Contains(strings.Join(Limits(cfg, Model{}), "\n"), "manual_release") {
		t.Fatal("a note without the setting")
	}
}

func TestActorAndSentence(t *testing.T) {
	env := map[string]string{"GITHUB_ACTOR": "mvd"}
	get := func(k string) string { return env[k] }
	if Actor(get) != "mvd" || Actor(func(string) string { return "" }) != "unknown" {
		t.Fatal(Actor(get))
	}
	s := ManualRelease{Bumps: []string{"mvd=patch"}, Force: true}.Sentence("mvd")
	if !strings.Contains(s, "started by mvd") || !strings.Contains(s, "--bump mvd=patch") {
		t.Fatal(s)
	}
}
