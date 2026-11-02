package guide

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipewizard"
)

// recipeLevelPrefix starts the answer key of a level of the recipe form:
// "level:<level>:<language>", where the language is the plugin the recipe
// fixes for its language, or empty when it fixes none.
const recipeLevelPrefix = "level:"

// recipeUnset is the value of the option that leaves a level to the recipe's
// users, or to the recipe it extends.
const recipeUnset = "unset"

// ParseRecipeLevels reads the output of `smith recipe new --list-levels --json`.
//
// text is that output. It returns the catalog, and an error when the text is
// not the JSON smith prints.
func ParseRecipeLevels(text string) (recipewizard.LevelCatalog, error) {
	var cat recipewizard.LevelCatalog
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &cat); err != nil {
		return recipewizard.LevelCatalog{}, fmt.Errorf("smith recipe new did not print its levels: %w", err)
	}
	return cat, nil
}

// recipeLevelKey is the answer key of a level, for the language the recipe fixes.
func recipeLevelKey(level, language string) string {
	return recipeLevelPrefix + level + ":" + language
}

// recipeLanguage is the language plugin the answers fix, or "" for none.
func recipeLanguage(a map[string][]string) string {
	if v := first(a, recipeLevelKey(recipewizard.LevelLanguage, "")); v != recipeUnset && v != recipeChosen {
		return v
	}
	return ""
}

// recipeFixed are the level=plugin values the answers fix, in the order a
// recipe is written in: those of the pick lists and tick boxes, then those
// typed in the plain form.
func recipeFixed(a map[string][]string) []string {
	byLevel := map[string][]string{}
	for key, values := range a {
		rest, ok := strings.CutPrefix(key, recipeLevelPrefix)
		if !ok {
			continue
		}
		level, _, _ := strings.Cut(rest, ":")
		for _, v := range values {
			if v != "" && v != recipeUnset && v != recipeChosen {
				byLevel[level] = append(byLevel[level], v)
			}
		}
	}
	var out []string
	levels := make([]string, 0, len(byLevel))
	for level := range byLevel {
		levels = append(levels, level)
	}
	for _, level := range recipewizard.OrderLevelKeys(levels) {
		out = append(out, level+"="+strings.Join(byLevel[level], ","))
	}
	return append(out, strings.Fields(first(a, "fixed"))...)
}

// recipeLevelCheck refuses a pick that makes a recipe `smith recipe new`
// would refuse, with its reason, before the next question.
func recipeLevelCheck(key string) func(map[string][]string, string) string {
	return func(a map[string][]string, value string) string {
		with := make(map[string][]string, len(a)+1)
		for k, v := range a {
			with[k] = v
		}
		with[key] = strings.Split(value, ",")
		return planProblem(first(a, "name"), first(a, "version"), recipeFixed(with), recipeChosenArgs(with)...)
	}
}

// RecipeAction is the "Write a recipe" form with a question for each level a
// plugin can fill, in place of one line of level=plugin text: a pick list for
// a level that takes one plugin (the language, the scaffold, a task slot), tick
// boxes for a level that takes any number (structure, deploy, toolchain). Each
// plugin shows what it is, and "Leave unset" is the first and default answer.
// After the language, only the plugins that fit it are offered.
//
// base is the plain form and cat the levels `smith recipe new --list-levels
// --json` printed. A catalog with no level returns the plain form. A level that
// takes one plugin ends its list with "Chosen by the people who use it": the
// form then asks the question the users read, the plugins they can end up
// with and the name of each answer, and writes one --chosen. A level that
// takes any number of plugins (structure, deploy, toolchain) is fixed only.
func RecipeAction(base Action, cat recipewizard.LevelCatalog) Action {
	if len(cat.Levels) == 0 {
		return base
	}
	var steps []Step
	for _, s := range base.Steps {
		if s.Key != "fixed" {
			steps = append(steps, s)
		}
	}
	languages := make([]string, 0, len(cat.Fits))
	for l := range cat.Fits {
		languages = append(languages, l)
	}
	slices.Sort(languages)
	for _, level := range cat.Levels {
		descriptions := map[string]string{}
		for _, p := range level.Plugins {
			descriptions[p.Name] = p.Description
		}
		if level.Level == recipewizard.LevelLanguage {
			steps = append(steps, recipeLevelSteps(level, "", level.Plugins)...)
			continue
		}
		steps = append(steps, recipeLevelSteps(level, "", level.Plugins)...)
		for _, lang := range languages {
			var offers []recipewizard.PluginOffer
			for _, name := range cat.Fits[lang][level.Level] {
				offers = append(offers, recipewizard.PluginOffer{Name: name, Description: descriptions[name]})
			}
			if len(offers) > 0 {
				steps = append(steps, recipeLevelSteps(level, lang, offers)...)
			}
		}
	}
	base.Steps = steps
	return base
}

// recipeLevelSteps are the questions for one level: the pick of the plugin,
// and for a level that takes one plugin the three questions of "chosen by the
// people who use it". lang is the language plugin the recipe fixes.
func recipeLevelSteps(level recipewizard.LevelOffer, lang string, plugins []recipewizard.PluginOffer) []Step {
	step := recipeLevelStep(level, lang, plugins)
	if level.Additive || len(plugins) < 2 {
		return []Step{step}
	}
	step.Options = append(step.Options, recipeChosenOption(level.Level))
	return append([]Step{step}, recipeChosenSteps(step.Key, level.Level, plugins, step.Skip)...)
}

// recipeLevelStep is the question for one level: lang is the language plugin
// the recipe fixes ("" for none, which offers every plugin of the level).
func recipeLevelStep(level recipewizard.LevelOffer, lang string, plugins []recipewizard.PluginOffer) Step {
	key := recipeLevelKey(level.Level, lang)
	unset := "Nothing is written for " + level.Level + ": the recipe it extends, or the person who uses the recipe, decides."
	s := Step{Key: key, Check: recipeLevelCheck(key), Optional: level.Additive}
	switch {
	case level.Level == recipewizard.LevelLanguage:
		s.Prompt = "Which language plugin does the recipe fix? Every other question follows this answer."
		s.Note = unset + "\nThe last answer lets the people who use the recipe choose between plugins."
	case level.Additive:
		s.Multi = true
		s.Prompt = fmt.Sprintf("Which %s plugins does the recipe fix? Space ticks one, enter confirms.", level.Level)
		s.Note = "Tick none to leave " + level.Level + " unset. " + unset
	default:
		s.Prompt = fmt.Sprintf("Which %s plugin does the recipe fix? Projects made from the recipe get it.", level.Level)
		s.Note = unset
	}
	if lang != "" {
		s.Skip = func(a map[string][]string) bool { return recipeLanguage(a) != lang }
		s.Note += "\nOnly plugins that fit " + lang + " are listed."
	} else if level.Level != recipewizard.LevelLanguage {
		s.Skip = func(a map[string][]string) bool { return recipeLanguage(a) != "" }
	}
	if !level.Additive {
		s.Options = append(s.Options, Option{Label: "Leave unset (default)", Value: recipeUnset, Hint: "no " + level.Level + " plugin is fixed by this recipe"})
	}
	for _, p := range plugins {
		s.Options = append(s.Options, Option{Label: p.Name, Value: p.Name, Hint: p.Description})
	}
	return s
}
