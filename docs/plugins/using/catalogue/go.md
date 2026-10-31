# The go plugin

The go plugin makes every `go.mod` below the workspace root a smith project. It gives each module the commands you would type yourself (`go build`, `go test`, and `golangci-lint` with a default configuration smith supplies), adds the module to `go.work` when it creates one, keeps `go.work` honest in `smith doctor`, installs Go and a module cache in CI, scans modules for known vulnerabilities, and releases Go packages by tag and Go apps as zipped binaries. It is a built-in plugin: it needs no installation and runs in-process with smith.

## At a glance

| | |
|---|---|
| Plugin name | `go` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold`. A language plugin gets all three by default; none of them has to be switched on. |
| `--lang` values | `go`, `golang` |
| Manifest that makes a directory a project | `go.mod` (a `go.mod` at the workspace root is skipped; see [What it detects](#what-it-detects)) |
| Requires / conflicts / provides | none / none / no task slots (`smith plugin list` shows `-` in the SLOTS and REQUIRES columns) |
| Capabilities advertised | `discover`, `generate`, `publish`, `ci`, `doctor`, `remove`, `dependencies`, `declarations`, `latest`. Not advertised: `version`, `published`. |
| Stacks it publishes | none. The add-on [go-smith-plugin](go-smith-plugin.md) publishes the one stack whose language is `go`. |
| Project kinds it can create | `app`, `lib`, `package` (`go-app`, `go-lib`, `go-package` in `smith kinds`) |

The capability list is the output of `smith plugin list --json`. `version` is absent because a Go module has no version in its manifest: its version is a git tag.

## What it detects

smith hands the plugin every `go.mod` it found. For each one that is not at the workspace root the plugin reads the file and the Go sources next to it (`libs/plugin-go/discovery/`).

| Field | How it is read |
|---|---|
| Name | the last element of the directory: `libs/util` is `util` |
| Kind | 1. a `// smith:kind app\|lib\|package` comment line in `go.mod`; 2. else the top directory (`apps/` is an app, `libs/` a lib, `packages/` a package); 3. else an app if a non-test `.go` file directly in the module directory declares `package main`, otherwise a lib. An unknown value in the comment is ignored and the next rule applies. |
| Exports | the module path from the `module` line (quoted or not) |
| Imports | every import path of every `.go` file under the module, parsed with `go/parser` in imports-only mode, sorted and unique. smith matches them against other projects' exports to draw graph edges. |
| Stack | a `// smith:stack <name>` comment line in `go.mod` |
| Tags | `native` when `go.mod` has a `// smith:native` line |
| Release tag prefix | for kind `package` only: `<directory>/v`, for example `packages/sdk/v` |
| Toolchain | `go <version>`, from `go version`; left out when `go` is not on `PATH` |

The import scan skips `vendor`, `testdata`, `dist` and `node_modules`, any directory starting with `.` or `_`, and any directory that holds its own `go.mod` (a nested module is its own project). A file that does not parse is ignored; `go build` reports it.

A `go.mod` at the workspace root is treated as a single-module repository, not as a project of a monorepo, and is not discovered.

Comment lines must start at the beginning of a line, as the generator writes them:

```
// smith:kind app
module example.com/go-demo/apps/api

go 1.27.0
```

Real output from the workspace built in [Verified](#verified):

```
$ smith projects
api                  app      go       apps/api
sdk                  package  go       packages/sdk
util                 lib      go       libs/util

$ smith graph          # after apps/api imports example.com/go-demo/libs/util/greeting
api  app · go · apps/api
  └─ util
sdk  package · go · packages/sdk
util  lib · go · libs/util
```

## Tasks

The plugin contributes these tasks (`libs/plugin-go/discovery/module_tasks_config.go`). All of them run in the module directory. Go builds a module's dependencies itself, so no task has `dependsOn`; the cache key still follows the sources of every project the module imports.

| Task | Command | Appears for | Cached | Outputs |
|---|---|---|---|---|
| `build` | `go build -trimpath -o {workspaceRoot}/dist/<dir>/<name>[.exe] .` | app | yes | `dist/<dir>` (for `apps/api`: `dist/apps/api`) |
| `build` | `go build ./...` | lib, package | yes | none (compiles, writes nothing) |
| `test` | `go test ./...` | every module | yes | none |
| `lint` | `golangci-lint run --allow-serial-runners --config {workspaceRoot}/.smith/defaults/golangci.yml` (no `--config` when the repository has its own, see [Lint configuration](#lint-configuration)) | every module | yes | none |
| `deps-check` | `smith go-deps-check <dir>` | package; its `lint` (or, without one, `build` or `test`) runs after it, so CI verify runs it | yes | none |
| `start` | `go run .` | app | no | none |
| `package` | `go build -trimpath -o {workspaceRoot}/dist/drop/<name>/<name>[.exe] .` | app | no | none declared; the binary lands in `dist/drop/<name>/` and `smith ci pack` keeps it |

`.exe` is added to the binary name only when smith itself runs on Windows. Every cached task has the inputs `**/*`, `{workspaceRoot}/go.work` and `{workspaceRoot}/go.work.sum`. A `native` module gets `CGO_ENABLED=1` in the environment of every task.

`deps-check` fails when a file of the module imports a package from a module its `go.mod` does not require. `go.work` merges the build lists of every module in the workspace, so a module can import a package that only another module requires and still build here, while whoever fetches the module on its own gets "no required module provides package". Never reported: the standard library, the module's own packages (and its nested modules' directories, which are skipped, as are `vendor/`, `testdata/`, and directories starting with `.` or `_`), files marked `//go:build ignore`, the modules of the `go.work` `use` directives, and modules a `replace` directive (in `go.mod` or `go.work`) names. A requirement marked `// indirect` counts as declared, versions are not compared, and a requirement nothing imports is not reported. A file that does not parse is left to the compiler. The output lines are `<file> imports <package>, which go.mod does not require`; the fix is `go mod tidy` in the module.

There is no `typecheck` and no `format` task: `smith run typecheck` and `smith run format` answer `no selected project has a [typecheck] task` (and `[format]`). `golangci-lint` is the lint (it includes `go vet`); the compiler is the type check inside `build`.

A convention adds tasks on top. With `[conventions.vertical-slices]` (written by `smith new`) every Go module also gets `slice-check`, and `lint` depends on it; see [vertical-slices](vertical-slices.md). Real `smith explain` output:

```
$ smith explain api
api has 6 tasks:
  build        go build -trimpath -o {workspaceRoot}/dist/apps/api/api.exe .  (cached)
  lint         golangci-lint run --allow-serial-runners --config {workspaceRoot}/.smith/defaults/golangci.yml  (cached, after slice-check)
  package      go build -trimpath -o {workspaceRoot}/dist/drop/api/api.exe .
  slice-check  smith slices check --lang go .  (cached)
  start        go run .
  test         go test ./...  (cached)

$ smith explain api:build
api:build — Build the binary into dist/apps/api/
  command    go build -trimpath -o {workspaceRoot}/dist/apps/api/api.exe .
  cwd        apps/api
  dependsOn  —
  inputs     **/*, {workspaceRoot}/go.work, {workspaceRoot}/go.work.sum
  outputs    ../../dist/apps/api
  env        —
  cache      true

$ smith explain sdk:build
sdk:build — Compile every package
  command    go build ./...
  cwd        packages/sdk
  ...
  cache      true
```

To change any of them for one project, or for all, use `[projects.<name>.tasks.<task>]` or `[tasks.<task>]` in `smith.toml` (see [tasks and cache](../../../reference/tasks-and-cache.md)). A `lint` command that starts with `golangci-lint` (the default) is what makes CI install it; see [CI](#ci).

### Lint configuration

smith owns the lint of a Go repository, so the repository need not carry a linter configuration. When no `.golangci.yml`, `.golangci.yaml`, `.golangci.toml` or `.golangci.json` exists in the module or in a directory above it (up to the workspace root), smith writes its default to `.smith/defaults/golangci.yml` (git ignores `.smith/`; the file is rewritten whenever it differs from the one in the smith binary) and passes it with `--config`. The default is golangci-lint v2 with its standard linters and `errcheck` told that `fmt.Fprint`, `fmt.Fprintf` and `fmt.Fprintln` need no error check.

Your own file is the override and replaces the default whole, because golangci-lint has no way to extend another file. When one exists smith passes no `--config` and golangci-lint finds it, as it would without smith. To start from the default, run [`smith eject golangci-lint`](../../../manual/eject.md): it copies the default to `.golangci.yml` in the workspace root, to edit. The workspace-root file is an input of the cached task, so editing it reruns the lint.

## Creating a project

```
smith add app|lib|package <name> --lang go
```

`smith add` writes a Go module with one worked example, a `greeting` slice (a package, a use case, a contract and a test), `go.mod` marked with its kind, and an entry in `go.work`. The module path is `[workspace] module_prefix` plus `/` plus the directory; with no prefix it is `example.com/<workspace name, lower-cased>`. Under the hood the plugin runs `go work init` (only when there is no `go.work`), `go mod init`, `go mod edit -go=<the go.work version>` (so a module takes the version the workspace builds with, not that of the Go that happens to run) and `go work use ./<dir>`. It refuses a directory that already exists.

Real runs:

```
$ smith add lib util --lang go
Adding lib util (go) in libs/util
go: creating new go.mod: module example.com/go-demo/libs/util
Wrote 10 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  azure-pipelines.yml
  go.work
  libs/util/doc.go
  libs/util/go.mod
  libs/util/greeting/doc.go
  libs/util/greeting/greet_use_case.go
  libs/util/greeting/greet_use_case_test.go
  libs/util/greeting/greeting_contract.go
Import it from another Go project as example.com/go-demo/libs/util/<slice>.
Try: smith test -p util

$ smith add app api --lang go
Wrote 8 file(s):
  .github/workflows/ci.yml
  apps/api/go.mod
  apps/api/greeting/doc.go
  apps/api/greeting/greet_use_case.go
  apps/api/greeting/greet_use_case_test.go
  apps/api/greeting/greeting_contract.go
  apps/api/main.go
  azure-pipelines.yml
```

The generated code:

| File | What it is |
|---|---|
| `<dir>/go.mod` | the module, with `// smith:kind <kind>` (and `// smith:stack <stack>` when `--stack` names one) as its first lines |
| `<dir>/greeting/greeting_contract.go` | `type Greeting struct{ Message string }` |
| `<dir>/greeting/greet_use_case.go` | `func Greet(name string) Greeting`, returning `Hello, <name>!` |
| `<dir>/greeting/greet_use_case_test.go` | a test of `Greet` |
| `<dir>/greeting/doc.go` | package comment |
| `<dir>/main.go` (app only) | prints `Greet("world")` |
| `<dir>/doc.go` (lib and package only) | package comment for the module root |

A `package` is generated exactly like a `lib`; the difference is the kind in `go.mod` and, therefore, release behaviour. Files in `<dir>/greeting` follow the vertical-slice naming of the [vertical-slices](vertical-slices.md) convention (`greet_use_case.go`), so the generated project passes `slice-check`. The new project builds, tests and lints at once (`smith run build`, `test`, `lint` succeeded, see [Verified](#verified)).

The plugin publishes no stack of its own. `--stack` for Go means `go-smith-plugin` (see [go-smith-plugin](go-smith-plugin.md)), or a stack you declare in `smith.toml` with `language = "go"`.

## Settings

The Go plugin has no `[projects.<p>.settings.go]` table and reads no environment variable of its own. What can be configured:

| Where | Key | Effect | Read in |
|---|---|---|---|
| `go.mod` comment | `// smith:kind app\|lib\|package` | the project kind | `discovery/go_module_mapper.go` |
| `go.mod` comment | `// smith:stack <name>` | the stack (letters, digits, `. @ / - _`) | same |
| `go.mod` comment | `// smith:native` | the module builds on each OS with `CGO_ENABLED=1`, and `smith ci verify` leaves it to `smith ci native` | same, `discovery/module_tasks_config.go` |
| `go.work` | `go` directive | the Go version CI installs, and the minimum `smith doctor` demands | `pipeline/go_work_mapper.go` |
| `smith.toml` | `[workspace] module_prefix` | module path prefix of new modules | `scaffolding/generate_module_use_case.go` |
| `smith.toml` | `[toolchains] go = "1.27"` | version pin checked by `smith doctor` | host (see [Toolchain](#toolchain)) |
| `smith.toml` | `[projects.<app>.release]` | release assets (`platforms`, `asset_name`, `product`, `os_alias`, `arch_alias`, `version_variable`, ...) | `releasing/` |
| `smith.toml` | `[ci] audit_policies = ["go"]`, `audit_warn = ["go"]` | adds `go mod verify` to the audit | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[tasks.*]`, `[projects.<p>.tasks.*]` | override any task above | host |

## Toolchain

The plugin reports the toolchain label `go` with the version from `go version`, once per process. That string is part of every task's cache key, so a machine on another Go version never reuses a result.

To pin it, put the version in `[toolchains]` (a prefix on dot boundaries). Real output with `go = "1.23"` on a machine running 1.27.0:

```
✖ go matches the pinned version 1.23
    [toolchains] pins go 1.23, but this machine runs 1.27.0: tasks cached on another version are misses, and results can differ from CI
    → install go 1.23 (`mise use go@1.23`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] go
```

smith does not download Go. What must be on the machine: `go`. Separately from the pin, the Go version that CI installs and that `smith doctor` requires comes from `go.work` (default `1.24` when `go.work` has no `go` line).

Where output goes: an app's `build` writes `dist/<module dir>/<name>[.exe]` (outside the module, so `go test ./...` and `golangci-lint run` never walk it); `package` writes `dist/drop/<name>/`; a release writes `dist/release/<app>/`. A lib or package build writes nothing.

## Releasing

Go has no manifest version, so the plugin has no `version` capability and nothing is written into `go.mod`.

| Kind | What a release does |
|---|---|
| package | smith computes the next version from conventional commits, creates the tag `<dir>/v<version>` (for example `packages/sdk/v0.1.0`) and pushes it. That tag is how `go get` finds a nested module's version. The plugin's publish step answers "published by its tag": there is no upload, no registry and no credential. |
| app with `[projects.<name>.release]` | tagged `<name>@<version>`, and the plugin cross-compiles it: for each platform `go build -trimpath -ldflags "-s -w -X main.version=<version>" -o ... .` with `CGO_ENABLED=0`, zips the binary at the archive root, and writes `checksums.txt` (SHA-256), all under `dist/release/<app>/`. smith attaches them to the GitHub Release. |
| app without a release section | not released. |
| lib | never released. |

Asset naming defaults to `{product}_{version}_{os}_{arch}.{ext}`; `{product}` is the project name unless `product` is set; `os_alias` and `arch_alias` rename `os` and `arch` in the file name only; an unknown placeholder is an error. `version_variable` (default `main.version`) is the symbol stamped with `-X`. For a `native` app the archives are not checksummed together: each OS builds only its own platform and `checksums.txt` is skipped; asking for another platform fails with `<app> is native (cgo): build <platform> on a <platform> runner (smith ci native)`.

Real output of a dry run (two commits in the demo workspace, one `[projects.api.release]` with `platforms = ["linux/amd64", "windows/amd64"]`):

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  api                  unreleased → 0.1.0  (1 releasing commit(s))
  sdk                  unreleased → 0.1.0  (1 releasing commit(s))
api: would attach dist/release/api/api_0.1.0_linux_amd64.zip, dist/release/api/api_0.1.0_windows_amd64.zip, dist/release/api/checksums.txt

$ smith release --dry-run --json      # excerpt for the package
  "project": "sdk", "plugin": "go", "next": "0.1.0", "level": "minor",
  "tag": "packages/sdk/v0.1.0", "publish": "published by its tag"
```

A dry run names the assets but builds nothing. There is no registry to configure: for a Go module the tag is the registry (releasing.md: "a Go module's tags are its registry"). Credentials: none for the module tag beyond push rights to `[release] remote`; `GITHUB_TOKEN` when `github_releases = true` or when an app attaches assets (see [releasing](../../../reference/releasing.md)).

The release's SBOM lists the modules `go.mod` requires, direct and indirect, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

`smith ci generate` writes the toolchain and cache steps from the plugin's `ci` method. With three modules, `go.work` saying `go 1.27.0`, real output:

GitHub (`.github/workflows/ci.yml`):

```yaml
      - uses: actions/setup-go@v7
        with:
          go-version: '1.27'
          cache: false
      - name: Cache the go download store
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: smith-go-${{ runner.os }}-${{ hashFiles('apps/api/go.sum', 'packages/sdk/go.sum', 'libs/util/go.sum') }}
```

Azure (`azure-pipelines.yml`):

```yaml
variables:
  GOMODCACHE: $(Pipeline.Workspace)/.smith-stores/go
  ...
  - task: GoTool@0
    inputs:
      version: '1.27.0'
```

| Contribution | What |
|---|---|
| Toolchain | Go at the `go.work` version (`1.24` when none): `actions/setup-go@v7` (with `cache: false`; the module cache is handled as a store), `GoTool@0` on Azure (a `major.minor` version gets `.0` appended), a `curl` of the release tarball for other providers, and the `go` dev container feature with `golangciLintVersion` set to the pinned version |
| Package store | `~/go/pkg/mod` (`GOMODCACHE` on Azure), keyed on each module's `go.sum` |
| Setup (`smith ci setup`) | `go mod download`, blocking. Real: `▶ Download Go modules` then `smith ci setup passed in 100ms`. While any module's `lint` command starts with `golangci-lint` (the default), also `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0` |
| Audit (`smith ci audit`) | `govulncheck` v1.9.0 (`go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...`), once per module, non-blocking, cached on that module's `go.mod` and `go.sum`. With `[ci] audit_policies = ["go"]` also `go mod verify` per module, blocking unless `"go"` is in `audit_warn` |
| Credentials, permissions | none; the plugin asks for no secret. The workflow-level `permissions` come from smith, not from this plugin. |

Real audit run: `smith ci audit` printed `No vulnerabilities found.` for each module and `smith ci audit passed in 5.9s`. A `native` module is built by the extra `native` job on each OS (`smith ci native`), see [CI](../../../reference/ci.md#native-projects).

## Doctor, dependencies and removal

Doctor checks (`pipeline/check_go_work_use_case.go`, `check_tool_versions_use_case.go`):

| Check | Fails when | Fix |
|---|---|---|
| `go.work exists` | the workspace has Go modules and no `go.work` | run `go work init` and `go work use` |
| `go.work entry <dir> exists` | a `use` entry has no `go.mod` | `go work edit -dropuse=./<dir>`; `smith doctor --fix` rewrites `go.work` |
| `go.work lists <dir>` | a Go project is not in `go.work` | `go work use ./<dir>`; `--fix` rewrites `go.work` |
| `<dir> requires the modules it imports` | a package (`kind = package`) fails `deps-check`; no automatic fix, `go mod tidy` in the module | `go mod tidy` |
| `Go is at least <v> (go.work)` | the installed Go is older than the `go.work` version (`go env GOVERSION`); also fails when `go` is not on PATH | install that Go |
| `golangci-lint is <v>, as in CI` | while a `lint` command starts with `golangci-lint` (the default); a warning, not a failure, when the installed version is not the pinned one (2.14.0) | `go install ...@v2.14.0` |

Real output after `go work edit -dropuse=./libs/util`:

```
✖ go.work lists libs/util
    libs/util has a go.mod but go.work does not use it
    → run `go work use ./libs/util`

19 passed, 1 failed, 0 warning(s)
1 fix(es) available: `smith doctor --fix`
```

`smith doctor --fix` printed a diff adding `./libs/util` to `use` and asked for `--yes`.

Dependencies:

| Command | What the plugin runs |
|---|---|
| `smith install` (no package) | `go mod download` in every module |
| `smith install -w <project> <module>[@v]` | `go get <module>` in that module; with `-D`, `go get -tool` (a Go tool dependency) |
| `smith uninstall -w <project> <module>` | `go get <module>@none` (`go get -tool ...@none` when `go.mod` has a `tool` directive for it) |
| `smith up` | asks the module proxy (`proxy.golang.org`) for the newest release of each module a `go.mod` requires directly, then runs `go get <module>@v<version>` in each module that declares it, so `go.mod` and `go.sum` change together. It also compares the pinned tools: Go and golangci-lint. Indirect requirements, pseudo-versions (a commit between tags) and `replace`d modules are left alone. |
| `smith sync` | brings the direct requirements of one module that several `go.mod` files declare to the highest version any of them has, with `go get`. A `go.work` already builds every module with the highest version required, so this does not change what ships; it keeps each module building the same on its own, and makes `smith doctor`'s one-version check meaningful for Go. |

Real: `smith install -w api github.com/pkg/errors@v0.9.1` printed `added github.com/pkg/errors@v0.9.1 (apps/api/go.mod, apps/api/go.sum)` and the note `go mod tidy drops a requirement nothing imports yet: import it before tidying`; the `go.mod` line was marked `// indirect` until code imports it. `smith uninstall` removed it again. `smith up --check` ended with:

```
Tools smith pins (informational):
  Go               1.27 ✔ (latest 1.27.3)
  golangci-lint    2.14.0 → 2.15.0 available
```

(Those figures are read from `go.dev/dl` and `proxy.golang.org` at run time; the plugin tolerates a failed lookup by leaving "latest" empty.)

Removal: `smith remove <project>` deletes the directory and the plugin drops its `use` entry from `go.work`, which would otherwise break every `go` command. Real dry run: `Would remove util (libs/util)`, `changes go.work`, `the pipelines would be regenerated`. With no matching entry it changes nothing.

## Requirements and limits

- `go` on `PATH` (the plugin shells out to `go`; there is no embedded compiler). Generation also needs it: `smith add` fails if `go work init`, `go mod init` or `go work use` fails.
- A workspace with Go modules needs a root `go.work`; the generator creates it, `smith doctor` demands it. A Go project outside the `go.work` is a doctor failure.
- Only `go.mod` marks a project. Not supported: a `go.mod` at the workspace root as a project (skipped by design), module-less Go code.
- The plugin has no `typecheck` and `format` tasks and no formatter; `gofmt` is not run (golangci-lint's `fmt` formatters are not enabled by the default).
- Tags: a Go app is tagged `<name>@<version>` while a package is tagged `<dir>/v<version>`. The plugin does nothing for a module whose import path needs a major-version suffix (`/v2`); not verified.
- Release assets are cross-compiled only when `CGO_ENABLED=0` is possible; a cgo app must declare `// smith:native`.
- `go` kind inference for a module outside `apps/`, `libs/` and `packages/` reads only the files directly in the module directory.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, go 1.27.0 windows/amd64 (the Go version `smith new` wrote into `go.work`), a throw-away workspace made with `smith new go-demo --ci github,azure`. Run: `smith add lib|app|package ... --lang go`, `smith projects`, `smith graph`, `smith explain` (project and tasks), `smith run build|test|lint|package|start|typecheck|format`, `smith doctor` (clean, with a missing `go.work` entry, and with `[toolchains] go = "1.23"`), `smith doctor --fix` (diff only), `smith ci generate`, `smith ci setup`, `smith ci audit` (network; govulncheck), `smith ci check`, `smith release --dry-run [--json] [--notes]`, `smith install -w` / `smith uninstall -w`, `smith up --check`, `smith remove util --dry-run`, `smith plugin check go` (all 11 checks ok), `smith stacks --json`.

Not verified by a run:

- The release asset build itself (`BuildAssets` cross-compiling, zipping, checksums): a real `smith release` would tag and push; only the dry run, which lists names, was run. The behaviour above is read from `releasing/build_assets_use_case.go`.
- A `// smith:native` module and `smith ci native` (code and unit tests only).
- The `lint` task as `go vet` was verified; the move to golangci-lint with smith's default config was verified only by `smith run lint` on this repository (golangci-lint 2.x installed) and the unit tests.
- GitLab CI output and the devcontainer feature on a non-Windows machine. The dev container file written for the demo workspace contained `ghcr.io/devcontainers/features/go:1` with `version` `1.27.0` and `golangciLintVersion` `2.14.0`.

## See also

- [The plugin catalogue](README.md)
- [go-smith-plugin](go-smith-plugin.md): the add-on that scaffolds a smith plugin in Go
- [vertical-slices](vertical-slices.md): the `slice-check` task Go modules get
- [docker](docker.md): the `image` task for a Go app with a `Dockerfile`
- [Releasing](../../../reference/releasing.md), [CI](../../../reference/ci.md), [Configuration](../../../reference/configuration.md#toolchains)
- [Writing a language plugin](../../creating/language-plugin.md)
