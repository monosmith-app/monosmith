# ADR 0029: One list of actions that the shell, MCP and serve read

- Status: accepted
- Date: 2026-10-22 (phases 3a and 3b merged 2026-10-23)

## Context

The terminal shell kept its own list of commands, so `adopt`, `convert` and `remove` could
only be run from the palette, and the MCP tools and `smith serve` methods were separate
tables. Three lists could drift from each other and from the commands.

## Options considered

Generating MCP tool schemas from the actions' arguments was left out, because the tool
descriptions are written for an assistant to read.

## Decision

`libs/actions` holds one catalog: every command with its safety, whether it asks first, and
why it has no menu entry, MCP tool or serve method (`NotInMCP`, `NotInServe`). The shell's
classifier and its `--yes` rule read it. Drift tests fail when a command has no action, an
action names no command, or an action is neither served nor excused.

## Consequences

Adding a command forces a decision about each front end. `adopt` without flags is read-only;
`--tags` and `--init` change the workspace. Guided forms for every command no menu covers are
generated from the commands themselves
([#414](https://github.com/monosmith-app/monosmith/pull/414)).

## Sources

- Issue [#302](https://github.com/monosmith-app/monosmith/issues/302); pull requests
  [#406](https://github.com/monosmith-app/monosmith/pull/406) (phase 1),
  [#415](https://github.com/monosmith-app/monosmith/pull/415) (MCP),
  [#418](https://github.com/monosmith-app/monosmith/pull/418) (serve)
- [TUI architecture](../tui-architecture.md)
