package projectgraph

import (
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/jest"
	slicesp "github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/plugin"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// connect runs a plugin in-process over pipes.
func connect(t *testing.T, p serve.Plugin) *host.Client {
	t.Helper()
	toPlugin, fromHost := io.Pipe()
	toHost, fromPlugin := io.Pipe()
	go func() { _ = serve.Serve(p, toPlugin, fromPlugin) }()
	c, err := host.Connect(toHost, fromHost, host.Options{WorkspaceRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func projects() []*Project {
	return []*Project{
		{Project: protocol.Project{Name: "tool", Language: "go", Tasks: map[string]protocol.Task{"lint": {Command: []string{"go", "vet"}}}}, Plugin: "go"},
		{Project: protocol.Project{Name: "old", Language: "go", Tasks: map[string]protocol.Task{"slice-check": {Command: []string{"mine"}}}}, Plugin: "go"},
		{Project: protocol.Project{Name: "kit", Language: "python"}, Plugin: "python"},
		{Project: protocol.Project{Name: "web", Language: "node"}, Plugin: "node"},
	}
}

func TestApplyConventions(t *testing.T) {
	set := &pluginhost.PluginSet{Clients: []*host.Client{connect(t, slicesp.New("1.0.0"))}}
	ps := projects()
	cfg := workspace.Config{Conventions: map[string]workspace.ConventionConfig{"vertical-slices": {Exclude: []string{"kit"}}}}
	applied, err := ApplyConventions(ps, set, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(applied["vertical-slices"], ",") != "tool,old,web" {
		t.Fatalf("applied to %v", applied)
	}
	tool := ps[0]
	if tool.Tasks["slice-check"].Command[0] != "{smith}" || strings.Join(tool.Tasks["lint"].DependsOn, ",") != "slice-check" {
		t.Fatalf("tool tasks %+v", tool.Tasks)
	}
	if err := ApplyOverrides(tool, cfg); err != nil {
		t.Fatal(err)
	}
	if tool.Sources["slice-check.command"] != "convention vertical-slices" || tool.Sources["lint.dependsOn"] != "convention vertical-slices" || tool.Sources["lint.command"] != "plugin go" {
		t.Fatalf("sources %v", tool.Sources)
	}
	if ps[1].Tasks["slice-check"].Command[0] != "mine" {
		t.Fatal("a task the project already has is kept")
	}
	if len(ps[2].Tasks) != 0 || len(ps[3].Tasks) != 0 {
		t.Fatal("an excluded project is left alone, and a Node project gets no slice-check task (@monosmith/eslint-config enforces its rules; doctor checks it)")
	}
}

func TestApplyConventionsOnlyWhenConfigured(t *testing.T) {
	set := &pluginhost.PluginSet{Clients: []*host.Client{connect(t, slicesp.New("1.0.0"))}}
	ps := projects()
	if applied, err := ApplyConventions(ps, set, workspace.Config{}); err != nil || len(applied) != 0 || len(ps[0].Tasks) != 1 {
		t.Fatalf("%v %v %+v", applied, err, ps[0].Tasks)
	}
}

func TestConventionWithoutTasksStillApplies(t *testing.T) {
	set := &pluginhost.PluginSet{Clients: []*host.Client{connect(t, jest.New("1.0.0")), connect(t, pythonLanguage{})}}
	ps := projects()
	applied, err := ApplyConventions(ps, set, workspace.Config{Conventions: map[string]workspace.ConventionConfig{"jest": {}}})
	if err != nil || strings.Join(applied["jest"], ",") != "web" || len(ps[3].Tasks) != 0 {
		t.Fatalf("%v %v %+v", applied, err, ps[3].Tasks)
	}
	_, err = ApplyConventions(ps, set, workspace.Config{Conventions: map[string]workspace.ConventionConfig{"python": {}}})
	if err == nil || !strings.Contains(err.Error(), "not a convention plugin") {
		t.Fatalf("a language plugin is not a convention: %v", err)
	}
}

// pythonLanguage is a language plugin named like a convention could be.
type pythonLanguage struct{}

func (pythonLanguage) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{Plugin: protocol.Peer{Name: "python"}, Kind: protocol.KindLanguage, Languages: []string{"python"}}, nil
}

func TestConventionTargetsRejectsUnsupportedLanguage(t *testing.T) {
	_, err := ConventionTargets("vertical-slices", projects(), []string{"go", "python"}, workspace.ConventionConfig{Languages: []string{"node"}})
	if err == nil || !strings.Contains(err.Error(), `does not apply to "node"`) {
		t.Fatal(err)
	}
	got, err := ConventionTargets("vertical-slices", projects(), []string{"go", "python"}, workspace.ConventionConfig{Languages: []string{"python"}})
	if err != nil || len(got) != 1 || got[0].Name != "kit" {
		t.Fatalf("%v %v", got, err)
	}
}

func TestServedBy(t *testing.T) {
	p := &Project{Plugin: "go", Conventions: []string{"vertical-slices"}}
	if !p.ServedBy("go") || !p.ServedBy("vertical-slices") || p.ServedBy("node") {
		t.Fatal("ServedBy")
	}
}
