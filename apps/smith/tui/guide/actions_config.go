package guide

import "github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"

// Step is one question: a choice from options, or free text when Options is nil.
type Step struct {
	Key     string
	Prompt  string
	Options []Option
	// Placeholder is shown in an empty text field.
	Placeholder string
	// Multi lets the user toggle several options (space) before confirming.
	Multi bool
	// Optional lets a free-text step be left empty; the answer is then empty.
	Optional bool
	// Filter lets typing narrow the options of a long list.
	Filter bool
	// PlaceholderFor, when set, picks the placeholder from the earlier answers.
	PlaceholderFor func(answers map[string][]string) string
	// Check says what is wrong with a typed answer ("" when it is fine); it
	// runs as the person types and again on enter.
	Check func(answers map[string][]string, value string) string
	// Skip, when set, leaves the question out when it says so, given the
	// earlier answers; the key then has no answer.
	Skip func(answers map[string][]string) bool
}

// Option is one answer.
type Option struct {
	Label, Value, Hint string
}

// Action is a menu entry: its questions, then the command they build.
type Action struct {
	Title, Hint string
	Steps       []Step
	// Argv turns the answers into smith's arguments.
	Argv func(answers map[string][]string) []string
	// NeedsWorkspace hides the action outside a workspace.
	NeedsWorkspace bool
	// OutsideOnly hides the action inside a workspace.
	OutsideOnly bool
	// NeedsContent hides the action in a folder that holds nothing.
	NeedsContent bool
	// Presents is the id of the action (see package actions) the form runs.
	Presents string
	// Wizard, when set, hands the action to the shell's wizard of that name
	// instead of asking its Steps: the plain, screen-reader menu still asks
	// them.
	Wizard string
}

// WizardAdopt is the one wizard that brings a folder in, whether it is a plain
// repository or an Nx or Turborepo workspace.
const WizardAdopt = "adopt"

// AdoptionFirst moves the entries the adopt wizard answers to the top, for a
// folder that holds something and is not a workspace yet: the first entry is
// the one the menu starts on, so Enter begins the adoption.
func AdoptionFirst(actions []Action) []Action {
	var first, rest []Action
	for _, a := range actions {
		if a.Wizard == WizardAdopt && len(first) == 0 {
			first = append(first, a)
		} else {
			rest = append(rest, a)
		}
	}
	return append(first, rest...)
}

func first(a map[string][]string, key string) string {
	if v := a[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// Actions is the menu, given the plugins that can generate and every built-in
// plugin.
func Actions(languages, plugins, enabled []string) []Action {
	var langs, pluginOpts []Option
	for _, l := range languagenames.Order(languages) {
		langs = append(langs, Option{Label: languagenames.Of(l), Value: l, Hint: languagenames.Describe(l)})
	}
	isEnabled := map[string]bool{}
	for _, p := range enabled {
		isEnabled[p] = true
	}
	for _, p := range plugins {
		hint := "disabled"
		if isEnabled[p] {
			hint = "enabled"
		}
		pluginOpts = append(pluginOpts, Option{Label: p, Value: p, Hint: hint})
	}
	return []Action{
		{
			Title: "Create a workspace", Presents: "new", Hint: "smith new", OutsideOnly: true,
			Steps: NewWorkspaceSteps(),
			Argv:  NewWorkspaceArgv,
		},
		{
			Title: "Run tasks", Presents: "run", Hint: "build, test, lint, typecheck", NeedsWorkspace: true,
			Steps: []Step{
				{Key: "tasks", Prompt: "Which tasks? (space toggles, enter confirms)", Multi: true, Options: []Option{
					{Label: "lint", Value: "lint"}, {Label: "typecheck", Value: "typecheck"}, {Label: "test", Value: "test"}, {Label: "build", Value: "build"},
				}},
				{Key: "scope", Prompt: "On which projects?", Options: []Option{
					{Label: "every project", Value: "all"}, {Label: "only what changed (affected)", Value: "affected"},
				}},
			},
			Argv: func(a map[string][]string) []string {
				argv := append([]string{"run"}, a["tasks"]...)
				if first(a, "scope") == "affected" {
					argv = append(argv, "--affected")
				}
				return argv
			},
		},
		{
			Title: "Add a project", Presents: "add", Hint: "app, lib or package", NeedsWorkspace: true,
			Steps: []Step{
				{Key: "kind", Prompt: "What are you adding?", Options: []Option{
					{Label: "An app", Value: "app", Hint: "something you run or deploy, like a website or an API (web, api)"},
					{Label: "A library", Value: "lib", Hint: "code other projects here share; it stays in this workspace"},
					{Label: "A package", Value: "package", Hint: "code you publish for others to install (npm, PyPI, crates.io and so on)"},
				}},
				{Key: "lang", Prompt: "Which language is it written in?", Options: langs, Filter: true},
				{Key: "name", Prompt: "Name it (lowercase words joined by dashes)", Placeholder: "web", PlaceholderFor: namePlaceholder, Check: checkProjectName},
			},
			Argv: func(a map[string][]string) []string {
				return []string{"add", first(a, "kind"), first(a, "name"), "--lang", first(a, "lang")}
			},
		},
		{
			Title: "Manage plugins", Presents: "plugin set", Hint: "choose which built-in plugins are on", NeedsWorkspace: true,
			Steps: []Step{{Key: "plugins", Prompt: "Enabled plugins (space toggles, enter saves)", Multi: true, Options: pluginOpts}},
			Argv: func(a map[string][]string) []string {
				return append([]string{"plugin", "set"}, a["plugins"]...)
			},
		},
		{Title: "Show the project graph", Presents: "graph", Hint: "smith graph", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"graph"} }},
		{Title: "Check the workspace", Presents: "doctor", Hint: "smith doctor", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"doctor"} }},
		{Title: "Preview the next release", Presents: "release", Hint: "smith release --dry-run", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"release", "--dry-run"} }},
		{Title: "Regenerate CI pipelines", Presents: "ci", Hint: "smith ci generate", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"ci", "generate"} }},
		{
			Title: "Bring this repository into MonoSmith", Wizard: WizardAdopt, Presents: "adopt", Hint: "smith reads what is here (a plain repository, or one that Nx or Turborepo manages today), asks how far to go and shows every file it would change before anything is written", OutsideOnly: true, NeedsContent: true,
			Steps: []Step{{Key: "step", Prompt: "What should smith do?", Options: []Option{
				{Label: "Report what bringing it in involves", Value: "report", Hint: "read-only: the projects smith would see, what blocks and the step that clears each"},
				{Label: "Create the baseline release tags", Value: "tags", Hint: "local only, nothing is pushed; needs a clean git tree"},
				{Label: "Write smith.toml and the pipelines, and commit", Value: "init", Hint: "needs a clean git tree; an Nx or Turborepo workspace is converted"},
				{Label: "Tags, then smith.toml and the pipelines", Value: "both", Hint: "both steps; needs a clean git tree"},
			}}},
			Argv: func(a map[string][]string) []string {
				switch first(a, "step") {
				case "tags":
					return []string{"adopt", "--tags"}
				case "init":
					return []string{"adopt", "--init"}
				case "both":
					return []string{"adopt", "--tags", "--init"}
				}
				return []string{"adopt"}
			},
		},
		{
			Title: "Remove a project", Presents: "remove", Hint: "smith remove: its directory, its traces in the workspace, its CI", NeedsWorkspace: true,
			Steps: []Step{
				{Key: "name", Prompt: "Which project?", Placeholder: "web"},
				{Key: "mode", Prompt: "Preview first?", Options: []Option{
					{Label: "Preview what would change", Value: "preview", Hint: "smith remove --dry-run: changes nothing"},
					{Label: "Remove it", Value: "remove", Hint: "refused while other projects import it"},
				}},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"remove", first(a, "name")}
				if first(a, "mode") != "remove" {
					argv = append(argv, "--dry-run")
				}
				return argv
			},
		},
	}
}
