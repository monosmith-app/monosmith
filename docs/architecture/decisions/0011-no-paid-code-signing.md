# ADR 0011: No paid code signing for now

- Status: accepted
- Date: 2026-10-16 (extended 2026-10-20 with SBOM and attestations)

## Context

Signing needs an Apple Developer ID and a Windows code-signing certificate, both paid.
The decision does not record the cost or the budget behind "for now".

## Options considered

- Paid signing: deferred.
- Free provenance instead: chosen.

## Decision

No Apple Developer ID and no Windows code-signing certificate for now. Free provenance
instead: SHA-256 `checksums.txt`, npm provenance, and
([#348](https://github.com/monosmith-app/monosmith/issues/348)) a CycloneDX SBOM per
released app plus GitHub artifact attestations (`actions/attest-build-provenance`, SLSA
provenance) over the archives, packages and SBOMs, verified with `gh attestation verify`.

## Consequences

Only a binary downloaded from GitHub Releases in a browser is affected: macOS quarantines
it, and Windows SmartScreen warns. Installs through package managers and the install
scripts are not affected.

The SBOM is written by smith from the lockfile or manifest (Go and npm so far; the plugin
protocol is not extended until more ecosystems need it). Azure Pipelines and GitLab have no
attestation service, so they get the SBOM and `checksums.txt` only. Native apps' archives
are attached by the per-OS jobs and are not attested or given an SBOM yet. The rendered
SBOM is validated against the CycloneDX 1.5 schema in a test
([#421](https://github.com/monosmith-app/monosmith/pull/421)).

## Sources

- Issue [#124](https://github.com/monosmith-app/monosmith/issues/124); issue
  [#348](https://github.com/monosmith-app/monosmith/issues/348); pull requests
  [#384](https://github.com/monosmith-app/monosmith/pull/384),
  [#421](https://github.com/monosmith-app/monosmith/pull/421)
- [Installing](../../getting-started/installing.md#unsigned-binaries),
  [Releasing](../../reference/releasing.md)
