# ADR 0010: One binary, re-packed for every package manager

- Status: accepted
- Date: 2026-10-16 to 2026-10-17

## Context

People install tools with the package manager of their language. Each wrapper that
downloads a binary at install time breaks offline and behind proxies.

## Options considered

- A wrapper per ecosystem that downloads the binary at install time: rejected for the
  reason above.
- One static binary per OS and architecture, re-packed for each channel: chosen.

## Decision

smith is one static binary per OS/arch (`CGO_ENABLED=0`, six targets). The release builds
the archives once, and re-packs the same binaries for each channel, with nothing
downloaded after install:

- npm: per-platform packages plus wrappers, with provenance
- PyPI: platform wheels with the binary in their scripts
- NuGet: a .NET 10 RID-specific tool
- RubyGems: platform gems with a Ruby launcher
- pub.dev: one package holding every platform's binary
- Linux: .deb and .rpm attached to the GitHub Release

## Consequences

pub.dev 1.0.0 was published by hand; its automation is
[#216](https://github.com/monosmith-app/monosmith/issues/216). Each registry has rules
that only showed up on the first real upload, and are now in the code:

- PyPI rejects wheels whose zip entries use data descriptors, so entries are written with
  their sizes up front ([#233](https://github.com/monosmith-app/monosmith/issues/233)).
- nuget.org needs a client version of 4.1.0 or later in `X-NuGet-Client-Version`
  ([#245](https://github.com/monosmith-app/monosmith/issues/245)), and a license
  expression needs `licenseUrl` too ([#247](https://github.com/monosmith-app/monosmith/issues/247)).
  The nuspec is written as `dotnet pack` writes it.

## Sources

- Issues [#121](https://github.com/monosmith-app/monosmith/issues/121),
  [#122](https://github.com/monosmith-app/monosmith/issues/122),
  [#217](https://github.com/monosmith-app/monosmith/issues/217),
  [#238](https://github.com/monosmith-app/monosmith/issues/238),
  [#240](https://github.com/monosmith-app/monosmith/issues/240),
  [#216](https://github.com/monosmith-app/monosmith/issues/216)
- [Installing](../../getting-started/installing.md), [Releasing](../../reference/releasing.md)
