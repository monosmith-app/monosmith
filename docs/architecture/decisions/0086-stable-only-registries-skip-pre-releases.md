# ADR 0086: A registry can be kept to stable versions with `stable_only`

- Status: accepted
- Date: 2026-11-05

## Context

An app is published as one package per platform to npm, PyPI, NuGet and RubyGems. PyPI limits a project to 10 GB.
After about 46 betas of smith, the upload of `monosmith 2.0.0b46` failed with `Project size too large`, and the
release stopped before its tag. A registry that never lets a version be removed fills up for good, so skipping
the upload at failure time is not enough: the pre-releases must not go there at all. The owner decided that
pre-releases go to npm only (the `next` channel) and PyPI receives stable releases only.

## Options considered

- A hidden rule (never upload a pre-release to PyPI): not chosen. A workspace that wants its betas on PyPI would
  lose them with no way to turn it back on, and the owner wants decisions to be shown.
- A workspace-wide `[release] prerelease_registries = [...]`: not chosen. The registry names of an app live in
  `[projects.<app>.release]`, and two apps in one workspace may differ (one on PyPI under a quota, one not).
- A boolean per registry (`pypi_stable_only = true`): not chosen. It adds three keys for one idea.
- One list in the table that names the registries, `stable_only = ["pypi"]`: chosen.

## Decision

1. `[projects.<app>.release] stable_only = [...]` lists registries from `pypi`, `nuget` and `rubygems`. A
   pre-release version (a version with a `-`) is not published to them. Unset means every registry receives every
   version, as before. `npm` cannot be listed, since the channel's dist-tag is how a pre-release is installed.
   Each listed registry must be set in the same table (a load error says so).
2. The release prints `<registry> <name>: skipped: pre-release <version>`, records it in the report
   (`skippedPackages`) and in the journal (a `package` entry, outcome `skipped`), and does not ask that registry in
   the check before a release. `PYPI_TOKEN`, `NUGET_API_KEY` and `GEM_HOST_API_KEY` are not required for a
   pre-release they do not publish. The generated workflow keeps the secret, since it also runs stable releases.
3. `smith doctor` warns, only in a workspace with release channels, about an app that sends pre-releases to
   PyPI, NuGet or RubyGems with no `stable_only`, and says how to set it.

## Consequences

- The npm `next` dist-tag flow is unchanged: the skip happens in the distribution step, before the tag, and the
  last phase that moves the tags runs as for any release.
- smith never removes a version a registry holds. Pre-releases uploaded earlier stay, which is harmless: pip and
  other installers ignore pre-releases unless asked.
- smith has no evidence of a project size limit on NuGet or RubyGems, and has not checked their current limits;
  they stay as they are until a registry refuses, and `stable_only` is ready for them.
