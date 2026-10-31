# The cpp plugin

The cpp plugin turns CMake projects into smith projects. A directory whose
`CMakeLists.txt` calls `project()` is a project; it is configured with Ninja
into `dist/<project dir>`, built with `cmake --build`, tested with CTest and
linted with cppcheck. Projects use each other with `add_subdirectory`, and
that is how smith knows one depends on another. The plugin also creates new C++
projects, writes release versions into the `project()` call, and tells CI to
install CMake, Ninja, a compiler and cppcheck.

## At a glance

| | |
|---|---|
| Plugin name | `cpp` (built in; runs as a separate process, `smith plugin serve cpp`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` values | `cpp`, `c++` |
| Manifest that makes a directory a candidate project | `CMakeLists.txt` (it is a project only if it calls `project()`) |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest`: `smith install`, `smith up` and `smith sync` do nothing for C++, and nothing is cleaned outside a removed project's directory |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app`, `lib`, `package` (`smith kinds` lists `cpp-app`, `cpp-lib`, `cpp-package`) |
| Tasks | `configure`, `build`, `test`, `lint`, `start` |

Roles and capabilities are what `smith plugin list --json` prints; languages,
manifests and the absence of requires, conflicts and provides come from the
plugin's `Describe` method (`libs/plugin-cpp/plugin/cpp_plugin_handler.go`).

## What it detects

smith hands the plugin every `CMakeLists.txt` in the files git lists. The
plugin reads each one with regular expressions (comments removed first), not
with CMake:

- **Is it a project.** Only a `CMakeLists.txt` with a plain `project(<name> ...)`
  call is. One without it is a subdirectory of a project and is skipped, and so
  is the one at the workspace root.
- **Project name.** The directory name (`libs/core` is `core`), not the CMake
  project name.
- **Exports.** The project's directory, for example `libs/core`.
- **Imports (graph edges).** Each `add_subdirectory(<dir> ...)` call, resolved
  to a workspace directory. A leading `${CMAKE_CURRENT_SOURCE_DIR}/`,
  `${CMAKE_CURRENT_LIST_DIR}/` or `${PROJECT_SOURCE_DIR}/` is understood; a
  source that still contains a variable, is absolute, leaves the workspace,
  or points inside the project's own directory is ignored.
- **Kind.** `set(SMITH_KIND "app")` (or `lib`, `package`) in the file; else the
  top directory (`apps/`, `libs/`, `packages/`); else `lib`.
- **Stack.** `set(SMITH_STACK "<name>")`, empty when absent.
- **Version.** The `VERSION` argument of `project()`. `cmake_minimum_required`'s
  `VERSION` is not mistaken for it.
- **Executable.** The first `add_executable(<target>` in the file, used by
  `start`.
- **Lint directories.** Whichever of `src`, `include` and `tests` exist.
- **Toolchain.** The first version in `cmake --version`, reported as
  `cmake 4.4.2`. The compiler is not part of it.

A real run, after `tool` added `core` with `add_subdirectory`:

```text
$ smith graph
core  lib · cpp · libs/core
sdk  package · cpp · packages/sdk
tool  app · cpp · apps/tool
  └─ core
```

## Tasks

All tasks run in the project's directory. `{workspaceRoot}/dist/<project dir>`
is the CMake build tree.

| Task | Command | Inputs | Outputs | dependsOn | Cached | Appears |
|---|---|---|---|---|---|---|
| `configure` | `cmake -S . -B {workspaceRoot}/dist/<dir> -G Ninja -DCMAKE_BUILD_TYPE=Debug` | every file of the project | none declared | none | no | every project |
| `build` | `cmake --build {workspaceRoot}/dist/<dir>` | every file of the project | none declared | `configure` | no | every project |
| `test` | `ctest --test-dir {workspaceRoot}/dist/<dir> --output-on-failure --no-tests=error` | `**/*`, `!build/**` | none | `build` | yes | every project |
| `lint` | `cppcheck --enable=warning,style,performance,portability --error-exitcode=1 --inline-suppr --quiet --suppress=missingIncludeSystem --std=c++20 -I include <src include tests>` | `**/*`, `!build/**` | none | none | yes | when `src`, `include` or `tests` exists (only the existing ones are passed) |
| `start` | `{workspaceRoot}/dist/<dir>/<executable>` (`.exe` on Windows) | every file of the project | none | `build` | no | kind `app` with an `add_executable` |

Notes from the real run:

- `build` and `configure` are not cached: CMake and Ninja do their own
  incremental build in the persistent tree under `dist/`. A second
  `smith run build` printed `core:build ✔ 86ms`.
- `test` is cached, so an unchanged project's second `smith run test` is
  restored.
- With the `vertical-slices` convention (on in a new workspace) each project
  also has `slice-check`, and `lint` depends on it. That task belongs to the
  convention.
- The test task needs at least one test: `--no-tests=error` fails a project
  that registers none.

```text
$ smith explain tool
tool has 6 tasks:
  build        cmake --build {workspaceRoot}/dist/apps/tool  (after configure)
  configure    cmake -S . -B {workspaceRoot}/dist/apps/tool -G Ninja -DCMAKE_BUILD_TYPE=Debug
  lint         cppcheck --enable=warning,style,performance,portability --error-exitcode=1 --in…  (cached, after slice-check)
  slice-check  smith slices check --lang cpp .  (cached)
  start        {workspaceRoot}/dist/apps/tool/cpp_demo_tool.exe  (after build)
  test         ctest --test-dir {workspaceRoot}/dist/apps/tool --output-on-failure --no-tests=…  (cached, after build)
```

```text
$ smith run start -p tool
tool:start ▶ {workspaceRoot}/dist/apps/tool/cpp_demo_tool.exe
tool:start │ Hello, world!
```

## Creating a project

```sh
smith add lib core --lang cpp
smith add app tool --lang cpp
smith add package sdk --lang cpp
```

Real output for the first two:

```text
$ smith add lib core --lang cpp
Adding lib core (cpp) in libs/core
Wrote 7 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  libs/core/CMakeLists.txt
  libs/core/include/cpp_demo_core/greeting.hpp
  libs/core/include/cpp_demo_core/greeting/greet_use_case.hpp
  libs/core/src/greeting/greet_use_case.cpp
  libs/core/tests/greeting/greet_use_case_test.cpp
Use it from another project with add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../libs/core cpp_demo_core) and target_link_libraries(<target> PRIVATE cpp_demo_core).
Try: smith test -p core

$ smith add app tool --lang cpp
Adding app tool (cpp) in apps/tool
Wrote 6 file(s):
  apps/tool/CMakeLists.txt
  apps/tool/include/cpp_demo_tool/greeting.hpp
  apps/tool/include/cpp_demo_tool/greeting/greet_use_case.hpp
  apps/tool/src/greeting/greet_use_case.cpp
  apps/tool/src/main.cpp
  apps/tool/tests/greeting/greet_use_case_test.cpp
Use it from another project with add_subdirectory(${CMAKE_CURRENT_SOURCE_DIR}/../../apps/tool cpp_demo_tool) and target_link_libraries(<target> PRIVATE cpp_demo_tool_lib).
Try: smith test -p tool
```

A package has the same files as a lib (5 files). The names:

- The **target** is `<workspace>_<project>` in lower-case with other
  characters turned into underscores (`cpp_demo_core`), so two projects added
  together never clash. It is also the C++ namespace.
- For an **app**, the code is a library named `<target>_lib` and the
  executable `<target>` links it; the tests link the library too.
- The generated `CMakeLists.txt` requires CMake 3.25, sets
  `project(<target> VERSION 0.0.0 LANGUAGES CXX)`, `set(SMITH_KIND "app")` (and
  `SMITH_STACK` when one was given), builds a static library from
  `src/greeting/greet_use_case.cpp` with C++20 and a public `include/`
  directory, and, only `if(PROJECT_IS_TOP_LEVEL)`, enables testing and adds a
  test executable. Tests therefore do not build when another project adds
  this one with `add_subdirectory`.
- A `vcpkg` block at the top: if a `vcpkg.json` sits beside the file,
  `VCPKG_ROOT` is set in the environment and no toolchain file is chosen, CMake
  uses `$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake`.
- The worked example is the greeting slice: `include/<target>/greeting/greet_use_case.hpp`
  declares `std::string greet(std::string_view name)` in namespace
  `<target>::greeting`, the `.cpp` returns `Hello, <name>!`, a `greeting.hpp`
  umbrella header includes it, and the test checks the result and returns
  `EXIT_FAILURE` if it differs. An app's `src/main.cpp` prints
  `Hello, world!`.

`smith add` fails if the directory exists. No tool is run, so it needs no
compiler at creation time. The generated `.gitignore` already ignores `dist/`,
the build tree's home. `--lang c++` is accepted as an alias (I created `libs/alt`
with it).

## Settings

The plugin has no `[projects.<p>.settings.cpp]` keys, no environment
variables of its own, and no plugin tables in `smith.toml`.

| Where | Key | Effect | Read in |
|---|---|---|---|
| `CMakeLists.txt` | `set(SMITH_KIND "app")` | the project's kind | `discovery/cmake_lists_mapper.go` |
| `CMakeLists.txt` | `set(SMITH_STACK "<name>")` | the project's stack | same |
| `CMakeLists.txt` | `set(SMITH_NATIVE ON)` | the app is built on each OS and its release archives come from the native jobs | same |
| `smith.toml` | `[projects.<app>.release] platforms`, `asset_name`, `product`, `os_alias`, `arch_alias` | which archive the release builds and what it is called | host, `releasing/build_assets_use_case.go` |
| `CMakeLists.txt` | `project(... VERSION x.y.z)` | the release version it starts from and rewrites | `releasing/project_version_mapper.go` |
| `CMakeLists.txt` | `add_subdirectory(<workspace dir>)` | a graph edge | `discovery/cmake_lists_mapper.go` |
| `CMakeLists.txt` | first `add_executable(<target>` | what `start` runs | same |
| `vcpkg.json` beside `CMakeLists.txt` | presence | adds the vcpkg toolchain and doctor check, vcpkg in CI | `pipeline/ci_contribution_use_case.go` |
| environment | `VCPKG_ROOT` | where vcpkg is; read by the CMake block and by doctor | `pipeline/`, generated `CMakeLists.txt` |
| `smith.toml` | `[toolchains] cmake = "3.25"` | doctor pin | host |
| `smith.toml` | `[tasks.<task>]`, `[projects.<p>.tasks.<task>]` | override an inferred task, for example `[tasks.configure]` to change the generator | host ([tasks and cache](../../../reference/tasks-and-cache.md)) |

The build type (`Debug`), generator (`Ninja`) and `-std=c++20` for cppcheck
are fixed in the plugin; override them with a task override.

## Toolchain

- **Probe.** `cmake --version`, reported as `cmake <version>`. smith pins and
  caches by CMake's version only; the compiler and Ninja versions are not
  probed.
- **Pin.** `[toolchains] cmake = "3.28"` is checked by `smith doctor`. A real
  run on CMake 4.4.2:

```text
✖ cmake matches the pinned version 3.28
    [toolchains] pins cmake 3.28, but this machine runs 4.4.2: tasks cached on another version are misses, and results can differ from CI
    → install cmake 3.28 (`mise use cmake@3.28`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] cmake
```

- **CI ignores the pin.** CI installs whatever the distribution packages
  (`Version: "system"`).
- **Needed on the machine.** `cmake`, `ninja`, a C++ compiler (`c++`, `g++`,
  `clang++` or `cl`), and `cppcheck`.
- **Where output goes.** The build tree is `dist/<project dir>` (for example
  `dist/apps/tool`), with `build.ninja`, `CMakeCache.txt`, the static library,
  the executables and CTest's `Testing/`. The tasks declare no `outputs`, and
  `smith doctor` still passed its output checks (`every build writes to dist/`,
  `every build declares the directories it writes as outputs`).

## Releasing

| | |
|---|---|
| Version in the manifest | the `VERSION` argument of `project()` in `CMakeLists.txt` |
| Version write | at release time the number is replaced in place; with no `VERSION` argument one is inserted right after the project name. The file is restored afterwards. A dry run lists `packages/sdk/CMakeLists.txt` and writes nothing |
| Tag | `<project>@<version>` (for example `sdk@0.1.0`); override with `[projects.<name>] tag_prefix` |
| Publish | nothing is uploaded for any kind: "released by its tag". The consumer chooses a vcpkg registry or Conan |
| Release assets of an app | With `[projects.<app>.release]` the plugin runs `cmake -S . -B dist/release/<app>/build -G Ninja -DCMAKE_BUILD_TYPE=Release`, `cmake --build` and `cmake --install --prefix dist/release/<app>/stage`, and writes `dist/release/<app>/<asset_name>` as a zip of what the install rules installed (or, with no install rules, the program of the first `add_executable`, named like the app's `Binary`), plus the README and LICENSE files of the project's directory, else the workspace root, and a checksum file. The version is the one the release wrote into `project(... VERSION)` (use `PROJECT_VERSION` in the code to show it). Each app of a workspace is its own asset set (`product` names it). |
| Platforms | C++ is compiled and linked for the system it runs on, so an app is built on each system. A plain app can release only this machine's platform; another is refused, in a dry run too: `this machine (linux/amd64) cannot build darwin/arm64 ... Mark the app native (set(SMITH_NATIVE ON) in apps/tool/CMakeLists.txt) ...`. A **native** app (`set(SMITH_NATIVE ON)`) is built by each `smith ci native` runner (`ubuntu-latest`, `windows-latest`, `macos-latest`; `[ci] native_runners` adds others such as `ubuntu-24.04-arm`) for its own platform only, and attaches `<asset>.zip` and `<asset>.zip.sha256` to the GitHub Release; `smith ci verify` leaves it to those jobs. A native app has no combined `checksums.txt` (builds from different runners), no package manager manifests and no registry packages. A plain app's one platform gets `checksums.txt`, and feeds those like a Go app's. Its dry run lists what this machine's native job would attach. |
| Credentials | none |
| Private registry | not applicable |
| Dry run | `smith release --dry-run` plans without tagging |

```text
$ smith release --dry-run
Dry run: nothing is published or tagged.
  sdk                  unreleased → 0.1.0  (1 releasing commit(s))
```

The JSON report for `sdk` shows `"tag": "sdk@0.1.0"`,
`"files": ["packages/sdk/CMakeLists.txt"]` and
`"publish": "released by its tag"`.

The release's SBOM lists the packages of `vcpkg.json` and `conanfile.txt`; a project with neither gets an SBOM that says it declares no dependency, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

| Provider | What is added |
|---|---|
| GitHub Actions, Azure Pipelines, GitLab | a step `Install cpp` (bash on every runner) that, on Linux, runs `apt-get install -y -qq cmake ninja-build g++ cppcheck git curl zip unzip tar pkg-config` (with `sudo` when present), on macOS `brew install cmake ninja cppcheck pkg-config`, and on Windows nothing (the runner has CMake, Ninja and MinGW) |
| Dev container | the `ghcr.io/devcontainers-extra/features/apt-get-packages:1` feature with `cmake,ninja-build,g++,cppcheck` |
| vcpkg | when any project has a `vcpkg.json`, a step `Install vcpkg` clones vcpkg, bootstraps it and exports `VCPKG_ROOT` (GitHub's `GITHUB_ENV` and Azure's `##vso[task.setvariable]` ways). Not in the dev container |

There is no setup step, no audit step, no package-manager store, no secret and
no permission: a C++ release is the tag. The GitHub fragment from a real run:

```yaml
      - name: Install cpp
        shell: bash
        run: |
          case "$(uname -s)" in
            Linux)
              SUDO=$(command -v sudo || true)
              $SUDO apt-get update -qq && $SUDO apt-get install -y -qq cmake ninja-build g++ cppcheck git curl zip unzip tar pkg-config ;;
            Darwin)
              brew install cmake ninja cppcheck pkg-config ;;
          esac
```

(The `case` and `shell: bash` are new with the native release jobs; the fragment above is the older one-line form.) The CI images are assumed to be Debian or Ubuntu; a native app's macOS and Windows runners use the script's other branches. Not verified: GitLab and
Azure output for cpp (generated from the same contribution, not run).

## Doctor, dependencies and removal

**Doctor.** Run only when the workspace has C++ projects.

| Check | Fails when |
|---|---|
| `CMake is installed` | `cmake --version` fails |
| `Ninja is installed` | `ninja --version` fails |
| `a C++ compiler is installed` | none of `c++`, `g++`, `clang++`, `cl` is on `PATH` |
| `cppcheck is installed` | `cppcheck --version` fails |
| `<app> releases for <platform>` | a platform in `[projects.<app>.release]` is not windows, linux or darwin with amd64 or arm64; remedy: change `platforms` |
| `<app> can build the <platform> archive here` | the platform is not this machine's and the app is not native; remedy: `set(SMITH_NATIVE ON)` in its `CMakeLists.txt` |
| `<app> has a program to release` | the app's `CMakeLists.txt` has no `add_executable` and no `install(` rule, so the archive would be empty |
| `vcpkg is set up` | a project has a `vcpkg.json` and `VCPKG_ROOT` is empty. Real output: `a project has a vcpkg.json, but VCPKG_ROOT is not set` |

All four passed on this machine; the vcpkg check failed as intended with an
empty `vcpkg.json` and no `VCPKG_ROOT`.

**Dependencies.** None: C++ projects do not advertise `dependencies`,
`declarations` or `latest`, so `smith install` ran silently, and
`smith sync --check` found nothing to compare. Libraries come from vcpkg or
other workspace projects, which you manage in `CMakeLists.txt` and
`vcpkg.json` yourself.

**Remove.** `smith remove <project>` deletes the directory and regenerates the
pipelines (`--dry-run` said `Would remove tool (apps/tool)` and
`the pipelines would be regenerated`). The plugin cleans nothing else. A
project that another still adds with `add_subdirectory` is refused unless
`--force`.

## Requirements and limits

- CMake 3.25 or newer for new projects, Ninja, a compiler, cppcheck.
- A C++ project is found by regular expressions: a `project()` call built by
  a macro, an `add_subdirectory` through a variable, or an `add_executable`
  generated by a function is not seen.
- `smith add` generates C++20 code only (`cxx_std_20`, `LANGUAGES CXX`), with no stacks.
- No `typecheck`, `format` or `package` tasks. `build` is not cached. Release archives are Release builds with Ninja and the compiler CMake finds (no cross-compiling, no multi-config generators), zip only, one platform per runner.
- The Windows executable path in `start` gets `.exe` appended on Windows only;
  multi-config generators (Visual Studio) are not supported because the
  configure step always passes `-G Ninja`.
- The second and later `smith add` calls leave the CI files untouched when
  they are already right (the first listed them in `Wrote 7 file(s)`).

## Verified

Windows 11 Pro, smith 1.62.0, CMake 4.4.2, Ninja 1.13.2, g++ 16.2.0 (WinLibs
MinGW-W64 UCRT), Cppcheck 2.21.0.

Run: `smith new`, `smith add lib|app|package --lang cpp`, `--lang c++`,
`smith projects`, `smith graph` (with a real `add_subdirectory`),
`smith explain` for each task, `smith run build`, `smith run test lint`,
`smith run start -p tool`, `smith doctor` (all checks, a CMake pin mismatch,
a `vcpkg.json` without `VCPKG_ROOT`), `smith ci generate` for GitHub (with and
without vcpkg), `smith release --dry-run` and `--json`, `smith remove --dry-run`,
`smith install`, `smith sync --check`.

Not verified: a real vcpkg manifest build; Linux or macOS runs; GitLab and Azure
output; a real (non-dry) release.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Tasks and cache](../../../reference/tasks-and-cache.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
- [The vertical-slices plugin](vertical-slices.md)
