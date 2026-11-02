package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// Keys of the environment questions of the add form (ADR 0081).
const (
	envsKey       = "environments"
	envNamesKey   = "environment-names"
	envDefaultKey = "environment-default"
	envOutputKey  = "environment-output"
	envFilesKey   = "env-files"
)

// Answers of the first environment question.
const (
	envsDefault = "default"
	envsCustom  = "custom"
	envsNone    = "none"
)

// webChoices are the choices that make a web app built with Vite.
var webChoices = []string{"react", "vue", "svelte", "html"}

// pickedChoices are the choices the person ticked in the add form: the ones of
// the node choices step and of its axis questions, and the outcome they picked.
func pickedChoices(a map[string][]string, offers []ChoiceOffer) []string {
	var picked []string
	for _, lang := range offerLanguages(offers) {
		picked = append(picked, a[withKeyPrefix+lang]...)
		for _, axis := range askedAxes(offersOf(offers, lang)) {
			picked = append(picked, a[axisKey(lang, axis)]...)
		}
	}
	return picked
}

// isWebAnswer reports whether the form describes a web app: an app made from
// react, vue, svelte or html.
func isWebAnswer(a map[string][]string, offers []ChoiceOffer) bool {
	if first(a, "lang") != "node" {
		return false
	}
	for _, c := range pickedChoices(a, offers) {
		if slices.Contains(webChoices, c) {
			return true
		}
	}
	return false
}

// webEnvironmentSteps are the questions about the environment builds of a web
// app: the names (the default four, names of your own, or one build), the
// environment that `smith run build` builds, where the builds go and whether
// the .env files are committed. Each has its default first and says what it
// costs. They are asked only for a web app.
func webEnvironmentSteps(offers []ChoiceOffer) []Step {
	notWeb := func(a map[string][]string) bool { return !isWebAnswer(a, offers) }
	none := func(a map[string][]string) bool { return notWeb(a) || first(a, envsKey) == envsNone }
	defaults := strings.Join(webenv.DefaultNames, ", ")
	return []Step{
		{Key: envsKey, Prompt: "A web app is built once per environment. Which environments?", Note: "Vite writes its VITE_ values into the bundle when it builds, so each environment is its own build with its own .env file.", Skip: notWeb, Options: []Option{
			{Label: defaults, Value: envsDefault, Hint: "default: one build and one .env file for each; lcl is your machine"},
			{Label: "Names of my own", Value: envsCustom, Hint: "you type the names: lowercase words, such as qa, stage"},
			{Label: "One build only", Value: envsNone, Hint: "smith add --environments none: a single vite build, as before"},
		}},
		{Key: envNamesKey, Prompt: "Environment names, separated by commas", Placeholder: defaults,
			Initial: func(map[string][]string) string { return strings.Join(webenv.DefaultNames, ", ") },
			Skip:    func(a map[string][]string) bool { return notWeb(a) || first(a, envsKey) != envsCustom },
			Check: func(_ map[string][]string, value string) string {
				if err := webenv.ValidateNames(webenv.ParseList(value)); err != nil {
					return err.Error()
				}
				return ""
			}},
		{Key: envDefaultKey, Prompt: "Which environment is `smith run build`? CI and the release build this one", Skip: none, OptionsFor: envDefaultOptions},
		{Key: envOutputKey, Prompt: "Where do the builds go?", Skip: none, Options: []Option{
			{Label: "In the workspace", Value: webenv.OutputWorkspace, Hint: "default: dist/apps/<app>/<name>, with every other build"},
			{Label: "In the project", Value: webenv.OutputProject, Hint: "dist/<name> in the app's own folder; smith doctor then accepts it"},
		}},
		{Key: envFilesKey, Prompt: "The .env files hold VITE_ values, which end up in the browser, so no secret. Commit them?", Skip: none, Options: []Option{
			{Label: "Yes, commit the .env files", Value: webenv.FilesCommit, Hint: "default: CI builds with them"},
			{Label: "Only .env.example", Value: webenv.FilesExample, Hint: "each machine and CI job makes its own files; git ignores them"},
		}},
	}
}

// envNamesOf are the environment names the answers so far name: the default
// four, or the ones typed.
func envNamesOf(a map[string][]string) []string {
	if first(a, envsKey) == envsCustom {
		return webenv.ParseList(first(a, envNamesKey))
	}
	return webenv.DefaultNames
}

// envDefaultOptions lists the environments for the question of which one `smith
// run build` builds: the one smith would take (prd, else the last) first and
// marked, then the others.
func envDefaultOptions(a map[string][]string) []Option {
	names := envNamesOf(a)
	fallback := webenv.Settings{Names: names}.Normalized().Default
	out := []Option{}
	if fallback != "" {
		out = append(out, Option{Label: fallback, Value: fallback, Hint: "default"})
	}
	for _, n := range slices.Backward(names) {
		if n != fallback {
			out = append(out, Option{Label: n, Value: n, Hint: "the others are build:<name>"})
		}
	}
	return out
}

// webEnvironmentArgs are the flags of smith add that carry the answers of the
// environment questions; nothing for an app that is not a web app.
func webEnvironmentArgs(a map[string][]string, offers []ChoiceOffer) []string {
	if !isWebAnswer(a, offers) {
		return nil
	}
	switch first(a, envsKey) {
	case envsNone:
		return []string{"--environments", "none"}
	case envsCustom:
		names := strings.Join(webenv.ParseList(first(a, envNamesKey)), ",")
		return []string{"--environments", names, "--default-environment", first(a, envDefaultKey), "--environment-output", first(a, envOutputKey), "--env-files", first(a, envFilesKey)}
	}
	return []string{"--environments", strings.Join(webenv.DefaultNames, ","), "--default-environment", first(a, envDefaultKey), "--environment-output", first(a, envOutputKey), "--env-files", first(a, envFilesKey)}
}

// withWebEnvironments adds the environment questions after the others of the
// add form, and their flags to its command.
func withWebEnvironments(action Action, offers []ChoiceOffer) Action {
	if !slices.ContainsFunc(offers, func(o ChoiceOffer) bool { return slices.Contains(webChoices, o.Name) }) {
		return action
	}
	out := action
	out.Steps = append(slices.Clone(action.Steps), webEnvironmentSteps(offers)...)
	argv := action.Argv
	out.Argv = func(a map[string][]string) []string {
		return append(argv(a), webEnvironmentArgs(a, offers)...)
	}
	return out
}
