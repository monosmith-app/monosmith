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
| In an empty folder | The **Start** menu: create a workspace (`smith new`). |
| In a folder that holds files but is not a workspace | The Start menu with **Adopt this repository** first, so Enter starts the adopt wizard. |

When the shell cannot be drawn (standard output is not a terminal, `SMITH_SCREEN_READER` is
set, or `SMITH_THEME=plain`), smith asks the same questions as plain numbered lines and runs
the command you end up with; see [Plain mode](#plain-mode-and-environment).

## The screen

From top to bottom:

1. **Header**: the workspace's name (`smith` when there is none, with a hint about
   `smith new` and `smith adopt` outside a workspace), the git branch, and a health summary:
   `checking…` until the first doctor run has finished (it starts with the shell), then
   the counts of failed checks and warnings, or a check mark when everything passed.
2. **Tabs**: Home, Projects, Run, Doctor, Deps, Release, Settings, numbered 1 to 7. When the names do
   not fit, they are shortened, then the strip scrolls with `‹` and `›`. On a short terminal
   the tabs are one row of names.
3. **The view.** Projects, Run and Doctor show a list with a **detail pane** beside it; on a
   narrow terminal the detail replaces the list when you press Enter, and Esc goes back.
4. **Footer**: key hints for the open view, then the ones every view ends with
   (`views`, `palette`, `help`, `quit`, `mouse`). Each hint can be clicked.
5. **Status bar**: the exact command line running or last run, its result, the git branch
   and the smith version.

A short **toast** in a corner reports a result, a refusal (with the reason) or a warning, and
goes away by itself. A result toast belongs to the view it was raised on.

## Views

| View | Key | Shows |
|---|---|---|
| Home | `1` | The health of the workspace in one headline, the facts behind it (checks, affected projects, the last run, outdated dependencies) and the one thing to do next, with its key. |
| Projects | `2` | Every project, ordered by kind and then name, with its language; the detail pane shows its properties and the actions that apply to it. |
| Run | `3` | The tasks of the current or last run: queued, running, cached, passed, failed or skipped, with durations, a progress bar and each task's log. |
| Doctor | `4` | The checks of [`smith doctor`](doctor.md), grouped by project or flat, each with its problem and remedy; fixes open the fix wizard. |
| Deps | `5` | Dependencies with newer releases, grouped patch, minor and major, and the one-version policy's verdict, with the `smith deps up` command line for a selection. |
| Release | `6` | What the next release would do (the plan of [`smith release --dry-run`](release.md)), the commits behind each bump. |
| Settings | `7` | Plugins, style and the settings of `smith.toml`, changed in place; see [Settings](#settings). |

Home chooses the **next action** by priority: what is broken first (`doctor could not run`,
failing checks with a fix ready, failing checks), then a last run that failed, then work the
changes left (`N projects affected: press w to lint and test them`), then warnings, then
outdated dependencies, and when there is nothing, `all good: press b to build`. Enter does
the suggested action.

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
| `1` to `6` | Show Home, Projects, Run, Doctor, Deps, Release. |
| `Up` `Down`, `j` `k` | Move in a list. |
| `Enter` | Open or run the selection. |
| `Esc` | Back, close, clear. |
| `?`, `F1` | The shortcuts overlay (Up, Down, PgUp, PgDn, Home, End scroll it; Esc, `?`, `F1`, `q` or Enter close it). |
| `:`, `Ctrl+P`, `/` | The command palette. |
| `q` | Quit (not while a run is going: `Ctrl+C` stops it first). |
| `Ctrl+C` | Stop the run in progress; pressed again, quit. |
| `m` | Turn the mouse on or off. Off lets the terminal select text; Shift+drag selects with it on. |
| `y` | Copy the focused item (remedy, command, project path or package) to the clipboard, through the terminal (OSC 52), so it also works over SSH and in tmux panes that allow it. Works in Projects, Run, Doctor, Deps. |
| `r` | Read the view again (Home, Doctor, Release, Deps). |

### Run tasks (in a workspace)

| Key | Does | Where |
|---|---|---|
| `b` | `smith run build -p <project>` on the project selected in the list; the footer names it (`b build orders`) | footer in every workspace view |
| `B` | `smith run build` on every project; the footer counts them (`B build all 23`) | footer in every workspace view |
| `t` | `smith run test -p <project>`, same; `T` tests every project | footer in every workspace view |
| `L` | `smith run lint` on every project | palette |
| `l` | `smith lint` | footer in Projects |
| `c` | `smith typecheck` | footer in Projects |
| `w` | `smith run lint test --affected` | footer in Home, Projects |

The run opens the Run view. The keys work in every workspace view; the footer shows only
the ones that make sense for the open view (the column above). The Run view's empty state
also has Build everything and Test everything as buttons, and the palette names the scope in
each title (`Build orders`, `Build all 23`).

### Per view

| View | Key | Does |
|---|---|---|
| Home | `Enter` | Do the suggested next action. |
| any workspace view | `d` | Open Doctor (shown in the footer of Home and Projects). |
| Projects | `a` | Add project (a guided form). |
| any workspace view | `+` | Install a package: a form that asks which project (the one selected, `Left` `Right` change it), which packages and whether they are development dependencies, shows the `smith install -w <project> ...` line, then runs it in the Run view. The palette's and the detail's "Install a package…" open the same form. |
| any workspace view | `-` | Remove a package: the same form for `smith uninstall -w <project> ...`. |
| any workspace view | `i` | Install dependencies (asks to confirm the command). |
| Projects | `z` | Zoom the detail to the full width, and back. |
| Projects, Doctor | `[` `]` | Make the detail narrower or wider (kept until smith closes). |
| Projects, Run, Doctor | `Ctrl+U` `Ctrl+D` | Scroll the detail. |
| Run | `r` | Run the last command again. |
| Run | `f` | Follow the log, or pause it to read. |
| Run | `Up` `Down` `PgUp` `PgDn` | Pick a task, scroll its log. |
| Run | `Esc` | Show all output. |
| Doctor | `Enter` | Expand the check (narrow terminals), or unfold the detail (wide). |
| Doctor | `f` | Fix this check: the wizard asks, shows what changes, then applies. |
| Doctor | `F` | Fix with doctor: the wizard on all the fixes that need no decision. |
| Doctor | `g` | Group the checks by project, or list them flat. |
| Doctor | `p` | Show or hide the passed checks. |
| Doctor | `e` | Show every line of the problem, or fewer. |
| Doctor | `PgUp` `PgDn` | Scroll the detail. |
| Deps | `Space` | Mark or unmark a dependency. |
| Deps | `s` | Mark the patch and minor ones. |
| Deps | `a` | Mark all, or none. |
| Deps | `u`, `Enter` | Update the marked ones (`smith deps up`). |
| Deps | `v` | Bring drifting versions to one (`smith deps up --align`). |
| Release | `Enter` | Show the commits behind the selected bump. |
| Release | `p` | Preview the next release (`smith release --dry-run`). |
| Release | `n` | Print the release notes (`smith release --dry-run --notes`). |
| Release | `x` | Release… (`smith release`, after a confirmation). |
| Settings | `Up` `Down` | Move in the list of sections, or in its rows. |
| Settings | `Tab`, `Left` `Right` | Switch between the list of sections and its rows. |
| Settings | `Space`, `Enter` | Change the selected plugin or setting. |
| Settings | `t` | Review the selected external plugin and trust it. |
| Settings | `f` | Forget that an external plugin is trusted. |
| Settings | `u` | Reset the selected setting to its default. |
| Settings | `Esc` | Back to the list of sections, or cancel the editor. |
| Start menu | `n` | New workspace… (outside a workspace). |
| Start menu | `a` | Adopt… (opens the adopt wizard). |

A key is refused with a reason in the toast when the view cannot do it now (nothing to
update, no safe fix), and its hint leaves the footer.

### In the palette

`:`, `Ctrl+P` or `/` opens a palette: type to find a command; Up, Down, PgUp and PgDn choose,
Enter picks, Esc or `Ctrl+C` closes. There is one line per intent, in plain words, grouped
under Build, Projects, Dependencies, Release and Health (then Go to and Help), each with its
key at the right and the command it runs as a muted hint. What cannot work now is hidden
until you type; searching finds it with the reason. Nothing is typed as a raw command line:
a command that needs an answer opens its own form, such as Install a package (project
preselected, then the packages) or Add project. Commands that have no key are reached here:
Regenerate CI pipelines (asks to confirm), Info, Show what changed (`smith affected`),
Project graph, Manage plugins, Upgrade workspace (asks to confirm) and Remove project.
The palette and every box drawn over the screen stay inside the frame; behind a box the
view is emptied so no text shows at its edges, and a toast sits inside the content area.

### In a confirmation

Commands that change the workspace or publish (`smith ci generate`, `smith install`,
`smith upgrade`, `smith release`) first show their command line in a dialog. `y` confirms;
`n`, `Esc`, `Ctrl+C` or `q` cancels; `Enter` or `Space` presses the focused button; `Tab`,
`Left`, `Right` and `Shift+Tab` move between the buttons.

## Settings

The Settings tab (key `7`) is where you change things without leaving the interface. A list
of sections is on the left; the open section's rows and, under them, what the selected row is
and does are on the right. `Tab` (or `Left` and `Right`) moves between the two; a click does
the same, and a double-click does what `Enter` does. Everything it writes is what the
matching command writes: a test compares `smith.toml` byte for byte with the result of
`smith plugin disable`, `smith plugin trust --pin` and `smith config set`, and comments and
every other line stay as they are.

**Plugins.** The built-in language plugins, each on or off (`Space`; at least one must stay
on, as with `smith plugin disable`). Below them, the external plugins the workspace declares
in `smith.toml`, with whether each is trusted and pinned: `Enter` (or `t`) shows the review
`smith plugin trust` shows, which is what the plugin is, where it comes from, the files its
command runs with their sha256 and the environment it may see, and `y` trusts it, `p` trusts
it and writes its sha256 to `smith.toml` for CI, `Esc` leaves it untrusted; `f` forgets a
decision (`smith plugin untrust`). A trusted plugin runs with your privileges and is not
sandboxed. Then the plugins installed on the plugin search path, and a row to install a
plugin: until `smith plugin install` exists in your smith it shows the command as a hint, and
then it asks for the name or path here.

**Style.** The colour theme (`auto`, `dark`, `light`, `plain`) and the symbols (`auto`,
`unicode`, `ascii`), with a preview of the highlighted choice. `Enter` applies the choice at
once to the whole interface and keeps it for you, in `tui.json` in the smith home
(`SMITH_HOME`, or `~/.smith`), not in `smith.toml`: it is a personal preference. The terminal
interface reads it when it starts; the command line does not use it. The variables in
[Plain mode and environment](#plain-mode-and-environment) still win over it, and the section
says so when one is set. `auto` means what smith did before: the terminal's background and
`TERM` decide.

**Workspace.** Every setting [`smith config list`](config-list.md) knows, grouped by the
table of `smith.toml` it belongs to, with what it does and its default. A switch flips with
`Space`; a list with fixed choices (`ci.providers`) is ticked with `Space` and saved with
`Enter`; text, numbers and lists without fixed choices (separate the items with commas) are
typed. The value is checked against the schema before anything is written, and the reason
stays in the editor when it is refused. `u` removes the setting, so it has its default.
`plugins.enabled` is edited in Plugins, and stacks, tasks and projects are edited in the file.

**Recipes.** The `*.recipe.toml` files in the workspace, with their name and version and why
one is not valid; [`smith recipe check <file>`](recipe-check.md) shows what a recipe would use.

**Forms.** The guided forms (add a project, remove a project and the rest of the Start menu)
that end as a smith command line.

Home points at the tab (`New here? ...`) until you have opened it once; that is remembered in
`tui.json`.

## Wizards

### The doctor fix wizard

Opened by `f` or `F` in Doctor. It is the interactive form of
[`smith doctor --fix`](doctor.md) and asks the same questions.

1. **Questions.** A fix may ask something (install the missing tool, or remove the script?).
   `Up` `Down` (or `j` `k`, `Tab`) move, a digit picks that option, `Enter` or `Space`
   chooses, `b`, `Left` or `Backspace` goes back one question, `Esc`, `q` or `Ctrl+C` leaves.
2. **Review.** What will change: files as diffs, the commands it will run, the projects it
   covers. `Enter` or `y` applies; `n`, `Esc` or `q` leaves; `Up` `Down` `PgUp` `PgDn` `Home`
   `End` scroll.
3. **Progress.** Step by step with a bar. `Esc` or `Ctrl+C` stops the step.
4. **Failure.** The failing step's output. `r` or `Enter` retries, `y` copies the failure
   text, `Esc`, `q`, `c` or `Ctrl+C` closes.

### The adopt wizard

Opened with `a` from the Start menu, or by Enter when the folder holds files but no
workspace. It combines [`smith adopt`](adopt.md) and [`smith convert`](convert.md) and uses
the fix wizard's keys and drawing. The stages:

1. **Reading** the repository.
2. **What was found**: the kind of repository (Nx, Turborepo, mnci, or plain), the projects
   (the first twelve), the old tooling in use, and blockers (`✖`) and warnings (`!`). Nothing
   has changed yet.
3. **How far to go**: conservative (keep the build and lint setup) or full (make smith the
   only tool; Nx and Turborepo only, like `--level full`).
4. For the full level: a **trial conversion on a scratch copy** to see what would be left, then
   **the four choices** (lint, test runner, build, CI), as in `smith convert --lint --test
   --build --ci`.
5. **The whole plan**, for confirmation: `Enter` converts. Nothing is written before.
6. **Converting**, then, for the full level, a question per **group of removals** (`We are
   going to delete N packages and M files. ... ok?`): `y` deletes them, `n` keeps them,
   `s` keeps these and skips the groups after.
7. **Finishing**: the install and the checks (build, test, lint, doctor).
8. **Done**: what was converted, the old tooling removed, what was left as you chose, files
   that look like old tooling but stay because something still uses them, and notes. `Enter`
   closes; run `smith` again to open the workspace.

While a job runs, `Ctrl+C` or `Esc` stops it. A failed step shows its output with Retry.
Everything is in the working tree and git undoes it.

### Guided forms

Entries of the Start menu and of the palette that need arguments (create a workspace, add a
project, manage plugins, remove a project, and forms generated from the command catalogue)
ask one question at a time: a choice (`Up` `Down`, `Enter`, `Esc` back one step), several
choices (`Space` toggles, `Enter` confirms) or free text with a placeholder. The answers
build the command line, which the shell runs and shows. A command that needs a terminal of
its own (it asks questions) is handed back to smith to run after the shell closes; the shell
prints `→ smith <command line>`.

## The mouse

The mouse is on by default: click a tab, a row, a footer hint, a detail action or a dialog
button; double-click a project, a guide option or a row of the Settings tab to open or change it; use the wheel to scroll. The wheel scrolls
a list and never moves the selection, and a thin scrollbar on the right edge shows where the visible part sits in a list or log that is longer than its pane; a click selects a row without moving the list. In Deps, click a row's box to
mark it, Shift+click to mark the rows from the last one to it (Shift+Up and Shift+Down do the same from the keyboard),
and click the box in the heading to mark all or none. Only
cell-motion reports are used, so Shift+drag still selects text in most terminals, and `m`
turns the mouse off.

## Plain mode and environment

| Variable | Effect |
|---|---|
| `CI` | Any value stops a bare `smith` from opening the shell. |
| `SMITH_SCREEN_READER` | Anything but empty, `0`, `false`, `no` or `off` draws no full-screen shell, no colour and no mouse: plain numbered questions, then the command, which smith runs in the terminal. |
| `SMITH_THEME` | `light`, `dark` or `plain`. `plain` also selects plain mode. Without it the choice saved in the Settings tab decides, and without that the terminal's background. |
| `NO_COLOR` | Any value gives the plain colours. |
| `SMITH_ASCII` | Any value (or `TERM=dumb` or `TERM=linux`) uses ASCII symbols instead of Unicode. Without it the symbols saved in the Settings tab decide. |
| `SMITH_HOME` | Where the smith home is (`~/.smith` by default): installed plugins, trust decisions and `tui.json`, the theme and symbols chosen in the Settings tab. |
| `SMITH_NONINTERACTIVE` | Read by the runner that starts the shell's child commands (its exact effect was not verified). |

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
