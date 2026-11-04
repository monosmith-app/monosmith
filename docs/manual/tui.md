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
| Home | `1` | The health of the workspace in one headline, the facts behind it (checks, affected projects, the last run, outdated dependencies) and the one thing to do next, with its key. "Where to go" starts with two buttons, "Add a project" (`a`) and "Compose a stack" (`A`); click one or press its key. |
| Projects | `2` | A "+ New project" button (key `a`, also with projects in the workspace) that opens the Add wizard, then every project, ordered by kind and then name, with its language; the detail pane shows its properties, its derived files (when the workspace derives files) and the actions that apply to it. |
| Run | `3` | The tasks of the current or last run: queued, running, cached, passed, failed or skipped, with durations, a progress bar and each task's log. |
| Doctor | `4` | The checks of [`smith doctor`](doctor.md), grouped by project or flat, each with its problem and remedy; fixes open the fix wizard. |
| Deps | `5` | Dependencies with newer releases, grouped patch, minor and major (columns Package, Now, Latest, Level, Kind, Projects), and the one-version policy's verdict, with the `smith deps up` command line for a selection. |
| Release | `6` | What the next release would do (the plan of [`smith release --dry-run`](release.md)), the commits behind each bump. On a release channel branch it says so first: ``Release channel `next`: prerelease versions…``, and the confirmation of the release repeats it. |
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
| `1` to `7` | Show Home, Projects, Run, Doctor, Deps, Release, Settings. |
| `Up` `Down`, `j` `k` | Move in a list. |
| `Enter` | Open or run the selection. |
| `Esc` | Back, close, clear. |
| `?`, `F1` | The shortcuts overlay (Up, Down, PgUp, PgDn, Home, End scroll it; Esc, `?`, `F1`, `q` or Enter close it). |
| `:`, `Ctrl+P`, `/` | The command palette. |
| `q` | Quit (not while a run is going: `Ctrl+C` stops it first). |
| `Ctrl+C` | Stop the run in progress; pressed again, quit. |
| `m` | Turn the mouse on or off. Off lets the terminal select text; Shift+drag selects with it on. |
| `y` | Copy the focused item (remedy, command, project path or package) to the clipboard, through the terminal (OSC 52), so it also works over SSH and in tmux panes that allow it. Works in Projects, Run, Doctor, Deps. |
| `r` | Reload this view: read the files and the checks again, run nothing (Home, Doctor, Release, Deps). The footer says `r reload view`. On the Run view `r` is `run again`: it runs the last command line again. |

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
| Home, Projects, Doctor, Deps, Release | `r` | Reload this view: read the files and the checks again. It runs nothing. The footer says `reload view`; on the Run view the same key is `run again`. |
| Projects | `a` | Add project (a guided form). |
| any workspace view | `+` | Install a package: a form that asks which project (the one selected, `Left` `Right` change it), which packages and whether they are development dependencies, shows the `smith install -w <project> ...` line, then runs it in the Run view. The palette's and the detail's "Install a package…" open the same form. |
| any workspace view | `-` | Remove a package: the same form for `smith uninstall -w <project> ...`. After the form, a confirmation shows the exact command line and what it changes, because a removal edits the manifest and the lock file. |
| any workspace view | `i` | Install everything every project declares, as after a fresh clone (asks to confirm the command). To add one package use `+`. |
| Projects | `z` | Zoom the detail to the full width, and back. |
| Projects, Doctor | `[` `]` | Make the detail narrower or wider (kept until smith closes). |
| Projects, Run, Doctor | `Ctrl+U` `Ctrl+D` | Scroll the detail. |
| Run | `r` | Run the last command again, as it was. A command that changes files asks first. A command that makes or removes something (`add`, `new`, `remove`, `adopt`, `convert`) is not run again: the key says why. |
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
| Deps | `i` | Install a package (`smith install -w <project> <package>`). It asks the names (a name or name@version, several with spaces), then the projects (space ticks; the projects that declare the package under the cursor start ticked, and nothing runs before you confirm), then the kind: production or development (`-D`). Each project is listed with its language and folder, so for a project with several parts, such as a Tauri app (a Rust part and a web part), you pick the part that gets the package. The confirmation opens at once and shows the plan: the latest release of each name and the section it goes into. Global installs, peer and optional dependencies are not offered: `smith install` has no way to do them. |
| Deps | `x` | Remove the package under the cursor (`smith uninstall -w <project> <package>`). A package that one project declares goes straight to the confirmation with that project as `-w`; the workspace root, shown as `(root)`, is `-w .`. When several projects declare it, a multiple choice asks from which ones (space toggles); smith never picks for you. A development dependency is removed like any other. |
| Deps | `v` | Bring drifting versions to one (`smith deps up --align`). |
| Release | `Enter` | Show the commits behind the selected bump. |
| Release | `p` | Preview the next release (`smith release --dry-run`). |
| Release | `n` | Print the release notes (`smith release --dry-run --notes`). |
| Release | `x` | Release… (`smith release`: first hold back or bump packages, then a confirmation that says it cannot be undone). |
| Settings | `Up` `Down` | Move in the list of sections, or in its rows. |
| Settings | `Tab`, `Left` `Right` | Switch between the list of sections and its rows. |
| Settings | `Space`, `Enter` | Change the selected plugin or setting. |
| Settings | `t` | Review the selected external plugin and trust it. |
| Settings | `f` | Forget that an external plugin is trusted. |
| Settings | `u` | Reset the selected setting to its default. |
| Settings | `Esc` | Back to the list of sections, or cancel the editor. |
| Start menu | `n` | New workspace… (outside a workspace). |
| Start menu | `a` | Bring into MonoSmith… (opens the wizard). |

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
Project graph, Manage plugins, Upgrade workspace (asks to confirm) and Remove project. The
Health section also holds the 2.0 commands, described in
[Undo, report, why and derived files](#undo-report-why-and-derived-files): Undo the last
change, Open the report, Write the agent file, Check the derived files, Update the derived
files and Why is this here?
The palette and every box drawn over the screen stay inside the frame; behind a box the
view is emptied so no text shows at its edges, and a toast sits inside the content area.

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

The dialog opens at once, also when smith must work out the plan first (the packages and versions
of `u` in Deps, the files of a swap, an add or an undo, the packages of a release). While it
waits, the dialog shows a spinner and a plain sentence ("Working out what will change…"), the
seconds after the second one, and the command that will run. `[ Run ]` is faint and does nothing:
`Enter`, `Space`, `y` and a click on it are ignored, so you cannot run a plan you have not seen.
`Esc`, `n`, `q`, `Ctrl+C` or a click outside cancels the wait: smith stops the command that reads
the plan, and a plan that arrives late opens nothing. When the plan is ready, it fills the dialog
and the usual buttons work. The wait shows no steps or percentage, because the command reports none.

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
  says how an app uses it: press `+` on the app and pick it, or `smith install -w <app> <library>`.
- **An add that could not be checked because a tool is missing**: the project exists on disk, and the dialog
  says so (`api-cs was created in apps/api-cs (12 files)`), then why the first build and test could not run
  (`the .NET SDK is not installed`). Choices: **Open the install page** (copies the link and opens it),
  **Install it with smith** (the tools dialog, which asks how before it runs anything, then returns here),
  **Check again** (runs the first check of the project again, not the add) and **Continue without checking**
  (Doctor keeps reminding you). Esc closes the dialog and changes nothing.

### Undo, report, why and derived files

These are in the palette (`:`), under Health. None has a key of its own.

| Palette entry | What it does |
|---|---|
| **Undo the last change…** | Runs `smith undo --dry-run` first and shows what it would revert in a dialog. `y` or `Enter` then runs `smith undo --yes`, which reverts it with a new commit; `Esc` leaves everything as it is. When smith has nothing to undo, a notice shows its reason and the fix (for example `commit_operations = true` under `[workspace]` in `smith.toml`); `Enter` closes it. |
| **Open the report** | Runs `smith report -o <file>` with a file in the temporary folder named after the workspace (`smith-report-<workspace>.html`), opens it in your browser and tells you the path in a toast, so you can open it again. |
| **Write the agent file…** | `smith agent file` after a confirmation: `AGENTS.md`, and a short `CLAUDE.md` that imports it. |
| **Check the derived files** | `smith derive --check` in the Run view: writes nothing, fails when a derived file is out of date or was edited. |
| **Update the derived files…** | `smith derive` after a confirmation: writes the files that follow from the project graph. A file you edited is never overwritten. |
| **Preview the derived files** | `smith derive --dry-run` in the Run view: the plan of the update, nothing is written. |
| **Why is this here?…** | A form that asks for a file, a folder or a setting and runs `smith why <it>`: who owns a file and how to change it, or why a setting has its value. |
| **Divergences from the defaults** | `smith divergences` in the Run view: every override, ejected file and drift, each with the reason and how to go back. |
| **Try a command in a copy…** | A form that asks for a smith command (`add package billing`) and runs `smith try` on it: the command runs in a temporary copy, doctor checks the result, and the Run view says what changed. Your files stay as they are. |
| **Write the launcher…** | `smith bootstrap` after a confirmation: writes `tools/smith/smith` and `tools/smith/smith.cmd`, so a laptop and CI run the same smith. |
| **Swap a choice…** | Under Projects in the palette. Pick the project, the decision (test, lint, style) and the answer. smith shows the plan of `smith swap --dry-run` first, with the unticked box "Commit this change so it can be undone"; Run does the swap. |

### Environments, cache, pipelines and the rest of the palette

Every command of smith that is not a dev server, `serve`, `mcp`, `lsp` or `agent` has an entry in
the palette (`:`). A command that changes files shows its plan first, as every other change does.

| Palette entry | Section | What it does |
|---|---|---|
| **Environments of the web apps** | Projects | `smith environments list` in the Run view. |
| **Add environments to a web app…** | Projects | Pick the web app (the selected project first), type the names. The plan of `smith environments add --dry-run` lists the files it writes. |
| **Remove environments from a web app…** | Projects | Same form. The plan says that the `.env.<name>` file stays and that the default environment cannot go. |
| **Make the icons of a Tauri app…** | Projects | Asks for the square PNG and, when more than one app has a `src-tauri`, which app. The plan lists the icons it overwrites. |
| **Add a ready-made pipeline…** | Release | Pick e2e, package-zip, deploy-pages or deploy-azure-function; a deploy pipeline asks for its app. The plan lists the files. |
| **Set the Docker Hub descriptions…** | Release | Asks to name the repositories only (the default, nothing is sent) or to set the descriptions. |
| **Cache: status**, **check for damage** | Health | `smith cache status` and `smith cache verify` in the Run view. |
| **Cache: remove what is broken…**, **trim to its limit…**, **delete it all…** | Health | `smith cache verify --fix`, `prune` (with its plan) and `clear`, each after a confirmation that says the next run is slower. |
| **Copies between projects**, **Tools the workspace needs** | Health | `smith copies` and `smith tools` in the Run view. |
| **Why would a task run again?…** | Health | Asks `project:task` and runs `smith why-miss`. |
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
cannot answer for the person (the trigger of a function app) is asked either way.

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

The Settings tab (key `7`) is where you change things without leaving the interface. A list
of sections is on the left; the open section's rows and, under them, what the selected row is
and does are on the right. `Tab` (or `Left` and `Right`) moves between the two; a click does
the same, and a double-click does what `Enter` does. Everything it writes is what the
matching command writes: a test compares `smith.toml` byte for byte with the result of
`smith plugin disable`, `smith plugin trust --pin` and `smith config set`, and comments and
every other line stay as they are. A setting whose default has a recorded reason shows it under
the setting as "Why:", the same text as `smith why <setting>`.

**Plugins.** The built-in language plugins, each on or off (`Space`; at least one must stay
on, as with `smith plugin disable`). Below them, the external plugins the workspace declares
in `smith.toml`, with whether each is trusted and pinned: `Enter` (or `t`) shows the review
`smith plugin trust` shows, which is what the plugin is, where it comes from, the files its
command runs with their sha256 and the environment it may see, and `y` trusts it, `p` trusts
it and writes its sha256 to `smith.toml` for CI, `Esc` leaves it untrusted; `f` forgets a
decision (`smith plugin untrust`). A trusted plugin runs with your privileges and is not
sandboxed. Then the plugins installed on the plugin search path, and a row to install a
plugin. `Enter` on that row asks where the plugin comes from (a directory, an archive, an https
URL or `github:owner/repo@tag`; an archive or URL needs its sha256 typed after the source) and
shows the command line as you type, `smith plugin install <source> [--sha256 <digest>]`. `Enter`
again shows it in the confirmation dialog, and `y` runs it in the Run view, so its output is the
command's own. The plugin is checked and never started; trusting it is the review above. The tab
reads the plugins again when the command ends. The row above it, **Find a plugin in the
marketplace…**, asks for words (or nothing, to list every plugin), runs
`smith plugin search <words> --json` and shows the plugins found as a list you can narrow by
typing, with the newest version, what it does and its roles. `Enter` on one shows the
`smith plugin install <name>` confirmation, as above. An empty result says to search with fewer
words, and a failed search shows the command's error in a notice.

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
table of `smith.toml` it belongs to, with what it does and its default. A switch flips with
`Space`; a list with fixed choices (`ci.providers`) is ticked with `Space` and saved with
`Enter`; text, numbers and lists without fixed choices (separate the items with commas) are
typed. The value is checked against the schema before anything is written, and the reason
stays in the editor when it is refused. `u` removes the setting, so it has its default.
`plugins.enabled` is edited in Plugins, and stacks, tasks and projects are edited in the file.

**Recipes.** The `*.recipe.toml` files in the workspace, with their name and version and why
one is not valid; [`smith recipe check <file>`](recipe-check.md) shows what a recipe would use.

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

## Wizards

### The doctor fix wizard

Opened by `f` or `F` in Doctor, or by **Fix smith.toml** on the Start menu when `smith.toml` uses a renamed key (the same fix as `smith doctor --fix` makes; after it is applied the shell starts again inside the workspace). It is the interactive form of
[`smith doctor --fix`](doctor.md) and asks the same questions.

1. **Questions.** A fix may ask something (install the missing tool, or remove the script?).
   `Up` `Down` (or `j` `k`, `Tab`) move, a digit picks that option, `Enter` or `Space`
   chooses, `b`, `Left` or `Backspace` goes back one question, `Esc`, `q` or `Ctrl+C` leaves.
2. **Review.** What will change: one line for each file (what happens and how many lines), the
   commands it will run, the projects it covers. `d` shows the diffs and hides them again.
   `Enter` or `y` applies; `n`, `Esc` or `q` leaves; `Up` `Down` `PgUp` `PgDn` `Home` `End`
   scroll. In the adopt wizard the review of a conversion has the same unticked box, "Commit this
   change so it can be undone": `c` or a click ticks it.
3. **Progress.** A spinner and the fix title, then "step n of m" with the step's name (for
   **Fix with doctor**, the steps of all the fixes together). After 2 seconds the line adds the
   time it has run. Under it, one line shows what smith writes ("Writing package.json") and then
   the last line the running command printed. Other keys do nothing while it runs, so a second
   fix cannot start. `Esc` or `Ctrl+C` stops the step.
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
   (`✖`) and warnings (`!`). Nothing has changed yet. When the packages are not installed, or
   are not the ones the repository lists, the button is **Install the packages and
   continue**: smith runs `npm ci` (or `npm install` without a lockfile), reads the
   repository again and goes on. git cannot take an install back, and the screen says so.
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
choices become `--with a,b`. A workspace with three or more projects is also asked whether the
new one should do what the others do (the default), follow one project you pick from a list
(`--like <project>`) or take the stack's own defaults (`--fresh`). Then the shell runs `smith add ... --dry-run` and shows its plan in a dialog:
the files it would write and, when it applies, which sibling project it follows
(`follows web: same test runner`). `y` runs the real add, `Esc` writes nothing. A wrong
combination (a choice for another language, one that needs a choice you left out) shows what
smith said, with its fix, in a notice instead of the plan. A command that needs a terminal of
its own (it asks questions) is handed back to smith to run after the shell closes; the shell
prints `→ smith <command line>`.

## The mouse

The mouse is on by default: click a tab, a row, a footer hint, a detail action or a dialog
button; double-click a project, a guide option or a row of the Settings tab to open or change it; use the wheel to scroll. The wheel scrolls
a list and never moves the selection, and a thin scrollbar on the right edge shows where the visible part sits in a list or log that is longer than its pane; a click selects a row without moving the list. The same rule holds in every list of a form
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
