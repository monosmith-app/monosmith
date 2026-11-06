package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
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
			Key: "name", Prompt: words.Say("guide.new_recipe_form.what_is_the_recipe_called_lowercase"), Placeholder: "team", Note: WhatIs("recipe"),
			Check: func(_ map[string][]string, v string) string { return planProblem(v, "", nil) },
		},
		{
			Key: "version", Prompt: words.Say("guide.new_recipe_form.version_major_minor_patch_enter_keeps"), Placeholder: "0.1.0", Optional: true,
			Check: func(a map[string][]string, v string) string {
				if v == "" {
					return ""
				}
				return planProblem(first(a, "name"), v, nil)
			},
		},
		{Key: "extends", Prompt: words.Say("guide.new_recipe_form.recipe_it_extends_a_path_or"), Placeholder: "base.recipe.toml", Optional: true, Path: true},
		{
			Key: "fixed", Prompt: words.Say("guide.new_recipe_form.fixed_plugins_as_level_plugin_separated"),
			Placeholder: words.Say("guide.new_recipe_form.language_python_lint_ruff"), Optional: true,
			Check: func(a map[string][]string, v string) string {
				return planProblem(first(a, "name"), "", strings.Fields(v))
			},
		},
	}
}

// planProblem says what is wrong with the recipe the answers so far describe,
// or "" when nothing is.
func planProblem(name, version string, fixed []string, chosen ...string) string {
	if name == "" {
		name = "team"
	}
	if _, err := recipewizard.PlanFromFlags(name, version, "", fixed, chosen); err != nil {
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
	for _, f := range recipeFixed(a) {
		argv = append(argv, "--fixed", f)
	}
	for _, c := range recipeChosenArgs(a) {
		argv = append(argv, "--chosen", c)
	}
	return argv
}
