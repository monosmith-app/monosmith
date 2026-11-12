package doctorchecks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func namingSession(t *testing.T, projects ...*projectgraph.Project) *session.Session {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, text string) {
		file := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("libs/core/package.json", "{\n  \"name\": \"@other/core\",\n  \"version\": \"1.0.0\"\n}\n")
	write("libs/rcore/Cargo.toml", "[package]\nname = \"rcore\"\nversion = \"0.1.0\"\n")
	graph, err := projectgraph.New(projects)
	if err != nil {
		t.Fatal(err)
	}
	return &session.Session{Root: dir, Graph: graph, Config: workspace.Config{Workspace: workspace.WorkspaceSection{Name: "shop", Namespace: "acme"}}}
}

func nodeProject(name, pkg string) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Root: "libs/" + name, Language: "node", PackageName: pkg}}
}

func TestAPackageNameOutsideTheNamespaceIsAWarningWithAFix(t *testing.T) {
	s := namingSession(t, nodeProject("core", "@other/core"))
	got := PackageNameFindings(s)
	if len(got) != 1 || got[0].OK || !got[0].Warning || got[0].Project != "core" {
		t.Fatalf("%+v", got)
	}
	if !strings.Contains(got[0].Problem, "@other/core") || !strings.Contains(got[0].Remedy, "@acme/core") {
		t.Errorf("text: %q / %q", got[0].Problem, got[0].Remedy)
	}
	fix := got[0].Fix
	if fix == nil || fix.Safe || len(fix.Files) != 1 || fix.Files[0].Path != "libs/core/package.json" || !strings.Contains(fix.Files[0].Content, `"name": "@acme/core"`) {
		t.Fatalf("a rename is a fix the person confirms: %+v", fix)
	}
	if !strings.Contains(fix.Files[0].Content, `"version": "1.0.0"`) {
		t.Errorf("the rest of the manifest is kept:\n%s", fix.Files[0].Content)
	}
}

func TestACargoRenameEditsOnlyTheName(t *testing.T) {
	s := namingSession(t, &projectgraph.Project{Project: protocol.Project{Name: "rcore", Root: "libs/rcore", Language: "rust", PackageName: "rcore"}})
	got := PackageNameFindings(s)
	if len(got) != 1 || got[0].Fix == nil || got[0].Fix.Files[0].Content != "[package]\nname = \"acme-rcore\"\nversion = \"0.1.0\"\n" {
		t.Fatalf("%+v", got)
	}
}

func TestAPackageOthersDependOnGetsNoFix(t *testing.T) {
	core := nodeProject("core", "@other/core")
	app := nodeProject("app", "@acme/app")
	app.Deps = []string{"core"}
	got := PackageNameFindings(namingSession(t, core, app))
	if len(got) != 1 || got[0].Fix != nil || !strings.Contains(got[0].Remedy, "app") {
		t.Fatalf("a dependent also names the package, so smith leaves the rename to the person: %+v", got)
	}
}

func TestPackagesThatFollowTheNamespacePass(t *testing.T) {
	s := namingSession(t, nodeProject("core", "@acme/core"), &projectgraph.Project{Project: protocol.Project{Name: "gocore", Root: "libs/gocore", Language: "go", PackageName: "anything"}})
	got := PackageNameFindings(s)
	if len(got) != 1 || !got[0].OK {
		t.Fatalf("%+v", got)
	}
}

func TestNoNamespaceSetMeansNoCheck(t *testing.T) {
	s := namingSession(t, nodeProject("core", "@other/core"))
	s.Config.Workspace.Namespace = ""
	if got := PackageNameFindings(s); len(got) != 0 {
		t.Fatalf("an adopted workspace keeps its names without a warning: %+v", got)
	}
}
