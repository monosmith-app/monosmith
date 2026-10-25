# ADR 0005: Generated CI pipelines

- Status: accepted
- Date: 2026-10-15

## Context

Rationale not recorded beyond the goal that CI and release are first-class and
everything is generated (the product north star in the [index](README.md)). The later
recipes design restates the problem: a workspace that builds and releases takes more
knowledge than it should ([Recipes and CI](../recipes-and-ci.md)).

## Options considered

Not recorded.

## Decision

Pipelines are generated from the workspace. Plugins contribute toolchains, setup and
audit steps and credentials. smith builds one provider-neutral model and renders it for
GitHub Actions and Azure Pipelines, later also GitLab CI. The YAML calls
`smith ci <phase>`, so a laptop runs what CI runs. Steps a team adds between
`smith:slot` markers survive regeneration.

## Consequences

`smith doctor` fails on a pipeline that no longer matches the workspace, and
`smith ci generate --check` reports drift. Later settings that change the output:
download-cache steps ([#346](https://github.com/monosmith-app/monosmith/issues/346)),
`[ci] branches` ([ADR 0035](0035-ci-branches-setting.md)).

## Sources

- [CI pipelines](../../reference/ci.md),
  [Release and CI internals](../release-and-ci-internals.md),
  [Recipes and CI](../recipes-and-ci.md)
