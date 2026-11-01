# smith derive

`smith derive` writes the files that follow from facts and the project graph (a tsconfig's
`references`, a Jest `displayName`) and keeps a record of what it wrote in `smith.lock`. A file that
smith owns is rewritten when its inputs change. A file that someone edited by hand is never
overwritten: smith reports the drift with a diff and waits for your choice.

Derived files are opt-in. Set `enabled = true` under `[derive]` in `smith.toml`. Without it the
command stops with a message and nothing else in smith looks at derived files.

## When to use it

- A dependency between projects changed, and the `references` of the projects must follow.
- You converted a workspace and want to delete the config that repeats in every package.
- A pull request must fail when a derived file is out of date (`--check`; `smith ci verify` runs it).
- A derived file has a hand edit, and you must decide what to do with it.

## Synopsis

```sh
smith derive [--check] [--explain <file>] [--replace <file>] [--adopt <file>] [--eject <file>] [--yes]
```

## Arguments and flags

<!-- smith:begin derive -->
Write the files that follow from facts and the project graph, and report hand edits

Usage: `smith derive [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--adopt <adopt>` | string, repeatable |  | take the hand edit of this keys-mode file as the new baseline |
| `--check` | boolean | off | write nothing; fail when a derived file or the record is out of date |
| `--eject <eject>` | string, repeatable |  | own this file: smith stops touching it |
| `--explain <explain>` | string, repeatable |  | say who owns this file and why it is in its state |
| `--replace <replace>` | string, repeatable |  | overwrite this drifted or diverged file with the derived content |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end derive -->

Paths are relative to the workspace, with forward slashes. A flag can be repeated.

## Examples

```sh
smith derive                                    # write what is missing or stale, record the rest
smith derive --check                            # for CI: fail when anything is out of date
smith derive --explain packages/web/tsconfig.lib.json
smith derive --replace packages/web/tsconfig.lib.json   # the derived content wins
smith derive --adopt packages/web/tsconfig.lib.json     # keep the hand edit as the baseline
smith derive --eject packages/web/tsconfig.lib.json     # the file is yours from now on
```

## What it does

1. Asks each plugin that derives for its files. The request carries the internal dependencies of
   every project, taken from the project graph, and the paths `smith.lock` says smith owns.
2. Compares, for each file, what the renderer wants, what is on disk and what the record says.
3. Writes the files that are missing or stale, records the files that already match, and prints the
   files that need a choice with a diff.
4. Saves `smith.lock` when it changed.

Before step 3, derive prints the plan: the files it creates or rewrites, grouped by renderer, with
their owner, and `smith.lock`. In a terminal, outside CI, it asks `Go ahead? [y/N]` unless you pass
`--yes` or a decision (`--replace`, `--adopt`, `--eject`). With `commit_operations = true` under
`[workspace]` and a clean git work tree, the run is one commit (`chore: derive files`) that
[`smith undo`](undo.md) reverts.

Besides what the project graph says, a plugin can derive a config file from the shape of the sibling
projects: when at least three projects of one kind hold the same file, and two thirds of those that
have it agree after the project name and references are substituted, that shape is derived for each
project that has the file. A key or file that differs is the project's own: smith leaves it alone, and
`--explain <file>` and `smith doctor` say that it diverges from its siblings (that is not drift and
needs no choice). The rule is in [Derived files](../reference/derived-files.md).

The states and the three ownership modes are in [Derived files](../reference/derived-files.md).

## Common errors

| Message | Cause and fix |
|---|---|
| `derived files are off in this workspace` | Set `enabled = true` under `[derive]` in `smith.toml`. |
| `N derived files need a choice` | A file is in `drift`, `diverged` or `broken`. Read the diff, then use `--replace`, `--adopt` or `--eject` for that file. |
| `N derived files are out of date` | Only from `--check`. Run `smith derive` and commit the files and `smith.lock`. |
| `<file> is a <mode> file: adopt applies to files that smith owns by keys` | `--adopt` works for a `keys` file that still has every derived key. Use `--replace` or `--eject`. |
| `the file has no smith:begin and smith:end markers` | Add the two marker lines around the lines smith owns, or `--eject` the file. |
| `smith does not derive <file>` | No renderer proposes this path. Run `smith derive --check` to list the files smith derives. |

## See also

[Derived files guide](../guides/derived-files.md), [Derived files reference](../reference/derived-files.md),
[`smith doctor`](doctor.md), [`smith ci verify`](ci.md), [`smith eject`](eject.md) (copies a tool's
default configuration: a different job).
