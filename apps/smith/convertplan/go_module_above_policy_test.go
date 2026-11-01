package convertplan

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

func goPackage(root string, files ...string) *nxworkspace.Project {
	return &nxworkspace.Project{Name: "mvd", Root: root, Files: files}
}

func TestAGoPackageOfARootModuleIsToldToGetItsOwnGoMod(t *testing.T) {
	p := goPackage("apps/mvd", "apps/mvd/main.go", "apps/mvd/project.json")
	dir, ok := GoModuleAbove(p, []string{"go.mod", "apps/mvd/main.go"})
	if !ok || dir != "the repository root" {
		t.Fatalf("got %q, %v", dir, ok)
	}
	if dir, ok := GoModuleAbove(p, []string{"apps/go.mod"}); !ok || dir != "apps" {
		t.Errorf("a module in a parent folder: %q, %v", dir, ok)
	}
	if _, ok := GoModuleAbove(p, []string{"libs/x/go.mod"}); ok {
		t.Error("a go.mod elsewhere is no module above")
	}
	if _, ok := GoModuleAbove(goPackage("apps/web", "apps/web/index.html"), []string{"go.mod"}); ok {
		t.Error("a project with no .go files is no Go package")
	}
}

func TestConvertDeclaresAGoPackageOfARootModuleAsAProject(t *testing.T) {
	p := goPackage("apps/mvd", "apps/mvd/main.go")
	p.Type = "app"
	plan := MakePlan(&nxworkspace.Workspace{Projects: []*nxworkspace.Project{p}}, Input{DirName: "acme", Files: []string{"go.mod", "apps/mvd/main.go"}, Read: func(string) string { return "" }})
	var got string
	for _, item := range plan.Findings {
		if strings.Contains(item.Message, "declared in smith.toml") {
			got = item.Message
		}
		if strings.Contains(item.Message, "smith v1 has no plugin for it") {
			t.Errorf("a Go package has a plugin: %q", item.Message)
		}
	}
	if !strings.Contains(got, "a Go project of the module at the repository root") {
		t.Errorf("the finding says it is declared: %q", got)
	}
	pp := plan.Projects[0]
	if !pp.Supported || !pp.Declared || pp.Language != "go" || pp.KindMarker {
		t.Errorf("a supported declared Go project, no kind marker in a manifest it lacks: %+v", pp)
	}
	if o := plan.Config.Projects[pp.Name]; o.Root != "apps/mvd" || o.Language != "go" || o.Kind != "app" {
		t.Errorf("the declaration goes to smith.toml: %+v", o)
	}
}
