# ADR 0032: The terminal interface uses single-key actions and one adaptive pane

- Status: accepted (replaces the desktop-style shell of pull request #382)
- Date: 2026-10-24

## Context

The first rich shell ([#382](https://github.com/monosmith-app/monosmith/pull/382),
2026-10-20) was desktop-styled: a menu bar with dropdowns, a button bar, a context menu,
Alt+1..9 tabs, a permanent navigator tree and an inspector. A first-time-user review
([#437](https://github.com/monosmith-app/monosmith/issues/437)) found among other things that
Ctrl+N meant two different things and that the palette outside a workspace listed commands
that could not work ([#472](https://github.com/monosmith-app/monosmith/issues/472)). The
redesign is described there as a terminal-native tool, released in smith 1.56.0.

## Options considered

Not recorded beyond the issues: the redesign epic discussion is in the thread of #437.

## Decision

- One keymap file (`apps/smith/tui/commands/keymap_config.go`) drives the keys, the per-view
  footer hints, the help (`?`, `F1`) and the palette entries, so a label cannot name a key
  that does nothing. `b t l c w d i y m q ? : /` are single keys, `1` to `6` open the views,
  and no plain key fires while a field or form has the keyboard
  ([#512](https://github.com/monosmith-app/monosmith/issues/512)).
- Delete the menu bar, dropdowns, button bar, context menu, Alt+1..9 and the Ctrl letter
  binds. Commands that need a workspace are left out of the palette outside one.
- An adaptive layout: header (workspace, branch, health), one content pane and a footer of
  the current view's keys. Under 100 columns the list fills the width and Enter opens the
  detail; wider shows both. Views implement a small page interface
  ([#513](https://github.com/monosmith-app/monosmith/issues/513)).
- Polish ([#516](https://github.com/monosmith-app/monosmith/issues/516)): one border weight,
  colour for state only, themes (`NO_COLOR`, `SMITH_THEME`), and a plain mode with numbered
  questions for screen readers.
- The principle kept from the first design: the shell never acts by itself; every action ends
  as a smith command line shown before it runs.

## Consequences

`tui/` still exceeds the slice size threshold and carries a written exception in
`apps/smith/tui/README.md`. The Plugins, Recipes, Settings and guide tabs were dropped;
those jobs were smith commands reached from the palette. That left a first-time user with
no way to switch a plugin or pick a style without leaving the interface, so a Settings tab
(plugins, style, the settings of `smith.toml`, recipes, the guided forms) brought them back as
one tab; see "What the Settings tab does" in [the interface design](../tui-and-ide.md).

## Sources

- Commit `82b490e` (closes #512, #472), commit `40cd473` (#513), pull request
  [#522](https://github.com/monosmith-app/monosmith/pull/522) (merge `2bde7df`)
- [The smith terminal interface](../../manual/tui.md), [TUI architecture](../tui-architecture.md),
  [The TUI and the IDE plugins](../tui-and-ide.md)
