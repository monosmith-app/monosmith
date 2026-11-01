# Changelog

The version of the extension is its release tag: a release writes it. Until
the first release everything below is under Unreleased.

## Unreleased

### Added

- Swap a Choice of a Project: quick picks for the project, the decision and the answer, the plan of `smith swap --dry-run` first, then the swap with an optional commit.
- Show Divergences from the Defaults lists what `smith divergences` finds.
- Add project asks which project to copy (`smith add --like`) when two or more projects share the language.
- Every derive change offers "and commit (so Undo works)", and deriving every file offers Preview (`smith derive --dry-run`).
- Add project offers the buttons Add and commit (so Undo works); smith commits nothing otherwise, and the notice says Not committed: review with git status. Add project refuses a name a project has already.

### Changed

- **VS Code 1.140 or newer is required** (it was 1.96). An older editor keeps the
  version it has and does not offer this one.

### Added

- **The 2.0 commands.** Add Project… (quick picks, the plan of `smith add --dry-run`, then the add),
  derived files in the Health view and the Problems panel with Explain, Apply, Adopt and Eject,
  Why… on a file, task or project, Undo the Last Change…, Open Report, Write AGENTS.md for AI
  agents, and the release channel in the Release Preview. Each command that writes asks first.
- **Start in an empty folder.** An empty folder says `This folder is empty.` and offers
  Create a new MonoSmith workspace (`smith new --here` in a terminal, which asks its questions);
  a folder with files but no tool smith recognises offers Bring this folder into MonoSmith
  (`smith adopt`) and See what would change. The same entries are in the status bar menu.
  The extension also wakes up for any folder with a file at its root, and when its view is opened.
  Nothing is written by the extension itself.
- **Bring a repository into MonoSmith.** The extension also wakes up in a folder
  with `nx.json`, `turbo.json`, `lerna.json`, `pnpm-workspace.yaml`, `rush.json`,
  `go.work` or a `package.json` that lists workspaces, and the Projects view says
  what it looks like with three links: Convert it to MonoSmith (opens `smith` in a
  terminal, which lists every change and asks before it makes any), See what would
  change (`smith adopt --dry-run`, writes nothing) and Learn more. A folder with
  none of those stays quiet, and nothing is written by the extension itself.
- **The MonoSmith agent.** Commands to install the MCP server and the agent
  (`.claude/agents/smith.md`, `.mcp.json`, `.vscode/mcp.json`) after showing the
  files, to install the MCP server alone, and to talk to the agent in a
  terminal (`smith agent`). A workspace without the files gets a line in the
  Projects view and a one-time notification with Install, Not now and Never for
  this workspace.
- **Health, Dependencies and Release Preview views** beside the projects. Health
  lists the checks that failed or warn with the problem and the remedy, and a
  Fix action applies the fixes `smith doctor` proposes after showing them.
  Dependencies lists what `smith up` finds, grouped patch, minor and major, with
  a button that opens `smith up` in a terminal and an Update action per
  dependency that asks first. Release Preview shows what `smith release` would
  publish, and publishes nothing.
- **Explain** and **Show Project Graph** on a project row: a readable page of
  how its tasks resolved (and where each field came from), and of what it uses
  and what uses it.
- **Problems panel.** The failures that tsc, ESLint, go, gcc, ruff and rustc
  print become entries of the Problems panel, with the file and the place, and
  are cleared when the task runs again. `monosmith.output.problems` turns it off.
- **Clickable paths.** `path:line:col` in the Output channel and in task logs
  opens the file at that place, also for paths relative to a project.
- **Show Output on a task row**, also for a failed task: the last log of each
  task opens as a document.
- **A quieter project tree.** An affected project shows a dot with a tooltip
  instead of the word, and no dot at all when every project is affected (the
  title says so). Kind and status icons, durations, the cache hit and the last
  result in tooltips. A filter, and an affected-only toggle.
- **A `smith` task type** for Run Task and `tasks.json`, with the `$smith`,
  `$smith-tsc` and `$smith-eslint` problem matchers.
- **Settings**: `monosmith.refresh.auto`, `monosmith.refresh.debounceMs`,
  `monosmith.run.parallel`, `monosmith.output.problems`.
- **A getting-started walkthrough** (Help, Get Started).
- Integration tests that start the extension in a VS Code host
  (`npm run test:integration`).
- The first version: the Projects view, running tasks from the keyboard, the
  status bar, refreshing by itself, workspace trust, and a state for every way
  smith can be missing. See the README.
