package discovery

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func rootModule(t *testing.T) string {
	t.Helper()
	ws := t.TempDir()
	files := map[string]string{
		"go.mod":                  "module example.com/mvd\n\ngo 1.22\n",
		"apps/mvd/main.go":        "package main\n\nimport \"example.com/mvd/libs/core/greeting\"\n\nfunc main() { greeting.Hi() }\n",
		"libs/core/greeting/g.go": "package greeting\n\nfunc Hi() {}\n",
		"libs/core/core.go":       "package core\n",
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

func TestDeclaredFoldersOfARootModuleAreProjectsOfThatModule(t *testing.T) {
	ws := rootModule(t)
	res, err := DiscoverWithDeclared(ws, []string{"go.mod"}, []protocol.DeclaredProject{{Name: "mvd", Root: "apps/mvd"}, {Name: "core", Root: "libs/core", Kind: protocol.ProjectLib}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Projects) != 2 {
		t.Fatalf("the root go.mod is no project, the two declared are: %+v", res.Projects)
	}
	got := map[string]protocol.Project{}
	for _, p := range res.Projects {
		got[p.Name] = p
	}
	app, lib := got["mvd"], got["core"]
	if app.Kind != protocol.ProjectApp || lib.Kind != protocol.ProjectLib {
		t.Errorf("an app for the folder with a main package, the declared lib: %v, %v", app.Kind, lib.Kind)
	}
	if !slices.Equal(app.Exports, []string{"example.com/mvd/apps/mvd"}) || !slices.Equal(lib.Exports, []string{"example.com/mvd/libs/core"}) {
		t.Errorf("exports are the package paths: %v, %v", app.Exports, lib.Exports)
	}
	if !slices.Contains(app.Imports, "example.com/mvd/libs/core/greeting") {
		t.Errorf("the app imports the lib's sub-package, which the graph turns into an edge: %v", app.Imports)
	}
	for name, task := range app.Tasks {
		if len(task.Inputs) > 0 && (!slices.Contains(task.Inputs, "{workspaceRoot}/go.mod") || !slices.Contains(task.Inputs, "{workspaceRoot}/go.sum")) {
			t.Errorf("task %s must be cached on the module's go.mod and go.sum: %v", name, task.Inputs)
		}
	}
	if app.Tasks["build"].Command == nil || app.Tasks["test"].Command == nil || app.Language != "go" {
		t.Errorf("tasks and language: %+v", app)
	}
}

func TestAModuleInAFolderAboveTheDeclaredOneIsFound(t *testing.T) {
	ws := t.TempDir()
	check(t, os.MkdirAll(filepath.Join(ws, "services", "api", "cmd", "serve"), 0o755))
	check(t, os.WriteFile(filepath.Join(ws, "services", "api", "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644))
	check(t, os.WriteFile(filepath.Join(ws, "services", "api", "cmd", "serve", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644))
	res, err := DiscoverWithDeclared(ws, nil, []protocol.DeclaredProject{{Name: "serve", Root: "services/api/cmd/serve"}}, nil)
	if err != nil || len(res.Projects) != 1 {
		t.Fatalf("%v %v", res.Projects, err)
	}
	p := res.Projects[0]
	if p.Exports[0] != "example.com/api/cmd/serve" || !slices.Contains(p.Tasks["test"].Inputs, "{workspaceRoot}/services/api/go.mod") {
		t.Errorf("%+v", p)
	}
}

func TestADeclarationThatCannotBeDescribedSaysWhy(t *testing.T) {
	ws := rootModule(t)
	check(t, os.WriteFile(filepath.Join(ws, "libs", "core", "go.mod"), []byte("module example.com/core\n\ngo 1.22\n"), 0o644))
	bare := t.TempDir()
	check(t, os.MkdirAll(filepath.Join(bare, "apps", "x"), 0o755))
	check(t, os.WriteFile(filepath.Join(bare, "apps", "x", "main.go"), []byte("package main\n"), 0o644))
	for name, c := range map[string]struct {
		root, workspace, want string
		kind                  protocol.ProjectKind
	}{
		"its own go.mod":  {"libs/core", ws, "go.mod of its own", ""},
		"no folder":       {"apps/missing", ws, "does not exist", ""},
		"no module above": {"apps/x", bare, "no go.mod in a folder above", ""},
		"a package kind":  {"apps/mvd", ws, "cannot be released on its own", protocol.ProjectPackage},
	} {
		_, err := DiscoverWithDeclared(c.workspace, nil, []protocol.DeclaredProject{{Name: "p", Root: c.root, Kind: c.kind}}, nil)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: want %q, got %v", name, c.want, err)
		}
	}
}

func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
