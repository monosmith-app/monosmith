package guide

import (
	"github.com/monosmith-app/monosmith/libs/actions/flagvocabulary"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pipelinetemplate"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite/webenv"
)

// The titles of the forms that change the environments of a web app.
var (
	// EnvironmentsAddTitle is the title of the form that adds environments to a web app.
	EnvironmentsAddTitle = words.Say("guide.parity_forms.add_title")
	// EnvironmentsRemoveTitle is the title of the form that removes environments from a web app.
	EnvironmentsRemoveTitle = words.Say("guide.parity_forms.remove_title")
)

// environmentsForm is the form of `smith environments add` or `remove`: the
// web app, then the names. The plan of the command (the files it writes or
// leaves) is shown by the confirmation before anything changes.
func environmentsForm(verb string) Action {
	title, names := EnvironmentsAddTitle, Step{
		Key: "names", Prompt: words.Say("guide.parity_forms.names_to_add_separated_by_commas"), Placeholder: words.Say("guide.parity_forms.qa_stage"),
		Note: words.Say("guide.parity_forms.each_name_gets_a_env_name"),
	}
	if verb == "remove" {
		title, names = EnvironmentsRemoveTitle, Step{
			Key: "names", Prompt: words.Say("guide.parity_forms.names_to_remove_separated_by_commas"), Placeholder: "dev",
			Note: words.Say("guide.parity_forms.the_env_name_file_stays_because"),
		}
	}
	names.Check = func(_ map[string][]string, value string) string {
		if err := webenv.ValidateNames(webenv.ParseList(value)); err != nil {
			return err.Error()
		}
		return ""
	}
	return Action{
		Title: title, Presents: "environments " + verb, Hint: words.Say("guide.parity_forms.hint", "verb", verb), NeedsWorkspace: true, // words:literal
		Steps: []Step{{Key: "project", Prompt: words.Say("guide.parity_forms.which_web_app"), Placeholder: "web"}, names},
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
	pick := Step{Key: key, Prompt: words.Say("guide.parity_forms.which_project"), Filter: len(names) > 10}
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
			hint += words.Say("guide.parity_forms.asks_for_the_app")
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
			Title: words.Say("guide.parity_forms.add_a_ready_made_pipeline"), Presents: "pipeline", Hint: words.Say("guide.parity_forms.smith_pipeline_a_pipeline_file_beside"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "template", Prompt: words.Say("guide.parity_forms.which_pipeline"), Options: templates},
				{Key: "project", Prompt: words.Say("guide.parity_forms.which_app_does_it_build"), Placeholder: "web", Skip: needsNoApp},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"pipeline", first(a, "template")}
				if p := first(a, "project"); p != "" {
					argv = append(argv, flagvocabulary.Arg("project"), p)
				}
				return argv
			},
		},
		{
			Title: words.Say("guide.parity_forms.make_the_icons_of_a_tauri"), Presents: words.Say("guide.parity_forms.tauri_icon"), Hint: words.Say("guide.parity_forms.smith_tauri_icon_every_icon_size"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "logo", Prompt: words.Say("guide.parity_forms.which_square_png_a_path_from"), Placeholder: "logo.png", Path: true},
				{Key: "project", Prompt: words.Say("guide.parity_forms.which_app_only_when_more_than"), Optional: true, Placeholder: "desktop"},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"tauri", "icon", first(a, "logo")}
				if p := first(a, "project"); p != "" {
					argv = append(argv, flagvocabulary.Arg("project"), p)
				}
				return argv
			},
		},
		{
			Title: words.Say("guide.parity_forms.set_the_docker_hub_descriptions"), Presents: words.Say("guide.parity_forms.docker_description"), Hint: words.Say("guide.parity_forms.smith_docker_description_the_description"), NeedsWorkspace: true,
			Steps: []Step{{Key: "mode", Prompt: words.Say("guide.parity_forms.preview_first"), Options: []Option{
				{Label: words.Say("guide.parity_forms.name_the_repositories_only"), Value: "preview", Hint: words.Say("guide.parity_forms.default_nothing_is_sent_to_docker")},
				{Label: words.Say("guide.parity_forms.set_the_descriptions"), Value: "set", Hint: words.Say("guide.parity_forms.calls_docker_hub_with_your_login")},
			}}},
			Argv: func(a map[string][]string) []string {
				if first(a, "mode") == "set" {
					return []string{"docker", "description"}
				}
				return []string{"docker", "description", "--dry-run"}
			},
		},
		{
			Title: words.Say("guide.parity_forms.why_would_a_task_run_again"), Presents: "why-miss", Hint: words.Say("guide.parity_forms.smith_why_miss_why_a_task"), NeedsWorkspace: true,
			Steps: []Step{{Key: "task", Prompt: words.Say("guide.parity_forms.which_task_project_task"), Placeholder: "web:build"}},
			Argv:  func(a map[string][]string) []string { return []string{"why-miss", first(a, "task")} },
		},
		{
			Title: words.Say("guide.parity_forms.connect_smith_to_a_coding_assistant"), Presents: words.Say("guide.parity_forms.mcp_install"), Hint: words.Say("guide.parity_forms.smith_mcp_install_writes_the_file"), NeedsWorkspace: true,
			Steps: []Step{{Key: "tools", Prompt: words.Say("guide.parity_forms.which_assistants_space_toggles_enter_con"), Multi: true, Options: []Option{
				{Label: words.Say("guide.parity_forms.claude_code"), Value: "--claude", Hint: words.Say("guide.parity_forms.writes_mcp_json")},
				{Label: words.Say("guide.parity_forms.vs_code"), Value: "--vscode", Hint: words.Say("guide.parity_forms.writes_vscode_mcp_json")},
			}}},
			Argv: func(a map[string][]string) []string { return append([]string{"mcp", "install"}, a["tools"]...) },
		},
	}
}
