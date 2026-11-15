package cipipeline

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

var phaseCommand = regexp.MustCompile(`smith ci [a-z]+`)

func phasesOf(s string) []string { return phaseCommand.FindAllString(s, -1) }

func TestGitLabCarriesTheSamePhases(t *testing.T) {
	m := model()
	gl := RenderGitLab(m)
	for _, want := range []string{
		"image: debian:bookworm", "GIT_DEPTH: \"0\"", `$CI_COMMIT_BRANCH == "main"`,
		"    # zig 0.14\n    - 'curl -fsSL https://example.com/zig.sh | sh'\n    - 'zig version'\n",
		"    # smith: no plugin says how to install node 24",
		"smith release --dry-run --comment", "GITLAB_TOKEN", "#   PYPI_TOKEN",
	} {
		if !strings.Contains(gl, want) {
			t.Errorf("GitLab pipeline lacks %q", want)
		}
	}
	if strings.Join(phasesOf(gl), ",") != strings.Join(phasesOf(RenderGitHub(m)), ",") {
		t.Fatalf("providers drifted:\n%v\n%v", phasesOf(gl), phasesOf(RenderGitHub(m)))
	}
	if strings.Contains(gl, "\t") {
		t.Fatal("YAML forbids tabs")
	}
}

func TestGitLabSlotsSurviveRegeneration(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"gitlab"}}}
	if _, _, err := Write(root, cfg, model()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, GitLabPath)
	body, _ := os.ReadFile(path)
	own := strings.Replace(string(body), "    # smith:slot after-verify\n", "    # smith:slot after-verify\n    - echo our own step\n", 1)
	if own == string(body) {
		t.Fatalf("no after-verify slot:\n%s", body)
	}
	_ = os.WriteFile(path, []byte(own), 0o644)
	if _, lost, err := Write(root, cfg, model()); err != nil || len(lost) > 0 {
		t.Fatal(err, lost)
	}
	if again, _ := os.ReadFile(path); !strings.Contains(string(again), "    - echo our own step\n") {
		t.Fatalf("the slot's step was lost:\n%s", again)
	}
}
