# The bicep plugin

The bicep plugin makes Azure infrastructure-as-code a smith workspace's
projects. Every directory with a `main.bicep` is a project: a deployment (an
app, usually with a `main.bicepparam` of parameters) or a module that
deployments use (a lib). The plugin lints each with the Azure CLI's Bicep linter,
compiles `main.bicep` to an ARM template and `main.bicepparam` to a parameters
file under `dist/`, follows `module` references between projects as dependencies,
creates new Bicep projects, writes a release version into `main.bicep`'s
metadata, and tells CI to install the Azure CLI with Bicep. It never deploys
anything: the pipeline that owns your environment does that with the files in
`dist/`.

## At a glance

| | |
|---|---|
| Plugin name | `bicep` (built in; runs as a separate process, `smith plugin serve bicep`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` value | `bicep` |
| Manifest that makes a directory a project | `main.bicep` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app` and `lib`. `package` is refused (see below) |
| Tasks | `lint`, `params`, `build` |

Roles and capabilities are what `smith plugin list --json` prints; the language,
manifest and the absence of requires, conflicts, provides and stacks come from
the plugin's `Describe` method (`libs/plugin-bicep/plugin/bicep_plugin_handler.go`).

## What it detects

smith hands the plugin every `main.bicep` in the files git lists. The one at the
workspace root is skipped. For each other one, the plugin reads the file with
regular expressions, not with the Bicep compiler:

- **Project name.** The directory name (`apps/infra` is `infra`).
- **Exports.** The project's directory (`apps/infra`).
- **Imports (graph edges).** Every `module <symbol> '<path>'` statement whose
  path is a local file (no `:` in it, so `br:` and `ts:` registry references are
  ignored). The path is resolved to the module file's directory, and that
  directory is the import when it is outside the project's own directory and
  inside the workspace. A module inside the project (`modules/shared.bicep`)
  makes no edge.
- **Kind.** `metadata smithKind = 'app'` (or `lib`, `package`); else the top
  directory (`apps/`, `libs/`, `packages/`); else **`app`** (the fallback here is
  app, not lib: a stray `main.bicep` is a deployment).
- **Stack.** `metadata smithStack = '<name>'`.
- **Version.** `metadata version = '<x.y.z>'`.
- **Parameters.** A `main.bicepparam` in the same directory adds the `params`
  task.
- **Toolchain.** None. The plugin does not report an Azure CLI or Bicep version,
  so neither is part of any task's cache key, and a `[toolchains]` pin for `az`
  produced no doctor finding.
- **Description, tags.** None.

A real run after `apps/infra/main.bicep` gained a `module` statement for
`libs/network` (and a `br:` module, which made no edge):

```text
$ smith graph
infra  app · bicep · apps/infra
  └─ network
network  lib · bicep · libs/network
```

## Tasks

All tasks run in the project's directory and share these inputs:
`**/*.bicep`, `**/*.bicepparam`, `bicepconfig.json`. All are cached. Because the
host mixes the sources of the projects a project imports into its cache key, a
change in `libs/network` invalidates `infra`'s tasks.

| Task | Command | Outputs | dependsOn | Appears |
|---|---|---|---|---|
| `lint` | `az bicep lint --file main.bicep` | none | none | every project |
| `params` | `az bicep build-params --file main.bicepparam --outfile ../../dist/<dir>/parameters/main.parameters.json` | `../../dist/<dir>/parameters` | none | when `main.bicepparam` exists |
| `build` | `az bicep build --file main.bicep --outdir ../../dist/<dir>/template` | `../../dist/<dir>/template` | `params` when that task exists | every project |

`az bicep lint` exits 0 on warnings, so the scaffold writes a `bicepconfig.json`
that raises three rules to errors (see below) and only those fail `lint`. The
`template` and `parameters` directories are separate because smith empties a
task's outputs before it runs. There is no `test`, `start`, `format` or
`typecheck` task (`smith test -p infra` says `no selected project has a [test]
task`).

```text
$ smith explain infra
infra has 3 tasks:
  build   az bicep build --file main.bicep --outdir ../../dist/apps/infra/template  (cached, after params)
  lint    az bicep lint --file main.bicep  (cached)
  params  az bicep build-params --file main.bicepparam --outfile ../../dist/apps/infra/pa…  (cached)

$ smith explain network
network has 2 tasks:
  build  az bicep build --file main.bicep --outdir ../../dist/libs/network/template  (cached)
  lint   az bicep lint --file main.bicep  (cached)
```

Real run and its output files:

```text
$ smith run lint build
...
✔ 5/5 tasks succeeded (0 from cache) in 8.996s

dist/apps/infra/parameters/main.parameters.json
dist/apps/infra/template/main.json
dist/libs/network/template/main.json
```

(Note: the vertical-slices convention does not apply to Bicep, so there is no
`slice-check`.)

## Creating a project

```sh
smith add app infra --lang bicep
smith add lib network --lang bicep
```

Real output:

```text
$ smith add app infra --lang bicep
Adding app infra (bicep) in apps/infra
Wrote 5 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  apps/infra/bicepconfig.json
  apps/infra/main.bicep
  apps/infra/main.bicepparam
After smith build, deploy it with az deployment group create --template-file dist/apps/infra/template/main.json --parameters dist/apps/infra/parameters/main.parameters.json.
Try: smith test -p infra

$ smith add lib network --lang bicep
Adding lib network (bicep) in libs/network
Wrote 2 file(s):
  libs/network/bicepconfig.json
  libs/network/main.bicep
Use it from a deployment: module network '<relative path>/libs/network/main.bicep' = { ... }.
Try: smith test -p network
```

What is generated:

- `main.bicep`: `metadata version = '0.0.0'`, `metadata smithKind = '<kind>'`
  (and `smithStack` with a stack), a `location` parameter defaulting to
  `resourceGroup().location`, a required `storageName` parameter
  (3 to 24 characters), one `Microsoft.Storage/storageAccounts@2023-05-01`
  resource (`Standard_LRS`, `StorageV2`) and an `output storageId`.
- `bicepconfig.json`: the linter rules `no-unused-params`, `no-unused-vars` and
  `no-hardcoded-location` at level `error`.
- For an app only, `main.bicepparam`: `using './main.bicep'` and
  `param storageName = 'st<workspace><name>'` (lower-case letters and digits,
  at most 22 after `st`, so `stbicepdemoinfra`).
- A lib is a module: it has no parameters file, because a deployment passes its
  parameters.

Limits:

- `smith add package x --lang bicep` fails with `smith: a Bicep project is an app
  (a deployment) or a lib (a module): publishing to a Bicep registry is not
  supported yet`, although `smith kinds` lists `bicep-package`. You can still
  make a Bicep project a package by hand with `metadata smithKind = 'package'`
  (see Releasing).
- The hint `Try: smith test -p infra` is wrong for Bicep: there is no test task.

## Settings

The plugin has no `[projects.<p>.settings.bicep]` keys and reads no environment
variable.

| Where | Key | Effect | Read in |
|---|---|---|---|
| `main.bicep` | `metadata smithKind = '...'` | `app`, `lib` or `package` | `discovery/bicep_file_mapper.go` |
| `main.bicep` | `metadata smithStack = '...'` | the project's stack | same |
| `main.bicep` | `metadata version = '...'` | the version a release starts from and rewrites | `releasing/bicep_version_mapper.go` |
| `main.bicep` | `module <name> '<local path>'` | a graph edge | `discovery/bicep_file_mapper.go` |
| project directory | `main.bicepparam` present | adds `params`, and `build` waits for it | `discovery/discover_projects_use_case.go` |
| project directory | `bicepconfig.json` | the linter's rules (and an input of every task) | Bicep itself |
| `smith.toml` | `[tasks.<task>]`, `[projects.<p>.tasks.<task>]` | override an inferred task | host ([tasks and cache](../../../reference/tasks-and-cache.md)) |
| `smith.toml` | `[projects.<p>] tag_prefix` | the release tag | host |

## Toolchain

- **Probe.** None. The plugin does not run a version command, so no toolchain
  version is recorded for Bicep projects and `[toolchains]` cannot pin it
  (`[toolchains] az = "2.1"` raised no finding).
- **Needed on the machine.** The Azure CLI (`az`) with its Bicep component.
  `az bicep install` installs the latter; on this machine Bicep was already
  present through `az`, so the standalone `bicep` CLI is not needed.
- **Where output goes.** `dist/<project dir>/template/main.json` and
  `dist/<project dir>/parameters/main.parameters.json`.
- **Registry modules.** A `module ... 'br:...'` reference makes `az bicep lint`
  and `build` contact that registry and sign in. In a test, a
  `br:contoso.azurecr.io/...` module made `lint` fail with
  `{"errors":[{"code":"UNAUTHORIZED","message":"token validation failed: ..."}]}
  ... [https://aka.ms/bicep/core-diagnostics#BCP192]`. This is the Bicep CLI's
  behaviour, not smith's; I removed the module after seeing it.

## Releasing

| | |
|---|---|
| Version in the manifest | `metadata version = '<x.y.z>'` in `main.bicep` |
| Version write | the statement is replaced in place, or added as the first line of the file when absent. A dry run lists `libs/network/main.bicep` and writes nothing |
| Tag | `<project>@<version>` (override with `tag_prefix`) |
| Publish | nothing is uploaded, for any kind: "released by its tag". Deployment is up to the pipeline that owns the environment |
| Registry query | none (`published` is not advertised) |
| Credentials | none |
| Private registry | none |
| Dry run | `smith release --dry-run` plans without tagging |

Only a `package` (or an app with a `[projects.<p>.release]` section) is released.
By default a Bicep project is an `app` or a `lib`, so `smith release` finds
nothing to do (`Nothing to release: no package has a feat: or fix: commit since
its last release ...`). With `metadata smithKind = 'package'` on `libs/network`
and two `feat:` commits, a real dry run said:

```text
$ smith release --dry-run
Dry run: nothing is published or tagged.
  network              unreleased → 0.1.0  (2 releasing commit(s))
```

and the JSON report had `"tag": "network@0.1.0"`, `"files": ["libs/network/main.bicep"]`
and `"publish": "released by its tag"`.

## CI

`smith ci generate` adds, when the workspace has a Bicep project, one toolchain
step on every provider (named `azure-cli`):

```yaml
      - name: Install azure-cli
        run: |
          command -v az >/dev/null || curl -fsSL https://aka.ms/InstallAzureCLIDeb | bash
          az bicep install
```

(GitHub's and Azure's hosted images carry `az`, so the download is skipped there;
`az bicep install` runs everywhere.) The dev container gets the
`ghcr.io/devcontainers/features/azure-cli:1` feature with `installBicep: "true"`.
The plugin adds no setup step, no audit step, no package store, no credential and
no permission; `smith ci check` printed `nothing to set up by hand`. The Azure
Pipelines and GitLab output uses the same install commands (not generated
separately by me for Bicep).

## Doctor, dependencies and removal

**Doctor.** Run only when the workspace has Bicep projects.

| Check | Fails when |
|---|---|
| `the Azure CLI is installed` | `az version` fails |
| `the Bicep CLI is installed` | (only if `az` works) `az bicep version` fails; the remedy is `az bicep install` |

Both passed. Doctor also warned, from the host, once per project:
`! infra can be verified: infra has no test task, so CI never checks it`
(and the same for `network`), because Bicep projects have no `test` task. The
run ended `16 passed, 0 failed, 2 warning(s)`.

**Dependencies.** Not supported: `smith install`, `smith sync` and `smith up`
do nothing for Bicep. Module versions live in `br:` references you edit.

**Remove.** The plugin has no remove support: `smith remove` deletes the
project's directory and regenerates the pipelines. A deployment's `module`
statement pointing at a removed lib is not edited (the host refuses to remove a
project others import unless `--force`).

## Requirements and limits

- Azure CLI with Bicep. Nothing else.
- Discovery is by regular expression: a `module` statement not at the start of
  a line, or a `metadata` statement in double quotes, is not seen. Bicep strings
  use single quotes, which the patterns expect.
- A Bicep module repository is not supported: `br:` and `ts:` references are
  ignored by the graph and there is no `package` scaffold.
- `lint` only fails on rules your `bicepconfig.json` sets to `error`.
- No toolchain pin, no test task, no format task.

## Verified

Windows 11 Pro, smith 1.62.0, Azure CLI 2.90.0, Bicep CLI 0.48.1 (installed
through `az bicep`, no download needed).

Run: `smith new`, `smith add app|lib --lang bicep`, the refusal of
`smith add package`, `smith projects`, `smith graph` (with a real `module`
statement and a `br:` one), `smith explain` for each task, `smith run lint build`
(real compilation to ARM JSON and a parameters file), `smith test -p infra`,
`smith doctor`, `smith ci generate`, `smith ci check`, `smith release --dry-run`
(plain and `--json`, with a hand-made package).

Not verified: a real `az deployment group create` (no Azure subscription was
used); modules from a Bicep registry (`br:`), beyond the failed sign-in above;
the Azure Pipelines and GitLab YAML for Bicep.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Tasks and cache](../../../reference/tasks-and-cache.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
