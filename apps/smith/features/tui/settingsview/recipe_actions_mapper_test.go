package settingsview

import (
	"slices"
	"strings"
	"testing"
)

// recipeForms are the forms the shell passes, with Write a recipe among them.
func recipeForms() []Form {
	return append(placedForms(), Form{Title: "Write a recipe", Hint: "smith recipe new", Presents: "recipe new"})
}

func TestRecipesStartWithTheTwoWaysToCreateARecipe(t *testing.T) {
	m := openedWith(t, recipeForms()).Show(Recipes)
	var kinds []kind
	for _, it := range m.items(Recipes) {
		kinds = append(kinds, it.kind)
	}
	want := []kind{kindWorkspaceRecipe, kindWorkspaceRecipe, kindRecipeWizard, kindRecipe, kindRecipe}
	if !slices.Equal(kinds, want) {
		t.Fatalf("rows %v, want %v", kinds, want)
	}
	out := text(m, 120, 30)
	for _, s := range []string{"Create a recipe from this workspace (go)", "Create a recipe from this workspace (node)", "Recipe creation wizard", "team.recipe.toml"} {
		if !strings.Contains(out, s) {
			t.Errorf("the section lacks %q:\n%s", s, out)
		}
	}
}

func TestTheWorkspaceRecipeShowsWhatItFixesBeforeEnter(t *testing.T) {
	m := openedWith(t, recipeForms()).Show(Recipes)
	// The person drags the detail taller, as the layout lets them.
	m, _ = press(m.WithSizes(0, 20), "down")
	out := text(m, 140, 40)
	for _, s := range []string{"Read from 1 node project, the stack node-jest", "lint", "eslint", "toolchain", "The projects also use go"} {
		if !strings.Contains(out, s) {
			t.Errorf("the detail lacks %q:\n%s", s, out)
		}
	}
}

func TestEnterAsksTheNameAndHandsTheCommandToTheShell(t *testing.T) {
	m := openedWith(t, recipeForms()).Show(Recipes)
	m, results := press(m, "enter", "enter")
	if !m.Modal() || results[len(results)-1].Run != nil {
		t.Fatal("an empty name runs nothing")
	}
	if out := text(m, 120, 30); !strings.Contains(out, "Type the recipe's name first") || !strings.Contains(out, "vertical-slices") {
		t.Errorf("the field shows neither the problem nor what is fixed:\n%s", out)
	}
	m, results = press(m, "Bad Name", "enter")
	if results[len(results)-1].Run != nil || !strings.Contains(text(m, 120, 30), "lowercase") {
		t.Fatal("a name a recipe cannot take is refused with the reason")
	}
	m, _ = press(m, "esc", "enter")
	m, results = press(m, "team-go", "enter")
	want := []string{"recipe", "new", "--from-workspace", "--lang", "go", "--name", "team-go"}
	if got := results[len(results)-1].Run; !slices.Equal(got, want) || m.Modal() {
		t.Fatalf("run %v, want %v", got, want)
	}
}

func TestTheWizardRowOpensTheWriteARecipeForm(t *testing.T) {
	m := openedWith(t, recipeForms()).Show(Recipes)
	_, results := press(m, "down", "down", "enter")
	if len(results) == 0 || results[len(results)-1].Form != "Write a recipe" {
		t.Fatalf("%+v", results)
	}
	if kinds := openedWith(t, nil).items(Recipes); slices.ContainsFunc(kinds, func(it item) bool { return it.kind == kindRecipeWizard }) {
		t.Error("the wizard row shows without the form")
	}
}

func TestAWorkspaceWithNothingToReadSaysWhy(t *testing.T) {
	f := sample()
	f.snap.WorkspaceRecipes, f.snap.WorkspaceRecipeErr = nil, "the workspace has no projects"
	m := opened(t, f).Show(Recipes)
	if out := text(m, 120, 30); !strings.Contains(out, "nothing to read yet") || !strings.Contains(out, "has no projects") {
		t.Fatalf("%s", out)
	}
	m, results := press(m, "enter")
	if m.Modal() || results[0].Run != nil || !strings.Contains(text(m, 120, 30), "has no projects") {
		t.Fatal("Enter says why instead of asking a name")
	}
}
