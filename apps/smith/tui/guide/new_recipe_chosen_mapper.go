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

// chosenGroups is the part of a choice that holds its answers, for a level that
// takes any number of plugins: each answer gives one or more plugins.
const chosenGroups = "groups"

// recipeModeKey is the answer key of "who decides" for a level that takes any
// number of plugins, from the key of the level's tick boxes.
func recipeModeKey(levelKey string) string {
	return "mode:" + strings.TrimPrefix(levelKey, recipeLevelPrefix)
}

// recipeAdditiveSteps are the questions of a level that takes any number of
// plugins (structure, deploy, toolchain): who decides it, the tick boxes of
// the plugins the recipe fixes, or the question the users read and the
// answers, each of which gives one plugin or several joined by "+".
func recipeAdditiveSteps(fixed Step, level recipewizard.LevelOffer, plugins []recipewizard.PluginOffer) []Step {
	levelKey, skipLevel := fixed.Key, fixed.Skip
	modeKey := recipeModeKey(levelKey)
	notChosen := func(a map[string][]string) bool {
		return (skipLevel != nil && skipLevel(a)) || first(a, modeKey) != recipeChosen
	}
	mode := Step{
		Key: modeKey, Skip: skipLevel, Prompt: "Who decides the " + level.Level + "?",
		Options: []Option{
			{Label: "Fixed by the recipe (default)", Value: "fixed", Hint: "you tick the plugins; every project made from the recipe gets them"},
			{Label: "Chosen by the people who use it", Value: recipeChosen, Hint: "they answer a question; each answer gives one plugin or several"},
		},
	}
	fixed.Skip = func(a map[string][]string) bool {
		return (skipLevel != nil && skipLevel(a)) || first(a, modeKey) == recipeChosen
	}
	var names []string
	initial := make([]string, 0, len(plugins))
	for _, p := range plugins {
		names = append(names, p.Name)
		initial = append(initial, p.Name+"="+p.Name)
	}
	question := Step{
		Key: recipeChosenKey(levelKey, chosenQuestion), Skip: notChosen,
		Prompt: "What question do the users read for the " + level.Level + "?", Placeholder: "Which " + level.Level + " do you want?",
		Note: "They see it when they make a project from the recipe. Do not use a colon.",
		Check: func(_ map[string][]string, v string) string {
			if strings.ContainsAny(v, ":\n") {
				return "write the question without a colon"
			}
			return ""
		},
	}
	groups := Step{
		Key: recipeChosenKey(levelKey, chosenGroups), Skip: notChosen,
		Prompt: "Name the answers, each with its plugins: answer=plugin, answer=plugin+plugin",
		Note:   "Each plugin starts as an answer of its own. Edit the names, or join plugins that go together with +. Plugins: " + strings.Join(names, ", "),
		Initial: func(map[string][]string) string {
			return strings.Join(initial, ", ")
		},
		Check: func(a map[string][]string, v string) string {
			pairs, problem := parseGroups(v, names)
			if problem != "" {
				return problem
			}
			with := make(map[string][]string, len(a)+1)
			for k, v := range a {
				with[k] = v
			}
			with[recipeChosenKey(levelKey, chosenGroups)] = []string{strings.Join(pairs, ",")}
			return planProblem(first(a, "name"), first(a, "version"), recipeFixed(with), recipeChosenArgs(with)...)
		},
	}
	return []Step{mode, fixed, question, groups}
}

// parseGroups reads answer=plugin+plugin pairs separated by commas. It returns
// the pairs without spaces, or what is wrong: fewer than two answers, a
// repeated answer, a plugin the level does not offer.
func parseGroups(text string, offered []string) (pairs []string, problem string) {
	seen := map[string]bool{}
	for _, part := range strings.Split(text, ",") {
		answer, plugs, ok := strings.Cut(strings.TrimSpace(part), "=")
		answer, plugs = strings.TrimSpace(answer), strings.TrimSpace(plugs)
		if !ok || answer == "" || plugs == "" || strings.ContainsAny(answer, ":+") {
			return nil, "write each answer as answer=plugin, or answer=plugin+plugin"
		}
		if seen[answer] {
			return nil, fmt.Sprintf("two answers are named %q: each needs its own name", answer)
		}
		seen[answer] = true
		var clean []string
		for _, p := range strings.Split(plugs, "+") {
			p = strings.TrimSpace(p)
			if !slices.Contains(offered, p) {
				return nil, fmt.Sprintf("%q is not a plugin of this level (%s)", p, strings.Join(offered, ", "))
			}
			clean = append(clean, p)
		}
		pairs = append(pairs, answer+"="+strings.Join(clean, "+"))
	}
	if len(pairs) < 2 {
		return nil, "name at least two answers: a choice with one answer is a fixed plugin"
	}
	return pairs, ""
}

// recipeAdditiveChosenArgs are the --chosen values of the levels that take any
// number of plugins and that the users choose.
func recipeAdditiveChosenArgs(a map[string][]string) map[string]string {
	out := map[string]string{}
	for key := range a {
		rest, ok := strings.CutPrefix(key, "mode:")
		if !ok || first(a, key) != recipeChosen {
			continue
		}
		level, lang, _ := strings.Cut(rest, ":")
		if lang != recipeLanguage(a) {
			continue
		}
		levelKey := recipeLevelPrefix + rest
		var pairs []string
		for _, part := range strings.Split(first(a, recipeChosenKey(levelKey, chosenGroups)), ",") {
			if part = strings.ReplaceAll(strings.TrimSpace(part), " ", ""); part != "" {
				pairs = append(pairs, part)
			}
		}
		if len(pairs) > 0 {
			out[level] = level + "=" + strings.TrimSpace(first(a, recipeChosenKey(levelKey, chosenQuestion))) + ":" + strings.Join(pairs, ",")
		}
	}
	return out
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
	for level, arg := range recipeAdditiveChosenArgs(a) {
		byLevel[level] = arg
		levels = append(levels, level)
	}
	var out []string
	for _, level := range recipewizard.OrderLevelKeys(levels) {
		out = append(out, byLevel[level])
	}
	return out
}
