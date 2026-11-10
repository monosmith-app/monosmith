# Turborepo 2.11.10 TUI: findings

Source: a real run on the same sample as Nx, with `--continue=always`. In one capture the search box showed "/ srv" but the list did not narrow. The log pane text was partly garbled, which may be an artefact of the screen capture.

## Layout
- The left column "Tasks (/ - Search)" is about 22 columns wide. Rows are `» @play/app-a#dev` (`»` selected, `•` persistent and running), a green tick, a red cross.
- The right pane shows logs, headed `@play/lib-b#dev > cache bypass, force executing <hash>`.
- The footer has two lines.

## Navigation
Arrows or j and k select. `/` searches and Esc clears. `h` shows full-screen logs. `s` streams all logs. `p` pins. `i` interacts with a task and Ctrl+z stops it. `u` and `d` scroll, `U` and `D` page, `t` and `b` jump. `c` copies. `m` opens a popup that lists all keys (verified).

## Keys and discoverability
The footer is context-sensitive. `i - Interact` shows for a running task and not for a finished one. The left foot reads `up down - Select` and `m - More binds`.

## Feedback
- Without `--continue`, one lint failure stopped everything in 1.3 seconds, closed the TUI and printed plain logs.
- The summary reads `Tasks: 1 successful, 7 total / Cached: 0 cached / Failed: ...`.
- Ctrl+C prints "Finishing writing to cache... Shutting down Turborepo tasks..." as plain text.

## Mouse
A click on a task row selects it and the logs switch. The wheel scrolls logs.

## What smith should copy
1. A short footer with `m - More binds` that opens the full key popup.
2. Click to select a task, wheel to scroll logs.
3. A full-screen logs toggle.
4. Pin.
5. An explicit Interact mode for a task that needs input.

## What to avoid
Fail-fast that kills unrelated long-running tasks and dumps plain logs. A task list that overflows into the log pane. A search with no visible effect.
