package discovery

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declaredWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	write(t, ws, "pyproject.toml", "[project]\nname = \"acme\"\nversion = \"1.0.0\"\n")
	write(t, ws, "src/core/__init__.py", "")
	write(t, ws, "apps/web/__main__.py", "import os\nfrom core import x\n")
	return ws
}

func TestDeclaredFoldersOfARootDistributionAreProjectsOfIt(t *testing.T) {
	ws := declaredWorkspace(t)
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "web", Root: "apps/web"}, {Name: "core", Root: "src/core"}})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	web, core := res.Projects[0], res.Projects[1]
	if web.Kind != protocol.ProjectApp || web.Language != "python" || !slices.Equal(web.Exports, []string{"web"}) || !slices.Equal(web.Imports, []string{"core", "os"}) {
		t.Errorf("web %+v", web)
	}
	if core.Kind != protocol.ProjectLib || !slices.Equal(core.Exports, []string{"core"}) {
		t.Errorf("core %+v", core)
	}
	for _, name := range []string{"lint", "typecheck", "test"} {
		task, ok := web.Tasks[name]
		if !ok || !slices.Contains(task.Inputs, "{workspaceRoot}/pyproject.toml") {
			t.Errorf("%s: %+v", name, task)
		}
	}
	if _, ok := web.Tasks["start"]; !ok {
		t.Error("an app can start")
	}
}

func TestDeclaredFolderBelowAnotherManifestNamesTheManifestAboveIt(t *testing.T) {
	ws := declaredWorkspace(t)
	write(t, ws, "tools/pyproject.toml", "[project]\nname = \"tools\"\n")
	write(t, ws, "tools/cli/main.py", "")
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "cli", Root: "tools/cli"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(res.Projects[0].Tasks["lint"].Inputs, "{workspaceRoot}/tools/pyproject.toml") {
		t.Errorf("inputs %v", res.Projects[0].Tasks["lint"].Inputs)
	}
}

func TestDeclaredFolderIsRefusedWithTheFix(t *testing.T) {
	ws := declaredWorkspace(t)
	write(t, ws, "own/pyproject.toml", "[project]\nname = \"own\"\n")
	write(t, ws, "loose/a.py", "")
	toolOnly := t.TempDir()
	write(t, toolOnly, "pyproject.toml", "[tool.ruff]\nline-length = 100\n")
	write(t, toolOnly, "apps/a/a.py", "")
	cases := []struct {
		ws   string
		d    protocol.DeclaredProject
		want string
	}{
		{ws, protocol.DeclaredProject{Name: "x", Root: "nope"}, "the folder does not exist"},
		{ws, protocol.DeclaredProject{Name: "own", Root: "own"}, "pyproject.toml of its own"},
		{ws, protocol.DeclaredProject{Name: "core", Root: "src/core", Kind: protocol.ProjectPackage}, "one version"},
		{toolOnly, protocol.DeclaredProject{Name: "a", Root: "apps/a"}, "[project] table"},
	}
	for _, c := range cases {
		_, err := DiscoverWithDeclared(c.ws, nil, []protocol.DeclaredProject{c.d})
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "[projects."+c.d.Name+"]") {
			t.Errorf("%s: %v, want %q", c.d.Root, err, c.want)
		}
	}
}
