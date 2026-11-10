# Yazi 26.9.1: findings

Source: a real run. The Tasks modal was empty during a 1.5 GB copy, so no progress bar was seen. One run started in `C:\` by mistake. It showed only a trash dialog and was cancelled with `n`. Nothing was deleted.

## Layout
Parent, current and preview columns. The header is the path. The status bar reads `NOR 6B name 5/5`. A task badge sits top right.

## Navigation
h, j, k, l and arrows. A prefix key opens a bottom panel: `g` for go-to, `,` for sort. `/` finds, `f` filters, `t` opens a tab, Space selects, `y` and `p` copy and paste, `d` trashes, `a` creates, `r` renames.

## Discoverability
There is no footer. `~` opens the full key list. The bottom panel after a prefix teaches the next keys.

## Feedback
- The trash dialog lists the full paths with `[Y]es` and `(N)o`.
- A filter with no match shows `No filter results`.
- A `Create:` input box asks for the name.
- The task badge counts background copies.

## Mouse
A click moves the cursor. Wheel and hover were not verified.

## What smith should copy
The bottom panel that lists the next keys after a prefix. The task badge (a count of background tasks in the tab strip). A confirmation that lists what changes. A status bar with position (`5/5`). A named empty state. Collapse on a narrow terminal.

## What to avoid
No start hints. A blank task list. Vim-only movement. Windows setup friction (it needs `file.exe` and misread a backslash path).
