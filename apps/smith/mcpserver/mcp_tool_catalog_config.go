package mcpserver

import (
	"fmt"
	"strings"
)

// Tool is one MCP tool: what a client lists, and how its input becomes the
// smith arguments.
type Tool struct {
	Name string
	// Action is the id of the smith action (libs/actions) the tool runs.
	Action      string
	Description string
	InputSchema map[string]any
	// ReadOnly tools change nothing, so a client may run them without asking.
	ReadOnly bool
	Args     func(input map[string]any) ([]string, error)
}

func object(props map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": props, "additionalProperties": false}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func str(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func boolean(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func list(description string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": description}
}

// word reads a string input that becomes one argument: no spaces, and not an
// option, so an input cannot smuggle a flag in.
func word(input map[string]any, key string) (string, error) {
	v, ok := input[key].(string)
	if !ok || v == "" || strings.HasPrefix(v, "-") || strings.ContainsAny(v, " \t\r\n") {
		return "", fmt.Errorf("%q must be a non-empty string with no spaces that does not start with '-'", key)
	}
	return v, nil
}

// choices reads a list of "name=answer" strings, each a recipe choice's
// answer, into repeated --choice arguments.
func choices(input map[string]any, key string) ([]string, error) {
	list, err := words(input, key, false)
	var args []string
	for _, c := range list {
		if name, answer, ok := strings.Cut(c, "="); !ok || name == "" || answer == "" {
			return nil, fmt.Errorf("%q entries must be name=answer, got %q", key, c)
		}
		args = append(args, "--choice", c)
	}
	return args, err
}

// pluginName reads a plugin's name. smith plugin check also takes a path and
// then runs that executable, which a tool must not: only names.
func pluginName(input map[string]any, key string) (string, error) {
	v, err := word(input, key)
	if err != nil {
		return "", err
	}
	for _, r := range v {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return "", fmt.Errorf("%q must be a plugin's name (lowercase letters, digits and -), not a path", key)
		}
	}
	return v, nil
}

func optional(input map[string]any, key string) (string, bool, error) {
	if _, present := input[key]; !present {
		return "", false, nil
	}
	v, err := word(input, key)
	return v, err == nil, err
}

func words(input map[string]any, key string, required bool) ([]string, error) {
	raw, present := input[key]
	if !present {
		if required {
			return nil, fmt.Errorf("%q must list at least one value", key)
		}
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok || (required && len(items) == 0) {
		return nil, fmt.Errorf("%q must be a list of strings", key)
	}
	var out []string
	for _, item := range items {
		v, err := word(map[string]any{key: item}, key)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// kindArgs reads the project and the kind smith kind takes.
func kindArgs(input map[string]any) ([]string, error) {
	project, err := word(input, "project")
	if err != nil {
		return nil, err
	}
	kind, err := word(input, "kind")
	if err != nil {
		return nil, err
	}
	if kind != "app" && kind != "lib" && kind != "package" {
		return nil, fmt.Errorf("\"kind\" must be app, lib or package, got %q", kind)
	}
	return []string{"kind", project, kind}, nil
}

func flag(input map[string]any, key string) bool { v, _ := input[key].(bool); return v }

func fixed(args ...string) func(map[string]any) ([]string, error) {
	return func(map[string]any) ([]string, error) { return args, nil }
}

var noInput = object(map[string]any{})

// Tools are the tools smith mcp serves.
var Tools = []Tool{
	{Name: "smith_projects", Action: "projects", Description: "List the workspace's projects as JSON: name, directory, kind, language, stack, dependencies, tasks.", InputSchema: noInput, ReadOnly: true, Args: fixed("projects", "--json")},
	{Name: "smith_graph", Action: "graph", Description: "The project graph as JSON (who imports whom); with focus, only those projects and everything connected to them.", ReadOnly: true,
		InputSchema: object(map[string]any{"focus": list("projects to focus on: names, globs, kind:x, lang:x")}),
		Args: func(in map[string]any) ([]string, error) {
			focus, err := words(in, "focus", false)
			args := []string{"graph", "--format", "json"}
			if len(focus) > 0 {
				args = append(args, "--focus", strings.Join(focus, ","))
			}
			return args, err
		}},
	{Name: "smith_affected", Action: "affected", Description: "The projects a change affects, as JSON, compared with a base branch (default: the release branch).", ReadOnly: true,
		InputSchema: object(map[string]any{"base": str("the branch or commit to compare with")}),
		Args: func(in map[string]any) ([]string, error) {
			base, ok, err := optional(in, "base")
			args := []string{"affected", "--json"}
			if ok {
				args = append(args, "--base", base)
			}
			return args, err
		}},
	{Name: "smith_kinds", Action: "kinds", Description: "The kinds of project smith_add_project can create, as JSON.", InputSchema: noInput, ReadOnly: true, Args: fixed("kinds", "--json")},
	{Name: "smith_choices", Action: "choices", Description: "The choices a project can be built from, by axis (framework, test, style), and what is wrong with an unusable one.", InputSchema: noInput, ReadOnly: true, Args: fixed("choices")},
	{Name: "smith_stacks",Action: "stacks", Description: "The stacks a project can be built with (the plugins of each), and which projects use them.", InputSchema: noInput, ReadOnly: true, Args: fixed("stacks")},
	{Name: "smith_commands", Action: "commands", Description: "Every smith command with its arguments and options, as JSON.", InputSchema: noInput, ReadOnly: true, Args: fixed("commands", "--json")},
	{Name: "smith_info", Action: "info", Description: "The workspace and the smith serving it, as JSON.", InputSchema: noInput, ReadOnly: true, Args: fixed("info", "--json")},
	{Name: "smith_doctor", Action: "doctor", Description: "Check the workspace's health, as JSON; every failed check names its remedy.", InputSchema: noInput, ReadOnly: true, Args: fixed("doctor", "--json")},
	{Name: "smith_release_preview", Action: "release", Description: "What the next release would version, tag and publish, as JSON; changes nothing.", InputSchema: noInput, ReadOnly: true, Args: fixed("release", "--dry-run", "--json")},
	{Name: "smith_updates", Action: "up", Description: "Declared dependencies with newer releases, grouped patch, minor and major, and the tools smith pins against their newest releases, as JSON; changes nothing.", InputSchema: noInput, ReadOnly: true, Args: fixed("up", "--json")},
	{Name: "smith_explain", Action: "explain", Description: "How a project's tasks resolved and where each field came from (the language plugin, a stack, a convention, smith.toml), as JSON; with recipe, only what the project's recipe decided: its choices, the plugin at each role and the settings.", ReadOnly: true,
		InputSchema: object(map[string]any{"target": str("a project, or project:task"), "recipe": boolean("only what the recipe decided")}, "target"),
		Args: func(in map[string]any) ([]string, error) {
			target, err := word(in, "target")
			args := []string{"explain", target, "--json"}
			if flag(in, "recipe") {
				args = append(args, "--recipe")
			}
			return args, err
		}},
	{Name: "smith_recipe_check", Action: "recipe", Description: "Resolve and validate a recipe without writing anything, as JSON: the choices asked, the plugin for each role and the merged settings with the file and line each came from, or why it is not valid. Use it to show the plan before smith_add_project or smith_recipe_new.", ReadOnly: true,
		InputSchema: object(map[string]any{
			"recipe":  str("a recipe file in the workspace, or an https URL pinned in smith.recipes.lock"),
			"choices": list("answers to its choices, each name=answer; the rest take their defaults"),
			"all":     boolean("check every combination of answers (up to 64)"),
		}, "recipe"),
		Args: func(in map[string]any) ([]string, error) {
			recipe, err := word(in, "recipe")
			if err != nil {
				return nil, err
			}
			answers, err := choices(in, "choices")
			if err != nil {
				return nil, err
			}
			args := append([]string{"recipe", "check", recipe, "--json"}, answers...)
			if flag(in, "all") {
				args = append(args, "--all")
			}
			return args, nil
		}},
	{Name: "smith_plugin_list", Action: "plugin", Description: "The built-in plugins and those installed on the plugin search path, as JSON: the levels each can be chosen for (roles), the slots it fills (provides), its requirements and conflicts, and what it implements.", InputSchema: noInput, ReadOnly: true, Args: fixed("plugin", "list", "--json")},
	{Name: "smith_plugin_check", Action: "plugin", Description: "Run the conformance suite against a plugin by name, as JSON: that it names itself, its roles match what it implements, its stacks are complete and every method it advertises answers. With generate, it also generates an app, a lib and a package in a temporary directory (this may install dependencies).",
		InputSchema: object(map[string]any{"plugin": str("the plugin's name"), "generate": boolean("also generate projects and discover them")}, "plugin"),
		Args: func(in map[string]any) ([]string, error) {
			name, err := pluginName(in, "plugin")
			args := []string{"plugin", "check", name, "--json"}
			if flag(in, "generate") {
				args = append(args, "--generate")
			}
			return args, err
		}},
	{Name: "smith_doctor_fixes", Action: "doctor", Description: "The fixes smith_doctor's findings have, as diffs, and which of them are safe (they only restore what smith or a plugin owns, like go.work or a generated pipeline); changes nothing. Apply the safe ones with smith_doctor_fix.", InputSchema: noInput, ReadOnly: true, Args: fixed("doctor", "--fix")},
	{Name: "smith_upgrade_preview", Action: "upgrade", Description: "What smith upgrade would change after a newer smith or recipe: the migration of smith.toml, the files smith owns it regenerates or deletes, and the projects a recipe brings up to date, as JSON; changes nothing.", InputSchema: noInput, ReadOnly: true, Args: fixed("upgrade", "--dry-run", "--json")},
	{Name: "smith_run", Action: "run", Description: "Run tasks (build, test, lint, typecheck, or any task a project has) in dependency order, from the cache when nothing changed.",
		InputSchema: object(map[string]any{
			"tasks":    list("the tasks to run"),
			"projects": list("the projects: names, globs, kind:x, lang:x; default every project"),
			"affected": boolean("only the projects affected by changes since the release branch"),
		}, "tasks"),
		Args: func(in map[string]any) ([]string, error) {
			tasks, err := words(in, "tasks", true)
			if err != nil {
				return nil, err
			}
			projects, err := words(in, "projects", false)
			args := append([]string{"run"}, tasks...)
			if len(projects) > 0 {
				args = append(args, "-p", strings.Join(projects, ","))
			}
			if flag(in, "affected") {
				args = append(args, "--affected")
			}
			return args, err
		}},
	{Name: "smith_add_project", Action: "add", Description: "Add a project: kind app (deployed), lib (imported internally) or package (published), or a combined kind from smith_kinds (node-app). Writes files, installs packages and regenerates the pipelines.",
		InputSchema: object(map[string]any{
			"kind":    str("app, lib, package, or a combined kind such as node-app"),
			"name":    str("the project name: lowercase words joined by hyphens"),
			"lang":    str("node, go, python, or a language a plugin adds"),
			"stack":   str("the stack to build it with (see smith_stacks); none for the language plugin alone"),
			"choices": list("answers to the workspace recipe's choices, each name=answer (language=python); check them first with smith_recipe_check"),
		}, "kind", "name"),
		Args: func(in map[string]any) ([]string, error) {
			kind, err := word(in, "kind")
			if err != nil {
				return nil, err
			}
			name, err := word(in, "name")
			if err != nil {
				return nil, err
			}
			args := []string{"add", kind, name}
			for _, key := range []string{"lang", "stack"} {
				v, ok, err := optional(in, key)
				if err != nil {
					return nil, err
				}
				if ok {
					args = append(args, "--"+key, v)
				}
			}
			answers, err := choices(in, "choices")
			return append(args, answers...), err
		}},
	{Name: "smith_remove_project", Action: "remove", Description: "Remove a project and its traces (go.work, the lockfile, the venv, smith.toml). Refused while another project imports it. With dry_run, only says what would change.",
		InputSchema: object(map[string]any{"name": str("the project"), "dry_run": boolean("report only")}, "name"),
		Args: func(in map[string]any) ([]string, error) {
			name, err := word(in, "name")
			if flag(in, "dry_run") {
				return []string{"remove", name, "--dry-run"}, err
			}
			return []string{"remove", name, "--yes"}, err
		}},
	{Name: "smith_change_kind_preview", Action: "kind", Description: "What smith_change_kind would do, as JSON: the project's new directory, every file it edits (manifest, workspace files, dependents), the references that still name the old folder and the warnings; changes nothing. A refusal (an app other projects import, a package already released) is part of the answer: only a person overrules it, with smith kind --force.", ReadOnly: true,
		InputSchema: object(map[string]any{"project": str("the project"), "kind": map[string]any{"type": "string", "enum": []string{"app", "lib", "package"}, "description": "the kind it becomes"}}, "project", "kind"),
		Args: func(in map[string]any) ([]string, error) {
			args, err := kindArgs(in)
			return append(args, "--dry-run", "--json"), err
		}},
	{Name: "smith_change_kind", Action: "kind", Description: "Make a project another kind: app (deployed), lib (imported internally) or package (published). Edits its manifest, moves its folder to libs/, packages/ or apps/, rewrites every reference to the old folder and regenerates the pipelines. Preview it first with smith_change_kind_preview. It does not overrule a refusal.",
		InputSchema: object(map[string]any{"project": str("the project"), "kind": map[string]any{"type": "string", "enum": []string{"app", "lib", "package"}, "description": "the kind it becomes"}}, "project", "kind"),
		Args: func(in map[string]any) ([]string, error) {
			args, err := kindArgs(in)
			return append(args, "--json"), err
		}},
	{Name: "smith_install", Action: "install", Description: "Add a package to one or more projects with each project's own toolchain (npm, pip, go). A dependency always belongs to a project, never the root.",
		InputSchema: object(map[string]any{"package": str("the package, optionally with a version"), "projects": list("the projects, by name or directory"), "dev": boolean("a development dependency")}, "package", "projects"),
		Args: func(in map[string]any) ([]string, error) {
			pkg, err := word(in, "package")
			if err != nil {
				return nil, err
			}
			projects, err := words(in, "projects", true)
			if err != nil {
				return nil, err
			}
			args := []string{"install"}
			for _, p := range projects {
				args = append(args, "-w", p)
			}
			if flag(in, "dev") {
				args = append(args, "-D")
			}
			return append(args, pkg), nil
		}},
	{Name: "smith_uninstall", Action: "uninstall", Description: "Remove a package from one or more projects with each project's own toolchain.",
		InputSchema: object(map[string]any{"package": str("the package"), "projects": list("the projects, by name or directory")}, "package", "projects"),
		Args: func(in map[string]any) ([]string, error) {
			pkg, err := word(in, "package")
			if err != nil {
				return nil, err
			}
			projects, err := words(in, "projects", true)
			if err != nil {
				return nil, err
			}
			args := []string{"uninstall"}
			for _, p := range projects {
				args = append(args, "-w", p)
			}
			return append(args, pkg), nil
		}},
	{Name: "smith_doctor_fix", Action: "doctor", Description: "Apply the safe fixes smith_doctor_fixes lists. Others are shown, never applied.", InputSchema: noInput, Args: fixed("doctor", "--fix", "--yes")},
	{Name: "smith_upgrade", Action: "upgrade", Description: "Migrate smith.toml to the current schema and regenerate the files smith owns; a project made from a recipe is brought up to date with it. Preview it first with smith_upgrade_preview.", InputSchema: noInput, Args: fixed("upgrade")},
	{Name: "smith_recipe_new", Action: "recipe new", Description: "Write a recipe from an answers file (TOML; see docs/reference/recipes.md) instead of asking questions; the recipe is checked for every combination of answers before it is written. Refuses to replace an existing file unless force is set.",
		InputSchema: object(map[string]any{
			"answers": str("path of the answers file, relative to the workspace"),
			"path":    str("where to write the recipe; default <name>.recipe.toml"),
			"force":   boolean("replace the file if it exists"),
		}, "answers"),
		Args: func(in map[string]any) ([]string, error) {
			answers, err := word(in, "answers")
			if err != nil {
				return nil, err
			}
			args := []string{"recipe", "new"}
			path, ok, err := optional(in, "path")
			if err != nil {
				return nil, err
			}
			if ok {
				args = append(args, path)
			}
			args = append(args, "--from-answers", answers)
			if flag(in, "force") {
				args = append(args, "--force")
			}
			return args, nil
		}},
	{Name: "smith_sync", Action: "sync", Description: "Make every project declare a shared dependency at the same version, then install. With check, only report.",
		InputSchema: object(map[string]any{"check": boolean("report without writing")}),
		Args: func(in map[string]any) ([]string, error) {
			if flag(in, "check") {
				return []string{"sync", "--check"}, nil
			}
			return []string{"sync"}, nil
		}},
	{Name: "smith_pipeline", Action: "pipeline", Description: "Add a ready-made CI pipeline (e2e, package-zip, deploy-pages, deploy-azure-function); without a template, list them.",
		InputSchema: object(map[string]any{"template": str("the template"), "project": str("the app a deploy template works on"), "ci": map[string]any{"type": "string", "enum": []string{"github", "azure", "both"}}}),
		Args: func(in map[string]any) ([]string, error) {
			args := []string{"pipeline"}
			for _, key := range []string{"template", "project", "ci"} {
				v, ok, err := optional(in, key)
				if err != nil {
					return nil, err
				}
				switch {
				case !ok:
				case key == "template":
					args = append(args, v)
				default:
					args = append(args, "--"+key, v)
				}
			}
			return args, nil
		}},
}
