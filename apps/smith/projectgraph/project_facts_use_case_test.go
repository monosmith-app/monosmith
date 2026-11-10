package projectgraph

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func factsProject() *Project {
	return &Project{Project: protocol.Project{Name: "web", Root: "apps/web"}, Plugin: "node"}
}

func TestProjectFactsBeatWorkspaceFactsBeatPluginDefaults(t *testing.T) {
	cfg := workspace.Config{
		Facts: map[string]any{"module.aliases": map[string]any{"@app": "src"}, "source.roots": []any{"app"}},
		Projects: map[string]workspace.ProjectOverride{
			"web": {Facts: map[string]any{"source.roots": []any{"lib"}}},
		},
	}
	p := factsProject()
	if err := SeedFacts([]*Project{p}, cfg); err != nil {
		t.Fatal(err)
	}
	merge(p, protocol.Contribution{Facts: map[string]any{"source.roots": []any{"src"}, "runtime.node": "22"}}, "plugin typescript")
	if err := ResolveFacts([]*Project{p}); err != nil {
		t.Fatal(err)
	}
	if got := p.FactDetail["source.roots"]; got.Source != "smith.toml [projects.web.facts]" {
		t.Errorf("source.roots came from %q, want the project table", got.Source)
	}
	if got := p.FactDetail["module.aliases"]; got.Source != "smith.toml [facts]" {
		t.Errorf("module.aliases came from %q, want [facts]", got.Source)
	}
	if got := p.Facts["runtime.node"]; got != "22" {
		t.Errorf("runtime.node = %v, want the plugin default in the request facts", got)
	}
}

func TestTwoPluginsWithDifferentDefaultsNameBothInTheError(t *testing.T) {
	p := factsProject()
	if err := SeedFacts([]*Project{p}, workspace.Config{}); err != nil {
		t.Fatal(err)
	}
	merge(p, protocol.Contribution{Facts: map[string]any{"runtime.node": "20"}}, "plugin a")
	merge(p, protocol.Contribution{Facts: map[string]any{"runtime.node": "22"}}, "plugin b")
	err := ResolveFacts([]*Project{p})
	if err == nil || !strings.Contains(err.Error(), "plugin a") || !strings.Contains(err.Error(), "plugin b") || !strings.Contains(err.Error(), "web") {
		t.Fatalf("err = %v", err)
	}
}

func TestAnInvalidFactInSmithTomlNamesTheTable(t *testing.T) {
	err := SeedFacts([]*Project{factsProject()}, workspace.Config{Facts: map[string]any{"source.roots": "src"}})
	if err == nil || !strings.Contains(err.Error(), "smith.toml [facts]") {
		t.Fatalf("err = %v", err)
	}
}

func TestConfigFactsIsTheGenerateRequestView(t *testing.T) {
	cfg := workspace.Config{Facts: map[string]any{"runtime.node": "22"}}
	facts, err := ConfigFacts(cfg, "new")
	if err != nil || facts["runtime.node"] != "22" {
		t.Fatalf("%v %v", facts, err)
	}
	if none, _ := ConfigFacts(workspace.Config{}, "new"); none != nil {
		t.Errorf("no facts must be nil, got %v", none)
	}
}
