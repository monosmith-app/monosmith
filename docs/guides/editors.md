# Editors

The [VS Code extension](../manual/vscode.md) is the editor integration smith
ships; it talks to `smith serve` (below). A first JetBrains plugin (`apps/jetbrains`, a tool window of the projects and tasks, task runs in the Run window) reads `smith projects --json`; see its [README](../../apps/jetbrains/README.md). Other editors and tools build on four
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

## Commands an editor can call besides the four

The VS Code extension reads the views from `smith serve`. For what changes files it runs the
`smith` command itself, in the workspace folder, so it prints the same text a terminal shows,
asks the person first, and passes `--yes` only after the person agreed. An editor can do the same:

| To offer | Run | Reads |
|---|---|---|
| Add a project | `smith kinds --json`, `smith stacks --json`, `smith choices --json` for the questions; `smith add <kind> <name> --lang <language> [--stack <stack>] [--with a,b] [--fresh] --dry-run` for the plan; the same with `--yes` to add | JSON for the lists. The plan is text. `smith choices --json` lists every choice with its `axis`, `language`, `requires`, `default` and `many`; filter by language. The first question of an add ("What are you building?") is the `outcomes` list of `smith compose --json`: each has an `id`, a `title`, a `kind` (app, lib or package), its `languages`, the `words` that find it and the `items` (frameworks, shells or hosts) of its layer. An editor shows that list and does not keep one of its own. |
| Derived files | `smith derive --check`; `--explain=<file>`, `--replace=<file>`, `--adopt=<file>`, `--eject=<file>` with `--yes` | Text only: one line per file that needs action, `State  path  (renderer)  detail`, then a count. A `--json` output for `--check` would let an editor stop reading lines. |
| Who owns a file | `smith why --json -- <file>` | `path`, `known`, `kind`, `owner`, `summary`, `project`, `source`, `rule` and `change` (the steps). |
| Why a task runs again | `smith why-miss <project>:<task> --json` | `status` (`hit`, `changed`, `no-earlier-run`), `key`, `lastRun` and `differences`. |
| Undo | `smith undo --dry-run`, then `smith undo --yes` | Text. Reverts an operation that was committed (`--commit`, `smith commit` or `commit_operations = true` under `[workspace]`). |
| A report | `smith report -o <file>` | An HTML page. |
| The agent file | `smith agent file --check`, then `smith agent file` | Text. `--check` fails when `AGENTS.md` or `CLAUDE.md` is missing or old. |
| The release channel | `smith release --dry-run --json` (or `release.preview` over `serve`) | `channel` is set on a channel branch. The Markdown form starts with the notice. |

A problem goes to standard error as `✖ what failed`, then a line `  → the fix` for each way out.
Show both to the person.

## More commands the VS Code extension offers

Each is a command in the palette (`MonoSmith: …`) and in the "…" menu of the Projects view. A
command that changes files shows smith's plan (`--dry-run`) or what it costs, asks, and only then
runs the command with `--yes`. Nothing is committed.

| Command | Runs | Reads from smith |
|---|---|---|
| Environments of the Web Apps; Add or Remove Environments… | `smith environments list\|add\|remove` | `environments list --json` for the web apps and their names; `config get workspace.environments` starts the names to add on the workspace's own |
| Cache: Status, Check for Damage, Remove What Is Broken, Trim to Its Limit, Delete It All | `smith cache status\|verify\|verify --fix\|prune\|clear` | text |
| Copies Between Projects; Tools the Workspace Needs; Info | `smith copies`, `smith tools`, `smith info` | text |
| Add a Ready-Made Pipeline…; Make the Icons of a Tauri App…; Docker Hub Descriptions | `smith pipeline`, `smith tauri icon`, `smith docker description` | `smith pipeline` lists the templates |
| Why Would a Task Run Again? | `smith why-miss <project>:<task>` | text |
| Regenerate the CI Pipelines…; Derived Files: Check, Preview, Update | `smith ci generate`, `smith derive` | text and the plan |
| Change a Setting of smith.toml…; Reset a Setting… | `smith config set\|unset` | `config list --json`: the type, the allowed values, the current value and the reason of the default |
| Turn a Plugin On or Off…; Choose the Architecture… | `smith plugin enable\|disable`, `smith architecture <name>` | `plugin list --json`; `smith architecture` |
| Install a Plugin from the Marketplace… | `smith plugin search <words> --json`, then `smith plugin install <name> [--trust --pin]` | the search result: name, newest version, description, roles; the trust question says the plugin runs unsandboxed with your privileges |
| Connect smith to a Coding Assistant… | `smith mcp install --claude --vscode` | none |
| New Recipe… | `smith recipe new --fixed …  --chosen …` | `recipe new --list-levels --json`; for each level that takes one plugin the extension asks who decides it: the recipe (default) or the people who use it |

The release picker does not list a package that was never released for a forced bump: it lists it
on one line and offers the commit that sets its first version (`neverReleased` in
`smith release --dry-run --json`).

Not in the extension: `smith plugin draft` (a plugin author works in a terminal on a sample project), and
`smith serve`, `mcp`, `lsp`, `agent` and the dev servers, which own a terminal.

## Committing, releasing and changing a project from an editor

smith commits and releases only when asked, and each of these has a plan flag that changes nothing.
The VS Code extension follows one order for all of them: it shows the plan, then asks, then runs
the command with `--yes`.

| Function | Plan | Change |
|---|---|---|
| Commit the changes of the last operation | `smith commit --dry-run` (it lists the files of the operation, and the other changed files when there are any) | `smith commit --only-operation --yes` or `smith commit --all --yes`, by the choice of the person |
| Release | `smith release --dry-run` (add `--json` to read the packages) | `smith release --yes [--hold a,b] [--bump pkg=minor]`. A release tags and publishes: it cannot be undone. |
| Remove a project | `smith remove <project> --dry-run` | `smith remove <project> --keep-copy --yes` or `--delete --yes` |
| Make a project an app, lib or package | `smith kind <project> <kind> --dry-run` | `smith kind <project> <kind> --yes` |
| Unlink a library | none | `smith uninstall -w <project> workspace:<library> --yes` |

`smith kinds --json` names the language of a kind by its name (`typescript` for `node-app`), and
`smith compose --json` by its id (`node`). Take the id from the kind (`node` in `node-app`) to match
the two. The `layers` list of `smith compose --json` gives one title for each layer
(`shell` is "Desktop shell"): use it wherever a choice is named.

An editor that asks the choices of an add follows these rules, so pressing Enter through the questions gives
the safe result: a list of one answer starts on its `default` (for example the test runner) or, when its axis
has no default (a desktop shell, a style), on "None", listed first. The answer of an earlier step is kept:
the question of a framework, shell or cloud is not asked twice, and a website does not list cloud hosts
(`layer` is `host`). The confirmation says the choices in words and marks the defaults ("React, ESLint
(default), Vitest (default); no desktop shell, no style"). A notification shows at most three buttons with
short verbs; the long text goes in the message, and more buttons go behind "More…".

## Starting in an empty folder

An editor that opens a folder with no `smith.toml` can offer two commands. `smith new --here`
creates the workspace in the current folder (the name comes from the folder unless one is given,
and it refuses a folder that holds anything besides `.git`, `.vscode`, `.DS_Store`, `Thumbs.db` or
`desktop.ini`), and `smith adopt` brings a folder with files into MonoSmith. The VS Code extension
offers them as "Create a new MonoSmith workspace" and "Bring this folder into MonoSmith", in the
Projects view and in the status bar menu.

An editor does not have to hand `smith new` a terminal. Every question of `smith new` has a flag, so
the editor can ask the questions itself, as the TUI form does ("Start with the defaults", or
"Customise"), say that the folder name is used, show `smith new <name> --here --yes <flags> --dry-run`
(it lists the files and writes nothing) and run the same line without `--dry-run` when the person
agrees. The VS Code extension does this and keeps "Answer in a terminal" as a third way.

## The language server (`smith lsp`)

`smith lsp` is a Language Server Protocol server for `smith.toml`, on standard input and
output. Any editor that starts a language server can use it: set `smith lsp` as the server
command for files named `smith.toml`. It reports what smith would reject (syntax, unknown
keys, wrong types, rejected values), completes table names, keys and enumerated values, and
shows a key's description on hover, all from the schema. Recipe files, doctor findings as
diagnostics and code actions are not built yet ([#303](https://github.com/monosmith-app/monosmith/issues/303)).
The [VS Code extension](../manual/vscode.md) does not start it yet. See [`smith lsp`](../manual/lsp.md).

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

`smith agent install` registers the server below for Claude Code and VS Code and adds a
Claude Code agent that knows the tools ([the agent guide](agent.md)).

`smith mcp` serves the workspace over the Model Context Protocol on stdin and
stdout. Register it as an MCP server whose command is `smith mcp`, started in
the workspace. For example, in a client's JSON configuration:

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

The read-only tools (`smith_projects`, `smith_graph`, `smith_affected`,
`smith_kinds`, `smith_stacks`, `smith_choices`, `smith_commands`, `smith_info`, `smith_doctor`,
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
| see the plan of a recipe: the choices asked, the plugin at each role, the merged settings and where each came from, or why it is invalid | `smith_recipe_check` (`choices`: `name=answer`, `all`) | `smith recipe check <recipe> --json` |
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
