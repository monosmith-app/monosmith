# ADR 0034: Adopt and convert have depth levels, conservative and full

- Status: accepted for Nx and Turborepo workspaces; proposed for plain repositories
- Date: 2026-10-25

## Context

`smith convert` and `smith adopt` left some things for the user to fix. The owner's idea
(issue [#548](https://github.com/monosmith-app/monosmith/issues/548), 2026-10-25) was to
add a choice of depth.

## Options considered

Where it lives was left as an open decision in the issue, with `--level` on both commands as
the proposed default. Starting with Node and adding other languages later was proposed. A
removal rule was proposed: remove unused packages only when no remaining file imports them.

## Decision

- Conservative (the default): keep what works, map what can be mapped, list the rest under
  things to check.
- Full: make smith the only tool. After converting, work out which files and `devDependencies`
  of the old tooling nothing that stays imports or runs, and delete them after asking about
  each group.
- Four choices, as flags or numbered questions: `--lint smith|keep`,
  `--test jest|vitest|keep`, `--build smith|keep`, `--ci smith|keep`.
- The plan is made on a scratch copy of the converted workspace, so nothing changes before the
  questions are answered; each group is planned again from the files as they are after the
  previous answer. A clean git tree is required so that git is the way back.
- Without a terminal the plan is printed and the command exits non-zero, naming the flags.
- The terminal interface's adopt wizard drives the same plan
  ([#554](https://github.com/monosmith-app/monosmith/pull/554)).

## Consequences

`smith adopt --level full` on a plain repository (npm workspaces, Go modules, Python
projects) says it is not built, because the removal plan assumes the conversion has already
put smith's build and ESLint config in place. That is tracked in
[#553](https://github.com/monosmith-app/monosmith/issues/553), still open. Other languages'
leftovers are not looked for yet.

## Sources

- Issues [#548](https://github.com/monosmith-app/monosmith/issues/548),
  [#553](https://github.com/monosmith-app/monosmith/issues/553); pull requests
  [#549](https://github.com/monosmith-app/monosmith/pull/549),
  [#554](https://github.com/monosmith-app/monosmith/pull/554)
- [Adopt and convert](../adopt-and-convert.md), [`smith convert`](../../manual/convert.md),
  [`smith adopt`](../../manual/adopt.md)
