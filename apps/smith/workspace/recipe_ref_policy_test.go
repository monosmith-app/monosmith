package workspace

import (
	"path/filepath"
	"testing"
)

func TestRecipeLocationIsAbsoluteForAPathAndLeavesAURL(t *testing.T) {
	if got := RecipeLocation("/ws", "https://example.com/t.toml"); got != "https://example.com/t.toml" {
		t.Errorf("url: %q", got)
	}
	if got := RecipeLocation(filepath.FromSlash("/ws"), "recipes/team.toml"); got != filepath.Join(filepath.FromSlash("/ws"), "recipes", "team.toml") {
		t.Errorf("path: %q", got)
	}
}

func TestValidRecipeRef(t *testing.T) {
	for ref, want := range map[string]bool{
		"recipes/team.recipe.toml":              true,
		"team.recipe.toml":                      true,
		"https://example.com/recipes/team.toml": true,
		"http://example.com/team.toml":          false,
		"https://":                              false,
		"":                                      false,
		" team.toml":                            false,
		`recipes\team.toml`:                     false,
		"/etc/team.toml":                        false,
		"C:/team.toml":                          false,
		"../shared/team.toml":                   false,
		"recipes/../../shared/team.toml":        false,
		"recipes/../team.toml":                  true,
		"file:///team.toml":                     false,
	} {
		if got := ValidRecipeRef(ref); got != want {
			t.Errorf("ValidRecipeRef(%q) = %v, want %v", ref, got, want)
		}
	}
}

func TestRecipeForPrefersTheKind(t *testing.T) {
	cfg := Config{Workspace: WorkspaceSection{Recipe: "team.toml"}, Recipes: map[string]string{"package": "pkg.toml"}}
	if got := RecipeFor(cfg, "package"); got != "pkg.toml" {
		t.Errorf("package: %q", got)
	}
	if got := RecipeFor(cfg, "app"); got != "team.toml" {
		t.Errorf("app: %q", got)
	}
	if got := RecipeFor(Config{}, "app"); got != "" {
		t.Errorf("none: %q", got)
	}
	if got := RecipeRefs(cfg); len(got) != 2 || got[0] != "team.toml" || got[1] != "pkg.toml" {
		t.Errorf("refs: %v", got)
	}
}
