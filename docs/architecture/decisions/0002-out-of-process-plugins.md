# ADR 0002: A language-agnostic core with out-of-process plugins

- Status: accepted
- Date: 2026-10-15

## Context

The tool has to cover many languages without the core knowing any of them, and plugin
authors should not have to write Go.

## Options considered

Not recorded beyond the decision itself; the prior-art survey in
[research 03](../research/03-prior-art-and-stack.md) is the input. Rationale for
rejecting in-process plugins is not recorded.

## Decision

Everything language-specific lives in plugins: separate processes speaking a versioned
JSON-RPC 2.0 protocol over stdio, framed like the Language Server Protocol, in any
language. Built-in plugins ship in the `smith` binary but run out of process too
(`smith plugin serve <name>`), so they speak the same protocol an external plugin does.

## Consequences

A new language is a plugin, with no change to the core. The protocol is versioned by an
integer, and additions are new capabilities that old peers ignore. The Go SDK
(`packages/plugin-sdk`) derives a plugin's capabilities from what it implements. How
plugins combine is refined in [ADR 0021](0021-plugin-roles-and-per-project-stacks.md) and
[ADR 0023](0023-recipes-with-levels-and-slots.md); how a plugin is found and installed by
name is still open (see the [index](README.md#open-not-decided-yet)).

## Sources

- [Plugin system](../plugin-system.md), [Plugins reference](../../reference/plugins.md)
- [Research 03](../research/03-prior-art-and-stack.md)
- Recorded in the first version of the decisions log, 2026-10-15
