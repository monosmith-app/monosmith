# ADR 0018: Improvements are judged by how much easier smith is than its competitors

- Status: accepted
- Date: 2026-10-18

## Context

The product owner keeps research on Nx, Turborepo, Bazel, Moon and Pants in Google Drive.
Its purpose is to find where MonoSmith must improve to be an easier and better tool than
they are, not to copy their feature lists.

## Options considered

Matching competitors feature for feature is the alternative the context rules out. Others
are not recorded.

## Decision

When ideas from that research ([#341 to #353](https://github.com/monosmith-app/monosmith/issues/341))
or elsewhere compete for time, the first question is how much easier, safer or more
predictable they make setting up, running and releasing a polyglot monorepo for a team. A
feature a competitor has that does not answer that is not a priority (distributed task
execution, hermetic sandboxing, a Starlark-style configuration language).

## Consequences

Several records below came out of that research: [ADR 0027](0027-one-version-per-production-dependency.md),
[ADR 0028](0028-graph-edges-from-shared-schemas.md), [ADR 0020](0020-toolchain-pins-checked-not-downloaded.md)
and the SBOM and provenance work in [ADR 0011](0011-no-paid-code-signing.md). The product
comparison is in `docs/product`.

## Sources

- Issue [#336](https://github.com/monosmith-app/monosmith/issues/336); pull request
  [#356](https://github.com/monosmith-app/monosmith/pull/356)
