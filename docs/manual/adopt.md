# smith adopt

`smith adopt` reads a repository that smith does not manage yet (npm workspaces, Go
modules, Python projects) and reports what bringing it under smith involves: the
projects smith would see, what blocks, and the step that clears each blocker. By
default it changes nothing. With `--init` it writes `smith.toml` and the pipelines.
It commits only with `--commit`.

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
smith adopt [--dry-run] [--json] [--tags] [--init [--install] [--eslint] [--move <folder>|none] [--yes]] [--level conservative|full]
```

## Arguments and flags

<!-- smith:begin adopt -->
Bring an existing repository under smith: report what it involves; --init takes the steps (start here)

Usage: `smith adopt [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--agent <agent>` | string |  | with --init, add the MonoSmith agent files (and merge .mcp.json), skip them for now, or never (records [agent] enabled = false) |
| `--ci <ci>` | string, repeatable |  | with --init, the CI providers to write pipelines for (github, azure, gitlab, circleci, bitbucket), or none |
| `--ci-existing <ci-existing>` | string |  | with --init, replace a pipeline of yours (kept as <file>.pre-smith) or keep it and write none for that provider |
| `--commit` | boolean | off | with --init, commit the change as one operation, so smith undo can revert it |
| `--dry-run` | boolean | off | only report (this is the default; the flag makes it explicit) |
| `--eslint` | boolean | off | with --init, write the two-line root eslint.config.mjs and add @monosmith/eslint-config to the root devDependencies |
| `--gitattributes <gitattributes>` | string |  | with --init, add or skip the .gitattributes that sets `* text=auto eol=lf` (a repository that has one keeps it) |
| `--init` | boolean | off | write smith.toml and the pipelines (nothing is committed unless --commit) |
| `--install` | boolean | off | with --init, also run npm install for the Node projects |
| `--json` | boolean | off | print the report as JSON |
| `--level <level>` | `conservative` or `full` | `conservative` | how far to go: conservative (keep the build and lint setup) or full (smith convert's, for Nx and Turborepo workspaces) |
| `--move <move>` | string |  | with --init (or --dry-run), the folder to move the project at the repository root into (apps/web); none keeps it where it is |
| `--tags` | boolean | off | create the baseline release tags the report asks for, locally |
| `--yes` | boolean | off | do not ask: a project at the root moves only when --move says where, and every other open choice of --init takes its safe default and prints it |

<!-- smith:end adopt -->

Interactions:

- Without `--init` or `--tags` the command only reports. `--dry-run` is the same and
  makes it explicit.
- `--tags` and `--init` both need a clean git tree. `--tags` creates, locally, the
  release tags a project needs where its tags were made under another name; nothing
  is pushed. `--init` writes `smith.toml` and the pipelines and leaves them in the
  work tree. With `--commit` (or `commit_operations = true`) it commits them as
  `chore: adopt smith`, and `smith undo` reverts that commit. They can be combined:
  `smith adopt --tags --init --commit`.
- A repository that uses pnpm (`pnpm-workspace.yaml`, or `workspace:` and `catalog:`
  ranges) blocks `--init`: smith installs with npm, which cannot read them. The report
  names the way on. A workflow of yours under `.github/workflows` is listed with a
  warning that CI would run twice beside the one smith generates.
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

## What `--init` asks

`--init` does not decide in silence. It asks four questions, each with its consequence, then prints
the plan with a diff of every file of yours it edits, and asks `Go ahead? [y/N]`:

| Question | Flag | Safe default |
|---|---|---|
| Write `.gitattributes` with `* text=auto eol=lf`? (asked only if the repository has none) | `--gitattributes add\|skip` | skip: your line endings stay |
| Add the agent files and merge the MCP server into `.mcp.json`? | `--agent add\|skip\|never` | skip; `never` records `[agent] enabled = false` |
| Which CI? GitHub, Azure, GitLab, CircleCI and Bitbucket are detected | `--ci <provider>\|none` | the provider the repository uses, else GitHub |
| A pipeline of yours at the path smith writes: replace it (yours moves to `<file>.pre-smith`) or keep it? | `--ci-existing replace\|keep` | replace, keeping the copy |

Without a terminal and without `--yes` it refuses and names these flags. With `--yes` each open
question takes its safe default and a line says which. `--init --dry-run` prints the same plan and
changes nothing.

## A project at the repository root

smith manages one folder per project (`apps/<name>`, `libs/<name>`, `packages/<name>`). A
repository whose one project sits at the root (a `package.json`, `go.mod`, `pyproject.toml`
or `Cargo.toml` there) has no such folder, so `--init` asks where to move it, and never picks
for you:

```text
Move it into a folder of its own so smith can manage it?
  1) apps/web-app (suggested)
  2) packages/web-app
  3) libs/web-app
  4) another folder
  5) no, leave it
```

- The suggestion is an app folder when the project looks like a program (a `start` script
  or `bin` in `package.json`, a `package main`, project scripts in `pyproject.toml`,
  `src/main.rs`), else a package folder.
- The name comes from the manifest: the npm scope or module path is dropped and the rest
  becomes kebab-case. You can edit it. Answer 4 asks for the whole folder.
- Without a terminal, or with `--yes`, `--init` moves nothing unless `--move <folder>` says
  where, for example `smith adopt --init --move apps/web`. Without it the command stops and
  prints the exact commands. `--move none` keeps the project at the root (an empty workspace,
  add projects with `smith add`).
- The move is `git mv` of the project's files. These stay at the root: `.git`, `.gitignore`,
  `.github` and the other CI files, editor and agent settings, `smith.toml`, README, LICENSE,
  CHANGELOG, `docs`, compose files, and for Node the lockfile (`package-lock.json` is
  rewritten by `npm install` for the workspace layout). Everything else moves, including
  source, tests, configs and, for Go, `go.mod` and `go.sum` (the module path does not change).
  Lines of `.gitignore` that start with `/` and name a moved entry follow it.
- It refuses, and changes nothing, when the root holds two manifests (a Go module and a
  `package.json`), when the folder already holds files, or when the folder's first part is a
  file or a folder that would move.
- `smith adopt --dry-run` prints the question and the `git mv` commands for the suggested
  folder; with `--move <folder>` it prints the commands for that folder.
- The move is staged, not committed. With `--commit` it is a commit of its own
  (`chore: move <name> into <folder>`, `git revert HEAD` takes it back), followed by the
  adoption commit that `smith undo` reverts.

The adopt wizard of the terminal interface and the VS Code adopt command ask the same
question, with a name field and a "Commit this change so it can be undone" box that starts
unticked, and then run the step with `--move`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The report was printed with nothing blocking, or the requested steps were applied. |
| 1 | Something blocks (dirty tree, a recognised Nx or Turborepo workspace, an unsupported `--level`), the folder is already a workspace (a `smith.toml` that cannot be read is reported with its file, line and column), a project at the root has no `--move` answer, or a step failed. |

## Files and environment

Reads the repository's manifests (`package.json` workspaces, `go.mod`, `pyproject.toml`)
and its git history. With `--init` it writes `smith.toml`, `.gitattributes`, `README.md`, `.github/workflows/ci.yml`
(and the other providers' files), `.devcontainer/devcontainer.json`, `tools/smith/install.sh`, the MonoSmith agent
(`.claude/agents/smith.md`,
`.mcp.json` and `.vscode/mcp.json`, see [the agent guide](../guides/agent.md); `smith agent install` adds it
later to a workspace that lacks it), `go.work` when there are Go modules, and edits `.gitignore`. It reads no
smith-specific environment variable.

A repository whose Go code is one module at the root (a `go.mod` there, and no other) has no manifest per
project. Adopt finds the programs (each folder with a `package main`, an app) and the libraries (each folder
under `libs/`, `pkg/` and `internal/`, a lib), lists them as projects, and with `--init` declares them in
`smith.toml` as `[projects.<name>]` tables with a `root`, `language = "go"` and a `kind`
([the Go plugin](../plugins/using/catalogue/go.md#projects-of-one-module)). It writes no `go.work` for them. A
program in the module's own folder cannot be a project: adopt says so and asks you to move it to `cmd/<name>`.

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
  wrote tools/smith/install.sh
  wrote .claude/agents/smith.md
  wrote .mcp.json
  wrote .vscode/mcp.json
  wrote go.work: ./services/api
  not committed: git status lists the changes and git can take them back; commit them when you are happy, or run smith adopt --commit next time
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
