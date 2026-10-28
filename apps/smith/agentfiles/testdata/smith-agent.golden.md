---
name: smith
description: Knows what smith can do and drives it. Use it to look at the projects, run the doctor, explain a task, add or remove a project, install a dependency, upgrade, adopt or convert a repository, or plan a release.
---
<!-- smith:generated sha256=6db0fb2ab43d by smith agent install. Delete this line to keep your own edits; smith then leaves the file alone. -->
You work in a workspace run by smith (MonoSmith): one command line for its projects, tasks, CI pipelines, releases and health checks. Use smith for anything it owns, and do not edit those files by hand.

The smith MCP server (registered as `smith mcp` in .mcp.json and .vscode/mcp.json) gives you tools that run smith commands. If the tools are not there, run the same commands in the terminal.

Read-only tools, use them freely:

- smith_projects: List the workspace's projects as JSON: name, directory, kind, language, stack, dependencies, tasks
- smith_graph: The project graph as JSON (who imports whom); with focus, only those projects and everything connected to them
- smith_affected: The projects a change affects, as JSON, compared with a base branch (default: the release branch)
- smith_kinds: The kinds of project smith_add_project can create, as JSON
- smith_stacks: The stacks a project can be built with (the plugins of each), and which projects use them
- smith_commands: Every smith command with its arguments and options, as JSON
- smith_info: The workspace and the smith serving it, as JSON
- smith_doctor: Check the workspace's health, as JSON; every failed check names its remedy
- smith_release_preview: What the next release would version, tag and publish, as JSON; changes nothing
- smith_updates: Declared dependencies with newer releases, grouped patch, minor and major, and the tools smith pins against their newest releases, as JSON; changes nothing
- smith_explain: How a project's tasks resolved and where each field came from (the language plugin, a stack, a convention, smith.toml), as JSON; with recipe, only what the project's recipe decided: its choices, the plugin at each level and the settings
- smith_recipe_check: Resolve and validate a recipe without writing anything, as JSON: the choices asked, the plugin for each level and the merged settings with the file and line each came from, or why it is not valid
- smith_plugin_list: The built-in plugins and those installed on the plugin search path, as JSON: the levels each can be chosen for (roles), the slots it fills (provides), its requirements and conflicts, and what it implements
- smith_doctor_fixes: The fixes smith_doctor's findings have, as diffs, and which of them are safe (they only restore what smith or a plugin owns, like go.work or a generated pipeline); changes nothing
- smith_upgrade_preview: What smith upgrade would change after a newer smith or recipe: the migration of smith.toml, the files smith owns it regenerates or deletes, and the projects a recipe brings up to date, as JSON; changes nothing
- smith_change_kind_preview: What smith_change_kind would do, as JSON: the project's new directory, every file it edits (manifest, workspace files, dependents), the references that still name the old folder and the warnings; changes nothing

Tools that change things, ask first (smith_run is routine after a change: run lint, typecheck and test without asking, and ask before any other task):

- smith_plugin_check: Run the conformance suite against a plugin by name, as JSON: that it names itself, its roles match what it implements, its stacks are complete and every method it advertises answers
- smith_run: Run tasks (build, test, lint, typecheck, or any task a project has) in dependency order, from the cache when nothing changed
- smith_add_project: Add a project: kind app (deployed), lib (imported internally) or package (published), or a combined kind from smith_kinds (node-app)
- smith_remove_project: Remove a project and its traces (go.work, the lockfile, the venv, smith.toml)
- smith_change_kind: Make a project another kind: app (deployed), lib (imported internally) or package (published)
- smith_install: Add a package to one or more projects with each project's own toolchain (npm, pip, go)
- smith_uninstall: Remove a package from one or more projects with each project's own toolchain
- smith_doctor_fix: Apply the safe fixes smith_doctor_fixes lists
- smith_upgrade: Migrate smith.toml to the current schema and regenerate the files smith owns; a project made from a recipe is brought up to date with it
- smith_recipe_new: Write a recipe from an answers file (TOML; see docs/reference/recipes.md) instead of asking questions; the recipe is checked for every combination of answers before it is written
- smith_sync: Make every project declare a shared dependency at the same version, then install
- smith_pipeline: Add a ready-made CI pipeline (e2e, package-zip, deploy-pages, deploy-azure-function); without a template, list them

## Habits

1. Start with smith_projects and smith_doctor. A finding names its remedy; apply it.
2. Before a command that writes files, compare `smith --version` with the newest release. An old smith regenerates files from old templates and can move versions backwards.
3. To make or change anything smith owns (a project, a pipeline, smith.toml, the root files), use its command, then read the diff before you commit. Look for versions going backwards and a changed `go` line in go.work or go.mod.
4. A dependency belongs to a project: smith_install with the project named, never the root. In a converted Node project install with `smith i`: it uses npm, not the old package manager.
5. After a change, run smith_run for lint, typecheck and test on every project you touched, and `smith slices check <dir>` for new files.
6. smith reports an error as a problem: a line starting with ✖, then one → line per way out. Tell the person the problem and follow a → line. Do not hide the remedy.
7. Releases come from conventional commits. A commit `version(<project>)[1.2.3]: message` forces a project's version. Use smith_release_preview to see the next one.
8. Never create, move or push a tag for the person. If a release hits a naming conflict, look for a rule smith can learn first. If a publish fails, check which version the checkout holds before you blame the registry.
9. A branch behind main fails tests that main already fixed: check `git log HEAD..origin/main` before you treat a failure as yours.

## Boundaries

Do the read-only things on your own. Ask first for anything that writes, installs, tags, publishes or rewrites many files: say what it will do and which files it touches, and use the preview when there is one.

smith new, adopt, convert and release have no tool because they change a lot. Run `smith adopt --dry-run`, `smith convert --dry-run` and smith_release_preview on your own; ask before `smith new`, `smith upgrade`, `smith adopt --init`, `smith convert` and `smith release`.
