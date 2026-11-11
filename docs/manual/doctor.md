# smith doctor

`smith doctor` checks the workspace's health: smith's own invariants and each plugin's
checks. Every problem names its remedy. With `--fix` it shows the fixes it can make (file
changes as diffs, commands it would run such as `npm install` or `git rm --cached`, and
questions it would ask) and applies the safe ones.

A **safe** fix only restores what smith or a plugin owns, such as `go.work` or a generated
pipeline with its slots kept; doctor applies those once you confirm (or with `--yes`). Any
other fix, and any fix with a question, is asked about in a terminal. Without a terminal it
is shown with the command that applies it. The same fix on many projects is one fix that
says which projects it applies to.

A fix runs the narrow command it names and nothing more. The step text is the command: for a
golangci-lint that differs from CI it is `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`,
and for a missing Python venv it is the two commands that create `.venv` and install
`requirements-dev.txt`. Every fix verifies itself: after the steps, doctor checks again and says
`Still not fixed: <check>` with what it found, or `Fixed`. The golangci-lint check always names the file it
found and the version line that file printed. When the pinned version is already in the folder `go install`
writes to (`go env GOBIN`, else `GOPATH/bin`) but another copy comes first on PATH, or that folder is not on
PATH, the remedy names both folders and says what to do, for example `PATH finds C:\old\golangci-lint.exe
(1.64.8) before C:\Users\you\go\bin (2.15.0): move go\bin earlier in PATH, or remove the old copy`. smith never
edits PATH or deletes a file.

When the golangci-lint that runs is not the version CI pins (older, newer, or the pinned copy sits behind
another one on PATH), the fix is a question. Nothing runs before you answer:

1. Install the pinned version into smith's own folder, `.smith/bin` in the workspace (recommended). It runs
   `GOBIN=<workspace>/.smith/bin go install ...@v<version>`. The lint task and this check use that copy first,
   whatever PATH says. Your other copy and PATH stay as they are.
2. Keep the one found. smith writes `golangci-lint = "<version>"` under `[toolchains]` in `smith.toml`, and
   doctor stops warning; lint results may differ from CI, which uses the pinned version.
3. Update the copy found, in place (offered when a version was found). The label states the exact command and
   that it changes every project that uses that copy: for a copy in a Go bin folder (`GOBIN`, a `GOPATH` entry's
   `bin`, or `~/go/bin`) built by Go, it runs `GOBIN=<that folder> go install ...@v<version>`. smith does not guess
   another origin (npm, pip, winget, brew, a system folder): the option shows greyed with `Disabled: smith does
   not know how this copy was installed`, and picking it does nothing. If the file is in use on Windows, the
   failure says to close the program and retry; the old copy is not changed. Doctor checks again after it ends.

Moving the CI pin to the version found is not offered: no setting moves the version CI installs. A fix never runs `smith ci setup`, because that command also installs the
dependencies of every other plugin (for example `npm ci`). In the TUI the fix dialog shows the
last six lines the step printed, and `l` opens its full log (see
[the fix wizard](tui.md#the-doctor-fix-wizard)).

In the interactive shell, the detail of a check shows its state, its action buttons, what is wrong and how to fix it, in that order: see [the Doctor detail](tui.md#the-doctor-detail).

## Local toolchain versus CI

For Node, Python, Go, Rust, .NET and the JDK, doctor compares the tool on this machine with
the version the generated CI workflow and dev container install, so no setting is needed
(`Rust is 1.90, as in CI`). A local tool on another version, newer or older, is a warning, never a
failure: a team may run it on purpose, and a fresh workspace on a common setup (Node.js 22 while CI
installs 24) is not unhealthy. The warning says what differs and why it can matter (an older tool
can refuse what CI builds, with an error that reads like a dependency problem). For Node.js the
fix is a question: keep it (the first answer, nothing changes) or get the version CI installs with
`smith tools install node`, which shows the commands and asks. No setting moves the version CI
installs, so no answer edits one. A tool that is
absent gives no version finding (the plugin's "installed" check already says so). Where smith
can apply a fix it offers one: for Rust, a `rust-toolchain.toml` pinning the version plus
`rustup toolchain install`, which also makes CI and the dev container follow the pin. A
`[toolchains]` pin in `smith.toml` is a separate check, against the same probed version.

## Checks that come from smith itself

Besides each plugin's checks, smith runs checks of its own. These are the ones a workspace
meets most often. Names are as doctor prints them. [Troubleshooting](../guides/troubleshooting.md#what-smith-doctor-reports-and-what-to-do)
lists every check with its remedy.

- `CI pipelines match the workspace`: the generated pipelines are current. The fix is
  `smith ci generate`.
- `The MonoSmith agent files are installed and current`: `.claude/agents/smith.md`,
  `.mcp.json` and `.vscode/mcp.json` match this smith. A warning, with `smith agent install`
  as the remedy; `enabled = false` under `[agent]` in `smith.toml` switches the check off.
- `derived files are up to date`, `derived file <path> (<renderer>) is up to date` and
  `config files that differ from their sibling projects are yours`: only when
  `[derive] enabled = true`. A hand edit to a derived file is reported with the choices
  `smith derive --replace`, `--adopt` or `--eject`; see [`smith derive`](derive.md).
- `release tags resolve under each project's tag prefix` and `release tags are not behind the
  registry`: every project's latest tag is found under its tag prefix, and no version is
  already published beyond it. When the registry cannot be asked, doctor says so and skips
  the comparison. See [Releasing](../reference/releasing.md).
- `the latest CI run on the default branch started its jobs`: a warning, shown only when the
  latest run of the default branch has a GitHub annotation saying a job was not started
  because of billing (failed payments or a spending limit). Such a run has no steps and no
  logs, so the annotation is the only clue. doctor prints it and names the remedy: check the
  GitHub billing page and the spending limit. It only reads (it never changes anything) and
  never fails the local checks. It says nothing without a GitHub remote, a GitHub login
  (`GITHUB_TOKEN`, `GH_TOKEN` or `gh auth login`) or a network; `SMITH_NO_UPDATE_CHECK=1`
  turns it off.
- `the clone has full history`, `tracked files are not generated files`, `every build writes to
  dist/`, `no task caches a work folder other projects share` and `projects depend only on what
  the boundaries allow`.
- `package-lock.json holds the packages every platform needs`: a lockfile made on one OS that
  would fail `npm ci` on another. `smith doctor --fix` regenerates it. If `npm ci` still fails,
  list the missing packages in the root `devDependencies`; when `overrides` names one of them
  (`@emnapi/core`, `@emnapi/runtime`), set the override to `"$<name>"`, or npm stops with
  `Override ... conflicts with direct dependency`.

## A build that writes outside `dist/`

smith builds every project into `dist/<project dir>` (`packages/kit` into `dist/packages/kit`).
A workspace that smith did not create keeps its own scripts, so `smith convert` and `smith adopt`
preserve them, and a package may build into its own `packages/kit/dist`. Doctor reports each such
project on its own line, `kit compiles into dist/packages/kit`, with where it builds now and where
it should.

For a Node project, doctor reads the build script (`--outDir`, `--outdir`, `--out-dir`,
`--outfile`, `--declarationDir`) and the tsconfig the script names (else `tsconfig.build.json`,
then `tsconfig.json`: `outDir`, `declarationDir`, `tsBuildInfoFile`, `outFile`). It also compares
the build task's cached outputs with the folder the script writes: a task that caches
`dist/packages/kit` while the script writes `packages/kit/dist` restores nothing from the cache.

Each project carries a fix that asks first. It never runs with `--yes` alone, and it never
commits. The question has two answers: move the build, or leave the project as it is. After you
choose "move", the review shows every changed line before anything is written:

- the build script flags and the tsconfig options move to `dist/<project dir>`; where the tsconfig
  `outDir` was the only place that named the folder, the script passes `--outDir` instead
- `main`, `types`, `module`, `typings`, `browser`, `bin` and every `exports` target that pointed
  into the old folder follow it; a `files` list that names only the old folder goes, because the
  release publishes the whole build folder
- every script that names the old folder as a path (`start`, `clean`, a test that loads the build)
  follows it
- when `smith.toml` gave the build task the old folder as `outputs`, a `smith config unset
  projects.<name>.tasks.build.outputs` step removes it, so the plugin's own outputs apply again

The old folder stays on disk, and ignore rules for it stay. Nothing is committed unless you pass
`--commit` (or set `commit_operations = true`); a committed fix is undone with `smith undo`.

Doctor refuses the move, and says why, when it cannot be made safe. The finding then becomes a
warning, not a failure, because the build still works: it names the reason and the manual steps, and
doctor stays green. It stays a failure only when something is broken now, for example a build task
that caches a folder the build never writes. The answer "leave it as it is" ends at once with
"Left as it is.": no review, no question.

- another project imports this one by name and its `main`, `types` or `exports` lead into the
  build folder. Node and TypeScript resolve those through the link in `node_modules`, where a
  path such as `../../dist/packages/kit` is not found, so the importers would stop compiling
- a published package (kind `package`, not private) whose tarball needs a file outside the build
  folder: a `files` entry such as `templates`, or a `bin` that is a source file. A release packs
  only the staged build folder, so that file would be missing
- the build writes a second folder that is not inside the first, or `smith.toml` gives the task an
  output that is not the build folder

### The other languages

A project that smith made builds into the central `dist/` already: its Rust crates through
`.cargo/config.toml`, its .NET projects through `Directory.Build.props`, its Maven and Gradle
modules through the build directory of their scripts, its CMake projects through `-B`, and Go,
Python and Dart through the output option of the build command. A project that was adopted,
converted or written by hand can build into the folder its tool picks (`target/`, `bin/` and
`obj/`, `build/`). Doctor reads the setting where it can be read reliably, names where the project
builds now and where it should, and offers the same fix as for Node: a question (move, or leave it
as it is), a review of every changed line, no change on `--yes` alone, `--commit` to commit it and
`smith undo` to revert it. The old folder is never deleted. It stays on disk, unused and
git-ignored.

| Language | Where smith builds | What doctor reads | What the fix writes |
|---|---|---|---|
| Rust | `dist/cargo`, one target folder | `build.target-dir` of the nearest `.cargo/config.toml` (also `.cargo/config`), going up from the crate; else the default `target/`; `CARGO_TARGET_DIR` of the environment wins | `.cargo/config.toml`: sets `[build] target-dir = "dist/cargo"`, every other line stays (a new file when there is none). One finding for the workspace, one for a crate that is its own workspace |
| .NET | `dist/dotnet` | `ArtifactsPath` of the nearest `Directory.Build.props` (MSBuild reads the first one it finds going up); else `bin/` and `obj/` of each project; `OutputPath`, `BaseOutputPath` of the project files | `Directory.Build.props`: adds a `PropertyGroup` with `UseArtifactsOutput` and `ArtifactsPath`, or updates the existing `ArtifactsPath`; the other content stays. A new file when there is none |
| Maven | `dist/maven/<artifactId>` | `<build><directory>` of the module pom, else of the parent poms it names (default `../pom.xml` or its `relativePath`), else `target/` | the module `pom.xml`: sets `<build><directory>${project.basedir}/../../dist/maven/${project.artifactId}</directory>` |
| Gradle (Kotlin and Groovy DSL) | `dist/gradle/<artifactId>` | `layout.buildDirectory` or `buildDir` in the module script, else `build/` | the module script only: adds or replaces `layout.buildDirectory.set(rootProject.layout.projectDirectory.dir("dist/gradle/<id>"))` after the `plugins` block. `settings.gradle(.kts)` and `gradle/libs.versions.toml` are never touched |
| CMake | `dist/<project dir>` | the `-B` of the configure task, and `binaryDir` of each configure preset in `CMakePresets.json` | `CMakePresets.json`: the `binaryDir` values (`${sourceDir}/../../dist/<project dir>`; with several presets `.../${presetName}`). A wrong `-B` in a task written by hand in `smith.toml` is reported with the manual step: smith does not rewrite a command line |
| Go, Python, Dart | `go build -o`, `python -m build --outdir`, `dart compile -o` / `flutter build -o` into `dist/<project dir>` | the build task: a command that passes no output option, or one that names a folder outside `dist/`; a task whose `outputs` do not hold the folder the command writes (a cache hit restores nothing) | nothing: the finding names the option to pass, or the outputs to list, in `[projects.<name>.tasks.build]` of `smith.toml` |

#### Two folders: the work folder and the shipped folder

A build has two folders (ADR 0092). The **shipped folder** `dist/<project dir>` holds what the
project ships: the binary, the jar, the package. The **work folder** `dist/<tool>/...` holds the
tool's intermediate files. Rust and .NET share one work folder among all projects (`dist/cargo`,
`dist/dotnet`). A Maven or Gradle module has its own (`dist/maven/<artifactId>`,
`dist/gradle/<id>`). Go, Python and Dart write straight to the shipped folder and have no work
folder in `dist/`. A Tauri app keeps its cargo target and its web build in `dist/tauri/<project
dir>`, and an Electron app what it runs and packs in `dist/electron/<project dir>`; both ship the
installers or the packed app into `dist/<project dir>`.

Doctor checks both folders:

- **The work folder is the one the build ships from.** A Rust crate links its program from
  `dist/cargo/<profile>`, a .NET app from `dist/dotnet/bin/<project>/<configuration>` and a Maven
  module its jar from `dist/maven/<artifactId>`. A `target-dir`, `ArtifactsPath` or
  `<build><directory>` that names another folder, also one inside `dist/`, leaves the shipped
  folder empty. Doctor reports it and offers the same move as above.
- **The build task caches the shipped folder, and never a shared work folder.** A cache restore
  empties each output folder first. If a crate's build listed `dist/cargo` as an output, a restore
  would delete the build state of every other crate. Doctor fails the check `<project> caches only
  dist/<project dir>` when a build task lists `dist/cargo` or `dist/dotnet` (or a folder in them,
  or `dist/` itself) as an output. It also reports a build task that ships files but does not list
  `dist/<project dir>`: a failure when the task is cached, else a warning. A module's own work
  folder (`dist/maven/<artifactId>`) can stay an output, because no other project builds there.
- **Across the workspace**, the check `no task caches a work folder other projects share` fails when
  a task of any plugin lists a folder that another project ships from.
- **An adopted Electron app packs in its work folder.** The check `<project> packs into
  dist/electron/<project dir>` (Electron Forge: `.../out`) is a warning when Forge's `outDir` or
  electron-builder's `directories.output` names another folder, such as the default `out/` or
  `dist/` in the project. The packing tasks ship the app into `dist/<project dir>` in both cases.
  Doctor offers the move as a choice (move, or leave as it is) when it can edit the file safely:
  `package.json`, `electron-builder.json`, an `electron-builder.yml` without a `directories` key,
  or a Forge configuration file with one `outDir` string or one plain exported object. Otherwise
  the remedy says what to set by hand. See [the node plugin](../plugins/using/catalogue/node.md#adopted-electron-setups).
- **A Tauri app takes its web build from its work folder.** The check `frontendDist of <project>
  is dist/tauri/<project dir>/frontend` is a warning when `tauri.conf.json` names another folder.

These findings name the manual step: list only `dist/<project dir>` in the `outputs` of the task in
`smith.toml`. smith does not rewrite them.

The move is refused, with the reason and the manual step, when it would break something:

- a file names the old folder as a path: a `Dockerfile`, `Makefile`, script or compose file of the
  project, a root script, a pipeline of `.github/workflows`, or `smith.toml` (`COPY target/release/app`,
  `src/App/bin/Release`). Doctor names the file
- another project reads a build of this one by its path: a .NET `HintPath` into `bin/` or `obj/`, a
  Maven pom or Gradle script that names `../kit/build/...`, a CMake file that includes
  `../engine/build/...`
- the project file names the folder itself: a Maven pom with `target/` as a path, a Gradle script
  with `file("build/...")`, a .NET project that sets `OutputPath` or `BaseOutputPath` (they win over
  `ArtifactsPath`)
- the Rust target folder is set by `CARGO_TARGET_DIR` in the environment (smith does not edit your
  environment) or is outside the workspace
- the file to edit is not valid, or is written in a form smith cannot edit without guessing

A published package is not a reason to refuse for these languages: `mvn deploy`, `gradle publish`,
`dotnet pack -o dist/<dir>`, `cargo package --target-dir` and the wheel build all write where smith
says, whatever the project's own build folder is.

What doctor does not check, and says nothing about: a `Cargo.toml` (cargo has no `target-dir` key
there), a `[env]` table or an inline-table `build = { target-dir = ... }` of cargo's configuration;
a `Directory.Build.props` that imports another file or builds `ArtifactsPath` from another
property; a Maven directory that uses a property other than `${project.basedir}`, `${basedir}`,
`${maven.multiModuleProjectDirectory}` and `${project.artifactId}`, or a parent with an unreadable
`relativePath`; a Gradle build directory set for all modules by the root script or in a form that
names no path; a CMake `binaryDir` that uses a variable other than `${sourceDir}`,
`${sourceParentDir}`, `${presetName}`, `${hostSystemName}` and `${generator}`, an `include` of
another preset file, `CMakeUserPresets.json`, and a `CMAKE_BINARY_DIR` set in `CMakeLists.txt`;
the folders a Python build backend makes inside the project (`build/`, `*.egg-info`); the
`build/` of `flutter run`; `.cargo`, `bin/` and `obj/` of a project that was never built. Nothing
is guessed: where the setting cannot be read, doctor reports nothing.

## Tools the workspace needs

A program that a language needs (Python, the .NET SDK, Go, Docker) and that is not installed is a
warning, worded "<tool> is installed". Its fix is a question with the same options as
[`smith tools install`](tools-install.md): install with a package manager, use the official
page, install it yourself, or skip it. A command that needs administrator rights is shown first
and is never run unless you pick "Let smith run it"; `--fix --yes` never picks it.

## Suggestions about the defaults

Two warnings read the per-project task overrides in `smith.toml` ([ADR 0056](../architecture/decisions/0056-repeated-overrides-say-what-the-default-should-be.md)).
They change nothing:

- `repeated project overrides are a workspace default`: three projects or more, and 80 percent of
  all projects, set the same value for the same task setting
  (`[projects.<name>.tasks.test] cache = false`). The remedy is to set it once under
  `[tasks.test]`. A smaller set of projects is folded into a `[[groups]]` entry by
  `task settings shared by projects are written once`.
- `a workspace default is not overridden by nearly every project`: a `[tasks.<task>]` setting that
  at least three projects, and 80 percent of all projects, set to another value. The default is
  probably wrong.

`smith.toml uses current key names` fires when `smith.toml` still has a key smith renamed
(`[[task_groups]]`, now `[[groups]]`, [ADR 0067](../architecture/decisions/0067-groups-replace-task-groups-and-also-set-tags-and-release.md)).
Every other smith command stops with one sentence that names `smith doctor --fix`; doctor is the one
that still runs. `--fix` shows the rename as a diff, keeps comments and every other line, and then
checks the workspace as usual.

`the workspace uses the latest defaults` compares the `[workspace] defaults` vintage with the
latest one this smith knows; `smith upgrade --defaults` moves it ([ADR 0054](../architecture/decisions/0054-defaults-have-vintages.md)).

## The packages that attach to GitHub Releases

A workspace that uses `registry = "github-release"` gets these checks:

- `release: github_releases is on ...`: fails when `[release] github_releases` is off, because the package files
  have no release to be attached to.
- `release: <name> can be packed`: fails when the plugin of the package cannot pack it. The release refuses to
  run, so nothing is published to a public registry.
- `release: <name> has the credential its public registry asks for`: a warning for a package that would publish
  to a public registry (npm, PyPI, NuGet, RubyGems, crates.io) whose credential variable is unset. The remedy
  offers `registry = "github-release"` or `"none"` for it.
- A package with `registry = "none"` or `"github-release"` is not asked about its registry by the lineage check
  (`release tags are not behind the registry`): its tags and release assets are the record.

## Suggestions about a project's kind

Two warnings point at [`smith kind`](kind.md). They are suggestions, never failures, and they
say nothing when there is nothing to say:

- `the lib <name> is used by a project`: no project imports the lib, and it has been in the
  repository for 14 days or more. A lib is shared inside the repository; this one is shared
  with nobody. A lib added today, or in a workspace of one project, is left alone. The remedy
  offers `smith kind <name> app`, `smith kind <name> package` or `smith remove <name>`, and
  says to leave it if something outside the graph uses it.
- `the package <name> is published`: a package that its own settings (`registry = "none"`
  under `[projects.<name>]`) or its manifest (Dart's `publish_to: none`) never publish, while
  the workspace does publish. The remedy offers `smith kind <name> lib`. A workspace that
  releases everything by tag (`registry = "none"` under `[release]`) is not asked.

## When to use it

- After cloning, after a merge, after changing tools, and before a release.
- In CI, as a gate: it exits 1 while any check fails.
- With `--fix` to repair generated files.
- A check whose remedy names a smith command (`smith upgrade --defaults`, `smith deps up --align`,
  `smith derive`, `smith cache prune`, `smith kind`) carries a fix of that command. It is never
  "safe": `smith doctor --fix --yes` skips it, and the interactive shell and the editors ask first,
  show the plan the command prints, and run it in place with a Fix action. Name one with
  `--only` and answer its questions with `--choose`.

## Synopsis

```sh
smith doctor [--verbose] [--json]
smith doctor --fix [--yes] [--commit] [--only <check or fix>]... [--choose <n>,<n>]
```

## Arguments and flags

<!-- smith:begin doctor -->
Check the workspace's health; every problem names its remedy

Usage: `smith doctor [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--choose <choose>` | integers, comma-separated |  | with --fix and --only, answer the fix's questions by option number, in order (1,2) |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--fix` | boolean | off | show the fixes doctor can make as diffs, and apply the safe ones |
| `--json` | boolean | off | print the report as JSON |
| `--only <only>` | string, repeatable |  | with --fix, fix only the named check, as doctor prints it (repeatable) |
| `--strict` | boolean | off | exit non-zero on a warning too, for a pipeline that allows none |
| `--verbose` | boolean | off | list every passed check, not one line for each check that repeats per project |
| `-y, --yes` | boolean | off | with --fix, apply without asking |

<!-- smith:end doctor -->

Interactions:

- `--only` and `--choose` pick what to fix, so they need `--fix` (`--only and --choose pick
  what to fix, so they go with --fix`).
- `--fix` alone shows the fixes as diffs and, in a terminal, asks. Without a terminal, or
  without `--yes`, it prints the summary first, before the checks (`3 fixes available: 2 applied by --yes,
  1 needs a choice`, which fixes are safe, which need your decision and the flags that apply them). More
  than eight passed lines are folded into one count (`--verbose` lists them). Then come the checks that
  failed, each fix with its diff or the command it would run, `Pass --yes to apply the N safe fix(es)
  above.`, and the count again as the last line. The exit code is 1. A fix that writes a file and then
  runs smith (pin a toolchain, then `smith ci generate`) shows the plan of that command as if the file
  were already written. A fix that asks (for
  example whether a Cargo crate belongs to the workspace) is never applied by `--yes` alone: it is listed
  with `--only "<name>" --yes --choose N`, one line per answer.
- A tool doctor needs and cannot find (such as `cargo-audit`) is one finding that says it is skipped.
  Doctor never prints the tool's own error lines.
- A Rust crate that has its own `[workspace]` table is its own root, so doctor does not ask for it in
  `members`. A crate it finds that is neither listed nor excluded is offered two answers: add it to
  `members`, or "not part of this workspace" (adds it to `[workspace] exclude`).
- A workspace that records no `[workspace] defaults` vintage is offered two answers: keep the current
  defaults (writes `defaults = "<vintage>"` to `smith.toml` and changes nothing else), or move to the
  latest (`smith upgrade --defaults`, which shows a plan).
- `--only "<name>" --yes` applies one fix by the name doctor prints (the check's or the
  fix's). Doctor exits non-zero while a named fix is not applied.
- `--choose 1,2` answers the fix's questions by option number, in order, for the fix `--only`
  names.
- Fix steps run in the workspace root and stop at the first that fails; the failing step's
  output is shown.
- `--verbose` lists every passed check instead of one line for each check that repeats per
  project.
- `--json` prints `findings` (each with `check`, `ok`, `detail`, `remedy`, `warning`,
  `source`), `passed`, `failed` and `warnings`. For a passing check `detail` and `remedy`
  describe the failure that would be reported.

A check is `✓` (passed), `×` (failed, exit 1) or `!` (a warning, not counted as a failure).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No check failed (warnings allowed). |
| 1 | A check failed, or `--fix` could not apply what was asked. |

## Files and environment

Reads the workspace and runs the plugins' checks (they may call `go`, `node`, `python`,
`git`). `--fix` rewrites generated files (pipelines, `go.work`, the dev container) and runs
the commands it displays. Variables that change checks: `SMITH_CACHE_REMOTE` and
`SMITH_CACHE_KEY` (the remote cache must be signed), `SMITH_UPDATE_CHECK` (a newer smith is named at the end), and the
release credentials ([`smith release`](release.md)).

## Examples

A failing workspace: the pipeline file is stale and `go.work` was deleted.

```sh
smith doctor
```

```text
× CI pipelines match the workspace
    .github/workflows/ci.yml is out of date
    → run `smith ci generate` and commit the result
× go.work exists
    the workspace has Go projects but no go.work
    → run `go work init` then `go work use` for each Go project
✓ the cache directory is ignored by git
✓ the repository has commits
...
✓ Node projects are checked for vertical slices (@monosmith/eslint-config)

25 passed, 2 failed, 0 warnings
1 fix(es) available: `smith doctor --fix`
```

The fix, shown as a diff, then applied:

```sh
smith doctor --fix
```

```text
...
● regenerate the pipelines, keeping your slot contents
--- a/.github/workflows/ci.yml
+++ b/.github/workflows/ci.yml
@@ -25,7 +25,7 @@
...
-
-# stale

Pass --yes to apply the 1 safe fix(es) above.
```

```sh
smith doctor --fix --yes
```

```text
...
✓ regenerate the pipelines, keeping your slot contents

after fixing: 26 passed, 1 failed, 0 warnings
```

Read the diff before applying: in this trial `go.work` was missing, and the regenerated
pipeline fell back to `go-version: '1.24'` where it had `'1.27'` (the version came from
`go.work`). Restoring `go.work` first (`go work init ./apps/api ./apps/web2`) and then
running `smith ci generate` put `1.27` back.

A repository adopted a moment ago, in a non-interactive run, names the command that applies
a fix that asks:

```text
● install the Node dependencies
  step 1: install the workspace's dependencies
          $ npm install --no-audit --no-fund

not applied: install the Node dependencies needs your decision. Run one of:
  smith doctor --fix --only "install the Node dependencies" --yes
```

## External plugins

For every external plugin `smith.toml` declares, doctor shows where it comes from, whether it is
trusted and pinned, and the environment variables it may see (names, never values):

```text
✓ external plugin widget is trusted (from .smith/plugins, pinned no, sees the baseline plus GITHUB_TOKEN, NPM_TOKEN)
```

An external plugin smith would refuse to start fails the check, and because no plugin starts until
it is trusted, doctor shows only the plugin checks and exits non-zero until you trust it:

```text
× external plugin widget is trusted
    the external plugin "widget" is not trusted yet, so smith did not start it (from .smith/plugins, pinned no, sees the baseline plus GITHUB_TOKEN, NPM_TOKEN)
    → run `smith plugin trust widget` to review it and trust it
    → run `smith plugin list` to see every external plugin and what it may see

0 passed, 1 failed; the other checks run once these plugins are trusted
```

A plugin installed on the search path but not declared in `smith.toml` is not needed by the
workspace and is not checked. See [Plugin security](../plugins/using/plugin-security.md).

## Common errors

| Message | Cause and remedy |
|---|---|
| `external plugin <name> is trusted` fails | smith will not start an external plugin you have not trusted (or that changed since you did). Run `smith plugin trust <name>`; in CI, `smith plugin trust <name> --pin` on your machine and commit the `sha256`. |
| `--only and --choose pick what to fix, so they go with --fix` | Add `--fix`. |
| `not applied: <fix> needs your decision. Run one of: smith doctor --fix --only "<fix>" --yes` | A fix with a question or a command is not applied without a terminal; run the command it prints. |
| Check-specific failures | Each prints its own remedy; the names are stable, for example `go.work exists`, `Node dependencies are installed`, `CI pipelines match the workspace`, `The MonoSmith agent files are installed and current`, `tracked files are not generated files`, `Jest helper files of an older convert are replaced by @monosmith/jest-preset`. |

## Related

[`smith ci generate`](ci-generate.md), [`smith upgrade`](upgrade.md),
[Troubleshooting](../guides/troubleshooting.md), [`smith explain`](explain.md).
