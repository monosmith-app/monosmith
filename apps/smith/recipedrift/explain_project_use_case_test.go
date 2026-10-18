package recipedrift

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

func TestValuesListChoicesPluginsAndSettingsWithTheirOrigin(t *testing.T) {
	res := recipe.Resolution{
		Asked: []recipe.Question{
			{Name: "language", Answer: "python", Origin: recipe.Origin{Recipe: "team", File: "team.recipe.toml", Line: 4}},
			{Name: "kind", Origin: recipe.Origin{Recipe: "team", File: "team.recipe.toml", Line: 8}},
		},
		Uses: []recipe.Use{
			{Key: "language", Plugins: []string{"python"}, Origin: recipe.Origin{Recipe: "team", File: "team.recipe.toml", Line: 14}},
			{Key: "lint", Plugins: []string{"ruff@^0.5"}, Origin: recipe.Origin{Recipe: "base", File: "base.recipe.toml", Line: 3}},
		},
		Settings: map[string]any{"ruff": map[string]any{"line": 88, "extend": map[string]any{"select": []string{"E"}}}},
		Origins: map[string]recipe.Origin{
			"settings.ruff.line":          {Recipe: "team", File: "team.recipe.toml", Line: 21},
			"settings.ruff.extend.select": {Recipe: "base", File: "base.recipe.toml", Line: 9},
		},
	}
	got := values(res)
	type row struct {
		name, value, from string
	}
	var rows []row
	for _, v := range got {
		from := ""
		if v.Origin != nil {
			from = v.Origin.String()
		}
		rows = append(rows, row{v.Name, v.Value, from})
	}
	want := []row{
		{"choices.language", "python", "team.recipe.toml:4 (recipe team)"},
		{"use.language", "python", "team.recipe.toml:14 (recipe team)"},
		{"use.lint", "ruff@^0.5", "base.recipe.toml:3 (recipe base)"},
		{"settings.ruff.extend.select", `["E"]`, "base.recipe.toml:9 (recipe base)"},
		{"settings.ruff.line", "88", "team.recipe.toml:21 (recipe team)"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("got\n%v\nwant\n%v", rows, want)
	}
}

func TestRelativeToShowsARecipeFileTheWaySmithTomlNamesIt(t *testing.T) {
	root := filepath.FromSlash("/ws")
	for file, want := range map[string]string{
		filepath.FromSlash("/ws/recipes/team.recipe.toml"): "recipes/team.recipe.toml",
		filepath.FromSlash("/elsewhere/team.recipe.toml"):  filepath.FromSlash("/elsewhere/team.recipe.toml"),
		"https://example.com/team.recipe.toml":             "https://example.com/team.recipe.toml",
	} {
		if got := relativeTo(root, file); got != want {
			t.Errorf("%s: got %s want %s", file, got, want)
		}
	}
}

func TestValuesOfAnEmptyResolution(t *testing.T) {
	if got := values(recipe.Resolution{}); len(got) != 0 {
		t.Errorf("%v", got)
	}
}
