package guide

// Step is one question: a choice from options, or free text when Options is nil.
type Step struct {
	Key     string
	Prompt  string
	Options []Option
	// Placeholder is shown in an empty text field.
	Placeholder string
	// Multi lets the user toggle several options (space) before confirming.
	Multi bool
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
	// Presents is the id of the action (see package actions) the form runs.
	Presents string
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
	for _, l := range languages {
		langs = append(langs, Option{Label: l, Value: l})
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
			Steps: []Step{{Key: "name", Prompt: "Workspace name (kebab-case)", Placeholder: "acme"}},
			Argv:  func(a map[string][]string) []string { return []string{"new", first(a, "name")} },
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
				{Key: "kind", Prompt: "What kind?", Options: []Option{
					{Label: "app", Value: "app", Hint: "deployed, never published"},
					{Label: "lib", Value: "lib", Hint: "internal, never published"},
					{Label: "package", Value: "package", Hint: "published on release"},
				}},
				{Key: "lang", Prompt: "Which language?", Options: langs},
				{Key: "name", Prompt: "Name (kebab-case)", Placeholder: "web"},
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
			Title: "Adopt this repository", Presents: "adopt", Hint: "smith adopt: bring an existing repository under smith", OutsideOnly: true,
			Steps: []Step{{Key: "step", Prompt: "What should smith do?", Options: []Option{
				{Label: "Report what adopting involves", Value: "report", Hint: "read-only: the projects smith would see, the blockers and the step that clears each"},
				{Label: "Create the baseline release tags", Value: "tags", Hint: "local only, nothing is pushed; needs a clean git tree"},
				{Label: "Write smith.toml and the pipelines, and commit", Value: "init", Hint: "needs a clean git tree"},
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
			Title: "Convert an Nx or Turborepo workspace", Presents: "convert", Hint: "smith convert: turn it into a smith workspace", OutsideOnly: true,
			Steps: []Step{{Key: "mode", Prompt: "Preview first?", Options: []Option{
				{Label: "Preview the conversion", Value: "preview", Hint: "smith convert --dry-run: reports what maps and what needs you, changes nothing"},
				{Label: "Convert now", Value: "convert", Hint: "the old tool is removed and the pipelines are regenerated"},
			}}},
			Argv: func(a map[string][]string) []string {
				if first(a, "mode") == "convert" {
					return []string{"convert"}
				}
				return []string{"convert", "--dry-run"}
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
