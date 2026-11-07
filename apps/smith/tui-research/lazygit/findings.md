# lazygit 0.66.2: findings

Source: a real run (screens in `screens/`). Items marked "docs" come from the documentation.

## Layout
Five left panels: [1] Status, [2] Files, [3] Branches, [4] Commits, [5] Stash. A detail pane [0] sits on the right. A Command log strip and a one-line footer sit at the bottom. Panel titles act as tabs, and the active one is green. The `n of m` counter and a scroll bar sit on the panel border.

## Navigation
- Digits 1 to 5 jump to a panel. Arrows move. Enter goes deeper. Esc goes back.
- There are no top-level tabs. `[` and `]` change the tab of a panel (docs).

## Keys and discoverability
- The footer follows the focus. Files shows `Stage: <space> | Commit: c | Stash: s | Discard: d | Reset: D | Keybindings: ?`.
- `?` opens the full key list with a "(Type to filter)" box. It is the closest thing to a command palette.
- A key that cannot run explains itself: `Disabled: The selected items don't have both staged and unstaged changes.`
- The footer right end holds Donate, Ask Question and the version. This is noise.

## Feedback
- `c` with nothing staged asks "No files staged. Commit all files?" and waits for Enter or Esc.
- The discard menu has hot keys, a Cancel row and a description box.
- The commit popup has a summary counter and a description box.
- A push with no remote asks for the upstream, then shows an Error popup with the raw git text and a copy key.
- A merge conflict shows `(merging)` and a "Conflicts!" menu (View conflicts, Abort).
- The Command log prints the real git commands.
- Empty states are plain text ("No remotes", "0 of 0"). Spinners and toasts were not seen.

## Mouse
Click focuses a panel, selects a row and switches a tab. Hover does nothing. Wheel and footer links were not verified.

## Colour and glyphs
Red is unstaged. Green is staged. Blue is the selected row (grey when unfocused). Green is the active border. `○` marks commits and `▼` marks folders.

## What smith should copy
1. A context footer in every tab (Home, Projects, Run, Doctor, Deps, Release).
2. "Disabled: reason" for a key that cannot run. Smith has this for some keys. Make it the rule.
3. A filterable key list as the command palette.
4. Described confirm menus with a Cancel row for Deps remove and Release publish.
5. A Command log line in Run and Release that shows the real smith command.
6. An error popup with a summary line and a copy key.

## What to avoid
All panels at once. Promo text in the footer. Key-only actions. A blocking first-run modal.
