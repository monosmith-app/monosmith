# The built-in plugin catalogue

smith ships 39 plugins inside its binary: 18 language plugins and 21 add-ons (toolchain, structure,
scaffold and deploy plugins, with the six architecture plugins counted one by one). This section has one page for each, written from the plugin's code
and from real runs: what it detects, the project kinds and tasks it adds, what it scaffolds, its
settings, how it handles the toolchain, how it releases, what it adds to CI, and where it is
limited. Every page ends with what was and was not verified.

The one-table summary is in [Plugins (reference)](../../../reference/plugins.md); the model behind
roles, stacks and conventions is in [What a plugin is and how plugins combine](../what-is-a-plugin.md).

## Language plugins

A language plugin finds projects from manifests and, by default, is also their toolchain and
scaffold. `smith plugin enable`, `disable` and `set` switch these on and off.

The last three columns are the columns of the same names in `smith plugin list`. **Declared**: the
plugin can describe a project that `smith.toml` declares with `[projects.<name>]` and no manifest. **Channels**: a
package of a `[[release.channels]]` branch can be published with it, a prerelease never replacing the stable version.
**Link**: `smith install -p <project> <library>` can make a project depend on a library of the workspace.
The values were read from `smith plugin list` of the smith 2.0 release candidate.

| Plugin | Projects it finds | Declared | Channels | Link | Page |
|---|---|---|---|---|---|
| `node` | `package.json` projects (npm workspaces); every script is a task | no | yes | yes | [node](node.md) |
| `go` | `go.mod` modules in a root `go.work` | yes | no | yes | [go](go.md) |
| `python` | `pyproject.toml` projects, one root `.venv` | yes | yes | yes | [python](python.md) |
| `rust` | `Cargo.toml` crates in a root Cargo workspace | yes | yes | yes | [rust](rust.md) |
| `dotnet` | `*.csproj` and `*.fsproj` projects | no | yes | yes | [dotnet](dotnet.md) |
| `ruby` | gems (`*.gemspec`) the root `Gemfile` lists | yes | yes | no | [ruby](ruby.md) |
| `php` | Composer packages the root `composer.json` requires | yes | yes | no | [php](php.md) |
| `cpp` | CMake projects (`CMakeLists.txt` that calls `project()`) | no | yes | no | [cpp](cpp.md) |
| `dart` | packages of a root pub workspace, Dart or Flutter | yes | yes | no | [dart](dart.md) |
| `jvm` | Maven modules (Java) and Gradle modules (Kotlin, Android) | no | yes | no | [jvm](jvm.md) |
| `bicep` | Azure infrastructure: each `main.bicep` | no | yes | no | [bicep](bicep.md) |
| `lua` | LuaRocks projects: each `*.rockspec` | yes | yes | no | [lua](lua.md) |
| `applescript` | AppleScript: each directory with a `.applescript` source | no | yes | no | [applescript](applescript.md) |
| `vba` | VBA modules exported as text (`.bas`, `.cls`, `.frm`) | no | yes | no | [vba](vba.md) |
| `swift` | Swift packages: each `Package.swift` | yes | yes | no | [swift](swift.md) |
| `objc` | CocoaPods libraries (Objective-C): each `*.podspec` | no | yes | no | [objc](objc.md) |
| `bundle` | apps with no language: a directory with `app.json` and `flow.json` | no | yes | no | [bundle](bundle.md) |
| `recipe` | a `*.recipe.toml` in a folder of its own | no | yes | no | [recipe](recipe.md) |

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
| `tauri` | toolchain | `tauri-build` | [tauri](tauri.md) |
| `vite` | scaffold, toolchain | `build`, `typecheck` | [vite](vite.md) |
| `functions` | scaffold | none | [functions](functions.md) |
| `electron` | scaffold | none | [Electron and Tauri](node.md#electron-and-tauri) (in the node page) |
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
