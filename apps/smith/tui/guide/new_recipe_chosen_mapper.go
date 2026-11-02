package guide

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
)

// recipeChosen is the value of the answer that leaves a level to the people
// who use the recipe: they answer a question, and each answer gives a plugin
// (`smith recipe new --chosen`, ADR 0079).
const recipeChosen = "chosen"

// chosenKeyPart names the sub-questions of a level that its users choose.
const (
	chosenQuestion = "question"
	chosenPlugins  = "plugins"
	chosenAnswers  = "answers"
)

// recipeChosenKey is the answer key of a sub-question of the level whose
// pick has the key levelKey.
func recipeChosenKey(levelKey, part string) string {
	return "chosen:" + strings.TrimPrefix(levelKey, recipeLevelPrefix) + ":" + part
}

// recipeChosenOption is the last answer of a level that takes one plugin:
// the recipe does not fix it, the people who use the recipe choose.
func recipeChosenOption(level string) Option {
	return Option{
		Label: "Chosen by the people who use it", Value: recipeChosen,
		Hint: "smith asks them a question for the " + level + " when they make a project; each answer gives a plugin. You write the question and the answers next",
	}
}

// recipeChosenSteps are the sub-questions of a level the users choose: the
// question they read, the plugins they can end up with, and the name of each
// answer. They are asked only when the pick of the level says so, and only
// where the level's own question is asked.
//
// levelKey is the key of the pick, level its name, plugins the ones that fit
// and skipLevel says when the level's own question is left out.
func recipeChosenSteps(levelKey, level string, plugins []recipewizard.PluginOffer, skipLevel func(map[string][]string) bool) []Step {
	skip := func(a map[string][]string) bool {
		return (skipLevel != nil && skipLevel(a)) || first(a, levelKey) != recipeChosen
	}
	picks := Step{
		Key: recipeChosenKey(levelKey, chosenPlugins), Multi: true, Skip: skip,
		Prompt: fmt.Sprintf("Which %s plugins can the users end up with? Space ticks one, enter confirms (two at least).", level),
		Check: func(_ map[string][]string, v string) string {
			if len(strings.Split(v, ",")) < 2 {
				return "tick at least two plugins: a choice with one answer is a fixed plugin"
			}
			return ""
		},
	}
	for _, p := range plugins {
		picks.Options = append(picks.Options, Option{Label: p.Name, Value: p.Name, Hint: p.Description})
	}
	return []Step{
		{
			Key: recipeChosenKey(levelKey, chosenQuestion), Skip: skip,
			Prompt: "What question do the users read for the " + level + "?", Placeholder: "Which " + level + " do you want?",
			Note: "They see it when they make a project from the recipe. Do not use a colon.",
			Check: func(_ map[string][]string, v string) string {
				if strings.ContainsAny(v, ":\n") {
					return "write the question without a colon"
				}
				return ""
			},
		},
		picks,
		{
			Key: recipeChosenKey(levelKey, chosenAnswers), Skip: skip,
			Prompt: "Name each answer, separated by commas, in the order of the plugins you ticked.",
			Note:   "The answer is what the user reads; the plugin behind it is what the project gets. Enter keeps the plugin names.",
			Initial: func(a map[string][]string) string {
				return strings.Join(a[recipeChosenKey(levelKey, chosenPlugins)], ", ")
			},
			Check: func(a map[string][]string, v string) string {
				answers := splitAnswers(v)
				if want := len(a[recipeChosenKey(levelKey, chosenPlugins)]); len(answers) != want {
					return fmt.Sprintf("name %d answers, one for each plugin you ticked (you named %d)", want, len(answers))
				}
				for i, answer := range answers {
					if strings.ContainsAny(answer, ":=+") || answer == "" {
						return "an answer is a word or two, without : = or +"
					}
					if slices.Contains(answers[:i], answer) {
						return fmt.Sprintf("two answers are named %q: each needs its own name", answer)
					}
				}
				with := make(map[string][]string, len(a)+1)
				for k, v := range a {
					with[k] = v
				}
				with[recipeChosenKey(levelKey, chosenAnswers)] = []string{v}
				return planProblem(first(a, "name"), first(a, "version"), recipeFixed(with), recipeChosenArgs(with)...)
			},
		},
	}
}

// splitAnswers are the names typed for the answers of a choice.
func splitAnswers(v string) []string {
	var out []string
	for _, part := range strings.Split(v, ",") {
		out = append(out, strings.TrimSpace(part))
	}
	return out
}

// recipeChosenArgs are the --chosen values of the levels the users choose,
// as level=question:answer=plugin,answer=plugin, in the order a recipe is
// written in. A level asked for another language than the recipe's is left
// out.
func recipeChosenArgs(a map[string][]string) []string {
	byLevel := map[string]string{}
	var levels []string
	for key, v := range a {
		rest, ok := strings.CutPrefix(key, recipeLevelPrefix)
		if !ok || first(a, key) != recipeChosen || len(v) == 0 {
			continue
		}
		level, lang, _ := strings.Cut(rest, ":")
		if lang != recipeLanguage(a) && level != recipewizard.LevelLanguage {
			continue
		}
		plugins := a[recipeChosenKey(key, chosenPlugins)]
		answers := splitAnswers(first(a, recipeChosenKey(key, chosenAnswers)))
		if len(plugins) == 0 || len(plugins) != len(answers) {
			continue
		}
		pairs := make([]string, len(plugins))
		for i := range plugins {
			pairs[i] = answers[i] + "=" + plugins[i]
		}
		byLevel[level] = level + "=" + strings.TrimSpace(first(a, recipeChosenKey(key, chosenQuestion))) + ":" + strings.Join(pairs, ",")
		levels = append(levels, level)
	}
	var out []string
	for _, level := range recipewizard.OrderLevelKeys(levels) {
		out = append(out, byLevel[level])
	}
	return out
}
