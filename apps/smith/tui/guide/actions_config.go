package guide

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/languagenames"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// WizardAdopt is the one wizard that brings a folder in, whether it is a plain
// repository or an Nx or Turborepo workspace.
const WizardAdopt = "adopt"

// WizardFixConfig is the review of the rename that repairs a smith.toml that
// uses a renamed key.
const WizardFixConfig = "fix-config"

// FixConfigFirst puts "Fix smith.toml" before the other entries, so the menu
// starts on it. The entry shows the change first and writes nothing until the
// user confirms.
func FixConfigFirst(actions []Action) []Action {
	fix := Action{
		Title: words.Say("guide.actions.fix_smith_toml"), Wizard: WizardFixConfig, Presents: "doctor",
		Hint: words.Say("guide.actions.smith_doctor_fix_renames_the_keys"),
	}
	return append([]Action{fix}, actions...)
}

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
	return append([]Action{
		{
			Title: words.Say("guide.actions.create_a_workspace"), Presents: "new", Hint: words.Say("guide.actions.smith_new"), OutsideOnly: true,
			Steps: NewWorkspaceSteps(),
			Argv:  NewWorkspaceArgv,
		},
		{
			Title: words.Say("guide.actions.run_tasks"), Presents: "run", Hint: words.Say("guide.actions.build_test_lint_typecheck"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "tasks", Prompt: words.Say("guide.actions.which_tasks_space_toggles_enter_confirms"), Multi: true, Options: []Option{
					{Label: "lint", Value: "lint"}, {Label: "typecheck", Value: "typecheck"}, {Label: "test", Value: "test"}, {Label: "build", Value: "build"},
				}},
				{Key: "scope", Prompt: words.Say("guide.actions.on_which_projects"), Options: []Option{
					{Label: words.Say("guide.actions.every_project"), Value: "all"}, {Label: words.Say("guide.actions.only_what_changed_affected"), Value: "affected"},
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
			Title: words.Say("guide.actions.add_a_project"), Presents: "add", Hint: words.Say("guide.actions.app_lib_or_package_to_search"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "kind", Prompt: words.Say("guide.actions.what_are_you_adding"), Options: []Option{
					{Label: words.Say("guide.actions.an_app"), Value: "app", Hint: words.Say("guide.actions.something_you_run_or_deploy_like")},
					{Label: words.Say("guide.actions.a_library"), Value: "lib", Hint: words.Say("guide.actions.code_other_projects_here_share_it")},
					{Label: words.Say("guide.actions.a_package"), Value: "package", Hint: words.Say("guide.actions.code_you_publish_for_others_to")},
				}},
				{Key: "lang", Prompt: words.Say("guide.actions.which_language_is_it_written_in"), Options: langs, Filter: true},
				{Key: "name", Prompt: words.Say("guide.actions.name_it_lowercase_words_joined_by"), Placeholder: "web", PlaceholderFor: namePlaceholder, Check: checkProjectName, Suggest: suggestName("")},
			},
			Argv: func(a map[string][]string) []string {
				return []string{"add", first(a, "kind"), first(a, "name"), "--lang", first(a, "lang")}
			},
		},
		{
			Title: words.Say("guide.actions.manage_plugins"), Presents: words.Say("guide.actions.plugin_set"), Hint: words.Say("guide.actions.choose_which_built_in_plugins_are"), NeedsWorkspace: true,
			Steps: []Step{{Key: "plugins", Prompt: words.Say("guide.actions.enabled_plugins_space_toggles_enter_save"), Note: WhatIs("plugin"), Multi: true, Options: pluginOpts}},
			Argv: func(a map[string][]string) []string {
				return append([]string{"plugin", "set"}, a["plugins"]...)
			},
		},
		{Title: words.Say("guide.actions.show_the_project_graph"), Presents: "graph", Hint: words.Say("guide.actions.smith_graph"), NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"graph"} }},
		{Title: words.Say("guide.actions.check_the_workspace"), Presents: "doctor", Hint: words.Say("guide.actions.smith_doctor"), NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"doctor"} }},
		{Title: words.Say("guide.actions.preview_the_next_release"), Presents: "release", Hint: words.Say("guide.actions.smith_release_dry_run"), NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"release", "--dry-run"} }},
		{Title: words.Say("guide.actions.regenerate_ci_pipelines"), Presents: "ci", Hint: words.Say("guide.actions.smith_ci_generate"), NeedsWorkspace: true, Argv: func(map[string][]string) []string { return []string{"ci", "generate"} }},
		{
			Title: words.Say("guide.actions.bring_this_repository_into_monosmith"), Wizard: WizardAdopt, Presents: "adopt", Hint: words.Say("guide.actions.smith_reads_what_is_here_a"), OutsideOnly: true, NeedsContent: true,
			Steps: []Step{{Key: "step", Prompt: words.Say("guide.actions.what_should_smith_do"), Options: []Option{
				{Label: words.Say("guide.actions.report_what_bringing_it_in_involves"), Value: "report", Hint: words.Say("guide.actions.read_only_the_projects_smith_would")},
				{Label: words.Say("guide.actions.create_the_baseline_release_tags"), Value: "tags", Hint: words.Say("guide.actions.local_only_nothing_is_pushed_needs")},
				{Label: words.Say("guide.actions.write_smith_toml_and_the_pipelines"), Value: "init", Hint: words.Say("guide.actions.needs_a_clean_git_tree_an")},
				{Label: words.Say("guide.actions.tags_then_smith_toml_and_the"), Value: "both", Hint: words.Say("guide.actions.both_steps_needs_a_clean_git")},
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
			Title: words.Say("guide.actions.convert_an_nx_or_turborepo_workspace"), Presents: "convert", Hint: words.Say("guide.actions.smith_convert_projects_keep_their_folder"), OutsideOnly: true, NeedsContent: true,
			Steps: []Step{
				{Key: "level", Prompt: words.Say("guide.actions.how_far_should_the_conversion_go"), Options: []Option{
					{Label: words.Say("guide.actions.conservative_keep_what_works_list_the"), Value: "conservative", Hint: words.Say("guide.actions.the_old_tool_is_removed_eslint")},
					{Label: words.Say("guide.actions.full_smith_becomes_the_only_tool"), Value: "full", Hint: words.Say("guide.actions.also_deletes_the_old_tooling_nothing")},
				}},
				{Key: "lint", Prompt: words.Say("guide.actions.lint_with"), Options: []Option{
					{Label: words.Say("guide.actions.keep_the_current_setup"), Value: "keep", Hint: words.Say("guide.actions.your_eslint_config_stays_default")},
					{Label: words.Say("guide.actions.smith_s_eslint_config"), Value: "smith", Hint: words.Say("guide.actions.your_own_eslint_configs_and_plugins")},
				}, Skip: func(a map[string][]string) bool { return first(a, "level") != "full" }},
				{Key: "test", Prompt: words.Say("guide.actions.test_runner"), Options: []Option{
					{Label: words.Say("guide.actions.keep_the_current_setup"), Value: "keep", Hint: "(default)"},
					{Label: words.Say("guide.actions.jest"), Value: "jest", Hint: words.Say("guide.actions.the_vitest_configs_and_packages_go")},
					{Label: words.Say("guide.actions.vitest"), Value: "vitest", Hint: words.Say("guide.actions.the_jest_configs_and_packages_go")},
				}, Skip: func(a map[string][]string) bool { return first(a, "level") != "full" }},
				{Key: "build", Prompt: words.Say("guide.actions.build_with"), Options: []Option{
					{Label: words.Say("guide.actions.smith_s_stack_build"), Value: "smith", Hint: words.Say("guide.actions.rollup_or_webpack_configs_nothing_uses")},
					{Label: words.Say("guide.actions.keep_the_current_build_setup"), Value: "keep", Hint: words.Say("guide.actions.nothing_about_the_build_changes")},
				}, Skip: func(a map[string][]string) bool { return first(a, "level") != "full" }},
				{Key: "ci", Prompt: words.Say("guide.actions.pipelines"), Options: []Option{
					{Label: words.Say("guide.actions.the_pipelines_smith_generates"), Value: "smith", Hint: words.Say("guide.actions.the_current_ci_files_are_replaced")},
					{Label: words.Say("guide.actions.keep_the_current_pipelines"), Value: "keep", Hint: words.Say("guide.actions.smith_writes_none")},
				}, Skip: func(a map[string][]string) bool { return first(a, "level") != "full" }},
				{Key: "commit", Prompt: words.Say("guide.actions.commit_the_conversion_so_it_can"), Options: []Option{
					{Label: words.Say("guide.actions.no_leave_the_files_uncommitted"), Value: "no", Hint: words.Say("guide.actions.review_the_changes_and_commit_when")},
					{Label: words.Say("guide.actions.yes_one_commit_that_smith_undo"), Value: "yes", Hint: words.Say("guide.actions.the_conversion_is_one_commit")},
				}},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"convert"}
				if first(a, "level") == "full" {
					argv = append(argv, "--level", "full", "--lint", first(a, "lint"), "--test", first(a, "test"), "--build", first(a, "build"), "--ci", first(a, "ci"))
				}
				if first(a, "commit") == "yes" {
					argv = append(argv, "--commit")
				}
				return argv
			},
		},
		ForcedReleaseAction(nil, nil),
		{
			Title: words.Say("guide.actions.remove_a_project"), Presents: "remove", Hint: words.Say("guide.actions.smith_remove_its_directory_its_traces"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "name", Prompt: words.Say("guide.actions.which_project"), Placeholder: "web"},
				{Key: "mode", Prompt: words.Say("guide.actions.preview_first"), Options: []Option{
					{Label: words.Say("guide.actions.preview_what_would_change"), Value: "preview", Hint: words.Say("guide.actions.smith_remove_dry_run_changes_nothing")},
					{Label: words.Say("guide.actions.remove_it"), Value: "remove", Hint: words.Say("guide.actions.refused_while_other_projects_import_it")},
				}},
				{Key: "folder", Prompt: words.Say("guide.actions.what_happens_to_the_project_s"), Options: []Option{
					{Label: words.Say("guide.actions.keep_a_copy"), Value: "keep", Hint: words.Say("guide.actions.the_folder_moves_to_smith_removed")},
					{Label: words.Say("guide.actions.delete_it_for_good"), Value: "delete", Hint: words.Say("guide.actions.nothing_is_kept_git_still_has")},
				}, Skip: func(a map[string][]string) bool { return first(a, "mode") != "remove" }},
			},
			Argv: func(a map[string][]string) []string {
				argv := []string{"remove", first(a, "name")}
				if first(a, "mode") != "remove" {
					return append(argv, "--dry-run")
				}
				if first(a, "folder") == "delete" {
					return append(argv, "--delete")
				}
				return append(argv, "--keep-copy")
			},
		},
		{
			Title: words.Say("guide.actions.why_is_this_here"), Presents: "why", Hint: words.Say("guide.actions.smith_why_who_owns_a_file"), NeedsWorkspace: true,
			Steps: []Step{{Key: "path", Prompt: words.Say("guide.actions.which_file_folder_or_setting_a"), Placeholder: "packages/web/tsconfig.json", Path: true, PathSettings: true}},
			Argv: func(a map[string][]string) []string {
				return []string{"why", first(a, "path")}
			},
		},
		{
			Title: words.Say("guide.actions.try_a_command_in_a_copy"), Presents: "try", Hint: words.Say("guide.actions.smith_try_runs_a_command_that"), NeedsWorkspace: true,
			Steps: []Step{{Key: "command", Prompt: words.Say("guide.actions.which_smith_command_add_remove_up"), Placeholder: words.Say("guide.actions.add_package_billing")}},
			Argv: func(a map[string][]string) []string {
				return append([]string{"try"}, strings.Fields(first(a, "command"))...)
			},
		},
		{
			Title: words.Say("guide.actions.swap_a_choice"), Presents: "swap", Hint: words.Say("guide.actions.smith_swap_change_a_test_runner"), NeedsWorkspace: true,
			Steps: []Step{
				{Key: "project", Prompt: words.Say("guide.actions.which_project"), Placeholder: "web"},
				{Key: "axis", Prompt: words.Say("guide.actions.which_decision_smith_choices_lists_them"), Placeholder: "test"},
				{Key: "choice", Prompt: words.Say("guide.actions.which_answer"), Placeholder: "vitest"},
			},
			Argv: SwapArgv,
		},
		{
			Title: words.Say("guide.actions.write_a_recipe"), Presents: words.Say("guide.actions.recipe_new"), Hint: words.Say("guide.actions.smith_recipe_new_a_recipe_whose"),
			Steps: NewRecipeSteps(),
			Argv:  NewRecipeArgv,
		},
		{
			Title: PluginSearchTitle, Hint: words.Say("guide.actions.smith_plugin_search_look_in_the"), NeedsWorkspace: true,
			Steps: []Step{{Key: "term", Prompt: words.Say("guide.actions.what_do_you_look_for_a"), Optional: true, Placeholder: "terraform"}},
			Argv: func(a map[string][]string) []string {
				return append([]string{"plugin", "search"}, strings.Fields(first(a, "term"))...)
			},
		},
	}, parityForms()...)
}
