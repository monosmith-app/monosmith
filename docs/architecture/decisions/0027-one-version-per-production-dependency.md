# ADR 0027: One version per production dependency

- Status: accepted
- Date: 2026-10-20

## Context

Projects of one ecosystem can declare the same dependency at different versions. The
research behind [ADR 0018](0018-judge-improvements-by-ease.md) (single-version policy
against per-project versioning) and the recipes design
([Recipes and CI](../recipes-and-ci.md), "Dependencies across the workspace") proposed one
version across the workspace, with dev tools at the root once and production dependencies in
each project. The design left open whether the policy is on by default, and recommended on,
with a per-dependency exception list.

## Options considered

On by default or off by default; the recommendation was on. What shipped: the setting
defaults to off, and `smith new` writes it as true.

## Decision

`[dependencies] one_version` (default off; `smith new` writes `true`) with `allow`
exceptions. `smith doctor` fails on a production dependency declared at different versions
by projects of one ecosystem, and warns on a dev tool declared outside the root.
`smith deps up [names]` (the same as `smith up`) makes one change across every declaring
project, and `--align` brings fragmented dependencies to one version without updating them.
`allow` is honoured by `smith sync` and by `--align`.

## Consequences

A recipe cannot carry the setting yet, because recipes have no workspace scope.

## Sources

- Issue [#350](https://github.com/monosmith-app/monosmith/issues/350); pull request
  [#386](https://github.com/monosmith-app/monosmith/pull/386)
- [`smith deps up`](../../manual/deps-up.md), [Configuration](../../reference/configuration.md)
