# ADR 0057: Try a change in a sandbox before it touches the tree

- Status: accepted
- Date: 2026-11-01

## Context

Idea D3 of the design review: `--dry-run` prints a plan; a dress rehearsal proves the result works.
smith already has plans (ADR 0047) and undo, but undo needs a commit and a plan is a promise, not a
proof.

## Decision

`smith try <command...>` runs one of the file-changing commands in a detached `git worktree` of
`HEAD` with the uncommitted changes laid over it, runs checks (`doctor` by default; `build`, `test`,
`lint`, `typecheck` by flag), prints the files the command changed, and offers to run the same
command for real. It never edits the real tree. It runs the child as this smith binary, with
`SMITH_CACHE_DIR` set to the real cache so cached tasks are shared.

Why a worktree and not a copy: it copies only tracked files, so `node_modules` and build output are
never copied. Why re-run the command and not copy the result back: the command is the contract the
person reviewed, and a copy-back would have to merge with edits made meanwhile.

## Cost

The copy has no installed dependencies, so `build` and `test` are meaningful only for workspaces that
build without an install (the default check is `doctor` for that reason). A command with a network
effect (`up`) can give a different result when applied. Neither is hidden: both are in the manual.
A try needs a git repository with a commit.

## Smallest first slice (done)

`try` for add, remove, up, upgrade, derive, kind, eject, sync and bootstrap. There is no `swap`
command in smith yet (idea B4), so `try swap` waits for it. Not done: installing dependencies in
the copy (`--install`), a TUI and extension surface (the action catalog lists the reason), a
comparison of doctor before and after.

## Sources

- `apps/smith/sandbox/`
