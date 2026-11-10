# k9s findings

Source of facts: a live run of k9s v0.51.0 against a local kind cluster (Kubernetes v1.31.2), namespace `demo`. Screens: `screens/*.png` (colour render of the real terminal screen) and `*.txt`. Items marked (docs) come from my knowledge of the k9s manual and were not run.

Screens by name: `01-pods`, `02-help`, `03-filter-typing`, `04-filter-applied`, `05-describe`, `06-logs-live`, `09-command-autocomplete`, `11-pulse`, `12-xray`, `13-bad-command`, `20-delete-dialog`, `23-help`, `26-mouse-click-row`, `27-services`.
Some shots from the first long run are out of order because a bare `Esc` needs about 2 seconds to act. Use the `2x-` shots when in doubt.

## Layout

- Top block, always visible: cluster facts on the left (Context, Cluster, User, K9s Rev, K8s Rev, CPU, MEM). Key hints for the current view in the middle, laid out in columns. The logo on the right.
- Centre: one bordered table. The border title holds the resource, the namespace and the row count: `pods(demo)[6]`. A filter adds the filter text to the title.
- Bottom: a breadcrumb of the view stack (`<pod>`, then `<logs>`). It is the only "where am I" marker.
- Command and filter prompt: a one-line box that appears above the table (`> deploy`, `/ web`).
- One view fills the whole screen. There are no tabs and no side panels.

## Navigation model

- Resource-first and keyboard-first. `:` plus an alias (`:pods`, `:svc`, `:deploy`, `:xray deploy`, `:pulse`) opens a resource view.
- Autocomplete shows the best match in dim text while you type, `Tab` or Right accepts it (docs; the dim match itself is in `09-command-autocomplete`).
- Drill down: `Enter` opens the selected object, `Esc` goes back up the breadcrumb. `d` describe, `y` YAML, `l` logs, `s` shell are one key each.
- Digits `0`, `1`, `2` switch namespace; the list sits in the header.
- `/` filters the table live while you type (`03-filter-typing`: 6 rows became 3). `Esc` clears it.
- Hand-off: `:xray` and `:pulse` are other lenses on the same data.

## Keys and discoverability

- The header always lists the keys that work in the current view. The list changes with the view (pods show Attach, Delete, Describe, Logs; logs show tail ranges and Toggle AutoScroll). This is a footer-hint pattern placed at the top.
- `?` opens a full help page in four columns (Resource, General, Navigation, Hotkeys), with `Esc` to go back (`23-help`).
- The command prompt is the palette. There is no fuzzy command list; you must know an alias, but `ctrl-a` lists them (the Help page shows `<ctrl-a> Aliases`; the aliases view was not opened).
- Key labels use one style: `<ctrl-d>`, `<shift-f>`, `<?>`. Keys sit in blue, actions in white, which scans fast.

## Feedback

- Errors use a modal with a sad-cow picture, the message and one `Dismiss` button (`13-bad-command`: "`nope` command not found"). Charming, but the picture costs 8 rows.
- Destructive action: `ctrl-d` opens a dialog "Delete pods demo/web-...?" with `Propagation: Background`, a `Force:` option and the buttons `Cancel` and `OK`. Cancel has focus by default (`20-delete-dialog`). The default is the safe choice.
- Live data: the table refreshes by itself (restart counts moved from 2 to 4 between shots; `AGE` grew). No spinner was needed.
- Logs stream live. A status line shows `Autoscroll:On ColumnLock:Off FullScreen:Off Timestamps:Off Wrap:Off`. When the container ended, a line says `stream closed: EOF for demo/crash (crash)`.
- Health by colour: orange-red rows for `ImagePullBackOff` and `CrashLoopBackOff`, pale blue for running pods. The xray view adds a red `TOAST` badge on broken deployments (`12-xray`).
- Empty state and toasts: not captured. CPU and MEM show `n/a` in red because the cluster has no metrics server (the cause is known; the red colour was seen).

## Mouse

- A click on a table row moved the selection to it (`26-mouse-click-row`). That is all that was verified.
- Wheel scroll, hover and clickable buttons in the dialog were not verified. Mouse is a minor feature in k9s (docs).

## Colour and glyphs

- Dark background, a cyan frame, orange for header labels and the logo, blue for keys, magenta for namespace digits and the selected namespace in the title.
- Status glyphs are small: `●` in the `PF` column, emoji in prompts. A tree glyph set in xray. Without Nerd Fonts the xray icons show as boxes in the PNG.

## What smith should copy

1. A header strip with the keys of the current view (map: smith footer key hints). The list follows the view, so the user never reads keys that do nothing. Put it in the footer in smith, one line.
2. Border title as a status line: `pods(demo)[6]` becomes `Projects (12)`, `Doctor (3 faults)`, `Deps (outdated 4)` (map: every tab).
3. `/` live filter with the count in the title (map: Projects, Deps, Run task list).
4. `:` command prompt with dim autocomplete (map: smith command palette).
5. Delete dialog with `Cancel` focused by default and the option fields visible (map: Release publish, Deps remove, Projects remove).
6. Row colour by health (map: Doctor faults in red, Deps outdated in amber).
7. A "stream closed" last line when a live log ends (map: Run tab output; show exit code on the last line).
8. A pulse view: one screen of counts per kind, as a Home summary (map: Home).

## What to avoid

- The big error picture: it hides the table. Use a one-line error bar or a small modal.
- Aliases you must know: add a list to the palette.
- Waiting 2 seconds for a bare `Esc` is an issue of the terminal layer here, not of k9s. Make sure smith handles `Esc` with no delay.
- Breadcrumb at the bottom beside the key hint area is easy to miss. Smith has tabs, so tabs are the breadcrumb.
