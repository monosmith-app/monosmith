# tui

The interactive shell a bare `smith` opens. The design is in
[docs/architecture/tui-and-ide.md](../../../docs/architecture/tui-and-ide.md).

## Layout

The pieces that stand alone are subfeatures with their own `doc.go`: `commands`
(the keymap), `theme` (tokens, themes and symbols), `pane` (frames, wrapping,
truncation), `dialog`, `palette`, `tabs`, `footer`, `statusbar`, `toast`, the
views (`dashboard`, `doctorview`, `depsview`, `releaseview`, `projectlist`,
`detail`, `runview`, `logview`, `guide`, `settingsview` (the Settings tab; its detail is drawn in the side panel of the shell, `WithSideDetail`)), `settingssource` (what the Settings
tab reads, and the functions it calls, which are the ones `smith plugin` and `smith config` call), `userprefs` (the
person's theme and symbols, kept in the smith home), `fixwizard` (the doctor fix dialog),
`adoptwizard` (the adopt and convert dialog), `cachewizard` (the remote cache setup, the screens of `smith cache setup`), `runner`, `sources`, `screenreader`, `clipboard`, `pointer`, `pathfinder` (the list of folders and files under a path field), `layout`,
`progress`, `splitter` (the dividers: keys, step and spots from plain values), `selection` (drag to select text: regions cut along the dividers, the selected text and its highlight), `runhistory` (the run panel's
history: ago text, header parts, recent rows, steps, the rerun-failed line) and `commandtext` (the
text a captured command printed: plan lines, failure text, report file, not-committed line).

## Text

The TUI shows no wording of its own. Every sentence, label and hint is a key in the shared catalogue
in [`libs/words`](../../../libs/words/README.md), read with `words.Say` and `words.Count`. The guard tests
that scan this tree are in `wordsguard`: a key the code asks for must exist, its placeholders must be
filled, a key must be claimed by a surface, and a migrated package holds no raw wording. The TUI
claims a key by asking for it in code; the editors claim theirs in `libs/words/surface/<name>.json`.

## The run panel

There is no Run tab. While a task runs, or when the user opens a past run, the right panel of Projects
shows the run (`shell_run_panel_mapper.go`); Recent runs sits on the bottom rows of the list's pane (three runs, or all of them while it has the keyboard). Esc or Back shows the
project again. What follows holds for that panel.

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

## Commands that show their text in a dialog

Some commands are not shown in the run panel: the shell runs the same smith as a child
(`sources.CaptureCommand`), reads what it printed and shows it in a dialog (`dialog.Confirm`,
or `dialog.NewNotice` for a text that only tells). `smith undo --dry-run` is read this way
before `smith undo --yes` asks; `smith add ... --dry-run` is read before the real add; `smith
choices --json` feeds the Add form (`guide.AddProjectAction`) and `smith plugin search --json`
feeds the plugin install form (`guide.PluginInstallAction`). The files of these flows are
`shell_capture_use_case.go` and `guide/add_project_choices_mapper.go`. The derived files of a
project come from `derive.BuildPlan` (`sources/load_derived_files_use_case.go`), shown by
`shell_derived_mapper.go`.

## Exception: the shell's own files stay in the package root

- **Path:** `apps/smith/tui/*.go` (`shell_*.go`, `tui_handler.go`, `splits_repository.go`,
  `page_contract.go`, `confirm_details_mapper.go`, `release_preview_mapper.go`,
  `no_workspace_hint_mapper.go` and the other mappers of the shell).
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

### What is left in the root, and the next slice to cut

The shell was split by role first: `shell_model.go` keeps the struct and its
queries, `shell_update_use_case.go` holds `Update`, `handle` and `resize`, and
`splits_repository.go` holds the one file write of the handler. The pure parts
left the root as slices that take plain values: `runhistory`, `commandtext` and
`splitter`. The linkable libraries of a project are a function of the project
list (`shell_libraries_mapper.go`), not a method.

Still methods on `Shell`, and why each stays for now:

- **Keys** (`shell_keys_use_case.go`): `key` and `perform` read and write most of
  the state (tabs, dialogs, runs, forms). Next cut: a `keyroute` slice that takes
  a small `Router` interface (`Open(dialog)`, `Say(text)`, `Go(tab)`), with the
  per-tab key handlers (`projectsKey`, `detailKey`, `runKey`) as the first to
  move because each touches one view.
- **Mouse** (`shell_mouse_use_case.go`): it reads the geometry of every view
  (`hit` maps). Next cut: move the hit-testing to a `hits` mapper that returns a
  target for (x, y), leaving the Shell only the reaction.
- **Capture** (`shell_capture_use_case.go`, 600 lines): each `*Read` method builds
  a dialog from a captured text and writes `s.confirm`. Next cut: a `captureflow`
  slice that returns the dialog to show (a value) and lets the Shell store it.
- **Forms** (`shell_adopt_`, `shell_fix_`, `shell_compose_`, `shell_tools_`,
  `shell_deps_*`): each is a dialog the Shell opens, feeds and closes. Next cut:
  one slice per dialog with an `Open`, `Key` and `Done` that return values, as
  `dialog.Confirm` already does.
- **Run** (`shell_run_use_case.go`): it starts processes and updates `runs`. Next
  cut: a `runflow` slice owning the run history state, with `runhistory` as its
  algorithms.
