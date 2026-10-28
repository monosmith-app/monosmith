# ADR 0039: A plugin declares the project kinds it scaffolds

- Status: accepted
- Date: 2026-10-28

## Context

A plugin that implements `generate` was assumed to scaffold all three kinds, app, lib and package.
Two things broke on that assumption (issue [#562](https://github.com/monosmith-app/monosmith/issues/562)):

- `smith plugin check --generate` asks every generating plugin for an app, a lib and a package, so
  it failed any scaffold that makes only an app, the built-in `go-smith-plugin` included (a smith
  plugin is an executable).
- `smith kinds` listed `<language>-<kind>` for all three kinds of every language plugin, so it
  offered `bicep-package`, which the Bicep plugin refuses to generate and `smith add` then failed.

## Options considered

**Hard-code the exceptions in the host and the conformance suites.** Rejected: an external plugin
could not say the same, and each suite (Go, TypeScript, Python, C#) would carry a list of plugin
names.

**Try each kind and treat a refusal as the answer.** Rejected: generation may install
dependencies and write files, and `smith kinds` must answer without generating anything.

**A mandatory `kinds` list.** Rejected: it would break every plugin written before it.

## Decision

The `initialize` result gains an optional `kinds` list: the project kinds `generate` can scaffold.
Absent or empty means all three, which is what every plugin that says nothing does today, so the
protocol stays at version 1.

- `smith kinds` offers `<language>-<kind>` only for the declared kinds.
- `smith add` refuses a kind the plugin that would generate it does not declare, with a message
  that names the kinds it does scaffold, before anything is written.
- The conformance checks (`smith plugin check --generate`, and the in-process suites of the Go,
  TypeScript, Python and C# SDKs, which stay the same) ask for and check only the declared kinds.
- The Bicep plugin declares `app` and `lib`; `go-smith-plugin` declares `app`.

## Consequences

- A plugin that scaffolds fewer kinds says so in one place, and no host or suite needs to know its
  name.
- The plugin still refuses a kind it was not asked for with its own error; `kinds` is what lets the
  host not ask.
- `smith kind` (`project/kind`) is unchanged: it has its own refusal.

## Sources

- Issue [#562](https://github.com/monosmith-app/monosmith/issues/562)
- [The plugin protocol](../plugin-protocol.md), [the protocol cheat sheet](../../plugins/creating/protocol-cheat-sheet.md)
- `packages/plugin-sdk/protocol/initialize_contract.go`, `packages/plugin-sdk/conformance`
