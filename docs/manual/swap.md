# smith swap

`smith swap` changes one decision of an existing project, such as its test runner, its linter or its
style framework. The decisions are the choices that [`smith choices`](choices.md) lists, one per
axis. The command checks the swap with the same choice solver that `smith add --with` uses, and
shows what leaves, what joins and what the project has afterwards.

smith moves the dependencies, scripts, config files and test imports of a project where the move is
mechanical. Today that is the test runner, Jest and Vitest. For the other axes it stops after the plan
and names what to change by hand. See
[ADR 0065](../architecture/decisions/0065-smith-swap-changes-one-choice-of-a-project.md).

## What a test runner swap does

1. Checks the swap with the choice solver and prints the plan.
2. Removes what the old runner wrote, only where the project still has it as written: the root dev
   dependencies (kept when another project still uses the runner), the `test` script, the `jest` field
   and the config files. Anything you edited stays and is listed as left to do by hand.
3. Has the new runner write its own dependencies, script and config, where they are absent.
4. Moves the test imports of `describe`, `it`, `expect` and the hooks to the new module. A file that
   uses `jest.fn`, `jest.mock` or `vi.*` stays, and is listed.
5. Records the new stack in `package.json`, installs the packages once, and runs the build and tests
   of the project.

`--dry-run` stops after step 1 and lists what the other steps would do. Without `--yes` the command
asks before it writes. `--commit` commits the swap as one operation that [`smith undo`](undo.md)
reverts.

## When to use it

- You want to know if a project can change its test runner, and what else would change with it.
- You move a project from Jest to Vitest, or back, without editing the dependencies and config by hand.

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
| `project` | yes | the project whose decision changes |
| `axis` | yes | the question to change: `test`, `lint` or `style` (`smith choices` lists the axes) |
| `choice` | yes | the answer that replaces the project's choice on the axis |

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
- An axis that smith cannot migrate by itself (style, for now): the message says to change the files by
  hand, then run [`smith doctor`](doctor.md).
- A run without `--yes` that is not answered with `y`: nothing changes.

If the build or the tests fail after the files changed, the files stay changed. `smith undo` reverts a
committed swap, and `git restore` reverts an uncommitted one.
