# smith slices

`smith slices` checks the vertical-slice layout of projects: that files are named for the
role they play, and that no folder is a junk drawer. It has one subcommand today,
[`smith slices check`](slices-check.md). For Node projects the same rules are enforced by
the ESLint config instead; the conventions are explained in
[Conventions](../reference/conventions.md).

## Subcommands

<!-- smith:begin slices -->
Check the vertical-slice layout of projects

Usage: `smith slices <command>`

| Subcommand | Meaning |
|---|---|
| `smith slices check` | Check that files are role-named and no folder is a junk drawer |

No flags.

<!-- smith:end slices -->

## Exit codes

0 for the help; see the subcommand.

## Files and environment

None of its own.

## Related

[`smith slices check`](slices-check.md), [Conventions](../reference/conventions.md),
[`smith lint`](lint.md).
