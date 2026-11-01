# smith swap

`smith swap` changes one decision of an existing project, such as its test runner, its linter or its
style framework. The decisions are the choices that [`smith choices`](choices.md) lists, one per
axis. The command checks the swap with the same choice solver that `smith add --with` uses, and
shows what leaves, what joins and what the project has afterwards.

smith does not move the dependencies, config files and imports of a project yet. No plugin offers
that migration, and a swap that rewrote only the stack record would leave the project half changed.
So `smith swap` stops after the plan and names what to change by hand. See
[ADR 0065](../architecture/decisions/0065-smith-swap-changes-one-choice-of-a-project.md).

## When to use it

- You want to know if a project can change its test runner, and what else would change with it.
- You plan the move before you do it by hand.

## Synopsis

```sh
smith swap <project> <axis> <choice> [--dry-run] [--yes] [--commit]
```

## Arguments and flags

<!-- smith:begin swap -->
Change one decision of a project, such as its test runner or linter

Usage: `smith swap <project> <axis> <choice> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes |  |
| `axis` | yes |  |
| `choice` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | show the plan and change nothing |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end swap -->

## Examples

```sh
smith swap web test vitest --dry-run   # show the plan for web and change nothing
smith swap api lint oxlint
```

## What it refuses

- A project that has no choice on the axis: the message lists the choices the project has.
- A choice that does not exist, or that is on another axis: the message lists the choices on the axis.
- A choice the project already has.
- A choice that does not fit the other choices of the project: the message is the solver's.
- Any swap that would need files to move: the message says to change them by hand, then run
  [`smith doctor`](doctor.md).
