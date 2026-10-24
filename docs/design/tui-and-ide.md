# Design: the TUI and the IDE plugins

Status: **proposal**, for discussion. Nothing here is built unless it says so.
It extends the TUI issue (#302) and the IDE plugins issue (#303).

## The goal

The product owner asked for a terminal interface that looks and behaves like
[cuppa](https://github.com/meta-ui-labs/cuppa) (a visual designer for Bubble Tea
interfaces): a full interface with **all the options visible**, **mouse
compatible** and **rich**. From inside it you can start executions, install
packages and run every function of smith, with **real-time feedback** of the
results. The same functions are available from the **IDE plugins**, which use
the **smith the team has installed** (as the ESLint and Jest extensions use the
workspace's own tools), so the team controls the version.

## What cuppa teaches

| cuppa | For MonoSmith |
|---|---|
| Three panes: component palette on the left, canvas in the centre, properties on the right | A navigator (workspace, projects, tasks), a main view, and an inspector for what is selected |
| A menu bar, with undo and redo | A menu bar that lists every command, with its shortcut; nothing is hidden behind a key you must know |
| Mouse first: click to select, drag to resize, shift-click for several | Click to select, double-click to run, right-click for a context menu, drag the splitters, scroll the logs, hover for help |
| Built on Bubble Tea, Lip Gloss and Bubbles, and also shipped as a desktop app and a WebAssembly build | smith's shell already uses the same libraries (Bubble Tea 1.3, Lip Gloss, Bubbles); one engine behind a terminal, and IDE views |
| Needs a terminal with mouse support and about 100 columns | The same minimum, with a compact layout below it |

## What exists today

`smith` with no arguments opens a shell with five tabs (Menu, Commands, Doctor,
Release, Settings) and `smith graph --format tui`. It is keyboard driven, one
view at a time, and **never acts by itself**: an action ends as a smith command
line, printed before it runs, and `smith` hands over to that command (see
docs/guides/getting-started.md). It has no mouse support, no live run view and
no panes.

## Principles

1. **Everything visible.** Every function is reachable from a menu or a pane
   without knowing a shortcut. The shortcut is shown beside it.
2. **Mouse and keyboard are equals.** Every mouse action has a key, and every
   key has a visible control.
3. **Real time.** A run shows each task as it starts, its output as it is
   written, cache hits, durations and failures, as they happen.
4. **A command, always.** Whatever the interface does is a smith command,
   shown in the interface and logged, so it can be repeated in a script or CI
   (see the open question below on how this works when actions run inside the
   app).
5. **One engine.** The terminal interface and the IDE plugins call the same
   engine through the same events, so they cannot drift.
6. **Accessible.** Colour is never the only signal, and there is a plain
   output mode for screen readers and for terminals without colour.

## Layout

```
┌ File  Workspace  Run  Projects  Recipes  Plugins  Release  Help ─────────────────────┐
│ ▸ libs (Python, C#)        │ Projects  Run  Doctor  Release  Recipes  Settings        │ Inspector          │
│   ▾ apps                   ├───────────────────────────────────────────────────────────┤ orders-api         │
│     ● orders-api   app     │  orders-api                                               │ kind  app          │
│     ○ web          app     │  Tasks      build   test   lint   package   image          │ stack team-csharp  │
│   ▾ libs                   │                                                           │ recipe team 1.3.0  │
│     ● kit          package │  [▶ Run build] [▶ Run all] [⟳ Affected] [⬇ Install…]       │ deps  kit, core    │
│     ○ core         lib     │                                                           │ affected ● yes     │
│ Plugins                    ├─ Output ───────────────────────────────────────────────── │ last   cached 2m   │
│   node  dotnet  docker …   │ orders-api:build ✔ 4.1s  (cached)                         │                    │
│ Recipes                    │ kit:test        ▶ running  42/60                          │ Actions            │
│   team 1.3.0               │ ...                                                       │ ▶ Run  ⟳ Why miss  │
├────────────────────────────┴───────────────────────────────────────────────────────────┴────────────────────┤
│ $ smith run build test -p orders-api --affected          ✔ 12 passed  ✖ 0  ⏱ 38s     main  smith 1.23  doctor ✔ │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

- **Menu bar** (top): every command, grouped, each with its shortcut. Opens with a
  click or `F10`/`Alt`.
- **Navigator** (left): the workspace as a tree of projects (kind, affected
  dot), their tasks, the plugins in use and the recipes. Selecting an item sets
  the main view and the inspector.
- **Main view** (centre): tabs for Projects, Run, Doctor, Release, Recipes and
  Settings. A bar of buttons runs the common actions for the selection.
- **Output** (centre, bottom): the live run, with a tree of tasks and the log of
  the selected one.
- **Inspector** (right): properties of the selection and its actions, as in
  cuppa's properties panel; where an option can be changed, it is edited here.
- **Status bar** (bottom): the exact command running or last run, the result,
  the branch, the smith version and doctor's status.

Panes can be resized by dragging their borders and hidden from the View menu.
Below 100 columns the inspector folds into the main view, and below 80 the
navigator becomes a tab, so the interface works in a small terminal too.

## Components

| Component | Is | Variants and states |
|---|---|---|
| Menu bar and menus | all commands, grouped, with shortcuts | closed, open, item disabled (with the reason on hover) |
| Tree | workspace, projects, tasks | collapsed, expanded, selected, focused, affected, failed |
| Tabs | the main views | active, inactive, with a badge (failures, findings) |
| Button | run an action | default, hover, pressed, disabled, running (spinner) |
| Task list | tasks of a run | queued, running, cached, passed, failed, skipped |
| Log viewer | output of the selected task | follow, paused, filtered, searched, wrapped |
| Form | a recipe's choices, a command's arguments | pending, valid, invalid with the reason |
| Inspector | properties and actions of the selection | read-only, editable |
| Command palette | find and run any command (`Ctrl+P`) | empty, results, no match |
| Context menu | actions for what was right-clicked | open, closed |
| Toast | a short result or warning | info, success, warning, error |
| Dialog | confirm something that changes the workspace or publishes | open, confirmed, cancelled |
| Status bar | the command line and the result | idle, running, passed, failed |

## Building blocks: the libraries behind Cuppa's palette

Cuppa's palette (79 components in the app) is made of the Charm libraries and a
few community packages. Every one of them is a Go package that smith's shell
can use directly, so the interface is designed in Cuppa and built from the same
parts:

| Library | What it gives | Used for |
|---|---|---|
| **Lip Gloss** | styles, borders, padding, layout (join, place), and its table, list and tree renderers | the panes, borders, tokens, the status bar, static tables and trees |
| **Bubbles** | spinner, progress, viewport, list, table, text input, text area, paginator, file picker, timer and stopwatch, help and key bindings | running indicators, the log viewer, project and task lists, the command palette's input, search boxes, the help footer |
| **Huh** | forms: input, text, select, multi-select, confirm, note | a recipe's choices, command arguments, confirmation dialogs |
| **Glamour** | markdown rendering with themes | showing docs, `smith explain` output, release notes and a recipe's description |
| **ntcharts** | terminal charts: bar, line, sparkline, time series | task durations, cache hit rate over time, release frequency |
| Community components | mouse hit-testing zones, menus, tab bars, tree views | the menu bar, tabs, clickable regions (a zone library gives Bubble Tea the mouse hit-testing it lacks) |

How the components above map to them:

| MonoSmith component | Built from |
|---|---|
| Menu bar, context menu | a community menu, with Lip Gloss borders, and mouse zones |
| Tabs | a community tab bar |
| Tree (workspace, projects, tasks) | Lip Gloss tree for display; a Bubbles list with indentation for the interactive one |
| Task list and progress | a Bubbles table or list, spinners and progress bars |
| Log viewer | a Bubbles viewport with follow, search and filter on top |
| Form (recipe choices, arguments) | Huh |
| Dialog | Huh confirm in a Lip Gloss double border |
| Command palette | a Bubbles text input over a Bubbles list |
| Inspector | a Lip Gloss column of key and value rows; Glamour for long text |
| Charts (durations, cache hits) | ntcharts |
| Status bar | Lip Gloss, with Bubbles help for the key hints |

**Designing it in Cuppa.** Cuppa exports pictures and text, not Go code, so it
is for the design stage: the screens above (the main layout, a run in progress,
a failed run, the recipe form, doctor's findings, the release preview, and the
compact layout) are drawn in it from this palette, saved as `.cuppa` files and
exported as images, and kept in `docs/design/tui/`. The Go interface is built
from the same libraries and checked against those pictures.

## Interaction

| Mouse | Keyboard |
|---|---|
| Click selects, double-click runs the default action (the task, the finding's fix) | Arrows move, `Enter` opens or runs |
| Right-click opens a context menu | `Ctrl+O` |
| Drag a pane border to resize | `Ctrl+Arrows` |
| Scroll the wheel in lists and logs | `PgUp` `PgDn`, `g` `G` |
| Click a tab or a menu | `Tab`/`Shift+Tab`, `F10`, `Alt+letter` |
| Hover is not tracked, so `Shift+drag` (`Option+drag` on macOS) selects text natively; `m` turns the mouse off | `y` copies the focused remedy, command, path or package |
| Click a task in a run to show its log | `Enter` on the task |
| Click `■ Stop` | `Ctrl+C` once stops the run, twice quits; `q` quits when nothing runs, `Esc` never quits |

Focus is always visible. A run can be left in the background while another
view is used, and the status bar keeps showing its progress.

## Design tokens

Colours are **semantic tokens**, never raw values in components:

| Token | Meaning | Also shown as |
|---|---|---|
| `success` | passed, cached, ready | ✔ |
| `warning` | needs attention, warning | ▲ |
| `error` | failed, blocked | ✖ |
| `info` | running, informational | ● ⟳ |
| `muted` | skipped, disabled, secondary text | dim |
| `accent` | selection, focus, the active tab | bold, reverse |

- Themes: dark, light, a 16-colour fallback and no colour (`NO_COLOR`, or
  `--plain`). The terminal's own palette is used where possible, so the
  interface follows the user's theme.
- Spacing is in cells: one-cell padding in panes, two-cell gaps between buttons.
- Borders: rounded for panes, single for the focused pane, double for dialogs.
- Icons have ASCII fallbacks (`[x]`, `[!]`, `[v]`) for terminals without Unicode.
- Motion: one spinner for anything running; nothing else animates.
- **Accessibility:** every state has a symbol and a word, not only a colour; a
  plain mode prints events one per line for screen readers; all controls are
  reachable by keyboard; contrast of the default theme meets WCAG AA.

## Real-time feedback: the event stream

A run emits **events**, and every interface renders them:

```
run.started   { run, tasks, command }
task.queued   { task }
task.started  { task, time }
task.output   { task, stream, line }
task.finished { task, status: passed|failed|cached|skipped, duration, cacheKey }
run.finished  { run, ok, summary }
```

- The engine has these facts today (`smith run` prints them, `--json` summarises);
  what is new is one **documented event stream** with a stable shape. The TUI
  consumes it in process; the IDE plugins and tools consume it as JSON lines or
  through the server below.
- The same stream carries `install`, `release`, `doctor` and `upgrade` progress,
  each as its own run.

## One engine for the terminal and the IDE plugins

Both need the engine to do things and to stream results, so the engine gets a
server mode, in the style of the language server protocol:

- `smith serve` speaks JSON-RPC over stdio: methods for each function (list
  projects, run tasks, install packages, doctor, recipes, release preview, `smith
  explain`), and the event stream as notifications. It is the same protocol
  family the plugins already speak.
- The terminal interface calls the same functions in process; the IDE plugins
  start the engine as a child process and talk to it.
- `smith lsp` (issue #303) is part of the same server: `smith.toml` and recipe
  completion, validation and fixes.

### The plugins use the installed smith

**Decided (2026-10-18, replacing the earlier requirement to embed the engine):**
the plugins **do not carry the engine**. They run the smith the team has
installed, the way the ESLint and Jest extensions run the workspace's own
tools. The reason is control: with an embedded engine the version is whatever
the plugin shipped, which can differ from the one CI and the rest of the team
use. With the installed one, the IDE, the terminal and CI run the same smith,
so a task gives the same result, cache key and release preview everywhere.

How the plugin finds smith, first match wins:

1. a setting (`smith.path`), for a smith somewhere unusual
2. the workspace's own install: `node_modules/.bin/smith` (the npm package
   `monosmith` as a dev dependency) or `.smith/bin/smith`, which pins the
   version per repository
3. `smith` on the `PATH`

If none is found, the plugin does not fail silently: it shows what is missing
and the install commands for each channel (npm, pipx, `dotnet tool`, gem, pub,
the install script, a package manager), with a button for the common one.

**Version compatibility.** `smith serve` (see above) reports its version and
the protocol version it speaks. The plugin states the range of protocol
versions it supports:
- older than the range: the plugin offers the upgrade command and limits itself
  to what the old smith can do
- inside the range: everything works
- newer: works if the protocol is compatible, and the plugin says when it is
  not, since additions to the protocol are optional fields as for plugins

**Updates** of smith are the team's decision, made in the repository (the
dev dependency, or the version pinned in CI), not by the plugin updating
itself. The plugin updates through its marketplace and only needs a smith in
its supported range.

**Not chosen:** embedding the binary in each extension package (platform
packages for VS Code, bundled binaries for JetBrains) was the earlier plan and
remains possible later as a **fallback** for a machine with no smith, but not
as the default, because it would let the IDE run a different version from the
rest of the team. A WebAssembly build of the engine (as cuppa has) is also not
the plan: the engine runs processes and reads the file system, which a browser
sandbox does not allow.

### What the IDE shows

The same functions as the terminal interface, in the editor's own style:

| Area | Terminal | IDE |
|---|---|---|
| Workspace | Navigator tree | Side-bar view of projects, tasks, plugins and recipes |
| Run | Output pane | Output panel with a task tree; problem matchers turn failures into editor problems |
| Install | A form in the main view | A command and a view; a package picker for the project's ecosystem |
| Doctor | Doctor tab | Problems panel, with fixes as quick fixes on the files they concern |
| Recipes | Recipes tab and wizard | New project from a recipe, with the choices as a form |
| Release | Release tab | A release preview view and the command |

## Reconciling "never acts by itself"

Today the shell hands over to `smith <command>` and exits the interface. The
requested behaviour runs everything **inside** the interface. This keeps the
principle that matters, that every action is a smith command the user can see
and repeat, by changing how: the interface **starts the command itself in a
pane**, shows its exact command line in the status bar and in the log, and
confirms before anything that publishes or changes the workspace. It still
never does something a command line could not.

**Open:** confirmation for local runs. Recommendation: running tasks, doctor
and previews need none; installing, adding or removing a project, upgrading
and releasing show a dialog with the command first.

## Delivery

1. **The event stream and the server mode** (`smith serve`), with the run
   events. Foundation for everything below.
2. **The layout and the run view** in the terminal: panes, menu bar, live task
   list and log, mouse support (Bubble Tea mouse events with a zone library).
3. **Install, doctor and release inside the interface**, with dialogs and the
   command line shown.
4. **Recipes in the interface**: the choices as a form, the creation wizard.
5. **The VS Code extension** on `smith serve`, finding the installed smith.
6. **The JetBrains plugin.**

## Open questions

- **Run inside the app, with confirmation:** as above.
- **Pinning smith in a repository:** the npm dev dependency `monosmith` pins it
  for a Node workspace, and the install script takes a version for CI. A
  repository with no Node has no equivalent. A `smith` version range in
  `smith.toml` (`[workspace] smith = "~1.24"`), checked by `smith doctor` and
  `smith ci check` and read by the IDE plugins, would give every repository one
  place to say which smith it uses. A decision for the product owner.
- **Minimum width:** 100 columns with a compact layout below. Is 80 enough as
  the floor?
- **Mouse zone library:** Bubble Tea's own mouse messages need hit-testing; a
  small library for zones is the usual answer. A choice for the implementation.
- **Who draws the screens in Cuppa:** the product owner knows what the
  interface should feel like; the screens listed above are the brief. Cuppa's
  file format is specified (`docs/spec/cuppa-format.md` in its repository), so
  a first draft of each can also be generated as a `.cuppa` file to refine by
  hand.
- **A desktop app:** cuppa also ships one. Not planned; the IDE plugins cover
  the graphical case.
