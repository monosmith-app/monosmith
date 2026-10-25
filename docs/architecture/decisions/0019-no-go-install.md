# ADR 0019: smith is not installed with `go install`

- Status: accepted
- Date: 2026-10-18

## Context

smith is built from the repository's `go.work`: `apps/smith` imports `packages/plugin-sdk`
and fifteen `libs/plugin-*` modules that are never tagged as Go modules (only the SDK is).
`go install github.com/monosmith-app/monosmith/apps/smith@latest` ignores `go.work`, so it
cannot resolve them.

## Options considered

Supporting `go install` would mean tagging every plugin module on every release, and a
release that first commits `apps/smith/go.mod` pinning those tags before it tags smith
itself. Not chosen.

## Decision

`go install` is not a supported way to install smith. The supported ways are the install
script, npm, PyPI, NuGet, RubyGems, pub.dev, the `.deb` and `.rpm` packages and the GitHub
Release ([Installing](../../getting-started/installing.md)). Those carry a stamped,
reproducible binary; `go install` would build one with no version.

## Consequences

The plugin modules stay unreleased internal modules and a release stays one tag. Someone who
wants to build from source clones the repository and builds inside the workspace
(`apps/smith/README.md`). Revisit if the plugins ever ship as separately versioned Go
modules.

## Sources

- Issue [#267](https://github.com/monosmith-app/monosmith/issues/267); pull request
  [#368](https://github.com/monosmith-app/monosmith/pull/368)
- [ADR 0010](0010-one-binary-for-every-package-manager.md)
