package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
	"github.com/monosmith-app/monosmith/libs/words/words"
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
// EnvUseTheseHint says what taking the default four means, in one sentence.
func EnvUseTheseHint() string { return words.Say("guide.web_environments.use_these_hint") }

// EnvChangeNames is the label of the answer that types other names.
func EnvChangeNames() string { return words.Say("guide.web_environments.change_names") }

// EnvChangeDetails is the label of the answer that opens the detailed questions.
func EnvChangeDetails() string { return words.Say("guide.web_environments.change_details") }

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
		{Key: envsKey, Prompt: words.Say("guide.web_environments_steps.a_web_app_is_built_once"),
			Take: func(a map[string][]string) ([]string, bool) {
				return []string{envsDefault}, first(a, addSetupKey) == setupDefaults
			},
			Note: words.Say("guide.web_environments_steps.vite_writes_its_vite_values_into"), Skip: notWeb, Options: []Option{
				{Label: words.Say("guide.web_environments.use_these", "defaults", defaults), Value: envsDefault, Hint: EnvUseTheseHint()},
				{Label: EnvChangeNames(), Value: envsCustom, Hint: words.Say("guide.web_environments_steps.you_type_the_names_lowercase_words")},
				{Label: EnvChangeDetails(), Value: envsDetails, Hint: words.Say("guide.web_environments_steps.which_environment_smith_run_build_builds")},
				{Label: words.Say("guide.web_environments_steps.one_build_only"), Value: envsNone, Hint: words.Say("guide.web_environments_steps.smith_add_environments_none_a_single")},
			}},
		{Key: envNamesKey, Prompt: words.Say("guide.web_environments_steps.environment_names_separated_by_commas"), Placeholder: defaults,
			Initial: func(map[string][]string) string { return defaults },
			Skip:    unless(envsCustom),
			Check: func(_ map[string][]string, value string) string {
				if err := webenv.ValidateNames(webenv.ParseList(value)); err != nil {
					return err.Error()
				}
				return ""
			}},
		{Key: envDefaultKey, Prompt: words.Say("guide.web_environments_steps.which_environment_is_smith_run_build"), Skip: unless(envsCustom, envsDetails), OptionsFor: envDefaultOptions},
		{Key: envOutputKey, Prompt: words.Say("guide.web_environments_steps.where_do_the_builds_go"), Skip: unless(envsDetails), Options: []Option{
			{Label: words.Say("guide.web_environments_steps.in_the_workspace"), Value: webenv.OutputWorkspace, Hint: words.Say("guide.web_environments_steps.default_dist_apps_app_name_with")},
			{Label: words.Say("guide.web_environments_steps.in_the_project"), Value: webenv.OutputProject, Hint: words.Say("guide.web_environments_steps.dist_name_in_the_app_s")},
		}},
		{Key: envFilesKey, Prompt: words.Say("guide.web_environments_steps.the_env_files_hold_vite_values"), Skip: unless(envsDetails), Options: []Option{
			{Label: words.Say("guide.web_environments_steps.yes_keep_the_env_files_in"), Value: webenv.FilesCommit, Hint: words.Say("guide.web_environments_steps.default_ci_builds_with_them")},
			{Label: words.Say("guide.web_environments_steps.only_env_example"), Value: webenv.FilesExample, Hint: words.Say("guide.web_environments_steps.each_machine_and_ci_job_makes")},
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
			out = append(out, Option{Label: n, Value: n, Hint: words.Say("guide.web_environments_steps.the_others_are_build_name")})
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
