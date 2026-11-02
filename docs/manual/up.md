# smith up

`smith up` shows every declared dependency that has a newer release, grouped patch, minor
and major, with the projects declaring it, then updates the ones you pick. Each declaration
keeps its operator and moves to the latest version, every project that declares the package
is updated in the same change, and the workspace is installed again so lockfiles and the
venv follow. A last section compares the tools smith pins (Node.js, Python, Go,
golangci-lint) with their newest releases.

`smith deps up` is the same command.

## When to use it

- Routine dependency updates across every language at once.
- `smith up --check` in a scheduled job to learn what is outdated.
- `smith up --align` to bring one dependency that projects declare at different versions to
  one (this updates nothing; it is [`smith sync`](sync.md) for the named packages).

## Major updates are a choice of their own

`--yes` takes patch and minor updates only. A major update can break the project, so smith
leaves it out, lists it ("name 1.2.0 -> 3.0.0") and says that `--major` takes it. A package you
name that has only a major update asks in a terminal and needs `--major` without one. With
`--align`, the plan and the questions are those of `smith sync`.

## Synopsis

```sh
smith up [package]... [--yes] [--major] [--check] [--json] [--no-install] [--align [--dry-run] [--prefer highest|lowest|installed] [--to <name>=<version>]]
```

## Arguments and flags

<!-- smith:begin up -->
Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)

Usage: `smith up [packages...] [flags]`

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

<!-- smith:end up -->

Behaviour:

- With names, those packages are updated without a prompt. `--yes` takes every update.
- Without a terminal, or with `--check`, `up` only reports; `--json` also never writes.
- With `--align` nothing is updated: the named packages (or, with none, every one) that
  projects declare at different versions are brought to one, leaving those on the
  `[dependencies] allow` list. With `--check` it fails when any differ.
- Tools are informational: `up` never changes the pinned toolchain versions.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The report was printed or the updates were applied. |
| 1 | A named package has no update (`smith: no update for zod`), `--align --check` found differences, or an install failed. |

## Files and environment

Queries the registries (network), rewrites manifests, then installs. `[workspace] registry`
selects the registry. Reads `[dependencies]` in `smith.toml`.

## Examples

Report only, from a script (the same text is printed by a terminal-less `smith up`):

```sh
smith up --check
```

```text
patch
  @monosmith/eslint-config     1.0.0 → 1.0.2  (root)
minor
  eslint                       10.12.0 → 10.13.0  (root)
major
  @types/node                  24.19.2 → 26.6.5  (root)
  typescript                   6.0.3 → 7.0.2  (root)

Tools smith pins (informational):
  Go               1.27 ✔ (latest 1.27.3)
  golangci-lint    2.14.0 → 2.15.0 available
  Node.js (LTS)    24 ✔ (latest 24.22.1)
  Python           3.12 → 3.15 available

Report only: pick with smith up in a terminal, name packages, or pass --yes.
```

The JSON form has `updates` (each with `plugin`, `name`, `current`, `latest`, `level`
and the `sites` it would rewrite: `file`, `spec`, `section`, `rewritable`) and `tools`
(`pinned`, `latest`, `newer`).

Take everything, or exactly some packages, in every project that uses them:

```sh
smith up --yes
smith up zod jest
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith: no update for zod` | The package is not declared, or already at the latest. |
| A registry error | The registry was unreachable; retry or set `[workspace] registry`. (Not captured.) |

## Related

[`smith sync`](sync.md), [`smith install`](install.md), [`smith deps up`](deps-up.md),
[`smith doctor`](doctor.md).
