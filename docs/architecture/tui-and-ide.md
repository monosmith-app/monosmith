# Design: the TUI and the IDE plugins

Status: **proposal, partly built**. It extends the TUI issue (#302) and the IDE plugins issue
(#303). Read the status box first; the sections below keep the original argument, and where
they disagree with the box, the box is right.

> **As built, October 2026 (smith 1.62.0), checked against the code**
>
> | Part of the proposal | State |
> |---|---|
> | Terminal shell with Home, Projects, Run, Doctor, Deps, Release, Settings | built; one keymap, no menu bar ([ADR 0032](decisions/0032-single-key-tui-keymap.md), [architecture](tui-architecture.md)) |
> | Navigator tree and permanent inspector | not built; replaced by a list with a detail pane beside it (from 100 columns) |
> | Right-click context menu | not built; removed with the menu bar in the redesign (ADR 0032) |
> | Drag a pane border to resize | not built; `[` and `]` make the detail narrower or wider |
> | Huh, Glamour and ntcharts | not used; the shell uses Bubble Tea, Bubbles and Lip Gloss |
> | The `.cuppa` design files in `docs/architecture/tui/` | not in the repository |
> | Event stream and `smith serve` | built ([ADR 0026](decisions/0026-event-stream-and-smith-serve.md), [Serve](../reference/serve.md)); the shell's Run tab does **not** consume it: it parses smith's text output (`tui/runview`) |
> | Install, doctor and release inside the interface | built as commands the shell starts; task runs and printing commands use the real terminal (`tea.Exec`) |
> | Recipes in the interface | partly: the Settings tab lists the workspace's recipes and whether each is valid; `a` opens a guided form, and the recipe's choices are asked by `smith add` |
> | Doctor fix wizard and adopt wizard | built ([ADR 0033](decisions/0033-doctor-fixes-with-steps-and-choices.md), [ADR 0034](decisions/0034-adopt-and-convert-depth-levels.md)) |
> | VS Code extension on `smith serve`, using the installed smith | built ([ADR 0025](decisions/0025-ide-plugins-run-the-installed-smith.md), [ADR 0030](decisions/0030-vscode-extension-is-a-serve-client.md), [architecture](vscode-architecture.md)); the setting is `monosmith.path` |
> | `smith lsp` (completion and validation of `smith.toml` and recipes) | **planned, not built** (issue #303); there is no such command |
> | JetBrains plugin | **planned, not built** (issue #303) |
> | A `smith` version range in `smith.toml` | not built; still an open question below |
> | Keys | the table under "Interaction" is a proposal; `Tab` and `Shift+Tab` move between buttons in dialogs and wizards, not between panes. The authoritative list is [the manual](../manual/tui.md) |

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
| Three panes: component palette on the left, canvas in the centre, properties on the right | Proposed: a navigator, a main view and an inspector. Built: a list with the detail of the selection beside it (see "As built") |
| A menu bar, with undo and redo | No menu bar (a terminal is not a desktop): every action is a single key, listed in a footer that changes with the view, and the command palette lists the rest |
| Mouse first: click to select, drag to resize, shift-click for several | Click to select, double-click to run, right-click for a context menu, drag the splitters, scroll the logs, hover for help |
| Built on Bubble Tea, Lip Gloss and Bubbles, and also shipped as a desktop app and a WebAssembly build | smith's shell already uses the same libraries (Bubble Tea 1.3, Lip Gloss, Bubbles); one engine behind a terminal, and IDE views |
| Needs a terminal with mouse support and about 100 columns | The same minimum, with a compact layout below it |

## What exists today

`smith` with no arguments opens a shell with seven tabs (Home, Projects, Run,
Doctor, Deps, Release, Settings) and `smith graph --format tui`. It is keyboard and mouse
driven, one view at a time, and **never acts by itself**: an action ends as a
smith command line, printed before it runs, and a task run uses the terminal
itself (see docs/getting-started/first-workspace.md and "As built" below). The
Settings tab is where plugins, the colour theme and the settings of `smith.toml` are
changed without leaving the interface (see "What the Settings tab does" below); the
guided forms of the old Guide tab are its last section.

## Principles

1. **Everything visible.** Every function is reachable from the command palette,
   a pane or a key the footer shows, without knowing a shortcut beforehand. A
   label is drawn from the same list as the key it names, so it cannot go stale.
2. **Mouse and keyboard are equals.** Every mouse action has a key, and every
   key hint in the footer can be clicked.
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

The drawing is the original proposal: a navigator tree on the left and an inspector on the
right. It was not built that way; "As built" below is what the shell is.

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│ ▸ libs (Python, C#)        │ Home  Projects  Run  Doctor  Deps  Release               │ Inspector          │
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
│ b build  t test  l lint  c typecheck  w changed  d doctor            1-6 views  : palette  ? help  q quit │
│ $ smith run build test -p orders-api --affected          ✔ 12 passed  ✖ 0  ⏱ 38s     main  smith 1.23  doctor ✔ │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

- **Navigator** (left): the workspace as a tree of projects (kind, affected
  dot), their tasks, the plugins in use and the recipes. Selecting an item sets
  the main view and the inspector.
- **Main view** (centre): tabs for Home, Projects, Run, Doctor, Deps and
  Release. The keys that run the common actions for the selection are in the
  footer.
- **Output** (centre, bottom): the live run, with a tree of tasks and the log of
  the selected one.
- **Inspector** (right): properties of the selection and its actions, as in
  cuppa's properties panel; where an option can be changed, it is edited here.
- **Footer** (second row from the bottom): the keys of the open view, then the
  keys every view ends with (`1-7` views, `m` mouse, `:` palette, `?` help,
  `q` quit). Only what can be done now is listed, and a hint that is one key
  can be clicked.
- **Status bar** (bottom): the exact command running or last run, the result,
  the workspace, the branch and the smith version.

Panes can be resized by dragging their borders.
**As built (the adaptive shell).** The navigator tree and the permanent
inspector are gone. The screen is a header (workspace, branch, doctor's health
as `✖2 ▲2`), the views as rounded tabs whose active tab opens into the pane
below, one content pane, the footer and the status bar. From 100 columns the
pane is the list and, beside it, the detail of the selection (a project's
tasks, dependencies and actions, a doctor finding's fix, a run's task); below,
only the list, Enter opens the detail over it and Esc goes back. The tabs shorten
their names and then scroll (`‹` `›`) so the open one is always in view, and are
one row on a terminal under 24 rows high (seven tabs fit the 72-column minimum by shortening the longest names to
`Proj`, `Rel` and `Sett`, so they do not scroll there). Views implement a small `Page` (title, footer keys,
update, view for a width and a height).

**Pages.** Each page leads with its headline, then the detail.

- **Home** is the first tab: one headline for the workspace's health ("2 checks
  fail"), the one suggested next action with its key ("press d", "press w to
  lint and test them", "all good: press b to build"; Enter or a click does it),
  and the facts behind them: checks, projects affected, how the last run ended,
  dependencies with newer releases (known once Deps has looked). Doctor starts in
  the background with the shell, so the header badge is filled without opening
  Doctor.
- **Doctor** shows "2 failed · 2 warnings", then the failing checks as one line
  each (state, title); passed checks stay behind `p`. Beside the list, the detail
  tells the selected check in full (problem, remedy, fix: files as diffs, command
  steps, the questions a wizard asks; a fix folded from many projects says
  "applies to 22 projects") and Enter is not
  offered; in the single pane Enter expands the check in place. Each part is said
  once. A check with a fix shows "Fix: <description>" (`f`; "Fix all N projects"
  for a folded fix) and opens a wizard dialog: the fix's questions as numbered
  options, a scrollable review (files as diffs, commands, the projects covered),
  then the apply with its progress; a failing step shows its output with Retry
  and Close, and a success re-runs doctor. `x` and the palette's "Fix with
  doctor" open the wizard on the fixes that need no decision, so no fix is
  applied without a review; nothing here runs in plain mode, where `smith doctor
  --fix` asks the same questions as numbered prompts.
- **Deps** and **Release** keep their behaviour: a headline ("3 with newer
  releases: 1 patch · 1 minor · 1 major", "2 packages to release (1 minor · 1
  patch)"), then the rows; long names are shortened in the middle. The footer
  lists only the keys that work now (no `u update` with nothing to update, no
  `x release` without a plan; the key says why when pressed).

**What the Settings tab does.** The last tab (key `7`) is a list of sections on the left
and the editor or detail of the open section on the right; `Tab` moves between the two.
It calls the same functions the commands call, in the process, so a change leaves
`smith.toml` byte for byte as the command would (a test compares them).

| Section | What it does | The command it stands for |
|---|---|---|
| Plugins | Switch the built-in language plugins on and off (`Space`); see the external plugins the workspace declares with their trust and pin; `t` shows the same review `smith plugin trust` shows (what it is, the files it runs with their sha256, the environment it sees) and trusts it (`p` also pins the sha256 for CI), `f` forgets the decision; the plugins installed on the search path; and a marked row for installing a plugin, a hint (`smith plugin install <name\|path>`) until `smith plugin install` exists in this smith | `smith plugin list`, `enable`, `disable`, `trust`, `untrust` |
| Style | The theme (auto, dark, light, plain) and the symbols (auto, unicode, ascii) with a preview of the highlighted choice; `Enter` applies it at once and keeps it for you in `$SMITH_HOME/tui.json` (`~/.smith/tui.json`), never in `smith.toml`. `SMITH_THEME`, `NO_COLOR`, `SMITH_ASCII` and `TERM=dumb\|linux` still win, and the section says so when one is set | the environment variables of "Themes, plain mode, mouse" |
| Workspace | Every setting `smith config list` knows, each with an editor for its type: a switch flips, a fixed list (`ci.providers`) is ticked, text and numbers are typed and checked against the schema; what it does and its default are shown; `u` resets it to its default | `smith config list`, `set`, `unset` |
| Recipes | The `*.recipe.toml` files of the workspace with their name, version and whether each is valid | `smith recipe` (`smith recipe check <file>` shows what a recipe would use) |
| Forms | The guided forms (add a project, remove a project...) that end as a smith command line | the guided forms |

The palette lists one **Settings** entry; a `smith plugin`, `plugin list`, `recipe` or
`config list` typed in the palette opens the tab on the matching section instead of leaving
the terminal. Home points a new user at the tab until it has been opened once (kept in
`tui.json` as `settingsTipSeen`). Not in the tab: `smith recipe new` (its wizard asks on the
terminal), and `smith plugin set`, which the Plugins section does one plugin at a time.

## Components

| Component | Is | Variants and states |
|---|---|---|
| Footer hints | the keys of the open view | a hint per key; clickable; absent when the key cannot be used now |
| Tree | workspace, projects, tasks | collapsed, expanded, selected, focused, affected, failed |
| Tabs | the main views | active, inactive, with a badge (failures, findings) |
| Task list | tasks of a run | queued, running, cached, passed, failed, skipped |
| Log viewer | output of the selected task | follow, paused, filtered, searched, wrapped |
| Form | a recipe's choices, a command's arguments | pending, valid, invalid with the reason |
| Inspector | properties and actions of the selection | read-only, editable |
| Command palette | find and run any command (`:`, `/` or `Ctrl+P`) | empty, results, no match |
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
| Community components | mouse hit-testing zones, tab bars, tree views | tabs, clickable regions (a zone library gives Bubble Tea the mouse hit-testing it lacks) |

How the components above map to them:

| MonoSmith component | Built from |
|---|---|
| Footer hints | Lip Gloss text, with mouse zones |
| Tabs | a community tab bar |
| Tree (workspace, projects, tasks) | Lip Gloss tree for display; a Bubbles list with indentation for the interactive one |
| Task list and progress | a Bubbles table or list, spinners and progress bars |
| Log viewer | a Bubbles viewport with follow, search and filter on top |
| Form (recipe choices, arguments) | Huh |
| Dialog | Huh confirm in the same rounded Lip Gloss border as a pane |
| Command palette | a Bubbles text input over a Bubbles list |
| Inspector | a Lip Gloss column of key and value rows; Glamour for long text |
| Charts (durations, cache hits) | ntcharts |
| Status bar | Lip Gloss, with Bubbles help for the key hints |

**Designing it in Cuppa.** Cuppa exports pictures and text, not Go code, so it
is for the design stage: the screens above (the main layout, a run in progress,
a failed run, the recipe form, doctor's findings, the release preview, and the
compact layout) are drawn in it from this palette, saved as `.cuppa` files and
exported as images, and would be kept in `docs/architecture/tui/` (the folder does not exist). The Go interface is built
from the same libraries and checked against those pictures.

## Interaction

No key needs a modifier except `Ctrl+C` (stop, then quit) and the palette's
`Ctrl+P`; there is no menu bar, no Alt chord and no function key to learn. The
keymap is one list (`apps/smith/tui/commands/keymap_config.go`), and the keys, the
footer hints, the help (`?`, `F1`) and the palette entries are all drawn from it.
A key that works only in one view is listed in that view's footer; when the same
letter is a global key, the view's own wins (Deps marks all with `a`, which adds a
project elsewhere). No plain key fires while a text field or a guided form has the
keyboard; `Ctrl+P` and `F1` still do.

| Key | Does | Where |
|---|---|---|
| `b` `t` `l` `c` | build, test, lint, typecheck the selected project (all when none) | workspace |
| `w` | lint and test what changed (`smith run lint test --affected`) | workspace |
| `d` | doctor | workspace |
| `i` | install dependencies (asks first) | workspace |
| `a` | add a project | inside a workspace |
| `n` | new workspace | outside a workspace |
| `1`..`6` | open a view: Home, Projects, Run, Doctor, Deps, Release | workspace |
| `Tab`, `Shift+Tab` | next, previous pane | anywhere |
| arrows, `j` `k`, `PgUp` `PgDn`, `Home` `End` | move | lists, trees, logs |
| `Enter` | open, run or edit the selection | anywhere |
| `Esc` | back, close, clear; never quits | anywhere |
| `y` | copy the focused remedy, command, path or package | Projects, Run, Doctor, Deps |
| `m` | turn the mouse off, so the terminal selects text (`Shift+drag` does too) | anywhere |
| `:`, `Ctrl+P`, `/` | command palette | anywhere |
| `?`, `F1` | help for the open view | anywhere |
| `q` | quit when nothing runs | anywhere |
| `Ctrl+C` | stop the run; quit when nothing runs | anywhere |
| `f` `x` `g` `p` `r` | Doctor: open the fix wizard for this check, for every fix that needs no decision, group by project, show passed, check again | Doctor |
| `p` `n` `x` | Release: preview, notes, release (asks first) | Release |
| `Space` `s` `a` `u` `v` | Deps: mark, mark the safe ones, mark all, update the marked, align versions | Deps |
| `Enter` | Home: do the suggested next action | Home |
| `f` `r` | Run: follow or pause the log (when the output is captured), run the last command again | Run |

| Mouse | Keyboard |
|---|---|
| Click selects, double-click runs the default action (the task, the finding's fix) | Arrows move, `Enter` opens or runs |
| Click a view name in the tab row | `1`..`6` |
| Click a hint in the footer | the key it shows |
| Drag a pane border to resize | not built |
| Scroll the wheel in lists and logs | `PgUp` `PgDn` |
| Hover is not tracked, so `Shift+drag` (`Option+drag` on macOS) selects text natively; `m` turns the mouse off | `y` copies the focused remedy, command, path or package |
| Click a task in a run to show its log | `Enter` on the task |

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

- Themes: dark, light and plain (no colour). `NO_COLOR` (any value) and
  `SMITH_THEME=plain` give plain; `SMITH_THEME=light` or `dark` forces one;
  with none of them the terminal decides (`COLORFGBG`, dark when unknown). The
  light tokens keep 4.5:1 contrast on white, checked by a test. In plain, the
  selection is bold and reverse video, and every state keeps its symbol and
  word.
- Spacing is in cells: one-cell padding in panes, two-cell gaps between buttons.
- Borders: one weight, rounded, for panes, dialogs and overlays. Only the colour
  changes: the focused pane or the open dialog has the accent border, the others
  the faint one.
- Colour is for state (success, warning, error, info, muted) and for one accent:
  the active tab, the cursor and the focused border. Titles, key labels and
  actions are bold, not coloured. Paths are shortened in the middle, and prose
  wraps to the pane's width.
- Icons have ASCII fallbacks (`[x]`, `[!]`, `[v]`) for terminals without Unicode.
- Motion: one spinner for anything running; nothing else animates.
- **Accessibility:** every state has a symbol and a word, not only a colour; a
  plain mode asks the questions as numbered lines for screen readers (`SMITH_SCREEN_READER=1`,
  `SMITH_THEME=plain`, or output that is not a terminal); all controls are
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

1. a setting (`monosmith.path` in the VS Code extension), for a smith somewhere unusual
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
| Doctor | Home and Doctor tabs | Problems panel, with fixes as quick fixes on the files they concern |
| Recipes | `smith recipe`, `smith add` (the palette lists them) | New project from a recipe, with the choices as a form |
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
2. **The layout and the run view** in the terminal: panes, key footer, live task
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
  hand. (Not done: see the status box.)
- **A desktop app:** cuppa also ships one. Not planned; the IDE plugins cover
  the graphical case.
