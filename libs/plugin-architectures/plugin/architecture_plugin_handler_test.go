package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

func TestEveryArchitectureIsServed(t *testing.T) {
	if len(Names()) < 2 {
		t.Fatalf("%v", Names())
	}
	for _, name := range Names() {
		p := New(name, "1.0.0")
		if p == nil {
			t.Fatalf("%s is not served", name)
		}
		res, err := p.Describe(protocol.InitializeParams{})
		if err != nil || res.Plugin.Name != name || res.Kind != protocol.KindConvention {
			t.Fatalf("%s: %+v %v", name, res, err)
		}
	}
	if New("onion", "1") != nil {
		t.Fatal("an unknown architecture has a plugin")
	}
}

func TestContribute(t *testing.T) {
	a := New("hexagonal", "1.0.0")
	if c := serve.CapabilitiesOf(a); !c.Contribute || c.Discover || !c.Doctor {
		t.Fatalf("capabilities %+v", c)
	}
	res, err := a.Contribute(protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "tool", Language: "go", Tasks: map[string]protocol.Task{"lint": {}}},
		{Name: "web", Language: "node"},
		{Name: "infra", Language: "bicep"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	tool := res.Projects["tool"]
	got := strings.Join(tool.Tasks[TaskName].Command, " ")
	if got != "{smith} slices check --architecture hexagonal --lang go ." || tool.Tasks[TaskName].Inputs[0] != "**/*.go" || tool.DependsOn["lint"][0] != TaskName {
		t.Fatalf("tool %+v", tool)
	}
	if web := res.Projects["web"]; web.Tasks[TaskName].Command[6] != "node" || len(web.DependsOn) != 0 {
		t.Fatalf("web %+v", web)
	}
	if _, ok := res.Projects["infra"]; ok {
		t.Fatal("a language the check cannot read got a task")
	}
}

func TestDoctorWarnsAboutAProjectWithNoLayerFolders(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"app/domain/order.go", "tool/main.go"} {
		p := filepath.Join(root, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	a := New("hexagonal", "1.0.0")
	if _, err := a.Describe(protocol.InitializeParams{WorkspaceRoot: root}); err != nil {
		t.Fatal(err)
	}
	res, err := a.Doctor(protocol.DoctorParams{Projects: []protocol.Project{
		{Name: "app", Root: "app", Language: "go"}, {Name: "tool", Root: "tool", Language: "go"},
	}})
	if err != nil || len(res.Findings) != 1 {
		t.Fatalf("%+v %v", res, err)
	}
	f := res.Findings[0]
	if f.OK || !f.Warning || !strings.Contains(f.Detail, "tool") || strings.Contains(f.Detail, "app,") || f.Remedy == "" {
		t.Fatalf("%+v", f)
	}
	res, _ = a.Doctor(protocol.DoctorParams{Projects: []protocol.Project{{Name: "app", Root: "app", Language: "go"}}})
	if !res.Findings[0].OK {
		t.Fatalf("%+v", res.Findings[0])
	}
}

func TestEveryArchitectureDeclaresTheLanguagesItAppliesTo(t *testing.T) {
	for _, name := range Names() {
		res, err := New(name, "1.0.0").Describe(protocol.InitializeParams{})
		if err != nil || len(res.Languages) == 0 {
			t.Fatalf("%s declares no languages: %+v %v", name, res, err)
		}
		if got := Languages(name); strings.Join(got, ",") != strings.Join(res.Languages, ",") {
			t.Fatalf("%s: Languages says %v, the handshake %v", name, got, res.Languages)
		}
	}
	if Languages("onion") != nil {
		t.Fatal("an unknown architecture has languages")
	}
}

func TestARulesetThatListsLanguagesAppliesToOnlyThose(t *testing.T) {
	a := New("hexagonal", "1.0.0")
	a.rules.Languages = []string{"go"}
	res, err := a.Describe(protocol.InitializeParams{})
	if err != nil || strings.Join(res.Languages, ",") != "go" {
		t.Fatalf("%v %v", res.Languages, err)
	}
	got, err := a.Contribute(protocol.ContributeParams{Projects: []protocol.Project{
		{Name: "tool", Language: "go"},
		{Name: "web", Language: "node"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Projects["tool"]; !ok {
		t.Fatal("a listed language got no task")
	}
	if _, ok := got.Projects["web"]; ok {
		t.Fatal("a language the ruleset leaves out got a task")
	}
}
