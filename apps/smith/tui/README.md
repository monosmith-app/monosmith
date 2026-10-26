# tui

The interactive shell a bare `smith` opens. The design is in
[docs/architecture/tui-and-ide.md](../../../docs/architecture/tui-and-ide.md).

## Layout

The pieces that stand alone are subfeatures with their own `doc.go`: `commands`
(the keymap), `theme` (tokens, themes and symbols), `pane` (frames, wrapping,
truncation), `dialog`, `palette`, `tabs`, `footer`, `statusbar`, `toast`, the
views (`dashboard`, `doctorview`, `depsview`, `releaseview`, `projectlist`,
`detail`, `runview`, `logview`, `guide`, `settingsview` (the Settings tab)), `settingssource` (what the Settings
tab reads, and the functions it calls, which are the ones `smith plugin` and `smith config` call), `userprefs` (the
person's theme and symbols, kept in the smith home), `fixwizard` (the doctor fix dialog),
`adoptwizard` (the adopt and convert dialog), `runner`, `sources`, `screenreader`, `clipboard`, `pointer`, `layout` and
`progress`.

## The Run tab

Every task run (build, test, lint, typecheck, `run <task>`, "Lint all", "changed", from the keys,
the palette or the Home next action) stays inside the interface. The shell starts the same smith as
a child (`runner`), captures its output and shows, like Nx does, one row per task with its state
(queued, running with a spinner and the elapsed time, passed, failed, cached, skipped), a progress
bar with totals, and the output of the selected row. Failed tasks are listed first and the first
failed one is selected when the run ends. `Ctrl+C` stops the run, `r` repeats the run shown, `R`
runs only the failed and skipped tasks again, `,` and `.` show the previous and the next run of the
session (the last 20 are kept), `y` copies the command line and `Y` the log shown. The footer lists
only the keys that work at that moment. The command line shown is the one to type in a shell to get
the same run.

`tea.Exec` (leaving the interface) is used only for a command that needs a terminal of its own, one
that asks questions or takes over the screen (`catalog.NeedsTerminal`). A task run no longer does:
the old way gave real scrollback, which this replaces with the copy keys. The full logs can still
be viewed and selected: `Y` copies the whole log (or one task's) to the clipboard, `m` turns mouse
reporting off so a plain drag selects text, and the command line, run in a terminal, prints
everything again.

A task run is started with `--json-events` (one JSON event per line, see
[ADR 0026](../../../docs/architecture/decisions/0026-event-stream-and-smith-serve.md)), which
`runview` reads; any other line, and every other command, is read as text.

## Exception: the shell's own files stay in the package root

- **Path:** `apps/smith/tui/*.go` (22 production files, plus `doc.go`: `shell_*.go`,
  `tui_handler.go`, `page_contract.go`, `confirm_details_mapper.go`,
  `release_preview_mapper.go`, `no_workspace_hint_mapper.go`).
- **Rule waived:** a subfeature holds at most 12 production files
  (`smith slices check` reports it for review).
- **Why:** they are the methods of one Bubble Tea model, `Shell`. Go allows a
  type's methods in one package only, and the keys, the mouse, the runs, the
  render and the detail panel all read and change the same state. Splitting
  them means first giving that state a public shape (a state struct that the
  `keys`, `mouse`, `runs` and `render` packages take and return), which is a
  rewrite of the model and its tests, not a move.
- **Done already:** everything that does not touch `Shell` was moved out
  (generated forms to `guide`, the Projects columns to `projectlist`, loading
  the views to `sources`).
- **Temporary.** Removal condition: when the shell state becomes a value passed
  to functions (the `Page` interface in `page_contract.go` is the first step:
  each tab already draws from a width and a height), move the key handling, the
  mouse handling, the run handling and the rendering into their own packages.
