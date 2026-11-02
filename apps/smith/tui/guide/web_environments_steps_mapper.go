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
	envsDetails = "details"
	envsNone    = "none"
)

// The words of the one question about the environments of a web app. The
// VS Code extension uses the same labels (docs/manual/tui.md).
const (
	// EnvUseThese is the label of the answer that takes the default four.
	EnvUseThese = "Use these"
	// EnvUseTheseHint says what taking them means, in one sentence.
	EnvUseTheseHint = "smith run build builds prd, the builds go to dist/apps/<name>/<env>, and the .env files are written for you and kept in git (smith commits nothing)"
	// EnvChangeNames is the label of the answer that types other names.
	EnvChangeNames = "Change the names…"
	// EnvChangeDetails is the label of the answer that opens the detailed questions.
	EnvChangeDetails = "Change the build default, the output or the .env files…"
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
// app. One question comes first: use the default four (the common case, which
// ends the questions), change the names, change the details, or build once.
// The names, the environment that `smith run build` builds, where the builds
// go and whether the .env files are committed are asked only on the answers
// that change them. They are asked only for a web app.
func webEnvironmentSteps(offers []ChoiceOffer) []Step {
	notWeb := func(a map[string][]string) bool { return !isWebAnswer(a, offers) }
	unless := func(answers ...string) func(map[string][]string) bool {
		return func(a map[string][]string) bool { return notWeb(a) || !slices.Contains(answers, first(a, envsKey)) }
	}
	defaults := strings.Join(webenv.DefaultNames, ", ")
	return []Step{
		{Key: envsKey, Prompt: "A web app is built once per environment. Which environments?",
			Take: func(a map[string][]string) ([]string, bool) {
				return []string{envsDefault}, first(a, addSetupKey) == setupDefaults
			},
			Note: "Vite writes its VITE_ values into the bundle when it builds, so each environment is its own build with its own .env file.", Skip: notWeb, Options: []Option{
				{Label: EnvUseThese + ": " + defaults, Value: envsDefault, Hint: EnvUseTheseHint},
				{Label: EnvChangeNames, Value: envsCustom, Hint: "you type the names (lowercase words, such as qa, stage), then say which one smith run build builds"},
				{Label: EnvChangeDetails, Value: envsDetails, Hint: "which environment smith run build builds, where the builds go, and whether the .env files are kept in git"},
				{Label: "One build only", Value: envsNone, Hint: "smith add --environments none: a single vite build, as before"},
			}},
		{Key: envNamesKey, Prompt: "Environment names, separated by commas", Placeholder: defaults,
			Initial: func(map[string][]string) string { return defaults },
			Skip:    unless(envsCustom),
			Check: func(_ map[string][]string, value string) string {
				if err := webenv.ValidateNames(webenv.ParseList(value)); err != nil {
					return err.Error()
				}
				return ""
			}},
		{Key: envDefaultKey, Prompt: "Which environment is `smith run build`? CI and the release build this one", Skip: unless(envsCustom, envsDetails), OptionsFor: envDefaultOptions},
		{Key: envOutputKey, Prompt: "Where do the builds go?", Skip: unless(envsDetails), Options: []Option{
			{Label: "In the workspace", Value: webenv.OutputWorkspace, Hint: "default: dist/apps/<app>/<name>, with every other build"},
			{Label: "In the project", Value: webenv.OutputProject, Hint: "dist/<name> in the app's own folder; smith doctor then accepts it"},
		}},
		{Key: envFilesKey, Prompt: "The .env files hold VITE_ values, which end up in the browser, so no secret. Keep them in git?", Skip: unless(envsDetails), Options: []Option{
			{Label: "Yes, keep the .env files in git", Value: webenv.FilesCommit, Hint: "default: CI builds with them"},
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

// orElse is the answer, or the default when the question was not asked.
func orElse(answer, fallback string) string {
	if answer == "" {
		return fallback
	}
	return answer
}

// webEnvironmentArgs are the flags of smith add that carry the answers of the
// environment questions; nothing for an app that is not a web app. A question
// that was not asked has its default, which the plan shows.
func webEnvironmentArgs(a map[string][]string, offers []ChoiceOffer) []string {
	if !isWebAnswer(a, offers) {
		return nil
	}
	if first(a, envsKey) == envsNone {
		return []string{"--environments", "none"}
	}
	names := envNamesOf(a)
	return []string{
		"--environments", strings.Join(names, ","),
		"--default-environment", orElse(first(a, envDefaultKey), webenv.Settings{Names: names}.Normalized().Default),
		"--environment-output", orElse(first(a, envOutputKey), webenv.OutputWorkspace),
		"--env-files", orElse(first(a, envFilesKey), webenv.FilesCommit),
	}
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
