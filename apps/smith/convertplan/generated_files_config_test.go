package convertplan

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
)

// TestGeneratedFilesPassTheESLintConfig lints every file smith writes into
// a workspace with the eslint config it installs (this repository's
// packages/eslint-config, through its node_modules), so a converted
// workspace's lint does not fail on files nobody there wrote. It needs the
// repository's npm dependencies, which CI installs, and skips without them.
func TestGeneratedFilesPassTheESLintConfig(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	eslint := filepath.Join(repo, "node_modules", "eslint", "bin", "eslint.js")
	config := filepath.Join(repo, "packages", "eslint-config", "index.js")
	if _, err := os.Stat(eslint); err != nil {
		t.Skip("the repository's npm dependencies are not installed")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	dir := t.TempDir()
	files := map[string]string{
		DeclarationSpecifiersFile:   DeclarationSpecifiers,
		AssetTransformFile:          AssetTransform,
		"jest.preset.js":            JestPreset,
		JestResolverFile:            JestResolver,
		cipipeline.DevcontainerPath: cipipeline.RenderDevcontainer("demo", cipipeline.NewModel(cipipeline.Contributions{}, "1.0.0", "main")),
		"package.json":              `{"name": "demo", "private": true}` + "\n",
		"eslint.config.mjs":         "import monosmith from '" + filepath.ToSlash("file:///"+strings.TrimPrefix(filepath.ToSlash(config), "/")) + "'\n\nexport default [\n  ...monosmith(),\n]\n",
	}
	var lint []string
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if name != "package.json" && name != "eslint.config.mjs" {
			lint = append(lint, filepath.FromSlash(name))
		}
	}
	cmd := exec.Command("node", append([]string{eslint, "--no-warn-ignored"}, lint...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("eslint fails a file smith writes:\n%s", out)
	}
}
