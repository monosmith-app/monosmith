package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
)

// NewRecipeSteps are the questions of "Write a recipe": its name, its version,
// the recipe it extends and the plugins that are fixed, as level=plugin. The
// rules are the ones `smith recipe new` enforces (recipewizard.PlanFromFlags),
// so a form that accepts an answer gives a command that accepts it. A choice
// the recipe's users make is not asked here: `smith recipe new` asks it on a
// terminal.
func NewRecipeSteps() []Step {
	return []Step{
		{
			Key: "name", Prompt: "What is the recipe called? (lowercase words joined by dashes)", Placeholder: "team",
			Check: func(_ map[string][]string, v string) string { return planProblem(v, "", nil) },
		},
		{
			Key: "version", Prompt: "Version (major.minor.patch). Enter keeps 0.1.0.", Placeholder: "0.1.0", Optional: true,
			Check: func(a map[string][]string, v string) string {
				if v == "" {
					return ""
				}
				return planProblem(first(a, "name"), v, nil)
			},
		},
		{Key: "extends", Prompt: "Recipe it extends (a path or an https URL; enter for none)", Placeholder: "base.recipe.toml", Optional: true},
		{
			Key: "fixed", Prompt: "Fixed plugins as level=plugin, separated by spaces (language=python lint=ruff structure=slices,layers)",
			Placeholder: "language=python lint=ruff", Optional: true,
			Check: func(a map[string][]string, v string) string {
				return planProblem(first(a, "name"), "", strings.Fields(v))
			},
		},
	}
}

// planProblem says what is wrong with the recipe the answers so far describe,
// or "" when nothing is.
func planProblem(name, version string, fixed []string) string {
	if name == "" {
		name = "team"
	}
	if _, err := recipewizard.PlanFromFlags(name, version, "", fixed); err != nil {
		return err.Error()
	}
	return ""
}

// NewRecipeArgv is the command the answers build: smith recipe new with the
// name, the version and the extends it was given, and one --fixed for each
// fixed plugin.
func NewRecipeArgv(a map[string][]string) []string {
	argv := []string{"recipe", "new", "--name", first(a, "name")}
	if v := first(a, "version"); v != "" {
		argv = append(argv, "--recipe-version", v)
	}
	if e := first(a, "extends"); e != "" {
		argv = append(argv, "--extends", e)
	}
	for _, f := range strings.Fields(first(a, "fixed")) {
		argv = append(argv, "--fixed", f)
	}
	return argv
}
