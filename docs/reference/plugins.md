# Plugins

Everything language-specific lives in a plugin: finding projects, their tasks,
scaffolding, writing versions, publishing, CI needs and health checks. smith's
core knows none of it. A plugin plays one or more roles (language, toolchain,
structure, scaffold, deploy), and each project picks the plugins it is built
with through its [stack](stacks.md).

## Built-in plugins

The table is a summary. Each plugin has a page in [the built-in plugin catalogue](../plugins/using/catalogue/README.md)
with what it detects, its tasks, scaffolds, settings, toolchain handling, release and registry
support, CI contributions and limits, written from its code and from real runs.

| Plugin | Roles | Does | Release |
|---|---|---|---|
| [`node`](../plugins/using/catalogue/node.md) | language (+ default toolchain, scaffold) | `package.json` projects (npm workspaces); every `package.json` script is a task; an app that builds also gets `prune`, a package gets `deps-check`; plain JavaScript with `--stack none`, TypeScript through a stack; `"smith": {"native": true}` builds a project on each OS | `npm publish`, skipping an existing version; a VS Code extension with `vsce publish`; an app is never sent to npm |
| [`go`](../plugins/using/catalogue/go.md) | language (+ default toolchain, scaffold) | `go.mod` projects (in a root `go.work`); `go build`, `go test`, `golangci-lint`; a package also gets `deps-check` | a `<dir>/vX.Y.Z` module tag; an app with a `[projects.<app>.release]` section also gets cross-compiled archives |
| [`python`](../plugins/using/catalogue/python.md) | language (+ default toolchain, scaffold) | `pyproject.toml` projects; one root `.venv`, Ruff, mypy and pytest on smith's defaults, PyPA `build`; a package also gets `deps-check` | `twine upload` of files the index lacks |
| [`rust`](../plugins/using/catalogue/rust.md) | language (+ default toolchain, scaffold) | `Cargo.toml` crates in a root Cargo workspace; `cargo build`, `test`, `clippy`, `check`, `fmt --check`, `package` | `cargo publish` of package crates (`CARGO_REGISTRY_TOKEN`) |
| [`dotnet`](../plugins/using/catalogue/dotnet.md) | language (+ default toolchain, scaffold) | `*.csproj`/`*.fsproj` projects; `dotnet build`, `test`, `format --verify-no-changes`, `pack` | `dotnet nuget push` of packages (`NUGET_API_KEY`), skipping a version the feed has |
| [`ruby`](../plugins/using/catalogue/ruby.md) | language (+ default toolchain, scaffold) | gems (`*.gemspec`) the root `Gemfile` lists with `gemspec path:`; RuboCop, RSpec, `exe/` | `gem build`, then `gem push` (`GEM_HOST_API_KEY`) |
| [`php`](../plugins/using/catalogue/php.md) | language (+ default toolchain, scaffold) | Composer packages the root `composer.json` requires through path repositories; PHPStan, PHPUnit | nothing is uploaded; Packagist reads the tags of a mirror repository (`split_repo`) |
| [`cpp`](../plugins/using/catalogue/cpp.md) | language (+ default toolchain, scaffold) | CMake projects (a `CMakeLists.txt` that calls `project()`); `configure`, `build`, CTest, cppcheck; vcpkg when configured | none: released by its tag |
| [`dart`](../plugins/using/catalogue/dart.md) | language (+ default toolchain, scaffold; the `flutter` stack) | the packages of a root pub workspace, Dart or Flutter; `analyze`, `format`, `test`, a web or CLI build | pub.dev through the generated `pub.yml` workflow; a private repository with `PUB_TOKEN` |
| [`jvm`](../plugins/using/catalogue/jvm.md) | language (+ default toolchain, scaffold; the `kotlin` and `android` stacks) | Maven modules (Java) and Gradle modules (Kotlin, Android); build, test, Checkstyle or ktlint | `mvn deploy` or Gradle `maven-publish` to `MAVEN_REPOSITORY_URL`, or `[registries.<name>] maven` |
| [`bicep`](../plugins/using/catalogue/bicep.md) | language (+ default toolchain, scaffold) | Azure infrastructure as code: each `main.bicep`; `az bicep lint`, `az bicep build` | none: released by its tag |
| [`lua`](../plugins/using/catalogue/lua.md) | language (+ default toolchain, scaffold) | LuaRocks projects: each `*.rockspec`; `luacheck`, `busted`, `luarocks make` | none: released by its tag |
| [`applescript`](../plugins/using/catalogue/applescript.md) | language (+ default toolchain, scaffold) | directories of `.applescript` sources; `osacompile` (macOS only) | none: released by its tag |
| [`vba`](../plugins/using/catalogue/vba.md) | language (+ default toolchain, scaffold) | directories of VBA modules exported as text; no tasks | none: released by its tag |
| [`swift`](../plugins/using/catalogue/swift.md) | language (+ default toolchain, scaffold) | Swift packages: each `Package.swift`; `swift build`, `swift test` | none: released by its tag |
| [`objc`](../plugins/using/catalogue/objc.md) | language (+ default toolchain, scaffold) | CocoaPods libraries: each `*.podspec`; `pod lib lint` (macOS) | none: released by its tag |
| [`recipe`](../plugins/using/catalogue/recipe.md) | language (+ default toolchain, scaffold) | a `*.recipe.toml` in a folder of its own is a package project; `test` runs `smith recipe check --all` | its tag, with the recipe attached to the GitHub Release |
| [`typescript`](../plugins/using/catalogue/typescript.md) | scaffold and toolchain, provides `build`, `typecheck`; publishes `node-ts` | scaffolds TypeScript projects that build and type-check with `tsc` | — |
| [`vite`](../plugins/using/catalogue/vite.md) | scaffold and toolchain, provides `build`, `typecheck`; publishes `react`, `react-p5`, `vue`, `svelte`, `html` | scaffolds browser apps in TypeScript that Vite serves and builds | released by its tag |
| [`angular`](../plugins/using/catalogue/angular.md) | scaffold and toolchain, provides `build`, `typecheck`; publishes `angular` | scaffolds an Angular app built with the Angular CLI | released by its tag |
| [`astro`](../plugins/using/catalogue/astro.md) | scaffold and toolchain, provides `build`; publishes `docs-site` | scaffolds an Astro Starlight documentation site | released by its tag |
| [`vscode`](../plugins/using/catalogue/vscode.md) | scaffold, toolchain and deploy, provides `build`, `typecheck`; publishes `vscode-extension` | scaffolds, bundles (esbuild) and packs a VS Code extension | `vsce publish` to the Marketplace (`VSCE_PAT`) and, with `OVSX_PAT`, Open VSX |
| [`eslint`](../plugins/using/catalogue/eslint.md) | toolchain, provides `lint` | Node projects lint with ESLint and `@monosmith/eslint-config` | — |
| [`jest`](../plugins/using/catalogue/jest.md) | toolchain, provides `test`; publishes `node-jest` | Node projects test with Jest (`ts-jest`) | — |
| [`vitest`](../plugins/using/catalogue/vitest.md) | toolchain, provides `test`; publishes `node-vitest` | Node projects test with Vitest | — |
| [`playwright`](../plugins/using/catalogue/playwright.md) | toolchain, provides `e2e`, and a convention (`[conventions.playwright]`); publishes `node-vitest-playwright` | the `e2e` task of every Node project with a Playwright suite, and an `e2e-browsers` task before it (see below) | — |
| [`docker`](../plugins/using/catalogue/docker.md) | deploy, a convention for every language (`[conventions.docker]`) | an `image` task for every app with a `Dockerfile`; writes a `Dockerfile` for a new Go, static web or Node app | the release builds the image once, tags it with the version and `latest`, and pushes to `[conventions.docker] registry` or `registries` (see [releasing](releasing.md#container-images)) |
| [`go-smith-plugin`](../plugins/using/catalogue/go-smith-plugin.md) | scaffold, requires the `go` language plugin; publishes the `go-smith-plugin` stack | `smith add app my-plugin --stack go-smith-plugin` writes a smith plugin in Go on the plugin SDK, with a conformance test | — |
| [`vertical-slices`](../plugins/using/catalogue/vertical-slices.md) | structure | a `slice-check` task for Go, Python, Dart, Rust, Ruby, C++, .NET, JVM and PHP projects, run before `lint`, in each language's own file-naming style | — |
| [`hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc`, `package-by-layer`](../plugins/using/catalogue/architectures.md) | structure, a convention for each language it reads (`[conventions.hexagonal]`) | an `architecture-check` task that reports the imports the architecture forbids, run before `lint` | — |

### playwright

- **Run the suites one at a time.** Suites usually share a fixture port and,
  under `xvfb-run`, a display, so run them with `smith run e2e --parallel 1`,
  as the `e2e` pipeline does. A plain `smith run e2e` runs them in parallel.
- **Reports** go to `dist/e2e/<project dir>`, beside the build rather than in
  `dist/<project dir>`: that directory is the build task's output and is
  replaced whenever the build is restored from the cache, reports with it.
- **Browsers** are installed by the `e2e-browsers` task. The `e2e` pipeline
  caches them under a key made of the lockfile (the Playwright version) and the
  manifests (the `browsers` lists); a cache that lacks one only makes
  `playwright install` download it.
- **The app under test** is started before the suite and stopped after by the
  task's `serve` setting; see
  [Serving an app for a task](tasks-and-cache.md#serving-an-app-for-a-task).
- **Not yet:** a starter for the library flavour (`smith add` writes the runner
  flavour), and doctor checks that the browsers are installed and that the unit
  test runner leaves `e2e/` alone (the starter's `*.e2e.ts` names keep Vitest
  and Jest away).

Built-in plugins ship inside the `smith` binary but still run as separate
processes (`smith plugin serve <name>`), speaking the same protocol an external
plugin does. `smith plugin list` shows each one's levels (roles), the slots it
fills and what it requires (`--json` adds its conflicts and capabilities), and
`smith stacks` the stacks they publish. `smith plugin enable`, `disable` and
`set` change which built-in language plugins load (`[plugins] enabled` in `smith.toml`);
see [`smith plugin`](../manual/plugin.md) and
[Installing, enabling and configuring plugins](../plugins/using/managing-plugins.md).

## Writing a plugin

This section is the reference. For a guide to each kind of plugin, with working
examples in Go, TypeScript, Python, C# and Node.js, read [Creating plugins](../plugins/creating/README.md)
(start with the [quickstart](../plugins/creating/quickstart.md)); for every method and
field in tables, [the protocol cheat sheet](../plugins/creating/protocol-cheat-sheet.md);
for using, installing and configuring plugins, [Using plugins](../plugins/using/README.md);
for why plugins are processes and how smith loads them, read
[the plugin system](../architecture/plugin-system.md).

A plugin is any executable that speaks the protocol on stdin/stdout; anything
it wants the user to see goes to stderr. Declare it in `smith.toml`, with its
command, or by name alone when it is installed:

```toml
[plugins.external.rust]
command = ["smith-plugin-rust"]   # any executable; on PATH, or a path

[plugins.external.kiwi]
version = "^1.2"                 # installed on the search path; "1.2.3", "~1.2", ">=1.2" also work
```

An external plugin with the name of a built-in one replaces it.

An external plugin starts only after you have trusted it (`smith plugin trust <name>`), and with a
minimal environment: the variables smith lists as the baseline, plus the ones the plugin declares
(`env` in its `plugin.json`) and the ones `smith.toml` grants (`env` under `[plugins.external.<name>]`).
In CI the plugin starts only when `smith.toml` pins its `sha256`. See
[Plugin security](../plugins/using/plugin-security.md). The built-in plugins need none of this.

A plugin builds every project into `protocol.ProjectBuildDir(root)`
(`dist/<project dir>`), and declares that directory as the task's outputs with
`protocol.BuildDirFromProject(root)` (`../../dist/apps/web`, relative to the
project). `smith doctor` fails any output elsewhere.

### Installing a plugin

An installed plugin is a directory holding a `plugin.json`:

```json
{
  "name": "kiwi",
  "version": "1.2.0",
  "description": "the kiwi language",
  "command": ["./bin/smith-plugin-kiwi-{os}-{arch}{exe}"],
  "env": ["NPM_TOKEN"]
}
```

A command path starting with `./` or `../` is relative to that directory;
`{os}`, `{arch}` and `{exe}` (`.exe` on Windows) are filled in, so one
directory can carry a binary per platform. `env` is optional: the environment
variables the plugin needs, as names or prefixes ending in `*`. smith looks for
`<dir>/<name>/plugin.json` in these directories, the first match winning:

1. the workspace's `.smith/plugins`
2. each directory of `SMITH_PLUGIN_PATH`
3. `$SMITH_HOME/plugins` (`SMITH_HOME` defaults to `~/.smith`)

None of it depends on how smith itself was installed. `smith plugin list` shows
the built-in plugins, the search path and what is installed on it.
How plugins are published, found and installed by name (a registry,
Homebrew/Yeoman style) is not decided yet (#164).

A project's cache keys include the version of every plugin that shapes it,
so upgrading a plugin re-runs its tasks.

### Checking a plugin

```sh
smith plugin check kiwi --generate
smith plugin check ./bin/smith-plugin-kiwi
```

runs the conformance suite: the plugin names itself, its roles match what it
implements, its stacks are complete, and every method it advertises answers.
With `--generate`, a plugin that generates projects makes an app, a lib and a
package in a temporary directory, and a language plugin must discover them as
generated, with their kinds, the stack they were asked to record, and tasks
that have commands. In Go, run the same checks in your tests with
`conformance.Check` (`packages/plugin-sdk/conformance`) over an in-process
connection; in TypeScript, with `checkPlugin`; in Python, with `check_plugin`; in C#, with `ConformanceChecks.CheckPluginAsync`. Every built-in plugin passes it.

The quickest start is a scaffold: `smith add app my-plugin --stack
go-smith-plugin` writes a Go module that is a working plugin (a structure
plugin giving projects a `hello` task) on the SDK, with a conformance test,
and its README says how to declare it.

In Go, the SDK (`github.com/monosmith-app/monosmith/packages/plugin-sdk`) does the
protocol for you:

```go
package main

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

type rust struct{ root string }

func (r *rust) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	r.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "rust", Version: "0.1.0"}, Kind: protocol.KindLanguage,
		Languages: []string{"rust"}, Manifests: []string{"Cargo.toml"},
	}, nil
}

// Implementing Discover is what advertises the discover capability.
func (r *rust) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) { /* ... */ }

func main() { serve.Main(&rust{}) }
```

`host.Connect` drives a plugin in-process over pipes, which is how to test one
(see `packages/plugin-sdk/host/plugin_client_test.go`).

In TypeScript, `@monosmith/plugin-sdk` does the same: a plugin is an object, its capabilities are
the members it has, and `connectInProcess` and `checkPlugin` test it
([The TypeScript SDK](../plugins/creating/typescript-sdk.md)):

```ts
import { serve } from '@monosmith/plugin-sdk'

await serve({
  describe: () => ({ plugin: { name: 'rust', version: '0.1.0' }, kind: 'language', languages: ['rust'], manifests: ['Cargo.toml'] }),
  discover: (params) => ({ projects: [ /* ... */ ] }),
})
```

In Python, `monosmith-plugin-sdk` (`monosmith_plugin_sdk`) does the same: a plugin is an object, its
capabilities are the methods it has, and `connect_in_process` and `check_plugin` test it
([The Python SDK](../plugins/creating/python-sdk.md)):

```python
from monosmith_plugin_sdk import serve


class Rust:
    def describe(self, params):
        return {"plugin": {"name": "rust", "version": "0.1.0"}, "kind": "language",
                "languages": ["rust"], "manifests": ["Cargo.toml"]}

    def discover(self, params):
        return {"projects": []}  # ...


if __name__ == "__main__":
    serve(Rust())
```

In C#, `MonoSmith.Plugin.Sdk` does the same: a plugin is a class, its capabilities are the interfaces it implements, and
`PluginClient.ConnectInProcessAsync` and `ConformanceChecks.CheckPluginAsync` test it
([The C# SDK](../plugins/creating/csharp-sdk.md)):

```csharp
using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

return await PluginServer.RunAsync(new Rust());

internal sealed class Rust : IPlugin, IDiscoverer
{
    public PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "rust", Version = "0.1.0" },
        Kind = PluginKind.Language,
        Languages = ["rust"],
        Manifests = ["Cargo.toml"],
    };

    public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation) =>
        Task.FromResult(new DiscoverResult { Projects = [] }); // ...
}
```

**The contract is a JSON Schema.** The Go types of the protocol are the source of truth, and
`packages/plugin-sdk/schema/protocol.schema.json` is generated from them (draft 2020-12): one definition
per message, `x-methods` (each of the 18 methods with its params, its result and the capability it
needs), `x-errorCodes`, `x-roles`, `x-protocolVersion` and `x-transport` (the framing). A member is
required when its Go field has no `omitempty`. A Go test fails when the file is stale, and the SDKs of other
languages generate their types from it, so a plugin in a language without an SDK can read it too.

## The protocol (version 1)

JSON-RPC 2.0 over stdio, framed like the Language Server Protocol:

```
Content-Length: 85\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1],...}}
```

| Method | Params → Result | Capability |
|---|---|---|
| `initialize` | protocol versions, host, workspace root → chosen version, plugin identity, kind (`language`/`convention`), roles (its levels), the tasks it provides (its slots), what it requires and conflicts with, the stacks it publishes, capabilities, languages, manifest names, and `concurrency` (how many requests it can work on at once; absent is 1) | always |
| `discover` | workspace-relative manifest paths the host found → projects | `discover` |
| `generate` | kind, name, directory, workspace info, stack, `convention` when the plugin is asked as a workspace convention rather than as one of the stack's plugins, and `settings` (the plugin's own table from the recipe the project is created from) → files written, notes | `generate` |
| `version/write` | project, version, new dependency versions, dry run → files changed | `version` |
| `publish` | project, version, dry run → published / skipped with a reason | `publish` |
| `ci` | the plugin's projects → toolchains (with their GitHub and Azure setup steps, or an install command), setup and audit steps (an audit step may name `cacheFiles`, the files that decide its verdict: smith skips it while they are unchanged on the same day), `postRelease` steps (run at the end of `smith ci release` after every push to the release branch, released or not; a step that is not blocking only warns), credentials (`setup` marks one installs need too), GitHub permissions, and `authenticate`: provider-native steps that sign in to a private registry before setup | `ci` |
| `doctor` | the plugin's projects → findings, each with a remedy and, when the plugin can write it, a fix (files' new contents, command steps and choices; marked safe when it only restores what the plugin owns) | `doctor` |
| `remove` | a project whose directory the host deleted, dry run → the files the plugin cleaned (go.work, the lockfile), notes | `remove` |
| `published` | a project → the newest version its registry has, or none, or that the registry could not be asked | `published` |
| `dependencies/install` | a project, packages, dev → the files changed, notes; with no project, the projects it serves, to install them all | `dependencies` |
| `dependencies/uninstall` | a project, packages → the files changed, notes | `dependencies` |
| `dependencies/declared` | the plugin's projects → every dependency their manifests declare (spec, section, peer, rewritable), and the versions installed | `declarations` |
| `dependencies/rewrite` | declarations and their new specs → the files changed | `declarations` |
| `dependencies/latest` | names, tools → each name's newest stable release, the names the registry could not be asked about, and the pinned tools against their newest releases | `latest` |
| `contribute` | the projects whose stack includes the plugin, or a convention applies to → new tasks, replaced tasks and extra task dependencies, by project | `contribute` |
| `shutdown` | → `{}`; the plugin then exits | always |
| `cancel` | a notification (no `id`, no response): the id of an in-flight request to stop → that request is answered with its result, or with error `-32800` | `cancel` |

A **project** carries its name, root, language, kind (`app`, `lib`,
`package`), the identifiers other projects import it by (`exports`), the
identifiers its sources import (`imports` — the host keeps those another
project exports and turns them into graph edges), its manifest version, an
optional release tag prefix, its tasks, and the stack its manifest records.
The host adds `registries`: the private registry the workspace installs from
and the one the project publishes to, from smith.toml `[registries]` (nil
means the public registries). A language plugin points its package manager at
them and publishes there.
A task marked `verify` is run by `smith ci verify` too, after the verify
tasks, for the affected projects that have it (the docker convention marks
an app's `image`); it is skipped, with a warning, where its program is not
installed.

Rules:

- The host never calls a method the plugin did not advertise. The Go SDK derives
  capabilities from the interfaces a plugin implements, and the TypeScript and Python SDKs from the
  members of the plugin object, and the C# SDK from the interfaces of the class, so they cannot lie.
- A breaking change raises the integer protocol version; additions are new
  capabilities or optional fields, which old peers ignore. `cancel` and `concurrency` are such
  additions ([ADR 0037](../architecture/decisions/0037-cancel-and-concurrency-in-the-plugin-protocol.md)).
- `initialize` must come first. A failure the user must act on returns error
  code `-32000` with a message saying what to do.
- Every call has a timeout. When it runs out, or the command is cancelled, the host sends `cancel` to a plugin
  that advertised it and waits five seconds for the answer; a plugin that does not answer, or cannot be
  asked, is killed.
- The host never has more requests in flight to a plugin than its `concurrency` allows (absent: one at a
  time); answers may come in any order and are matched by `id`.
- Paths in messages are workspace-relative with forward slashes.

## Edges the host adds

A plugin reports a project's `imports` and `exports`, and smith links projects
by them. Edges between languages come from the host, not from a plugin: a
project owns the OpenAPI, AsyncAPI, Protocol Buffers and GraphQL files inside
its directory (and the `schemas` globs of `[projects.<name>]`), and depends on
the owner of every file its `consumes_schemas` globs match. A plugin needs
nothing for it; the edge shows in `deps` of the graph and in `smith graph
--blast-radius`. See [Shared schemas](tasks-and-cache.md#shared-schemas).

## Roles

A plugin declares its roles in `initialize` (`roles`; a language plugin that
declares none is language, toolchain and scaffold, any other is structure):

- A **language** plugin implements `discover`, and usually `generate`,
  `version/write` and `publish`. It is what makes a new language work in smith:
  the graph, the cache, affected and releases need nothing else.
- A **toolchain**, **structure** or **deploy** plugin implements
  `contribute` (`serve.Contributor` in the Go SDK) to add or replace tasks of
  the projects whose stack includes it, `generate` (`serve.Generator`) to add
  its files, scripts and dependencies when such a project is created, and
  `ci` and `doctor` for those projects. A toolchain plugin lists the tasks it
  owns in `provides`, so a stack cannot hold two test runners.
- A **scaffold** plugin implements `generate` and writes a new project in
  place of the language plugin. It records `params.stack` in the manifest the
  way the language plugin does, so discovery reports it back.

A plugin can publish ready-made `stacks` in `initialize`. Any plugin of kind
`convention` can also be applied workspace-wide with `[conventions.<name>]`;
see [conventions](conventions.md). [Stacks](stacks.md) explains how a
project's plugins combine.

## Levels, slots, requirements and conflicts

A role is a **level** in a project's make-up. There is no separate `level`
field: one plugin can play several levels (`typescript` is a scaffold and a
toolchain), so `roles` lists them. Some levels are exclusive and the others
additive:

| Level | A project has |
|---|---|
| `language` | exactly one |
| `scaffold` | at most one |
| `toolchain` | any number, but each **slot** (a task in `provides`, such as `test` or `lint`) is filled by one |
| `structure`, `deploy` | any number; they only add |

`provides` is the toolchain's slots: there is no separate `slots` field. Two
more optional `initialize` fields say what a plugin needs around it:

```json
{
  "roles": ["toolchain"],
  "provides": ["test"],
  "requires": [{"level": "language", "name": "node"}],
  "conflicts": [{"name": "karma"}]
}
```

A reference (`requires` and `conflicts` entries) names plugins by any of:
- `level`: the level the plugin is chosen for
- `name`: the plugin's name, or for a language plugin one of its `languages`
- `slot`: a task the plugin provides

Every field it sets must match, and it sets at least one. A plugin that
declares neither field requires nothing, so plugins written before these
fields keep working; the protocol stays at version 1.

In Go:

```go
Requires:  []protocol.PluginRef{protocol.RequiresLanguage("node")},
Conflicts: []protocol.PluginRef{protocol.ConflictsWith("karma")},
```

smith checks a project's plugins in one place, whatever chose them: a stack
or a recipe. It checks that:
- there is one language plugin and at most one scaffold
- every plugin plays the level it is chosen for and applies to the language
- each slot has one plugin
- every requirement is met and no two plugins conflict

An error names both sides, for example `jest and vitest both provide the test
task; keep one` or `vitest conflicts with karma (plugin karma); keep one`.
`smith plugin check` also verifies the declarations:
- a convention plugin with the toolchain role lists its slots
- only a toolchain provides tasks
- every reference names a known level, a name or a slot, and a conflict
  never names the plugin itself
