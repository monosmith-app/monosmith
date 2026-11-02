# smith commit

`smith commit` commits the changes that are in the work tree as one commit named after an operation
(`chore: add app web`), and records it so that [`smith undo`](undo.md) can revert it. It is what ticking
"Commit this change" before a `smith add` does, for a change that is already in the files.

smith never commits unasked. [`smith add`](add.md), `remove`, `upgrade`, `derive` and `swap` leave their files
uncommitted unless you pass `--commit` or set `commit_operations = true`. When you decide afterwards that you
want the commit, this command makes it. The terminal interface and the VS Code extension call it from
"Commit now", after a change.

## When to use it

- You ran `smith add` without `--commit`, looked at the files and want them in one undoable commit.
- You want `smith undo` to be able to revert a change that is already written.

## Synopsis

```sh
smith commit [operation words] [--only-operation | --all] [--dry-run] [--yes]
```

## Arguments and flags

<!-- smith:begin commit -->
Commit the changes of an operation as one commit that smith undo can revert

Usage: `smith commit [operation words] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `operation words` | no |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--all` | boolean | off | commit everything git shows as changed, your own changes included |
| `--dry-run` | boolean | off | list the files that would be committed and change nothing |
| `--only-operation` | boolean | off | commit only the files the last operation changed; the other changes stay in the work tree |
| `-y, --yes` | boolean | off | do not ask for the final confirmation (the choice of files still needs --only-operation or --all) |

<!-- smith:end commit -->

## What it does

1. It lists every changed file of the work tree (smith's own `.smith` folder is left out), each as created,
   changed or removed.
2. If your own changes are in the work tree, it asks which files (above). In a terminal it asks; with no terminal, or in CI, it refuses unless you pass `--yes`. `--dry-run` prints
   the list and commits nothing.
3. It makes one commit, `chore: <operation words>`, and writes it in `.smith/operations.json`.

The operation words name the commit: `smith commit add app web` makes `chore: add app web`. Without words the
commit is named `chore: save the changes`.

## Which files

smith writes down which files an operation changed, at the time of the operation, in
`.smith/uncommitted-operation.json` (the files that changed during it and not before it). That keeps your own
work in progress out of the commit. When the work tree holds other changes too, `smith commit` asks:

1. **only the N files this operation changed** (the default),
2. **everything git shows changed (M files)**, with the extra files listed, or
3. cancel.

With no terminal, or in CI, it refuses and names the answers: `--only-operation` or `--all`. `--yes` skips only
the final "Go ahead?" question, not this choice. With nothing else changed, or when smith has no record of the
operation, there is nothing to choose: it commits what changed. `--dry-run` lists both sets and commits nothing.
With `--only-operation`, a file you had already changed before the operation stays out, even if the operation
touched it too; use `--all` for that.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The changes are committed, or there was nothing to commit. |
| 1 | The folder is not in a git work tree, the repository has no commit yet, git refused the commit, or there is no terminal and no `--yes`. |

## Files and environment

Reads `git status`. Writes one commit and `.smith/operations.json`. The repository's hooks run, as they do for
any commit; if git refuses the commit, the files stay in the work tree, unstaged.

## Examples

```sh
smith add app web --lang node --yes
smith commit add app web --dry-run
smith commit add app web --only-operation --yes
```

## Common errors

| Message | Fix |
|---|---|
| `this workspace is not in a git work tree` | Run `git init`, then commit again. |
| `the repository has no commit yet` | Make the first commit yourself; smith commits the changes after it. |

## Related

- [`smith undo`](undo.md) reverts the commit this command makes.
- [`smith add`](add.md) and its `--commit` flag commit while the command runs.
