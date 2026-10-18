package recipewizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

func teamPlan() Plan {
	return Plan{
		Name: "team", Version: "0.1.0",
		Choices: map[string]recipe.Choice{"language": {Ask: "Language", Options: []string{"python", "node"}, Default: "python"}},
		Levels: map[string]LevelPlan{
			"language":  {Choice: "language", Plugins: map[string]any{"python": "python", "node": "node"}},
			"lint":      {Choice: "language", Plugins: map[string]any{"python": "ruff", "node": "eslint"}},
			"test":      {Choice: "language", Plugins: map[string]any{"python": "pytest", "node": []string{"jest"}}},
			"structure": {Fixed: []string{"slices", "docker"}},
		},
	}
}

const teamRecipe = `# Recipe team, created with ` + "`smith recipe new`" + `.
# Edit it by hand if you like; ` + "`smith recipe check <this file> --all`" + ` checks it
# for every answer.
name = "team"
version = "0.1.0"

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[use]
structure = ["slices", "docker"]

[when.language.python.use]
language = "python"
lint = "ruff"
test = "pytest"

[when.language.node.use]
language = "node"
lint = "eslint"
test = "jest"
`

func TestRenderWritesTheRecipeAPersonWouldWrite(t *testing.T) {
	if got := Render(teamPlan()); got != teamRecipe {
		t.Errorf("rendered:\n%s\nwant:\n%s", got, teamRecipe)
	}
}

func TestRenderedRecipesParseAndResolve(t *testing.T) {
	doc, err := recipe.ParseRecipe([]byte(Render(teamPlan())), recipe.Source{Ref: "team.recipe.toml"})
	if err != nil {
		t.Fatal(err)
	}
	res, err := recipe.Resolve(recipe.Loaded{Chain: []recipe.Document{doc}}, map[string]string{"language": "node"})
	if err != nil {
		t.Fatal(err)
	}
	if use, ok := res.UseFor("test"); !ok || len(use.Plugins) != 1 || use.Plugins[0] != "jest" {
		t.Errorf("test = %+v", use)
	}
}

func TestRenderQuotesAnswersAndExtends(t *testing.T) {
	p := Plan{Name: "team", Version: "1.0.0", Extends: "../base.recipe.toml",
		Choices: map[string]recipe.Choice{"stack": {Options: []string{"c#", "go"}}},
		Levels:  map[string]LevelPlan{"language": {Choice: "stack", Plugins: map[string]any{"c#": "dotnet", "go": "go"}}},
	}
	out := Render(p)
	for _, want := range []string{`extends = "../base.recipe.toml"`, `options = ["c#", "go"]`, `[when.stack."c#".use]`, `language = "dotnet"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if _, err := recipe.ParseRecipe([]byte(out), recipe.Source{Ref: "x"}); err != nil {
		t.Errorf("a quoted answer does not parse: %v", err)
	}
}

func TestRenderUsesAChoiceTheParentDeclares(t *testing.T) {
	p := Plan{Name: "team", Version: "1.0.0", Extends: "base.recipe.toml",
		Levels: map[string]LevelPlan{"lint": {Choice: "language", Plugins: map[string]any{"python": "ruff", "node": "eslint"}}},
	}
	out := Render(p)
	if strings.Contains(out, "[choices.") || !strings.Contains(out, "[when.language.node.use]") || !strings.Contains(out, "[when.language.python.use]") {
		t.Errorf("a parent's choice is used, not declared again:\n%s", out)
	}
}
