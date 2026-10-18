# Working on smith

## Layout

This repository was bootstrapped with the old, Nx-based mnci (#19) and
converted with `smith convert` (#129): it is a smith workspace, built, tested and
released by the smith it contains.

```
apps/smith/                     the CLI: the language-agnostic core, one package per slice
packages/plugin-sdk/            the plugin protocol, framing, runtime, host client and helpers (published)
libs/plugin-node/               the Node language plugin
libs/plugin-node-toolchains/    the typescript, eslint, jest, vitest, playwright, vite, angular, astro and vscode plugins and their stacks
libs/plugin-go/                 the Go language plugin, and the go-smith-plugin scaffold
libs/plugin-python/             the Python language plugin
libs/plugin-rust/, plugin-dotnet/, plugin-ruby/, plugin-php/, plugin-cpp/,
  plugin-dart/, plugin-jvm/, plugin-bicep/    the other language plugins
libs/plugin-docker/             the docker deploy plugin and convention
libs/plugin-vertical-slices/    the vertical-slices structure plugin
install.sh, install.ps1         installers used by people and by generated pipelines
docs/                           documentation, decisions and research
```

Every plugin under `libs/` is built into the `smith` binary and served as a
separate process (`smith plugin serve <name>`).

Go modules are tied together by `go.work`. Code follows vertical slices: a
package per outcome, files named `<name>_<role>.go` (`use_case`, `policy`,
`model`, `contract`, `mapper`, `algorithm`, `client`, `repository`, `store`,
`handler`, `config`, …), and each slice owns its CLI handler.

Core slices in `apps/smith`:

| Slice | Outcome |
|---|---|
| `workspace` | find the root, load `smith.toml` |
| `pluginhost` | find and start plugins (the search path, `plugin.json`), route calls, the stack catalog, `smith plugin list/check` |
| `projectgraph` | projects, edges from imports, overrides with provenance, selection |
| `taskgraph` | expand requested tasks with their dependencies |
| `taskcache` | cache keys, the `Storage` interface and local backend, save/restore/prune |
| `taskrun` | the scheduler, process trees, `smith run/build/test/lint/typecheck` |
| `affected` | changed projects and their dependents |
| `release` | plan and execute releases, `smith release` |
| `splitrepo` | mirror a released project to a repository of its own (`split_repo`) |
| `releaselineage` | decide whether release tags can be trusted: old names, published but untagged, behind the registry |
| `releasepreview` | post the release preview on the pull request (GitHub, Azure Repos) |
| `cipipeline` | pipeline model, renderers, `smith ci` |
| `nativebuild` | `smith ci native`: build native (cgo) projects on each OS and attach their release assets |
| `scaffold` | `smith new`, `smith add` |
| `upgrade` | `smith upgrade`: migrate smith.toml, regenerate what smith owns, bring recipe projects up to date |
| `recipe`, `recipechoice`, `recipestack` | load, resolve and validate a recipe (`smith recipe check`); ask its choices; turn it into a stack and the declarations `smith add` writes |
| `recipedrift` | compare each recipe project with its recipe as it is now (`smith doctor`), plan the upgrade, and list where a recipe's values came from (`smith explain`) |
| `nxworkspace`, `convertplan`, `convert` | read an Nx workspace, plan its conversion, apply it (`smith convert`) |
| `githubrelease` | GitHub Releases and their assets |
| `packagemanifests` | the Homebrew, Scoop and winget manifests attached to an app's GitHub Release |
| `npmbinaries`, `pypiwheels`, `nugettool`, `rubygems` | an app's binaries as npm packages, PyPI wheels, a .NET tool and platform gems |
| `npmpack`, `npmaudit` | the node plugin's `smith npm-pack` and `smith npm-audit` steps |
| `remotecache` | the shared remote cache (HTTP, S3, Azure Blob) behind the local one |
| `adoption` | `smith adopt` |
| `dependencies` | `smith install`, `smith uninstall`, `smith sync`, `smith up` |
| `devservers` | `smith dev` |
| `configedit` | `smith config` |
| `pipelinetemplate` | `smith pipeline`: ready-made pipelines |
| `slicecheck` | `smith slices check`, the vertical-slices convention's check |
| `doctor`, `explain`, `whymiss`, `stackview`, `graphview` | inspection commands |
| `clicontract` | `smith commands`, `smith kinds`, `smith info`: the --json contracts editors read |
| `mcpserver` | `smith mcp` |
| `tui` | the interactive shell a bare `smith` opens |
| `session` | open a workspace for a command |
| `gitclient` | every git call |
| `commandline` | assemble the command tree |
| `goldenpath` | the end-to-end tests (build tag `e2e`): the golden path, a language added by an external plugin (`testdata/kiwiplugin`), and converting Nx and old-mnci workspaces |

## Checks

```sh
# every module go.work lists
for m in $(go work edit -json | sed -n 's/.*"DiskPath": "\.\/\(.*\)".*/\1/p'); do
  (cd $m && go vet ./... && go test ./... && golangci-lint run ./...)
done

# the golden path, end to end (needs every language's toolchain and network)
cd apps/smith && go test -tags e2e -timeout 40m -v -run '^TestGoldenPath$' ./goldenpath
cd apps/smith && go test -tags e2e -timeout 40m -v -skip '^(TestConvert.*|TestGoldenPath)$' ./goldenpath   # each language's workspace
cd apps/smith && go test -tags e2e -timeout 85m -v -run TestConvert ./goldenpath                          # converting Nx and old-mnci workspaces
```

CI runs the same on Linux, macOS and Windows
(`.github/workflows/golden-path.yml`): `unit` on every pull request and push,
and the two golden-path suites once a day (03:00 UTC) and when started by hand
(Actions → golden-path → Run workflow). Run them by hand before merging a
change the unit tests can't judge, such as a plugin's scaffold or the release.
The convert suite builds real Nx and old-mnci workspaces and takes over half
an hour on Windows, so it runs only when started by hand.

`smith run lint test build` runs the same through the workspace (lint includes
the vertical-slices convention's `slice-check`), with the cache. Use the smith you are working on:
`go run ./apps/smith run lint test build`.

## Building

```sh
cd apps/smith && go build -o ../../dist/dev/smith .
```

`smith release` builds the release binaries: the Go plugin stamps `main.version`
and builds six static binaries (`CGO_ENABLED=0`, `-trimpath`), as configured in
`[projects.smith.release]` of `smith.toml`.

## Releasing smith

smith releases itself. The generated pipeline builds smith from the commit (the
`before-setup` slot of `.github/workflows/ci.yml`) and runs `smith ci release` on
`main`. A conventional commit touching `apps/smith` (or a lib it bundles) tags
`smith@<version>` and attaches `monosmith_<version>_<os>_<arch>.zip` for six
platforms, `checksums.txt` and the Homebrew, Scoop and winget manifests to its
GitHub Release, which is what `install.sh` downloads. The same archives are
published to npm (`monosmith`, `@monosmith/cli`), PyPI, NuGet and RubyGems
(all `monosmith`); the secrets are organisation secrets of `monosmith-app`.
pub.dev is published by hand for now. `packages/plugin-sdk` releases as the Go
module tag `packages/plugin-sdk/v<version>`.

Merging to `main` releases: check what `smith release --dry-run` (and the
release preview on the pull request) says before merging.
