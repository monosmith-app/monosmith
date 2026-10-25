# ADR 0022: The MCP server runs smith commands

- Status: accepted
- Date: 2026-10-16

## Context

AI assistants need to read and change a workspace. The old CLI had an `mnci mcp` command;
the new one rebuilt it for the new commands.

## Options considered

Not recorded beyond the design in the pull request: implement the protocol directly with no
SDK dependency.

## Decision

`smith mcp` serves the workspace over the Model Context Protocol, JSON-RPC 2.0 over
newline-delimited stdio. Each tool is a smith command line: the server runs its own
executable with `CI=1` (so no command stops to ask) and returns the output headed by
`$ smith <args>`, so the assistant and the user see exactly what ran. Read-only tools carry
`readOnlyHint`. Every value becomes exactly one argument, and a value that looks like an
option is refused, so `--registry=evil` cannot be injected as a package. The command runs
directly, not through a shell.

## Consequences

An assistant cannot do more than the command line can. `release` and settings edits have no
tool (an editor calls them through `smith serve`). Tools for recipes, plugins, explain,
doctor fixes and upgrade were added on 2026-10-19
([#372](https://github.com/monosmith-app/monosmith/pull/372)), and every tool is tied to an
action in [ADR 0029](0029-one-list-of-actions.md).

## Sources

- Pull request [#193](https://github.com/monosmith-app/monosmith/pull/193) (closes issue
  [#118](https://github.com/monosmith-app/monosmith/issues/118)), commit `1fb00f7`
- [Editors](../../guides/editors.md#ai-assistants-mcp), [`smith mcp`](../../manual/mcp.md)
