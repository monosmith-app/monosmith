# The built-in plugin catalogue

smith ships 24 plugins inside its binary: 12 language plugins and 12 add-ons (toolchain, structure,
scaffold and deploy plugins). This section has one page for each, written from the plugin's code
and from real runs: what it detects, the project kinds and tasks it adds, what it scaffolds, its
settings, how it handles the toolchain, how it releases, what it adds to CI, and where it is
limited. Every page ends with what was and was not verified.

The one-table summary is in [Plugins (reference)](../../../reference/plugins.md); the model behind
roles, stacks and conventions is in [What a plugin is and how plugins combine](../what-is-a-plugin.md).

## Language plugins

A language plugin finds projects from manifests and, by default, is also their toolchain and
scaffold. `smith plugin enable`, `disable` and `set` switch these on and off.

| Plugin | Projects it finds | Page |
|---|---|---|
| `node` | `package.json` projects (npm workspaces); every script is a task | [node](node.md) |
| `go` | `go.mod` modules in a root `go.work` | [go](go.md) |
| `python` | `pyproject.toml` projects, one root `.venv` | [python](python.md) |
| `rust` | `Cargo.toml` crates in a root Cargo workspace | [rust](rust.md) |
| `dotnet` | `*.csproj` and `*.fsproj` projects | [dotnet](dotnet.md) |
| `ruby` | gems (`*.gemspec`) the root `Gemfile` lists | [ruby](ruby.md) |
| `php` | Composer packages the root `composer.json` requires | [php](php.md) |
| `cpp` | CMake projects (`CMakeLists.txt` that calls `project()`) | [cpp](cpp.md) |
| `dart` | packages of a root pub workspace, Dart or Flutter | [dart](dart.md) |
| `jvm` | Maven modules (Java) and Gradle modules (Kotlin, Android) | [jvm](jvm.md) |
| `bicep` | Azure infrastructure: each `main.bicep` | [bicep](bicep.md) |
| `lua` | LuaRocks projects: each `*.rockspec` | [lua](lua.md) |
| `applescript` | AppleScript: each directory with a `.applescript` source | [applescript](applescript.md) |
| `vba` | VBA modules exported as text (`.bas`, `.cls`, `.frm`) | [vba](vba.md) |
| `recipe` | a `*.recipe.toml` in a folder of its own | [recipe](recipe.md) |

## Toolchain, scaffold, structure and deploy plugins

These are always loaded and act on projects whose stack names them, or as a convention you
configure.

| Plugin | Roles | Provides | Page |
|---|---|---|---|
| `typescript` | scaffold, toolchain | `build`, `typecheck` | [typescript](typescript.md) |
| `eslint` | toolchain | `lint` | [eslint](eslint.md) |
| `jest` | toolchain | `test` | [jest](jest.md) |
| `vitest` | toolchain | `test` | [vitest](vitest.md) |
| `playwright` | toolchain | `e2e` | [playwright](playwright.md) |
| `vite` | scaffold, toolchain | `build`, `typecheck` | [vite](vite.md) |
| `angular` | scaffold, toolchain | `build`, `typecheck` | [angular](angular.md) |
| `astro` | scaffold, toolchain | `build` | [astro](astro.md) |
| `vscode` | scaffold, toolchain, deploy | `build`, `typecheck` | [vscode](vscode.md) |
| `vertical-slices` | structure | | [vertical-slices](vertical-slices.md) |
| `hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc`, `package-by-layer` | structure | | [architectures](architectures.md) |
| `docker` | deploy | | [docker](docker.md) |
| `go-smith-plugin` | scaffold | | [go-smith-plugin](go-smith-plugin.md) |

## What was verified, and where

Each page records the exact tool versions it was run with. In short, on Windows 11 with smith
1.62.0:

| Verified by running the toolchain | Derived from the code (marked "not verified" on the page) |
|---|---|
| `node` (Node 24.12, npm 11.6), `go` (1.27.0), `python` (3.12 through `py -3`), `rust` (1.76), `dotnet` (SDK 10.0.401), `ruby` (3.4.11), `php` (8.4.25), `cpp` (CMake 4.4, g++ 16, Ninja), `dart` (Dart 3.12, Flutter 3.44), `jvm` Maven (JDK 21, Maven 3.10), `bicep` (Azure CLI 2.90, Bicep 0.48), `recipe` | `jvm` Gradle, Kotlin and Android (no `gradle` installed); every real publish to a registry (npm, PyPI, crates.io, NuGet, RubyGems, Maven, pub.dev, Packagist, the Marketplace, container registries); a real `docker build` (the Docker daemon was not running); the Playwright browser download; a successful Angular build (the Node version was too old for the Angular CLI); Linux and macOS behaviour |

## See also

- [Using plugins](../README.md)
- [Creating plugins](../../creating/README.md): writing one
