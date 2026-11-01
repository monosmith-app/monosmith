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

func TestConvertSaysWhatToDoForAGoPackageOfARootModule(t *testing.T) {
	p := goPackage("apps/mvd", "apps/mvd/main.go")
	plan := MakePlan(&nxworkspace.Workspace{Projects: []*nxworkspace.Project{p}}, Input{DirName: "acme", Files: []string{"go.mod", "apps/mvd/main.go"}, Read: func(string) string { return "" }})
	var got string
	for _, item := range plan.Findings {
		if strings.Contains(item.Message, "Go package") {
			got = item.Message
		}
		if strings.Contains(item.Message, "smith v1 has no plugin for it") {
			t.Errorf("the generic message is wrong for Go: %q", item.Message)
		}
	}
	for _, want := range []string{"a Go package of the module at the repository root", "go mod init <module>/apps/mvd in apps/mvd", "go work use ./apps/mvd"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
}
