# Editors

The [VS Code extension](../manual/vscode.md) is the editor integration smith
ships; it talks to `smith serve` (below). Other editors and tools build on four
commands whose `--json` output is a stable contract. They keep the shapes the mnci VS Code extension reads, which was
written against the old TypeScript mnci.

| Command | Prints |
|---|---|
| `smith commands --json` | every command: `name`, `aliases`, `group` (`workspace`, `projects`, `dependencies`, `pipeline`, `inspect`; an editor offers everything outside `inspect`), `description`, `arguments` (`name`, `description`, `required`, `variadic`) and `options` (`name`, `flags`, `short`, `description`, `takesValue`: `none`/`required`/`optional`, `variadic`, `repeatable`, `negated`, `defaultValue`) and, for a command with subcommands (`deps up`), `subcommands`: the same description of each, in the same group |
| `smith kinds --json` | what `smith add <kind>` accepts: `kind` (`node-app`, `go-lib`, `python-package`, …, one per language plugin and kind), `language`, `label`, `description`, `flags` |
| `smith projects --json` | each project: `name`, `dir`, `kind`, `ecosystem` (`npm`, `pip`, `go`, or the plugin's name), `targets` (its tasks), and also `root`, `language`, `plugin`, `stack`, `deps`, `tasks` |
| `smith info --json` | `cli` (`version`, `latest` — the newest release, or null when it cannot be found out quickly — and `updateAvailable`) and `workspace` (`root` and `smith.toml` as JSON), or `"workspace": null` outside a workspace |

For example, the start of `smith projects --json` in a workspace with a Go app
that imports a Go lib:

```json
[
  {
    "name": "api",
    "root": "apps/api",
    "dir": "apps/api",
    "kind": "app",
    "language": "go",
    "ecosystem": "go",
    "plugin": "go",
    "deps": ["core"],
    "tasks": ["build", "lint", "package", "slice-check", "start", "test"],
    "targets": ["build", "lint", "package", "slice-check", "start", "test"]
  }
]
```

(Real output is indented one item per line; shortened here.) Run a command with
`--json` and pipe it to `jq` to explore the rest.

`smith doctor --json` prints `findings` (`check`, `ok`, `detail`, `remedy`,
`warning`, `source`) with `passed`, `failed` and `warnings` counts. A finding
that can be fixed carries a `fix`, and the report lists the fixes folded in
`fixes`: each has `description`, `safe`, `ask` (the user must decide), `checks`
it clears, `applies_to` (the projects, when folded), `files` (whole new
contents), `steps` (`argv` commands run in order) and `choices` (`question`
with `options`, each `label` and the `fix` it adds). An editor shows the files
as diffs and walks the choices; `smith doctor --fix --only "<fix>" --choose 1,2
--yes` applies one.

`smith schema` prints the JSON Schema of `smith.toml` (see configuration).

`smith add node-app web` is `smith add app web --lang node`: the combined kinds
`smith kinds` lists are accepted as they are.

Set `SMITH_NO_UPDATE_CHECK=1` to stop `smith info` asking GitHub for the newest
release.

## The engine over JSON-RPC (`smith serve`)

An editor plugin starts the installed smith as a child process, `smith serve`
in the workspace, and talks JSON-RPC 2.0 to it over stdin and stdout, one JSON
message per line. It opens no port. A method stands for each function the
`--json` outputs above give (`projects`, `graph`, `affected`, `doctor`,
`release.preview`, `explain`, ...), so the answer of a query is the same JSON
and the contracts above do not change. What is new is the work that takes time:

- `run` runs tasks and answers when the run ends, and `cancel` stops it and
  its processes. `install`, `uninstall`, `add`, `upgrade` and `doctor.fix` are
  runs too.
- While a run goes, the server sends `event` notifications: `run.started`,
  `task.queued`, `task.started`, `task.output`, `task.finished` (`passed`,
  `failed`, `cached` or `skipped`, with its duration and cache key) and
  `run.finished`. The plugin renders them live.
- `initialize` reports smith's version, the protocol version and the event
  version, so a plugin can tell what the smith it found can do.

Without a server, `smith run build --json-events` prints the same events as
JSON lines. Both are described in [Serve and the event stream](../reference/serve.md).

## AI assistants (MCP)

`smith mcp` serves the workspace over the Model Context Protocol on stdin and
stdout. Register it as an MCP server whose command is `smith mcp`, started in
the workspace. For example, in a client's JSON configuration:

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

The read-only tools (`smith_projects`, `smith_graph`, `smith_affected`,
`smith_kinds`, `smith_stacks`, `smith_commands`, `smith_info`, `smith_doctor`,
`smith_release_preview`, `smith_updates`, `smith_explain`, `smith_recipe_check`,
`smith_plugin_list`, `smith_doctor_fixes`, `smith_upgrade_preview` and
`smith_change_kind_preview`) are
marked so a client may run them without asking. The others change the
workspace or run a plugin: `smith_run`, `smith_add_project`,
`smith_remove_project`, `smith_install`, `smith_uninstall`, `smith_sync`,
`smith_pipeline`, `smith_doctor_fix`, `smith_upgrade`, `smith_recipe_new`,
`smith_change_kind` and
`smith_plugin_check`. Each tool runs the smith command it stands for
and returns its output headed by that command line (`$ smith install -w web
zod`), so you can see and repeat exactly what was done. Values become single
arguments, and one that looks like an option is refused.

### Helping with recipes, plugins and new projects

An assistant creating a recipe, a plugin or a project can look before it
writes:

| To | Tool | Runs |
|---|---|---|
| see the plan of a recipe: the choices asked, the plugin at each level, the merged settings and where each came from, or why it is invalid | `smith_recipe_check` (`choices`: `name=answer`, `all`) | `smith recipe check <recipe> --json` |
| write a recipe from an answers file, checked for every combination first | `smith_recipe_new` (`answers`, `path`, `force`) | `smith recipe new --from-answers` |
| create a project from the workspace's recipe, answering its choices | `smith_add_project` (`choices`) | `smith add <kind> <name> --choice name=answer` |
| see why a task or setting has its value | `smith_explain` (`target`, `recipe`) | `smith explain <project>[:<task>] --json` |
| know what plugins exist: levels, slots, requirements, conflicts | `smith_plugin_list` | `smith plugin list --json` |
| check a plugin keeps the protocol's promises | `smith_plugin_check` (`plugin`, `generate`) | `smith plugin check <name> --json` |
| see, then apply, the fixes doctor can make | `smith_doctor_fixes`, `smith_doctor_fix` | `smith doctor --fix`, `--fix --yes` |
| see, then apply, what a newer smith or recipe changes | `smith_upgrade_preview`, `smith_upgrade` | `smith upgrade --dry-run --json`, `smith upgrade` |

`smith_plugin_check` takes a plugin's name, never a path, because `smith plugin
check ./path` would run that executable. `smith_doctor_fixes` shows diffs and
changes nothing; only `smith_doctor_fix` applies the safe ones. A fix that
needs a decision (steps, questions) is reported with the command that applies
it, and the call fails until someone runs it.
