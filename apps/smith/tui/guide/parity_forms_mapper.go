package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pipelinetemplate"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// The titles of the forms that change the environments of a web app.
const (
	EnvironmentsAddTitle    = "Add environments to a web app"
	EnvironmentsRemoveTitle = "Remove environments from a web app"
)

// environmentsForm is the form of `smith environments add` or `remove`: the
// web app, then the names. The plan of the command (the files it writes or
// leaves) is shown by the confirmation before anything changes.
func environmentsForm(verb string) Action {
	title, names := EnvironmentsAddTitle, Step{
		Key: "names", Prompt: "Names to add, separated by commas (lowercase words, such as qa, stage)", Placeholder: "qa, stage",
		Note: "Each name gets a .env.<name> file, a build:<name> task and an output folder. The plan shows the files before they are written.",
	}
	if verb == "remove" {
		title, names = EnvironmentsRemoveTitle, Step{
			Key: "names", Prompt: "Names to remove, separated by commas (smith environments list shows them)", Placeholder: "dev",
			Note: "The .env.<name> file stays, because it holds your values. The default environment cannot be removed.",
		}
	}
	names.Check = func(_ map[string][]string, value string) string {
		if err := webenv.ValidateNames(webenv.ParseList(value)); err != nil {
			return err.Error()
		}
		return ""
	}
	return Action{
		Title: title, Presents: "environments " + verb, Hint: "smith environments " + verb + ": the plan comes first; the .env files you wrote stay", NeedsWorkspace: true,
		Steps: []Step{{Key: "project", Prompt: "Which web app?", Placeholder: "web"}, names},
		Argv: func(a map[string][]string) []string {
			return append([]string{"environments", verb, first(a, "project")}, webenv.ParseList(first(a, "names"))...)
		},
	}
}

// WithProjectPick puts a pick list of projects in the place of the free-text
// step with the key, so a name cannot be mistyped. A form with no project is
// returned as it is.
func WithProjectPick(base Action, key string, names []string) Action {
	if len(names) == 0 {
		return base
	}
	pick := Step{Key: key, Prompt: "Which project?", Filter: len(names) > 10}
	for _, n := range names {
		pick.Options = append(pick.Options, Option{Label: n, Value: n})
	}
	out := base
	out.Steps = slices.Clone(base.Steps)
	for i := range out.Steps {
		if out.Steps[i].Key == key {
			pick.Prompt = out.Steps[i].Prompt
			out.Steps[i] = pick
		}
	}
	return out
}

// WithWorkspaceEnvironments starts the free-text question of the names of an
// app's own environments on the names the workspace declares
// (`[workspace] environments`), so the person edits a list and does not type
// one. An empty list keeps the default four.
func WithWorkspaceEnvironments(action Action, names []string) Action {
	if len(names) == 0 {
		return action
	}
	out := action
	out.Steps = slices.Clone(action.Steps)
	for i := range out.Steps {
		if out.Steps[i].Key == envNamesKey {
			out.Steps[i].Initial = func(map[string][]string) string { return strings.Join(names, ", ") }
		}
	}
	return out
}

// parityForms are the forms of commands that have a screen of their own in
// the palette: a ready-made pipeline, the icons of a Tauri app, the Docker
// Hub descriptions, why a task missed the cache, the connection of an
// assistant, and the environments of a web app.
func parityForms() []Action {
	var templates []Option
	for _, t := range pipelinetemplate.Catalog {
		hint := t.Description
		if t.NeedsProject {
			hint += " (asks for the app)"
		}
		templates = append(templates, Option{Label: t.Name, Value: t.Name, Hint: hint})
	}
	needsNoApp := func(a map[string][]string) bool {
		t, ok := pipelinetemplate.Find(first(a, "template"))
		return !ok || !t.NeedsProject
	}
	return []Action{
		environmentsForm("add"),
		environmentsForm("remove"),
		{
			Title: "Add a ready-made pipeline", Presents: "pipeline", Hint: "smith pipeline: a pipeline file beside the generated CI; it is yours from then on and smith upgrade never rewrites it", NeedsWorkspace: true,
			Steps: []Step{
				{Key: "template", Prompt: "Which pipeline?", Options: templates},
				{Key: "project", Prompt: "Which app does it build?", Placeholder: "web", Skip: needsNoApp},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"pipeline", first(a, "template")}
				if p := first(a, "project"); p != "" {
					argv = append(argv, "--project", p)
				}
				return argv
			},
		},
		{
			Title: "Make the icons of a Tauri app", Presents: "tauri icon", Hint: "smith tauri icon: every icon size from one square PNG; the plan lists the files it overwrites", NeedsWorkspace: true,
			Steps: []Step{
				{Key: "logo", Prompt: "Which square PNG? (a path from where you started smith)", Placeholder: "logo.png"},
				{Key: "project", Prompt: "Which app? (only when more than one has a src-tauri; Enter leaves it out)", Optional: true, Placeholder: "desktop"},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"tauri", "icon", first(a, "logo")}
				if p := first(a, "project"); p != "" {
					argv = append(argv, "-p", p)
				}
				return argv
			},
		},
		{
			Title: "Set the Docker Hub descriptions", Presents: "docker description", Hint: "smith docker description: the description of each app's image on Docker Hub", NeedsWorkspace: true,
			Steps: []Step{{Key: "mode", Prompt: "Preview first?", Options: []Option{
				{Label: "Name the repositories only", Value: "preview", Hint: "default: nothing is sent to Docker Hub"},
				{Label: "Set the descriptions", Value: "set", Hint: "calls Docker Hub with your login"},
			}}},
			Argv: func(a map[string][]string) []string {
				if first(a, "mode") == "set" {
					return []string{"docker", "description"}
				}
				return []string{"docker", "description", "--dry-run"}
			},
		},
		{
			Title: "Why would a task run again?", Presents: "why-miss", Hint: "smith why-miss: why a task did not come from the cache", NeedsWorkspace: true,
			Steps: []Step{{Key: "task", Prompt: "Which task? (project:task)", Placeholder: "web:build"}},
			Argv:  func(a map[string][]string) []string { return []string{"why-miss", first(a, "task")} },
		},
		{
			Title: "Connect smith to a coding assistant", Presents: "mcp install", Hint: "smith mcp install: writes the file that lets the assistant start smith's MCP server", NeedsWorkspace: true,
			Steps: []Step{{Key: "tools", Prompt: "Which assistants? (space toggles, enter confirms)", Multi: true, Options: []Option{
				{Label: "Claude Code", Value: "--claude", Hint: "writes .mcp.json"},
				{Label: "VS Code", Value: "--vscode", Hint: "writes .vscode/mcp.json"},
			}}},
			Argv: func(a map[string][]string) []string { return append([]string{"mcp", "install"}, a["tools"]...) },
		},
	}
}
