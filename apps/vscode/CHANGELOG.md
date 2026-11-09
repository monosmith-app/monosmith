# Changelog

The version of the extension is its release tag: a release writes it. Until
the first release everything below is under Unreleased.

## Unreleased

### Added

- Convert an Nx or Turborepo Workspace with Your Choices: asks the level and, at the full level, lint, tests, build and pipelines from `smith convert --questions --json`, shows the plan with the files it deletes, asks the unticked commit box, then converts.
- New workspace: the Customise questions (CI, language, Node tests, code layout, agent files, dev container) come from `smith new --questions --json`, so a question that smith adds appears without a new version of the extension.
- Environments of a web app: "Environments of the Web Apps", "Add Environments to a Web App…" (the names start on `[workspace] environments`, to edit; the plan lists the files) and "Remove Environments from a Web App…" (only the ones that can go; the `.env` file stays).
- New Recipe: for each level that takes one plugin it asks who decides it, the recipe (default) or the people who use it. For the second it asks the question they read, the plugins they can end up with and a name for each answer, and writes `--chosen`.
- One command each, with the plan or the cost shown first: Cache (status, check for damage, remove what is broken, trim to its limit, delete it all), Copies Between Projects, Tools the Workspace Needs, Info, Add a Ready-Made Pipeline, Make the Icons of a Tauri App, Make an App a Tauri Desktop App, Docker Hub descriptions, Why Would a Task Run Again, Connect smith to a Coding Assistant, Regenerate the CI Pipelines, and Derived Files (check, preview, update).
- Change a Setting of smith.toml and Reset a Setting: a pick list of the settings with their value and the reason of the default; a boolean or an enum is a pick list, and any other starts on the current value.
- Turn a Plugin On or Off (smith's warning about the projects it leaves comes with the answer) and Choose the Architecture of the Workspace.
- Install a Plugin from the Marketplace: search, pick with the description, the trust question in plain words (install only, or trust and pin this version), then the confirmation. An empty marketplace says so.
- New Recipe: a level that takes any number of plugins (structure, deploy, toolchain) asks who decides it as well; an answer of a choice gives one plugin or several joined by `+`.
- Add project: the language question starts on the language most projects of the workspace use and says so.
- Release without a releasing commit: a package that was never released is not offered, because smith has no version to bump. One line names it and offers to copy the commit that sets its first version.

### Fixed

- Add project: the choices lists show only the choices that fit the kind you picked (`kinds` of `smith choices --json`), so a package is not offered aws-lambda, electron or tauri. Before the lists are left, smith checks the picks (`smith compose --json --with`): a conflict or a missing piece is said in plain words, and a website, desktop or function app without its framework, shell or host is told "this kind of project needs a framework: tick one of html, react, svelte, vue, or untick everything" and can pick again. Lint and test choices alone are fine. Compose a stack no longer says that a choice that writes the project is needed.
- Add project: pressing Enter through the steps gives the safe defaults. Every list of one answer starts on its default (Vitest for the test runner) or on "None" when the axis has none (desktop shell, style), listed first; it no longer starts on Tauri, Bootstrap or Jest. The confirmation says the choices in words and marks the defaults ("React, ESLint (default), Vitest (default); no desktop shell, no style").
- Add project: the framework is asked once. The answer of "Which framework?" is kept, a website no longer lists AWS Lambda or Electron, and the framework list starts on what you typed in the first question (`react`).
- Notifications show at most three buttons with short verbs (Install tools, Open Health, More…); the long explanation is in the message and the other buttons are behind "More…".
- `MonoSmith: Release…` says why when nothing releases and offers "Open Preview" and, when a package exists that no commit releases, "Without a commit…".

- New workspace is a flow of its own, not a bare terminal: "How do you want to set it up?" (the defaults, which list what they choose; Customise, which asks CI, language, Node tests, code layout, agent files, dev container and the first commit; or answer in a terminal), the name with the folder name filled in and said to be the folder's, then a summary and the list of files before anything is written.
- Terminals show `smith new --here`, not the full path of smith: the folder of smith goes first on the PATH of the terminal.
- The Projects view of a workspace with no project shows "Add your first project".
- Health has its own panel (MonoSmith Health), wide enough for its messages.

- Compose a stack: the plan lists each tool with its state (`Node.js (installed, for node)`, `the Tauri system libraries (missing, for tauri): smith tools install tauri-linux-libs`) and offers "Install the tools first" when one is missing; the plan is read again after the tools flow.
- Compose a stack: the picker keeps what you ticked visible when you change the search, names it in the title ("Selected: React, Electron"), and an axis that takes one choice behaves like a radio button. The axis comes from the catalogue, not from names.
- Compose a stack: a layer that answers another question is labelled so (Electron: "shell (the framework choice)"), which agrees with the conflict message.
- Compose a stack: the steps are numbered 1, 2, 3 (pick, name, plan); the name prompt names the stack and says why the name comes before the plan; the example name is `web`.

- A project that smith wrote and then found failing `smith doctor` (a system library is missing) is no longer reported as "refused" and "cancelled". A notice gives the facts (created, N files, what could not be checked and why, whether the commit happened) and the next step as buttons: Install the tools, Open Health, Commit anyway, Leave it uncommitted. smith's full steps go to the Output channel. Compose a stack and Add project both do this.

- Add project: the first question lists what smith lists (`outcomes` of `smith compose --json`), so "A website or web app" and "A desktop app" are there; the extension keeps no list of its own.
- Add project: typing `react`, `tauri` or `electron` in that question finds the website or the desktop app, as in the TUI; a search with no match shows a "No result" row with the closest answers.
- Add project: "Compose a stack" opens the composer after the add ends, and the Output channel says nothing was added, not "cancelled".
- Add project: the language question of a function app names each language ("C# / .NET") with its clouds, and lists every language smith can generate.

- Add project: the website and the desktop app were missing from the first question. `smith kinds` names the language of `node-app` "typescript" and `smith compose` names it "node", so the extension dropped every outcome only node builds. The language of a kind is now the id before its role; a spec reads the real JSON of smith as a fixture.
- Add project and Compose a stack: the multi-select keeps the cursor and the focus (the list no longer jumps to the first row, and a click outside does not close it).
- Add project: a framework, a desktop shell and a cloud host carry one name in every step, the title of its layer in the catalogue ("Desktop shell"), not `framework: electron` in one step and "desktop shell" in the next.

### Added

- **Commit the changes…** (`smith commit`): shows the files of `smith commit --dry-run`, then asks for only the files the last operation changed (the default), everything git shows as changed (the extra files are listed), or cancel. "Commit anyway" and the buttons of a failed doctor use it.
- **Release…**: the `smith release --dry-run` plan in a modal, a multi-select of packages to hold back (`--hold`), a pick of package and level to change a bump (`--bump`), then the last question, which says that tags and registry publish cannot be undone, and `smith release --yes`.
- The menu of a project: Add a Library, Install a Package, Unlink a Library, Swap a Choice, Make it an App, Library or Package (`smith kind`) and Remove (keep a copy, delete or cancel). Each shows smith's plan first and commits nothing.
- The Release Preview shows the explanation of the CLI when nothing releases (which commits count, that apps and libraries never release). The Projects, Dependencies and Release Preview views share the height by weight, so none is squeezed off the screen.
- New Recipe: input boxes for the name, version and parent, then a pick list or tick boxes for each level of `smith recipe new --list-levels --json`, and the exact command and file before it writes.
- Swap a Choice of a Project: quick picks for the project, the decision and the answer, the plan of `smith swap --dry-run` first, then the swap with an optional commit.
- Show Divergences from the Defaults lists what `smith divergences` finds.
- Why asks about a setting of smith.toml (`smith why <setting>`): what it does, its value and the reason of its default.
- Try a Command in a Copy runs `smith try` and shows the files it changed and the checks; it runs the command for real only on a yes.
- Write the smith Launcher writes `tools/smith/smith` (`smith bootstrap`) after a question.
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
