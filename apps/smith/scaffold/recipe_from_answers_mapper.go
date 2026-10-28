package scaffold

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
)

// nodeStackPlugins are the plugins each Node stack smith new offers is made
// of: its scaffold and its toolchain, in order. A recipe names these, not the
// stack, so the stack is read back from the toolchain.
var nodeStackPlugins = map[string]struct {
	Scaffold  string
	Toolchain []string
}{
	"node-jest":   {"typescript", []string{"typescript", "eslint", "jest"}},
	"node-vitest": {"typescript", []string{"typescript", "eslint", "vitest"}},
	"node-ts":     {"typescript", []string{"typescript"}},
}

// planFromAnswers is the recipe plan the answers of smith new make: the main
// language, for Node its scaffold and toolchain, and the code layout as the
// structure plugin. The recipe does not hold what is not about a project's
// plugins (the CI, the release branch, the cache): those are in smith.toml.
func planFromAnswers(opts NewOptions, name string) (recipewizard.Plan, error) {
	lang := opts.Language
	if lang == "" {
		return recipewizard.Plan{}, fmt.Errorf("a recipe needs a main language; answer that question or pass --language")
	}
	levels := map[string]recipewizard.LevelPlan{recipewizard.LevelLanguage: {Fixed: []string{lang}}}
	if lang == "node" {
		stack := opts.DefaultStacks["node"]
		if stack == "" {
			stack = DefaultNodeStack
		}
		plugins, ok := nodeStackPlugins[stack]
		if !ok {
			return recipewizard.Plan{}, fmt.Errorf("no recipe for the stack %q", stack)
		}
		levels[recipewizard.LevelScaffold] = recipewizard.LevelPlan{Fixed: []string{plugins.Scaffold}}
		levels[recipewizard.LevelToolchain] = recipewizard.LevelPlan{Fixed: plugins.Toolchain}
	}
	if layout := orDefault(opts.Architecture, DefaultArchitecture); layout != "none" {
		levels[recipewizard.LevelStructure] = recipewizard.LevelPlan{Fixed: []string{layout}}
	}
	return recipewizard.Plan{Name: name, Version: "1.0.0", Levels: levels}, nil
}

// renderRecipe is the recipe file for the answers, checked the way
// `smith recipe new` checks one, so a recipe saved here resolves.
func renderRecipe(opts NewOptions, name string) (string, error) {
	plan, err := planFromAnswers(opts, name)
	if err != nil {
		return "", err
	}
	creator := recipewizard.Creator{Loader: &recipe.Loader{}, Lookup: recipe.Catalog{Version: opts.Version}.Lookup}
	text, err := creator.Render(plan, recipePath(name))
	if err != nil {
		return "", err
	}
	return strings.Replace(text, "created with `smith recipe new`", "saved by `smith new`", 1), nil
}

// recipePath is where a saved recipe goes in the workspace.
func recipePath(name string) string { return "recipes/" + name + ".recipe.toml" }

// answersFromRecipe are the answers of smith new that a recipe settles: the
// language it fixes, for Node the stack its toolchain makes, and the code
// layout of its structure plugin. A recipe that names none of these settles
// nothing, and a recipe that cannot be resolved without answers settles
// nothing either (smith add asks its choices later).
func answersFromRecipe(loaded recipe.Loaded) map[string]string {
	res, err := recipe.Resolve(loaded, nil)
	if err != nil {
		return nil
	}
	answers := map[string]string{}
	if u, ok := res.UseFor(recipewizard.LevelLanguage); ok && len(u.Plugins) == 1 {
		lang := pluginName(u.Plugins[0])
		for _, q := range NewQuestions() {
			if q.Key == AnswerLanguage && hasOption(q.Options, lang) {
				answers[AnswerLanguage] = lang
			}
		}
	}
	if answers[AnswerLanguage] == "node" {
		if u, ok := res.UseFor(recipewizard.LevelToolchain); ok {
			names := make([]string, len(u.Plugins))
			for i, p := range u.Plugins {
				names[i] = pluginName(p)
			}
			for stack, plugins := range nodeStackPlugins {
				if slices.Equal(names, plugins.Toolchain) {
					answers[AnswerNodeStack] = stack
				}
			}
		}
	}
	if u, ok := res.UseFor(recipewizard.LevelStructure); ok {
		for _, p := range u.Plugins {
			if _, known := layoutHints[pluginName(p)]; known {
				answers[AnswerArchitecture] = pluginName(p)
			}
		}
	}
	return answers
}

// pluginName drops the version constraint a recipe may put after an @.
func pluginName(plugin string) string {
	name, _, _ := strings.Cut(plugin, "@")
	return name
}
