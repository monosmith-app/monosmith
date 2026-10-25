# ADR 0021: Plugin roles and per-project stacks

- Status: accepted (refined by [ADR 0023](0023-recipes-with-levels-and-slots.md))
- Date: 2026-10-16

## Context

One monorepo should be able to hold a React app that uses a plugin bringing only Vite and
React, next to an app that uses an opinionated stack with a scaffold, lint rules, a test
library, CI steps and a deploy. Each project decides which plugins apply to it; there is no
single workspace-wide choice. The earlier workspace-wide conventions
([#161](https://github.com/monosmith-app/monosmith/issues/161)) were closed unmerged and
this record supersedes them.

## Options considered

- Workspace-wide conventions (issue #161): closed unmerged.
- Plugins declare roles; a project picks a stack, a named bundle of plugins: chosen.

## Decision

A plugin declares the roles it fills in `initialize`, the way it declares capabilities. The
roles proposed in the issue were language (exactly one per project), toolchain (one per job,
for example one test runner), structure (any number), scaffold (one) and deploy or CI (any
number). Any role may add doctor checks. A stack is a named bundle of plugins, declared in
`smith.toml` or published by a plugin author. A project records its stack, and
`smith add <kind> <name> --stack <stack>` picks it. `default_stacks` in `smith.toml` names
the stack of projects that record none.

## Consequences

`smith stacks` lists the stacks and who uses each. The role model was extended into levels
and slots two days later (ADR 0023).

## Sources

- Issue [#163](https://github.com/monosmith-app/monosmith/issues/163) (closed 2026-10-16)
- [Stacks](../../reference/stacks.md), [Plugin system](../plugin-system.md)
