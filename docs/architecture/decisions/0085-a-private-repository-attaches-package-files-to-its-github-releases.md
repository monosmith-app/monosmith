# ADR 0085: A private repository attaches package files to its GitHub Releases

- Status: accepted
- Date: 2026-11-05

## Context

A private repository must not publish to npm, PyPI, NuGet, RubyGems or crates.io: a public registry shows the
package to everyone. smith had two answers. A private registry (`[registries.<name>]`) needs a feed to run and a
token for every consumer. `registry = "none"` (ADR 0024) only tags, which suits JitPack and Packagist, but it leaves
a npm, Python, NuGet, RubyGems or Rust package with no installable file at all.

A GitHub Release can hold files, and a private repository's releases are visible to whoever can read the repository.
The owner wants the package files there, one release for each package tag, and a way to install the set together.

## Options considered

- Rewrite each package's manifest to point at the sibling files (`file:` paths in `package.json`): not chosen. It
  changes the files that are published, and a dependency range then means two things.
- One release for the whole run that holds every file: not chosen. The tag, the notes and the changelog of a package
  belong to the package's own release, and `github_releases` already makes one for each tag.
- A new registry value, `github-release`, that packs and attaches: chosen.

## Decision

1. `[release] registry = "github-release"`, or `[projects.<name>] registry = "github-release"`, makes the release
   pack each package instead of publishing it, and attach the file to the package's GitHub Release. It needs
   `[release] github_releases = true` (a load error says so). `github-release` is reserved as a registry name and is
   not an install source. Version stamping, tags, changelog and the journal are as before.
2. The plugin protocol gains the `packs` capability and `PublishParams.packDir`. A plugin that declares it writes the
   package file and publishes nowhere: the npm tarball (`npm pack` of the staged package), the wheel and the sdist,
   the `.nupkg`, the `.gem` and the `.crate` (`cargo package --no-verify`, since a sibling released in the same run is
   on no registry). The host never calls `publish` on a plugin without the capability: it refuses the release
   (preflight) because the plugin would publish to the public registry. Go, PHP, JVM and Dart have no package file
   to attach: they keep their tag-based behaviour, and a Dart package is not dispatched to pub.dev.
3. Each package file gets a `<file>.sha256` file, as the other release assets do (`sha256sum -c` reads it). The
   checksum files are uploaded after all the package files, so a release that holds a package file without its
   checksum file is an attach that stopped. A re-run keeps an attached file that has the same bytes, replaces one that
   differs (a repacked wheel), and skips a release that is complete: "the registry already has the version" becomes
   "the release for the tag already has the asset".
4. Dependencies keep their normal ranges. Each GitHub Release of a run instead gets the same two bundle assets,
   `release-bundle.json` and `INSTALL.md`, once every package of the run is attached. They list each package with
   its files, checksums and tag, the `gh release download` commands, and one install command per ecosystem
   (`npm i ./core-0.5.2.tgz ./cli-0.2.4.tgz`). The bundle is on every package's release, not on one release: a
   consumer who opens any package finds the whole set, and no release is special. A re-run merges with the bundle a
   release holds, so finishing a few packages never drops the others. `INSTALL.md` says a private repository needs
   `gh auth login` or a token that reads its contents.
5. Nothing asks a registry for a project in a non-publishing mode (`none` or `github-release`): the lineage check
   and `smith doctor` skip it (the tags and the release assets are the record), the registry probes and the
   rehearsal skip it, the generated workflow lists no registry secret for it (it needs `GITHUB_TOKEN` with
   `contents: write`), and an app's npm, PyPI, NuGet and RubyGems distributions are not published either.
6. `smith release --rehearse` checks, with a read-only request, that the token may write to the repository.
   `smith doctor` reports the packages that attach, and warns when another package would publish to a public registry
   with a credential that is not set. The token comes from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`, and is never
   printed.

## Consequences

A private repository releases packages with no registry and no registry secret. Consumers download and install from
the release, so they need read access to the repository. The existing confirmation flow still presents the plan
before anything is attached. SBOM and attestation are unchanged, since they belong to apps with a `[release]` table.
A package whose plugin lacks the capability, or an old plugin, stops the release before the first attach.

## Sources

- [Releasing](../../reference/releasing.md), [Configuration](../../reference/configuration.md),
  [Release files of a private repository](../../guides/private-release-assets.md)
- ADR 0024, ADR 0083
