# ADR 0053: Every language plugin follows one contract

- Status: accepted
- Date: 2026-11-01

## Context

By 2.0 the language plugins differed in what a user could rely on: only the Go plugin described a
project declared in `smith.toml` ([ADR 0052](0052-a-project-without-a-manifest-is-declared-in-smith-toml.md)),
only three plugins declared the `channels` capability ([ADR 0051](0051-a-release-channel-gives-a-branch-its-own-line-of-prerelease-versions.md)),
and a Python app was sent to PyPI where a Rust app was not. The owner's rule: all languages behave
the same.

## Decision

The behaviour a user can rely on is one contract, written in
`docs/plugins/creating/language-plugin-contract.md` and tested by the conformance suite
(`smith plugin check`):

- A project is found by its manifest or declared; a plugin that can describe declared projects has
  the new `declared` capability, and one that cannot is refused by the host with one message.
- An app is never published to a registry by its plugin (the core decides).
- A package takes part in a release channel only with the `channels` capability.
- The same task names for the same kinds, and the same words for the same failure.

`declared` is a modifier capability like `channels`: no method needs it, it changes how `discover`
behaves. `smith plugin list` shows both for each language plugin. Every built-in language plugin
is brought to the contract before the stable 2.0: it describes declared projects where a folder is
a unit its tools can build, and otherwise does not declare the capability; it declares `channels`
once its registry's prerelease spelling is checked.

## Consequences

- A user meets the same rules and the same refusal in every language.
- A new language plugin has a checklist and a test.
- The TypeScript, Python and C# SDKs carry the capability in their generated types but cannot
  declare it yet.

## Sources

- `packages/plugin-sdk/conformance/check_declared_use_case.go`, `docs/plugins/creating/language-plugin-contract.md`
