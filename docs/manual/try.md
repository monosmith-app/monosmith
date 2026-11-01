# smith try

`smith try <command>` runs a command that changes files in a temporary copy of the work tree,
runs checks on the result, and tells you what changed. Your files are not touched. Then it offers
to run the same command for real. `smith add app web --lang go --dry-run` prints a plan;
`smith try add app web --lang go` proves that the result passes `smith doctor`.

## When to use it

- Before `smith add`, `up`, `upgrade` or `derive` in a workspace you care about.
- To see what a command would change, as a file list, without a commit and without `smith undo`.

## Synopsis

```sh
smith try [--checks doctor,build,test] [--keep] [--json] [--yes] <command> [args...]
```

Flags go before the command; everything after the command name belongs to the command.
The commands it accepts: add, remove, up, upgrade, derive, kind, eject, sync, bootstrap.

## Arguments and flags

<!-- smith:begin try -->
Try a command that changes files in a temporary copy, check the result, then offer to run it for real

Usage: `smith try <command> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `command` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--checks <checks>` | string, repeatable |  | checks to run in the copy after the command: doctor, build, test, lint, typecheck |
| `--json` | boolean | off | print the result as JSON and do not apply |
| `--keep` | boolean | off | leave the copy on disk and print its folder |
| `-y, --yes` | boolean | off | run the command for real when the try passed, without asking |

<!-- smith:end try -->

## What the copy is

A detached `git worktree` of `HEAD`, with your uncommitted changes laid over it. Folders git
ignores (`node_modules`, `dist`) are not copied, and the task cache is shared with your tree, so a
check that was cached is not run again. Dependencies are not installed in the copy: `doctor` needs
none, but `build` and `test` need them, so ask for those only when the workspace can build without
an install step (Go, Rust), or use `--keep` and install by hand. The copy is removed when the try
ends.

`doctor` reports problems that were there before the command too. Read its output against a plain
`smith doctor`.

## Applying

In a terminal smith shows the files the try changed and asks `Go ahead? [y/N]`. In CI or without a
terminal it never asks: it prints `smith <command>` for you to run. `--yes` runs the command when the
command and every check passed. A failed try never offers to apply. Applying runs the command again
in your tree, so a command whose result depends on the network can differ ([ADR 0057](../architecture/decisions/0057-try-a-change-in-a-sandbox-before-it-touches-the-tree.md)).
