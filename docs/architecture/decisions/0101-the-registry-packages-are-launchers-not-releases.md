# ADR 0101: The registry packages are launchers, not releases

- Status: accepted
- Date: 2026-11-15
- Amends: [ADR 0055](0055-a-committed-launcher-runs-the-pinned-smith.md) (the pin moves to `[workspace] smith`)

## Context

Every smith release published a package for each platform plus a wrapper on npm, PyPI and NuGet,
and `smith update` ran the original installer again. A release therefore depended on every
registry publish succeeding, and an install from a registry was a release of its own. smith now
updates itself, so the registry packages need not carry smith at all. Bazelisk, rustup and Volta
show the pattern: a small launcher that rarely changes picks a version and runs it.

Owner decisions of 2026-11-15:

- The workspace pins its smith in `smith.toml`: `[workspace] smith`. It replaces `[ci] smith_version`.
- A network that reaches only its package registry points the launcher at a mirror of the GitHub
  Release archives (`SMITH_DOWNLOAD_URL`); the per-platform packages are no longer published.
- The launcher checks what it downloads with a signed `checksums.txt`.

## Decision

1. **The pin.** `[workspace] smith` is `latest` or an exact version. The launchers, the committed
   launcher of `smith bootstrap` and the generated pipelines run it. `[ci] smith_version` is a
   removed key: smith names its new place, and `smith doctor --fix` moves it.
2. **The launchers.** `monosmith` on npm, PyPI and NuGet holds a launcher with its own version line
   starting at 3.0.0, published only when the launcher changes. It starts above every smith release on
   the registries, because pip and NuGet install the highest version: the first npm launcher continues
   npm's line from a baseline tag `launcher-node@2.50.0` and releases as a breaking change; the PyPI
   and .NET launchers do the same from `launcher-python@2.53.0` and `launcher-dotnet@2.53.0` (the newest smith on PyPI and NuGet). No
   launcher has a dependency: Python and .NET have no Ed25519 of their own, so those two check the
   signature with the reference algorithm of RFC 8032, tested against the release key's signatures. Each run it picks a version: `SMITH_VERSION`,
   then the pin of the workspace it runs in, then the newest release. It runs
   `~/.smith/versions/<version>/smith`, and downloads that version first when it is missing: the
   download happens on the first run, never in an install script (pnpm 10 and `--ignore-scripts`
   skip those).
3. **Where it downloads from.** `https://github.com/monosmith-app/monosmith/releases/download/smith@<v>/monosmith_<v>_<os>_<arch>.zip`
   and that release's `checksums.txt`; `SMITH_DOWNLOAD_URL` replaces the part before `/smith@<v>/`.
   The newest release is the newest `smith@` tag GitHub lists (a mirror serves `latest.txt`).
4. **Integrity.** The release signs `checksums.txt` with minisign (`checksums.txt.minisig`). The
   public key is built into every launcher; the private key is the repository secret
   `SMITH_SIGNING_KEY`. An archive whose checksum is not in a correctly signed list is not run.
5. **Updates.** `smith update` downloads into `~/.smith/versions` and makes it the current version;
   it no longer runs npm, pip or dotnet.

## Consequences

- A release uploads the archives and the signature; no registry publish can stop it.
- A lockfile pins the launcher, not smith: the pin in `smith.toml` is the one source of truth.
- The first run needs the network (or the mirror); CI caches `~/.smith/versions`.
- A lost signing key stops signed releases until a new key and a new launcher are published.
- The first launcher release must reach the registries before the per-platform packages stop.
