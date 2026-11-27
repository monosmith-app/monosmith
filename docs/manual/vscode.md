# The MonoSmith VS Code extension

The MonoSmith extension runs the tasks of a smith workspace from VS Code. It adds an
activity-bar container with the seven views of the terminal interface (Home, Projects, Runs, Health, Dependencies, Release, Settings), a list of every command with its smith command line,
commands and keyboard shortcuts to run tasks, a `smith` task type with problem matchers, a
status-bar item and a walkthrough. It does not carry smith: it starts the `smith` your team
installed and talks to it with [`smith serve`](serve.md), so the editor, the terminal and
CI give the same result.

This page documents every command, view, setting, task field and key that
`apps/vscode/package.json` of this version contributes. The extension's own README
(`apps/vscode/README.md`) is the short introduction with screenshots. Installation from the
Marketplace is not described here.

## Requirements

| Item | Value |
|---|---|
| VS Code | `^1.140.0` |
| Activation | When the workspace contains a `smith.toml`, a marker of another monorepo tool (`nx.json`, `turbo.json`, `lerna.json`, `pnpm-workspace.yaml`, `rush.json`, `go.work`, `package.json`) or any file at its root (`workspaceContains:*`); and when the MonoSmith view is opened, which is how an empty folder gets its offer. |
| smith | A smith whose `serve` protocol version the extension speaks (version 1; a newer one is worked with, an older one cannot be served). |

## Which smith runs

In order: the `monosmith.path` setting; the workspace's own install (`node_modules/.bin/smith`,
for example after `npm i -D monosmith`, or `.smith/bin/smith`), **only in a trusted
workspace**; `smith` on the `PATH`. The status-bar tooltip says which one runs. When the
workspace's smith is not the version a terminal gets from the `PATH`, the extension warns once
per session and offers to open the settings or copy the update command (`npm i -D monosmith`).

In an untrusted workspace (Restricted Mode) the extension does not run the workspace's own
smith and does not run tasks, because they execute code from the workspace. It uses the
`monosmith.path` user setting or the `PATH` to show the projects, says why in the view, and
starts again by itself when you trust the workspace. The workspace value of `monosmith.path`
is ignored while the workspace is untrusted.

## Views

All views are in the **MonoSmith** container of the activity bar, with the names and the order
of the terminal interface: **Home, Projects, Runs, Health, Dependencies, Release, Settings**.
Home and Projects show open. The other views show collapsed, and only when smith is ready.
Every row that does something names the smith command line it stands for, and every command
that changes the workspace asks first and shows that command.

### Home

What needs you now, as the Home of the terminal shows it. The four boxes come first: **Changed**
(the projects your changes affect), **Last run**, **Health** and **Release**. A click on a box
opens its view. **Needs you now** follows: a failed run (See why, Run it again), checks that fail
or warn (Fix, Open Health), changed projects (Test what changed, `smith run lint test
--affected`), a release that is ready (Open Release) and the updates (Open Deps). A click on an
item runs its first button. The buttons are the rows under the item, each with its smith
command. When nothing asks for attention, the first item is **All good** with **Build all**.
**Start something** ends the view: Build all, Lint all, Run a task…, Add a project, Compose a
stack and Settings.

Home reads nothing of its own. Doctor and the release plan start in the background the first
time Home shows, as the terminal starts them with the shell. The updates stay "not looked at
yet" until the Dependencies view opens, because that view asks the registries.

### Projects

Apps, libs and packages with their language. A dot marks a project the changes affect (left
out when every project is affected; the title then says `all N affected`). The icon of a
project is the mark of its last run: the worst state of its tasks (failed, running, passed).
Before a run, the icon is the icon of its kind. A multi-root workspace gets one section per
folder that has a `smith.toml`.

Under a project are the four sections of the terminal's project detail:

| Section | Holds |
|---|---|
| Tasks (open) | the tasks, each with an icon for how it ended, how long it took, and a tooltip with the last result and whether it came from the cache |
| Dependencies | **Uses** and **Used by**, then **Install a package** (`smith install -p <project> <package>`) and **Unlink a library** |
| Runs | the runs of this session that ran a task of the project; a click opens the run in Runs |
| Settings | **Change a project setting** (`smith config set -p <project> <setting> <value>`) |

When smith is not usable the view says why instead of staying empty, with buttons:

| State | Message | Buttons |
|---|---|---|
| starting | `Starting smith…` | Show log |
| missing | `smith was not found.` | Copy install command, Open settings, Retry, Show log |
| failed | `smith could not start, or it stopped.` | Restart smith, Show log, Open settings |
| untrusted | `This workspace is not trusted.` | Manage Workspace Trust, Open settings, Retry |
| empty | `This folder is empty.` | Create a new MonoSmith workspace (`smith new --here` in a terminal), Learn more |
| unrecognised | `This folder is not a MonoSmith workspace yet.` | Bring this folder into MonoSmith (`smith adopt` in a terminal), See what would change, Learn more |
| nofolder | `No folder is open.` | Open Folder |

The status bar menu offers the same entries in those states. Nothing is written by the click:
smith asks its questions in the terminal, lists what it will create or change, and `smith new --here`
refuses a folder that holds files. A folder counts as empty when it holds only `.git`, `.vscode`,
`.DS_Store`, `Thumbs.db` or `desktop.ini`; the extension lists the top of the folder and reads no file.

Title-bar buttons (when ready): Filter Projects…, Show Affected Only / Show All Projects,
Run Affected, Re-run Last (not while running), Cancel Run (while running), Refresh. The
`…` menu: Run Task, Run Task in All Projects, Clear Filter (when filtered), Run Doctor, Clear
Problems, Restart smith, Show Output, Open Settings, Open Getting Started.

Row actions: on a task, **Run Task** (inline and context menu) and, when it has run, **Show
Task Output**; on a project, group or folder, **Run All Tasks**; on a project, **Explain
Project** and **Show Project Graph**. Enter on a row runs it (`Run Focused Task`).

The filter narrows the view as you type, and it is remembered per workspace. It reads the
words of the terminal's filter: `kind:<prefix>` (`kind:app`), `lang:<prefix>` (`lang:go`),
`changed` (the projects the changes affect), and any part of a name, a kind, a language or a
task. Every word must match.

### Runs

The runs you start from the editor open here, as a run started anywhere opens the Runs view of
the terminal. The editor keeps the keyboard. The newest run is open: its state and its tasks
(`Failed 4 tasks`), how far it is (`3/4 75%`) and how long it took. The command line is in the
title and the tooltip. Under it are **All output** (the MonoSmith output channel) and the tasks
grouped by project, each with its mark, its time, `hit` for a cached task, or the reason it did
not pass. A click on a task opens its log as a document. A progress bar runs on the view while
the run goes. **Runs this session** holds the older runs (smith keeps 20, in memory).

Title-bar buttons: **Show Errors Only** / **Show All Tasks** (the terminal's `e`), **Re-run
Last**, **Cancel Run** (while running) and **Show Output**.

### Health

The checks of `smith doctor`, grouped by state as the terminal groups them: **Failed**, then
**Warnings**, each heading with its count, then **Passed**, closed. Each check says its project
(or `workspace`) and its file, and `fix` when smith can fix it. Under it are the problem, the
remedy and the fix.

The fix has the three steps of the terminal's inline fix (`1 Review ── 2 Apply ── 3 Check
again`). A click on **Fix: …** asks the questions of the fix. Then smith's review opens beside the
code, and the question asks whether to apply it. After the fix, doctor runs again and the
message says whether the check passes now, that nothing was committed, and offers **Check
again**. The **The same as** row shows the command for a terminal:
`smith doctor --fix --only '<fix>'`.

**Fix Health Problems…** (shown when a fix exists) shows the changes smith proposes and applies
them only after you agree. **Refresh Health** reads again.

When `[derive] enabled = true` is set in `smith.toml`, a **Derived files** row follows the
checks. It says how many files are up to date, or lists each file that needs action with its
state (`drift`, `diverged`, `stale`, `missing`, `unrecorded` or `broken`), the renderer and the
fix. The same files get a warning in the Problems panel (when `monosmith.output.problems` is on).
The editor reads them from `smith derive --check`, which has no JSON form yet, so an older or
newer smith that words the lines differently shows the failure as one row instead. Without
`[derive] enabled` the row is left out.

Right-click a derived file in the Health view, in the Explorer or in an editor:

| Action | Runs | Does |
|---|---|---|
| **Explain Derived File** | `smith derive --explain <file>` | A page with the owner, the renderer, the state and the diff. |
| **Apply Derived File** | `smith derive`, or `smith derive --replace <file>` | Writes a stale or missing file. For a drifted or diverged file, overwrites it with the derived content. |
| **Adopt Edit of Derived File** | `smith derive --adopt <file>` | Keeps your edit as the new baseline (only for a file smith owns by keys). |
| **Eject Derived File (Own It)** | `smith derive --eject <file>` | The file is yours; smith stops touching it. |

Each action names its effect and asks first. Commit the file together with `smith.lock`.

### Dependencies

What has a newer release, in the two groups of the terminal: **Patch and minor releases**, safe
to take together, and **Major releases**, to take one at a time. The title says the headline
(`↑ 3 with newer releases · 1 patch · 1 minor · 1 major`). A row says the versions, the level and
where the package is used. Under it is the package's detail: the versions, a warning for a
major release, **Used by** (each file that declares it), **Declared as**, **Registry** (a click
opens the page), **The same as** (`smith deps up <package>`, with `--major` for a major release)
and **Update <package>**. The tools smith pins come last.

**Update Dependency…** asks first and shows the command, then runs `smith deps up <names>` in a
terminal. Select several rows (Ctrl or Shift and a click) to update them in one command, as the
terminal's marked packages. **Update All Safe…** does the same for every patch and minor release.
**Repair Dependencies…** opens `smith deps repair` in a terminal after a yes; smith asks there
what to repair. **Run smith deps up…** opens the picker of `smith deps up`. The editor never
rewrites a manifest itself: smith has no method for it in the editor protocol.

### Release

What `smith release` would publish, as the terminal's Release view shows it. The title says
`✓ 2 packages to release · 1 minor · 1 patch`. Each package has the version it moves to and,
under it, why it releases, its changes (the commits), its tag and **The same as**
(`smith release --dry-run --notes`). Then come the **Checks before release** (Health, the tests of
this session, and the registry sign-in, which smith checks when you release), the note that
nothing is published until you confirm, and the actions with their commands: **Release now**
(`smith release`, which shows the plan and asks), **Dry run** and **Release notes** (each opens a
page). **Refresh Release Preview** reads again.

On a release channel (see [`smith release`](release.md)), the first row reads **Release channel
`next`** and carries the notice of smith: the versions are prereleases, published under the
channel tag, and the stable version stays the latest. The title of the view repeats it
(`2 to release · channel next`), so nobody mistakes a prerelease line for the stable one.

### Settings

The sections of the terminal's Settings view. **Workspace** is open: every setting of
`smith.toml` that `smith config list` knows, each with its value (`● on`, `○ off`, `○ not decided`
for a switch with no value, `risky ▾` for a fixed set, or the default), its type, and in the
tooltip what it does, why the default is what it is, and the command it stands for
(`smith config set <setting> <value>`). A click opens **Change a Setting of smith.toml…** on that
setting: the value is asked, and the confirmation shows the command before smith writes the
file, as the command does. A set setting has **Reset to Default…** (`smith config unset`).

**Plugins** lists the plugins; a click asks On or Off, each with its command, then the
confirmation. **Recipes** lists the recipes of the workspace, with what is wrong with a file that
is not a recipe; a click opens the file. **Architecture**, **Assistants**, **CI** and **Forms**
open the matching guided commands. Per-project settings are in Projects, under a project, in
Settings. The colour theme of the terminal (its Style section) has no place here: VS Code uses
its own theme.

### Explain and graph pages

**Explain Project** opens a page with how each task of the project resolved and where each
field came from ([`smith explain`](explain.md)). **Show Project Graph** opens what the project
uses and what uses it ([`smith graph`](graph.md)).

### Why pages

**Why…** answers "who owns this, and why is it so?" in a page:

- On a **file** (Explorer or editor context menu, or the palette with a file open): `smith why
  <file> --json`. It names the owner (smith or you), what produced the file and how to change
  it. For a file smith does not know it says so.
- On a **task** row of the Projects view: `smith why-miss <project>:<task> --json`. It says
  whether the task would come from the cache or run again, and which inputs changed since the
  last run.
- On a **project** row: the explain page above.

From the palette, **Why…** asks what you want to know about.

### Add a project

**Add Project…** (the `+` button of the Projects view, the palette and the status-bar menu)
asks four things with quick picks: the kind (`app`, `lib` or `package`), the language, how the
project is built (the default of the language; a stack from `smith stacks`; or **Pick
choices…**, which asks one question per axis such as framework, test and style, from `smith
choices`), and the name (kebab-case). Then it runs `smith add --dry-run` and shows the plan in a
dialog; **Open the plan** puts it in an editor. Nothing is written until you choose **Add
project**. When the plan says the project follows its sibling projects, **Add with the stack
defaults** adds it with `--fresh` instead. The add runs in the background with a progress
notice (smith installs and tests the new project, which can take minutes) and its output goes to
the MonoSmith output channel. smith leaves the files uncommitted, or commits them when
`commit_operations = true` is set under `[workspace]`.

If the workspace uses a recipe (`[workspace] recipe`), the recipe asks its own questions in a
terminal. The command does not answer them: use `smith add` in a terminal, or pass `--choice`
there.

### Undo, report and the agent file

- **Undo the Last Change…** runs `smith undo --dry-run`, shows what it would revert and asks.
  Only then does it run `smith undo --yes`, which reverts with `git revert`: a new commit and
  your history stays. A notice offers it after **Add Project…** and after **Convert This
  Folder to MonoSmith** (the offer after a conversion needs the shell integration of the
  terminal). smith refuses when a later commit or an uncommitted change touches the same
  files, and when `commit_operations` is not set; the message says how to fix it.
- **Open Report** runs `smith report -o <temp file>` and opens the page in your browser. The
  report is not written into the workspace.
- **Write AGENTS.md for AI agents** runs `smith agent file --check`, names what is missing or
  old, and writes `AGENTS.md` and the `CLAUDE.md` region only after you agree. Text you wrote
  outside the markers stays. The notice after **Install the MCP server and agent** offers it,
  and it is in the Projects view menu and the palette.

## Commands

All are in the Command Palette under the category **MonoSmith** unless the table says they are
hidden there.

| Command (id) | Title | Does |
|---|---|---|
| `monosmith.refresh` | Refresh | Read the projects and what is affected again. |
| `monosmith.runTask` | Run Task | Ask for a project, then a task or all of them, and run it. |
| `monosmith.runSelected` | Run Focused Task | Run the focused row of the Projects view (Enter). Hidden from the palette. |
| `monosmith.runAllTasks` | Run All Tasks | All the tasks of the selected project, group or folder. |
| `monosmith.runAffected` | Run Affected | One task in the projects the changes affect. |
| `monosmith.runAll` | Run Task in All Projects | One task in every project that has it. |
| `monosmith.rerunLast` | Re-run Last | Repeat the last run (needs a previous run, not while running). |
| `monosmith.cancelRun` | Cancel Run | Stop the run and its processes. |
| `monosmith.doctor` | Run Doctor | Run `smith doctor`. |
| `monosmith.restart` | Restart smith | Stop and start the smith process. |
| `monosmith.showMenu` | Show Menu | The status-bar menu. |
| `monosmith.showOutput` | Show Output | Open the MonoSmith output channel. |
| `monosmith.openSettings` | Open Settings | Open the settings filtered to `monosmith`. |
| `monosmith.copyInstallCommand` | Copy Install Command | Copy `npm i -D monosmith`. |
| `monosmith.filter` | Filter Projects… | Ask for text to filter the Projects view. |
| `monosmith.clearFilter` | Clear Filter | Remove the text filter. |
| `monosmith.affectedOnly.on` | Show Affected Only | Keep only affected projects. |
| `monosmith.affectedOnly.off` | Show All Projects | Show every project again. |
| `monosmith.showTaskOutput` | Show Task Output | Open the last log of the selected task as a document. |
| `monosmith.clearProblems` | Clear Problems | Clear the entries the extension put in the Problems panel. |
| `monosmith.explainProject` | Explain Project | The explain page. |
| `monosmith.showGraph` | Show Project Graph | The graph page. |
| `monosmith.health.refresh` | Refresh Health | |
| `monosmith.health.fix` | Fix Health Problems… | |
| `monosmith.dependencies.refresh` | Refresh Dependencies | |
| `monosmith.dependencies.up` | Run smith deps up… | |
| `monosmith.dependencies.update` | Update Dependency… | A row action; hidden from the palette. |
| `monosmith.release.refresh` | Refresh Release Preview | |
| `monosmith.addProject` | Add Project… | Quick picks, the plan of `smith add --dry-run`, then the add. |
| `monosmith.undo` | Undo the Last Change… | `smith undo --dry-run`, then `smith undo --yes` after you agree. |
| `monosmith.openReport` | Open Report | Write `smith report` to a temp file and open it in the browser. |
| `monosmith.why` | Why… | Who owns a file, why a task runs again, how a project resolved. |
| `monosmith.agentFile` | Write AGENTS.md for AI agents | `smith agent file`, after you agree. |
| `monosmith.derive.explain` | Explain Derived File | Hidden from the palette unless the open file needs action. |
| `monosmith.derive.apply` | Apply Derived File | Same. |
| `monosmith.derive.adopt` | Adopt Edit of Derived File | Same. |
| `monosmith.derive.eject` | Eject Derived File (Own It) | Same. |
| `monosmith.openWalkthrough` | Open Getting Started | Open the walkthrough. |
| `monosmith.commands` | Commands… | Every command in one list, grouped as the terminal's palette (Start, Build, Projects, Dependencies, Release, Health, Go to, Help), each with the smith command line it runs under its title. |
| `monosmith.buildAll` | Build All | `smith run build` in every project, after a yes. |
| `monosmith.lintAll` | Lint All | `smith run lint` in every project, after a yes. |
| `monosmith.testChanged` | Lint and Test What Changed | `smith run lint test --affected`. |
| `monosmith.runs.open` | Open the Run | Show a run in the Runs view, the newest when none is given. |
| `monosmith.runs.showLog` | Show the Log of This Task | A row action of Runs; hidden from the palette. |
| `monosmith.runs.errorsOnly.on` / `.off` | Show Errors Only / Show All Tasks | Narrow the Runs view to the failed tasks, or widen it again. |
| `monosmith.dependencies.updateSafe` | Update All Safe… | `smith deps up <every patch and minor release>` in a terminal, after a yes. |
| `monosmith.dependencies.repair` | Repair Dependencies… | `smith deps repair` in a terminal, after a yes. |
| `monosmith.release.dryRun` | Release Dry Run | The output of `smith release --dry-run` as a page. |
| `monosmith.release.notes` | Release Notes | The output of `smith release --dry-run --notes` as a page. |
| `monosmith.settings.refresh` | Refresh Settings | |
| `monosmith.settings.change`, `.reset`, `.plugin`, `.openRecipe` | | Row actions of Settings; hidden from the palette. Change and reset open **Change a Setting of smith.toml…** and **Reset a Setting…** on that setting; the plugin row asks On or Off, each with its command, then the confirmation. |

`Re-run Last`, `Cancel Run` and similar commands are enabled by the context keys
`monosmith.hasLastRun` and `monosmith.running`; the palette shows the others only when
`monosmith.state == ready`.

## Keyboard shortcuts

| Keys | Command | When |
|---|---|---|
| `Enter` | Run Focused Task | the Projects view has focus |
| `Ctrl+Alt+M T` | Run Task | ready |
| `Ctrl+Alt+M A` | Run Affected | ready |
| `Ctrl+Alt+M R` | Re-run Last | ready, there is a last run, nothing is running |
| `Ctrl+Alt+M C` | Cancel Run | a run is going |
| `Ctrl+Alt+M P` | Commands… | ready |

## Status bar

An item on the left shows the smith that runs, `· N affected` when known, and `· N failed`
for tasks that failed in the runs so far. The icon is a spinner during a run, a warning
triangle when tasks failed, and a slashed circle, shield or error icon when smith is missing,
untrusted or failed. The tooltip says which smith runs and why. Clicking opens a menu:
Cancel run (while running), Run affected…, Re-run last, Run doctor, Refresh, Open output,
Open settings.

## Settings

| Setting | Type | Default | Effect |
|---|---|---|---|
| `monosmith.path` | string | empty | The smith executable to use. Empty: the workspace's own (`node_modules/.bin`, `.smith/bin`; only in a trusted workspace), then the `PATH`. |
| `monosmith.run.bail` | boolean | `false` | Stop a run at the first failed task. |
| `monosmith.run.noCache` | boolean | `false` | Run tasks again even when their result is in the cache. |
| `monosmith.run.parallel` | integer, at least 0 | `0` | How many tasks a run may run at once. 0: smith chooses (the number of CPUs). |
| `monosmith.refresh.auto` | boolean | `true` | Refresh the projects and what is affected by themselves when `smith.toml` changes, the git branch changes, a file is saved or a run ends. Off: use the Refresh button. |
| `monosmith.refresh.debounceMs` | integer, 100 to 30000 | `1500` | How long to wait, in milliseconds, for a burst of changes to end before refreshing once. |
| `monosmith.output.problems` | boolean | `true` | Put the failures found in task output (tsc, ESLint, go and the like) in the Problems panel. |

## Tasks and problem matchers

A **task type `smith`** appears in Run Task and can be written in `tasks.json`:

| Field | Type | Meaning |
|---|---|---|
| `task` | string, required | The task to run, such as `build`, `test` or `lint`. |
| `project` | string | The projects to run it in: names, globs, `kind:app`, `lang:go`. Every project that has the task when empty. |
| `affected` | boolean | Only the projects the changes affect. |
| `bail` | boolean | Stop starting tasks after the first failure. |
| `noCache` | boolean | Run the tasks even when their result is in the cache. |
| `parallel` | integer, at least 1 | How many tasks to run at once. |

```json
{ "type": "smith", "task": "test", "affected": true, "problemMatcher": ["$smith", "$smith-tsc", "$smith-eslint"] }
```

Problem matchers (they accept the task-name prefix `<project>:<task> │ ` that smith prints):

| Matcher | Reads |
|---|---|
| `$smith` | `path:line:col: message` as printed by go, gcc, ruff, mypy. |
| `$smith-tsc` | `file(line,col): error TSnnnn: message` from tsc. |
| `$smith-eslint` | ESLint's stylish output (a file line, then `line:col severity message rule` lines). |

The matchers know a path relative to the workspace folder; to place a failure by the project
the task ran in, run the task from the view instead. With `monosmith.output.problems` on, the
view's runs put the failures of tsc, ESLint, go, gcc, ruff and rustc in the Problems panel and
clear them when the task runs again. A path in the output (`src/main.ts:12:5`) is a link, also
when relative to the project.

## Walkthrough

**Help, Get Started, Get started with MonoSmith** (or `MonoSmith: Open Getting Started`) has
six steps: have smith in the workspace, see your projects, run a task, run only what changed,
check the health of the workspace, and use it from `tasks.json` and the keyboard.

## Output

The **MonoSmith** output channel shows the exact command and the task output as it arrives, and
what `smith serve` writes to standard error. A smith that ends by itself is noticed at once,
with an offer to restart it.

## Related

[`smith serve`](serve.md), [`smith doctor`](doctor.md), [`smith run`](run.md),
[`smith up`](up.md), [`smith release`](release.md), [Editors](../guides/editors.md),
[The architecture of the extension](../architecture/README.md).
