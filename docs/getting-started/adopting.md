# Adopting an existing repository

A repository with npm workspaces, Go modules or Python projects but no `smith.toml`
can be brought under smith. `smith adopt` reads the repository, tells you what smith
would see and what stands in the way, and with `--init` writes the workspace files.
Nothing about your build or lint setup changes unless you ask for it. smith commits
nothing unless you add `--commit`. If the only project of the repository sits at its
root, smith asks where to move it (see [A project at the root](#a-project-at-the-root)).

An Nx, Turborepo or mnci workspace is for [`smith convert`](converting.md); `adopt`
recognises them and hands them over, so `smith adopt` is the one way in for any repository.

## 1. Ask what adopting involves

This changes nothing:

```sh
smith adopt --dry-run
```

```text
~/work/old would be workspace "old" (scope @old, CI github) with:
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
✓ nothing blocks: smith adopt --init makes it a workspace (add --install to run npm install, --eslint to write the lint config), then says what smith doctor still reports

Nothing changed (adopt only reports). To apply:
  → smith adopt --init
```

Without `--dry-run` the command behaves the same: with neither `--init` nor `--tags` it
only reports. `--json` prints the report for a script.

When something blocks, the report lists each blocker and warning with the step that
clears it, and `smith adopt` exits with status 1. Blockers include:

- a folder that is not a git repository (`this is not a git repository: smith versions
  from tags and finds files through git`; run `git init` and commit what you have);
- a folder that is already a smith workspace;
- an Nx or Turborepo workspace (use `smith convert`).

Uncommitted changes do not block the report, which notes them
(`uncommitted changes: commit them before --tags or --init`), but `--init` and `--tags`
refuse to run until you commit or stash them.

Warnings include a root `package.json` with runtime dependencies (in smith each project
declares what it uses) and a pipeline of yours at a path smith's pipelines use, which
`--init` keeps as `<file>.pre-smith`.

## 2. Take the steps

```sh
smith adopt --tags --init
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
✓ adopted. smith doctor still reports 2 thing(s) for you:
  × node_modules is missing at the workspace root
    → run `npm install` at the workspace root
  ...
  next: follow the steps above, then run smith doctor
```

In a terminal, `--init` first asks a few questions, each with its consequence and a safe
default: the CI provider, whether to add the agent files, what to do with a pipeline of
yours, and whether to set line endings. It prints the plan, with a diff of every file of
yours it edits, and asks `Go ahead? [y/N]`. Without a terminal it refuses and names the
flags that answer (`--ci`, `--agent`, `--ci-existing`, `--gitattributes`); `--yes` takes
each default.

`--init` leaves the new files in the work tree, so you can review them with `git status`.
Add `--commit` to commit them as one commit, `chore: adopt smith`, that `smith undo` can
revert. If you forgot, [`smith commit`](../manual/commit.md) does it afterwards.

`--init` takes the workspace name from the directory, the npm scope from the root
`package.json` and the CI providers from the ones the repository already uses. It
runs `go work init` and `go work use` for the Go projects, adds the Node projects to
the root `package.json` `"workspaces"`, adds `.smith/` to your `.gitignore`,
writes the MonoSmith agent files (`.claude/agents/smith.md`, `.mcp.json` and `.vscode/mcp.json`, see
[the agent guide](../guides/agent.md)) and ends by checking the workspace as `smith doctor` does, so anything left over is listed with its
first step. `--install` also runs `npm install`; `--eslint` writes the two-line root
`eslint.config.mjs` and adds `@monosmith/eslint-config` to the root devDependencies.

`--tags` fixes release history. When a project's existing release tags are named
differently from the way smith names them (`<project>@<version>`, for example
`sdk@0.1.0`; `<dir>/v<version>` for a Go package), its next release would start from
scratch. `--tags` creates the missing tags locally, on the same commits; pushing them is
yours. Both `--tags` and `--init` need a clean git tree.

### A project at the root

smith keeps one folder for each project (`apps/<name>`, `libs/<name>`, `packages/<name>`).
If the only project of the repository has its `package.json`, `go.mod`, `pyproject.toml`
or `Cargo.toml` at the root, `--init` asks where to move it. It never picks for you.
Without a terminal it moves nothing unless `--move <folder>` says where (`--move none`
keeps the project in place). The move is `git mv`; with `--commit` it is a commit of its
own, before the adoption's. `smith adopt --dry-run` lists the `git mv` commands.

The move splits what belongs to the workspace from what belongs to the project
([ADR 0090](../architecture/decisions/0090-adopt-splits-the-workspace-from-the-project.md)).
The `.code-workspace` file, README, LICENSE and tool configuration such as `.prettierrc`
stay at the root, and other documents go to `doc/`. For a Node project, the tools in
`devDependencies` (typescript, jest, prettier, `@types/*`) move to the root
`package.json`, the compiler options every project shares move to a root
`tsconfig.base.json` that the project's `tsconfig.json` extends. A project's own ESLint setup is a question: use the shared lint
config (`--eslint-setup shared`) or keep your own in the project (`--eslint-setup keep`). `dependencies` and `jest.config.js` stay with the
project. A development dependency smith does not know as a tool is asked about; `--hoist
<name>` answers for it. Other languages move their manifest whole for now. The
[manual](../manual/adopt.md#a-project-at-the-repository-root) has the details.

## 3. Choose how far to go

`--level conservative` is the default: smith keeps the repository's own build and lint
setup and adds itself around it. `--level full` makes smith the only tool; it exists for
Nx and Turborepo workspaces, where it is `smith convert --level full`. On a plain
repository, `smith adopt --level full` says so and stops; it is not built yet.

## In the terminal interface

In a terminal, a bare `smith` in a folder that holds files but no workspace opens the
adopt wizard, which does the same in screens and also covers conversion. See
[the adopt wizard](../manual/tui.md#the-adopt-wizard).

## Next

Run `smith doctor` and follow what it reports, then `smith build`. If a program the
workspace needs is missing, `smith tools` lists it and `smith tools install` offers to
install it. The full list of
options is in [`smith adopt`](../manual/adopt.md); the tour of a workspace continues in
[Your first workspace](first-workspace.md#3-add-projects).
