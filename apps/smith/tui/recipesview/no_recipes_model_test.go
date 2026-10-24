package recipesview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipecatalog"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAWorkspaceWithoutRecipesExplainsThemAndOffersOnlyTheWizard(t *testing.T) {
	v := New().FromLoaded(Loaded{Catalog: Catalog{ByKind: map[string]string{}}})
	if !v.NoRecipes() {
		t.Fatal("an empty catalog has no recipes")
	}
	lines, _ := v.View(theme.Default(), 80)
	out := pane.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(out, "What is a recipe?") || !strings.Contains(out, "You do not need one") {
		t.Fatalf("the screen should lead with what a recipe is:\n%s", out)
	}
	if strings.Contains(out, "Kind") || strings.Contains(out, "Create ") {
		t.Fatalf("the project form must be hidden:\n%s", out)
	}
	argv, _ := v.Argv()
	if strings.Join(argv, " ") != "recipe new team.recipe.toml" {
		t.Fatalf("only the recipe wizard can run, got %v", argv)
	}
	for _, l := range lines {
		if pane.Width(l) > 80 {
			t.Fatalf("line too wide: %q", pane.Strip(l))
		}
	}
}

func TestAWorkspaceWithARecipeKeepsTheForm(t *testing.T) {
	v := New().FromLoaded(Loaded{Catalog: Catalog{Recipes: []recipecatalog.Listed{{Path: "team.recipe.toml"}}}})
	if v.NoRecipes() {
		t.Fatal("a recipe file is present")
	}
	lines, _ := v.View(theme.Default(), 100)
	if out := pane.Strip(strings.Join(lines, "\n")); !strings.Contains(out, "Kind") {
		t.Fatalf("the form should show:\n%s", out)
	}
}
