<!-- lore-master
parent: deps.md
-->
# smith deps up

`smith deps up` is `smith up`: it shows the dependencies with newer releases and updates the
ones you pick, in every project that declares them. The full description, the exit codes and
examples are on [`smith up`](up.md); the two commands share one implementation and one set of
flags.

## Major updates are a choice of their own

`--yes` takes patch and minor updates only. A major update can break the project, so smith
leaves it out, lists it ("name 1.2.0 -> 3.0.0") and says that `--major` takes it. A package you
name that has only a major update asks in a terminal and needs `--major` without one. With
`--align`, the plan and the questions are those of `smith sync`.

## Synopsis

```sh
smith deps up [package]... [--yes] [--major] [--check] [--json] [--no-install] [--align [--dry-run] [--prefer highest|lowest|installed] [--to <name>=<version>]]
```

## Arguments and flags

<!-- smith:begin deps up -->
Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)

Usage: `smith deps up [packages...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `packages...` | no | packages to update; none lists the updates |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--align` | boolean | off | bring dependencies declared at different versions to one (the named ones, or all); with --check, fail when any differ |
| `--check` | boolean | off | report only; never prompt, never write |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | with --align: show the plan and change nothing |
| `--json` | boolean | off | print the updates and tools as JSON; never write |
| `--major` | boolean | off | also take major updates (with --yes or named packages), and with --align let a declaration move to another major version |
| `--no-install` | boolean | off | rewrite the manifests without installing |
| `--prefer <prefer>` | string |  the installed one, else the highest | with --align: which version wins when projects disagree: highest, lowest or installed (default: the installed one, else the highest) |
| `--to <to>` | string, repeatable |  | with --align: the version of one dependency, as <name>=<version> |
| `-y, --yes` | boolean | off | take every patch and minor update without prompting; major updates stay out unless --major is given |

<!-- smith:end deps up -->

## Example

```sh
smith deps up --check
```

prints the same report as `smith up --check`; see [the example there](up.md#examples).

## Related

[`smith up`](up.md), [`smith sync`](sync.md).
