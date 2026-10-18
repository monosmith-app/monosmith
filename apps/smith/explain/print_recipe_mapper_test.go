package explain

import (
	"bytes"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipedrift"
)

func TestPrintRecipeShowsEachValueWithItsRecipeAndLine(t *testing.T) {
	var out bytes.Buffer
	printRecipe(&out, recipedrift.Explained{
		Recipe: "team", Version: "1.3.0", Source: "recipes/team.recipe.toml",
		Values: []recipedrift.Value{
			{Name: "use.lint", Value: "ruff", Origin: &recipe.Origin{Recipe: "team", File: "recipes/team.recipe.toml", Line: 14}},
			{Name: "settings.ruff.line", Value: "88"},
		},
	})
	want := "recipe team 1.3.0 — recipes/team.recipe.toml\n" +
		"  use.lint            ruff\n" +
		"                        from recipes/team.recipe.toml:14 (recipe team)\n" +
		"  settings.ruff.line  88\n\n"
	if out.String() != want {
		t.Errorf("got\n%q\nwant\n%q", out.String(), want)
	}
}
