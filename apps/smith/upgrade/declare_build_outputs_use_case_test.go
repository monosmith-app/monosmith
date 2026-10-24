package upgrade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func outputsWorkspace(t *testing.T) (*session.Session, string) {
	t.Helper()
	dir := t.TempDir()
	for name, text := range map[string]string{
		"packages/auth/package.json": `{"scripts": {"build": "esbuild src/index.ts --bundle --outfile=dist/index.esm.js && tsc --declarationDir dist/src"}}`,
		"packages/ok/package.json":   `{"scripts": {"build": "tsc --outDir ../../dist/packages/ok"}}`,
		"smith.toml":                 "[workspace]\n",
	} {
		file := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	build := func(name string) map[string]protocol.Task {
		return map[string]protocol.Task{"build": {Command: []string{"npm", "run", "build", "--silent"}, Outputs: []string{"../../dist/packages/" + name}}}
	}
	graph := &projectgraph.Graph{Projects: []*projectgraph.Project{
		{Project: protocol.Project{Name: "auth", Root: "packages/auth", Tasks: build("auth")}, Plugin: "node"},
		{Project: protocol.Project{Name: "ok", Root: "packages/ok", Tasks: build("ok")}, Plugin: "node"},
	}}
	return &session.Session{Root: dir, Graph: graph}, dir
}

func TestUpgradeDeclaresTheOutputsABuildWrites(t *testing.T) {
	s, dir := outputsWorkspace(t)
	var r Result
	if err := declareBuildOutputs(s, false, &r); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "smith.toml"))
	if !strings.Contains(string(raw), "[projects.auth.tasks.build]\noutputs = [\"dist\"]") || strings.Contains(string(raw), "projects.ok") {
		t.Fatalf("%s", raw)
	}
	if len(r.Migrations) != 1 || !strings.Contains(r.Migrations[0], "auth (dist)") || len(r.Files) != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestUpgradeCheckLeavesTheOutputsAlone(t *testing.T) {
	s, dir := outputsWorkspace(t)
	var r Result
	if err := declareBuildOutputs(s, true, &r); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "smith.toml"))
	if string(raw) != "[workspace]\n" || len(r.Migrations) != 1 {
		t.Fatalf("%s %+v", raw, r)
	}
}
