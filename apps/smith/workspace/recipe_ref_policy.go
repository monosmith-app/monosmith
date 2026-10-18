package workspace

import (
	"path"
	"strings"
)

// ValidRecipeRef reports whether a recipe setting names a recipe: an https
// URL, or a path relative to smith.toml with forward slashes that does not
// climb out of the workspace (a recipe shared with other repositories is
// reached by URL, not by "..").
func ValidRecipeRef(ref string) bool {
	if strings.TrimSpace(ref) == "" || ref != strings.TrimSpace(ref) {
		return false
	}
	if strings.Contains(ref, "://") {
		return strings.HasPrefix(ref, "https://") && len(ref) > len("https://")
	}
	if strings.ContainsAny(ref, `\:`) || path.IsAbs(ref) {
		return false
	}
	clean := path.Clean(ref)
	return clean != ".." && !strings.HasPrefix(clean, "../")
}

// RecipeKinds are the project kinds [recipes] may name.
var RecipeKinds = []string{"app", "lib", "package"}

// RecipeRefs lists the recipes smith.toml names: [workspace] recipe, then
// [recipes] by kind (app, lib, package). Empty when it names none.
func RecipeRefs(cfg Config) []string {
	var refs []string
	if cfg.Workspace.Recipe != "" {
		refs = append(refs, cfg.Workspace.Recipe)
	}
	for _, kind := range RecipeKinds {
		if ref := cfg.Recipes[kind]; ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

// RecipeFor is the recipe new projects of a kind are created from: the one
// [recipes] names for the kind, else [workspace] recipe, else "".
func RecipeFor(cfg Config, kind string) string {
	if ref := cfg.Recipes[kind]; ref != "" {
		return ref
	}
	return cfg.Workspace.Recipe
}
