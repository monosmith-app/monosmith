# MonoSmith for JetBrains IDEs

The JetBrains plugin of issue #303. It offers the same functions as the VS Code extension
([feature parity](../../docs/reference/feature-parity.md), JetBrains column), in IntelliJ's own way.

## What it does

- The **MonoSmith** tool window (right side) has three tabs.
  - **Projects**: the projects of the workspace and their tasks, the affected ones marked, a filter box and
    an "Affected only" box. Double-click a task to run it. A row
    **Add your first project** (or **New project…** after the projects, or **New workspace…** when there is
    no `smith.toml`) starts the flow. Right-click a project for **Add library…**, **Install package…**,
    **Unlink library…**, **Swap a choice…**, **Make app, library or package…**, **Remove…**, **Why…**,
    **Explain** and **Run task**. The title bar has a Refresh button.
  - **Dependencies**: the updates of `smith up --json` grouped by risk (major first) and the tools with a
    newer version. **Update dependencies…** picks every patch and minor update or a list (a major update is
    never ticked), shows the plan, then writes; **Align versions…** asks which version wins (`up --align`).
  - **Release**: the plan of `smith release --dry-run --json` with the reason and commits of each package,
    the channel, and why nothing releases; **Release…** starts the flow (hold, bump, force).
  - **Health**: the findings of `smith doctor --json`, failed checks first. A finding with a structured fix
    has a **Fix this check** row: smith's questions, then its review (the diffs and the plan of each command),
    then Apply.
  - **Output**: the command line and what smith prints, as it prints it.
- **Tools > MonoSmith** holds the other functions: Add Project, Compose a Stack, Add a Library, Install a
  Package, Create a New Workspace, New Recipe, Install the Tools the Workspace Needs, the Project group
  (Remove, Make a project an app/library/package, Unlink, Swap a choice, Sync versions), Environments,
  Derived Files, Commit the Changes, Undo, Release, Release Preview, Doctor, Doctor: Apply the Safe Fixes,
  Divergences, Run Task in All Projects, Run Affected, Start Dev Servers, Update and Align Dependencies, Open the
  Report, Search and Install a Plugin, Settings, Bring This Folder In (adopt), Convert This Workspace, Try a
  Command in a Copy, Write the smith Launcher, Set Up Tab Completion, Install the MCP Server and Agent, Explain a Project and Why.
  **Set Up Tab Completion** types `smith completion install` in a MonoSmith tab of the terminal, which lists
  the files and asks before it writes; the plugin writes nothing. After a project opens, a notification offers
  it once (**Set up**, **Not now**, **Don't ask again**) when `smith completion status --json` says it is not set
  up. The answers **Set up** and **Don't ask again** are kept for you in the IDE settings
  (`monosmith.tabCompletionOffer`). The plugin needs the bundled Terminal plugin.
  **Why** is also in the editor and project view menus. Five groups hold the rest:
  - **Plugins**: **Turn a Plugin On…** and **Off…**. The plugins come from `smith plugin list --json`. When
    projects use a plugin, smith refuses to turn it off and names them; the plugin shows that list and runs
    `--force` only after a second yes.
  - **Cache**: status and check for damage (they print in the Output tab), then **Remove What Is Broken…**,
    **Trim to Its Limit…** and **Delete It All…**, each after its plan or what it costs.
  - **Pipelines and Apps**: **Add a Ready-Made Pipeline…** (the templates come from `smith pipeline`),
    **Regenerate the CI Pipelines…** (a diff of every file), **Make an App a Tauri Desktop App…** (asks the
    bundle identifier), **Make the Icons of a Tauri App…** and **Set the Docker Hub Descriptions…** (the
    repositories first, the call only after a yes).
  - **Look Up**: **Why Would a Task Run Again?…**, **Tools the Workspace Needs**, **Info** and **Copies Between
    Projects**. They only read.
  - **Dependencies: Require One Version…** edits `dependencies.one_version` through the settings editor with its
    reason and the plan of `smith config set`. **Write the Agent File** and **Connect smith to a Coding
    Assistant…** (`smith mcp install`; the assistants start unticked) are in the menu too.
- The projects, kinds, choices, outcomes, tools, recipe levels and doctor fixes all come from smith's JSON
  (`projects`, `affected`, `kinds`, `choices`, `compose`, `tools install --dry-run`, `recipe new --list-levels`,
  `doctor`, `release --dry-run`, `up`, `config list`, `plugin list`, `plugin search`, `adopt --dry-run`,
  `new --questions`, `convert --questions`). The plugin holds no list of languages, frameworks, tools, CI
  systems or layouts. The questions of `smith new` come from `smith new --questions --json`, and the level and
  the four choices of `smith convert` come from `smith convert --questions --json`.
- Each question shows the one-line "what is this" of its layer from the catalogue (`about` in `smith compose --json`).

## The rules of every flow

- **Every question is a presented choice.** The default is the first row and says `(default)`. A list of one
  answer is radio buttons, and an axis with no default starts on **None**. A list of many is check boxes.
- **Every change shows its plan first** (the text of a `--dry-run`), and nothing is written before the person
  answers. A flow ends with `--yes`, because the person already agreed; a command run with no terminal never
  asks.
- **smith never commits unasked.** The plan dialog has an **unticked** box "Commit this change so it can be
  undone"; `--commit` goes on the command only when it is ticked. After a change that is not committed, a
  **Commit changes…** button offers to commit it (only the files of the operation, everything git shows, or
  cancel).
- **A failure is a notice with at most three short buttons.** The rest are behind **More…**, which lists the
  long sentence of each. A project that was written and then failed its checks says "created", not
  "refused".
- smith is the **installed** one ([ADR 0025](../../docs/architecture/decisions/0025-ide-plugins-run-the-installed-smith.md)):
  `node_modules/.bin/smith`, then `.smith/bin/smith`, then `smith` on the `PATH`. When it does not start, a
  notice offers `npm i -g monosmith@latest` to copy.

## Not built

The live `smith serve` connection and the language server (`smith lsp`, not built in the engine either).
The Go module prefix question of `smith new` is not asked: smith writes its default, as the VS Code flow
does. "Let smith run
it" for a tool that needs administrator rights runs in the Output tab, where a `sudo` password cannot be
typed. Four functions have no JetBrains version, and each has its reason in the
[feature parity](../../docs/reference/feature-parity.md) table: the footer keys of the TUI, the files named
with `-o` or `--force`, `smith plugin draft` (a plugin author works at a terminal on a sample project) and
the scroll of the lists (the IDE's own lists scroll as the IDE does).

The dialogs, notifications and tool window have not been run in a real IDE: CI compiles them against the
platform, and every flow has unit tests with a fake smith and a fake person (`FakePortsHarness`).

## Build and test

```
smith run build test lint -p jetbrains
```

Needs a JDK 21 and Gradle (smith's CI installs both). The first build downloads IntelliJ IDEA
Community 2025.1 (large, cached in `.intellijPlatform/`). `gradle :apps:jetbrains:buildPlugin`
writes the installable zip to `dist/gradle/monosmith-jetbrains/distributions/`.
`gradle :apps:jetbrains:runIde` starts a sandbox IDE to try it.

## Publishing

Nothing publishes this plugin to the JetBrains Marketplace; it needs a Marketplace token and
signing certificate, and deciding to publish is the owner's. The project is a Kotlin `app` of
the workspace, so `smith release` does not release it.

## Where the texts live

Every text of the plugin comes from the shared catalogue in `libs/words` ([ADR 0089](../../docs/architecture/decisions/0089-one-text-catalogue-for-every-interface.md)).
Edit `libs/words/catalog/en/jetbrains_*.json` (keys `jetbrains.<area>.<what>`), or a shared key claimed in `libs/words/surface/jetbrains.json`.
Then run `go run ./wordgen -surface jetbrains` in `libs/words`. It writes `src/main/resources/messages.properties` and `words/SmithWordsConfig.kt`.
Do not edit those two files by hand. `plugin.xml` reads the bundle through `<resource-bundle>` (keys `action.<id>.text`, `action.<id>.description`, `group.<id>.text`).
Kotlin code calls `say(KEY, values)` from `SmithWordsConfig`. A Go test (`libs/words/words/jetbrains_literals_validator_test.go`) fails on a new raw sentence.

## Layout

Slices under `src/main/kotlin/dev/monosmith/jetbrains/`, named `<Name><Role>` as `smith slices check --lang jvm`
requires. The pure parts (mappers, algorithms, policies, contracts) have unit tests; the IDE parts are
compiled against the platform.

- Foundation: `flowports` (the three ports a flow is written against: smith, the person, the notices),
  `jsonread` (read smith's JSON), `smithrun` (run smith, the Output console, background flows),
  `notice` (balloons and the three-button split), `prompt` (the radio, check box and plan dialogs),
  `planapply` (plan, ask, apply), `workspace` (find the root, list the projects), `smithcommand`, `taskrun`.
- Features, one slice each: `addproject`, `addchoices`, `composestack`, `installdependency`, `projectactions`,
  `swapchoice`, `syncversions`, `commitchanges`, `undochange`, `releaseflow`, `health`, `installtools`,
  `derivefiles`, `newworkspace`, `newrecipe`, `trycommand`, `workspacetools`, `smithversion`, `runacross`,
  `dependencies`, `workspacereport`, `pluginmarket`, `pluginswitch`, `settings`, `adoptfolder`, `convertfolder`,
  `devservers`, `cachecontrol`, `pipelinetemplate`, `ciregenerate`, `tauriapp`, `dockerhub`, `mcpinstall`,
  `agentfile`, `readcommands`.
- Entry points: `toolwindow` (the tool window) and `smithaction` (one handler for each menu action).

## Written exceptions to the slice rules

### `smithaction` is a catalogue

- Path: `src/main/kotlin/dev/monosmith/jetbrains/smithaction`.
- Rule waived: the review trigger of 12 production files in one slice.
- Constraint: the slice holds 62 handlers, one class for each menu action of `plugin.xml`. The IDE needs one
  class for each action, and the actions share no outcome. Each handler only calls one use case of another
  slice. Splitting the folder by action would add a slice for each class and no cohesion.
- Owner: the smith maintainers.
- Temporary: no. Review it when the shared command catalogue (`guided-commands.config.ts`) exists.

### Package dependencies

No package imports a package that imports it back. `smithrun` owns `FailureTextMapper` and
`ProblemTextMapper`. `flowports` owns `WorkspaceChangedStore`. `workspace` may use both.
