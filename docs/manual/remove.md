# smith remove

`smith remove` deletes a project: its directory, what its plugin owns outside it (the
`go.work` entry, the npm lockfile, the Python venv), its tables in `smith.toml`, and its
place in the regenerated CI pipelines. A project that other projects import is refused
unless `--force`. Its release tags are kept.

## When to use it

- A project is no longer wanted and you want its traces gone in one step.

Do not use it to stop a project from being a smith project while keeping its code: add
its folder to `[workspace] exclude` with [`smith config set`](config-set.md) instead.

## Synopsis

```sh
smith remove <project> [--dry-run] [--force] [--keep-copy | --delete] [--json] [--yes]
```

## Arguments and flags

<!-- smith:begin remove -->
Remove a project: its directory, its traces in the workspace, its CI

Usage: `smith remove <project> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes | the project to remove |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--delete` | boolean | off | delete the folder for good, including files git cannot bring back |
| `--dry-run` | boolean | off | show what would change; change nothing |
| `--force` | boolean | off | remove a project other projects still import |
| `--json` | boolean | off | print what changed as JSON |
| `--keep-copy` | boolean | off | move the project folder to .smith/removed/<name>-<time>/ instead of deleting it |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end remove -->

- `<project>` is a project name, as [`smith projects`](projects.md) lists it.
- `--dry-run` prints what would change and changes nothing.
- `--yes` skips the confirmation. smith prints a plan first (the directory, then the workspace
  files that change, with their owner where `smith.lock` records one). In a terminal it asks
  `Go ahead? [y/N]`; without a terminal it refuses, so pass `--yes` in scripts.
- With `commit_operations = true` under `[workspace]` and a clean git work tree, the removal is one
  commit (`chore: remove core`); [`smith undo`](undo.md) reverts it.
- `--force` removes a project other projects still import; their imports are left
  broken.
- `--json` prints what changed: `project`, `root`, `files`, `notes` and `mentions`.

## Files git cannot bring back

Deleting the folder loses what git does not hold: untracked files and uncommitted changes. When the
folder has any, smith lists them and asks what to do with the folder:

1. Move it to `.smith/removed/<name>-<time>/` (the default). The project leaves the workspace and
   nothing is lost. `.smith/` is ignored by git.
2. Delete it for good. The files listed cannot be brought back.
3. Cancel. Nothing changes.

Without a terminal smith refuses and names the flags: `--keep-copy` (answer 1) or `--delete`
(answer 2). `--yes` alone does not choose, because it cannot know which you want. A folder that
holds nothing git cannot restore is deleted as before, after the plan.

## Files that still name the project

smith cannot know about a hand-written workflow, a `.code-workspace` task or a doc that names the
project. After the removal (and in `--dry-run`), it searches the files git tracks for the project name
as a whole word and for its former root path. It prints the result under `Still mentions <name>`:
at most 20 files with the first matching line, then `and N more`, in three groups: workflow files
(`.github/workflows`, `azure-pipelines*`, `.gitlab-ci*`), editor files (`*.code-workspace`, `.vscode`,
`.idea`) and other files.

smith does not change these files. Review them. The search skips binary files, lockfiles, the removed
directory and files smith regenerates, and it never fails the command. A name of 4 characters or fewer
(`web`, `api`) is too common as a word, so only the root path is searched, and the output says so.
In `--json`, `mentions` is a list of `{file, line, kind}` with `kind` one of `workflow`, `editor` or
`other`; `mentionsRootOnly` is true for a short name.

```text
Still mentions mvd-tui
  smith does not change these files. Review them.
  workflow files
    .github/workflows/release.yml:3
  editor files
    mvd.code-workspace:2
  other files
    docs.md:2
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The project was removed, or the dry run printed. |
| 1 | No such project, other projects import it (without `--force`), or a step failed. |

## Files and environment

Deletes the project folder; edits `smith.toml` (its `[projects.<name>]` tables), the
lockfile or `go.work`, and regenerates the CI files. Run `git status` afterwards: nothing
is committed.

## Examples

```sh
smith remove core --dry-run
```

```text
Would remove core (libs/core)
  changes go.work
  the pipelines would be regenerated
```

```sh
smith remove core --yes
```

```text
Removed core (libs/core)
  changes .github/workflows/ci.yml
  changes go.work
```

The dry run as JSON, for a Python package whose venv entry is uninstalled:

```sh
smith remove tools --json --dry-run
```

```json
{
  "project": "tools",
  "root": "packages/tools",
  "notes": [
    "would uninstall acme_tools from .venv",
    "the pipelines would be regenerated"
  ]
}
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `no project named "nope"` | Check the name with `smith projects`. |
| `refusing to remove without a terminal; pass --yes` | Standard input is not a terminal, so the `Go ahead? [y/N]` question cannot be asked. Pass `--yes`. |
| A refusal naming the projects that import it | Remove or change those first, or pass `--force`. (The exact wording was not captured for this page.) |

## Related

[`smith add`](add.md), [`smith projects`](projects.md), [`smith graph`](graph.md) (who imports
what), [`smith config`](config.md).
