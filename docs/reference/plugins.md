# Plugins

Everything language-specific lives in a plugin: finding projects, their tasks,
scaffolding, writing versions, publishing, CI needs and health checks. smith's
core knows none of it. A plugin plays one or more roles (language, toolchain,
structure, scaffold, deploy), and each project picks the plugins it is built
with through its [stack](stacks.md).

## Built-in plugins

| Plugin | Roles | Does | Release |
|---|---|---|---|
| `node` | language (+ default toolchain, scaffold) | `package.json` projects (npm workspaces); every package.json script is a task; a package (`kind = package`) also gets `deps-check`, which fails when it imports a package its package.json does not declare (see tasks and cache); scaffolds plain JavaScript; `"smith": {"native": true}` builds a project on each OS (see CI pipelines, Native projects); `smith doctor` checks that a built package has every file its `main`, `module`, `types` and `exports` name | `npm publish`, skipping an existing version; a VS Code extension with `vsce publish` (`--target` per platform when native); an app is never sent to npm, and a native app's release asset is its build, zipped |
| `cpp` | language (+ default toolchain, scaffold) | CMake projects (a `CMakeLists.txt` that calls `project()`), using each other with `add_subdirectory`; `configure` (Ninja, into `dist/<project dir>`), `build`, CTest (test), cppcheck (lint); vcpkg when a project has a `vcpkg.json` and `VCPKG_ROOT` is set | none: released by its tag |
| `go` | language (+ default toolchain, scaffold) | `go.mod` projects (in a root `go.work`); `go build`, `go test`, `go vet` | a `<dir>/vX.Y.Z` module tag |
| `dotnet` | language (+ default toolchain, scaffold) | `*.csproj`/`*.fsproj` projects, each with its `tests/` project; `dotnet build`, `test`, `format --verify-no-changes` (lint), `pack`; bin/ and obj/ under `dist/dotnet` | `dotnet nuget push` of packages (`NUGET_API_KEY`), skipping a version the feed has |
| `php` | language (+ default toolchain, scaffold) | Composer packages the root `composer.json` requires through path repositories, one `vendor/`; PHPStan level 8 (lint), PHPUnit (test), the `bin` script (start) | none to upload: Packagist reads the release's tags |
| `dart` | language (+ default toolchain, scaffold; the `flutter` stack) | the packages of a root pub workspace (`pubspec.yaml`), Dart or Flutter; `analyze --fatal-infos` (lint), `dart format` (format), `test`; a Flutter app builds for the web (`flutter build web`), a Dart app with `dart build cli`, both into `dist/<app dir>` | `pub publish` of packages: pub.dev, or the `publish_to` repository with `PUB_TOKEN` |
| `jvm` | language (+ default toolchain, scaffold; the `kotlin` and `android` stacks) | Maven modules of the root aggregator pom (`pom.xml`), in Java; `mvn install` (build), `test`, Checkstyle with Google's rules (lint), `exec:java` (start); every build under `dist/maven`, installed with a flattened pom. Gradle modules (`build.gradle.kts`, `build.gradle`) of the root settings, in Kotlin with the `kotlin` stack; `gradle assemble` (build), `test`, ktlint or Checkstyle (lint), `run` (start); builds under `dist/gradle`. Android apps with the `android` stack (AGP, Kotlin built in): the APKs under `dist/gradle/<module>/outputs/apk`, ktlint and Android Lint, one task after another (build, test, lint) | `mvn deploy` of packages to `MAVEN_REPOSITORY_URL` (`MAVEN_USERNAME`, `MAVEN_PASSWORD`), the version written into dependent modules too; Gradle packages publish to the same repository with `maven-publish` |
| `bicep` | language (+ default toolchain, scaffold) | Azure infrastructure as code: each `main.bicep` is a project (a deployment in `apps/`, a module in `libs/`), using the workspace modules it references; `az bicep lint` with `bicepconfig.json`'s rules raised to errors (lint), `az bicep build` into `dist/<project dir>/template` (build), after `build-params` into `dist/<project dir>/parameters` (params) | none to upload: released by its tag (`metadata version` in `main.bicep`) |
| `python` | language (+ default toolchain, scaffold) | `pyproject.toml` projects (src layout); one root `.venv`, Ruff, pytest, PyPA `build` | `twine upload` of files the index lacks |
| `ruby` | language (+ default toolchain, scaffold) | gems (`*.gemspec`) the root `Gemfile` lists with `gemspec path:`, one Bundler resolution; RuboCop (lint), RSpec (test), `exe/` (start) | `gem build` into `dist/<gem dir>`, then `gem push` (`GEM_HOST_API_KEY`) |
| `rust` | language (+ default toolchain, scaffold) | `Cargo.toml` crates in a root Cargo workspace; `cargo build`, `test`, `clippy` (lint), `check` (typecheck), `fmt --check` (format), `package`; every build under `dist/cargo` | `cargo publish` of package crates (`CARGO_REGISTRY_TOKEN`) |
| `typescript` | scaffold and toolchain, provides `build`, `typecheck`; publishes `node-ts` | scaffolds TypeScript projects that build and type-check with `tsc` | — |
| `vite` | scaffold and toolchain, provides `build`, `typecheck`; publishes `react`, `react-p5`, `vue`, `svelte`, `html` | scaffolds browser apps in TypeScript (React, React with p5, Vue, Svelte, or a page with no framework) that Vite serves (`start`) and builds into `dist/<app dir>`; type-checked by `tsc`, `vue-tsc` or `svelte-check`; each stack lints with eslint and tests with Vitest | released by its tag |
| `angular` | scaffold and toolchain, provides `build`, `typecheck`; publishes `angular` | scaffolds an Angular app (standalone root component, no zone.js) that the Angular CLI serves (`ng serve`) and builds into `dist/<app dir>`; `ngc` type-checks its templates; linted by eslint, tested with Vitest | released by its tag |
| `astro` | scaffold and toolchain, provides `build`; publishes `docs-site` | scaffolds an Astro Starlight documentation site (a splash page, guides and reference with generated sidebars); `build` runs `astro check` then `astro build` into `dist/<app dir>`, `start` runs `astro dev`; linted by eslint | released by its tag |
| `vscode` | scaffold and toolchain, provides `build`, `typecheck`; publishes `vscode-extension` | scaffolds a VS Code extension in TypeScript (a command showing the greeting) that esbuild bundles into `dist/<dir>/main.js`; the node plugin's `package` packs it into `dist/drop/<name>-<version>.vsix`; linted by eslint, tested with Vitest | `vsce publish` of the staged build to the Marketplace (`VSCE_PAT`), skipping a version it has |
| `eslint` | toolchain, provides `lint` | Node projects lint with ESLint and `@monosmith/eslint-config` (TypeScript rules, Standard style, vertical slices) | — |
| `jest` | toolchain, provides `test`; publishes `node-jest` | Node projects test with Jest (`ts-jest`) | — |
| `vitest` | toolchain, provides `test`; publishes `node-vitest` | Node projects test with Vitest | — |
| `playwright` | toolchain, provides `e2e`, and a convention (`[conventions.playwright]`); publishes `node-vitest-playwright` | the `e2e` task of every Node project with a Playwright suite: the Playwright test runner (a `playwright.config.*` or `@playwright/test`) runs `playwright test`, and a project that drives the `playwright` library from its own runner keeps its `e2e` script (the library may be declared by the project or, hoisted by npm workspaces, by the root `package.json`). Before it, an `e2e-browsers` task installs the browsers (`playwright install`, with `--with-deps` in CI on Linux), so nothing else downloads one. `"smith": {"playwright": {"display": true}}` runs the suite under `xvfb-run -a` on Linux, `"browsers"` names what to install (default `chromium`). The task is cached like `test`, with `dist/e2e/<project dir>` (where the starter's config writes reports and traces) as its outputs. Doctor checks Playwright is installed, `xvfb-run` for display suites, and that an e2e pipeline exists | — |
| `docker` | deploy, a convention for every language (`[conventions.docker]`) | an `image` task (`docker build`, the workspace as context, after `build`) for every app with a `Dockerfile`; writes one for a new Go app (built in Docker, shipped on distroless) or static web app (nginx serving `dist/<app dir>`) | the release builds the image tagged with the version and `latest` and pushes both to `[conventions.docker] registry` (see [releasing](releasing.md#container-images)); with no registry the image is built and not pushed; an app with a `Dockerfile` releases without a `[release]` section |
| `vertical-slices` | structure | a `slice-check` task for Go, Python, Dart, Rust, Ruby, C++, .NET, JVM and PHP projects, run before `lint`, in each language's own file-naming style | — |

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
- **Not yet:** a starter for the library flavour (`smith add` writes the runner
  flavour), starting the app under test before the suite (#261), and doctor
  checks that the browsers are installed and that the unit test runner leaves
  `e2e/` alone (the starter's `*.e2e.ts` names keep Vitest and Jest away).

Built-in plugins ship inside the `smith` binary but still run as separate
processes (`smith plugin serve <name>`), speaking the same protocol an external
plugin does. `smith plugin list` shows each one's roles and capabilities, and
`smith stacks` the stacks they publish.

## Writing a plugin

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
  "command": ["./bin/smith-plugin-kiwi-{os}-{arch}{exe}"]
}
```

A command path starting with `./` or `../` is relative to that directory;
`{os}`, `{arch}` and `{exe}` (`.exe` on Windows) are filled in, so one
directory can carry a binary per platform. smith looks for
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
connection. Every built-in plugin passes it.

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

## The protocol (version 1)

JSON-RPC 2.0 over stdio, framed like the Language Server Protocol:

```
Content-Length: 85\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1],...}}
```

| Method | Params → Result | Capability |
|---|---|---|
| `initialize` | protocol versions, host, workspace root → chosen version, plugin identity, kind (`language`/`convention`), roles, the tasks it provides, the stacks it publishes, capabilities, languages, manifest names | always |
| `discover` | workspace-relative manifest paths the host found → projects | `discover` |
| `generate` | kind, name, directory, workspace info, stack → files written, notes | `generate` |
| `version/write` | project, version, new dependency versions, dry run → files changed | `version` |
| `publish` | project, version, dry run → published / skipped with a reason | `publish` |
| `ci` | the plugin's projects → toolchains (with their GitHub and Azure setup steps, or an install command), setup and audit steps, credentials | `ci` |
| `doctor` | the plugin's projects → findings, each with a remedy and, when the plugin can write it, a fix (files' new contents, marked safe when it only restores what the plugin owns) | `doctor` |
| `remove` | a project whose directory the host deleted, dry run → the files the plugin cleaned (go.work, the lockfile), notes | `remove` |
| `published` | a project → the newest version its registry has, or none, or that the registry could not be asked | `published` |
| `dependencies/install` | a project, packages, dev → the files changed, notes; with no project, the projects it serves, to install them all | `dependencies` |
| `dependencies/uninstall` | a project, packages → the files changed, notes | `dependencies` |
| `dependencies/declared` | the plugin's projects → every dependency their manifests declare (spec, section, peer, rewritable), and the versions installed | `declarations` |
| `dependencies/rewrite` | declarations and their new specs → the files changed | `declarations` |
| `dependencies/latest` | names, tools → each name's newest stable release, the names the registry could not be asked about, and the pinned tools against their newest releases | `latest` |
| `contribute` | the projects whose stack includes the plugin, or a convention applies to → new tasks, replaced tasks and extra task dependencies, by project | `contribute` |
| `shutdown` | → `{}`; the plugin then exits | always |

A **project** carries its name, root, language, kind (`app`, `lib`,
`package`), the identifiers other projects import it by (`exports`), the
identifiers its sources import (`imports` — the host keeps those another
project exports and turns them into graph edges), its manifest version, an
optional release tag prefix, its tasks, and the stack its manifest records.

Rules:

- The host never calls a method the plugin did not advertise. The Go SDK derives
  capabilities from the interfaces a plugin implements, so it cannot lie.
- A breaking change raises the integer protocol version; additions are new
  capabilities or optional fields, which old peers ignore.
- `initialize` must come first. A failure the user must act on returns error
  code `-32000` with a message saying what to do.
- Every call has a timeout; a plugin that does not answer is killed.
- Paths in messages are workspace-relative with forward slashes.

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
