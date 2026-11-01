package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func declaredWorkspace(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{
		"kit-1.0-1.rockspec":       "package = \"kit\"\nversion = \"1.0-1\"\n",
		"src/core/init.lua":        "local M = {}\nreturn M\n",
		"apps/web/main.lua":        "local c = require(\"core.init\")\nlocal s = require 'strings'\n",
		"apps/web/spec/a_spec.lua": "",
		"own/own-1.0-1.rockspec":   "package = \"own\"\n",
	}
	for name, body := range files {
		full := filepath.Join(ws, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func TestDeclaredFoldersAreModulesOfTheRockAboveThem(t *testing.T) {
	ws := declaredWorkspace(t)
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "web", Root: "apps/web"}, {Name: "core", Root: "src/core"}})
	if err != nil || len(res.Projects) != 2 {
		t.Fatal(err, res.Projects)
	}
	web, core := res.Projects[0], res.Projects[1]
	if web.Kind != protocol.ProjectApp || !slices.Equal(web.Exports, []string{"web"}) || !slices.Equal(web.Imports, []string{"core", "strings"}) {
		t.Errorf("web %+v", web)
	}
	if core.Kind != protocol.ProjectLib {
		t.Errorf("core %+v", core)
	}
	if _, ok := core.Tasks["test"]; ok {
		t.Error("no spec/ folder, no test task")
	}
	if _, ok := web.Tasks["test"]; !ok {
		t.Error("spec/ gives a test task")
	}
	if !slices.Contains(web.Tasks["lint"].Inputs, "{workspaceRoot}/kit-1.0-1.rockspec") {
		t.Errorf("inputs %v", web.Tasks["lint"].Inputs)
	}
	if _, ok := web.Tasks["build"]; ok {
		t.Error("a rock is built whole: no build task for a folder")
	}
}

func TestDeclaredFolderIsRefusedWithTheFix(t *testing.T) {
	ws := declaredWorkspace(t)
	bare := t.TempDir()
	if err := os.MkdirAll(filepath.Join(bare, "lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		ws   string
		d    protocol.DeclaredProject
		want string
	}{
		{ws, protocol.DeclaredProject{Name: "x", Root: "nope"}, "the folder does not exist"},
		{ws, protocol.DeclaredProject{Name: "own", Root: "own"}, "rockspec of its own"},
		{ws, protocol.DeclaredProject{Name: "core", Root: "src/core", Kind: protocol.ProjectPackage}, "one version"},
		{bare, protocol.DeclaredProject{Name: "lib", Root: "lib"}, "no rockspec is above"},
	}
	for _, c := range cases {
		_, err := DiscoverWithDeclared(c.ws, nil, []protocol.DeclaredProject{c.d})
		if err == nil || !strings.Contains(err.Error(), c.want) || !strings.Contains(err.Error(), "[projects."+c.d.Name+"]") {
			t.Errorf("%s: %v, want %q", c.d.Root, err, c.want)
		}
	}
}
