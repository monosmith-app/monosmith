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
			Title: "Create a workspace", Hint: "smith new", OutsideOnly: true,
			Steps: []Step{{Key: "name", Prompt: "Workspace name (kebab-case)", Placeholder: "acme"}},
			Argv:  func(a map[string][]string) []string { return []string{"new", first(a, "name")} },
		},
		{
			Title: "Run tasks", Hint: "build, test, lint, typecheck", NeedsWorkspace: true,
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
			Title: "Add a project", Hint: "app, lib or package", NeedsWorkspace: true,
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
			Title: "Manage plugins", Hint: "choose which built-in plugins are on", NeedsWorkspace: true,
			Steps: []Step{{Key: "plugins", Prompt: "Enabled plugins (space toggles, enter saves)", Multi: true, Options: pluginOpts}},
			Argv: func(a map[string][]string) []string {
				return append([]string{"plugin", "set"}, a["plugins"]...)
			},
		},
		{Title: "Show the project graph", Hint: "smith graph", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"graph"} }},
		{Title: "Check the workspace", Hint: "smith doctor", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"doctor"} }},
		{Title: "Preview the next release", Hint: "smith release --dry-run", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"release", "--dry-run"} }},
		{Title: "Regenerate CI pipelines", Hint: "smith ci generate", NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"ci", "generate"} }},
	}
}
