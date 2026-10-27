package doctor

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func proj(name, language string) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Language: language}}
}

func TestNothingIsSaidWhenAConventionCoversEveryProject(t *testing.T) {
	cov := projectgraph.Coverage{Checked: []*projectgraph.Project{proj("api", "go")}}
	if _, ok := CoverageFinding("hexagonal", []string{"go"}, cov); ok {
		t.Fatal("a finding for full coverage")
	}
}

func TestAConventionThatSkipsSomeProjectsNamesThem(t *testing.T) {
	cov := projectgraph.Coverage{
		Checked: []*projectgraph.Project{proj("api", "go")},
		Skipped: []*projectgraph.Project{proj("web", "node"), proj("infra", "bicep")},
	}
	f, ok := CoverageFinding("hexagonal", []string{"go"}, cov)
	if !ok || !f.Warning || f.OK {
		t.Fatalf("%+v", f)
	}
	for _, want := range []string{"infra (bicep), web (node)", "applies to go"} {
		if !strings.Contains(f.Detail, want) {
			t.Fatalf("detail %q lacks %q", f.Detail, want)
		}
	}
	if !strings.Contains(f.Remedy, "exclude") {
		t.Fatalf("remedy %q", f.Remedy)
	}
}

func TestAConventionThatCoversNoProjectSaysSo(t *testing.T) {
	cov := projectgraph.Coverage{Skipped: []*projectgraph.Project{proj("web", "node")}}
	f, ok := CoverageFinding("hexagonal", []string{"go"}, cov)
	if !ok || !strings.Contains(f.Detail, "checks none of your projects") {
		t.Fatalf("%+v", f)
	}
}
