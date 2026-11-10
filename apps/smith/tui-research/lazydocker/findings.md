# lazydocker 0.25.2: findings

Source: a real run against a three-service Compose project. Confirm, progress and error states for stop, remove and restart were not seen. Mouse was not verified. The no-daemon screen was not captured. Redraws lagged 3 to 9 seconds because the machine was busy.

## Layout
Left panels: [1] Project, [2] Services, [3] Standalone Containers, [4] Images, [5] Volumes, [6] Networks. The right pane title holds tabs: `Logs - Stats - Container Env - Container Config - Top`. A row reads `exited (3) failing 0.00%` or `running web 0.00% 18080->80/tcp`.

## Navigation
- `2` focuses a panel. Left and right arrows move between panels, not tabs.
- `[` and `]` change the tab.
- Selecting a row streams its logs at once, with no key press.

## Keys and discoverability
- The footer is static: `PgUp/PgDn: scroll, b: view bulk commands, q: quit, x: menu, arrows: navigate`.
- `x` opens a menu of all actions with hot keys (up, remove, stop, pause, restart, start, attach, logs, exec shell, open in browser). Its footer is `enter: execute, esc: close`.

## Feedback
The exit code and the running or exited colour are in the row. The screen gave no spinner or message during a slow redraw.

## What smith should copy
1. Live detail on selection, with no extra key.
2. A one-key action menu with hot keys. It can seed the palette.
3. The exit code in the row (Run, Doctor).
4. Tabs in the detail title.
5. An "open in browser" action for a dev server.

## What to avoid
The static footer. Silent waiting. Empty boxes with no words. Dark-blue text on black. Listing host-wide items.
