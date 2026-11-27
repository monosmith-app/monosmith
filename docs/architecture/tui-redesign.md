# Design: the TUI redesign

Status: **built**. Phase 1 (theme and chrome), phase 2 (layout), phase 3 (Home),
phase 4 (Projects), phase 5 (Runs), phase 6 (the palette and Health), phase 7 (Dependencies and
Release) and phase 8 (Settings, the help and the dialogs) are built. The VS Code extension has the
same views ([VS Code as built](#vs-code-as-built)). What the redesign left open is listed in
[Left open](#left-open), and what was built of it since in [Follow-up as built](#follow-up-as-built).
This page holds the owner's decisions, the colour tokens, the phases and the drawn screens.
The earlier design and its "as built" box are in [tui-and-ide.md](tui-and-ide.md), and the
code layout is in [tui-architecture.md](tui-architecture.md).

## Screen sizes and pieces

A wide screen is 150 x 44 cells and a narrow one 80 x 30. Every screen is drawn only with
Lip Gloss and Bubbles pieces (Bubble Tea 1.3, Bubbles 1.0, Lip Gloss 1.1).

## Decisions

The owner made these decisions for the whole redesign. Later phases do most of them.

| Decision | Phase |
|---|---|
| The palette is the MonoSmith one (table below). Plain mode does not change. | 1 |
| The title bar has the ` smith ` tag, the workspace in bold, the branch muted, the palette hint with its key and the smith version at the right. | 1 |
| The footer is the short style of `bubbles/help`: `key desc • key desc`, the key in the accent and bold, the description muted, the separator in the border colour. | 1 |
| The status bar has the last command (muted), the result counts with their symbols, the time, and the branch, the version and the doctor state at the right. | 1 |
| A wide screen has a rail of views at the left. A narrow screen has a one-line tab strip. One view shows at a time: `Enter` opens, `Esc` goes back. | 2 |
| A Runs view on key `3`; the other views move up one (Health `4` to Settings `7`). It replaces the run panel of Projects and the older "no Run tab" rule. | 5 |
| The Doctor view is renamed Health, in the TUI and in the VS Code extension. The `smith doctor` command keeps its name. | 2 |
| The doctor fix shows inline in the detail (review, apply, check again), not as a dialog. | 6 |
| One pull request for each phase. | all |

## Tokens

The tokens are in `apps/smith/features/tui/theme/tokens_config.go`. They are hex colours, and
Lip Gloss makes them fit the terminal's colour profile. Each text colour keeps a contrast of
4.5:1 on the terminal's background, the selected row (Highlight), the row under the pointer
(RowHover) and the bars (Bar). `tokens_config_test.go` checks this.

| Token | Dark | Light | Use |
|---|---|---|---|
| Accent, Focus, Info | `#5fb6d1` | `#176a86` | selection, focus, the active tab, running |
| Success | `#7ccf96` | `#1b7537` | passed, cached, ready |
| Warning | `#e3c16f` | `#855600` | needs attention |
| Error | `#ff9a90` | `#b3261e` | failed, blocked |
| Text | `#d0d0d0` | `#2b2b2b` | normal text |
| Muted | `#b0b0b0` | `#5c5c5c` | secondary text, the footer descriptions, the last command |
| Faint | `#a5a5a5` | `#686868` | hints: a dimmer grey that still has 4.5:1 on the selected row |
| Highlight | `#2f3d44` | `#dcebf1` | background of the selected row |
| RowHover | `#263238` | `#eef5f8` | background of the row under the pointer, between the terminal's background and Highlight |
| Bar | `#303030` | `#f2f2f2` | background of the title bar and the status bar |
| Border | `#5a5a5a` | `#b8b8b8` | rules, pane borders, the footer separator |
| Primary button | Accent background, text `#141414`, bold | Accent background, text `#ffffff`, bold | `ButtonActive` and `ButtonOn`; also the ` smith ` tag |
| Secondary button | Highlight background, Text | Highlight background, Text | `Button` and `ButtonText` |
| Button under the pointer | `#3d5560` | `#a9cfdd` | `ButtonHover`, between the secondary and the primary background |

The palette gives `#1d7a3a` for the light Success. That colour has 4.4:1 on the light
Highlight, so the token is `#1b7537`, one step darker, with 4.7:1.

## Phases

1. Theme and chrome: the tokens, the title bar, the footer and the status bar. The layout
   keeps the top tabs. From 120 columns the header stays at the right end of the tab row
   without the palette hint and the version. Phase 2 gives the title bar its own row there.
2. Layout: the rail of views on a wide screen, the tab strip on a narrow one, one view at a
   time. **Built** (see [Phase 2 as built](#phase-2-as-built)).
3. Home. **Built** (see [Phase 3 as built](#phase-3-as-built)).
4. Projects. **Built** (see [Phase 4 as built](#phase-4-as-built)).
5. The Runs view. **Built** (see [Phase 5 as built](#phase-5-as-built)).
6. The command palette, and Health with the inline fix. **Built** (see [Phase 6 as built](#phase-6-as-built)).
7. Dependencies and Release. **Built** (see [Phase 7 as built](#phase-7-as-built)).
8. Settings, the help and the dialogs. **Built** (see [Phase 8 as built](#phase-8-as-built)).

## Phase 2 as built

- The wide layout starts at `layout.WideFrom`, 100 columns, the width from which the detail
  already sat beside the list. Row 0 is the title bar at every width, with the palette hint
  (`Ctrl+K`) and the version at the right. From 20 rows a wide screen keeps a blank row under
  the title bar and one over the footer, as the screens show.
- The rail (`rail` package) is 24 columns in a rounded frame titled `Views`, with a blank row
  between the views when they fit (none on a short screen). It lists today's views: Home,
  Projects, Health, Dependencies, Release and Settings, on keys 1 to 6. There is no Runs row
  (phase 5) and no Pinned group, because the shell has no pinned projects.
- The badges: Projects carries the run (spinner, `×n`, `✓`), Health `×n` or `▲n`, Dependencies
  `↑n` once the registries answered, Release `✓n` when a plan can be released. A view that
  loads shows a spinner.
- The narrow strip is `1 Home  2 Projects  3 Health ×2  4 Deps …`; Dependencies is `Deps`
  there, and the strip fits 80 columns.
- Every pane is a rounded frame with its title in the top border and one cell of padding. The
  list's right border is the divider the mouse drags. Inner dividers (Settings, the run panel)
  meet the frame in `┬`, `┴` and `├`. The detail keeps at most 68 columns (it was 64), so its
  text keeps 64 columns inside the frame.
- On a narrow screen Enter opens the detail of Projects and Health in the frame, titled
  `‹ Projects / orders`, and Esc goes back. Health used to expand a check in place there.
- The footer is one run, `key desc • … • ? help • q quit`; the global hints go first when it
  does not fit, then the view's last ones, then help and quit.
- `Ctrl+K` opens the palette, as `:` and `Ctrl+P` do.

Deviations from the brief:

- The symbols stay `×` and `▲`. The design draws `✗` and `⚠`, but
  `apps/smith/shared/marks` bans both in source (they overlap the next cell in common
  fonts, and `⚠` has an emoji form), and its test guards that. The owner decides whether
  to change that rule; the ASCII set is unchanged.
- Deps, Release and Settings keep their narrow behaviour: Deps and Settings show their detail
  under the list, and Release has no detail. Enter already has a meaning in those lists.
- The VS Code extension already names the tree Health; nothing changed there. Its
  `Run Doctor` command and its `Doctor:` messages are about the `smith doctor` command.

## Phase 3 as built

- Home is drawn by the `dashboard` package (`dashboard.Render`). The page is wide from
  `dashboard.WideFrom`, 88 cells of the pane, which a 120-column terminal gives. A narrower
  pane has the narrow look of the NarrowHome screen, also inside the wide layout from 100 to 119
  columns.
- The top line is the workspace in bold, then, muted, `N projects (languages) · N changed since
  <base> · last run <ago>`. The base is the ref the shell counts the affected projects from
  (`ProjectsLoaded.Base`). A part without facts is left out, and the languages go when the
  line is too long.
- The four boxes are `pane.Frame` boxes joined with `lipgloss.JoinHorizontal`, in one row (five
  lines) or a 2x2 grid (four lines, without the sub-line, as NarrowHome draws them). The row
  fills the pane: the first boxes take the cells the division leaves. A box that loads shows the
  spinner and `Loading…`, the waiting style of the other views. A box without facts says so
  (`Nothing ran yet`, `Nothing changed`, `Nothing ready`, `No projects yet`).
- The Last run box tells the newest run of the session and opens it on a click, in the run
  panel of Projects. It replaces the line of recent runs, which the design does not draw.
- "Needs you now" is the queue of `dashboard.Queue`, unchanged in its items, actions and keys,
  "All good" included. The count is `1/5`. Each item has two lines; a wide page puts a blank line
  between two items, a narrow one none, as the screens draw them. The tag is the kind of the item
  (`Failed`, `Warning`, `Changed`, `Release`, `Updates`, `Cache`, `Setup`, `Good`), in the colour of
  its severity.
- The keyboard is on an item (`Up`, `Down`, `Home`, `End`) and on one of its buttons (`Left`,
  `Right`). That button is primary and shows `(Enter)`; the others are secondary. The footer is
  `↑/↓ move • Enter <the label of that button> • …`, through the new `commands.State.Labels`.
- Start something is `Build all`, `Lint all`, `Run a task…` (the palette, `Ctrl+K`),
  `Add a project`, `Compose a stack` and `Settings`. These are the actions with a key on Home today.
  It shows on a wide page only. A workspace without projects leaves it out, because its one
  item already offers the two buttons that add a project. The settings tip ends the page.
- A click on a line of an item selects it, a click on a button or on the key of a narrow item
  runs it, and the wheel scrolls the page. The page scrolls as one, as before.

Deviations from the brief and the screens:

- The marks stay `×` and `▲` (see phase 2). The changed item keeps the warning mark of its
  severity; the screens draw `●`.
- There is no "Build changed" command, so Start something has `Build all`. `Run a task…` has the
  palette key `Ctrl+K`, because `r` reloads the view. Settings has its key `6` (no Runs view yet).
- The screens show the button labels without parentheses (`See why enter`); the buttons keep the
  `label (key)` form of every button of the TUI.
- An item without a note keeps its second line empty, so every item takes two lines.

## Phase 4 as built

- The `projectpage` package draws the view from plain values (`ListHead`, `ListBody`, `Legend`,
  `Head`, `TaskTable`, `RunList`, `Uses`); the shell gives it its state
  (`shell_projects_page_mapper.go`) and handles its keys and clicks
  (`shell_projects_page_use_case.go`).
- The list is the filter line, the muted count line, the projects under muted bold headings with
  their count (`APPS 2`, `LIBS 1`, `PACKAGES 1`, then other kinds in capitals) and the legend on the
  last row. A row is the last-run mark, the name, the language (the kind and the version too when
  the list has the room) in muted text and `● changed` in the accent. The selected row is `▶ ` on
  the Highlight background with the name in the accent and bold.
- The filter is new: the shell had none. `/` on the list types it, Enter keeps it, Esc clears it.
  Its words are `kind:<prefix>`, `lang:<prefix>`, `changed` and any text of the name, kind or
  language (`projectlist.Matches`). The frame title is `Projects 7/42` while a filter is set, and a
  reload keeps the filter.
- The mark is the worst state of the project's tasks in the newest run of the session that ran one
  of them: `✓` passed or cached, `×` failed, `–` not run (`Icons.NotRun`, `-` in ASCII), `»`
  skipped, `⊘` stopped, the spinner while it runs.
- The list takes two fifths of the content width, from 36 to 44 columns; the detail takes the rest,
  81 columns at 150. A run beside the list still takes half, and a divider the person moved wins.
- The detail has a top that never scrolls: `path · kind · language · stack · version` muted, with
  `● changed`, and the buttons `Run… (R)` (primary) and `Open` at the right (on the next line when
  they do not fit), a blank line, the tabs and a rule. `Open` opens the folder with
  `sources.Open`; it is left out when the shell has no such source.
- Run… shows the Run… menu in the detail: the action panel the detail had before (the task buttons,
  the packages, Explain, Why, Graph, the derived files, and under More the settings, the forms,
  Change kind, Eject and Remove). Esc closes it. The palette lists the same actions of the selected
  project (`How was this resolved? (orders)`).
- The tabs are Tasks, Dependencies (Deps on a narrow screen), Runs and Settings. Tab, Shift+Tab,
  Left and Right change them while the detail has the keyboard, and a click on one shows it.
- Tasks is the table Task, Last result (Result under 64 columns), Time, Cache and Key, with a header
  and a rule as `bubbles/table` draws them. Enter or a double click runs the task as its key does. A
  `start` task says `own terminal` and a line under the table explains it.
- Dependencies has the boxes Uses and Used by (two rows under 60 columns), the buttons Install a
  package and Remove a package, and the derived files, which the detail showed before.
- Runs lists the runs of the session that ran a task of the project: mark, tasks, result, time and
  how long ago. Enter or a double click opens one in the run panel.
- Settings is the Project section of `settingsview`, drawn alone (`ProjectPane`, `ProjectKey`,
  `ProjectClick`). The Settings tab no longer lists it: opening the tab forgets the project
  (`CloseProject`). While the section has the keyboard its letters are its own (`t`, `f`, `u`).
- The footer on the list is `Tab detail • / filter • R run… • b build orders • …` (`Enter detail` on
  a narrow screen); in the detail it is `Tab section • Enter run • Esc back • R run… • …`, with
  `Enter open` in Runs and `Enter change` in Settings. The new state is `State.ProjectDetail` with the
  command fields `NeedsProjectDetail` and `NotInProjectDetail`.

Deviations from the brief and the screens:

- The marks stay `×` and `–`; the screens draw `✗`, which `shared/marks` bans.
- The shell had no filter and no multi-select. The filter is new. Space selects nothing: running
  several chosen projects together is not built.
- The `↑ ↓` columns of dependency counts are gone: the Dependencies section names them.
- The row of workspace buttons (New project, Verify, Build, Test, Lint, Typecheck, Build release,
  Clear cache) left the list. Their keys (`a`, `V`, `B`, `T`, `L`, `C`), the footer and the palette
  keep them. An empty workspace still shows New project and Compose a stack.
- The block of recent runs under the list is gone: the runs of a project are its Runs section, and
  Home has the last run. The commands `run.recent-focus`, `run.recent-move` and `run.open-recent`
  went with it.
- Run… has the key `R` (shift+r): `r` is already Replace the files I edited. Open has no key: `o` is
  Why and `O` is Own the files I edited.
- Enter on the wide list moves the keyboard to the detail; it does not run the task, as the footer of
  the screen (`enter run test`) draws.
- The Tasks section shows the table only. The screens also draw Uses, Used by and Recent runs under
  it; those are the Dependencies and Runs sections.
- The Cache column says `miss`, without the reason: the run output does not give it to the shell.

## Phase 5 as built

- The rail and the strip list Home `1`, Projects `2`, Runs `3`, Health `4`, Dependencies `5`,
  Release `6` and Settings `7`. The footer, the help, the palette (Go to: Runs) and the Home
  buttons (`Open Health (4)`, `Settings (7)`) follow the new keys. Only the views that can carry a
  badge (Runs, Health, Deps, Release) keep its room in the strip, so the seven names fit 80
  columns until three badges show; then Settings is the first to shorten (`7 Sett`).
- The run badge (spinner, `×n`, `✓`, `»` for a dry run, `⊘`) moved from Projects to Runs.
- The `runpage` package draws the parts from plain values: `Tree` (the tasks grouped by project,
  in the order they were first seen, `Last` closing the branch), `TaskRow`, `HeadingLine`,
  `AllOutputRow`, `SessionRow`, `Bar`, `Switch` and `Empty`. The shell measures and composes them
  (`shell_runs_page_mapper.go`) and handles the keys and the wheel (`shell_runs_page_use_case.go`,
  `runKey`).
- A run started from any view (a task key, the Tasks section or the Run… menu of a project, Home,
  the palette) opens the Runs view (`revealRun`). A Release preview still stays on Release, as
  before. Esc shows all output, then goes back to the view the run was opened from (`runFrom`).
- The header is the state and the time (`× Failed 4 tasks  12s`) with the buttons at its right end
  when they fit, else on the lines under it; the history line `run 2 of 5 · ‹ older  newer ›` when
  the session has more than one run; a blank line; the bar of `bubbles/progress` with `4/4 100%`
  (a stopped run keeps the neutral bar of `progress.DrawCancelled`); the counts with their marks.
  `bubbles/progress` draws without colour (`termenv.Ascii`) and the theme colours the cells, so the
  plain theme stays plain.
- The rule under the header is a fixed divider (`splitter.RunsRuleID`, never selected or moved),
  so it runs through the padding to the frame. The tree and the log are divided by
  `splitter.RunsTreeID`, which the mouse and the keys move; it meets the rule in `┬` (a new case of
  `splitter.Junctions`) and the bottom border in `┴`. The tree takes a third of the width, from 24
  columns, and the log keeps at least 30.
- The tree starts with `All output`, then each project in bold and its tasks under `├─` and `└─`,
  with the mark, the time and `hit` or `remote` for a cached task. Under it, **Runs this session**
  lists at most five of the kept runs (the newest first; smith keeps 20, in memory, for the
  session only); a click opens one.
- The log column has the title (the task or `All output`, and `following` or `paused`), the command
  of the task and its state, time and reason (the row of the tree has no room for them), then the log.
- `e` (Errors only, `run.errors`) keeps only the failed tasks; a project without one leaves the tree.
  The footer then says `e all tasks`.
- Under 100 columns the view shows the tree or the log: `Tab` (`run.switch`), a click on the switch
  line (` Tasks   Log: orders:build   tab switches`) or Enter on a task changes it. The header keeps only
  the buttons that answer the run (Stop, Run again, Run unfinished). The frame is
  `‹ Runs / smith run build -p orders`, and `Run · …` on a wide screen.
- Before anything ran, the view says `Nothing ran yet` and lists `b`, `t`, the palette and Projects.
- The new words are in `libs/words/catalog/en/runpage.json`; `runpage` is in the migrated list of
  the words guard. The words of the old run panel that nothing uses left `tui_shell.json`.

Deviations from the brief and the screens:

- The tree groups by project, as the brief asks; the Run screen groups by state (Failed, Running,
  Queued, Done). The counts and the marks say the state.
- The marks stay `×`, `≈` and `▲`; the screens draw `✗`, `⚡` and `⚠`, which `shared/marks` bans. Running
  is the spinner, queued `○`.
- There is no time estimate (`about 20s left`): smith has no history of task times to base it on.
- The `Errors only   Full log   / search` line of the log is not built: Errors only filters the tree, as
  the brief asks, and the log has no search. The cache line under the log (`Cache miss: 2 files
  changed`) is not built: the run output does not give the reason to the shell.
- `run.list`, `RunListSpot`, `RunListRows`, `MinRunListRows` and `MinLogRows` of `splitter` stay, unused,
  because removing an exported symbol would force a major release.
- The VS Code extension and the JetBrains plugin are unchanged: they have their own run output.

## Phase 6 as built

- The palette (`palette` package) groups its results under headings with their counts (`Build  13`).
  With nothing typed the groups are the sections of `commands.PaletteEntries` (Start, Build,
  Projects, Dependencies, Release, Health, Go to, Help). A search puts its first usable match under
  **Best match**, the other matches under their sections in the order they first show, and the
  commands that cannot be used now under **Not usable now**, with their reason on the second line.
  The cursor counts in that drawn order (`groupRows`).
- Each entry has two lines: the title with its key at the right, then the command line in the faint
  colour. A hint `smith <command>: what it does` shows only `smith <command>`; a hint that is no
  command line shows as it is (`commandLine`). The typed text (or each typed word) is lit in the
  accent and bold in the title (`matchSpans`). The selected entry has `▶ ` and the selected
  background.
- The input line has `esc` at its right and a rule under it; the list ends with a rule (it carries
  `N more not usable here` while nothing is typed) and the key hints in the footer style
  (`↑/↓ move • enter run • esc close`). Nothing found says `no command matches` and how to find
  more; a typo still offers the closest commands (`Did you mean`).
- `Palette.At(y, h)` maps a line inside the frame to an entry (both of its lines); the shell uses it
  for the hover and the click. `RowAt` stays, for the list area.
- Health lists the checks by state (`doctorview.GroupByState`, `Model.ByState`, which the shell
  sets): Failed, Warnings, then Passed when shown, each heading in its colour with its count. Each
  check has a second muted line with its project (or `workspace`) and its file (`FileOf`), and `fix`
  at the right when it has a fix. `g` still turns the groups off; the grouping by project stays in
  `doctorview` for the model's own use.
- The fix shows in the detail (`fixwizard.Wizard.Inline`): the wizard keeps its phases, keys and
  `doctorfix` machinery, and draws without a frame, its title on its first line and its buttons from
  the left, with no Cancel (the footer says Esc). The head over it is the state line
  (`doctorview.StateLine`), what is wrong (three lines at most), `or in a terminal: smith doctor
  --fix --only '<check>'` and `doctorview.Steps` (`1 Review ── 2 Apply ── 3 Check again`).
- After the fix ran the detail shows step 3: `Checking again…` while doctor runs, then the new
  state of the check, what was applied, `Not committed: review with git status` and
  **Check again (r)**. Esc or a move in the list leaves it. A check that still fails opens the
  failure of the wizard in the detail, as before in the dialog.
- The inline fix is no modal: the screen is not dimmed, the footer lists the keys of the step, a
  click beside the fix leaves it (not while a step runs). The smith.toml repair and the dependency
  repair keep the dialog. While a fix shows, the Health list takes two fifths of the width, at most
  48 columns, so the detail has the room of the screens.
- `F`, the palette's Fix with doctor and Home's Fix button open the Health view first. On a narrow
  screen the fix opens in the frame `‹ Health / <check>`; Esc closes the fix, then the check.
- The new words are in `libs/words/catalog/en/palette_view.json`, `health_list.json` and
  `healthfix.json`. `palette.footer` left the catalog.

Deviations from the brief and the screens:

- The marks stay `×` and `▲`; the screens draw `✗` and `⚠`, which `shared/marks` bans.
- The palette has no Settings group and no Tasks group: `commands.PaletteEntries` has no such
  sections, and the project's own tasks are in Build, as before. The screens' `tab edit the command
  first` is not built: the palette has no command editor.
- The detail of a check before `f` keeps its action panel (Fix, Check again, Why); the review shows
  once the fix is open, because building the review reads the files.
- The check-again key is `r` (`view.reload`), as before; the screens draw `c`, and one command keeps
  one key. `x review all` is `F` (fix all), as before: `x` is the filter of the failed checks.
- The wizard keeps its own button labels (`Apply (Enter)`, `Show changes (d)`); the screens draw
  `Apply fix f` and `Not now esc`. The note `Undo with smith undo` is not shown: smith commits
  nothing unasked, so `smith undo` cannot revert the fix.
- The VS Code extension and the JetBrains plugin already call the view Health (phase 2); their
  `Doctor` labels name the `smith doctor` command, so nothing changed there.

## Phase 7 as built

- Dependencies (`depsview`) draws the headline `↑ 3 with newer releases · 1 patch · 1 minor · 1 major`
  with the marked count, the chips, then two groups: `Patch and minor releases · 2 · safe to take
  together` (success) and `Major releases · 1 · may break your code, take one at a time` (warning). Each
  group has its column heading (Package, Now, Latest, Level, Used by; the box of the heading marks every
  row) and a rule. A level has its mark: `●` patch (success), `●` minor (info), `▲` major (warning). The
  Used by column shows `web +1` when the names do not fit. The body scrolls by lines around the cursor
  row; the headline and the chips stay.
- The box at the foot of the list (`shell_deps_buttons_mapper.go`) says what an update takes: `1 marked ·
  axios 1.6.8 in 2 projects`, `2 marked: axios, zod`, or with nothing marked `Nothing marked: u updates
  axios 1.6.8`. Its buttons are **Update marked (u)** (**Update axios (u)** with nothing marked),
  **Update all safe (U)** and **Repair dependencies (H)**, which left the top of the list; a button that
  does not fit moves to the next line. Under them the command line of `u`, dim (`$ smith deps up axios`).
  Each button keeps its command: `u` asks with the confirmation that shows `smith deps up … --yes`, `U`
  opens the `smith deps up` form, `H` the repair dialog.
- The detail (`depsview.Model.Detail`, titled by the package) is the versions and the level (with the
  warning of a major release), **Used by** with the folder of each project, **Declared as** (the kinds,
  which left the table), **Registry** (the link and `g`/`Y`, which left the list), **One version** (the
  policy, which shows in every package's detail) and **The same as** `$ smith deps up <package>` (with
  `--major` for a major one, `depsview.PackageArgv`). With no package it is the policy alone; a narrow
  screen with no package shows the policy under the list.
- Release (`releaseview`) draws `✓ 2 packages to release · 1 minor · 1 patch · from main`, the table
  Package, Now, Next, Bump and Why with a rule, the selected row on the Highlight background, and the
  unchanged packages. Under it (`shell_release_page_mapper.go`) the box **Checks before release** with
  the checks of the confirmation (`releaseconfirm.Preflight`: Health, the tests of the session, the
  branch) and `○ Registry sign-in: checked when you release, or before with smith release --rehearse`;
  `Nothing is published until you confirm the plan.`; the buttons `Release now (x)`, `Dry run (p)` and
  `Release notes (n)`; their command lines, dim. While a dry run shows its output (the preview of
  before, unchanged), the checks leave their room to it; during a release the view shows its steps as
  before.
- The detail of Release (`releaseview.Model.Detail`, titled `web 1.0.0 → 1.1.0`) is the bump and the
  plugin, **Why it releases**, **Changes** (the commits, as the drill-down drew them), **Tag**,
  **Public API** when the API check warns, and **The same as** `$ smith release --dry-run --notes`.
- Narrow: Enter (or a click on the selected row, or a double click) opens the package in the frame,
  `‹ Dependencies / zod` or `‹ Release / kit 0.2.0 → 0.2.1`, and Esc goes back
  (`listWithDetailKey`, the commands `deps.open`, `deps.back`, `release.open`, `release.back`). While the
  package is open, mark and the chips are unavailable. Beside the detail Enter does nothing.
- The release plan and the dependency list show a spinner at once while they load (`releaseview.Model.Tick`
  is new).
- The new words are in `libs/words/catalog/en/deps_page.json` and `release_page.json`. The words of the
  old summary, the Kind column, the registry label, the group headings without counts and the tag line
  left `depsview.json` and `releaseview.json`; `release.expand` left `commands.json`.

Deviations from the brief and the screens:

- The marks stay `▲` and `●`; the screens draw `⚠` and `•`, which `shared/marks` bans or does not have.
  Patch and minor share `●` and differ in colour and name.
- The screens draw tick boxes `☐ ☑`; the list keeps `[ ]` and `[x]` of `choicerow.Box`, as every pick list.
- The design's `smith up` is `smith deps up`, the command the actions run. There is no `smith deps up
  --dry-run`: `--dry-run` works only with `--align`, and `--check` only reports what the view already
  lists, so no dry-run button is drawn for Dependencies. `checked 2 min ago` is not drawn: the shell does
  not keep the time of the load.
- The screens' `1-3 level`, `o one version` (a detail line `✓ both use 19.1.0` per package) and `Used by`
  with the manifest file are not built: the chips keep `e`, `p`, `n`, `M`; the policy is the workspace
  verdict of `smith deps up --align --check`, not one per package; smith knows the project's folder, not
  the manifest that declares the package.
- Release has no `Publishes to` or `Commits` column: the plan does not say which registry a package goes
  to. The column is Why (the reason of the plan). The Dry run key stays `p` (the screens draw `d`): one
  command keeps one key. There is no `smith release -p <package>`, so The same as is the notes command.
- `releaseview.Model.Open` stays, no longer set, and `Update` no longer opens the commits: removing an exported
  field would force a major release. The `←→ expand/collapse` keys (`release.expand`) are gone; the
  detail shows the commits.
- The VS Code extension and the JetBrains plugin are unchanged: their Dependencies and Release surfaces
  are their own trees and forms.

## Phase 8 as built

- Settings (`settingsview`) draws its two frames itself. On a wide screen the main frame `Settings` is
  `SectionsPane`: a blank row, each section (`▶ Workspace` on the selected background while the keyboard is
  there, bold while it is in the open section), a blank row between them when they fit, and
  `Per-project settings are in Projects › a project › Settings` at the foot. The frame beside it is
  `FieldsPane`, titled `PaneTitle` (`Workspace · smith.toml`; `Style` and the sections of forms name no file).
  The shell gives the frame of the sections 36 columns (`layout.MinList`) and the rest to the open section;
  the divider between them is the only one of the view (the sections | rows divider and the rows | detail
  divider are gone). Tab, Left and Right move the keyboard between the frames, and the focused frame has the
  accent border (`mainFocused`).
- The open section has its line about it, then its rows. Workspace and the Project section are a table with a
  heading and a rule (Setting, Value, Default, Type; the type and then the default leave first on a narrow
  pane). A switch is `● on` or `○ off`, a value of a fixed set ends in `▾`, an unset switch with no default is
  `○ not decided`. Under the rows the selected row is told in a rounded box titled by its name (the setting's
  path, the plugin, the form) with what it does and, for a setting, the command line it stands for, dim
  (`$ smith config set <path> <value>`). Under the box a setting has `Writes smith.toml exactly as the command
  does.` and the buttons **Save (Enter)** (primary) and **Reset to default (u)** (muted while nothing is written);
  a click on them does what the key does.
- An editor opens in the box, titled by what it changes, with its keys under it and the command line of the
  value as it is typed or picked. A switch no longer flips on `Enter`: it opens the choice On, Off and, for a
  switch with no default, Not decided (an unset), the cursor on what it is now, and nothing is written until
  `Enter` saves the answer (ask, never impose). The plugins of the Plugins section keep their one-key switch with
  the `undo (u)` offer.
- Narrow (`Narrow`, `NarrowClick`): the frame lists the sections, each with a line about it (`18 on · 0 external`,
  `57 settings, 2 set`, `theme auto · symbols auto`, else the section's About). The tab opens there
  (`ToSections`); Enter, a click on the selected section or a double click opens one as `‹ Settings / Workspace`,
  Enter on a field edits it, and Esc closes the editor, then goes back. The footer says `Enter open` on the
  sections (`labels`).
- The strip fits `7 Settings` at 80 columns with three or four badges: the fit counted a gap after the last tab,
  which is never drawn (`tabs.renderFlat` gives the room that gap). The badge spacing is unchanged.
- The help (`helpdialog`) is titled `Keys · <view>` (`Help.Titled`). It lists the keys by scope: `<view> view`, then
  `Everywhere` with the menus as muted headings, each key in the accent and bold in a column of its own and what
  it does beside it (muted with its reason when it cannot run now). Under a rule come the notes (every action is a
  smith command, plain mode) and the mouse. The last line stays while the keys scroll:
  `Find a command by name: Ctrl+K` and `↑/↓ scroll • Esc close`. It is up to 104 columns and 32 rows, and fits
  80 x 30. `commands.HelpRow` and `helpdialog.HelpRow` have the new field `View`.
- The dialogs already were rounded frames with the title in the top border, the primary button in the accent, and
  Esc closing them; a test now checks the confirmation, the help, the palette and the dependency repair. The
  command line a dialog runs is dim (`th.Faint`) in the confirmation, the package prompt, the release adjustment,
  the composer plan and the tools dialog (it was bold, or in the accent). The fix wizard has `Style.Command`: the
  dependency repair shows `$ smith deps repair`, and a doctor fix that stays a dialog (the smith.toml repair, a fix
  outside a workspace) shows `$ smith doctor --fix --only '<check>'` over its buttons. What the dialogs do did not
  change. The toast keeps its rounded box in the border colour.
- An `X all N` label shows no count when no project has the task (`State.allCount`): the palette says `Lint all`, not
  `Lint all 0`, and the footer `build all`, not `build all 0`. A count above zero shows as before.
- The new words are in `libs/words/catalog/en/settings_page.json` and `help_page.json`; `helpdialog` joined the
  migrated list of the words guard. `dialog.help.title`, `settingsview.sections.title`,
  `settingsview.setting.command` and `settingsview.setting.keys.flip` left the catalog.

Deviations from the brief and the screens:

- The sections keep their order and their nine entries (Plugins first); the screens draw five, Workspace first.
  The Workspace filter (`Filter: release▏`) is not built: no section had a filter, and the table scrolls.
- The help is one column: the titles of the keys are sentences, which the screens' three columns would wrap into
  short lines. The keys of the open view come first, as before (the screens put Everywhere first). The screens'
  `Lists` and `Runs` columns are the menus under Everywhere.
- A switch's choice is a column of radios, as every pick list (the screens draw them on one line); a click on one
  picks it.
- The buttons keep the `label (key)` form (`Save (Enter)`); the screens draw `Save enter`.
- `WithSideDetail`, `DetailLines`, `Geometry`, `WithSizes`, `splitter.SettingsListSpot` and `SettingsDetailSpot` stay,
  no longer used by the shell, and `ProjectClick` stays beside the new `ProjectClickAt`: removing an exported
  symbol would force a major release. The project section in the detail of Projects still clicks without a
  column, so its Save and Reset buttons answer the keys, not the mouse.
- The marks stay `×` and `▲` (see phase 2).
- The VS Code extension and the JetBrains plugin are unchanged: their settings are their own forms.

## Follow-up as built

These items of the left-open list were built after phase 8.

- Pinned projects (`rail.Props.Pinned`, `rail.Pin`): under the views, after a blank row, the muted bold heading
  `Pinned`, a blank row when the rail has the room, then one row per pin: the mark of its last run, the name and the
  kind, muted, at the right. A pin that does not fit is not drawn; a pin of a project that left the workspace is kept
  but not drawn. `P` on Projects (`projects.pin`, also in the palette, `Unpin the project from the rail` while it is
  pinned) pins or unpins the selected project; a click on a pin opens Projects on it and clears a filter that hides it.
  The owner decided the pins are per person: they are in `tui.json` in the smith home (`userprefs.Preferences.Pins`,
  by workspace root), never in `smith.toml` or a file git sees.
- Multi-select on Projects: `Space` (`projects.select`) selects the row and moves down, `[x]` and `[ ]` boxes lead
  every row while projects are selected, the count line starts with `N selected`, and a box over the legend lists
  them with the buttons of their tasks and the dim command line (`$ smith run build -p a -p b`). The run keys and
  the buttons run `smith run <task> -p a -p b` (`Command.TaskArgvAcross`), which opens in the Runs view. `Esc`
  (`projects.unselect`, `NeedsPicked`) clears the selection before the filter. The new words are in
  `libs/words/catalog/en/projects_pick.json`.

- The time estimate of the Runs view: `· about 20s left` after the time while a run goes. The owner decided the
  durations live only in the workspace's local cache, which git ignores (`<local cache>/tui-durations.json`,
  `rundurations`), never in the remote cache. Each task keeps its last five passed times; the estimate is the
  median of each task left, less what a running one spent, divided by the tasks running at once. A task with no
  history gives no estimate.
- The log search: `/` (`run.search`) types `Search: text▏  1 of 3` in the log's title, the matches are in the accent and
  bold and the current one is also on the selected background, `n`/`N` (`run.search-next`) go round the matches and
  scroll to them, `Esc` clears the search first.
- The cache miss reason: the run engine now says why a task that uses the cache ran (`runevents.Miss`, the `miss`
  field of `task.finished`: `no-entry`, `inputs-changed` with the count of files and the other kinds, `key-changed`,
  `no-cache`). The Runs view shows `Cache miss: 2 files changed` over the log of the task, and the Tasks table of a
  project `miss · 2 files changed`. The new words are in `libs/words/catalog/en/runs_follow_up.json`.

- The dry run of Dependencies: `smith deps up` (and `smith up`) take `--dry-run` for an update, not only with `--align`.
  It prints each declaration the update would rewrite (`apps/web/package.json  axios  ^1.6.0 → ^1.6.8`), whether it
  would install, and that nothing changed; with no package named it is what `--yes` would take. The action catalogue
  classifies it as reading. The Dependencies box has **Dry run (D)** (`deps.dry-run`, also in the palette) and its
  command line, dim, under the one of the update; it runs at once and opens in the Runs view.
- The per-package one-version line: the detail's `One version` starts with `✓ both use ^19.1.0` (`web uses …`, `all 3
  use …`), or `▲ api ^1.5.0 · web ^1.6.0` in the warning colour when the projects differ, read from the declarations of
  the update (`depsview.Row.Declared`). The workspace's policy lines follow it. The new words are in
  `libs/words/catalog/en/deps_dry_run.json` and `dependencies_dry_run.json`.

Deviations:

- The dry run key is `D`: `d` opens Health everywhere and one command keeps one key. The screens' footer `o one version`
  and `1-3 level` stay as they were (phase 7). The one-version line covers the packages with a newer release only: the
  view lists no other package.
- The estimate counts only passed tasks (a cached, failed or stopped one says nothing about how long it takes), and
  it shares the work by the tasks running now, not by the parallelism smith may reach later.
- The screens' `Errors only   Full log   / search` line is not drawn: `e` already filters the tree, and the search
  shows in the title of the log. `n` and `N` come after `Esc back` in the footer, so a narrow footer drops them first;
  the help lists them.
- The cache line is over the log, with the task's facts, not under it: the bottom of the log is where it grows.
  The screens' `in src/Orders since the last pass` is not said: the engine counts the files, it does not group
  them by folder; `smith why-miss` (`w`) lists them.
- The narrow strip has no Pinned group: it is one line. A narrow screen pins and unpins with `P`, and shows the pins
  once it is wide.
- The selection keeps the `[ ]`/`[x]` boxes of every pick list (the screens draw no mark for it), and it counts only
  on Projects: on another view `b` still builds the selected project, so a key never runs projects the person cannot see.
- A project without the task is passed on with `-p` too: `smith run` skips the projects that lack a task. The buttons
  offer only the tasks one of the selected projects has.
- The pin key is `P`: the screens draw no key for it, and `p` is taken by the project tasks (`package`).

## VS Code as built

The VS Code extension (`apps/vscode`) took the views of the redesign in the editor's own idiom:
tree views in the MonoSmith side bar, the editor's theme colours (`charts.*`) and codicons, and
the editor's dialogs, quick picks and pages. The JetBrains plugin follows in a later pull request.

- The side bar holds the seven views in the order of the rail: Home, Projects, Runs, Health,
  Dependencies, Release, Settings (`package.json` `contributes.views.monosmith`). Health left its
  panel. Release Preview is named Release. Home and Projects show open; the others are collapsed
  and show only when smith is ready.
- Each view is a list view of the shared `list-view` slice, except Projects (`workspace-view`). The
  rows are plain values made by a mapper per view (`home-view`, `runs-view`, `health`,
  `dependencies`, `release-preview`, `settings-view`), and the mappers are tested. A row that does
  something names the smith command it stands for in its description or its tooltip.
- **Home** (`home-view`): the four boxes as rows, Needs you now (`homeQueue`, the rules of
  `dashboard.Queue`) with each item's buttons as rows, and Start something. Doctor and the release
  plan start in the background the first time Home shows. The texts are the terminal's catalogue
  keys (`dashboard.*`), claimed in `libs/words/surface/vscode.json`.
- **Projects**: the mark of the last run is the project's icon; the filter is `filterProject`
  (the words of `projectlist.Matches`); the sections Tasks, Dependencies, Runs and Settings are rows
  under the project, Tasks open.
- **Runs** (`runs-view`): `createRunHistory` keeps 20 runs of the session from the event stream; a
  run started from the editor opens there without the keyboard (`ListView.reveal`). The view shows
  the progress bar of the editor while a run goes. Errors only is a title button with a context key.
- **Health**: grouped by state, the project and file of each check (`fileOf`, the rule of
  `doctorview.FileOf`), the fix in three steps: the review opens as a page beside the code, the
  modal asks to apply, and doctor runs again and says whether the check passes now and that nothing
  was committed.
- **Dependencies**: the safe and major groups of `depsview`, the package detail as rows, and
  marking as the tree's own multi-select. The updates are `smith deps up`, opened in a terminal
  after a yes that shows the command.
- **Release**: the plan, the detail of each package, the checks before release
  (`releaseChecks`: Health and the tests of the session) and the three actions.
- **Settings** (`settings-view`): Plugins, Architecture, Workspace, Recipes, Assistants, CI and
  Forms. A setting row opens the guided `smith config set` on that setting
  (`settingRowAnswers`), so the value is asked and the command is confirmed first.
- **The palette**: `MonoSmith: Commands…` (`Ctrl+Alt+M P`) lists every command in the sections of
  `commands.PaletteEntries`, each with its smith command line under its title. The line of each
  command is in `command-palette/command-lines.config.ts`; a test fails when a command of
  `package.json` has no entry there.

Deviations, and what smith does not give the extension:

- The branch check before a release is left out: `smith serve` gives no git facts (branch,
  upstream, commits ahead), and the extension runs no git of its own.
- The Plugins section cannot show which plugins are on: `plugins` (`smith plugin list --json`) has
  no on/off field. A plugin row asks On or Off, each with its command.
- The Cache column of Tasks, the cache reason of a task and the time estimate of a run are not
  shown yet: the terminal got them after this pull request ([Follow-up as built](#follow-up-as-built)).
- The Run unfinished button of the terminal is not built: a run plan of the editor is tasks across
  projects, not a list of task ids.
- The remote cache item and the tab-completion item of the Home queue are not built: `smith serve`
  has no method for the remote cache, and the extension offers tab completion as its own notice.
- Style (the terminal's theme and symbols) has no section: VS Code uses its own theme.
- The fix review opens as a page and the apply is a modal question. A tree view has no inline
  detail, so the review cannot sit inside the view as in the terminal.
- The keys of the terminal (`1`-`7`, `f`, `x`, `u`) are not bound. VS Code gives each view a
  Focus command, which the palette lists under Go to.

## Left open

What the redesign did not build, across the phases, for the owner to decide:

- The marks `✗`, `⚠` and `⚡` the screens draw: `shared/marks` bans them (phases 2 to 8).
- The Settings and Tasks groups of the palette and its command
  editor (phase 6), and the Workspace filter of Settings (phase 8).
- The keys the screens change (`c` check again, `d` dry run): one command keeps one key.
- Unused exported symbols kept to avoid a major release: `run.list`, `RunListSpot`, `RunListRows`, `MinRunListRows`,
  `MinLogRows`, `releaseview.Model.Open`, and the side-detail functions of `settingsview` and `splitter`.
- The VS Code extension follows the new views (see [VS Code as built](#vs-code-as-built)). The JetBrains plugin
  follows none of the new layouts; only the Health name (phase 2) reached it.

## Screens

The screens below are plain-text renders of the design.

### Main

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │                                                                                                                           │
│   Projects         2 │ │  ╭───────────────────────────╮ ╭───────────────────────────╮ ╭───────────────────────────╮ ╭───────────────────────────╮  │
│                      │ │  │ Changed                   │ │ Last run                  │ │ Health                    │ │ Release                   │  │
│   Runs         ✗1  3 │ │  │ ● 7 projects              │ │ ✗ 1 failed                │ │ ⚠ 2 warnings              │ │ ✓ 2 ready                 │  │
│                      │ │  │ since main                │ │ 17 passed · 9 cached      │ │ 0 failed checks           │ │ 1 minor · 1 patch         │  │
│   Health       ⚠2  4 │ │  ╰───────────────────────────╯ ╰───────────────────────────╯ ╰───────────────────────────╯ ╰───────────────────────────╯  │
│                      │ │                                                                                                                           │
│   Dependencies ↑3  5 │ │  Needs you now  5 items                                                                                                   │
│                      │ │                                                                                                                           │
│   Release      ✓2  6 │ │  │ ✗ Failed   test failed in orders-api                                                                  See why enter    │
│                      │ │  │            2 of 140 tests · 4 min ago                                                                                  │
│   Settings         7 │ │                                                                                                                           │
│                      │ │    ● Changed  7 projects changed and not tested                                                         Test changed t    │
│                      │ │               web, kit, orders-api and 4 more                                                                             │
│                      │ │                                                                                                                           │
│ Pinned               │ │    ⚠ Warning  2 health checks warn                                                                            Review 4    │
│                      │ │               both have a fix that needs no decision                                                                      │
│ ✗ orders-api     app │ │                                                                                                                           │
│ ✓ web            app │ │    ↑ Updates  3 dependencies have newer versions                                                              Review 5    │
│ ● kit            lib │ │               1 major · 1 minor · 1 patch                                                                                 │
│                      │ │                                                                                                                           │
│                      │ │    ✓ Release  2 packages are ready to release                                                                Preview 6    │
│                      │ │               kit 1.4.0 · web 2.0.1                                                                                       │
│                      │ │                                                                                                                           │
│                      │ │  ───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────  │
│                      │ │                                                                                                                           │
│                      │ │  Start something                                                                                                          │
│                      │ │                                                                                                                           │
│                      │ │   Build changed b     Lint all l     Run a task… r     Add a project a     Settings 7                                     │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### Projects

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Projects 7/42 ──────────────────────────╮╭─ orders-api ──────────────────────────────────────────────────────────────────╮
│                      │ │                                          ││                                                                               │
│   Home             1 │ │ Filter: changed kind:app▏                ││  apps/orders-api · app · C# (.NET 9)  ● changed         Run… r    Open o      │
│                      │ │ 4 of 12 apps · 3 of 30 libs              ││                                                                               │
│ ▶ Projects         2 │ │                                          ││   Tasks    Dependencies    Runs    Settings                                   │
│                      │ │ APPS                                     ││  ───────────────────────────────────────────────────────────────────────────  │
│   Runs         ✗1  3 │ │ ▶ ✗  orders-api         C#       ● changed│                                                                               │
│                      │ │   ✓  web                React    ● changed│  Task          Last result       Time       Cache                     Key     │
│   Health       ⚠2  4 │ │   ✓  ms-teams-bot       Go       ● changed│  ───────────────────────────────────────────────────────────────────────────  │
│                      │ │   –  billing-fn         Python   ● changed│  build         ✓ passed          4.1s       hit                       b       │
│   Dependencies ↑3  5 │ │ LIBS                                     ││  test          ✗ 2 failed        21.6s      miss · src changed        t       │
│                      │ │   ✓  kit                TS       ● changed│  lint          ✓ passed          2.3s       hit                       l       │
│   Release      ✓2  6 │ │   ✓  core               C#       ● changed│  package       – not run         –          –                         p       │
│                      │ │   ✓  words              Go       ● changed│  start         own terminal      –          never cached              s       │
│   Settings         7 │ │                                          ││                                                                               │
│                      │ │                                          ││  start runs in your terminal: smith leaves the screen and comes back when it en
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││  ╭─ Uses ────────────────────────────╮ ╭─ Used by ─────────────────────────╮  │
│ Pinned               │ │                                          ││  │                                   │ │                                   │  │
│                      │ │                                          ││  │ core · kit · words                │ │ web · billing-fn                  │  │
│ ✗ orders-api     app │ │                                          ││  │                                   │ │ a change here tests them too      │  │
│ ✓ web            app │ │                                          ││  ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│ ● kit            lib │ │                                          ││                                                                               │
│                      │ │                                          ││  Recent runs                                                                  │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││  ✗ test   2 failed   21.6s               4 min ago                            │
│                      │ │                                          ││  ✓ build  passed     4.1s (cache)        4 min ago                            │
│                      │ │                                          ││  ✓ test   passed     19.8s               yesterday                            │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │                                          ││                                                                               │
│                      │ │ ✓ passed  ✗ failed  – not run            ││                                                                               │
╰──────────────────────╯ ╰──────────────────────────────────────────╯╰───────────────────────────────────────────────────────────────────────────────╯

 / filter • enter run test • space select • tab next pane • ctrl+k commands • ? more
 $ smith run test -p orders-api                                                                                                 main  ·  2.64.1  ·  ⚠2
```

### Run

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Run · smith test --changed ──────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│   Home             1 │ │  ⣾ Running 18 tasks  00:24 · about 20s left                                             Copy command y     Stop ctrl+c    │
│                      │ │                                                                                                                           │
│   Projects         2 │ │  ████████████████████████████████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░  67%       │
│                      │ │  ✗ 1 failed    ✓ 9 passed    ⚡ 2 cached    ◐ 2 running    ○ 4 queued                                                      │
│ ▶ Runs         ✗1  3 │ │                                                                                                                           │
│                      │ │───────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────│
│   Health       ⚠2  4 │ │                                       │                                                                                   │
│                      │ │ Failed                                │ orders-api:test    Errors only   Full log   / search                              │
│   Dependencies ↑3  5 │ │ ├─ ✗ orders-api:test           21.6s  │                                                                                   │
│                      │ │ Running                               │   Restored 3 projects (0.4s)                                                    ┃ │
│   Release      ✓2  6 │ │ ├─ ⣾ web:test                    14s  │   Running 140 tests in Orders.Api.Tests                                         ┃ │
│                      │ │ ├─ ⣾ billing-fn:test              6s  │                                                                                 ┃ │
│   Settings         7 │ │ Queued · waits on web:test            │ ✗ OrderTotals_IncludeTax          tests/OrderTotalsTests.cs:42                  ┃ │
│                      │ │ ├─ ○ web-e2e:test                     │     Expected: 119.00                                                            ┃ │
│                      │ │ ├─ ○ docs:test                        │     Actual:   100.00                                                            ┃ │
│                      │ │ Done                                  │                                                                                 ┃ │
│ Pinned               │ │ ├─ ✓ kit:test                   3.2s  │ ✗ Refund_AfterShipping_IsPartial  tests/RefundTests.cs:88                       │ │
│                      │ │ ├─ ⚡ core:test                 cache  │     Expected status Partial, got Full                                           │ │
│ ✗ orders-api     app │ │ ├─ ✓ ms-teams-bot:test          1.9s  │                                                                                 │ │
│ ✓ web            app │ │ ├─ ✓ … 7 more                         │ ✓ 138 passed   ✗ 2 failed   21.6s                                               │ │
│ ● kit            lib │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │                                                                                 │ │
│                      │ │                                       │ ───────────────────────────────────────────────────────────────────────────────── │
│                      │ │                                       │ Cache miss: 2 files changed in src/Orders since the last pass.                    │
│                      │ │                                       │  Rerun failed R     Open file o                                                   │
╰──────────────────────╯ ╰───────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ task • e errors only • R rerun failed • y copy command • Y copy log • , . older/newer • esc back
 ◐ running  $ smith test --changed  00:24                                                                                       main  ·  2.64.1  ·  ⚠2
```

### Palette

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │           ╭─ Commands ───────────────────────────────────────────────────────────────╮                                    │
│   Projects         2 │ │  ╭────────│                                                                          │───╮ ╭───────────────────────────╮  │
│                      │ │  │ Changed│  › test orders▏                                                   esc    │   │ │ Release                   │  │
│   Runs         ✗1  3 │ │  │ ● 7 pro│──────────────────────────────────────────────────────────────────────────│   │ │ ✓ 2 ready                 │  │
│                      │ │  │ since m│  Best match                                                              │   │ │ 1 minor · 1 patch         │  │
│   Health       ⚠2  4 │ │  ╰────────│  ▶ Test orders-api                                                enter  │───╯ ╰───────────────────────────╯  │
│                      │ │           │    smith run test -p orders-api                                          │                                    │
│   Dependencies ↑3  5 │ │  Needs you│  ▶ Test orders-api and what uses it                                      │                                    │
│                      │ │           │    smith run test -p orders-api --dependents                             │                                    │
│   Release      ✓2  6 │ │  │ ✗ Faile│  ↻ Rerun the 2 failed tests                                           R  │                   See why enter    │
│                      │ │  │        │    from the last run · 4 min ago                                         │                                    │
│   Settings         7 │ │           │  Projects                                                                │                                    │
│                      │ │    ● Chang│  ✗ orders-api                                                            │                  Test changed t    │
│                      │ │           │    app · C# · apps/orders-api                                            │                                    │
│                      │ │           │  Settings                                                                │                                    │
│ Pinned               │ │    ⚠ Warni│  ⚙ orders-api: test timeout                                              │                        Review 4    │
│                      │ │           │    smith config set -p orders-api tasks.test.timeout                     │                                    │
│ ✗ orders-api     app │ │           │                                                                          │                                    │
│ ✓ web            app │ │    ↑ Updat│                                                                          │                        Review 5    │
│ ● kit            lib │ │           │──────────────────────────────────────────────────────────────────────────│                                    │
│                      │ │           │  ↑/↓ move • enter run • tab edit the command first                       │                                    │
│                      │ │    ✓ Relea╰──────────────────────────────────────────────────────────────────────────╯                       Preview 6    │
│                      │ │               kit 1.4.0 · web 2.0.1                                                                                       │
│                      │ │                                                                                                                           │
│                      │ │  ───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────  │
│                      │ │                                                                                                                           │
│                      │ │  Start something                                                                                                          │
│                      │ │                                                                                                                           │
│                      │ │   Build changed b     Lint all l     Run a task… r     Add a project a     Settings 7                                     │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### Health

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Health ───────────────────────────────────╮╭─ Lock file out of date ─────────────────────────────────────────────────────╮
│                      │ │                                            ││                                                                             │
│   Home             1 │ │ ⚠ 2 warnings  · 0 failed · 46 passed       ││  ⚠ Warning                                                                  │
│                      │ │                                            ││                                                                             │
│   Projects         2 │ │ ▶ ⚠ Lock file out of date             fix  ││  package.json in web asks for react 19.2, but package-lock.json still holds 1
│                      │ │     web · package-lock.json                ││  A clean install in CI will fail.                                           │
│   Runs         ✗1  3 │ │                                            ││                                                                             │
│                      │ │   ⚠ TypeScript config drifts          fix  ││   1 Review  ── 2 Apply ── 3 Check again                                     │
│ ▶ Health       ⚠2  4 │ │     applies to 22 projects                 ││                                                                             │
│                      │ │                                            ││  ╭─ apps/web/package-lock.json  +2 −2 ───────────────────────────────────╮  │
│   Dependencies ↑3  5 │ │ ▸ 46 passed checks            p            ││  │                                                                       │  │
│                      │ │                                            ││  │     "node_modules/react": {                                           │  │
│   Release      ✓2  6 │ │                                            ││  │ -     "version": "19.1.0",                                            │  │
│                      │ │                                            ││  │ +     "version": "19.2.0",                                            │  │
│   Settings         7 │ │                                            ││  │ -     "resolved": "…/react-19.1.0.tgz",                               │  │
│                      │ │                                            ││  │ +     "resolved": "…/react-19.2.0.tgz",                               │  │
│                      │ │                                            ││  │       … 1 more line                                                   │  │
│                      │ │                                            ││  │                                                                       │  │
│ Pinned               │ │                                            ││  │                                                                       │  │
│                      │ │                                            ││  ╰───────────────────────────────────────────────────────────────────────╯  │
│ ✗ orders-api     app │ │                                            ││                                                                             │
│ ✓ web            app │ │                                            ││  The same as                                                                │
│ ● kit            lib │ │                                            ││  $ smith doctor --fix lockfile-sync -p web                                  │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │                                            ││                                                                             │
│                      │ │ ╭────────────────────────────────────────╮ ││                                                                             │
│                      │ │ │ Both warnings have a fix that          │ ││                                                                             │
│                      │ │ │ needs no decision.                     │ ││                                                                             │
│                      │ │ │                                        │ ││                                                                             │
│                      │ │ │  Review all fixes x                    │ ││  Nothing is committed. Undo with smith undo.                                │
│                      │ │ ╰────────────────────────────────────────╯ ││                                                                             │
│                      │ │                                            ││   Apply fix f     Not now esc                                               │
╰──────────────────────╯ ╰────────────────────────────────────────────╯╰─────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ check • f apply fix • x review all • c check again • p show passed • ctrl+k commands
 $ smith doctor  ⚠ 2 warnings  46 passed                                                                                        main  ·  2.64.1  ·  ⚠2
```

### Deps

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Dependencies ─────────────────────────────────────────────────────────────────────╮╭─ react ─────────────────────────────╮
│                      │ │                                                                                    ││                                     │
│   Home             1 │ │  ↑ 3 with newer releases  · 1 patch · 1 minor · 1 major · checked 2 min ago        ││  19.1.0 → 19.2.0  ▲ minor           │
│                      │ │                                                                                    ││                                     │
│   Projects         2 │ │   All 3   Patch 1   Minor 1   Major 1                                              ││  Used by                            │
│                      │ │                                                                                    ││  web       apps/web/package.json    │
│   Runs         ✗1  3 │ │  Patch and minor releases: safe to take together                                   ││  kit       libs/kit/package.json    │
│                      │ │                                                                                    ││                                     │
│   Health       ⚠2  4 │ │     Package             Now       Latest    Level     Projects                     ││  One version                        │
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  ✓ both use 19.1.0                  │
│ ▶ Dependencies ↑3  5 │ │  ☐  react               19.1.0    19.2.0    ▲ minor   web, kit                     ││                                     │
│                      │ │  ☑  Newtonsoft.Json     13.0.3    13.0.4    • patch   orders-api, core             ││  Registry                           │
│   Release      ✓2  6 │ │                                                                                    ││  npmjs.com/package/react            │
│                      │ │  Major releases: may break your code, take one at a time                           ││  g opens it · Y copies the link     │
│   Settings         7 │ │                                                                                    ││                                     │
│                      │ │     Package             Now       Latest    Level     Projects                     ││  The same as                        │
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  $ smith up react                   │
│                      │ │  ☐  cobra               1.8.1     2.0.0     ⚠ major   ms-teams-bot                 ││                                     │
│ Pinned               │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│ ✗ orders-api     app │ │                                                                                    ││                                     │
│ ✓ web            app │ │                                                                                    ││                                     │
│ ● kit            lib │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │  ╭──────────────────────────────────────────────────────────────────────────────╮  ││                                     │
│                      │ │  │ 1 marked · Newtonsoft.Json 13.0.4 in 2 projects                              │  ││                                     │
│                      │ │  │                                                                              │  ││                                     │
│                      │ │  │  Update marked u     Update all safe U                                       │  ││                                     │
│                      │ │  ╰──────────────────────────────────────────────────────────────────────────────╯  ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
╰──────────────────────╯ ╰────────────────────────────────────────────────────────────────────────────────────╯╰─────────────────────────────────────╯

 ↑/↓ move • space mark • u update marked • 1-3 level • g release notes • o one version • ctrl+k commands
 $ smith up --dry-run  ↑ 3 newer                                                                                                main  ·  2.64.1  ·  ⚠2
```

### Release

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Release ──────────────────────────────────────────────────────────────────────────╮╭─ kit 1.3.2 → 1.4.0 ─────────────────╮
│                      │ │                                                                                    ││                                     │
│   Home             1 │ │  ✓ 2 packages to release  · 1 minor · 1 patch · from main                          ││  Changes since kit-v1.3.2           │
│                      │ │                                                                                    ││                                     │
│   Projects         2 │ │  Package           Now       Next      Bump       Commits   Publishes to           ││  feat  add a retry option to the clie
│                      │ │  ────────────────────────────────────────────────────────────────────────────────  ││  feat  export the time helpers      │
│   Runs         ✗1  3 │ │  kit               1.3.2     1.4.0     ▲ minor    4         npm                    ││  fix   keep the order of headers    │
│                      │ │  web               2.0.0     2.0.1     • patch    1         container registry     ││  docs  describe the retry option    │
│   Health       ⚠2  4 │ │                                                                                    ││                                     │
│                      │ │  unchanged: orders-api, core, words and 35 more                                    ││  feat → minor bump                  │
│   Dependencies ↑3  5 │ │                                                                                    ││                                     │
│                      │ │  ╭─ Checks before release ──────────────────────────────────────────────────────╮  ││  Tag                                │
│ ▶ Release      ✓2  6 │ │  │                                                                              │  ││  kit-v1.4.0                         │
│                      │ │  │ ✓ Working tree is clean                                                      │  ││                                     │
│   Settings         7 │ │  │ ✓ Tags give the right versions                                               │  ││  The same as                        │
│                      │ │  │ ✓ CI passed on main                                                          │  ││  $ smith release -p kit             │
│                      │ │  │ ○ Registry sign-in: checked when you release                                 │  ││                                     │
│                      │ │  │                                                                              │  ││                                     │
│ Pinned               │ │  ╰──────────────────────────────────────────────────────────────────────────────╯  ││                                     │
│                      │ │                                                                                    ││                                     │
│ ✗ orders-api     app │ │                                                                                    ││                                     │
│ ✓ web            app │ │                                                                                    ││                                     │
│ ● kit            lib │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │  Nothing is published until you confirm the plan.                                  ││                                     │
│                      │ │                                                                                    ││                                     │
│                      │ │   Release 2 packages x     Dry run d                                               ││                                     │
│                      │ │                                                                                    ││                                     │
╰──────────────────────╯ ╰────────────────────────────────────────────────────────────────────────────────────╯╰─────────────────────────────────────╯

 ↑/↓ package • enter changes • x release • d dry run • ctrl+k commands
 $ smith release --dry-run  ✓ 2 to release                                                                                      main  ·  2.64.1  ·  ⚠2
```

### Settings

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Settings ─────────────────────╮╭─ Workspace · smith.toml ────────────────────────────────────────────────────────────────╮
│                      │ │                                ││                                                                                         │
│   Home             1 │ │ ▶ Workspace                    ││  Filter: release▏                                                                       │
│                      │ │                                ││                                                                                         │
│   Projects         2 │ │   Plugins                      ││  Setting                             Value                   Default         Type       │
│                      │ │                                ││  ─────────────────────────────────────────────────────────────────────────────────────  │
│   Runs         ✗1  3 │ │   Style                        ││  release.nuget_trusted_publishing    ● on                    not decided     switch     │
│                      │ │                                ││  release.channel                     stable ▾                stable          choice     │
│   Health       ⚠2  4 │ │   Recipes                      ││  release.tag_prefix                  {project}-v             {project}-v     text       │
│                      │ │                                ││  release.rubygems_trusted_publishing ○ off                   not decided     switch     │
│   Dependencies ↑3  5 │ │   Forms                        ││                                                                                         │
│                      │ │                                ││  ╭─ release.nuget_trusted_publishing ────────────────────────────────────────────────╮  │
│   Release      ✓2  6 │ │                                ││  │                                                                                   │  │
│                      │ │                                ││  │ Publish NuGet packages with a short-lived key from GitHub OIDC,                   │  │
│ ▶ Settings         7 │ │ Per-project settings are in    ││  │ so the repository keeps no NuGet API key.                                         │  │
│                      │ │ Projects › a project › Settings││  │                                                                                   │  │
│                      │ │                                ││  │ ( ) Not decided    (•) On    ( ) Off                                              │  │
│                      │ │                                ││  │                                                                                   │  │
│ Pinned               │ │                                ││  │ $ smith config set release.nuget_trusted_publishing true                          │  │
│                      │ │                                ││  ╰───────────────────────────────────────────────────────────────────────────────────╯  │
│ ✗ orders-api     app │ │                                ││                                                                                         │
│ ✓ web            app │ │                                ││                                                                                         │
│ ● kit            lib │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││  Writes smith.toml exactly as the command does.                                         │
│                      │ │                                ││                                                                                         │
│                      │ │                                ││   Save enter     Reset to default u                                                     │
╰──────────────────────╯ ╰────────────────────────────────╯╰─────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • tab next pane • enter change • u reset • ctrl+k commands
 $ smith config list                                                                                                            main  ·  2.64.1  ·  ⚠2
```

### Help

```text
  smith   monosmith    main ↑2                 ⌕ Type a command, project or task   ctrl+k                                               smith 2.64.1

╭─ Views ──────────────╮ ╭─ Home ────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                      │ │                                                                                                                           │
│ ▶ Home             1 │ │  monosmith  42 projects · 7 changed since main · last run 3 min ago                                                       │
│                      │ │ ╭─ Keys · Home ────────────────────────────────────────────────────────────────────────────────╮                          │
│   Projects         2 │ │ │                                                                                              │───────────────────────╮  │
│                      │ │ │  Everywhere                     Lists                          Runs                          │ease                   │  │
│   Runs         ✗1  3 │ │ │                                                                                              │ ready                 │  │
│                      │ │ │  1-7      open a view           ↑/↓ j/k  move                  ctrl+c   stop                 │inor · 1 patch         │  │
│   Health       ⚠2  4 │ │ │  ctrl+k   commands              enter    open or run           r        run again            │───────────────────────╯  │
│                      │ │ │  tab      next pane             /        filter                R        rerun failed         │                          │
│   Dependencies ↑3  5 │ │ │  ?        these keys            space    select several        e        errors only          │                          │
│                      │ │ │  m        mouse on / off        y        copy the command      Y        copy the log         │                          │
│   Release      ✓2  6 │ │ │  q        quit                  esc      back                  , .      older / newer run    │         See why enter    │
│                      │ │ │                                                                                              │                          │
│   Settings         7 │ │ │                                                                                              │                          │
│                      │ │ │──────────────────────────────────────────────────────────────────────────────────────────────│        Test changed t    │
│                      │ │ │                                                                                              │                          │
│                      │ │ │  Every action is a smith command. The status bar shows the last one, and y copies it.        │                          │
│ Pinned               │ │ │                                                                                              │              Review 4    │
│                      │ │ │  Mouse: click a row, a tab or a key hint. Drag a divider to resize. Wheel scrolls.           │                          │
│ ✗ orders-api     app │ │ │  Plain mode (NO_COLOR or SMITH_THEME=plain) keeps every symbol, so no state is told by colour alone.                    │
│ ✓ web            app │ │ │                                                                                              │              Review 5    │
│ ● kit            lib │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │             Preview 6    │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │────────────────────────  │
│                      │ │ │                                                                                              │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ │  Find a command by name:  ctrl+k                                                   esc close │                          │
│                      │ │ │                                                                                              │                          │
│                      │ │ ╰──────────────────────────────────────────────────────────────────────────────────────────────╯                          │
│                      │ │                                                                                                                           │
│                      │ │  Tip: every action here is a smith command. Press y on any row to copy it.                                                │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
│                      │ │                                                                                                                           │
╰──────────────────────╯ ╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯

 ↑/↓ move • enter see why • t test changed • 1-7 views • ctrl+k commands • ? more • q quit
 $ smith test --changed  ✗ 1 failed  ✓ 17 passed  38s                                                                           main  ·  2.64.1  ·  ⚠2
```

### NarrowHome

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Home ───────────────────────────────────────────────────────────────────────╮
│ 42 projects · 7 changed · last run 3 min ago                                 │
│                                                                              │
│ ╭───────────────────────────────────╮ ╭───────────────────────────────────╮  │
│ │ Changed                           │ │ Last run                          │  │
│ │ ● 7 projects                      │ │ ✗ 1 failed                        │  │
│ ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│ ╭───────────────────────────────────╮ ╭───────────────────────────────────╮  │
│ │ Health                            │ │ Release                           │  │
│ │ ⚠ 2 warnings                      │ │ ✓ 2 ready                         │  │
│ ╰───────────────────────────────────╯ ╰───────────────────────────────────╯  │
│                                                                              │
│ Needs you now 1/5                                                            │
│                                                                              │
│ │ ✗ test failed in orders-api                                         enter  │
│ │   2 of 140 tests · 4 min ago                                               │
│   ● 7 projects changed, not tested                                        t  │
│     web, kit, orders-api +4                                                  │
│   ⚠ 2 health checks warn                                                  4  │
│     both have a fix                                                          │
│   ↑ 3 dependency updates                                                  5  │
│     1 major                                                                  │
│   ✓ 2 packages ready to release                                           6  │
│     kit 1.4.0 · web 2.0.1                                                    │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter see why • t test changed • ? more
 $ smith test --changed  ✗ 1  ✓ 17  38s                                 2.64.1
```

### NarrowProjects

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Projects 7/42 ──────────────────────────────────────────────────────────────╮
│ Filter: changed▏                                        ✓ pass ✗ fail – none │
│                                                                              │
│ APPS                                                                         │
│ ▶ ✗  orders-api           app   C#        ● changed               test ✗     │
│   ✓  web                  app   React     ● changed                          │
│   ✓  ms-teams-bot         app   Go        ● changed                          │
│   –  billing-fn           app   Python    ● changed                          │
│ LIBS                                                                         │
│   ✓  kit                  lib   TS        ● changed                          │
│   ✓  core                 lib   C#        ● changed                          │
│   ✓  words                lib   Go        ● changed                          │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│ enter opens the project · space selects several to run together              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 / filter • enter open • space select • t test
 $ smith run test -p orders-api                                         2.64.1
```

### NarrowProject

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Projects / orders-api ────────────────────────────────────────────────────╮
│ app · C# · apps/orders-api  ● changed                                Run… r  │
│                                                                              │
│  Tasks   Deps   Runs   Settings                                              │
│ ──────────────────────────────────────────────────────────────────────────── │
│ Task        Result          Time     Cache               Key                 │
│ build       ✓ passed        4.1s     hit                 b                   │
│ test        ✗ 2 failed      21.6s    miss · changed      t                   │
│ lint        ✓ passed        2.3s     hit                 l                   │
│ package     – not run       –        –                   p                   │
│ start       own terminal    –        never               s                   │
│                                                                              │
│ Uses      core · kit · words                                                 │
│ Used by   web · billing-fn  (a change here tests them too)                   │
│                                                                              │
│ Recent runs                                                                  │
│                                                                              │
│ ✗ test    2 failed   21.6s   4 min ago                                       │
│ ✓ build   passed     4.1s    4 min ago                                       │
│ ✓ test    passed     19.8s   yesterday                                       │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 esc back • tab section • enter run • r run…
 $ smith run test -p orders-api                                         2.64.1
```

### NarrowRunTasks

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ smith test --changed ───────────────────────────────────────────────────────╮
│ ⣾ 18 tasks · 00:24  about 20s left                              Stop ctrl+c  │
│                                                                              │
│ ██████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░  67%  │
│ ✗ 1   ✓ 9   ⚡ 2   ◐ 2   ○ 4                                                  │
│                                                                              │
│  Tasks   Log: orders-api:test                                   tab switches │
│──────────────────────────────────────────────────────────────────────────────│
│ Failed                                                                       │
│ ├─ ✗ orders-api:test                                                  21.6s  │
│ Running                                                                      │
│ ├─ ⣾ web:test                                                           14s  │
│ ├─ ⣾ billing-fn:test                                                     6s  │
│ Queued · waits on web:test                                                   │
│ ├─ ○ web-e2e:test                                                            │
│ ├─ ○ docs:test                                                               │
│ Done                                                                         │
│ ├─ ✓ kit:test                                                          3.2s  │
│ ├─ ⚡ core:test                                                        cache  │
│ ├─ ✓ … 8 more                                                                │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ task • enter log • tab log • R rerun failed • esc back
 ◐ running  00:24                                                       2.64.1
```

### NarrowRunLog

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ smith test --changed ───────────────────────────────────────────────────────╮
│ ⣾ 18 tasks · 00:24  about 20s left                              Stop ctrl+c  │
│                                                                              │
│ ██████████████████████████████████████████████░░░░░░░░░░░░░░░░░░░░░░░░  67%  │
│ ✗ 1   ✓ 9   ⚡ 2   ◐ 2   ○ 4                                                  │
│                                                                              │
│  Tasks   Log: orders-api:test                                   tab switches │
│──────────────────────────────────────────────────────────────────────────────│
│  Errors only   Full log                                 ✗ 2 · lines 1-11/11  │
│                                                                              │
│ ✗ OrderTotals_IncludeTax                                                    ┃│
│   tests/OrderTotalsTests.cs:42                                              ┃│
│     Expected: 119.00                                                        ┃│
│     Actual:   100.00                                                        ┃│
│                                                                             ┃│
│ ✗ Refund_AfterShipping_IsPartial                                            ┃│
│   tests/RefundTests.cs:88                                                   ││
│     Expected status Partial, got Full                                       ││
│                                                                             ││
│ ✓ 138 passed   ✗ 2 failed   21.6s                                           ││
│                                                                             ││
│                                                                             ││
│                                                                             ││
│ Cache miss: 2 files changed in src/Orders.       Rerun failed R              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ scroll • e errors only • / search • Y copy log • tab tasks
 ◐ running  00:24                                                       2.64.1
```

### NarrowPalette

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Home ───────────────────────────────────────────────────────────────────────╮
│ 42 projects · 7 changed · last run 3 min ago                                 │
│   ╭─ Commands ───────────────────────────────────────────────────────────╮   │
│ ╭─│                                                                      │╮  │
│ │ │ › test orders▏                                                  esc  ││  │
│ │ │──────────────────────────────────────────────────────────────────────││  │
│ ╰─│ ▶ Test orders-api                                             enter  │╯  │
│ ╭─│   smith run test -p orders-api                                       │╮  │
│ │ │ ▶ Test orders-api and dependents                                     ││  │
│ │ │   smith run test -p orders-api --dependents                          ││  │
│ ╰─│ ↻ Rerun the 2 failed tests                                           │╯  │
│   │   from the last run · 4 min ago                                      │   │
│ Ne│ ✗ Open orders-api                                                    │   │
│   │   project · app · C#                                                 │   │
│ │ │ ⚙ orders-api: test timeout                                           │r  │
│ │ │   smith config set -p orders-api …                                   │   │
│   │                                                                      │t  │
│   │                                                                      │   │
│   │──────────────────────────────────────────────────────────────────────│4  │
│   │ ↑/↓ move • enter run • tab edit first                                │   │
│   ╰──────────────────────────────────────────────────────────────────────╯5  │
│     1 major                                                                  │
│   ✓ 2 packages ready to release                                           6  │
│     kit 1.4.0 · web 2.0.1                                                    │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter see why • t test changed • ? more
 $ smith test --changed  ✗ 1  ✓ 17  38s                                 2.64.1
```

### NarrowHealth

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Health ─────────────────────────────────────────────────────────────────────╮
│ ⚠ 2 warnings  · 0 failed · 46 passed · 1 min ago                             │
│                                                                              │
│ ▶ ⚠ Lock file out of date                                             fix    │
│     web · package-lock.json                                                  │
│                                                                              │
│   ⚠ TypeScript config drifts from the standard                        fix    │
│     applies to 22 projects                                                   │
│                                                                              │
│ ▸ 46 passed checks                                                           │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│ ╭──────────────────────────────────────────────────────────────────────────╮ │
│ │ Both warnings have a fix that needs no decision.                         │ │
│ │                                                                          │ │
│ │  Review all fixes x                                                      │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ check • enter open • x review all • p passed
 $ smith doctor  ⚠ 2  46 passed                                         2.64.1
```

### NarrowHealthFix

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Health / Lock file out of date ───────────────────────────────────────────╮
│ ⚠ Warning · web                                                              │
│                                                                              │
│ package.json asks for react 19.2, but package-lock.json                      │
│ still holds 19.1. A clean install in CI will fail.                           │
│                                                                              │
│  1 Review  ── 2 Apply ── 3 Check again                                       │
│                                                                              │
│ ╭─ package-lock.json  +2 −2 ───────────────────────────────────────────────╮ │
│ │                                                                          │ │
│ │ -   "version": "19.1.0",                                                 │ │
│ │ +   "version": "19.2.0",                                                 │ │
│ │ -   "resolved": "…/react-19.1.0.tgz",                                    │ │
│ │ +   "resolved": "…/react-19.2.0.tgz",                                    │ │
│ │     … 1 more line                                                        │ │
│ │                                                                          │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│ Nothing is committed. Undo with smith undo.                                  │
│                                                                              │
│  Apply fix f     Not now esc                                                 │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 esc back • f apply • ↑/↓ scroll • c check again
 $ smith doctor --fix lockfile-sync -p web                              2.64.1
```

### NarrowDeps

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Dependencies ───────────────────────────────────────────────────────────────╮
│ ↑ 3 newer · 1 patch · 1 minor · 1 major                                      │
│                                                                              │
│  All 3   Patch 1   Minor 1   Major 1                                         │
│                                                                              │
│ Safe together                                                                │
│    Package            Now      Latest   Level     Used by                    │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ☐  react              19.1.0   19.2.0   ▲ minor   web +1                     │
│ ☑  Newtonsoft.Json    13.0.3   13.0.4   • patch   orders-api +1              │
│                                                                              │
│ Take one at a time                                                           │
│    Package            Now      Latest   Level     Used by                    │
│ ──────────────────────────────────────────────────────────────────────────── │
│ ☐  cobra              1.8.1    2.0.0    ⚠ major   ms-teams-bot               │
│                                                                              │
│ 1 marked: Newtonsoft.Json 13.0.4                                             │
│                                                                              │
│  Update marked u     Update all safe U                                       │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • space mark • u update • enter details
 $ smith up --dry-run                                                   2.64.1
```

### NarrowRelease

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Release ────────────────────────────────────────────────────────────────────╮
│ ✓ 2 packages to release · 1 minor · 1 patch                                  │
│                                                                              │
│ Package       Now      Next     Bump      To                                 │
│ ──────────────────────────────────────────────────────────────────────────── │
│ kit           1.3.2    1.4.0    ▲ minor   npm                                │
│ web           2.0.0    2.0.1    • patch   containers                         │
│                                                                              │
│ unchanged: orders-api, core, words +35                                       │
│                                                                              │
│ Checks                                                                       │
│                                                                              │
│ ✓ Working tree is clean                                                      │
│ ✓ Tags give the right versions                                               │
│ ✓ CI passed on main                                                          │
│                                                                              │
│                                                                              │
│ Nothing is published until you confirm the plan.                             │
│                                                                              │
│  Release 2 x     Dry run d                                                   │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ package • enter changes • x release • d dry run
 $ smith release --dry-run                                              2.64.1
```

### NarrowSettings

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ Settings ───────────────────────────────────────────────────────────────────╮
│                                                                              │
│ ▶ Workspace                                                                  │
│   every smith.toml setting, 42                                               │
│                                                                              │
│   Plugins                                                                    │
│   6 on · 2 external, trusted                                                 │
│                                                                              │
│   Style                                                                      │
│   theme auto · symbols unicode                                               │
│                                                                              │
│   Recipes                                                                    │
│   1 recipe, valid                                                            │
│                                                                              │
│   Forms                                                                      │
│   add or remove a project…                                                   │
│                                                                              │
│                                                                              │
│ Per-project settings: Projects › a project › Settings                        │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ move • enter open • esc back
 $ smith config list                                                    2.64.1
```

### NarrowSetting

```text
  smith   monosmith   main ↑2                               ctrl+k commands
1 Home  2 Projects  3 Runs ✗1  4 Health ⚠2  5 Deps ↑3  6 Release ✓2  7 Settings
╭─ ‹ Settings / Workspace ─────────────────────────────────────────────────────╮
│ Filter: release▏                                                             │
│                                                                              │
│ release.nuget_trusted_publishing                              ● on           │
│ release.channel                                               stable ▾       │
│ release.tag_prefix                                            {project}-v    │
│ release.rubygems_trusted_publishing                           ○ off          │
│                                                                              │
│ ╭─ release.nuget_trusted_publishing ───────────────────────────────────────╮ │
│ │                                                                          │ │
│ │ Publish NuGet packages with a short-lived key from                       │ │
│ │ GitHub OIDC, so the repository keeps no NuGet API key.                   │ │
│ │                                                                          │ │
│ │ ( ) Not decided   (•) On   ( ) Off                                       │ │
│ │                                                                          │ │
│ │ Default: not decided · switch                                            │ │
│ │                                                                          │ │
│ ╰──────────────────────────────────────────────────────────────────────────╯ │
│                                                                              │
│  Save enter     Reset u                                                      │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
╰──────────────────────────────────────────────────────────────────────────────╯
 ↑/↓ setting • enter change • u reset • esc back
 $ smith config set release.nuget_trusted_publishing true               2.64.1
```

