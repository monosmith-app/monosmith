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

## capability

One of the optional methods a [plugin](#plugin) announces when it starts
(`discover`, `generate`, `publish`, …). smith never calls a method the plugin
did not announce. See [Writing a plugin](../guides/writing-a-plugin.md).

## convention

A [plugin](#plugin) applied to every project of some languages at once, turned
on in `smith.toml` with `[conventions.<name>]`. `vertical-slices` and `docker`
are the built-in ones. A convention adds tasks, CI steps and doctor checks; it
does not replace a project's language. See [Conventions](conventions.md).

## CI phase

One step of the generated pipeline. Every phase is a command you can run
yourself: `smith ci setup`, `audit`, `verify`, `pack`, `native` and `release`.
See [CI pipelines](ci.md).

## depth level

How far `smith adopt` and `smith convert` go (`--level`). `conservative`, the
default, keeps the repository's own build and lint setup and lists what is left.
`full` makes smith the only tool and deletes the old tooling nothing uses any
more, after asking. See [Converting](../getting-started/converting.md#going-further---level-full).

## dist

The one directory at the workspace root where every build writes, in a path
that mirrors the project's (`apps/web` builds into `dist/apps/web`). Delete
`dist/` to clean everything. `smith doctor` fails a task whose output goes
anywhere else.

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

## lib

One of the three [kinds](#kind). A lib is imported by other projects and never
published by itself. It lives in `libs/<name>`.

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

## toolchain

Two meanings. The **toolchain role** is described under
[role](#role). A **toolchain** in CI is a runtime the
pipeline must install (Node.js 24, Go, Python 3.12); a plugin says which and
how. See [CI pipelines](ci.md#how-it-is-built).

## workspace

Your repository as smith sees it: a directory with a `smith.toml` at its root.
Every command looks upward from where you run it until it finds one.
`SMITH_WORKSPACE_ROOT` overrides that. See [Configuration](configuration.md).
