# smith undo

`smith undo` reverts the last operation that smith committed: a `smith add`, `remove`, `upgrade`,
`derive` or `sync`. It uses `git revert`, so the result is a new commit and the history stays as it
was.

Operations are committed only when you opt in. Set `commit_operations = true` under `[workspace]`
in `smith.toml`. Then each of those commands, started in a git work tree with no uncommitted
change, makes one commit (`chore: add package billing`) and records it in `.smith/operations.json`.
A run in a dirty work tree makes no commit and says why.

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
| `--dry-run` | boolean | off | show what would be reverted and change nothing |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end undo -->

## What it checks

`smith undo` prints what it will revert (an added file is removed, a changed file is restored, a
deleted file is created) and asks in a terminal. With no terminal, or in CI, it never asks. It
refuses, with the reason and the next step, when:

- no operation is recorded (the setting was off, or everything was undone);
- the commit is not in the history of the current branch (it was rewritten);
- a later commit changed one of its files: undo the later work first, or use `git revert`;
- one of its files has an uncommitted change.

Operations undo one at a time, newest first. See [ADR 0047](../architecture/decisions/0047-mutating-commands-show-a-plan-and-can-be-undone.md).

## Related

- [`smith add`](add.md), [`smith remove`](remove.md), [`smith upgrade`](upgrade.md),
  [`smith derive`](derive.md): the commands that print a plan before they write.
