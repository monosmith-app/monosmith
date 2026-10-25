# ADR 0008: Containers through a docker plugin

- Status: accepted
- Date: 2026-10-16

## Context

The old mnci kept container images in a separate container project. Why that was
replaced is not recorded beyond the decision.

## Options considered

Not recorded.

## Decision

An app with a `Dockerfile` gets an `image` task from the `docker` plugin, and the release
pushes the image tagged with the version. This replaced the old mnci's separate
container project.

## Consequences

Later changes build on it: an image can be pushed to several registries, tagged with the
commit and branch, and have its Docker Hub description set without a release.
`smith convert` turns an app with a `Dockerfile` into the same app plus
`[conventions.docker]`.

## Sources

- Issues [#83](https://github.com/monosmith-app/monosmith/issues/83),
  [#213](https://github.com/monosmith-app/monosmith/issues/213)
- Pull requests [#322](https://github.com/monosmith-app/monosmith/pull/322),
  [#354](https://github.com/monosmith-app/monosmith/pull/354),
  [#371](https://github.com/monosmith-app/monosmith/pull/371)
- [Releasing](../../reference/releasing.md)
