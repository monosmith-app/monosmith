package factmerge

import (
	"strings"
	"testing"
)

func TestTheNarrowestScopeWins(t *testing.T) {
	merged, err := Merge([]Layer{
		{Scope: ScopeProject, Source: "project", Facts: map[string]any{"source.roots": []any{"lib"}}},
		{Scope: ScopePluginDefault, Source: "plugin ts", Facts: map[string]any{"source.roots": []any{"src"}, "runtime.node": "22"}},
		{Scope: ScopeWorkspace, Source: "workspace", Facts: map[string]any{"source.roots": []any{"app"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := merged["source.roots"]; got.Source != "project" || got.Scope != ScopeProject {
		t.Errorf("source.roots = %+v, want the project's", got)
	}
	if got := merged["runtime.node"]; got.Source != "plugin ts" {
		t.Errorf("runtime.node = %+v, want the plugin default", got)
	}
}

func TestTwoPluginsThatDisagreeAreAnErrorNamingBoth(t *testing.T) {
	_, err := Merge([]Layer{
		{Scope: ScopePluginDefault, Source: "plugin a", Facts: map[string]any{"source.roots": []any{"src"}}},
		{Scope: ScopePluginDefault, Source: "plugin b", Facts: map[string]any{"source.roots": []any{"lib"}}},
	})
	if err == nil || !strings.Contains(err.Error(), "plugin a") || !strings.Contains(err.Error(), "plugin b") {
		t.Fatalf("err = %v, want one that names both plugins", err)
	}
}

func TestTwoPluginsThatAgreeShareTheFact(t *testing.T) {
	merged, err := Merge([]Layer{
		{Scope: ScopePluginDefault, Source: "plugin a", Facts: map[string]any{"source.roots": []string{"src"}}},
		{Scope: ScopePluginDefault, Source: "plugin b", Facts: map[string]any{"source.roots": []any{"src"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := merged["source.roots"].Source; got != "plugin a and plugin b" {
		t.Errorf("source = %q", got)
	}
}

func TestTheWorkspaceDecidesAConflictOfPlugins(t *testing.T) {
	merged, err := Merge([]Layer{
		{Scope: ScopePluginDefault, Source: "plugin a", Facts: map[string]any{"runtime.node": "20"}},
		{Scope: ScopePluginDefault, Source: "plugin b", Facts: map[string]any{"runtime.node": "22"}},
		{Scope: ScopeWorkspace, Source: "workspace", Facts: map[string]any{"runtime.node": "24"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := merged["runtime.node"]; got.Value != "24" || got.Source != "workspace" {
		t.Errorf("runtime.node = %+v, want the workspace's", got)
	}
}

func TestValuesIsNilForNoFacts(t *testing.T) {
	if Values(nil) != nil {
		t.Error("want nil")
	}
}
