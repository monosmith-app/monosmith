# ADR 0035: `[ci] branches` sets where the pipelines run

- Status: accepted
- Date: 2026-10-25

## Context

A user working on a long-lived `dev` branch found that the generated GitHub workflow ran only
on pushes to `main` and on pull requests into `main`, so pull requests into `dev` got no CI.
Nothing in `smith.toml` changed it.

## Options considered

The issue proposed a `[ci] branches` list for the triggers and the release dry-run comment,
with `smith ci release` staying limited to the release branch under its own setting.

## Decision

`[ci] branches = [...]` sets the push and pull request triggers for the GitHub and Azure
pipelines. The release branch is always included and stays the only branch a release runs
from. The default output is unchanged.

## Consequences

`smith doctor` and `smith ci generate --check` see a pipeline that does not match the setting
as drift. GitLab is not covered by the setting as shipped (the pull request names GitHub and
Azure).

## Sources

- Issue [#550](https://github.com/monosmith-app/monosmith/issues/550); pull request
  [#552](https://github.com/monosmith-app/monosmith/pull/552)
- [CI pipelines](../../reference/ci.md)
