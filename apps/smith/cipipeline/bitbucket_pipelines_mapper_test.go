package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"go.yaml.in/yaml/v3"
)

func bitbucketConfig() workspace.Config {
	return workspace.Config{CI: workspace.CISection{Providers: []string{"bitbucket"}}}
}

func TestBitbucketMatchesTheGoldenFile(t *testing.T) {
	got := RenderBitbucket(multiLanguageModel())
	golden := filepath.Join("testdata", "bitbucket_multi_language.golden.yml")
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
		t.Fatalf("the Bitbucket config changed; review it, then rerun with SMITH_UPDATE_GOLDEN=1:\n%s", got)
	}
}

func TestBitbucketIsValidYAMLWithTheExpectedShape(t *testing.T) {
	type step struct {
		Step struct {
			Name   string
			Caches []string
			Script []string
		}
	}
	var doc struct {
		Image       string
		Definitions struct {
			Caches map[string]any
			Steps  []step
		}
		Pipelines struct {
			PullRequests map[string][]step `yaml:"pull-requests"`
			Branches     map[string][]step
		}
	}
	if err := yaml.Unmarshal([]byte(RenderBitbucket(multiLanguageModel())), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Image != BitbucketImage || len(doc.Definitions.Steps) != 2 || doc.Definitions.Caches["npm"] == nil {
		t.Fatalf("definitions: %+v", doc)
	}
	for name, steps := range map[string][]step{"pull requests": doc.Pipelines.PullRequests["**"], "main": doc.Pipelines.Branches["main"]} {
		if len(steps) != 2 || steps[0].Step.Name != "Verify" || len(steps[0].Step.Script) == 0 || steps[1].Step.Name != "Native (Linux)" {
			t.Errorf("pipeline %s: %+v", name, steps)
		}
	}
	if got := doc.Definitions.Steps[0].Step.Caches; len(got) != 2 {
		t.Errorf("caches: %v", got)
	}
}

func TestBitbucketCarriesTheSamePhasesAsGitHub(t *testing.T) {
	m := multiLanguageModel()
	bb, gh := RenderBitbucket(m), RenderGitHub(m)
	if got, want := strings.Join(uniquePhases(bb), ","), strings.Join(uniquePhases(gh), ","); got != want {
		t.Fatalf("providers drifted:\n%s\n%s", got, want)
	}
	for _, want := range []string{
		"image: debian:bookworm", "npm install --global --registry https://registry.npmjs.org/ monosmith@1.4.0",
		`if [ -n "$BITBUCKET_PR_ID" ]; then smith release --dry-run --comment; fi`,
		`[ "$BITBUCKET_BRANCH" = "main" ]`, "x-token-auth:${BITBUCKET_TOKEN}", "- dist/drop/**",
		"#   NODE_AUTH_TOKEN: publishes npm packages", "# smith: no plugin says how to install rust 1.90",
		"      key:\n        files:\n          - package-lock.json\n      path: ~/.npm",
		"# smith:slot before-setup", "# smith:slot after-verify",
	} {
		if !strings.Contains(bb, want) {
			t.Errorf("Bitbucket config lacks %q", want)
		}
	}
	if strings.Contains(bb, "\t") {
		t.Fatal("YAML forbids tabs")
	}
}

func TestBitbucketBuildsNativeProjectsOnLinuxAndSaysWhatItSkipped(t *testing.T) {
	m := multiLanguageModel()
	out := RenderBitbucket(m)
	if !strings.Contains(out, "&native") || !strings.Contains(out, "Not built on Bitbucket: windows-latest, macos-latest") {
		t.Errorf("no Linux native step or skipped runners unnamed:\n%s", out)
	}
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"bitbucket"}, Branches: []string{"main", "dev"}}, Release: workspace.ReleaseSection{Branch: "main"}}
	limits := strings.Join(Limits(cfg, m), "\n")
	for _, want := range []string{"Bitbucket: native projects build on Linux only, not on windows-latest, macos-latest", "Bitbucket: [ci] branches does not apply"} {
		if !strings.Contains(limits, want) {
			t.Errorf("limits lack %q:\n%s", want, limits)
		}
	}
	m.NativeRunners = []string{"windows-latest"}
	if out := RenderBitbucket(m); strings.Contains(out, "&native") || strings.Contains(out, "*native") {
		t.Errorf("a native step with no Linux runner:\n%s", out)
	}
}

func TestBitbucketWritesNoNativePackOrCacheWithoutThem(t *testing.T) {
	out := RenderBitbucket(NewModel(Contributions{}, "1.0.0", "main"))
	for _, unwanted := range []string{"native", "smith ci pack", "artifacts", "caches"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("unexpected %q in\n%s", unwanted, out)
		}
	}
	if err := yaml.Unmarshal([]byte(out), &map[string]any{}); err != nil {
		t.Fatal(err)
	}
}

func TestBitbucketSlotsSurviveRegeneration(t *testing.T) {
	root := t.TempDir()
	cfg := bitbucketConfig()
	if _, _, err := Write(root, cfg, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, BitbucketPath)
	body, _ := os.ReadFile(path)
	own := strings.Replace(string(body), "          # smith:slot after-verify\n", "          # smith:slot after-verify\n          - echo our own step\n", 1)
	if own == string(body) {
		t.Fatalf("no after-verify slot:\n%s", body)
	}
	_ = os.WriteFile(path, []byte(own), 0o644)
	if _, lost, err := Write(root, cfg, multiLanguageModel()); err != nil || len(lost) > 0 {
		t.Fatal(err, lost)
	}
	again, _ := os.ReadFile(path)
	if !strings.Contains(string(again), "          - echo our own step\n") {
		t.Fatalf("the slot's step was lost:\n%s", again)
	}
	if drift := Drift(root, cfg, multiLanguageModel()); len(drift) > 0 {
		t.Fatalf("a kept slot counts as drift: %v", drift)
	}
}

func TestBitbucketDriftSeesAnEditAndAMissingFile(t *testing.T) {
	root := t.TempDir()
	cfg := bitbucketConfig()
	m := multiLanguageModel()
	if drift := Drift(root, cfg, m); len(drift) != 1 || drift[0] != BitbucketPath+" is missing" {
		t.Fatalf("a missing config is drift: %v", drift)
	}
	if _, _, err := Write(root, cfg, m); err != nil {
		t.Fatal(err)
	}
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("a fresh config drifted: %v", drift)
	}
	path := filepath.Join(root, BitbucketPath)
	body, _ := os.ReadFile(path)
	_ = os.WriteFile(path, []byte(strings.Replace(string(body), "smith ci verify", "smith ci verify --fast", 1)), 0o644)
	if drift := Drift(root, cfg, m); len(drift) != 1 || drift[0] != BitbucketPath+" is out of date" {
		t.Fatalf("an edited config is not drift: %v", drift)
	}
}

func TestDroppingBitbucketRemovesOnlyAGeneratedFile(t *testing.T) {
	root := t.TempDir()
	if _, _, err := Write(root, bitbucketConfig(), multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, BitbucketPath)
	other := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	if _, _, err := Write(root, other, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the generated Bitbucket config stays after the provider is dropped")
	}
	_ = os.WriteFile(path, []byte("image: node\n# ours\n"), 0o644)
	if _, _, err := Write(root, other, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("a config the team wrote was deleted")
	}
}

func TestBitbucketManualStepsNameTheVariablesAndTheToken(t *testing.T) {
	steps := strings.Join(ManualSteps(bitbucketConfig(), multiLanguageModel()), "\n")
	if !strings.Contains(steps, "Bitbucket: add the secured repository variables NODE_AUTH_TOKEN, PYPI_TOKEN") || !strings.Contains(steps, "BITBUCKET_TOKEN") {
		t.Errorf("steps:\n%s", steps)
	}
}
