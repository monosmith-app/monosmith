# Adopting an existing repository

A repository with npm workspaces, Go modules or Python projects but no `smith.toml`
can be brought under smith without moving a file. `smith adopt` reads the repository,
tells you what smith would see and what stands in the way, and with `--init` writes
the workspace files and commits them. Nothing about your build or lint setup changes
unless you ask for it.

An Nx, Turborepo or mnci workspace is for [`smith convert`](converting.md); `adopt`
recognises them, says so and stops.

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
✔ nothing blocks: smith adopt --init makes it a workspace (add --install to run npm install, --eslint to write the lint config), then says what smith doctor still reports

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
  wrote go.work: ./services/api
  committed: chore: adopt smith
✔ adopted. smith doctor still reports 2 thing(s) for you:
  ✖ node_modules is missing at the workspace root
    → run `npm install` at the workspace root
  ...
  next: follow the steps above, then run smith doctor
```

`--init` takes the workspace name from the directory, the npm scope from the root
`package.json` and the CI providers from the ones the repository already uses. It
runs `go work init` and `go work use` for the Go projects, adds the Node projects to
the root `package.json` `"workspaces"`, adds `.smith/` to your `.gitignore` and ends by
checking the workspace as `smith doctor` does, so anything left over is listed with its
first step. `--install` also runs `npm install`; `--eslint` writes the two-line root
`eslint.config.mjs` and adds `@monosmith/eslint-config` to the root devDependencies.

`--tags` fixes release history. When a project's existing release tags are named
differently from the way smith names them (`<project>@<version>`, for example
`sdk@0.1.0`; `<dir>/v<version>` for a Go package), its next release would start from
scratch. `--tags` creates the missing tags locally, on the same commits; pushing them is
yours. Both `--tags` and `--init` need a clean git tree.

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

Run `smith doctor` and follow what it reports, then `smith build`. The full list of
options is in [`smith adopt`](../manual/adopt.md); the tour of a workspace continues in
[Your first workspace](first-workspace.md#3-add-projects).
