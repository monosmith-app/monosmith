package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"go.yaml.in/yaml/v3"
)

// multiLanguageModel is a workspace with Go, Node and Python projects: each
// plugin's install command, a package store and a secret each, packed apps
// and a native project.
func multiLanguageModel() Model {
	var c Contributions
	c.Add(protocol.CIResult{
		Toolchains: []protocol.Toolchain{{Name: "go", Version: "1.26", Install: "curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz | tar -C /usr/local -xz\nexport PATH=/usr/local/go/bin:$HOME/go/bin:$PATH"}},
		Stores:     []protocol.PackageStore{{Name: "go", Path: "~/go/pkg/mod", Env: "GOMODCACHE", Files: []string{"apps/api/go.sum", "libs/core/go.sum"}}},
	})
	c.Add(protocol.CIResult{
		Toolchains:  []protocol.Toolchain{{Name: "node", Version: "24", Install: "curl -fsSL https://deb.nodesource.com/setup_24.x | bash -\napt-get install -y -qq nodejs"}},
		Stores:      []protocol.PackageStore{{Name: "npm", Path: "~/.npm", Env: "NPM_CONFIG_CACHE", Files: []string{"package-lock.json"}}},
		Credentials: []protocol.Credential{{Env: "NODE_AUTH_TOKEN", Description: "publishes npm packages", Required: true}},
	})
	c.Add(protocol.CIResult{
		Toolchains:  []protocol.Toolchain{{Name: "python", Version: "3.12", Install: "curl -LsSf https://astral.sh/uv/install.sh | sh\nexport PATH=\"$HOME/.local/bin:$PATH\""}, {Name: "rust", Version: "1.90"}},
		Credentials: []protocol.Credential{{Env: "PYPI_TOKEN", Description: "publishes to PyPI", Required: true}},
	})
	c.Pack, c.Native = true, true
	c.NativeRunners = DefaultNativeRunners
	return NewModel(c, "1.4.0", "main")
}

func TestCircleCIMatchesTheGoldenFile(t *testing.T) {
	got := RenderCircleCI(multiLanguageModel())
	golden := filepath.Join("testdata", "circleci_multi_language.golden.yml")
	if os.Getenv("SMITH_UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.ReplaceAll(string(want), "\r\n", "\n") {
		t.Fatalf("the CircleCI config changed; review it, then rerun with SMITH_UPDATE_GOLDEN=1:\n%s", got)
	}
}

func TestCircleCIIsValidYAMLWithTheExpectedShape(t *testing.T) {
	var doc struct {
		Version   float64
		Executors map[string]any
		Commands  map[string]struct{ Steps []any }
		Jobs      map[string]struct{ Steps []any }
		Workflows map[string]struct{ Jobs []any }
	}
	if err := yaml.Unmarshal([]byte(RenderCircleCI(multiLanguageModel())), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != 2.1 || doc.Executors["smith"] == nil {
		t.Fatalf("header: %+v", doc)
	}
	for _, name := range []string{"install-tools", "save-stores"} {
		if len(doc.Commands[name].Steps) == 0 {
			t.Errorf("command %s has no steps", name)
		}
	}
	if len(doc.Jobs["ci"].Steps) == 0 || len(doc.Jobs["native"].Steps) == 0 {
		t.Fatalf("jobs: %+v", doc.Jobs)
	}
	if len(doc.Workflows["ci"].Jobs) != 2 {
		t.Fatalf("workflow: %+v", doc.Workflows)
	}
}

func TestCircleCICarriesTheSamePhasesAsGitHub(t *testing.T) {
	m := multiLanguageModel()
	circle, gh := RenderCircleCI(m), RenderGitHub(m)
	if got, want := strings.Join(uniquePhases(circle), ","), strings.Join(uniquePhases(gh), ","); got != want {
		t.Fatalf("providers drifted:\n%s\n%s", got, want)
	}
	for _, want := range []string{
		"image: debian:bookworm", "sh tools/smith/install.sh 1.4.0", "smith release --dry-run --comment",
		"equal: [ main, << pipeline.git.branch >> ]", "path: dist/drop\n                destination: drop",
		"#   NODE_AUTH_TOKEN: publishes npm packages", "# smith: no plugin says how to install rust 1.90",
		"smith-npm-v1-{{ arch }}-{{ checksum \"/tmp/smith-npm.lock\" }}",
	} {
		if !strings.Contains(circle, want) {
			t.Errorf("CircleCI config lacks %q", want)
		}
	}
	if strings.Contains(circle, "\t") {
		t.Fatal("YAML forbids tabs")
	}
}

// uniquePhases are the smith ci phases in order of first appearance.
func uniquePhases(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, p := range phasesOf(s) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

func TestCircleCIBuildsNativeProjectsOnLinuxAndSaysWhatItSkipped(t *testing.T) {
	m := multiLanguageModel()
	out := RenderCircleCI(m)
	if !strings.Contains(out, "  native:\n    executor: smith\n") || !strings.Contains(out, "destination: native-linux") {
		t.Errorf("no Linux native job:\n%s", out)
	}
	if !strings.Contains(out, "Not built on CircleCI: windows-latest, macos-latest") {
		t.Errorf("the skipped runners are not named:\n%s", out)
	}
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"circleci"}, Branches: []string{"main", "dev"}}, Release: workspace.ReleaseSection{Branch: "main"}}
	limits := strings.Join(Limits(cfg, m), "\n")
	for _, want := range []string{"native projects build on Linux only, not on windows-latest, macos-latest", "[ci] branches does not apply"} {
		if !strings.Contains(limits, want) {
			t.Errorf("limits lack %q:\n%s", want, limits)
		}
	}
	m.NativeRunners = []string{"windows-latest"}
	if out := RenderCircleCI(m); strings.Contains(out, "  native:\n") || strings.Contains(out, "requires: [ci]") {
		t.Errorf("a native job with no Linux runner:\n%s", out)
	}
}

func TestCircleCIWritesNoNativeOrPackStepsWithoutThem(t *testing.T) {
	out := RenderCircleCI(NewModel(Contributions{}, "1.0.0", "main"))
	for _, unwanted := range []string{"native", "smith ci pack", "store_artifacts", "save-stores", "restore_cache"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("unexpected %q in\n%s", unwanted, out)
		}
	}
}

func circleciConfig() workspace.Config {
	return workspace.Config{CI: workspace.CISection{Providers: []string{"circleci"}}}
}

func TestCircleCISlotsSurviveRegeneration(t *testing.T) {
	root := t.TempDir()
	cfg := circleciConfig()
	if _, _, err := Write(root, cfg, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(CircleCIPath))
	body, _ := os.ReadFile(path)
	own := strings.Replace(string(body), "      # smith:slot after-verify\n", "      # smith:slot after-verify\n      - run: echo our own step\n", 1)
	if own == string(body) {
		t.Fatalf("no after-verify slot:\n%s", body)
	}
	_ = os.WriteFile(path, []byte(own), 0o644)
	if _, lost, err := Write(root, cfg, multiLanguageModel()); err != nil || len(lost) > 0 {
		t.Fatal(err, lost)
	}
	again, _ := os.ReadFile(path)
	if !strings.Contains(string(again), "      - run: echo our own step\n") {
		t.Fatalf("the slot's step was lost:\n%s", again)
	}
	if drift := Drift(root, cfg, multiLanguageModel()); len(drift) > 0 {
		t.Fatalf("a kept slot counts as drift: %v", drift)
	}
}

func TestCircleCIDriftSeesAnEditAndAMissingFile(t *testing.T) {
	root := t.TempDir()
	cfg := circleciConfig()
	m := multiLanguageModel()
	if drift := Drift(root, cfg, m); len(drift) != 2 || !strings.Contains(strings.Join(drift, ";"), CircleCIPath+" is missing") {
		t.Fatalf("a missing config is drift: %v", drift)
	}
	if _, _, err := Write(root, cfg, m); err != nil {
		t.Fatal(err)
	}
	if drift := Drift(root, cfg, m); len(drift) != 0 {
		t.Fatalf("a fresh config drifted: %v", drift)
	}
	path := filepath.Join(root, filepath.FromSlash(CircleCIPath))
	body, _ := os.ReadFile(path)
	_ = os.WriteFile(path, []byte(strings.Replace(string(body), "smith ci verify", "smith ci verify --fast", 1)), 0o644)
	if drift := Drift(root, cfg, m); len(drift) != 1 || drift[0] != CircleCIPath+" is out of date" {
		t.Fatalf("an edited config is not drift: %v", drift)
	}
}

func TestDroppingCircleCIRemovesOnlyAGeneratedFile(t *testing.T) {
	root := t.TempDir()
	if _, _, err := Write(root, circleciConfig(), multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, filepath.FromSlash(CircleCIPath))
	other := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	if _, _, err := Write(root, other, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("the generated CircleCI config stays after the provider is dropped")
	}
	_ = os.WriteFile(path, []byte("version: 2.1\n# ours\n"), 0o644)
	if _, _, err := Write(root, other, multiLanguageModel()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("a config the team wrote was deleted")
	}
}

func TestCircleCIManualStepsNameTheVariablesAndTheWriteKey(t *testing.T) {
	steps := strings.Join(ManualSteps(circleciConfig(), multiLanguageModel()), "\n")
	if !strings.Contains(steps, "CircleCI: add the environment variables NODE_AUTH_TOKEN, PYPI_TOKEN") || !strings.Contains(steps, "user key with write access") {
		t.Errorf("steps:\n%s", steps)
	}
}
