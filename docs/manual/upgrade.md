# smith upgrade

After you install a newer smith, `smith upgrade` brings the workspace up to it. It
migrates `smith.toml` to the current schema (comments and layout kept) and regenerates
the files smith owns, deleting one the workspace no longer asks for (a dropped CI
provider, say). Files smith wrote once (`README.md`, `.gitignore`) and your own are left
alone, and steps inside `smith:slot` markers in the pipelines are kept.

A project made from a recipe is brought up to date with the recipe as it is now: smith
resolves it again with the answers the project recorded, updates the recipe version, the
stack and the settings recorded in `smith.toml`, and has the project's plugins rewrite the
files they own. Starter files, which a plugin wrote once when the project was created, are
never touched; when the recipe adds a plugin whose starter files the project lacks,
`upgrade` says so.

## When to use it

- After `npm i -g monosmith@latest` (or any other way of updating smith).
- After changing a setting that decides which files exist, such as `ci.providers` or
  `ci.devcontainer` (`smith ci generate` regenerates only the pipelines).
- In CI, as `smith upgrade --check`, to fail a pull request that leaves the workspace
  behind.

Do not use it to apply dependency updates ([`smith up`](up.md)) or to bring in a repository
([`smith adopt`](adopt.md)).

## Synopsis

```sh
smith upgrade [--dry-run] [--check] [--json] [--yes]
```

## Arguments and flags

<!-- smith:begin upgrade -->
Bring the workspace up to this smith and its recipes: migrate smith.toml, regenerate the files smith owns

Usage: `smith upgrade [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | change nothing; fail when something would change |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--defaults` | boolean | off | move the workspace to the latest defaults vintage (smith doctor says which it uses) |
| `--dry-run` | boolean | off | show the plan, with the change to smith.toml, and change nothing |
| `--json` | boolean | off | print the result as JSON |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end upgrade -->

- `--dry-run` shows the plan, including the change to `smith.toml`, and changes nothing.
  Exit code 0 either way.
- `--check` also changes nothing, and fails (exit 1) when something would change.
- `--defaults` moves the workspace to the latest defaults vintage, with a plan, and does nothing
  else. A workspace keeps the defaults of the vintage in `[workspace] defaults` whatever smith version
  runs ([ADR 0054](../architecture/decisions/0054-defaults-have-vintages.md)); `smith doctor` says
  which vintage it uses.
- `--json` prints the result as JSON (`{}` when nothing changed).
- In a terminal, outside CI, smith upgrade first prints the plan (from a dry run) and asks
  `Go ahead? [y/N]`. `--yes` skips the question; without a terminal it refuses unless `--yes` is given.
- With `commit_operations = true` under `[workspace]` and a clean git work tree, the upgrade is one
  commit (`chore: upgrade the workspace`); [`smith undo`](undo.md) reverts it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The workspace is up to date, or the upgrade was applied, or `--dry-run` printed a plan. |
| 1 | `--check` found something that would change, or a step failed. |

## Files and environment

Reads and may rewrite `smith.toml`; regenerates `.github/workflows/ci.yml`,
`azure-pipelines.yml`, `.gitlab-ci.yml`, `.devcontainer/devcontainer.json` as configured,
and the files plugins own for recipe projects. It also rewrites `.claude/agents/smith.md` when
smith wrote it and nobody edited it; an edited file is left as it is and named in a hint, and a
workspace without the file gets none (`smith agent install` adds it). `[agent] enabled = false`
turns this off. Nothing is committed.

## Examples

Nothing to do:

```sh
smith upgrade --check
```

```text
The workspace is up to date.
```

A setting that decides a file's existence was changed (`ci.devcontainer` was removed, so
the default `false` applies, and the file is still on disk):

```sh
smith upgrade --check
```

```text
would change: .devcontainer/devcontainer.json is no longer configured
smith: the workspace is not up to date: run smith upgrade
```

The exit code is 1. Turning the setting back on and upgrading:

```sh
smith config set ci.devcontainer true
smith upgrade
```

```text
✔ ci.devcontainer = true
upgraded: .devcontainer/devcontainer.json
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith: the workspace is not up to date: run smith upgrade` | `--check` found a difference. Run `smith upgrade` and commit. |
| `there is no smith workspace here` | Run it inside a workspace. |

## Related

[`smith ci generate`](ci-generate.md), [`smith doctor`](doctor.md), [`smith config`](config.md),
[Installing](../getting-started/installing.md).
