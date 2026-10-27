package kindchange

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func edited(r protocol.ChangeKindResult) map[string]string {
	out := map[string]string{}
	for _, f := range r.Files {
		out[f.Path] = f.Content
	}
	return out
}

func projects() []protocol.Project {
	return []protocol.Project{
		{Name: "core", Root: "libs/core", Language: "python"},
		{Name: "util", Root: "libs/util", Language: "python"},
		{Name: "api", Root: "apps/api", Language: "python"},
	}
}

func workspace() map[string]string {
	return map[string]string{
		"pyproject.toml":       "[tool.uv.workspace]\nmembers = [\"libs/*\", \"apps/*\"]\n",
		"requirements-dev.txt": "-e libs/core\n-e libs/util\n",
		"uv.lock":              "[[package]]\nname = \"core\"\nsource = { editable = \"libs/core\" }\n",
		"libs/core/pyproject.toml": "[project]\nname = \"core\"\nversion = \"0.0.0\"\ndependencies = [\"util\"]\n\n" +
			"[tool.smith]\nkind = \"lib\"\n\n[tool.uv.sources]\nutil = { path = \"../util\", editable = true }\n",
		"libs/core/src/core/__init__.py": "",
		"libs/util/pyproject.toml":       "[project]\nname = \"util\"\nversion = \"0.0.0\"\n\n[tool.smith]\nkind = \"lib\"\n",
		"apps/api/pyproject.toml": "[project]\nname = \"api\"\nversion = \"0.0.0\"\ndependencies = [\"core\"]\n\n" +
			"[tool.smith]\nkind = \"app\"\n\n[tool.uv.sources]\ncore = { path = \"../../libs/core\" }\n",
		"apps/api/src/api/__main__.py": "",
	}
}

func TestLibBecomesPackage(t *testing.T) {
	dir := writeFiles(t, workspace())
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[0], To: protocol.ProjectPackage, NewRoot: "packages/core", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := edited(r)
	core := got["libs/core/pyproject.toml"]
	if !strings.Contains(core, `kind = "package"`) || !strings.Contains(core, `path = "../../libs/util"`) {
		t.Errorf("kind changes and the sibling stays reachable:\n%s", core)
	}
	if !strings.Contains(got["apps/api/pyproject.toml"], `path = "../../packages/core"`) {
		t.Errorf("dependents follow:\n%s", got["apps/api/pyproject.toml"])
	}
	root := got["pyproject.toml"]
	if !strings.Contains(root, `"packages/*"`) || !strings.Contains(root, `"libs/*"`) {
		t.Errorf("uv members cover the new directory:\n%s", root)
	}
	if got["requirements-dev.txt"] != "-e packages/core\n-e libs/util\n" {
		t.Errorf("requirements follow: %q", got["requirements-dev.txt"])
	}
	if !strings.Contains(got["uv.lock"], `editable = "packages/core"`) {
		t.Errorf("lockfile follows:\n%s", got["uv.lock"])
	}
	joined := strings.Join(r.Warnings, "\n")
	if !strings.Contains(joined, "license") || !strings.Contains(joined, "smith install") {
		t.Errorf("license and reinstall are named: %v", r.Warnings)
	}
	if _, touched := got["libs/util/pyproject.toml"]; touched {
		t.Error("a project that does not use core is not edited")
	}
}

func TestAnAppBecomingALibKeepsItsProgramAndSaysSo(t *testing.T) {
	dir := writeFiles(t, workspace())
	r, err := Change(dir, protocol.ChangeKindParams{
		Project: projects()[2], To: protocol.ProjectLib, NewRoot: "libs/api", Projects: projects(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "program") {
		t.Fatalf("the program is named: %v", r.Warnings)
	}
	if got := edited(r)["apps/api/pyproject.toml"]; !strings.Contains(got, `kind = "lib"`) || !strings.Contains(got, `path = "../core"`) {
		t.Fatalf("kind changes; the path to core is now a sibling's:\n%s", got)
	}
}
