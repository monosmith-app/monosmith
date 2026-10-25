# The architecture of the terminal interface

A bare `smith` in a terminal opens a full-screen shell for the workspace. This page explains how
that shell is put together: one Bubble Tea model, a single keymap that the keys, footer, help and
palette are all drawn from, pages that draw themselves into one pane, wizards built on one dialog,
and a runner that starts the same `smith` binary as a child. The organising rule is that the shell
never does what a command line could not.

Audience: contributors to `apps/smith/tui`. For what the screens and keys are, see
[The smith terminal interface](../manual/tui.md). The design proposal that preceded the build, with
its history, is [The TUI and the IDE plugins](tui-and-ide.md).

## The rule: the shell runs smith commands

The shell is a front end. Every action it offers ends as a `smith` command line that is printed in
the status bar before it runs, and a command that changes the workspace, installs or publishes asks
first in a dialog that shows the line. The point is that anything done in the shell can be repeated
in a script or in CI, and that the shell and the command line cannot disagree.

Two things need qualifying, because the code does both:

- **Reading is in process.** The views (Home, Projects, Doctor, Deps, Release) are loaded by
  `tui/sources` calling the engine's Go packages directly: `doctor`, `release`, `affected`,
  `projectgraph`, `taskgraph` (`sources/load_views_use_case.go`). Reading changes nothing and does
  not need a command line.
- **Acting is a child process.** Anything that runs, writes or publishes goes through
  `tui/runner`, which starts `os.Executable()`, that is this very smith, as `smith <argv>`
  (`runner/child_process_client.go`).

The decision is recorded in [ADR 0032](decisions/0032-single-key-tui-keymap.md) ("the shell never
acts by itself") and, for the list of what the shell may offer, [ADR 0029](decisions/0029-one-list-of-actions.md).

## Components

```mermaid
flowchart TD
  handler["tui_handler.go<br/>Interactive(), Choose()"] --> shell["Shell<br/>(one Bubble Tea model)"]
  shell --> keys["shell_keys_use_case.go<br/>commands.ByKey(key, state)"]
  keys --> keymap["tui/commands<br/>keymap_config.go: the one list"]
  keymap --> footer["footer hints<br/>footer_hints_mapper.go"]
  keymap --> help["help overlay<br/>help_rows_mapper.go"]
  keymap --> palette["tui/palette"]
  shell --> pages["Page interface<br/>page_contract.go"]
  pages --> views["dashboard, projectlist, runview+logview,<br/>doctorview, depsview, releaseview, guide"]
  shell --> dialogs["dialog (confirm, help)<br/>fixwizard, adoptwizard"]
  shell --> exec["execute(argv)"]
  exec -->|"catalog.Classify"| classify["libs/actions/catalog"]
  exec --> runner["tui/runner<br/>child smith process"]
  exec --> inline["runner.Inline<br/>tea.Exec: real terminal"]
  shell --> sources["tui/sources<br/>load views in process"]
  runner --> runview["runview.ParseOutputLine"]
```

| Package | Responsibility |
|---|---|
| `tui` (root, 21 production files) | the `Shell` model: keys, mouse, runs, render, detail panel. One type's methods cannot be split across Go packages, which is the written reason the root exceeds the slice size limit (`apps/smith/tui/README.md`, temporary) |
| `tui/commands` | the keymap and everything computed from it: availability, footer hints, help rows, key labels, parsing a run request |
| `tui/theme` | tokens (dark, light, plain), icons (Unicode, ASCII), the theme chosen from the environment |
| `tui/layout`, `pane`, `tabs`, `footer`, `statusbar`, `toast`, `detail` | geometry and drawing of the pieces |
| `tui/dialog`, `palette`, `guide` | confirmation and help dialogs, the command palette, guided forms |
| `tui/dashboard`, `projectlist`, `runview`, `logview`, `doctorview`, `depsview`, `releaseview` | the views |
| `tui/fixwizard`, `tui/adoptwizard` | the doctor fix dialog and the adopt dialog |
| `tui/runner` | starting and stopping child smith processes |
| `tui/sources` | loading what the views show (replaceable in tests) |
| `tui/screenreader`, `clipboard`, `pointer`, `progress` | plain mode, OSC 52 copy, double-click detection, progress bars |

## One keymap, three consumers

`tui/commands/keymap_config.go` holds `all`, a list of `Command`s. A `Command` has an ID, a group
for the help, a title, its keys, a `Kind` (what happens when chosen: `ShowTab`, `Execute`, `RunTask`,
`Guided`, `OpenDoctor`, `OpenFixWizard`, `OpenAdoptWizard`, `Stop`, `PaneKey`, and others), the
views where it applies (`Views`, `ShowIn`), its preconditions (`NeedsWorkspace`, `NeedsIdle`,
`NeedsRunning`, `NeedsNarrowList`, `OnlyInside`, `OnlyOutside`) and the footer label.

The same list produces:

| Consumer | Function |
|---|---|
| key dispatch | `commands.ByKey(key, state)` returns the command, or the reason it is unavailable, which becomes the toast |
| footer hints | `commands.Hints(state)` returns the view's keys then the tail every view ends with |
| help overlay (`?`, `F1`) | `commands.HelpRows(state, viewTitle)` |
| palette | the commands that have no key are reached here, plus every command of the CLI catalogue |

Because one list feeds all of them, a hint cannot name a key that does nothing; that is the design
claim of [ADR 0032](decisions/0032-single-key-tui-keymap.md). Keys are single keys, no plain key
fires while a text field or form has the keyboard (`plain(key)` in `command_availability_policy.go`),
and the keys of a view's own pane come first in the list so they win over a global key of the same letter
(Deps marks all with `a`, which elsewhere adds a project). `PaneKey` entries are listed for the footer
and help but never run by the shell: the open view handles them.

## Pages

A view that appears in the one content pane implements `Page` (`page_contract.go`): a title, the
footer hints, `Update(msg)` for keys meant for it, and `View(width, height)`. The shell asks the
open tab for its `Page` and draws nothing else of the view. The layout is adaptive: below 100
columns the list takes the whole width and Enter opens the detail; wider, the list and the detail
sit side by side (`layout.divide_screen_algorithm.go`). Tabs shorten their names before the strip
would scroll, and below 24 rows they become one plain row.

Views load lazily: Doctor, Deps and Release load on first open; Doctor also starts in the
background when the shell opens, so the header shows its counts before any tab is visited.

## Running commands

```mermaid
sequenceDiagram
  participant P as person
  participant S as Shell
  participant C as libs/actions/catalog
  participant R as runner or Inline
  participant X as child: smith argv
  P->>S: press b
  S->>S: commands.ByKey -> RunTask build
  S->>S: execute(argv)
  S->>C: Classify(argv)
  alt changes the workspace
    S->>P: confirmation dialog shows the command line
    P->>S: y
    S->>S: argv = WithConfirmation(argv)
  else needs a terminal
    S-->>P: quit the shell, smith runs argv in the terminal
  end
  alt task run or a printing command
    S->>R: Inline: tea.Exec hands the terminal over
    R->>X: smith build (inherits the terminal)
    X-->>P: output in the real scrollback
    R-->>S: exit, then a key brings the shell back
  else other commands
    S->>R: runner.Process
    R->>X: smith argv
    X-->>R: lines
    R-->>S: runview parses lines into task states
  end
```

`Shell.execute` (`shell_run_use_case.go`) classifies the command line with
`libs/actions/catalog.Classify` into three kinds: needs a terminal (a command that asks questions;
the shell quits and smith runs it), changes the workspace (ask first), and the rest. A task run, and
the commands in `cliVerbs` that print and end (`plugin`, `recipe`, `config`, `info`, `graph`,
`affected`, `explain`, `projects`, `stacks`, `kinds`, `commands`, `help`), use `runner.Inline`: Bubble
Tea gives the terminal to the child through `tea.Exec`, so output goes to the real scrollback with
the child's own status line and stays after the interface ends. Everything else runs through
`runner.Process` and appears in the Run tab. Outside a workspace there is no Run tab, so `smith new`,
`adopt` and `convert` quit the shell and run in the terminal.

### runview parses lines, not events

The Run tab turns the child's output into task states with `runview.ParseOutputLine`, which
recognises `<task> ▶ <command>`, `<task> ✔ <duration>`, `<task> ✔ from cache`, `<task> ✖ <reason>`,
`<task> ■ stopped`, `<task> ! <message>`, the skip and failure lines of the closing summary and the
`✔ 3/4 tasks succeeded` line. It does not use the JSON event stream that `smith run --json-events`
and `smith serve` emit ([ADR 0026](decisions/0026-event-stream-and-smith-serve.md)). The package
comment says so, and the shell therefore depends on smith's text format; moving it to the event
stream is possible but is not recorded as planned or decided.

## Wizards

The doctor fix wizard (`tui/fixwizard`) is a state machine of stages: questions (one per
decision), a scrollable review (files as diffs, commands, projects covered), progress with a bar,
and a failure screen with the step's output and Retry. It renders a `doctorfix.Planned` and applies
it through `doctorfix`, so it asks exactly what `smith doctor --fix` asks
([ADR 0033](decisions/0033-doctor-fixes-with-steps-and-choices.md)).

The adopt wizard (`tui/adoptwizard`) is a second state machine layered on the first: every screen is
a `fixwizard.Wizard`, so keys, mouse and drawing are identical. It holds no logic; steps that take
time come out as `Job`s the shell runs off the UI goroutine, and `adoptflow.Session` does the work.
The stages and calls are in [How adopt and convert work](adopt-and-convert.md).

## Themes, plain mode, mouse

| Setting | Effect | Code |
|---|---|---|
| `NO_COLOR` (any value) | plain tokens | `theme.Choose` |
| `SMITH_THEME=light\|dark\|plain` | forces tokens; `plain` also requests plain mode | `theme.Choose`, `screenreader.Plain` |
| terminal background | decides light or dark when `SMITH_THEME` is unset | `theme/dark_background_policy.go` |
| `SMITH_ASCII`, `TERM=dumb\|linux` | ASCII symbols | `theme.Choose` |
| `SMITH_SCREEN_READER` (anything but empty, `0`, `false`, `no`, `off`) | no full-screen shell, no colour, no mouse; numbered questions | `screenreader.Wanted` |
| stdout not a terminal | plain questions on stderr | `screenreader.Plain` |
| stdin not a terminal, or `CI` set | `Interactive()` is false: `smith` prints its help | `tui_handler.go` |

Every state keeps its word beside its symbol, so colour is never the only signal. The mouse reports
cell motion only, so Shift+drag still selects text, and `m` turns it off. Copy uses the OSC 52
terminal sequence, which also works over SSH.

## Where in the code

| Concern | Path |
|---|---|
| Entry and mode choice | `apps/smith/tui/tui_handler.go` |
| Shell model | `apps/smith/tui/shell_*.go` |
| Keymap | `apps/smith/tui/commands/keymap_config.go` |
| Classifying actions | `libs/actions/catalog` |
| Runner | `apps/smith/tui/runner` |
| Output parsing | `apps/smith/tui/runview` |
| Wizards | `apps/smith/tui/fixwizard`, `apps/smith/tui/adoptwizard`, `apps/smith/adoptflow`, `apps/smith/doctorfix` |

## Decisions

| Decision | Record |
|---|---|
| Single-key keymap, one adaptive pane, no menu bar | [ADR 0032](decisions/0032-single-key-tui-keymap.md) |
| One list of actions shared by shell, MCP and serve | [ADR 0029](decisions/0029-one-list-of-actions.md) |
| Event stream and `smith serve` (not used by the shell's Run tab) | [ADR 0026](decisions/0026-event-stream-and-smith-serve.md) |
| Doctor fixes with steps and choices | [ADR 0033](decisions/0033-doctor-fixes-with-steps-and-choices.md) |
| Adopt and convert depth levels | [ADR 0034](decisions/0034-adopt-and-convert-depth-levels.md) |

The choice of Bubble Tea as the library has no decision record; the dependency is in `go.mod` and
the package comments name it.
