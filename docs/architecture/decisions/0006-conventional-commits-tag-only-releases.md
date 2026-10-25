# ADR 0006: Conventional commits and tag-only releases

- Status: accepted
- Date: 2026-10-15

## Context

Release commits back to the branch race with other merges and clutter history.

## Options considered

Committing the new version back to the branch is the alternative the context argues
against. Others are not recorded.

## Decision

Versions come from conventional commits. The model is tag-only: manifests are stamped
for the build and publish, then restored, and the tag is the record. Releases run in
dependency order and are resumable (a re-run skips what is tagged and what a registry
already has). `--dry-run` follows the same code path. A commit `version(<project>)[x.y.z]`
forces a version, the way `version(smith)[1.0.0]` started MonoSmith at 1.0.0. A released
dependency gives its dependents a patch, and no setting changes that (2026-10-16).

## Consequences

Every merge to `main` that touches a released project releases it. Commit subjects
matter, so commitlint guards them. Before 1.0 a `feat:` releases a minor version.
[ADR 0024](0024-release-without-publishing-and-bare-tags.md) adds two opt-outs from the
registry and the tag format.

## Sources

- [Releasing](../../reference/releasing.md),
  [Release and CI internals](../release-and-ci-internals.md)
