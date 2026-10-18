package workspace

import (
	"strings"
	"testing"
)

func TestRemoveProjectTables(t *testing.T) {
	src := `[workspace]
name = "acme" # keep

[projects.web]
tag_prefix = "web-v"

[projects.web.tasks.build]
command = ["make"]

[projects.webkit]
stack = "node-ts"

[projects."web".release]
product = "web"

[tasks.lint]
cache = false
`
	want := `[workspace]
name = "acme" # keep

[projects.webkit]
stack = "node-ts"

[tasks.lint]
cache = false
`
	got, removed := RemoveProjectTables(src, "web")
	if !removed || got != want {
		t.Fatalf("removed=%v\n%s", removed, got)
	}
	if _, err := Parse([]byte(got)); err != nil {
		t.Fatal(err)
	}
	if _, removed := RemoveProjectTables(want, "api"); removed {
		t.Fatal("nothing to remove for api")
	}
}

func TestRemoveProjectTablesTakesTheRecipeComment(t *testing.T) {
	src := `[workspace]
name = "acme"

# Written by smith from the recipe team for the project billing.
[stacks.team-python-library]
language = "python"

[projects.billing.recipe]
name = "team"
source = "t.toml"

[projects.billing.settings.python]
layout = "src"

# Written by smith from the recipe team for the project invoices.
[projects.invoices.recipe]
name = "team"
source = "t.toml"
`
	got, removed := RemoveProjectTables(src, "billing")
	if !removed || strings.Contains(got, "projects.billing") || !strings.Contains(got, "[projects.invoices.recipe]") || !strings.Contains(got, "for the project invoices.") {
		t.Fatalf("removed=%v\n%s", removed, got)
	}
	// The stack is shared, so it and its comment stay; the project's own
	// comment goes with its tables.
	got, _ = RemoveProjectTables(got, "invoices")
	if strings.Contains(got, "invoices") || !strings.Contains(got, "[stacks.team-python-library]") {
		t.Fatalf("after removing both:\n%s", got)
	}
	if _, err := Parse([]byte(got)); err != nil {
		t.Fatal(err)
	}
}
