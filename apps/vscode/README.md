# MonoSmith for VS Code

Run the tasks of a [MonoSmith](https://github.com/monosmith-app/monosmith)
workspace from the editor: see the projects and what the changes affect, run a
task from the keyboard, read the failures in the Problems panel, and check the
health of the workspace, its dependencies and its next release.

![The projects of a workspace, a dot on those the changes affect, and the Health, Dependencies and Release Preview views](https://github.com/monosmith-app/monosmith/raw/main/apps/vscode/media/screenshot-views.png)

## What you get

- **Projects view.** Apps, libs and packages, each with its language. A dot
  marks a project the changes affect (and is left out when every project is
  affected, since it would tell nothing; the title says `all 6 affected`).
  Under a project, its tasks, each with an icon for how it ended, how long it
  took, and a tooltip with the last result and whether it came from the cache.
  A **filter** (the search button) narrows the view by project, task or
  language as you type, and **Show Affected Only** keeps the affected projects.
  A multi-root workspace gets one section per folder that has a `smith.toml`.

  ![A project with the tasks of its last run: passed with a duration, failed with the reason](https://github.com/monosmith-app/monosmith/raw/main/apps/vscode/media/screenshot-tasks.png)
- **Run from the keyboard or the mouse.** Enter on a task runs it. **Run Task**
  (`Ctrl+Alt+M T`) asks for a project, then a task or all of them. The context
  menu and the inline buttons run a task, or all the tasks of a project, a
  group or a folder. **Run Affected** (`Ctrl+Alt+M A`) runs one task in the
  projects the changes affect, **Run Task in All Projects** wherever it
  exists, **Re-run Last** (`Ctrl+Alt+M R`) repeats the last run, and
  **Cancel** (`Ctrl+Alt+M C`) stops the run and its processes.
- **Problems panel.** The failures that tsc, ESLint, go, gcc, ruff and rustc
  print become entries in the Problems panel, with the file and the place,
  and are cleared when the task runs again. A path in the output
  (`src/main.ts:12:5`) is a link, also when it is relative to the project.

  ![Failures of a typecheck and a lint run in the Problems panel](https://github.com/monosmith-app/monosmith/raw/main/apps/vscode/media/screenshot-problems.png)

- **Output of every task.** The MonoSmith output channel shows the exact
  command and the task output as it arrives. **Show Task Output** on a task
  row (also a failed one) opens the last log of that task as a document.
- **Health.** The checks of `smith doctor` that failed or warn, each with the
  problem and the remedy. **Fix** shows the changes smith proposes and applies
  them only after you agree.
- **Dependencies.** What has a newer release, grouped patch, minor and major,
  and the tools smith pins. **Run smith up** opens `smith up` in a terminal;
  **Update** on a dependency asks first, then runs `smith up <name>` there.
  Smith has no method to update through the editor, so the editor never
  rewrites a manifest itself.
- **Release Preview.** What `smith release` would publish: each package with
  the version it moves to, why, its tag and the commits. It publishes nothing.
- **Explain** and **Show Project Graph** on a project (context menu or the
  palette) open a page: how each task of the project resolved and where each
  field came from, and what the project uses and what uses it.

  ![Explain page of a project](https://github.com/monosmith-app/monosmith/raw/main/apps/vscode/media/screenshot-explain.png)

- **The MonoSmith agent.** A workspace without the agent files shows one line
  in the Projects view, **Add the MonoSmith agent**, and the extension offers
  it once as a notification (**Install**, **Not now**, **Never for this
  workspace**, which records `[agent] enabled = false` through `smith config
  set`). Three commands do the same from the palette:
  **MonoSmith: Install the MCP server and agent** writes
  `.claude/agents/smith.md`, `.mcp.json` and `.vscode/mcp.json`;
  **MonoSmith: Install the MCP server only** writes `.vscode/mcp.json`;
  **MonoSmith: Talk to the MonoSmith agent** opens a terminal that runs
  `smith agent` (Claude Code with the agent), or says how to install Claude
  Code and offers the MCP config to copy. Every install shows the files first
  (with the diff on request) and writes only after you agree; smith merges into
  an existing MCP file and never overwrites a file you edited, and says so
  when it stops. See [the agent guide](https://github.com/monosmith-app/monosmith/blob/main/docs/guides/agent.md).

  To use the server in VS Code (1.99 or later), open `.vscode/mcp.json`,
  choose **Start** above the `smith` server (or run **MCP: List Servers**),
  trust it when VS Code asks, and pick its tools in Chat's agent mode.
- **Tasks and tasks.json.** Tasks of type `smith` appear in **Run Task** and
  can be written in `tasks.json`; the `$smith`, `$smith-tsc` and
  `$smith-eslint` problem matchers read the failures out of the terminal.

  ```json
  { "type": "smith", "task": "test", "affected": true, "problemMatcher": ["$smith", "$smith-tsc", "$smith-eslint"] }
  ```

  The matchers know a path relative to the workspace folder; to place a
  failure by the project the task ran in, run the task from the view instead.
- **Status bar.** The smith that is running, how many projects are affected
  and how many tasks failed. The tooltip says which smith runs and why; a
  click opens a menu.
- **Refreshes by itself** when `smith.toml` changes, the git branch changes, a
  file is saved and after each run; bursts are debounced into one refresh.

Every action is reachable without the mouse, and the rows carry text for
screen readers. **Help, Get Started, Get started with MonoSmith** walks
through it.

## Settings

| Setting | Default | |
| --- | --- | --- |
| `monosmith.path` | empty | The smith to use. Empty: the workspace's own, then the PATH (see below). |
| `monosmith.run.bail` | `false` | Stop a run at the first failed task. |
| `monosmith.run.noCache` | `false` | Run tasks again even when their result is in the cache. |
| `monosmith.run.parallel` | `0` | How many tasks a run may run at once; 0 lets smith choose. |
| `monosmith.refresh.auto` | `true` | Refresh by itself; off: use the Refresh button. |
| `monosmith.refresh.debounceMs` | `1500` | How long a burst of changes may last before one refresh. |
| `monosmith.output.problems` | `true` | Put the failures found in task output in the Problems panel. |

## Which smith

The extension does not carry smith. It runs the one your team installed, so the
editor, the terminal and CI give the same result. It looks for it in this order:

1. the `monosmith.path` setting
2. the workspace's own install: `node_modules/.bin/smith` (`npm i -D monosmith`)
   or `.smith/bin/smith`, **only in a trusted workspace**
3. `smith` on the `PATH`

The status bar tooltip says which one runs. When the workspace's smith is not
the version a terminal gets from the `PATH`, the extension warns once per
session and offers to open the settings or copy the update command.

## Workspace trust

A workspace's own smith is an executable the folder brings, so in an untrusted
workspace (Restricted Mode) the extension does not run it, and it does not run
tasks either; they execute code from the workspace. It uses the `monosmith.path`
user setting or the `PATH` to show the projects, says why in the view, and
starts again by itself when you trust the workspace. The workspace value of
`monosmith.path` is ignored while the workspace is untrusted.

## When there is no smith

The view tells what is wrong instead of staying empty, with buttons for what to
do: smith not found (copy the install command, open settings, retry, show log),
could not start or stopped (restart, show log, open settings), not trusted
(manage workspace trust), starting. What `smith serve` writes to stderr goes to
the MonoSmith output channel, and a smith that ends by itself is noticed at
once, with an offer to restart it.

It talks to `smith serve` (JSON-RPC over stdin and stdout, no network port) and
works with any smith whose protocol it knows; see
[Serve and the event stream](https://github.com/monosmith-app/monosmith/blob/main/docs/reference/serve.md).

## Developing

```sh
smith run typecheck lint test build -p vscode   # the checks
npm run test:integration -w monosmith-vscode    # starts VS Code with a fake smith
```

The integration tests start a real VS Code (downloaded the first time, or the
one `VSCODE_EXECUTABLE_PATH` names) with the extension loaded and a stand-in
`smith serve`, in three scenarios: a smith that runs, no smith, and a workspace
that is not trusted. On Linux without a display run them under `xvfb-run -a`.

`smith run package -p vscode` packs a .vsix into dist/drop, and a release
publishes it to the Marketplace, with the token in `VSCE_PAT` or, when the
workspace's `smith.toml` sets `[release] vsce_entra = true`, as a Microsoft Entra
ID app signed in by the workflow (no token). When `OVSX_PAT` is set it also
publishes to Open VSX. See
[Releasing](https://github.com/monosmith-app/monosmith/blob/main/docs/reference/releasing.md#the-marketplace-without-a-token-microsoft-entra-id).

The full list of commands, views and keys is in the
[manual](https://github.com/monosmith-app/monosmith/blob/main/docs/manual/vscode.md).
