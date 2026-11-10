# Nx 24.0.5 TUI: findings

Source: a real run on a sample of 3 projects (the lint of app-a fails, serve is continuous). The first run printed plain text although `tui.enabled` was set in `nx.json`. The TUI appeared only with `NX_TUI=true`. The cause is unknown. The help overlay rendered only in part, so its full list comes from the docs.

## Layout
- Header: `NX  3 build, lint, serve tasks, and 6 others they depend on   Cache  Duration`.
- The list has three groups. Running tasks sit in a cyan bracket group with "Waiting for task..." slots. Done tasks come next (red cross, green tick). Pending tasks come last (`·`).
- A braille spinner marks running rows. The label "Continuous" replaces the duration for serve tasks.
- The footer left shows a counter: `4/6 (5.2s)`.
- Enter opens a right output pane with a status dot.

## Navigation
Arrows select. Enter shows output (Enter again for full screen). `1` or `2` pins output and a `[1]` marker shows on the row. `/` filters (the bottom bar reads `/lint 6 tasks filtered out  <enter> confirm, <esc> cancel`). `q` quits. `?` opens help.

## Keys and discoverability
- List footer: `pin output: 1 or 2  show output: <enter>  filter: /  navigate: up down  quit: q  help: ?`.
- Output pane footer: `scroll: up down  copy: c  search: /  full screen: <enter>  quit: q  help: ?`.

## Feedback
- A red cross shows the failure, with the raw npm error in the pane.
- When the tasks finish, a dialog says "exiting in 1..." and offers "Press q to exit immediately or any other key to keep the TUI running".
- Cache hit or miss is a column.

## Mouse
The wheel scrolls the output pane. One click on a row did not change the selection (a single try, lag possible).

## What smith should copy
1. The Run tab as grouped running, done and pending.
2. The spinner, the duration and the `n/m (time)` counter.
3. The "Continuous" label for dev servers.
4. Pin and filter.
5. A different footer in each pane.
6. A cache column (smith already knows cache hits).

## What to avoid
The auto-exit countdown dialog. Raw npm noise with no summary line. A thin help dialog.
