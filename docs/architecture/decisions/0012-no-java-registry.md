# ADR 0012: No Java registry; OS package managers later

- Status: accepted
- Date: 2026-10-16

## Context

smith is published to the registries of several languages ([ADR 0010](0010-one-binary-for-every-package-manager.md)).
Java has no CLI registry worth an approval request: Maven Central and SDKMAN! both
require one.

## Options considered

- Publish to Maven Central or SDKMAN!: not chosen, because both require an approval
  request.
- Defer OS package managers and further registries: chosen.

## Decision

- Java: smith is not published to a Java registry. The jvm plugin publishes a workspace's
  own libraries to a Maven repository it is given.
- OS package managers (Homebrew tap, Scoop bucket, winget, Chocolatey, AUR, Snap, apt/dnf
  repositories) and crates.io and Packagist are deferred to
  [#225](https://github.com/monosmith-app/monosmith/issues/225).

## Consequences

The Homebrew, Scoop and winget manifests are already attached to each GitHub Release, but no
tap, bucket or winget-pkgs submission uses them yet.

## Sources

- Issue [#225](https://github.com/monosmith-app/monosmith/issues/225)
- [Installing](../../getting-started/installing.md#not-yet)
