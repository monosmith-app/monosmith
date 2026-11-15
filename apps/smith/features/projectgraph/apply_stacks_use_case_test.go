package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// language is a language plugin for the made-up language "x".
type language struct{}

func (language) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "xlang", Version: "2.0.0"}, Kind: protocol.KindLanguage, Languages: []string{"x"}}, nil
}

// toolchain owns lint: it replaces the language plugin's and adds format.
type toolchain struct{ name string }

func (t toolchain) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: t.name, Version: "1.0.0"}, Kind: protocol.KindConvention, Languages: []string{"x"},
		Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{"lint"},
		Stacks: []protocol.Stack{{Name: "x-" + t.name, Language: "xlang", Toolchain: []string{t.name}}},
	}, nil
}

func (t toolchain) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		res.Projects[p.Name] = protocol.Contribution{
			Replace: map[string]protocol.Task{"lint": {Command: []string{t.name, "check"}}},
			Tasks:   map[string]protocol.Task{"format": {Command: []string{t.name, "fmt"}}, "test": {Command: []string{"ignored"}}},
		}
	}
	return res, nil
}

func stackSet(t *testing.T) *pluginhost.PluginSet {
	return &pluginhost.PluginSet{Clients: []*host.Client{connect(t, language{}), connect(t, toolchain{"xfmt"}), connect(t, toolchain{"xlint"})}}
}

func xProjects() []*Project {
	task := func(cmd string) protocol.Task { return protocol.Task{Command: []string{cmd}} }
	return []*Project{
		{Project: protocol.Project{Name: "a", Language: "x", Tasks: map[string]protocol.Task{"lint": task("old"), "test": task("xtest")}}, Plugin: "xlang"},
		{Project: protocol.Project{Name: "b", Language: "x", Stack: "x-xlint", Tasks: map[string]protocol.Task{"lint": task("old")}}, Plugin: "xlang"},
		{Project: protocol.Project{Name: "c", Language: "x", Stack: "x-xlint"}, Plugin: "xlang"},
	}
}

func TestApplyStacksPerProject(t *testing.T) {
	set := stackSet(t)
	cfg := workspace.Config{
		DefaultStacks: map[string]string{"xlang": "x-xfmt"},
		Projects:      map[string]workspace.ProjectOverride{"c": {Stack: "x-xfmt"}},
	}
	catalog, err := pluginhost.NewCatalog(set, cfg)
	if err != nil {
		t.Fatal(err)
	}
	ps := xProjects()
	if err := ApplyStacks(ps, set, catalog, cfg); err != nil {
		t.Fatal(err)
	}
	a, b, c := ps[0], ps[1], ps[2]
	if a.Stack != "x-xfmt" || b.Stack != "x-xlint" || c.Stack != "x-xfmt" {
		t.Fatalf("stacks: default %q, manifest %q, smith.toml %q", a.Stack, b.Stack, c.Stack)
	}
	if a.Tasks["lint"].Command[0] != "xfmt" || b.Tasks["lint"].Command[0] != "xlint" {
		t.Fatalf("a stack's toolchain replaces lint: %v %v", a.Tasks["lint"], b.Tasks["lint"])
	}
	if a.Tasks["test"].Command[0] != "xtest" || a.Tasks["format"].Command[0] != "xfmt" {
		t.Fatalf("added tasks keep the project's own: %v", a.Tasks)
	}
	if !a.ServedBy("xfmt") || a.ServedBy("xlint") || !b.ServedBy("xlint") {
		t.Fatal("ServedBy covers the project's stack plugins only")
	}
	a.Conventions = []string{"xlint"}
	if got := toolset(a, set); got != "xlang@2.0.0+xfmt@1.0.0+xlint@1.0.0" {
		t.Fatalf("toolset %s", got)
	}
	if err := ApplyOverrides(a, cfg); err != nil {
		t.Fatal(err)
	}
	if a.Sources["lint.command"] != "plugin xfmt (stack x-xfmt)" || a.Sources["test.command"] != "plugin xlang" {
		t.Fatalf("sources %v", a.Sources)
	}
}

func TestApplyStacksRejectsUnknownAndForeignStacks(t *testing.T) {
	set := stackSet(t)
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{"a": {Stack: "nope"}}}
	catalog, _ := pluginhost.NewCatalog(set, cfg)
	if err := ApplyStacks(xProjects(), set, catalog, cfg); err == nil || !strings.Contains(err.Error(), `no stack named "nope"`) {
		t.Fatalf("unknown: %v", err)
	}
	ps := []*Project{{Project: protocol.Project{Name: "web", Language: "node", Stack: "x-xfmt"}, Plugin: "node"}}
	if err := ApplyStacks(ps, set, catalog, workspace.Config{}); err == nil || !strings.Contains(err.Error(), "is for xlang projects") {
		t.Fatalf("foreign: %v", err)
	}
}
