package projectgraph

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestDeclaredProjectsGoToThePluginThatOwnsTheirLanguage(t *testing.T) {
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"mvd":  {Root: "apps/mvd", Language: "go"},
		"core": {Root: "libs/core", Language: "go", Kind: "lib"},
		"api":  {Root: "services/api", Language: "py"},
		"web":  {Tags: []string{"x"}},
	}}
	got := declaredFor("go", nil, cfg)
	if len(got) != 2 || got[0] != (protocol.DeclaredProject{Name: "core", Root: "libs/core", Kind: protocol.ProjectLib}) || got[1] != (protocol.DeclaredProject{Name: "mvd", Root: "apps/mvd"}) {
		t.Fatalf("sorted by name, only this language's, kind passed: %+v", got)
	}
	if py := declaredFor("python", []string{"py", "python"}, cfg); len(py) != 1 || py[0].Name != "api" {
		t.Errorf("a plugin answers to its languages too: %+v", py)
	}
	if none := declaredFor("rust", nil, cfg); len(none) != 0 {
		t.Errorf("%+v", none)
	}
}

func TestDiscoverSettingsAreThePluginsTablesByProjectName(t *testing.T) {
	cfg := workspace.Config{Projects: map[string]workspace.ProjectOverride{
		"mvd": {Settings: map[string]map[string]any{"go": {"tags": []any{"production"}}, "docker": {"x": 1}}},
		"web": {Settings: map[string]map[string]any{"docker": {"x": 1}}},
	}}
	got := discoverSettings("go", cfg)
	if len(got) != 1 || got["mvd"]["tags"] == nil {
		t.Fatalf("only the go table of mvd: %+v", got)
	}
	if none := discoverSettings("rust", cfg); none != nil {
		t.Errorf("%+v", none)
	}
}
