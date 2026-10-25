# ADR 0004: Configuration in one minimal TOML file

- Status: accepted
- Date: 2026-10-15

## Context

Rationale not recorded. The product north star in the [index](README.md) (a tool where
you ask for a project and it just works) is the only stated driver.

## Options considered

Not recorded.

## Decision

`smith.toml`: every setting optional, so most workspaces only name themselves. Overrides
record their source for `smith explain`, and the TUI edits settings without losing
comments.

## Consequences

`smith schema` prints a JSON Schema for editors, and `smith config get/set/unset` edits
the file while keeping its comments. Later records add workspace-level tables
(for example `[dependencies]`, `[toolchains]`, `[ci]`).

## Sources

- [Configuration reference](../../reference/configuration.md)
- Recorded in the first version of the decisions log, 2026-10-15
