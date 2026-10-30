# The MonoSmith VS Code extension

The MonoSmith extension runs the tasks of a smith workspace from VS Code. It adds an
activity-bar container with four views (Projects, Health, Dependencies, Release Preview),
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
| Activation | When the workspace contains a `smith.toml` (`workspaceContains:smith.toml`). |
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

All views are in the **MonoSmith** container of the activity bar. Health, Dependencies and
Release Preview appear (collapsed) only when smith is ready.

### Projects

Apps, libs and packages with their language. A dot marks a project the changes affect (left
out when every project is affected; the title then says `all N affected`). Under a project, its
tasks, each with an icon for how it ended, how long it took, and a tooltip with the last result
and whether it came from the cache. A multi-root workspace gets one section per folder that has
a `smith.toml`.

When smith is not usable the view says why instead of staying empty, with buttons:

| State | Message | Buttons |
|---|---|---|
| starting | `Starting smith…` | Show log |
| missing | `smith was not found.` | Copy install command, Open settings, Retry, Show log |
| failed | `smith could not start, or it stopped.` | Restart smith, Show log, Open settings |
| untrusted | `This workspace is not trusted.` | Manage Workspace Trust, Open settings, Retry |
| empty | `No open folder has a smith.toml.` | Learn about MonoSmith |

Title-bar buttons (when ready): Filter Projects…, Show Affected Only / Show All Projects,
Run Affected, Re-run Last (not while running), Cancel Run (while running), Refresh. The
`…` menu: Run Task, Run Task in All Projects, Clear Filter (when filtered), Run Doctor, Clear
Problems, Restart smith, Show Output, Open Settings, Open Getting Started.

Row actions: on a task, **Run Task** (inline and context menu) and, when it has run, **Show
Task Output**; on a project, group or folder, **Run All Tasks**; on a project, **Explain
Project** and **Show Project Graph**. Enter on a row runs it (`Run Focused Task`).

The filter narrows the view by project, task or language as you type; it is remembered per
workspace.

### Health

The checks of `smith doctor` that failed or warn, each with the problem and the remedy.
**Fix Health Problems…** (shown when a fix exists) shows the changes smith proposes and
applies them only after you agree. **Refresh Health** reads again.

### Dependencies

What has a newer release, grouped patch, minor and major, and the tools smith pins.
**Run smith up…** opens `smith up` in a terminal. **Update Dependency…** on a dependency asks
first, then runs `smith up <name>` in a terminal. The editor never rewrites a manifest
itself: smith has no method for it in the editor protocol.

### Release Preview

What `smith release` would publish: each package with the version it moves to, why, its tag
and the commits. It publishes nothing. **Refresh Release Preview** reads again.

### Explain and graph pages

**Explain Project** opens a page with how each task of the project resolved and where each
field came from ([`smith explain`](explain.md)). **Show Project Graph** opens what the project
uses and what uses it ([`smith graph`](graph.md)).

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
| `monosmith.dependencies.up` | Run smith up… | |
| `monosmith.dependencies.update` | Update Dependency… | A row action; hidden from the palette. |
| `monosmith.release.refresh` | Refresh Release Preview | |
| `monosmith.openWalkthrough` | Open Getting Started | Open the walkthrough. |

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
