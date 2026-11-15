package workspace

import "testing"

const withRecipes = `# Hand-written, kept.
[workspace]
name = "acme"

# Written by smith from the recipe team for the project billing.
[stacks.team-python]
language = "python"
toolchain = ["ruff"]

[projects.billing.recipe]
name = "team"
version = "1.2.0"

[projects.billing.recipe.answers]
language = "python"

[projects.billing.settings.ruff]
line = 100

[projects.billing.tasks.test]
cache = false

# Written by smith from the recipe team for the project invoices.
[projects.invoices.recipe]
name = "team"
version = "1.2.0"

[projects.invoices.settings.ruff]
line = 100
`

func TestRemoveRecipeTablesKeepsEverythingElse(t *testing.T) {
	got, removed := RemoveRecipeTables(withRecipes, []string{"billing"}, []string{"team-python"})
	if !removed {
		t.Fatal("nothing was removed")
	}
	want := `# Hand-written, kept.
[workspace]
name = "acme"

[projects.billing.tasks.test]
cache = false

# Written by smith from the recipe team for the project invoices.
[projects.invoices.recipe]
name = "team"
version = "1.2.0"

[projects.invoices.settings.ruff]
line = 100
`
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestRemoveRecipeTablesOnlyTakesTheNamedProject(t *testing.T) {
	got, _ := RemoveRecipeTables(withRecipes, []string{"invoices"}, nil)
	want := withRecipes[:len(withRecipes)-len(`
# Written by smith from the recipe team for the project invoices.
[projects.invoices.recipe]
name = "team"
version = "1.2.0"

[projects.invoices.settings.ruff]
line = 100
`)]
	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}

func TestRemoveRecipeTablesAcceptsQuotedNamesAndReportsNothingToRemove(t *testing.T) {
	src := "[projects.\"my-app\".recipe]\nname = \"team\"\n\n[projects.\"my-app\".settings.ruff]\nline = 1\n\n[stacks.\"team-x\"]\nlanguage = \"go\"\n"
	got, removed := RemoveRecipeTables(src, []string{"my-app"}, []string{"team-x"})
	if !removed || got != "" {
		t.Errorf("quoted keys: removed=%v got=%q", removed, got)
	}
	if same, removed := RemoveRecipeTables(withRecipes, []string{"nobody"}, []string{"nothing"}); removed || same != withRecipes {
		t.Errorf("an unknown project changed the text: removed=%v", removed)
	}
}
