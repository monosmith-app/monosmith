# btop4win 1.0.5: findings

Source: a real run. Screens show real process names and a user name: do not publish them. A `k` key sent with a process selected may have killed a process. See `install-notes.md`.

## Layout
CPU on top. Memory, disks and network on the left. Processes on the right. Boxes are labelled `1cpu` `2mem` `3net` `4proc`, and the digit toggles the box. The CPU title holds `menu`, `preset`, the clock and `- 1500ms +`.

## Navigation
No tabs. `p` cycles presets. Enter opens process detail. Left and right pick the sort column. `e` shows a tree. Esc or `m` opens a menu (Options, Help, Quit).

## Discoverability
A red hotkey letter sits in each label. Help opens with F1 or `h` and has two pages. `?` does nothing.

## Feedback
- A filter title `f python del` shows a `0/12` counter and a delete button.
- A green "Running" status shows in the detail pane.
- Bars go green, yellow, red.
- At 80x24 the process box is dropped.
- No error or empty state was seen.

## Mouse
Documented as click and wheel. One click changed the layout and the cause is unknown.

## What smith should copy
The hotkey letter inside a label (tabs, footer). A title that holds state. A counter in the border. A refresh control. Reflow on a small terminal. A health colour ramp.

## What to avoid
Kill on one key with no confirmation. Superscript digits. No hover. Dense graphics.
