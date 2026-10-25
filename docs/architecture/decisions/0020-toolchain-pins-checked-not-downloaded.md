# ADR 0020: Toolchains are pinned by smith and provisioned by the developer's manager

- Status: accepted
- Date: 2026-10-20

## Context

smith pins runtime versions only in the generated CI and dev container; locally it runs
whatever is installed. The cache key already includes the installed version
(`Project.Toolchain`, probed by each plugin), so two machines on different versions never
share a result, but nothing says which version a workspace should use. Moon downloads and
pins runtimes itself (through proto).

## Options considered

- Download runtimes inside smith, as Moon does: not chosen. Fetching a compiler per OS and
  architecture means a registry of URLs and checksums for every language smith supports,
  signature and offline handling, and a path shim so tasks find the binary: a second
  package manager inside smith, kept current against ten ecosystems. mise, asdf and the dev
  container already do that and are what teams use.
- Pin and check only: chosen.

## Decision

smith owns the pin and the check, not the download. `[toolchains]` in `smith.toml` names a
version per toolchain label (`node = "22"`, `go = "1.23"`; a prefix on dot boundaries).
`smith doctor` compares each pin with the version the plugins probed and fails on a
mismatch, with the remedy (`mise use node@22`) and a fix that writes the pins to
`.tool-versions`, the file mise and asdf both read. The fix is safe (applied by `--fix`)
only when the file is new. A language plugin that declares how CI installs its runtime
(`protocol.Toolchain`) can declare the fetch later without a protocol change to the pin.

## Consequences

A pin for a tool no project reports on (not installed, or unused) gives no finding yet,
because plugins leave an unprobed toolchain out. Not done, to follow as separate issues: CI
and the dev container reading the pin instead of each plugin's version; the pin in the cache
key when the tool is missing; `smith run` refusing a mismatch (today it only warns through
doctor); a plugin-declared fetch for managers other than mise and asdf; checksums and offline
use, which only matter if smith ever downloads itself.

## Sources

- Issue [#342](https://github.com/monosmith-app/monosmith/issues/342); pull requests
  [#365](https://github.com/monosmith-app/monosmith/pull/365),
  [#393](https://github.com/monosmith-app/monosmith/pull/393)
- [Configuration](../../reference/configuration.md)
