# Please compared with MonoSmith

Please is a cross-language build system, in its own words "with an emphasis on high
performance, extensibility and reproducibility". It is written in Go and supports
Linux, macOS and FreeBSD.

## What Please is best at

- **Hash-based caching.** It caches artifacts by hashes of each rule's inputs, not
  file timestamps, and its documentation says shared remote caches make builds faster.
- **Cross-language builds**, with "a number of popular languages" (Go, Python and
  Java are mentioned in passing).
- **Open source.** Apache-2.0.

The pages read do not describe remote execution, affected selection, release or CI
generation, so none is claimed either way.

## Where MonoSmith differs

| | Please | MonoSmith |
|---|---|---|
| Platforms | Linux, macOS, FreeBSD | Linux, macOS, Windows (x64 and arm64) |
| What you write | BUILD files | `smith.toml`; the rest inferred |
| Release, CI, health check | not covered by the pages read | all three |

The Windows row compares the two tools' stated platforms: the Please page lists three
and does not list Windows. It does not say Windows is unsupported.

## When to choose Please instead

- You want a BUILD-file based, reproducibility-focused build system written in Go, and
  you run on Linux, macOS or FreeBSD.

## When MonoSmith is the better fit

- You work on Windows, or want a tool that infers projects from manifests and also
  generates CI and releases.

## Moving between Please and MonoSmith

No automated path in either direction, and neither was tried.

## Sources

Retrieved 2026-10-25.

- Repository (description, platforms, licence): <https://github.com/thought-machine/please>
