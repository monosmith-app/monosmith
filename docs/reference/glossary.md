# Glossary

The words smith uses, in alphabetical order. Each entry says what the thing is
and where to read more.

## affected

A project is **affected** by a change when it owns a changed file, or when it
depends (directly or not) on a project that does. `smith affected` lists them;
`--affected` on `smith run`, `smith build` and the like runs tasks only on
them. smith counts the changes since the merge base with a base branch, plus
what you have not committed. When it cannot tell (no base branch, a shallow
clone, a changed `smith.toml` or lockfile) every project is affected. More in
[Tasks, cache and affected](tasks-and-cache.md#affected).

## app

One of the three [kinds](#kind). An app is deployed or run and never published
to a package registry. It lives in `apps/<name>`. A Go app can still release
per-platform binaries when `smith.toml` gives it a release section. See
[Releasing](releasing.md#apps).

## cache key

The fingerprint of a task run. smith computes a SHA-256 over the task's
command, the content of its input files, the sources of the projects it
depends on, the keys of the tasks it depends on, the operating system, the
smith version, the version of every plugin that shapes the project and the
version of the compiler or runtime. Same key, same result: smith replays the
stored output and log instead of running the task. `smith why-miss` says which
part of the key changed. See [Cache](tasks-and-cache.md#cache).

## axis

The question a [choice](#choice) answers: `framework`, `test`, `lint` or `style`. A project has at
most one choice on an axis unless its choices allow more. `smith choices` lists the axes. See
[Stacks](stacks.md#choices-a-stack-as-a-saved-selection).

## capability

One of the optional methods a [plugin](#plugin) announces when it starts
(`discover`, `generate`, `publish`, …). smith never calls a method the plugin
did not announce. See [Writing a plugin](../guides/writing-a-plugin.md).

## choice

One option on an [axis](#axis) of a project's make-up: React or Vue on the framework axis, Jest or
Vitest on the test axis. A plugin ships its choices as data. The solver checks a selection (`requires`,
`conflicts`, what an axis needs) and turns it into a [stack](#stack). `smith choices` lists them, and
`smith compose` searches them. `smith swap` changes one choice of an existing project.

## convention

A [plugin](#plugin) applied to every project of some languages at once, turned
on in `smith.toml` with `[conventions.<name>]`. `vertical-slices` and `docker`
are the built-in ones. A convention adds tasks, CI steps and doctor checks; it
does not replace a project's language. See [Conventions](conventions.md).

## CI phase

One step of the generated pipeline. Every phase is a command you can run
yourself: `smith ci setup`, `audit`, `verify`, `pack`, `native` and `release`.
See [CI pipelines](ci.md).

## declared project

A project that `smith.toml` names with `[projects.<name>]`, a `root` and a `language`, because its
folder has no manifest of its own (a Go program that is a package of the module above it). Only a
language plugin with the `declared` capability can describe one. See
[Configuration](configuration.md).

## depth level

How far `smith adopt` and `smith convert` go (`--level`). `conservative`, the
default, keeps the repository's own build and lint setup and lists what is left.
`full` makes smith the only tool and deletes the old tooling nothing uses any
more, after asking. See [Converting](../getting-started/converting.md#going-further---level-full).

## divergence

A place where a workspace differs from smith's defaults: a task override, an ejected file, a derived
file that drifted, a file that differs from its siblings. `smith divergences` lists them, each with
the reason and the way back. See [`smith divergences`](../manual/divergences.md).

## dist

The one directory at the workspace root where every build writes, in a path
that mirrors the project's (`apps/web` builds into `dist/apps/web`). Delete
`dist/` to clean everything. `smith doctor` fails a task whose output goes
anywhere else.

## framework

How an app or a library is written: a web framework such as React or Vue, or a whole starter. It is a
choice on the framework layer of a project (see [`smith compose`](../manual/compose.md)). A terminal form
shows it as `framework = how the app or library is written`.

## group

A `[[groups]]` table in `smith.toml`: a selector (names, folder globs, `kind:`, `lang:`, `tag:`) and
what it gives every project it selects: task overrides, tags and release settings. Groups apply in
file order. They replace the `[[task_groups]]` table of smith 1.x. See
[Configuration](configuration.md#groups).

## host

Where code runs in the cloud: an Azure Functions, AWS Lambda or Google Cloud Functions app, with the trigger
and the runtime it asks for. It is a choice on the host layer of a project (see
[`smith compose`](../manual/compose.md)).

## fix

What `smith doctor --fix` applies for a failed check: new file contents, commands
to run (steps) and questions to ask (choices). A **safe** fix only restores what
smith or a plugin owns and is applied after one confirmation; any other, and any
with a question, is asked about. See [`smith doctor`](../manual/doctor.md).

## kind

What a project is for: `app`, `lib` or `package`. The kind decides where the
project lives and what `smith release` does with it. Do not confuse it with
the combined names `smith kinds` lists (`node-app`, `go-lib`, …), which name a
language and a kind together.

| Kind | Directory | Release |
|---|---|---|
| `app` | `apps/<name>` | never published; deployed by your pipeline |
| `lib` | `libs/<name>` | never published; imported by other projects, and released inside the packages that use it |
| `package` | `packages/<name>` | published to its registry when it has releasable commits |

## language plugin

The [plugin](#plugin) that understands one language: it finds the projects
from their manifests (`package.json`, `go.mod`, …), turns imports into graph
edges and writes versions. A project has exactly one. See [Roles](#role).

## launcher

The two static files `tools/smith/smith` and `tools/smith/smith.cmd`, written by `smith bootstrap`.
They run the smith version the workspace pins in `[ci] smith_version`, so a clone needs no install.
See [`smith bootstrap`](../manual/bootstrap.md).

## lib

One of the three [kinds](#kind). A lib is imported by other projects and never
published by itself. It lives in `libs/<name>`.

## library of the workspace

A project of kind `lib` or `package` that another project of the same workspace and language depends
on. `smith install -w <project> <library>` writes that dependency in the manifest, through the
plugin's `link` capability, instead of asking the registry. See [`smith install`](../manual/install.md).

## native project

A project that must be built on each operating system instead of cross-compiled
(a Go app using cgo, a Node project with per-platform builds). `smith ci native`
builds it in a job per OS. See [CI pipelines](ci.md#native-projects).

## package

One of the three [kinds](#kind). A package is published to a registry (npm,
PyPI, crates.io, NuGet, …) when its commits call for a release. It lives in
`packages/<name>`. (The word is also used, as everywhere, for a dependency you
install.)

## plugin

A program that teaches smith something: a language, a test runner, a layout
rule, a deploy target. smith's core knows no language; every plugin runs as a
separate process and speaks a small JSON-RPC protocol on stdin and stdout. The
built-in ones ship inside the `smith` binary. You can write one in any
language. See [Plugins](plugins.md) and
[Writing a plugin](../guides/writing-a-plugin.md).

## project

A directory with a manifest a language plugin recognises (`package.json`,
`go.mod`, `pyproject.toml`, `Cargo.toml`, a `*.csproj`, …). Every project has a
name, a [kind](#kind), a language, [tasks](#task) and the other projects it
imports. smith finds them; you do not list them anywhere. `smith projects`
prints them.

## recipe

A TOML file describing which plugin fills each [role](#role) of a
project and with what settings, plus a few questions to ask when someone adds
a project. It is data, never code. A team writes one so every new project is
made the same way. See [Recipes](recipes.md).

## release branch

The branch real releases run from (`[release] branch`, default `main`).
`smith release` refuses to publish from any other.

## release channel

A branch that releases its own line of prerelease versions (`2.0.0-beta.1`), named in
`[[release.channels]]`. A package of a channel never takes the place of the stable version, and only a
language plugin with the `channels` capability can publish it. See
[Releasing](releasing.md#release-channels).

## role

The part a [plugin](#plugin) plays in a project. The word was once "level" as
well; "role" is the one word now (`smith plugin list` shows ROLES, `smith recipe new`
takes `--fixed role=plugin`, the protocol field is `roles`). The recipe file
still writes the table `[levels.<role>]` and the protocol still writes `level` in
`requires`: those are names in a format, kept for the files that exist.

| Role | Does | A project has |
|---|---|---|
| `language` | finds projects, writes versions, publishes | exactly one |
| `scaffold` | writes the starter files of a new project | at most one |
| `toolchain` | decides how a project builds, tests and lints | any number, one per [slot](#slot) |
| `structure` | checks layout and code rules | any number |
| `deploy` | adds pipeline steps and deploy targets | any number |

## SBOM

The CycloneDX software bill of materials `smith release` attaches to the GitHub
Release of an app, listing its dependencies. See
[Releasing](releasing.md#sbom-and-build-provenance).

## shell

What wraps an app as a program you install on a computer: Tauri or Electron. It is a choice on the shell layer
of a project (see [`smith compose`](../manual/compose.md)).

## slice

A folder that holds one outcome of a program, with flat files named for the
role they play (`greet.use-case.ts`, `greet_use_case.go`). The
`vertical-slices` [convention](#convention) checks this layout. See
[Conventions](conventions.md#vertical-slices).

## slot

Two unrelated things carry this name, so check which one you mean.

- A **toolchain slot** is a task name a toolchain plugin owns (`test`, `lint`,
  `build`). A project can have only one plugin per slot, so a stack cannot hold
  both `jest` and `vitest`. A plugin lists its slots in `provides`.
- A **pipeline slot** is a pair of `smith:slot` comments in a generated CI file
  (`before-setup`, `after-verify`). Steps you put between the markers survive
  `smith ci generate`. See [Your own steps](ci.md#your-own-steps).

## stack

The set of plugins one project is built with: a language plugin plus
toolchain, structure, scaffold and deploy plugins on top. Each project records
its own, so one workspace can hold a React app and a plain Go library.
`smith stacks` lists them; `smith add app web --stack react` picks one. See
[Stacks](stacks.md).

## task

A command a plugin works out for a project: `build`, `test`, `lint`,
`typecheck`, `start`, and any npm script. A task has a command, the files it
reads (inputs), the directories it writes (outputs) and the tasks it waits for
(`dependsOn`). smith runs tasks in dependency order and caches the result. See
[Tasks](tasks-and-cache.md#tasks).

## tool

A program on the machine that a plugin needs to build, test or run its projects: Python, the .NET SDK,
Docker. The plugin declares it as data. `smith tools` lists the tools and `smith tools install` offers
the ways to get what is missing. Not to be confused with a [toolchain](#toolchain). See
[`smith tools`](../manual/tools.md).

## toolchain

Two meanings. The **toolchain role** is described under
[role](#role). A **toolchain** in CI is a runtime the
pipeline must install (Node.js 24, Go, Python 3.12); a plugin says which and
how. See [CI pipelines](ci.md#how-it-is-built).

## vintage

The `[workspace] defaults = "2026.11"` value: the version of smith's defaults a workspace follows. A
newer smith keeps them, `smith doctor` warns when the workspace is behind, and
`smith upgrade --defaults` moves it with a plan. See [Configuration](configuration.md).

## workspace

Your repository as smith sees it: a directory with a `smith.toml` at its root.
Every command looks upward from where you run it until it finds one.
`SMITH_WORKSPACE_ROOT` overrides that. See [Configuration](configuration.md).
