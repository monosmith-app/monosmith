# ADR 0070: Adopt asks where to move the project at the root

- Status: accepted
- Date: 2026-11-01

## Context

smith manages one folder per project. A plain repository with one project at its root (a
`package.json`, `go.mod`, `pyproject.toml` or `Cargo.toml` there) has no such folder, so
`smith adopt --init` found no project and made an empty workspace. The person had to move the files
by hand first. A tool that moves a whole repository must not guess where, so smith asks.

## Decision

- `smith adopt --init` detects one project at the root. If smith sees no project yet, it asks:
  "Move it into a folder of its own so smith can manage it?" with five answers: `apps/<name>`,
  `packages/<name>`, `libs/<name>`, another folder, or no. The suggested one is an app folder when
  the project looks like a program (a start script or `bin`; a `package main`; project scripts;
  `src/main.rs`), else a package folder. The name comes from the manifest (scope and module path
  dropped, kebab-case) and can be edited.
- smith never picks the folder. Without a terminal, or with `--yes`, `--init` moves nothing unless
  `--move <folder>` is given, and it stops with the exact commands. `--move none` keeps the project
  where it is (the empty workspace of before).
- The move is `git mv` of every top-level entry except a fixed list that belongs to the repository:
  version control, CI, editor and agent settings, smith's files, repository documents (README,
  LICENSE, CHANGELOG, `docs`), compose files, and for Node the lockfiles. The list is in
  `apps/smith/adoptmove`, not in the language plugins, because the protocol has no method that
  answers it and the list is the same for every ecosystem except the Node lockfiles. Leading-slash
  lines of `.gitignore` that name a moved entry are rewritten. Go module paths and file contents are
  not changed.
- It refuses a root with two manifests, a folder that already holds files, a folder whose first part
  is a file or a moving entry, and a manifest that is not tracked.
- The move is staged, not committed. With `--commit` it is a commit of its own
  (`chore: move <name> into <folder>`, undone with `git revert HEAD`) before the adoption commit
  that `smith undo` reverts.
- `--dry-run` prints the question and the `git mv` commands for the suggested folder, or for
  `--move`. The adopt wizard and the VS Code extension ask the same question and pass the answer as
  `--move`, with the commit box unticked. The JetBrains plugin has no adopt flow yet.

## Consequences

- A repository with one project at the root can be adopted in one run.
- The classification is a list: an unknown top-level name moves with the project. A repository file
  with an unusual name needs a `git mv` back, and the dry run shows it first.
- `package-lock.json` stays at the root; `npm install` rewrites it for the workspace layout.
- The move commit is made before `smith.toml` exists, so `smith undo` does not cover it.
