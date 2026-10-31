# ADR 0043: A release is rehearsed against its registries with read-only requests

- Status: accepted
- Date: 2026-10-31

## Context

A first CI run on a new feed often fails on something a person could have seen earlier: a wrong
URL, a feed that wants Basic where the pipeline sends a Bearer token, a missing variable, a
pipeline identity without the right to publish. Each cost a fix-and-retry commit. The existing
pre-publish check ([releasing](../../reference/releasing.md#checked-before-anything-is-published))
covers the binaries of an app, not the packages that language plugins publish, and a dry run does
not ask any registry. Issue [#824](https://github.com/monosmith-app/monosmith/issues/824), part (a), of the smith 2.0
epic ([#819](https://github.com/monosmith-app/monosmith/issues/819)) asks for a rehearsal.

## Options considered

- **A new plugin method, `rehearse`, so each plugin checks its own registry.** Rejected for now: it
  changes the protocol schema that other 2.0 work also changes, and the plugins already tell the
  host what it needs: the registry in `smith.toml`, the variables of the `ci` method, the
  language's usual configuration (`.npmrc`, `TWINE_REPOSITORY_URL`, `NUGET_SOURCE`).
- **Run the tools' own dry runs (`npm publish --dry-run`, `cargo publish --dry-run`).** Rejected: they
  need built packages and most do not contact the registry or check the credential.
- **Fold the checks into the real release.** Rejected as the only place: the point is to learn
  before the first push. The real release keeps its own check.

## Decision

`smith release --rehearse` (a flag, not a new command) lists the package projects whose plugin
publishes to npm, PyPI, NuGet or crates.io, and for each target asks four things: the URL resolves
and which credential the feed challenges for (`WWW-Authenticate`), the credential is set, the
registry accepts it, and in GitHub Actions or Azure Pipelines which identity publishes. The code is
the `releaserehearsal` slice; the host maps a project's plugin name to an ecosystem and reads the
registry from the settings it already holds. No protocol type changes.

Requests are reads. The one exception is PyPI, which has no read that checks an upload credential: an
upload with no file, which the index refuses after the credential check. Nothing can publish. A
credential is never printed or encoded in JSON. What cannot be known without publishing (OIDC
trusted-publisher registration, a NuGet push key on nuget.org, an Azure feed role) is said as a
warning or a named setting, not claimed as checked.

## Consequences

- A failure exits 1 and prints the reason and the exact fix, in the style of `smith doctor`;
  `--json` prints the same as data.
- Only the four ecosystems are rehearsed. Maven, RubyGems, pub.dev and Packagist are skipped;
  adding one is a case in `ListTargets` and in the access check.
- The mapping from plugin name to ecosystem lives in the host. A plugin that publishes elsewhere
  needs an optional method later, which would be additive.
- Parts (b) (public export removed on a `fix:` commit) and (c) (plans and `smith undo`) of #824 are
  separate.

## Sources

[#824](https://github.com/monosmith-app/monosmith/issues/824),
[#819](https://github.com/monosmith-app/monosmith/issues/819).
