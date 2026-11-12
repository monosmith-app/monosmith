# The smith terminal interface

Typing `smith` with no arguments in a terminal opens an interactive shell: a full-screen
view of the workspace with tabs for its health, projects, runs, dependencies, release and
settings, single-key commands, a command palette, and wizards for the jobs that ask questions. It is
a front end for the same commands this manual documents. It never does what a command line
could not: every action runs the equivalent `smith` command, shows it in the status bar,
and can be repeated in a script or in CI.

This page describes the interface as of smith 1.62.0. It was written from the source of the
shell (`apps/smith/tui`, in particular its keymap in `commands/keymap_config.go`) and from
running the commands the shell stands for. The full-screen views were not captured live for
this page, so no screenshots are included and the exact layout of a screen may differ in
detail from the descriptions here.

## When it starts

A bare `smith` opens the shell when standard input is a terminal and the `CI` environment
variable is empty. Otherwise, with no arguments, smith prints its help like any other
command.

| Where you start it | What you see |
|---|---|
| Inside a workspace (a folder with `smith.toml`, or below one) | The **Home** view, with the seven views as tabs. |
| In an empty folder | The **Start** menu: create a workspace (`smith new`): the name, then the same few questions the command asks (CI, main language, code layout, more options). |
| In a folder that holds files but is not a workspace | The Start menu with **Bring this repository into MonoSmith** first, so Enter starts the wizard. |
| In a folder whose `smith.toml` uses a renamed key (for example `[[task_groups]]`, now `[[groups]]`) | The Start menu with **Fix smith.toml** first. Enter shows the change as a diff and writes nothing until you confirm; then the shell opens the workspace. Any other problem in the file keeps its message, with its line and column. |

When the shell cannot be drawn (standard output is not a terminal, `SMITH_SCREEN_READER` is
set, or `SMITH_THEME=plain`), smith asks the same questions as plain numbered lines and runs
the command you end up with; see [Plain mode](#plain-mode-and-environment).

## The screen

From top to bottom:

1. **Header**: the workspace's name (`smith` when there is none, with a hint about
   `smith new` and `smith adopt` outside a workspace), the git branch, and a health summary:
   `checking…` until the first doctor run has finished (it starts with the shell), then
   the counts of failed checks and warnings, or a check mark when everything passed. When the
   workspace has a remote cache, the last mark is its state: `≈ checking…` until the first
   answer, then `≈✓` (it answers and accepts the credentials), `≈▲` (it does not answer) or
   `≈×` (it refuses the credentials, or `SMITH_CACHE_KEY` or the store's credentials are not
   set). ASCII terminals show `[c][v]`, `[c][!]` and `[c][x]`. smith asks the remote in the
   background with a 5 second timeout, so the screen never waits. There is no mark without a
   remote. The header sits at the right end of its row.
2. **Tabs**: Home, Projects, Doctor, Deps, Release, Settings, numbered 1 to 6. When the names do
   not fit, they are shortened, then the strip scrolls with `‹` and `›`. On a short terminal
   the tabs are one row of names.
3. **The view.** Projects and Doctor show a list with a **detail pane** beside it; on a
   narrow terminal the detail replaces the list when you press Enter, and Esc goes back.
   The main screen has no box. One rule runs under the tabs across the whole width, a blank
   cell keeps the text one cell in from each side, and thin lines divide the panels: the
   vertical line between a list and its detail, and the horizontal line between stacked
   panels (the rows and the help of a setting, the task list and the log). The lines meet
   in tees (`┬` under the rule, `├` and `┤` where a horizontal line meets a vertical one).
   There is no left, right or bottom edge. The rule has no title: the lit tab says where
   you are, and a fact shows once. A form on the Start tab names itself on its first line.
4. **Footer**: key hints for the open view, then the ones every view ends with
   (`views`, `palette`, `help`, `quit`, `mouse`). Each hint can be clicked.
5. **Status bar**: the exact command line running or last run, its result, the git branch
   and the smith version.

A short **toast** in a corner reports a result, a refusal (with the reason) or a warning, and
goes away by itself. A result toast belongs to the view it was raised on.

## Views

| View | Key | Shows |
|---|---|---|
| Home | `1` | An attention queue: what needs a look, the most urgent first, each with the buttons that deal with it (see [Home](#home)). The first line says what the workspace holds (`24 projects (23 TypeScript / JavaScript, 1 Go)`; its name and branch are in the header). Under the queue: the last three runs and two buttons, "Add a project" (`a`) and "Compose a stack" (`A`); click one or press its key. |
| Projects | `2` | The top panel is the workspace: one line with the count (`24 projects · 0 affected by the changes (●)`) and the legend of the arrows (`↑ = dependencies · ↓ = dependants`), then one line of buttons for the whole workspace: `+ New project (a)` (opens the Add wizard), `Verify (V)`, `Build (B)`, `Test (T)`, `Lint (L)`, `Typecheck (C)` and `Build release` (`smith run build --release`), each on every project, and `Clear cache` (`smith cache clear`: with a remote cache set it asks first which cache goes, the local one, the remote one (`--remote`) or both (`--all`), then shows the same question as the command line, with the store, the workspace and the count; without a remote it deletes the local cache after you confirm); a button that does not fit moves to the next line. Then every project, ordered by kind and then name, with its language. The detail pane has two parts split by a thin rule. The top part never scrolls: the project's name, kind, language and stack (beside the list it leaves out the ones the list shows as columns; opened full-screen it shows them), root, version, whether it is affected, and its dependencies and dependants as names in a line that wraps. The bottom part scrolls (`Ctrl+U` `Ctrl+D`, the wheel, or the keyboard cursor after `Tab`): the actions that apply to the project, under `Run on <project>` (its tasks, the packages, How was this resolved?, Why, Graph), then **More**: the forms that apply to it (swap a choice, environments of a web app, Tauri, Docker Hub), `Change kind…`, `Eject a tool's configuration…` and `Remove project…`, one button per line. Under them is the description of the chosen action and the one button that runs it (see [Action panels](#action-panels)). Then come its derived files (when the workspace derives files). The version reads `current → next` while the plan of the Release tab releases the project, and the current version alone otherwise; the list has the same Version column, which shortens to `→ next` when it is narrow. At the foot of the pane, under a rule, **Recent runs** lists the last three runs of this session and says how many more are kept; `Tab` or a click opens it upward to list every kept run. The list has the columns Project, Kind, Language, Stack, Version and `↑ ↓` (the number of dependencies and of dependants); a narrow pane drops Stack, Language and Kind before Version. While a task runs, or when you open a run, the detail pane shows that run instead (see [The run panel](#the-run-panel)). |
| Doctor | `3` | The checks of [`smith doctor`](doctor.md), grouped by project or flat, each with a detail pane in four parts (see [The Doctor detail](#the-doctor-detail)); fixes open the fix wizard. |
| Deps | `4` | A row of chips at the top, `All`, `Patch`, `Minor` and `Major` with the number of releases of each (keys `e`, `p`, `n`, `M`; click one too), limits the list to one level. Under it the summary, the warning about major releases among the chosen packages (so it shows without scrolling), and the dependencies with newer releases, grouped patch and minor, then major (columns Package, Now, Latest, Level, Kind, Projects), with the `smith deps up` command line for a selection. The one-version policy's verdict (`One version`) is in the detail beside the list; on a narrow screen it is under the list. A line under the table gives the registry page of the selected package ([Registry link](#the-registry-link-in-the-deps-view)). |
| Release | `5` | What the next release would do (the plan of [`smith release --dry-run`](release.md)), the commits behind each bump, and a button `Release now (x)` that releases it. On a release channel branch it says so first: ``Release channel `next`: prerelease versions…``, and the confirmation of the release repeats it. When nothing releases, the view says so in one line (`Nothing to release: no feat: or fix: commits since the last tag`) and keeps the reason behind a fold, `Why? (w)`. For a package that was never released it also offers `Create the first release commit (c)`. |
| Settings | `6` | Plugins, style and the settings of `smith.toml`, changed in place; what the selected row is and does is in the detail beside the list (under the list on a narrow screen); see [Settings](#settings). |

### Home

Home is a queue of what needs attention, not a table of facts. Only the items that apply
appear, in this order:

1. **The last run failed** (a stopped run is not a failure): `orders:build failed`, the
   task that failed (the command when no task failed). Buttons: `See why` (opens that task's
   log in the run panel of Projects) and `Run it again` (not offered for a command that makes something,
   such as `add`).
2. **Setup is broken**: `2 checks fail, 6 can be fixed now`, with the first two failing
   checks named under it. Buttons: `Fix` (only when a fix is ready) and `Open Doctor`.
   When doctor could not run, the item says so and offers `Retry`.
3. **Checks warn**: `1 warning: golangci-lint is 2.14.0, CI uses 2.15.0`. Buttons: `Fix`
   (only when a fix is ready, the same button and wording as for a failure) and `Open Doctor`. Home never says `All good` while a check warns.
4. **Changes not tested**: `N projects are affected by your changes`. Button: `Test what changed`.
5. **A release is ready**: `Ready to release: 8 packages, smith 1.132.0 → 1.133.0 (292 commits)`,
   naming the package with the most commits. Button: `Open Release`.
6. **Updates**: `N dependencies have newer releases` (`Open Deps`), or `Updates not looked at yet` (`Open Deps` too: its key, `4`, opens Deps and reads the registries the first time).
7. **Tab completion** (only while it waits): `Tab completion is not set up`. Buttons: `Set up` and
   `Not now`. `Set up` (also `Set up tab completion` in the palette) runs `smith completion install --dry-run`
   and shows the files and lines in the confirmation; nothing is written until you press `Run`, and the
   confirmed line runs in the run panel. `Not now` is remembered for you in `tui.json` (`completionOffer`),
   and the item never comes back; so does a finished setup. The item is muted, so it never takes Enter or
   the `All good` line away. See [`smith completion`](completion.md).
8. **Remote cache** (only when the workspace has one): its state and how many local entries it
   does not hold, for example `Remote cache: connected · 12 local entries not uploaded`, with
   `Push` (runs [`smith cache push`](cache-push.md) after a confirmation) and `Cache status`.
   smith counts in the background: it asks the remote about at most 500 entries, 8 at a time,
   for at most 10 seconds, and says `at least 500` when it stops early. When the remote holds
   everything the item says so and is muted. A remote that does not answer or refuses the
   credentials is a warning, with the reason under it; for a refusal the first button is
   `Set up again`. The count is made again after each run.
9. When nothing needs attention, one calm line, `All good` (with the last run that passed under it), and `Build all`.

A workspace without projects shows one item with `Add a project` and `Compose a stack`.
While the first doctor run goes on, one item says `Running the health checks…` and nothing
can be pressed.

Every item is drawn the same way: its title in bold, an optional note under it, and its buttons on one line
below, left-aligned under the title, never at the right. A button names its destination (`Open Doctor`,
`Open Deps`) and shows its key. **Enter runs the first button of the first item**, which shows `(Enter)`;
the footer says `next (Enter)`. Every button has three looks:
at rest, under the pointer and pressed. A click on a button runs its command; a click
elsewhere on an item does nothing.

Under the queue, `Recent runs` lists the last three runs of the session on one line (mark,
command, duration); a click on one opens it in the run panel of Projects. The line under it names the
workspace, its branch and what it holds, with the two buttons that add a project.

The release item needs the plan of the Release view. The shell reads that plan in the
background when it starts, with the checks, and the Release view shows the same read. Home never waits for
it: the item appears when the plan arrives, and nothing shows if reading it failed. After a command
that changes the workspace, the plan is read again in the background.

Doctor and Release load the first time you open them; Deps and Settings load their data on first open too.
`r` reads a view again.

## Keys

Every key is a single key; none needs a modifier except the Ctrl combinations noted. No key
works while a text field is editing. The same table is in the help overlay (`?` or `F1`),
which is built from the same keymap as the footer and the palette, so it cannot list a key
that does nothing. Esc never quits.

### Everywhere

| Key | Does |
|---|---|
| `1` to `6` | Show Home, Projects, Doctor, Deps, Release, Settings. |
| `Up` `Down`, `j` `k` | Move in a list. |
| `Left` `Right` | Move between buttons on Home. In Projects and Doctor, `Right` moves the keyboard to the detail and `Left` back to the list. In Release, `Right` shows the commits behind a bump and `Left` hides them. |
| `Tab` | Move the keyboard to the next panel of a view that has two (Projects: the list, then Recent runs, then the detail; Doctor: the list, then the detail). The footer names where `Tab` goes. |
| `Enter` | Open or run the selection. |
| `Esc` | Back, close, clear. |
| `?`, `F1` | The shortcuts overlay (Up, Down, PgUp, PgDn, Home, End scroll it; Esc, `?`, `F1`, `q` or Enter close it). |
| `:`, `Ctrl+P`, `/` | The command palette. |
| `q` | Quit (not while a run is going: `Ctrl+C` stops it first). |
| `Ctrl+C` | Stop the run in progress; pressed again, quit. A stopped run is **cancelled**, not failed (see [A stopped run](#a-stopped-run)). |
| `m` | Turn the mouse on or off. Off lets the terminal select text itself; with it on, a plain drag selects too (see Select text). |
| `y` | Copy the focused item (remedy, command, project path or package) to the clipboard, through the terminal (OSC 52), so it also works over SSH and in tmux panes that allow it. Works in Projects, Run, Doctor, Deps. |
| `r` | Reload this view: read the files and the checks again, run nothing (Home, Doctor, Release, Deps). The footer says `r reload view`. While the run panel is open `r` is `run again`: it runs the last command line again. |

### Run tasks (in a workspace)

| Key | Does | Where |
|---|---|---|
| `b` | `smith run build -p <project>` on the project selected in the list; the footer names it (`b build orders`) | footer in every workspace view |
| `B` | `smith run build` on every project; the footer counts them (`B build all 23`) | footer in every workspace view |
| `t` | `smith run test -p <project>`, same; `T` tests every project | footer in every workspace view |
| `l` | `smith run lint -p <project>`; `L` checks the style of every project | `l` in the footer of Projects, `L` in the palette |
| `c` | `smith run typecheck -p <project>`; `C` checks the types of every project | `c` in the footer of Projects, `C` in the palette |
| `w` | `smith run lint test --affected` | footer in Home, Projects |
| `Ctrl+B` | `smith run build --release`: the optimised build (Rust, .NET, Android), on the selected project, else on every project (`release build orders`) | last in the footer of every workspace view; the button `Build release` of Projects and the palette build every project |

One key has one meaning on every view, the run panel and Home included: the lower-case keys
`b`, `t`, `l` and `c` act on the project selected in the list, and the capital keys `B`, `T`,
`L` and `C` act on every project. Every label, in the footer, the palette and the project detail,
names the scope (`build orders`, `build all 23`). With no project selected, a
lower-case key covers the workspace.

A run opens in the right panel of Projects, from any view: the keys work in every workspace view and
switch to Projects. The footer shows only
the ones that make sense for the open view (the column above). The top panel of Projects
has Verify, Build, Test, Lint and Typecheck of every project as buttons, and the palette names the scope in
each title (`Build orders`, `Build all 23`). `Verify (V)` runs [`smith verify`](verify.md), the run CI
makes: the dependency boundaries, then lint, typecheck, test and build, and a check that no file git
tracks changed. The run panel lists its tasks as it does for `smith run`.

### The run panel

There is no Run tab. While a task runs, or when you open a past run, the right panel of Projects
shows that run, as Nx and Turborepo do, and the list of projects stays on the left. From the top
the panel has:

- the header, `<project> · <task>`, then a spinner (or the mark of how it ended) and `n/m`;
- the buttons `Back (Esc)`, `Stop (Ctrl+C)` (faint once nothing runs), `Run again (r)`,
  `Run unfinished (R)` (only when a task did not finish) and `Copy log (Y)`;
- the line `run 2 of 5 · ‹ older  newer ›` (the keys `,` and `.` do the same);
- the command line and the progress bar with the counts, one row for each task with its mark
  (`✓` passed, `×` failed, `≈` from cache, `»` skipped, `⊘` stopped), and the log of the
  task chosen (or all output) under a rule that says if it follows or is paused. A task the
  remote cache served says `cached … from remote cache` in its row, and the closing line counts
  them: `31 from cache, 29 of them remote`.

A failed run opens with its first failed task chosen, so the log says why. Starting a run
from any view (a key, a Home button, the palette) switches to Projects with the panel open.
`Esc` shows all output, and `Esc` again, or `Back`, shows the project again; the run goes on,
and Recent runs under the list opens it again. On a narrow terminal the panel replaces the list
and `Esc` goes back to it. Home has `See why` and `Run it again`; they open the panel too.

### A stopped run

`Ctrl+C` asks the run to stop. The status line says "Stopping" until the process ends, and
then it stops saying it. A run that was stopped is neutral, not red: the marks use `⊘`
(cancelled) for the task that was running and `»` (skipped) for those that never started, the
bar reads `3/8 done, 5 skipped` and never 100%, the task that was running is not counted as
failed, and the toast says "Cancelled". `R` runs the tasks that did not finish again.

### Per view

| View | Key | Does |
|---|---|---|
| Home | `Enter` | Run the first button of the first item of the queue (`Fix`, `Test what changed`, `Build all`, …). |
| any workspace view | `d` | Open Doctor (shown in the footer of Home and Projects). |
| Home, Projects, Doctor, Deps, Release | `r` | Reload this view: read the files and the checks again. It runs nothing. The footer says `reload view`; while the run panel is open the same key is `run again`. |
| Projects | `a` | Add project (a guided form). |
| any workspace view | `+` | Install a package: a form that asks which project (the one selected, `Left` `Right` change it), which packages and whether they are development dependencies, shows the `smith install -p <project> ...` line, then runs it in the run panel of Projects. The palette's and the detail's "Install a package…" open the same form. |
| any workspace view | `-` | Remove a package: the same form for `smith uninstall -p <project> ...`. After the form, a confirmation shows the exact command line and what it changes, because a removal edits the manifest and the lock file. |
| any workspace view | `i` | Install everything every project declares, as after a fresh clone (asks to confirm the command). To add one package use `+`. |
| Projects | `z` | Zoom the detail to the full width, and back. |
| Projects | `e`, `o`, `g` | The buttons of the detail: How was this resolved? (`smith explain` of the project), Why (who owns its files) and Graph (the dependencies and the dependants). A project with derived files also has `E` explain them, `u` update them, `K` keep my edit, `O` own the files I edited and `r` replace them. `+` and `-` install and remove a package. A task with no key of its own takes a free letter shown on its button. The buttons of the well-known tasks (`build`, `test`, `lint`, `typecheck`, `package`) name what the task does to the kind of the project: `Build app`, `Test library`, `Package app`, `Package library`, and `Pack for publishing` for a project of kind `package`, because `package` is also a kind. The hint under the button keeps the raw task name (`smith run package`). Any other task (`stage-web`, `package-dmg`) reads `Run <task>`. The keys do not change. |
| Projects | (buttons only) | Under **More**, the detail opens the forms that apply to the selected project, each one on that project. They have no key, and the palette keeps its entries. A project with a stack has `Change my test tool or linter (Swap a choice)…`. A Node app has `Add environments to a web app…` and `Remove environments from a web app…`. A Tauri app (one with a `tauri-build` task) has `Make the icons of a Tauri app…`. A Node app that is not a Tauri app yet has `Make it a Tauri desktop app…` (`smith tauri add`). An app with an `image` task has `Set the Docker Hub descriptions…`. Last come `Change kind…` (`smith kind`), `Eject a tool's configuration…` (`smith eject`) and `Remove project…`. A button that does not apply to the project does not show. Each form shows the plan or the command line and asks before it changes anything. |
| Projects, Run, Doctor, Settings | `\` `|`, `[` `]` | `\` selects the next divider between two panels and `|` the previous one. `Shift`+arrows select one too. `[` moves it left or up and `]` right or down. `Tab` picks the next divider and `Esc` lets go. With one divider, `[` and `]` need no selection. Drag a divider with the mouse to do the same. A divider you drag goes back to its resting look when
you let go; a divider you select with the keys stays lit until `Esc`. Each view keeps its sizes in the preferences file. |
| Projects, Doctor, Deps, Settings | `Ctrl+U` `Ctrl+D` | Scroll the detail (in Projects its bottom part: the top part stays). |
| Run | `r` | Run the last command again, as it was. A command that changes files asks first. A command that makes or removes something (`add`, `new`, `remove`, `adopt`, `convert`) is not run again: the key says why. |
| Run | `f` | Follow the log, or pause it to read. |
| Run | `w` | On a task that ran (not one replayed from the cache), open the **Why would a task run again?** form with the task filled in: `Enter` runs `smith why-miss <project:task>`. On a cached task, or one that did not run, the key says why it does nothing. Outside the run panel `w` still runs the affected projects. |
| Run | `Up` `Down` `PgUp` `PgDn` | Pick a task, scroll its log. |
| Run panel | `Esc` | Show all output; press it again to go back to the project (the run goes on in the background). The `Back (Esc)` button does the same. |
| Run | `,` `.` | Show the older or the newer run of this session (smith keeps the last 20). The line under the buttons, `run 2 of 5 · ‹ older  newer ›`, has the same two actions as clickable parts: `‹ older` and `newer ›` step through the runs. A part that cannot be used now (no older run, a run in progress) is faint, and a click on it says why. |
| Projects, Recent runs | `Tab`, `Up` `Down`, `Enter` | **Recent runs** sits on the bottom rows of the pane under a rule. Closed, it shows the three newest runs (command, how it ended, how long it took, how long ago) and `N more runs` when more are kept. `Tab` or a click opens it upward to list every kept run (the last 20) with the cursor on a row; `Up` and `Down` choose a row, `Enter` or a double click opens that run in the right panel, and a single click only chooses it. `Esc` closes the list and gives the keyboard back to the projects; `Tab` closes it and moves on to the detail. |
| Projects, Doctor | `Tab`, `Right`, `Left`, `Up` `Down`, `Enter` | `Tab` (or `Right`, except on a Doctor group header, where it unfolds the group) moves the keyboard to the detail beside the list. There `Up` and `Down` choose one of its buttons, `Enter` presses it and `PgUp` `PgDn` scroll it. `Tab`, `Left` or `Esc` gives the keyboard back to the list. With a run open, the keyboard is on the run until `Tab` gives it to the list; the run stays open. A click in a panel gives it the keyboard. |
| Home | `Up` `Down` `Left` `Right`, `Enter`, `PgUp` `PgDn`, `Ctrl+U` `Ctrl+D` | The arrows move through the buttons and runs in reading order (`Up` and `Left` go back, `Down` and `Right` go on, `Home` and `End` jump to the first and the last); `Enter` presses the one the keyboard is on, and that one shows `(Enter)`. A page taller than the terminal scrolls: `PgUp` `PgDn`, `Ctrl+U` `Ctrl+D` and the wheel move it, the first or last row says `↑ more (PgUp, Ctrl+U)` or `↓ more (PgDn, Ctrl+D)`, and the arrows keep the target in view. |
| Doctor | `Enter` | Expand the check (narrow terminals), or unfold the detail (wide). |
| Doctor | `f` | Fix this check: the wizard asks, shows what changes, then applies. |
| Doctor | `F` | Fix with doctor: the wizard on all the fixes that need no decision. |
| Doctor | `g` | Group the checks by project, or list them flat. |
| Doctor | `x` | Show or hide the failed checks in every group. The headline has a button for it, `× 2 failed (x)`. |
| Doctor | `w` | Show or hide the warnings in every group. The headline button reads `▲ 1 warning (w)` (`warnings` for any other count). |
| Doctor | `p` | Show or hide the passed checks (the headline button `✓ 20 passed (p)`). By default they are one collapsed line, so this filter starts off. |
| Doctor | `r` | Check again, for when you fixed things outside smith. The headline has a button, `Check again (r)`, beside the filter buttons. The headline stays at the top of the panel while the checks below it scroll. The list stays while doctor runs, with its groups, filters and cursor; the button gives way to `Checking again…`, and a toast says what doctor found (`Checked again: 3 failed, 5 warnings`). After a fix ends, whether it worked or not, doctor checks again by itself in the same way. |
| Doctor | `e` | Show every line of the problem, or fewer. |
| Doctor | `PgUp` `PgDn` | Scroll the detail. |
| Deps | `Space` | Mark or unmark a dependency. |
| Deps | `e`, `p`, `n`, `M` | Show every release, only the patch, only the minor or only the major ones (the chips at the top). Marks stay when the filter changes; the box in the heading marks the rows shown, or unmarks them when all are marked. |
| Deps | `u`, `Enter` | Update the marked ones (`smith deps up`). |
| Deps | `g` | Open the registry page of the package under the cursor in your browser, and copy the link. It does nothing when no page is known. |
| Deps | `Y` | Copy the registry link of the package under the cursor. |
| Deps | `i` | Install a package (`smith install -p <project> <package>`). It asks the names (a name or name@version, several with spaces), then the projects (space ticks; the projects that declare the package under the cursor start ticked, and nothing runs before you confirm), then the kind: production or development (`-D`). Each project is listed with its language and folder, so for a project with several parts, such as a Tauri app (a Rust part and a web part), you pick the part that gets the package. The confirmation opens at once and shows the plan: the latest release of each name and the section it goes into. The first choice of the projects step is `Workspace root: shared by every project, in the root package.json` (`-p .`), shown only in a workspace with a Node project; it can be ticked together with projects, starts ticked when the cursor is on a `(root)` row, and the kind step applies to it. The plan and the confirmation list it with the exact `npm install` command. Machine-wide installs (`npm -g`), peer and optional dependencies are not offered: `smith install` has no way to do them. |
| Deps | `x` | Remove the package under the cursor (`smith uninstall -p <project> <package>`). A package that one project declares goes straight to the confirmation with that project as `-p`; the workspace root, shown as `(root)`, is `-p .`. When several projects declare it, a multiple choice asks from which ones (space toggles); smith never picks for you. A development dependency is removed like any other. |
| Deps | `v` | Bring drifting versions to one (`smith deps up --align`). |
| Deps | `U` | Update all dependencies: opens the `smith deps up` form, the one the Start tab lists, which asks the options and ends at the confirmation. |
| Deps | `S` | Sync the versions: opens the `smith sync` form in the same way. |
| Deps | `H` | Repair the dependencies (`smith deps repair`), also the **Repair dependencies** button above the list and the palette (type heal, broken, node_modules, reinstall or lockfile). The dialog opens at once with a spinner while each plugin says what it can repair. Then it shows the four repairs, lightest first, each with what it deletes: Prune, Clean reinstall, Rebuild the lockfile, Clear the download cache. None is marked as suggested. A repair that no tool of the workspace has is shown and disabled, with the reason. The next screen shows the plan tool by tool (what is deleted, what runs, the install) and the warnings (a cache that other workspaces share, versions that can change). `Enter` repairs, and the progress names each step. `Esc` goes back one screen, and on the first screen it closes the dialog. A failed step shows what the tool printed. Nothing is committed. |
| Release | `Enter`, `Right`, `Left` | Show the commits behind the selected bump (`Enter` toggles, `Right` shows, `Left` hides). Each row starts with `▸` while its commits are folded and `▾` while they show (`>` and `v` with ASCII symbols); a click on the row does the same as `Enter`. The footer says `←→ expand/collapse`. |
| Release | `p` | Preview the next release (`smith release --dry-run`). |
| Release | `n` | Print the release notes (`smith release --dry-run --notes`). |
| Release | `x` | Release now… (`smith release`: first hold back or bump packages, then a confirmation that says it cannot be undone). The button `Release now (x)` does the same. |
| Release | `w` | Open or close `Why?`, the explanation of an empty plan: which commits make a release. Only when nothing releases; otherwise `w` runs the affected lint and test, as in the other views. |
| Release | `c` | Create the first release commit: `git commit --allow-empty -m "feat: first version"`, after a confirmation. Only for a package that was never released; otherwise `c` runs the typecheck. |
| Settings | `Up` `Down` | Move in the list of sections, or in its rows. |
| Settings | `Tab`, `Left` `Right` | Switch between the list of sections and its rows. |
| Settings | `Space`, `Enter` | Change the selected plugin or setting. |
| Settings | `t` | Review the selected external plugin and trust it. |
| Settings | `f` | Forget that an external plugin is trusted. |
| Settings | `u` | Reset the selected setting to its default. |
| Settings | `Esc` | Back to the list of sections, or cancel the editor. |
| Start menu | `n` | New workspace… (outside a workspace). |
| Start menu | `a` | Bring into MonoSmith… (opens the wizard). |

A key is refused with a reason in the status line when the view cannot do it now (nothing to
update, no safe fix), and its hint leaves the footer. The palette and the help list it greyed,
with the reason (`Disabled: a run is in progress`). In a folder that is not a workspace the help
lists only the keys that can run. Until the registries answer, the Deps footer keeps only
`reload`: the keys that need the list wait for it.

### The registry link in the Deps view

Under the table, the Deps view shows the registry page of the selected package: `Registry:` and
the address, then the keys `g` (open it in your browser, and copy the link) and `Y` (copy the link).
The address is printed in full, so it can be read and copied; in a terminal that draws hyperlinks
(Windows Terminal, iTerm2, WezTerm, kitty, VS Code, GNOME Terminal and other VTE 0.50 or newer, Konsole,
foot, Alacritty, Ghostty) it is also a clickable OSC 8 link. Inside tmux or screen, and in a terminal
smith does not know, only the plain address is printed; `FORCE_HYPERLINK=1` turns links on and `=0`
turns them off. smith opens a browser only when you press `g`, never by itself.

The page is on the registry of the package's language: npmjs.com, pypi.org, nuget.org, crates.io,
Maven Central, pub.dev, rubygems.org, packagist.org or pkg.go.dev. A project that installs from a
private registry named in `smith.toml` (`[registries.<name>]`) uses that registry: for Azure Artifacts the
page of the feed (`https://dev.azure.com/<org>/<project>/_artifacts/feed/<feed>/Npm/<package>/overview`)
is worked out from the feed address. For any other private registry, and when an `.npmrc`,
`PIP_INDEX_URL` or `NuGet.Config` points the tools at one, smith cannot work out the web address, so the
line says "no page known" and shows no link rather than a wrong one. The same address is the `url` field
of each update in [`smith up --json`](up.md).

### The footer follows the mode

- While a text box is open, the footer shows only `Enter confirm` and `Esc back`.
- When the run shown has a failed task, `r run again` and `Y copy log` lead the footer.
- The `Esc` label says what Esc does: in the run panel, `all output` while one task's log is
  shown, then `back`; it leaves the footer while a run shows all output.
- A tab shows a spinner after its name while its work goes on in the background: the checks
  of Doctor, the registries asked for Deps, the plan being read for Release. The Deps view
  shows a spinner and the seconds while it waits.

### In the palette

`:`, `Ctrl+P` or `/` opens a palette: type to find a command; Up, Down, PgUp and PgDn choose,
Enter picks, Esc or `Ctrl+C` closes. There is one line per intent, in plain words, grouped
under Build, Projects, Dependencies, Release and Health (then Go to and Help), each with its
key at the right and the command it runs as a muted hint. What cannot work now is hidden
until you type; searching finds it with the reason. Nothing is typed as a raw command line:
a command that needs an answer opens its own form, such as Install a package (project
preselected, then the packages) or Add project. Commands that have no key are reached here:
Regenerate CI pipelines (asks to confirm), Info, Show what changed (`smith affected`),
Project graph, Manage plugins, Upgrade workspace (asks to confirm) and Remove project. The
Health section also holds the 2.0 commands, described in
[Undo, report, why and derived files](#undo-report-why-and-derived-files): Undo the last
change, Open the report, Write the agent file, Check the derived files, Update the derived
files and Why is this here?
The palette and every box drawn over the screen stay inside the content area; behind a box the
view is emptied so no text shows at its edges, and a toast sits inside the content area.

While a box is open (the palette, help, a confirmation, a picker, the fix, adopt and compose
screens, the tools and result dialogs) it blocks the screen. The whole screen behind it, from the
header to the status bar, is drawn dimmed in the faint style, with its colours removed, so
nothing behind it looks clickable. The mouse acts only on the box: a hover, a click, a press,
the wheel or a drag over the tabs, the header, the footer, the panels or the dividers does
nothing, and the pointer position behind the box is not tracked. A click outside a box does what
the box does with it: the palette closes, the other boxes ignore it. The keyboard works as before.
Without colour (`NO_COLOR`), the dimming is the faint attribute of the terminal, when the
terminal prints styles at all.

### In a confirmation

Commands that change the workspace or publish (`smith ci generate`, `smith install`,
`smith upgrade`, `smith release`) first show their command line in a dialog. `y` confirms;
`n`, `Esc`, `Ctrl+C` or `q` cancels; `Enter` or `Space` presses the focused button; `Tab`,
`Left`, `Right` and `Shift+Tab` move between the buttons. A plan too long for the dialog scrolls
with `Up`, `Down`, `PgUp` and `PgDn` under a line that says how many lines are hidden; the
commit box, the command that will run and the buttons stay on screen. The dialog is placed again when
the window changes size, and it stays between the tab bar and the status bar. On the Release view `x`
first asks which packages to hold back and which bump to change (see [`smith release`](release.md)),
then shows this confirmation with the `--hold` and `--bump` it built.

The release confirmation is titled `Release now` and lists, in this order:

1. `Before you release:` three checks. The doctor (no findings, or how many are open, or not checked yet), the last test run of this session (passed, failed or none) and the branch (pushed, commits not pushed, no upstream, no remote). A check that fails is drawn in the warning colour with the warning mark. It never blocks: you decide.
2. `Packages and versions to publish:` each package with its old and new version and its tag, and the packages you held back.
3. `What will happen:` smith makes no commit (it tags the current commit), how many tags it creates, whether the tags go to the remote (and which) or stay on this computer (no remote, or `--no-push`), and that each package goes to its registry.
4. The exact command line, for example `smith release --yes`.

Cancel has the focus, so a hurried `Enter` does not publish; `Esc` cancels too. `y` confirms. The confirming button reads `Release now (y)`, or `Release anyway (y)` when a check warns. While the release runs, the Release view shows the command, one step per package (waiting, building with a spinner, tagged) and `Ctrl+C` stops it. When it ends, a dialog lists the new tags, read from what the release printed, and a line `What next:` (`smith release --status`, or the `git push --tags` to run when the tags stayed local, or what to do after a failure). The run panel of Projects keeps the full output.

The preview of an empty plan does not repeat the headline, and the preview of a plan ends with a pointer to `Release now (x)` and not with the advice to run `smith release` in a terminal.

The dialog opens at once, also when smith must work out the plan first (the packages and versions
of `u` in Deps, the files of a swap, an add or an undo, the packages of a release). While it
waits, the dialog shows a spinner and a plain sentence ("Working out what will change…"), the
seconds after the second one, and the command that will run. `[ Run ]` is faint and does nothing:
`Enter`, `Space`, `y` and a click on it are ignored, so you cannot run a plan you have not seen.
`Esc`, `n`, `q`, `Ctrl+C` or a click outside cancels the wait: smith stops the command that reads
the plan, and a plan that arrives late opens nothing. When the plan is ready, it fills the dialog
and the usual buttons work. The wait shows no steps or percentage, because the command reports none.
`Esc` on the release confirmation ends the wait and drops its plan, but the read of the commits
itself runs on: it is the load of the Release tab, which the tab shows when you open it, so
stopping it would leave the tab empty.

Every screen that reads before it opens waits the same way: the Add a project and Swap a choice
forms (they read the choices of the plugins), the Write a recipe form, the plugin search, the
Compose a stack screen and the check for missing tools. The dialog is titled with the screen, shows
the spinner, the seconds after the second one and "The command that is running:", and `Esc` stops
the command and leaves you on the tab you were on. Changes on the Settings tab (a plugin on or off,
the architecture, a setting, the style, trust) show a spinner line at the bottom of the tab with the
seconds after the second one. A change is a write in this process, so `Esc` says it cannot stop it;
the review that opens before you trust a plugin only reads, so `Esc` stops it and drops its answer.

### Where Esc leaves you

A form that a tab opens (install or remove a package from Deps, add a project, swap a choice) gives
you back to that tab. `Esc` on its first question, `Esc` on the confirmation it leads to, and the
close of a notice about it all show the tab you came from, with its cursor and marks as you left
them. The confirmations that a tab opens itself (update, one version, release, doctor fix) never
leave the tab.

### What a word means

A form that uses a word of smith says what it means on the same screen, in one line, from the
[glossary](../reference/glossary.md) and from the layers of the composer: the choices question explains
`stack`, `framework`, `shell` and `host`, Manage plugins explains `plugin`, Write a recipe explains
`recipe`, and the code layout question explains a `slice` (`vertical-slices`).

### After a change: the result dialog

When a run ends, a dialog says what happened and offers the next steps as choices, each with its
consequence. The first line is always the one that does nothing, because smith never commits or installs unasked.

- **A change that is not committed** (`add`, `remove`, `upgrade`, `derive`, `swap`): "Added app web", what was
  created and how many files, then **Leave them uncommitted** (the default), **Commit only the files this operation changed** or
  **Commit everything git shows as changed** (your own work included). A commit choice reads the files first
  (`smith commit <operation> --only-operation --dry-run`, or `--all`), shows them in the usual
  confirmation, and then makes one commit named from the operation (`chore: add app web`) that `smith undo`
  can revert. The box in the confirmation before the change stays unticked. After a library, the dialog also
  says how an app uses it: press `+` on the app and pick it, or `smith install -p <app> <library>`.
- **An add that could not be checked because a tool is missing**: the project exists on disk, and the dialog
  says so (`api-cs was created in apps/api-cs (12 files)`), then why the first build and test could not run
  (`the .NET SDK is not installed`). Choices: **Open the install page** (copies the link and opens it),
  **Install it with smith** (the tools dialog, which asks how before it runs anything, then returns here),
  **Check again** (runs the first check of the project again, not the add) and **Continue without checking**
  (Doctor keeps reminding you). Esc closes the dialog and changes nothing.

The mouse works in the result dialog and in the tools dialog ("Install the missing tools"): one click
on an entry picks it, as `Enter` does. The last line of the dialog says so.
The dialog covers the content of the open view, but the rule under the tabs and the divider
between the panes stay. Glyphs the log drew itself (a rule, a scrollbar thumb) are cleared with the rest, so no
stray line shows around the dialog.

### Undo, report, why and derived files

These are in the palette (`:`), under Health. None has a key of its own.

| Palette entry | What it does |
|---|---|
| **Undo the last change…** | Runs `smith undo --dry-run` first and shows what it would revert in a dialog. `y` or `Enter` then runs `smith undo --yes`, which reverts it with a new commit; `Esc` leaves everything as it is. When commits are off, the notice reads "Undo is not on" and has a button, **Choose how smith commits**, that opens the question below. When commits are on and smith has nothing to undo, a notice shows its reason; `Enter` closes it. |
| **Open the report** | Runs `smith report -o <file>` with a file in the temporary folder named after the workspace (`smith-report-<workspace>.html`), opens it in your browser and tells you the path in a toast, so you can open it again. |
| **Write the agent file…** | `smith agent file` after a confirmation: `AGENTS.md`, and a short `CLAUDE.md` that imports it. |
| **Check the derived files** | `smith derive --check` in the run panel of Projects: writes nothing, fails when a derived file is out of date or was edited. |
| **Update the derived files…** | `smith derive` after a confirmation: writes the files that follow from the project graph. A file you edited is never overwritten. |
| **Preview the derived files** | `smith derive --dry-run` in the run panel of Projects: the plan of the update, nothing is written. |
| **Why is this here?…** | A form that asks for a file, a folder or a setting (with the [path finder](#the-path-finder) to browse them) and runs `smith why <it>`: who owns a file and how to change it, or why a setting has its value. The Doctor detail of a check about a file opens it with that file filled in. |
| **Divergences from the defaults** | `smith divergences` in the run panel of Projects: every override, ejected file and drift, each with the reason and how to go back. |
| **Try a command in a copy…** | A form that asks for a smith command (`add package billing`) and runs `smith try` on it: the command runs in a temporary copy, doctor checks the result, and the run panel says what changed. Your files stay as they are. |
| **Write the launcher…** | `smith bootstrap` after a confirmation: writes `tools/smith/smith` and `tools/smith/smith.cmd`, so a laptop and CI run the same smith. |
| **Swap a choice…** | Under Projects in the palette. Pick the project, the decision (test, lint, style) and the answer. smith shows the plan of `smith swap --dry-run` first, with the unticked box "Commit this change so it can be undone"; Run does the swap. |

### Undo is asked once

The first time smith is about to change the workspace and nothing says how to treat undo (no
`commit_operations` in `smith.toml`, no saved answer), a box asks "Let smith commit each change
so it can be undone?" before the change goes on. Each choice says what it saves:

- **Yes, commit each change** writes `commit_operations = true` in `smith.toml` with
  `smith config set workspace.commit_operations true --commit`, as a commit of its own, so the
  work tree is clean for the change that waits. The change then goes on to its confirmation.
- **No, leave my files uncommitted** saves the answer for you only, in `tui.json` in the smith
  home, not in `smith.toml`. Undo has nothing to revert.
- **Ask me each time** saves the answer for you only. Each change then shows the box "Commit this
  change so it can be undone" in its confirmation.
- **Cancel this change** (`Esc`) changes and saves nothing, and smith asks again at the next change.

The palette entry **Undo the last change…** opens the same question from its "Undo is not on" notice.

### Environments, cache, pipelines and the rest of the palette

Every command of smith that is not a dev server, `serve`, `mcp`, `lsp` or `agent` has an entry in
the palette (`:`). A command that changes files shows its plan first, as every other change does.

| Palette entry | Section | What it does |
|---|---|---|
| **Environments of the web apps** | Projects | `smith environments list` in the run panel of Projects. |
| **Add environments to a web app…** | Projects | Pick the web app (the selected project first), type the names. The plan of `smith environments add --dry-run` lists the files it writes. |
| **Remove environments from a web app…** | Projects | Same form. The plan says that the `.env.<name>` file stays and that the default environment cannot go. |
| **Make the icons of a Tauri app…** | Projects | Asks for the square PNG and, when more than one app has a `src-tauri`, which app. The plan lists the icons it overwrites. |
| **Add a ready-made pipeline…** | Release | Pick e2e, package-zip, deploy-pages or deploy-azure-function; a deploy pipeline asks for its app. The plan lists the files. |
| **Set the Docker Hub descriptions…** | Release | Asks to name the repositories only (the default, nothing is sent) or to set the descriptions. |
| **Cache: status**, **check for damage** | Health | `smith cache status` and `smith cache verify` in the run panel of Projects. |
| **Cache: remove what is broken…**, **trim to its limit…**, **delete it all…** | Health | `smith cache verify --fix`, `prune` (with its plan) and `clear`, each after a confirmation that says the next run is slower. |
| **Copies between projects**, **Tools the workspace needs** | Health | `smith copies` and `smith tools` in the run panel of Projects. |
| **Why would a task run again?…** | Health | Asks `project:task` and runs `smith why-miss`. `w` on a task of the run panel opens it with that task filled in. |
| **Connect smith to a coding assistant…** | Health | Tick Claude Code and VS Code; writes `.mcp.json` and `.vscode/mcp.json` after a confirmation. |

In **Add project**, a web app gets ONE question about its environments, not four. The labels
below are fixed so that the VS Code extension uses the same words:

| Option | What it does |
| --- | --- |
| `Use these: lcl, dev, uat, prd` (default, first) | Ends the questions. The hint says: smith run build builds prd, the builds go to `dist/apps/<name>/<env>`, and the `.env` files are written for you and kept in git (smith commits nothing). |
| `Change the names…` | Asks the names (pre-filled with the default four, edit the list), then which one `smith run build` builds (prd, else the last name, marked default). |
| `Change the build default, the output or the .env files…` | Asks which environment `smith run build` builds, where the builds go, and whether the `.env` files are kept in git. |
| `One build only` | `--environments none`: one Vite build. |

The command always carries the answers (`--environments`, `--default-environment`,
`--environment-output`, `--env-files`), so the plan shows the default that was taken. The CLI keeps
its flags and its own defaults.

After the name, **Add project** asks one question, like a new workspace does: `Start with the
defaults` (first, default) goes straight to the plan, with the pick of the first question ticked,
the shape of the other projects followed, and for a web app the environments above; every default
is a flag in the command. `Customise` asks the questions one at a time. A question that smith
cannot answer for the person (the trigger of a function app) is asked either way. `Make my own
recipe…` leaves the form for **Write a recipe** (`smith recipe new`), the same form the Start tab lists.

A framework, shell or host picked in the question before the name is ticked in the choices
question and is in the command (`--with react`). The cursor starts on "Decide in the next
question" only when nothing was typed to search for: Enter then ticks nothing.

`Esc` on the plan of an add goes back to the last question, with the answers filled in.

**Write a recipe** asks, for each level that takes one plugin, a pick list whose last answer is
"Chosen by the people who use it". Then it asks the question the users read, which plugins they can
end up with and the name of each answer, and writes a `--chosen` flag (see
[`smith recipe new`](recipe-new.md)).

**Release without a releasing commit** does not list a package that was never released: smith has
no version to bump. A line names it and gives the commit that sets its first version, `git commit
--allow-empty -m "version(<name>)[1.0.0]"`.

While smith convert or smith new runs from the form (before a workspace exists), the box is titled with what it does (Converting this workspace to smith…, Creating a new workspace…). A spinner line says what it is doing, with the time once it passes 2 seconds, above the last lines it printed.

After an add, a convert or an adopt finishes, the toast says it can be undone and where.

**Derived files on the Projects tab.** When the workspace turned derived files on
(`[derive] enabled = true`), the detail of a project has a **Derived files** section with one
row per file and its state in words: *up to date*, *out of date*, *missing*, *not recorded
yet*, *edited by hand*, *differs from what smith derives*, *yours* (ejected) and *the
project's own* (it differs from its sibling projects). The heading counts the files that need
action, and the Workspace detail (no project selected) counts them for the whole workspace.
The Doctor tab keeps reporting them as checks. The actions of the detail run the matching
command, and each one that changes a file shows its command line in a confirmation first:

| Action | Command |
|---|---|
| Explain the derived files | `smith derive --explain <file>` for each file (owner, renderer, why it is in its state) |
| Update the derived files | `smith derive` (the whole workspace) |
| Keep my edit of a derived file | `smith derive --adopt <file>` for each hand-edited keys file |
| Own the files I edited | `smith derive --eject <file>` for each file that waits for a choice |
| Replace the files I edited | `smith derive --replace <file>` for each file that waits for a choice; the edit is lost |

An action that no file of the project takes is shown off, with the reason under it. The
**Why** action of the same detail runs `smith why <project folder>`.

## Settings

The Settings tab (key `6`) is where you change things without leaving the interface. A list
of sections is on the left; the open section's rows and, under them, what the selected row is
and does are on the right. `Tab` (or `Left` and `Right`) moves between the two; a click does
the same, and a double-click does what `Enter` does. Everything it writes is what the
matching command writes: a test compares `smith.toml` byte for byte with the result of
`smith plugin disable`, `smith plugin trust --pin` and `smith config set`, and comments and
every other line stay as they are. The name, type, description, default and current value of the selected row, and the key that changes it, are in the detail beside the list on a wide screen. A setting whose default has a recorded reason shows it there
as "Why:", the same text as `smith why <setting>`.

**Plugins.** The built-in language plugins, each on or off (`Space`; at least one must stay
on, as with `smith plugin disable`). The status line says what changed and offers `undo (u)`:
`u` switches the plugin back, and any other key lets the offer go. Switching a language off also
shows, in a toast, which projects are no longer built, tested or checked. Below them, the external plugins the workspace declares
in `smith.toml`, with whether each is trusted and pinned: `Enter` (or `t`) shows the review
`smith plugin trust` shows, which is what the plugin is, where it comes from, the files its
command runs with their sha256 and the environment it may see, and `y` trusts it, `p` trusts
it and writes its sha256 to `smith.toml` for CI, `Esc` leaves it untrusted; `f` forgets a
decision (`smith plugin untrust`). A trusted plugin runs with your privileges and is not
sandboxed. Then the plugins installed on the plugin search path, and a row to install a
plugin. `Enter` on that row asks where the plugin comes from (a directory, an archive, an https
URL or `github:owner/repo@tag`; an archive or URL needs its sha256 typed after the source) and
shows the command line as you type, `smith plugin install <source> [--sha256 <digest>]`. `Enter`
again shows it in the confirmation dialog, and `y` runs it in the run panel of Projects, so its output is the
command's own. The plugin is checked and never started; trusting it is the review above. The tab
reads the plugins again when the command ends. The row above it, **Find a plugin in the
marketplace…**, asks for words (or nothing, to list every plugin), runs
`smith plugin search <words> --json` and shows the plugins found as a list you can narrow by
typing, with the newest version, what it does and its roles. `Enter` on one shows the
`smith plugin install <name>` confirmation, as above. An empty result says to search with fewer
words, and a failed search shows the command's error in a notice. Under them, **Manage plugins**
opens the `smith plugin set` form, the one the Start tab lists.

**Architecture.** How the code of the workspace is laid out, as radio rows with what each
choice means beside it: `vertical-slices` (code grouped by what it does, one flat folder per
outcome, files named for their role) and `none` (smith checks no code layout). `Enter`
chooses the highlighted one and saves it in `smith.toml` at once; the choice in use is marked
`(●)`. It writes what [`smith architecture`](architecture.md) writes: choosing
`vertical-slices` adds `[conventions.vertical-slices]`, and choosing another removes it
together with the settings under it. The panel under the rows says in plain words what the
highlighted choice does, which languages it applies to, and which projects of this workspace
it would check or skip; a choice that would check none of them asks for a second `Enter`. More architectures will be offered here as smith learns them.

**Style.** The colour theme (`auto`, `dark`, `light`, `plain`) and the symbols (`auto`,
`unicode`, `ascii`), with a preview of the highlighted choice. `Enter` applies the choice at
once to the whole interface and keeps it for you, in `tui.json` in the smith home
(`SMITH_HOME`, or `~/.smith`), not in `smith.toml`: it is a personal preference. The terminal
interface reads it when it starts; the command line does not use it. The variables in
[Plain mode and environment](#plain-mode-and-environment) still win over it, and the section
says so when one is set. `auto` means what smith did before: the terminal's background and
`TERM` decide.

**Workspace.** Every setting [`smith config list`](config-list.md) knows, grouped by the
table of `smith.toml` it belongs to, with what it does and its default. A switch shows a box
(`[x]` or `[ ]`) and flips with `Enter` or `Space`; a setting that takes one word of a fixed set
(`ci.attest`, `ci.audit`, `release.api_check`) opens a pick list: the arrows move, `Enter` picks, `Esc`
cancels; a list with fixed choices (`ci.providers`) is ticked with `Space` and saved with
`Enter`; text, numbers and lists without fixed choices (separate the items with commas) are
typed. The value is checked against the schema before anything is written, and the reason
stays in the editor when it is refused. `u` removes the setting, so it has its default.
`plugins.enabled` is edited in Plugins, and stacks, tasks and projects are edited in the file.

**Recipes.** The `*.recipe.toml` files in the workspace, with their name and version and why
one is not valid; [`smith recipe check <file>`](recipe-check.md) shows what a recipe would use.

**Assistants.** The forms that connect smith to a coding assistant (`smith mcp install`) and
install the agent files (`smith agent install`). `Enter` opens the form, the same one the Start
tab and the palette open.

**CI.** The forms that regenerate the CI pipelines (`smith ci generate`) and add a ready-made
pipeline (`smith pipeline`). `Enter` opens the form. The doctor still offers its own fix when a
pipeline drifts from what smith would write.

**Forms.** The guided forms (add a project, remove a project and the rest of the Start menu)
that end as a smith command line. Every action in the action catalog has one: a command with
no hand-written form gets a form built from its own arguments and options, subcommands
included (`deps up`, `docker description`, `slices check`, `cache prune`). **Write a recipe**
asks for the name, version and parent, then one question for each level a plugin can fill. A
level with one plugin (language, scaffold, lint, test) is a pick list that starts on "Leave
unset". A level with any number (structure, deploy) is a list of tick boxes (space ticks, enter
confirms; nothing ticked leaves the level unset). Each plugin shows a line about what it is, and
after the language only the plugins that fit it are listed. The list of plugins comes from
`smith recipe new --list-levels --json`. A pick is refused with its reason, and the
confirmation shows the exact `smith recipe new --name ... --fixed ...` command and the file it
writes. This form fixes plugins only: a level whose plugin the recipe's users choose is still
written with `smith recipe new` on a terminal, which asks the questions.

Home points at the tab (`New here? ...`) until you have opened it once; that is remembered in
`tui.json`.

## The Doctor list

The first line of the Doctor view holds three filter buttons, one for each state: `× 2 failed (x)`,
`▲ 1 warning (w)` and `✓ 20 passed (p)`. Each is a toggle: click it or press its key. A filter that is
off hides the checks of that state in every group; the group headings still count all of their
checks. The failed checks and the warnings start on. The passed checks start off, as one line
(`20 checks passed`) that `p` or `Enter` opens. When the filters hide every problem, the list says
`No checks match the filters. Turn one on (x, w or p).` The filters last while the Doctor view is
open, also across a new check. The footer names what each key does now: `hide failed` or `show failed`.

## The Doctor detail

The detail of the selected check has four parts, always in this order, and each fact appears once:

1. **State**, under the check's name: the icon and the word only (`× failed`, `▲ warning`, `✓ passed`).
2. **Actions**, as an [action panel](#action-panels): `[ Fix  f ]` (only when the check has a fix), **Run this check now** (only for a check that a command runs on its own: the boundary checks run `smith boundaries`), `[ Check again  r ]` and **Why is <file> here?** (only for a check about a file: the file its fix writes, or a path its problem names). Why opens the **Why is this here?** form with the file filled in, so `Enter` runs `smith why <file>`. Each button shows its key. A click on a button chooses it. Under the rule, a plain line says what the chosen action does (for a fix, what it will change), and the run button under that line runs it. The key of an action runs it at once.
3. **What is wrong**: the problem the check found. A long problem shows its first three items; `e` shows all.
4. **How to fix**: the remedy, then what the fix will do (the projects it covers, the files it writes, the commands it runs, the questions it asks), and last the same fix in a terminal (`or in a terminal: smith doctor --fix --only '<check>'`). A check without a fix has no Fix button; its remedy says what to do.

`PgUp` and `PgDn` scroll the detail (`↑ more` and `↓ more` tell when there is more). Under 100 columns, `z` shows the detail alone.

Under the checks, **Checks doctor does not run itself** lists the checks that only run on request:
**Vertical-slice layout**. `Enter` on the row, or the **Run this check now** button of its detail,
runs the check at once, with no question. It runs `slice-check`, the cached task of each project in a
language the checker knows (Go, Python, Rust and others), and `lint` when the workspace has Node
projects, because ESLint's vertical-slices rules check a Node project's layout as part of its lint.
A workspace with a Go app and Node packages runs `smith run slice-check lint`. Each project's
result, and the output of a failure, show in the run panel. The row is absent when no project has
either task.

## Wizards

### The doctor fix wizard

Opened by `f` or `F` in Doctor, or by **Fix smith.toml** on the Start menu when `smith.toml` uses a renamed key (the same fix as `smith doctor --fix` makes; after it is applied the shell starts again inside the workspace). It is the interactive form of
[`smith doctor --fix`](doctor.md) and asks the same questions.

1. **Questions.** A fix may ask something (install the missing tool, or remove the script?).
   `Up` `Down` (or `j` `k`, `Tab`) move, a digit picks that option, `Enter` or `Space`
   chooses, `Esc`, `b`, `Left` or `Backspace` goes back one question, `q` or `Ctrl+C` leaves.
   On the first question, with no screen before it, `Esc` leaves.
   An option the plugin cannot offer is greyed, with `Disabled: <reason>` under it; a digit or `Enter` on it
   moves the cursor there and picks nothing.
2. **Review.** What will change: one line for each file (what happens and how many lines), the
   commands it will run, the projects it covers. `d` shows the diffs and hides them again.
   `Enter` or `y` applies; `Esc` or `b` goes back to the last question (or leaves when there is
   none); `n`, `q` or `Ctrl+C` leaves; `Up` `Down` `PgUp` `PgDn` `Home` `End`
   scroll. In the adopt wizard the review of a conversion has the same unticked box, "Commit this
   change so it can be undone": `c` or a click ticks it.
3. **Progress.** A spinner and the fix title, then "step n of m" with the step's name (for
   **Fix with doctor**, the steps of all the fixes together). After 2 seconds the line adds the
   time it has run. Under it, six rows (always there, so the dialog keeps its height) show what
   smith writes ("Writing package.json") and then the last six lines the running command
   printed. `l` (or a click on "l full log" in the hint line) opens the whole log of the step in
   a scrollable view: the arrows, `PgUp`, `PgDn`, `Home` and `End` or the wheel scroll it, and it
   follows the newest line until you scroll up. `Esc` or `l` closes the log; the step keeps
   running. Other keys do nothing while it runs, so a second fix cannot start. `Esc` or `Ctrl+C`
   stops the step; when the log is open, `Esc` closes the log first and `Ctrl+C` still stops.
   The log keeps the last 2000 lines.
   A fix can be a smith command (`smith upgrade --defaults`, `smith deps up --align`,
   `smith derive --replace`, `smith kind`, `smith cache prune`). The review then shows the plan
   the command prints with its plan flag, so you see what it changes before you agree, and
   Doctor checks again after it ends. Nothing runs in a terminal: the listing in
   [the terminal hand-over reference](../reference/tui-terminal-handover.md) says what is left.
4. **Failure.** The failing step's output. `r` or `Enter` retries, `y` copies the failure
   text, `Esc`, `q`, `c` or `Ctrl+C` closes.

### The adopt wizard

Opened with `a` from the Start menu, or by Enter when the folder holds files but no
workspace, under the one name "Bring this repository into MonoSmith". It names what it
found first (a plain repository, or an Nx or Turborepo workspace) and does the work of
[`smith adopt`](adopt.md) or [`smith convert`](convert.md) to fit; it uses the fix
wizard's keys and drawing. The stages:

1. **Reading** the repository.
2. **What was found**: the kind of repository (Nx, Turborepo, or plain), the projects
   (the first twelve, with their language in words), the tools in use today, and blockers
   (`×`) and warnings (`!`). Nothing has changed yet. When the packages are not installed, or
   are not the ones the repository lists, the button is **Install the packages and
   continue**: smith runs `npm ci` (or `npm install` without a lockfile), reads the
   repository again and goes on. git cannot take an install back, and the screen says so.
   For a plain repository whose project is at the root, the wizard asks **where it goes**
   (apps, packages, libs, another folder, or leave it), then **what goes where**: one
   question for each development dependency smith does not know as a tool (the project,
   or the workspace) and one for the project's own ESLint setup (keep it, or use the shared
   config). The first answer of each changes the least, `b` goes back, and the answers reach
   the step as `--hoist` and `--eslint-setup`. Nothing is asked when smith can place everything.
3. **How far to go**: conservative (keep the build and lint setup) or full (make smith the
   only tool; Nx and Turborepo only, like `--level full`).
4. For the full level: a **trial conversion on a scratch copy** to see what would be left, then
   **the four choices** (lint, test runner, build, CI), as in `smith convert --lint --test
   --build --ci`.
5. **The whole plan**, for confirmation, from a trial conversion on a scratch copy. It
   starts with **In short** (how many files are created, edited and removed, how many things
   to check, how many things behave differently), then **Check these before you go ahead**
   (every item, with its reason), **Behaves differently afterwards** (the first five; the
   rest go to `.smith/convert-report.txt` once it is converted), then the files in detail
   (a file that changes in many folders is one line) and, at the full level, the removals.
   `Enter` converts. Nothing is written before.
6. **Converting**, then, for the full level, a question per **group of removals** (`We are
   going to delete N packages and M files. ... ok?`): `y` deletes them, `n` keeps them,
   `s` keeps these and skips the groups after.
7. **Finishing**: the install, then the proof that the conversion works, at both levels:
   the build, the tests and doctor (and lint at the full level, where smith owns it). A
   failed check stops here with its output and Retry.
8. **Done**, in three parts: what is done and what was proven (`It works: everything builds
   (9.8s), every test passes (31s), doctor finds nothing wrong (4s)`, then how many things
   need you), **Needs you** (what the conversion could not decide for you, each with its
   reason) and **Behaves differently now** (the first five, the rest in
   `.smith/convert-report.txt`), then the files that look like old tooling but stay.
   **Open the workspace** (`Enter`) reopens the interface inside the new workspace; **Read
   the report** (`s`) shows `.smith/convert-report.txt` inside the wizard. When no check
   ran, **Open it and build everything** (`n`) starts the build once the workspace is open.

While a job runs, `Ctrl+C` or `Esc` stops it. A failed step shows its output with Retry.
Everything is in the working tree and git undoes it (installed packages stay installed).

A new workspace made from the Start menu (`n`) opens the same way: when `smith new` ends, the
interface starts again inside the new folder, and Home offers to add the first project.

### Guided forms

Entries of the Start menu and of the palette that need arguments (create a workspace, add a
project, manage plugins, remove a project, and forms generated from the command catalogue)
ask one question at a time: a choice (`Up` `Down`, `Enter`, `Esc` back one step), several
choices (`Space` toggles, `Enter` confirms) or free text with a placeholder. The answers
build the command line, which the shell runs and shows.

### The path finder

A question that asks for a path (a file, a folder, or for "Why is this here?" a setting such as
`ci.devcontainer`) shows a list under the text field: the folders and files of the folder you are
typing in, starting where you started smith. What you type is still the answer; the list only helps,
so a path typed in full always works, even one that is not in the list.

| Key | What it does |
|---|---|
| type | Narrows the list to the names that hold your text; the names that start with it come first. |
| `Up` `Down` `PgUp` `PgDn` | Moves the highlight. |
| `Tab` | Puts the highlighted name in the field; a folder ends with `/`, so the list goes inside it. |
| `Enter` after `Up`/`Down` | Opens the highlighted folder, or chooses the highlighted file or setting. |
| `Enter` without moving | Confirms the text as you typed it. |
| `..` row, `Backspace` | Goes up one folder. |
| click, double-click | Click highlights a row; a double-click does what `Enter` does on it. |

A folder is shown with a closing `/`. Folders that git ignores (`dist`, `node_modules`, anything in
`.gitignore`) are shown dimmed with "git-ignored". Names that start with a dot appear when you type a
dot. Only the folder you are in is read, never the whole tree; a folder with many names shows the first
rows and a line "N more: keep typing to narrow the list". The list fits 80 columns.

The finder is one component (`apps/smith/tui/pathfinder`). It is used by: Why is this here? (files,
folders and settings), the logo of Tauri icons, the recipe a new recipe extends, every argument or option of
a generated form whose name says file, folder, directory, path or recipe (for example
`recipe check <file-or-url>`, `plugin draft <sample-folder>`, `--dir`, `--cache-dir`, `--recipe-file`), and
in the Settings tab the path settings (`cache.dir`, `workspace.recipe`) and the source of
"Install a plugin". Choices from a fixed list, project names (`add --like` is a list of the projects) and
names (a project, a workspace, a package) do not use it.

**Create a workspace** asks the name, then "How do you want to set it up?". **Start with the defaults** (the
first line) shows what will be chosen (CI, language, code layout, agent files, dev container), says that CI is
where your tests run on every push, and tells how to change each later; with it the name is the only answer.
**Customise** asks the questions one at a time. Either way, smith asks what the first commit holds when the folder
has files and no git repository, and about files that look like secrets.

**Compose a stack** (palette, or `A` in the Projects view, or Home when the workspace has no project) is one screen to search languages, frameworks, desktop shells and cloud hosts, tick them, see each conflict explained, and end in the plan; see [`smith compose`](compose.md).

**Add a project** starts with "What are you building?": a website or web app, an API or backend
service, a function app (Azure, AWS, Google), a desktop app (Tauri, Electron), a library for
my other projects, or a package I publish. Each line shows what people usually pick and one
sentence about what you get. Type to narrow the list with the words people use: `react`,
`azure function`, `electron`, `tauri`, `lambda`, `api`. The words come from the same catalogue
as the composer, so a plugin that adds a framework or a shell makes the line findable. The
outcome answers the kind. The language is asked only when more than one fits (a website is
Node only, so none is asked; a function app asks Node, Python or C#), with the line
"Preselected: N of the M projects here use X." when that language is in the list. A website,
a function app and a desktop app then ask which framework, cloud or desktop shell; what you
type in the first question starts that list on the match, and the pick starts ticked in the
choices below. The last line, "I know the language: show every language", keeps the long way:
the kind, then the full list of languages. After the outcome and the language, the form asks
the name, then, from what the plugins
publish (`smith choices`), a list of choices for that language: a framework, a test runner, a
style. Tick any with `Space`; `Enter` with nothing ticked uses the stack's defaults. The
list shows only the choices that fit the kind: a package is offered lint and test choices, not
a framework or a desktop shell. `Enter` checks the ticked set before the form goes on, and says
in plain words what is missing (a website needs a framework, a style needs the framework it
goes with) instead of failing at the end. Lint and test choices alone are enough for a package:
the stack's own scaffold writes it. The choices become `--with a,b`. A workspace with three or more projects is also asked whether the
new one should do what the others do (the default), follow one project you pick from a list
(`--like <project>`) or take the stack's own defaults (`--fresh`). Then the shell runs `smith add ... --dry-run` and shows its plan in a dialog:
the files it would write and, when it applies, which sibling project it follows
(`follows web: same test runner`). `y` runs the real add, `Esc` writes nothing. A wrong
combination (a choice for another language, one that needs a choice you left out) shows what
smith said, with its fix, in a notice instead of the plan. A command that needs a terminal of
its own (it asks questions) is handed back to smith to run after the shell closes; the shell
prints `→ smith <command line>`.

## Action panels

A detail pane with two or more actions, of which at least one has a description, shows its actions
in two parts. This rule is the same in every detail pane: Projects (the tasks and the More forms) and
Doctor (Fix, Run this check now, Check again, Why).

1. **The top part** is one compact button per line, with no description. A button here only chooses
   an action. The chosen action has the selected look. The first action is chosen until you choose
   another one, and a new selection starts at its first action again.
2. **The bottom part**, under a thin rule, says what the chosen action does: what runs, what runs
   first, whether smith remembers the result, and the `smith run …` line. Under the description is
   the one button that runs it. That button has the label and the key of the chosen action.
   When the pane is too short for the whole panel, this part stays at the bottom of the pane and
   the buttons above it scroll, so the description and the run button are always in view. A
   description too long for half of the pane is cut and ends with `…`.

With the keyboard (after `Tab`, or in the opened detail), `Up` and `Down` choose an action. `Down`
after the last action moves to the run button, and `Up` goes back. `Enter` on an action moves to the
run button, and `Enter` on the run button runs the action. The footer says which: `go to run (Enter)`
on an action, `run (Enter)` on the run button. The shortcut letters (`b`, `t`, `l`, `c`,
`p` and the others the buttons show) do not choose: they run the action at once. They are the fast
path that the footer and the help promise.

A panel with one action, or with actions that have no description, is not split. Its buttons run at
a click, for example **Run this check now** of a check that doctor does not run itself, or **Doctor**
and **Preview the release** in the Workspace detail. The Deps detail has no actions, and the Release
view draws its buttons in the view, not in a detail pane.

## The mouse

**Buttons.** Every action in a detail pane, and the buttons of the Projects list (`New project`,
`Compose a stack`) and of the run panel (`Back`, `Stop`, `Run again`, `Run unfinished`, `Copy log`), is a button that shows its key in
parentheses, for example `Test app (t)`. A button has three looks. It rests; it is lit while the
pointer is over it; and it is pressed while you hold the mouse button on it. The action runs when you
let go over the same button. Press a button, drag away and let go, and nothing runs. In an
[action panel](#action-panels) a click on a button above the rule chooses the action, and the run
button under the rule runs it. A muted line under a single button says what it does. A task that the keymap has no key for (`package`, `start`, a task you wrote)
gets a letter that no other key of the view uses, shown on its button. In the opened detail on a
narrow screen the row at the keyboard cursor carries the cursor mark, and `Enter` runs it (in an
action panel, `Enter` first moves to the run button).

The mouse is on by default: click a tab, a row, a footer hint, a detail action or a dialog
button; click an answer of a single-choice list to pick it; double-click a project or a row of the Settings tab to open or change it; use the wheel to scroll. The wheel scrolls
a list and never moves the selection, and a thin scrollbar in the last column of the content (a dashed line with a solid, muted thumb, `#` over `|` in the ASCII theme) shows where the visible part sits in a list or log that is longer than its pane; a click selects a row without moving the list. The same rule holds in every list of a form
(the choices of Add a project, the languages, the plugin lists of a recipe) and in the Compose a
stack screen: the wheel scrolls and the selection stays; Up, Down, PgUp, PgDn, Home and End move the
selection and keep it in view; a scrollbar and a line "↑ 3 more · ↓ 12 more" say what is hidden.
The words list under the choices question folds to one line, "what do these words mean? press ?",
whenever the pane would leave the list fewer than 16 lines (the height of the pane decides, not that of the
terminal; `?` opens it); the list always keeps at least 8 rows. The confirmation of an add says the choices in
words and marks the defaults: "React, ESLint (default), Vitest (default); no desktop shell, no style". With nothing
ticked, the project takes the stack's defaults and the confirmation says so. While the confirmation is open the
tab bar names the form ("Add a project"), not "Start"; "Decide in the next question" leaves the choices list empty. In Deps, click a row's box to
mark it, Shift+click to mark the rows from the last one to it (Shift+Up and Shift+Down do the same from the keyboard),
and click the box in the heading to mark every row shown, or none. Shift+drag still selects
text in most terminals, and `m` turns the mouse off.

**Tick boxes and pick lists.** Every list of choices and every tick box in smith is the same
component, so it looks and acts the same everywhere: Compose a stack, the guided forms (Add a
project and the others), the doctor fix wizard and the dependency repair, the result dialog after a
change (with the cache clear choice), the missing tools dialog, the remote cache wizard, the
release adjust list, the Settings editors and the commit box of a confirmation. The row at the
keyboard cursor has the cursor mark (`›`, `>` in the ASCII theme) and an accent label. A box reads
`[ ]`, `[x]` or `[-]` (part ticked, or cannot be ticked), and a one-of-several choice reads `( )`
or `(●)`. A row that cannot be chosen now is muted. The row under the pointer takes the row hover
look. With the mouse:

- A click on a tick box row ticks or unticks it (Compose a stack, a form with several answers,
  a list setting in Settings, the box of a Deps row, the commit box). In Compose a stack and the Settings
  list editor, a click is the same as the key that ticks (`Space`).
- One click on a row of a list with one answer picks it, as `Enter` does: the guided forms, the
  doctor fix wizard and the dependency repair, the result dialog, the missing tools dialog, the
  remote cache wizard, a pick-list setting in Settings and the open question of Compose a stack.
  There is no "click moves, double click picks" anywhere. The path finder is a browser, not a
  choice: there a click selects a folder and a double click goes into it.
- `Space` ticks a tick box in every list. In Compose a stack the search box takes the typed
  letters: `Down` puts the keyboard on the list, where `Space` ticks the row under the cursor, and
  a letter puts it back in the search box. `Enter` goes on to the next step there, as `Ctrl+N` does.

**Key hints.** The line of keys under a list or a question names each key the way the buttons and
the footer do (`Enter`, `Esc`, `Space`, `Tab`, `↑↓`), then what it does, with ` · ` between:
`↑↓ move · Enter or a click picks · Esc goes back`.

**Hover.** Everything you can click shows a hover look while the pointer is over it. A button,
a tab and a footer hint take the button hover colour, which is between the resting colour and the
pressed (or selected) colour. A row that a click selects or ticks (a project, a check, a
dependency, an option of a form, a run) takes the row hover background, which is between the
terminal's own background and the background of the selected row. The selected row keeps its own
look. Nothing behind a dialog shows a hover. Without colour (`NO_COLOR` or the plain theme) a
hovered button is in reverse video and a hovered row is underlined. smith draws the hover at
most about 40 times a second, so a fast pointer does not slow the screen.

### Select text

Drag across text with the left button to select it, with the mouse on. When you let go, smith
copies the plain text to the clipboard (OSC 52, as `y` does) and the toast says so. The selected cells stay lit
until you press Esc or any other key (Esc only clears the highlight; another key also does its work), press
again, scroll the wheel, change tab, resize the terminal, or the text under the highlight changes.

- A drag stays inside the panel where it started: the list or the detail; in Settings the sections, the
  settings and their help; in the run panel the task list and the log. The rule, the margins, the header, the tabs and the footer are never selected.
- A press on a button, a tab, a divider or a dialog starts no selection. A plain click selects nothing.
- Without colour (`NO_COLOR`) the selection is drawn in reverse video.
- Shift+drag, and `m` with a plain drag, still select with the terminal itself, across panels.

## Plain mode and environment

| Variable | Effect |
|---|---|
| `CI` | Any value stops a bare `smith` from opening the shell. |
| `SMITH_SCREEN_READER` | Anything but empty, `0`, `false`, `no` or `off` draws no full-screen shell, no colour and no mouse: plain numbered questions, then the command, which smith runs in the terminal. |
| `SMITH_THEME` | `light`, `dark` or `plain`. `plain` also selects plain mode. Without it the choice saved in the Settings tab decides, and without that the terminal's background. |
| `NO_COLOR` | Any value gives the plain colours, and the tools a run starts print plain text. Without it, the log of a run shows each tool's own colour and smith paint (go test pass green, fail red). The interface hands this on to the smith it starts with `SMITH_TOOL_COLOR=1`; output the interface reads (a plan, a dry run) stays plain. |
| `SMITH_ASCII` | Any value (or `TERM=dumb` or `TERM=linux`) uses ASCII symbols instead of Unicode. Without it the symbols saved in the Settings tab decide. |
| `SMITH_HOME` | Where the smith home is (`~/.smith` by default): installed plugins, trust decisions and `tui.json`, the theme and symbols chosen in the Settings tab. |

When standard output is not a terminal the plain questions go to standard error so you can
still read them. The plain questions have no Settings tab: there, change plugins and settings
with `smith plugin enable|disable|trust` and `smith config set|unset`, and the theme with the
variables above. In the full-screen interface with `NO_COLOR` or the plain theme, the Settings
tab keeps every state as a word and a symbol (`[x]`, `trusted`, `not trusted`), so colour is
never the only signal.

## Related

[`smith doctor`](doctor.md), [`smith adopt`](adopt.md), [`smith convert`](convert.md),
[`smith run`](run.md), [`smith release`](release.md), [`smith up`](up.md),
[The architecture of the terminal interface](../architecture/README.md).
