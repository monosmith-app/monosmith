# tui

The interactive shell a bare `smith` opens. The design is in
[docs/design/tui-and-ide.md](../../../docs/design/tui-and-ide.md).

## Layout

The pieces that stand alone are subfeatures with their own `doc.go`: `commands`
(the keymap), `theme` (tokens, themes and symbols), `pane` (frames, wrapping,
truncation), `dialog`, `palette`, `tabs`, `footer`, `statusbar`, `toast`, the
views (`dashboard`, `doctorview`, `depsview`, `releaseview`, `projectlist`,
`detail`, `runview`, `logview`, `guide`), `runner`, `sources`, `screenreader`,
`clipboard`, `pointer`, `layout` and `progress`.

## Exception: the shell's own files stay in the package root

- **Path:** `apps/smith/tui/*.go` (19 production files: `shell_*.go`,
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
