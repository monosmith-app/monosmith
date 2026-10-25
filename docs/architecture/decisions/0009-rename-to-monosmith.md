# ADR 0009: The rename to MonoSmith

- Status: accepted
- Date: 2026-10-16

## Context

The project outgrew the MoNecromanCI name, and its first release was a chance to start
clean.

## Options considered

Keeping the old name: not recorded as considered.

## Decision

The product is MonoSmith, its command is `smith`, the repository is
`monosmith-app/monosmith`, the packages are named `monosmith` (`@monosmith/cli` too on
npm), and the first release is `smith@1.0.0`. The old Nx-based tool keeps the name mnci,
and `smith convert` still reads its workspaces.

## Consequences

References to the old tool (`@mnci/*` packages, `mnci:slot` markers, the `mnci` block of
`nx.json`) stay where convert needs to recognise them.

## Sources

- Commit [a17bcfd](https://github.com/monosmith-app/monosmith/commit/a17bcfd)
- Commit 77b66b6 (2026-10-16, move `apps/mnci` to `apps/smith`)
- [Converting](../../getting-started/converting.md#the-old-mnci)
