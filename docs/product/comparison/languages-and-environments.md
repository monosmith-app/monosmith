# Languages and environments: a ready-made setup for eleven ecosystems

**Add a Rust service, a Python package or a .NET library to your monorepo with one command, and the project arrives with a worked example that lints, tests and builds. CI, the dev container, the cache and the release already know about it.** You write no BUILD files, no per-language rules and no pipeline steps.

This page defines what a "ready-made environment" is, checks it language by language, and sets MonoSmith against the fourteen other tools in the [comparison](README.md). It is one of three pages on what smith does best: see also [plugins and flexibility](plugins-and-flexibility.md) and [the dependency graph](dependency-graph.md).

## What "ready-made" means here

Six things, checked for each language:

| Element | What smith does |
|---|---|
| **A project to start from** | `smith add app\|lib\|package <name> --lang <language>` writes the manifest, a worked example (a small slice with a test) and the files the language needs |
| **The workspace around it** | the root files a language wants once per repository: a `.venv`, a `go.work`, a Cargo workspace, `Directory.Build.props`, a `Gemfile`, a pub workspace, a root `pom.xml` |
| **Tasks that work immediately** | `lint`, `typecheck`, `test` and `build` as the language's own tools define them, cached and in dependency order |
| **The right toolchain, everywhere** | CI installs and caches each toolchain; the dev container carries the same versions; `[toolchains]` pins a version and `smith doctor` checks it |
| **Audit and alignment** | a dependency audit in CI where the ecosystem has one, one version of each shared dependency |
| **A way to ship it** | `smith release` publishes to the language's registry, or tags it |

## Day one in three languages

All three were added to one scratch workspace with smith 1.65.0 on 2026-10-25 (Windows 11, toolchains installed by hand).

```text
$ smith add lib core --lang go
Adding lib core (go) in libs/core
Wrote 9 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  go.work
  libs/core/doc.go
  libs/core/go.mod
  libs/core/greeting/doc.go
  libs/core/greeting/greet_use_case.go
  libs/core/greeting/greet_use_case_test.go
  libs/core/greeting/greeting_contract.go
Import it from another Go project as example.com/acme/libs/core/<slice>.

$ smith add package sdk --lang python
Adding package sdk (python) in packages/sdk
Wrote 11 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  packages/sdk/README.md
  packages/sdk/pyproject.toml
  packages/sdk/src/acme_sdk/__init__.py
  packages/sdk/src/acme_sdk/greeting/__init__.py
  packages/sdk/src/acme_sdk/greeting/greet_use_case.py
  packages/sdk/src/acme_sdk/greeting/greeting_contract.py
  packages/sdk/tests/test_greet_use_case.py
  requirements-dev.txt
  ruff.toml
Import it from another Python project as `import acme_sdk`.

$ smith add lib engine --lang rust
Adding lib engine (rust) in libs/engine
Wrote 8 file(s):
  .cargo/config.toml
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  Cargo.toml
  libs/engine/Cargo.toml
  libs/engine/src/greeting/greet_use_case.rs
  libs/engine/src/greeting/mod.rs
  libs/engine/src/lib.rs
Import it from another crate as acme-engine = { path = "<relative path>" }.
```

The Python add also created the workspace's one `.venv` and installed the pinned tools into it; it took 70 seconds, most of it the network. Then, with the cache off, lint, test and build all three:

```text
$ smith run lint test build -p core,sdk,engine --no-cache
core:build ✔ 1.435s
core:test ✔ 2.777s
core:lint ✔ 1.807s
engine:build ✔ 1.49s
engine:test ✔ 1.728s
engine:lint ✔ 1.727s
sdk:test ✔ 1.727s
sdk:lint ✔ 716ms
sdk:build ✔ 24.157s
...
✔ 12/12 tasks succeeded (0 from cache) in 24.173s
```

The twelve tasks include a layout check per project. The commands behind them are the language's own: `go vet` and `go test`, `cargo clippy -D warnings` and `cargo test`, `ruff check`, `pytest` and PyPA `build` through the workspace `.venv`. `smith explain sdk:build` shows each one, and where each setting came from.

## All eleven languages in one workspace

The same workspace held eleven projects, one per language. On the first run, on a machine with the toolchains installed:

| Run | Result |
|---|---|
| `smith test` | 14 of 14 tasks succeeded in 21 seconds |
| `smith build` | 10 of 10 tasks succeeded in 26 seconds; the second run took 3.7 seconds with 5 tasks from cache |
| `smith lint` | 21 of 22 tasks succeeded; the Dart one did not, see below |
| `smith doctor` | 47 passed, 0 failed, 1 warning (a Bicep project with no test task) |

The one lint failure is `dart analyze`, which crashed with an analysis-server error on this machine (Dart 3.12.0 on Windows). The same crash happens when `dart analyze` is run by hand, outside smith.

### What each language gets

| Language | `smith add` writes | Workspace root | Default tasks | CI installs | Audit, then release |
|---|---|---|---|---|---|
| Node, TypeScript | `package.json`, `tsconfig`, a slice with a spec | npm workspaces, `eslint.config.mjs`, `tsconfig.base.json` | `lint`, `typecheck`, `test`, `build` | Node 24, `~/.npm` cache | `npm audit`, then npm |
| Python | `pyproject.toml`, `src/<name>`, `tests` | one `.venv`, `requirements-dev.txt`, `ruff.toml` | `lint` (Ruff), `test` (pytest), `typecheck` (mypy, when configured), `build` | Python 3.12, pip cache | `pip-audit`, then PyPI |
| Go | `go.mod`, a slice with a test | `go.work` | `build`, `test`, `lint` (`go vet`) | Go at the `go.work` version, module cache | `govulncheck`, then a module tag |
| Rust | crate `Cargo.toml`, `src` | root Cargo workspace, `.cargo/config.toml` | `build`, `test`, `lint` (Clippy, deny warnings), `typecheck`, `format` | Rust with Clippy and rustfmt, `~/.cargo/registry` | none by default, then crates.io |
| .NET | `.csproj`, a tests project | `Directory.Build.props` | `build`, `lint` (`dotnet format`), `test`, `pack` | .NET SDK from `global.json`, NuGet cache | none by default, then NuGet |
| Ruby | gemspec, `lib`, `spec` | root `Gemfile`, `.rubocop.yml` | `lint` (RuboCop), `test` (RSpec) | Ruby from `.ruby-version` | none by default, then RubyGems |
| PHP | `composer.json`, `src`, `tests` | root `composer.json`, `vendor/` | `lint` (PHPStan level 8), `test` (PHPUnit) | PHP 8.4 and Composer, Composer cache | none by default, then Packagist by tag |
| C and C++ | `CMakeLists.txt`, `include`, `src`, `tests` | CMake and Ninja, build directory under `dist/` | `configure`, `build`, `test` (CTest), `lint` (cppcheck) | CMake, Ninja, g++, cppcheck (and vcpkg when `vcpkg.json` exists) | none, then a tag |
| Dart, Flutter | `pubspec.yaml`, `lib`, `test` | pub workspace | `lint` (`analyze`), `format`, `test`, `build` (apps) | Dart or Flutter SDK, pub cache | none by default, then pub.dev |
| Java, Kotlin, Android | `pom.xml` (Maven) or Gradle scripts | root `pom.xml`, `.mvn` | `build`, `test`, `lint` (Checkstyle or ktlint) | JDK 21 (Temurin), Gradle and the Android SDK when needed | none by default, then a Maven repository |
| Azure Bicep | `main.bicep`, `main.bicepparam` | Azure CLI with Bicep | `lint`, `params`, `build` | Azure CLI and Bicep | none, then a tag |

Every language adds its toolchain to the one generated pipeline. For the eleven above that is eleven toolchain installs and seven package-store caches (cargo, composer, go, npm, NuGet, pip, pub) in a single GitHub Actions workflow, followed by `smith ci setup`, `audit`, `verify`, `pack` and `release`. The matching `.devcontainer/devcontainer.json` carries the same toolchains at the same versions, so a new teammate opens the repository in a container and finds the toolchains CI uses.

### Frameworks it scaffolds

A stack is a project's plugins: `smith add app site --stack react`. These were scaffolded in a clean workspace and run:

| Stack | Command run | Result |
|---|---|---|
| `react` (Vite, ESLint, Vitest) | `smith run build typecheck lint test` | 4 of 4 tasks |
| `vscode-extension` (esbuild, ESLint, Vitest, vsce) | `... package` | 6 of 6 tasks, including packing the extension |
| `docs-site` (Astro Starlight) | `build lint` | 2 of 2 tasks |
| `angular` | `typecheck lint test` | 3 of 3 tasks; `build` needs the Angular CLI's Node 24.15 or newer, and this machine had 24.12 |
| `node-vitest-playwright` | `build typecheck lint test` | 4 of 4 tasks; browsers were not downloaded |
| `node-jest` (the Node default: TypeScript, ESLint, Jest) | `lint typecheck test build` | passed in the eleven-language workspace |

`vue`, `svelte`, `html`, `react-p5`, `kotlin`, `android`, `flutter` and `go-smith-plugin` are listed by `smith stacks` and described in the [catalogue](../../plugins/using/catalogue/README.md); they were not run for this page.

## Compared with fourteen tools

**Only MonoSmith scaffolds, runs, caches, generates CI for and releases eleven ecosystems from one tool and one config file with no per-language rules to write.** The nearest in breadth are Nx (JavaScript by generators, other ecosystems by plugins) and Gradle (the JVM and C++).

On this page a ✅ in the environment rows means the tool's documentation shows a built-in generator for a new project in that language and tasks that run it, with no rules for you to write. ◐ means the tool supports the language, but you write BUILD files, rules, scripts or configuration, or add a plugin. ❌ means the documentation scopes the tool to another language. ❔ means the sources were searched and say nothing either way. Gradle gets ✅ for the JVM and C++ because `gradle init` creates applications and libraries in Java, Kotlin, Groovy, Scala and C++; Cargo gets ✅ for Rust because of `cargo new`; Nx gets ✅ for JavaScript and TypeScript because its generators create projects whose tasks its plugins infer.

| Row | **MonoSmith** | Nx | Turbo | Bazel | Pants | Lerna | Rush | moon | Buck2 | Gradle | Earthly | Lage | JS pkg mgrs | Cargo | Please |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| JavaScript, TypeScript | **✅** | ✅ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ❔ | ❔ | ◐ | ◐ | ◐ | ❌ | ❔ |
| Python | **✅** | ◐ | ◐ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❔ | ❌ | ◐ |
| Go | **✅** | ◐ | ◐ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❌ | ❌ | ◐ |
| Rust | **✅** | ◐ | ◐ | ◐ | ❔ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❔ | ✅ | ◐ |
| .NET | **✅** | ◐ | ◐ | ❔ | ❔ | ❌ | ❌ | ◐ | ❌ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ |
| Ruby | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ |
| PHP | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ |
| C and C++ | **✅** | ❔ | ❔ | ◐ | ❔ | ❌ | ❌ | ◐ | ◐ | ✅ | ◐ | ❌ | ❌ | ❌ | ◐ |
| Dart and Flutter | **◐** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ |
| Java, Kotlin, Android | **◐** | ◐ | ❔ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ✅ | ◐ | ❌ | ❌ | ❌ | ◐ |
| Azure Bicep | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ |
| A new project passes lint, test and build on first run | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Writes a dev container with CI's toolchains | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Generated CI installs and caches each toolchain | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Downloads and pins toolchains | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ✅ | ❔ | ✅ | ❔ | ❔ | ◐ | ❔ | ❔ |

Seven of the cells are worth reading closely.

- **Nx, JavaScript:** the most complete rival. Generators (`nx g`) create libraries and apps, plugins infer their tasks, and the documentation lists Gradle, Maven and .NET plugins and community plugins for Python, Go and Rust. smith matches it for JavaScript and adds the other ecosystems in the same tool and the same config.
- **Gradle and Cargo** are as ready-made as smith inside their own ecosystems. Gradle can also download a matching JDK; smith pins and checks versions and leaves installing to your manager (mise, asdf or the dev container), by [decision 0020](../../architecture/decisions/0020-toolchain-pins-checked-not-downloaded.md).
- **Bazel, Buck2, Please, Pants and moon** support many of these languages, with a BUILD file, rule set or `moon.yml` that you write. Bazel's documentation says the Node, Rust and Python rules are separately maintained ("rules_nodejs", "rules_rust", "rules_python"). Pants's `tailor` writes BUILD files for code that already exists; it does not create a project.
- **Turborepo, Lerna, Rush and Lage** are JavaScript tools. They run what `package.json` scripts say.
- **Ruby, PHP, Dart and Bicep:** no other tool in this set documents language-aware support; the nearest are moon (partial, experimental for Ruby; project categorisation only for PHP) and Earthly (any language, as containers).
- **Dev container:** no tool in this set documents writing one; Pants, moon and Earthly mention dev containers as places to run the tool.
- **Generated CI that installs toolchains:** setup steps appear in the other tools' hand-written examples (`actions/setup-node` in the Nx, Turborepo, Rush and moon guides). smith writes them from the languages in the workspace.

## What this costs you with other tools

- **In Bazel, Buck2 or Please** each language is a rule set and each package a BUILD file. You, or the people who maintain the repository, write and keep them in step with `go.mod`, `package.json` and `Cargo.toml`.
- **In Pants** you run `pants tailor` to generate BUILD files for the code you wrote. Its strongest languages are Python, Go and the JVM; JavaScript is documented as beta.
- **In Nx** every non-JavaScript ecosystem is another plugin to install, pin and keep current, some from the community.
- **In Turborepo, Lerna, Rush or Lage** the other languages are outside the tool, so their tests and builds are `package.json` scripts that call them, and their CI is yours.
- **In moon** each project has a `moon.yml` and each toolchain is configured; moon then installs the toolchain for you, which smith does not.
- **Everywhere**, the pipeline that installs eleven toolchains is hand-written, per CI provider, and nothing fails when a new language is added and the pipeline is not.

## Where another tool is the better choice

- **Your repository is one ecosystem, and its own tool is the best at it.** Gradle for the JVM and Android (smith's `jvm` plugin wraps Gradle and Maven), Cargo for Rust-only, Nx for a deep JavaScript monorepo.
- **You want the tool to install and pin toolchains.** moon downloads Node, Go and Rust versions through proto; Gradle provisions JDKs. smith pins versions, checks them in `smith doctor` and writes `.tool-versions` for mise and asdf.
- **You need hermetic builds or remote execution.** Bazel, Pants and Buck2 are built for it; smith runs tasks on one machine.
- **Your language is not on the list.** Swift, Scala, Haskell, Bun and Deno have no built-in plugin. You can [write one](plugins-and-flexibility.md) in any language, or use moon (Bun, Deno) or Gradle (Scala, Swift).
- **What was not verified for this page.** Gradle, Kotlin and Android projects were not run (no Gradle on the test machine); only Maven ran. Linux and macOS were not run. Dependency alignment with `smith sync` covers Node and Python; the dependency audit covers Node, Python and Go, with Rust opt-in. On 2026-10-25 `dart analyze` crashed here, as above.
- **A known wrinkle.** Adding the `react` stack to a workspace whose Node project already used the default Jest stack failed on 2026-10-25 with an npm `ERESOLVE` peer-dependency conflict (`@vitejs/plugin-react` against Jest's Babel 7). The reverse order, and a clean workspace, worked.

## Try it

```sh
smith new acme && cd acme
smith add app web --lang node
smith add lib core --lang go
smith add package sdk --lang python
smith build && smith doctor
```

Already have a repository? `smith adopt --dry-run` reports which of your projects smith would take on and what blocks, and changes nothing. Coming from Nx or Turborepo? `smith convert --dry-run` (see [Migration stories](../migration-stories.md)). Language-by-language detail, with the commands behind every cell, is in the [plugin catalogue](../../plugins/using/catalogue/README.md).

## Method and evidence

- **MonoSmith.** The runs on this page were made on 2026-10-25 with smith 1.65.0 on Windows 11 in throwaway directories, never in this repository's own workspace. Toolchains present: Node 24.12, Go 1.27, Python 3.12, Rust, .NET 10, Ruby 3.4, PHP 8.4 with Composer, CMake with Ninja and g++, Dart 3.12 with Flutter, JDK 21 with Maven, Azure CLI with Bicep. Not present: Gradle. The files each language gets, the CI steps and the dev container were read from the generated workspace; the catalogue pages in `docs/plugins/using/catalogue` give the release targets and the audit steps, from the code and from earlier runs on 1.62.0.
- **Other tools.** Read raw from each tool's own documentation source on 2026-10-25. Language support, generators and scope are the statements already recorded on the [tool pages](README.md#the-tools) and the language rows of the [master table](everything-compared.md); this page adds the generator statements below. ❔ is "searched, not stated", never "absent". Nothing was measured: the durations above are single runs on one machine and are not benchmarks.

### Sources

Retrieved 2026-10-25.

- Gradle (`gradle/gradle`, `platforms/documentation/docs/src/docs/userguide/reference/core-plugins/build_init_plugin.adoc`): the build types `java-application`, `kotlin-application`, `groovy-application`, `scala-application`, `cpp-application` and `swift-application` ("A command-line application implemented in C++").
- Cargo (`rust-lang/cargo`, `doc/book/src`): `cargo new` and `cargo init`, as recorded on the [Cargo page](cargo.md).
- Nx (`nrwl/nx`, `astro-docs/src/content/docs`): `features/generate-code.mdoc` ("Generators come as part of Nx plugins"); `features/multi-language-support.mdoc` (language plugins for Gradle, Maven, .NET; community plugins for Python, Go, Rust); the CI guides that use `actions/setup-node`.
- Bazel (`bazelbuild/bazel`, `docs`): `docs/bazel-and-javascript.mdx`, `external/overview.mdx` and `concepts/runfiles.mdx` for the separately maintained rule sets, as recorded on the [Bazel page](bazel.md).
- Pants (`pantsbuild/pants`, `docs/docs`): `getting-started/incremental-adoption.mdx` and the `tailor` goal, as recorded on the [Pants page](pants.md).
- moon (`moonrepo/moon`, `website/docs`): `config/toolchain.mdx` and the `proto` documentation (toolchain installation).
- A search for "devcontainer" and "dev container" in all fourteen sources found Pants, moon and Earthly using dev containers as places to run the tool, and none as generated output.
- Per-tool language support, scaffolding and CI statements: the Sources of each [tool page](README.md#the-tools).
- MonoSmith: `docs/plugins/using/catalogue/*.md`, `docs/reference/ci.md`, `docs/reference/configuration.md`, `docs/architecture/decisions/0020-*.md`, and the generated files of the scratch workspace.
