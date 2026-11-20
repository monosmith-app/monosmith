package templating

import (
	"maps"
	"slices"
	"testing"
)

func TestManifestOnlyKeepsTheTopLevelConfiguration(t *testing.T) {
	templates := map[string]string{
		"[[.Dir]]/package.json":         "{}",
		"[[.Dir]]/tsconfig.json":        "{}",
		"[[.Dir]]/[[.ID]].csproj":       "<Project/>",
		"[[.Dir]]/Program.cs":           "x",
		"[[.Dir]]/main.ts":              "x",
		"[[.Dir]]/README.md":            "# x",
		"[[.Dir]]/src/index.ts":         "x",
		"[[.Dir]]/tests/x.Tests.csproj": "<Project/>",
		"Directory.Build.props":         "<Project/>",
		".npmrc":                        "x",
	}
	got := slices.Sorted(maps.Keys(ManifestOnly(templates, []string{"*.ts", "*.cs"})))
	want := []string{"[[.Dir]]/[[.ID]].csproj", "[[.Dir]]/package.json", "[[.Dir]]/tsconfig.json"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
