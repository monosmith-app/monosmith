# ADR 0024: Release a project without publishing it, and tag it with the bare version

- Status: accepted
- Date: 2026-10-18

## Context

A Java package built for a user was published through JitPack, which builds the artifact
from a git tag and uses the tag name as the version. smith only had to tag and create the
GitHub Release, but it always tried to publish to a registry and prefixed tags with the
project name.

## Options considered

- An empty `tag_prefix = ""` for bare tags, as the issue proposed: not chosen, because an
  empty string cannot be told from "unset" without changing the configuration type that
  convert and its tests use throughout.
- A flag, `bare_tags = true`: chosen.

## Decision

`registry = "none"`, under `[release]` or `[projects.<name>]`, versions the project, tags it
and creates its GitHub Release, and publishes to no registry. `Project.Publishes()` in the
plugin SDK is what the plugins that list publish credentials ask, so none lists a secret for
such a project. `none` is reserved as a registry name and refused for `[workspace] registry`
(installs always come from somewhere). `bare_tags = true` on one project makes its tags the
bare version (`0.1.0`); it excludes `tag_prefix`, and only one project can use it. `tag_prefix
= "v"` still works for `v0.1.0`.

## Consequences

The release summary says "published by its tag". An app that builds release assets still
does. Related: a tag prefix counts as deliberate only when it comes from `smith.toml`
([#361](https://github.com/monosmith-app/monosmith/pull/361)).

## Sources

- Pull request [#358](https://github.com/monosmith-app/monosmith/pull/358) (closes issue
  [#357](https://github.com/monosmith-app/monosmith/issues/357)), commit `3811c19`
- [Releasing](../../reference/releasing.md)
