# smith adopt

`smith adopt` reads a repository that smith does not manage yet (npm workspaces, Go
modules, Python projects) and reports what bringing it under smith involves: the
projects smith would see, what blocks, and the step that clears each blocker. By
default it changes nothing. With `--init` it writes `smith.toml` and the pipelines
and commits.

## When to use it

- You have a repository with packages or modules and want smith to run it.
- You want to know first what smith would do (`smith adopt`, or `--dry-run`).

You do not have to choose between this command and [`smith convert`](convert.md): in an
Nx or Turborepo workspace (or an mnci one) `adopt` recognises it and hands over to
`convert`, so `smith adopt` is the one way in for any repository.
Do not use it in a folder that already has a `smith.toml` (`smith doctor` checks that
workspace, [`smith upgrade`](upgrade.md) brings it up to date).

In a terminal, a bare `smith` in a folder that holds files but no workspace opens the
adopt wizard, which does the same in screens; see
[The smith terminal interface](tui.md#the-adopt-wizard).

## Synopsis

```sh
smith adopt [--dry-run] [--json] [--tags] [--init [--install] [--eslint]] [--level conservative|full]
```

## Arguments and flags

<!-- smith:begin adopt -->
Bring an existing repository under smith: report what it involves; --init takes the steps (start here)

Usage: `smith adopt [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | only report (this is the default; the flag makes it explicit) |
| `--eslint` | boolean | off | with --init, write the two-line root eslint.config.mjs and add @monosmith/eslint-config to the root devDependencies |
| `--init` | boolean | off | write smith.toml and the pipelines, and commit |
| `--install` | boolean | off | with --init, also run npm install for the Node projects |
| `--json` | boolean | off | print the report as JSON |
| `--level <level>` | `conservative` or `full` | `conservative` | how far to go: conservative (keep the build and lint setup) or full (smith convert's, for Nx and Turborepo workspaces) |
| `--tags` | boolean | off | create the baseline release tags the report asks for, locally |

<!-- smith:end adopt -->

Interactions:

- Without `--init` or `--tags` the command only reports. `--dry-run` is the same and
  makes it explicit.
- `--tags` and `--init` both need a clean git tree. `--tags` creates, locally, the
  release tags a project needs where its tags were made under another name; nothing
  is pushed. `--init` writes `smith.toml` and the pipelines and commits them
  (`chore: adopt smith`). They can be combined: `smith adopt --tags --init`.
- `--install` and `--eslint` only mean something with `--init`. `--install` runs
  `npm install` for the Node projects. `--eslint` writes the two-line root
  `eslint.config.mjs` and adds `@monosmith/eslint-config` to the root devDependencies.
- `--level conservative` (the default) keeps the repository's own build and lint
  setup. `--level full` makes smith the only tool; it exists for Nx and Turborepo
  workspaces, where it is `smith convert --level full`. For a plain repository
  `adopt --level full` says so and stops (the work is tracked in issue 553 of the
  project's repository).
- A pipeline of yours at a path smith owns is kept beside the new one as
  `<file>.pre-smith`. A `.gitignore` of yours is kept and gets `.smith/` added.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The report was printed with nothing blocking, or the requested steps were applied. |
| 1 | Something blocks (dirty tree, a recognised Nx or Turborepo workspace, an unsupported `--level`), the folder is already a workspace, or a step failed. |

## Files and environment

Reads the repository's manifests (`package.json` workspaces, `go.mod`, `pyproject.toml`)
and its git history. With `--init` it writes `smith.toml`, `.gitattributes`,
`README.md`, `.github/workflows/ci.yml` (and the other providers'
files), `.devcontainer/devcontainer.json`, `go.work` when there are Go projects, and
edits `.gitignore`. It reads no smith-specific environment variable. It does not write the
MonoSmith agent: the result names `smith agent install`, which adds the agent for Claude Code and
VS Code after showing the files.

## Examples

Report only, on a repository with an npm package and a Go service:

```sh
smith adopt
```

```text
~/work\old would be workspace "old" (scope @old, CI github) with:
  PROJECT              LANGUAGE KIND     FOLDER
  lib                  node     package  packages/lib
  api                  go       app      services/api
  kind: apps/ holds apps, libs/ libs, packages/ packages; elsewhere a Go main package is an app,
  a public Node package is a package and the rest are libs. To change one: "smith": {"kind": "app"}
  in its package.json, or a // smith:kind app comment in its go.mod.

! smith doctor will report: the workspace has Go projects but no go.work
  → --init runs go work init and go work use for you
! smith doctor will report: node_modules is missing at the workspace root
  → --init --install runs npm install for you (or run it yourself afterwards)
✔ nothing blocks: smith adopt --init makes it a workspace (add --install to run npm install, --eslint to write the lint config), then says what smith doctor still reports

Nothing changed (adopt only reports). To apply:
  → smith adopt --init
```

Apply it, with the baseline tags:

```sh
smith adopt --init --tags
```

```text
  kept your .gitignore
  added .smith/ to your .gitignore
  wrote .gitattributes
  wrote smith.toml
  wrote README.md
  wrote .github/workflows/ci.yml
  wrote .devcontainer/devcontainer.json
  wrote go.work: ./services/api
  committed: chore: adopt smith
✔ adopted. smith doctor still reports 2 thing(s) for you:
  ✖ node_modules is missing at the workspace root
    → run `npm install` at the workspace root
  ✖ the workspace root has no eslint.config file, so the slice rules do not run on lib
    → make the root eslint.config.mjs: ...
  next: follow the steps above, then run smith doctor
```

(The last remedy is cut here; the real output prints it in full.) The same report as
JSON, for a script, has `root`, `name`, `scope`, `providers`, `clean`, `projects`,
`blockers` and `warnings` (each warning has `message`, `step` and `steps`):

```sh
smith adopt --json
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `this is a Turborepo workspace: its tasks are turbo.json's to describe` | Run `smith convert --dry-run`. A recognised Nx workspace is refused the same way (not captured for this page). |
| `this is already a smith workspace` | The folder has a `smith.toml`. Run `smith doctor` or `smith upgrade`. |
| `something blocks adoption; see smith adopt` | `--init` or `--tags` was refused. Run `smith adopt` to read the blockers (usually a dirty git tree: commit or stash). |
| `--level full makes smith the only tool, ... not yet for a plain repository` | Use the default level, or `smith convert --level full` in an Nx or Turborepo workspace. |

## Related

[`smith convert`](convert.md), [`smith doctor`](doctor.md), [`smith upgrade`](upgrade.md),
[Adopting an existing repository](../getting-started/adopting.md),
[Converting from Nx or Turborepo](../getting-started/converting.md).
