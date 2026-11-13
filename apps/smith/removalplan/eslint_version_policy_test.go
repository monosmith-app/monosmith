package removalplan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestESLintMeetsConfig(t *testing.T) {
	for version, want := range map[string]bool{
		"9.39.3": false, "10.0.0": false, "10.8.0": false, "10.8.1": true, "10.12.0": true, "11.0.0": false,
	} {
		if got := ESLintMeetsConfig(version); got != want {
			t.Errorf("ESLintMeetsConfig(%q) = %v, want %v", version, got, want)
		}
	}
}

// The range smith raises eslint to is the config package's own peer
// dependency: when the package moves to a newer ESLint, this fails until the
// constant follows.
func TestLintESLintRangeIsTheConfigPackagesPeer(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "packages", "eslint-config", "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Name  string            `json:"name"`
		Peers map[string]string `json:"peerDependencies"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Name != LintConfigPackage || manifest.Peers["eslint"] != LintESLintRange || "^"+LintESLintFloor != LintESLintRange {
		t.Fatalf("%s peers eslint %q; LintESLintRange = %q, floor %q", manifest.Name, manifest.Peers["eslint"], LintESLintRange, LintESLintFloor)
	}
}
