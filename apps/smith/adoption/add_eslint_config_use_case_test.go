package adoption

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWithDevDependencyAddsOrKeeps(t *testing.T) {
	for name, raw := range map[string]string{
		"none":     "{\n  \"name\": \"x\"\n}\n",
		"empty":    "{\n  \"name\": \"x\",\n  \"devDependencies\": {}\n}\n",
		"existing": "{\n  \"devDependencies\": {\n    \"jest\": \"^30.0.0\"\n  }\n}\n",
	} {
		out, changed, err := withDevDependency([]byte(raw), eslintConfigPackage, "^1.0.0")
		if err != nil || !changed || !strings.Contains(string(out), `"@monosmith/eslint-config": "^1.0.0"`) {
			t.Errorf("%s: %v %v\n%s", name, changed, err, out)
		}
		if name == "existing" && !strings.Contains(string(out), `"jest": "^30.0.0"`) {
			t.Errorf("%s lost an entry:\n%s", name, out)
		}
		again, changed, _ := withDevDependency(out, eslintConfigPackage, "^2.0.0")
		if changed || string(again) != string(out) {
			t.Errorf("%s: a listed package is left alone", name)
		}
	}
}

func TestAddESLintConfigWritesTheFileAndKeepsYours(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{\n  \"name\": \"x\"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := AddESLintConfig(root, &out); err != nil {
		t.Fatal(err)
	}
	cfg, _ := os.ReadFile(filepath.Join(root, "eslint.config.mjs"))
	pkg, _ := os.ReadFile(filepath.Join(root, "package.json"))
	if !strings.Contains(string(cfg), "monosmith({ workspaceRoot: import.meta.dirname })") || !strings.Contains(string(pkg), eslintConfigPackage) {
		t.Fatalf("config %q package %q", cfg, pkg)
	}
	if !strings.Contains(out.String(), "wrote eslint.config.mjs") || !strings.Contains(out.String(), "devDependencies") {
		t.Errorf("the report must say what was done: %s", out.String())
	}

	mine := t.TempDir()
	if err := os.WriteFile(filepath.Join(mine, "eslint.config.js"), []byte("export default []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := AddESLintConfig(mine, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(mine, "eslint.config.mjs")); err == nil || !strings.Contains(out.String(), "kept your eslint.config.js") {
		t.Errorf("a config of the user's is kept: %s", out.String())
	}
}
