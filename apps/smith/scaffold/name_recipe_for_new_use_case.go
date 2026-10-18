package scaffold

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// checkRecipe looks at the recipe `smith new --recipe` is given, before
// anything is created. A URL must be reachable and a recipe: nothing is
// pinned yet (the first `smith add` pins it). A path is relative to the new
// smith.toml; one that is not there yet is only said, since a recipe kept in
// the same repository may be added next.
func checkRecipe(dir, ref string, out io.Writer) error {
	if ref == "" {
		return nil
	}
	if !workspace.ValidRecipeRef(ref) {
		return fmt.Errorf("--recipe %q is not an https URL or a path relative to smith.toml with forward slashes", ref)
	}
	location := workspace.RecipeLocation(dir, ref)
	if !strings.Contains(ref, "://") {
		if _, err := os.Stat(location); errors.Is(err, fs.ErrNotExist) {
			fmt.Fprintf(out, "  the recipe %s is not there yet; add it before `smith add`\n", ref)
			return nil
		}
	}
	if _, err := (&recipe.Loader{}).Load(location); err != nil {
		return fmt.Errorf("--recipe %s: %w", ref, err)
	}
	return nil
}

// withRecipe puts [workspace] recipe in the configuration text smith new
// writes, right after the module prefix.
func withRecipe(config, modulePrefix, ref string) string {
	after := fmt.Sprintf("module_prefix = %q\n", modulePrefix)
	line := "# The recipe `smith add` creates projects from (docs/reference/recipes.md).\n" + fmt.Sprintf("recipe = %q\n", ref)
	return strings.Replace(config, after, after+line, 1)
}
