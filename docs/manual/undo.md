# smith undo

`smith undo` reverts the last operation that smith committed: a `smith add`, `remove`, `upgrade`,
`derive` or `sync`. It uses `git revert`, so the result is a new commit and the history stays as it
was.

Operations are committed only when you opt in. Pass `--commit` to the command, or set
`commit_operations = true` under `[workspace]` in `smith.toml`. Then the command, started in a git
work tree with no uncommitted change, makes one commit (`chore: add package billing`) and records
it in `.smith/operations.json`. A run in a dirty work tree makes no commit and says why
(`not committing this operation: the work tree has uncommitted changes`). If you did not opt in,
[`smith commit`](commit.md) makes the commit afterwards, and then `smith undo` can revert it.

## When to use it

- You added a package and want the workspace as it was before.
- A `smith upgrade` or `smith derive` changed files you did not expect.

## Synopsis

```sh
smith undo [--dry-run] [--yes]
```

## Arguments and flags

<!-- smith:begin undo -->
Revert the last operation smith committed (add, remove, upgrade, derive, sync)

Usage: `smith undo [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--clean` | boolean | off | also delete the generated folders the project left behind (node_modules, dist), which git ignores |
| `--dry-run` | boolean | off | show what would be reverted and change nothing |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end undo -->

## What it checks

`smith undo` prints what it will revert (an added file is removed, a changed file is restored, a
deleted file is created) and asks in a terminal. With no terminal, or in CI, it refuses unless you pass `--yes`. It
refuses, with the reason and the next step, when:

- no operation is recorded (the setting was off, or everything was undone);
- the commit is not in the history of the current branch (it was rewritten);
- a later commit changed one of its files: undo the later work first, or use `git revert`;
- one of its files has an uncommitted change.

`git revert` cannot remove files that git ignores. When the undone project left generated folders behind
(`node_modules`, `dist`, a venv), `smith undo` lists them under "Left behind". In a terminal it asks whether to
delete them; `--clean` deletes them without asking; with neither it only says what is left and names `--clean`. Each folder is listed with its ignored
entries in parentheses, as `apps/web/  (dist/, node_modules/)`; a folder with no entry to name is listed alone.
A folder that holds a tracked file, or a file of yours that git does not ignore, is never listed or deleted.

Operations undo one at a time, newest first. See [ADR 0047](../architecture/decisions/0047-mutating-commands-show-a-plan-and-can-be-undone.md).

## Related

- [`smith add`](add.md), [`smith remove`](remove.md), [`smith upgrade`](upgrade.md),
  [`smith derive`](derive.md): the commands that print a plan before they write.
