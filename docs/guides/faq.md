# FAQ

Short answers, with a link to the page that has the detail. For problems, see
[troubleshooting](troubleshooting.md).

## Using smith

### What do I have to write to make smith work?

A `smith.toml` that names the workspace, and nothing more. smith finds
projects from their manifests (`package.json`, `go.mod`, `pyproject.toml`,
`Cargo.toml`, a `*.csproj`, ...), learns the dependency graph from what the
sources import, and infers the tasks. `smith.toml` is for the things smith
cannot guess: overrides, a remote cache, a private registry, release settings.
See [configuration](../reference/configuration.md).

### Do I have to use `smith add`?

No. A folder with a manifest smith recognises is a project, however you made it.
`smith add` is the easy path because it also wires the toolchain files (the
`go.work` entry, the npm workspace), writes a starter example, and regenerates
CI. If you add a project by hand, `smith doctor` tells you what is missing.

### Which languages are supported?

Node (JavaScript and TypeScript), Python, Go, Rust, .NET (C#, F#), Ruby, PHP,
C and C++ (CMake), Dart and Flutter, the JVM (Java with Maven, Kotlin with
Gradle, Android) and Bicep, each as a built-in plugin. Anything else is a
[plugin you write](writing-a-plugin.md), in any language.

### What is the difference between an app, a lib and a package?

Their **kind** says what release does with them. An `app` is deployed and never
published. A `lib` is imported by other projects and never published itself; it
ships inside the packages that use it. A `package` is published to its
registry. See the [glossary](../reference/glossary.md#kind).

### What is a stack?

The set of plugins one project is built with: its language plus a test runner,
a linter, a scaffold, and so on. `smith stacks` lists them. Two Node projects
in one workspace can use different stacks. See [stacks](../reference/stacks.md).

### What is a recipe, and do I need one?

A recipe is a file that records your team's choice of plugins and settings, with
a few questions asked when someone adds a project. You do not need one. It is
for teams that want every new project made the same way. See
[recipes](../reference/recipes.md).

### How do I run one project, or only what changed?

```sh
smith test -p web          # web and what it depends on
smith test --affected      # only what a change since main can break
```

### How do I run an app?

```sh
smith run start -p web     # one app
smith dev web api          # several together; Ctrl+C stops all of them
smith dev --all            # every project that has a start task
```

### Do I have to remember all these commands?

Run `smith` with no arguments in a terminal. The interactive shell shows the
workspace's health and the one thing to do next, runs tasks with single keys,
and shows the `smith` command line behind every action. See
[the terminal interface](../manual/tui.md); every command is in the
[manual](../manual/README.md).

### Can I run a task that is not `build`, `test`, `lint` or `typecheck`?

Yes: `smith run <task> [-p <projects>]`. A task is whatever a plugin inferred
(including every npm script) or what you declare under `[projects.<name>.tasks]`
in `smith.toml`. `smith explain <project>` lists them.

### How do I run an end-to-end suite against a running app?

Give the suite's task a `serve` setting, and smith starts the app before it and
stops it after. See
[Serving an app for a task](../reference/tasks-and-cache.md#serving-an-app-for-a-task).

### How do I leave a folder out?

Add it to `exclude` under `[workspace]` in `smith.toml`:

```toml
[workspace]
exclude = ["examples/*", "tools/scratch"]
```

Directories listed there have manifests that are not projects.

### How do I switch a plugin off?

`smith plugin disable <name>`, or `smith plugin set a b c` to enable exactly
these. They edit `plugins.enabled` in `smith.toml`. `smith plugin list` shows
what is installed.

### Does smith work offline?

Yes, once the toolchains and dependencies are installed. The local cache is a
folder (`.smith/cache`). Commands that ask a registry (`smith up`, `smith info`'s
latest-release line, releasing) need the network.

### Where does smith put build output?

Everything builds into `dist/` at the workspace root, in a path that mirrors the
project's (`apps/web` builds into `dist/apps/web`). Delete `dist/` to clean.
`smith doctor` fails a task that writes elsewhere.

## The cache

### What is cached?

A task's output directories, its log and its exit status, under a key made from
the command, the input files, the projects it depends on, the OS, and the
versions of smith, its plugins and your compiler. See
[tasks, cache and affected](../reference/tasks-and-cache.md#cache).

### Why did a task run again?

`smith why-miss <project>:<task>` says. See
[troubleshooting](troubleshooting.md#a-task-ran-again-and-i-expected-the-cache).

### Can my team share the cache?

Yes. Set `[cache] remote` to an HTTP server, an S3 bucket or an Azure Blob
container, and share the signing key as `SMITH_CACHE_KEY`. CI writes, developers
read. Without the key nothing is shared. See
[the remote cache](../reference/tasks-and-cache.md#remote-cache).

### Is it safe to delete `.smith/`?

The cache lives in `.smith/cache`; deleting it only costs you the cached results.
`.smith/` also holds plugins you installed there (`.smith/plugins`), so delete
`.smith/cache` rather than the whole folder. Keep `.smith/` out of git
(`smith doctor` checks that).

## CI

### Do I write the pipeline?

No. `smith new` and `smith add` generate it for GitHub Actions, Azure Pipelines
or GitLab CI (`[ci] providers`), and `smith ci generate` regenerates it. Put your
own steps between the `smith:slot` markers and they are kept. See
[CI pipelines](../reference/ci.md).

### Can I use my own pipeline?

Yes, with a caveat: `smith doctor` fails when a generated pipeline is out of date,
and `smith ci generate` rewrites the file. Keep your own steps between the
`smith:slot` markers, or call the phases from a pipeline you own: `smith ci setup`,
`audit`, `verify`, `pack`, `native` and `release` are ordinary commands.
`smith adopt` keeps a pipeline of yours as `<file>.pre-smith`.

### Why does CI install smith from a script?

The generated pipeline downloads the `smith` binary by running `tools/smith/install.sh`,
a copy of the installer that `smith ci generate` commits into your project (never edit
it; `smith doctor --fix` refreshes it), pinned to the version that generated the pipeline (a development build writes
`SMITH_VERSION: latest`). Change that value to pin another version. See
[installing](../getting-started/installing.md).

### How does CI cache package downloads?

Each generated pipeline has a cache step per package manager in use (npm, pip,
Go modules, cargo, NuGet, Composer, ...), keyed on the lockfiles. See
[CI pipelines](../reference/ci.md#package-download-stores).

## Releasing

### How does smith pick a version?

From Conventional Commits since the project's last tag: `feat` is a minor bump,
`fix` and `perf` a patch, a breaking change a major (a minor before 1.0). `docs`,
`chore`, `test`, `ci`, `refactor`, `build` and `style` release nothing. See
[releasing](../reference/releasing.md).

### Where does it release to?

Packages go to their language's public registry (npm, PyPI, crates.io, NuGet,
RubyGems, a Maven repository, Go module tags), or to a private one named in
`smith.toml`. Apps can release per-platform binaries to GitHub Releases, and
container images for apps with a Dockerfile (the `docker` convention) get the
version, `latest`, the commit and the branch as tags. `smith release --dry-run`
shows the plan.

### Can I try a release without publishing?

`smith release --dry-run`. It changes nothing.

### What if a release fails half way?

Run it again. A version a registry already has is skipped, so a re-run finishes
the job.

### What tag does smith create?

`<project>@<version>` for each released project, for example `sdk@0.1.0`
(`tag_prefix` changes the prefix). A Go package is tagged `<dir>/v<version>`
(for example `packages/gokit/v0.1.0`), which is how `go get` finds a nested
module's version. See [releasing](../reference/releasing.md#tags).

## Installing and upgrading

### Can I `go install` smith?

No, by decision: smith is built from the repository's `go.work`, which
`go install` does not read. Use the install script, a package manager or a
download. See [installing](../getting-started/installing.md).

### How do I upgrade smith?

Install the newer one (rerun the install script, or use your package manager),
then run `smith upgrade` in each workspace. See
[getting started](../getting-started/first-workspace.md#10-upgrade-smith).

### Is there an editor or AI assistant integration?

Every command has a stable `--json` output, `smith schema` prints the JSON
Schema of `smith.toml`, and `smith mcp` serves the workspace to AI assistants
over the Model Context Protocol. See [editors](editors.md).

## Coming from somewhere else

### I have an Nx or Turborepo workspace.

`smith convert` turns it into a smith workspace, keeping its projects and
release tags, and reports what it cannot map. `--level full` also removes the
old tooling nothing uses any more. See [converting](../getting-started/converting.md).

### I have a repository with no monorepo tool.

`smith adopt` reports what bringing it under smith involves, then takes the
steps. See [adopting an existing repository](../getting-started/adopting.md).

## Extending smith

### My language is not supported.

Write a plugin. It is a program that speaks JSON-RPC over stdin and stdout, so it
can be written in any language. See [writing a plugin](writing-a-plugin.md).

### How do I check a plugin?

`smith plugin check <name or path>` runs the conformance suite.
