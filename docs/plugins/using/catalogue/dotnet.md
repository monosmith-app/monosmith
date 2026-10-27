# The dotnet plugin

The dotnet plugin makes every `.csproj` and `.fsproj` below the workspace root a smith project, pairs each project with its xUnit tests project, and runs the .NET SDK for it: `dotnet build`, `dotnet test`, `dotnet format --verify-no-changes` (as lint), and `dotnet pack` for packages. It sends every `bin/` and `obj/` to `dist/dotnet`, installs the SDK in CI, restores each project, and publishes packages to nuget.org or a private feed. It is a built-in plugin: no installation needed.

## At a glance

| | |
|---|---|
| Plugin name | `dotnet` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold` (all three by default for a language plugin) |
| `--lang` values | `dotnet`, `csharp`, `fsharp` (see the note in [Creating a project](#creating-a-project)) |
| Manifests that make a directory a project | `*.csproj`, `*.fsproj` |
| Requires / conflicts / provides | none / none / no task slots |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `published`, `dependencies`, `declarations`, `latest`. Not advertised: `remove`. |
| Stacks it publishes | none |
| Project kinds it can create | `app`, `lib`, `package` (`dotnet-app`, `dotnet-lib`, `dotnet-package`) |

Capabilities are from `smith plugin list --json`.

## What it detects

(`libs/plugin-dotnet/discovery/`)

| Field | How it is read |
|---|---|
| Project | each project file in a subdirectory of the workspace. A project file at the workspace root is not a project. |
| Tests project | a project file is a tests project if any `PropertyGroup` has `<IsTestProject>true</IsTestProject>` or an `ItemGroup` has `<PackageReference Include="Microsoft.NET.Test.Sdk">`. A tests project is not a smith project of its own; it belongs to the project it references with `ProjectReference`, and that project's `build` and `test` use it. |
| Name | the last element of the directory (`libs/core` is `core`); the project file name is not used |
| Kind | the `<SmithKind>` property (`app`, `lib`, `package`); else the top directory (`apps/`, `libs/`, `packages/`); else `app` when `<OutputType>` is `Exe` or `WinExe`, otherwise `lib` |
| Stack | the `<SmithStack>` property |
| Exports | the package id: first non-empty `<PackageId>`, else `<AssemblyName>`, else the project file name without extension (`NetDemo.Core`) |
| Imports (graph edges) | the `<ProjectReference Include="...">` entries (backslashes accepted) that resolve to another project in the workspace, by its package id; plus the references of the project's tests project |
| Version | the first non-empty `<Version>` |
| Toolchain | `dotnet <version>` from `dotnet --version` |

Real output after `apps/web` got a `ProjectReference` to `libs/core`:

```
$ smith graph
core  lib · dotnet · libs/core
kit  package · dotnet · packages/kit
web  app · dotnet · apps/web
  └─ core
```

`PackageReference` entries are not graph edges.

## Tasks

All tasks run in the project directory. MSBuild builds a project's references itself, so two builds sharing a project must not run at once: `build` waits for the builds of the projects it imports, and `test`, `lint` and `start` wait for this project's `build` and do not build again. Inputs of the cached tasks: `**/*`, `!bin/**`, `!obj/**`, `{workspaceRoot}/Directory.Build.props`, `{workspaceRoot}/Directory.Packages.props`, `{workspaceRoot}/global.json`.

| Task | Command | Appears for | dependsOn | Cached |
|---|---|---|---|---|
| `build` | `dotnet build <tests project, else project file> --nologo` | every project | `^build` | no |
| `lint` | `dotnet format <project file> --verify-no-changes --no-restore` | every project | `build` | yes |
| `test` | `dotnet test <tests project> --nologo --no-build` | a project that has a tests project | `build` | yes |
| `start` | `dotnet run --project <project file> --no-build` | `app` | `build` | no |
| `pack` | `dotnet pack <project file> --nologo -c Release -o {workspaceRoot}/dist/<project dir>` | `package` | `^build` | yes; outputs `dist/<project dir>` |

`dotnet build` builds the tests project when there is one, and that builds the project. `dotnet format` also runs the analyzers; the generated `Directory.Build.props` sets `TreatWarningsAsErrors`. There is no `typecheck` and no `format` task, and the packaging task is called `pack`, not `package`, so it is a different task from the `package` task other plugins give apps for `smith ci pack`; apps have no packaging task here (not verified in a CI run).

With `[conventions.vertical-slices]` each project also gets `slice-check` (inputs `**/*.cs`, `**/*.fs`) and `lint` runs after it; see [vertical-slices](vertical-slices.md).

Real output:

```
$ smith explain web
web has 5 tasks:
  build        dotnet build tests/NetDemo.Web.Tests.csproj --nologo  (after ^build)
  lint         dotnet format NetDemo.Web.csproj --verify-no-changes --no-restore  (cached, after build, slice-check)
  slice-check  smith slices check --lang dotnet .  (cached)
  start        dotnet run --project NetDemo.Web.csproj --no-build  (after build)
  test         dotnet test tests/NetDemo.Web.Tests.csproj --nologo --no-build  (cached, after build)

$ smith explain kit:pack
kit:pack — dotnet pack into dist/packages/kit/
  command    dotnet pack NetDemo.Kit.csproj --nologo -c Release -o {workspaceRoot}/dist/packages/kit
  cwd        packages/kit
  dependsOn  ^build (build of the projects it imports, first)
  inputs     **/*, !bin/**, !obj/**, {workspaceRoot}/Directory.Build.props, {workspaceRoot}/Directory.Packages.props, {workspaceRoot}/global.json
  outputs    ../../dist/packages/kit
  env        —
  cache      true
```

`smith run build`, `test`, `lint` and `pack` all succeeded; `smith run start -p web` printed `Hello, world!`.

## Creating a project

```
smith add app|lib|package <name> --lang dotnet
```

The first project also writes `Directory.Build.props` at the workspace root. Each project is a C# project with its tests project in `<dir>/tests/`. Names are `<Workspace>.<Name>` in PascalCase (`net-demo` + `core` = `NetDemo.Core`). `smith add` does not run `dotnet restore`; the first `smith run build` restores.

Real output:

```
$ smith add lib core --lang dotnet
Adding lib core (dotnet) in libs/core
Wrote 8 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  Directory.Build.props
  libs/core/Greeting/GreetUseCase.cs
  libs/core/Greeting/Greeting.cs
  libs/core/NetDemo.Core.csproj
  libs/core/tests/Greeting/GreetUseCaseTests.cs
  libs/core/tests/NetDemo.Core.Tests.csproj
Reference it from another project with <ProjectReference Include="<relative path>/NetDemo.Core.csproj" />.
Try: smith test -p core

$ smith add app web --lang dotnet
Wrote 6 file(s):
  apps/web/Greeting/GreetUseCase.cs
  apps/web/Greeting/Greeting.cs
  apps/web/NetDemo.Web.csproj
  apps/web/Program.cs
  apps/web/tests/Greeting/GreetUseCaseTests.cs
  apps/web/tests/NetDemo.Web.Tests.csproj
```

| Generated | What it is |
|---|---|
| `Directory.Build.props` (once) | `UseArtifactsOutput=true`, `ArtifactsPath=$(MSBuildThisFileDirectory)dist/dotnet` (so every `bin/` and `obj/` lives under `dist/dotnet`), `Nullable=enable`, `ImplicitUsings=enable`, `TreatWarningsAsErrors=true` |
| `<dir>/<ID>.csproj` | `net10.0`, `RootNamespace` and `AssemblyName` = `<ID>`, `<Version>0.0.0</Version>`, `<SmithKind>`, `<SmithStack>` when a stack was named. An app has `<OutputType>Exe</OutputType>`; a package has `PackageId`, `Description` and `PackageLicenseExpression` MIT; apps and libs have `IsPackable=false`. `tests/**` is removed from the `Compile` and `None` items, since the tests project sits inside the directory. |
| `<dir>/Greeting/Greeting.cs`, `GreetUseCase.cs` | a worked slice: `record Greeting(string Message)` and `GreetUseCase.Greet(name)` returning `Hello, <name>!` |
| `<dir>/Program.cs` (app) | prints `Hello, world!` |
| `<dir>/tests/<ID>.Tests.csproj` | xUnit tests project referencing the project: `Microsoft.NET.Test.Sdk` 17.14.1, `xunit` 2.9.3, `xunit.runner.visualstudio` 3.1.4 |
| `<dir>/tests/Greeting/GreetUseCaseTests.cs` | one xUnit test of `Greet` |

`fsharp` is accepted as a `--lang` value, but the templates are C#: a project made with `--lang fsharp` is a C# project. Existing `.fsproj` projects are discovered and built, but the plugin does not generate F#. No F# project was run here.

## Settings

| Where | Key | Effect | Read in |
|---|---|---|---|
| `.csproj` | `<SmithKind>`, `<SmithStack>` | kind and stack | `discovery/msbuild_project_mapper.go` |
| `.csproj` | `<PackageId>`, `<AssemblyName>`, `<Version>`, `<OutputType>`, `<IsTestProject>` | export, version, app detection, tests detection | same |
| `Directory.Build.props`, `Directory.Packages.props`, `global.json` | whole file | part of every cached task's inputs | `discovery/project_tasks_config.go` |
| `global.json` | `sdk.version` | the `major.minor` channel CI installs (default `10.0`) | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[registries.<name>] nuget`, `auth`, `token_env`; `[workspace] registry`; `[release] registry`; `[projects.<p>] registry` | private NuGet feed for restore and push | `registries/` |
| `smith.toml` | `[toolchains] dotnet = "10"` | version pin checked by doctor | host |
| env | `NUGET_API_KEY` | key for `dotnet nuget push` | `releasing/release_project_use_case.go` |
| env | `NUGET_SOURCE` | push to another source instead of `https://api.nuget.org/v3/index.json` | same |

The plugin reads no `[projects.<p>.settings.dotnet]` keys. `smith recipe check` accepts `[projects.<p>.settings.dotnet] target_framework = "net10.0"` (older docs showed it), but nothing in `libs/plugin-dotnet` reads project or generate settings; the target framework of new projects is the constant `net10.0`. (Contrast with `NUGET_PUSH_URL`, which belongs to the separate app-as-tool release, see [releasing](../../../reference/releasing.md).)

## Toolchain

The pin label is `dotnet`, probed with `dotnet --version`. Real output with `[toolchains] dotnet = "9"` on a machine whose `dotnet --version` is 10.0.401:

```
✖ dotnet matches the pinned version 9
    [toolchains] pins dotnet 9, but this machine runs 10.0.401: tasks cached on another version are misses, and results can differ from CI
    → install dotnet 9 (`mise use dotnet@9`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] dotnet
```

The SDK must be installed; smith does not install it. Generated projects target `net10.0`, so they need the .NET 10 SDK (the test machine also has 3.1 and 7.0 SDKs, which the generated projects do not use). The doctor checks only that `dotnet` runs; it also compares the SDK with CI's channel (older fails, newer warns).

Output: all `bin/` and `obj/` go to `dist/dotnet` (set by the `Directory.Build.props` the plugin writes: `dist/dotnet/bin/<Project>/debug/...`); `pack` writes the `.nupkg` to `dist/<project dir>` (real: `dist/packages/kit/NetDemo.Kit.0.0.0.nupkg`).

## Releasing

| Aspect | Behaviour |
|---|---|
| Version | `<Version>` in the project file. The release writes it into the first `<Version>` element (or adds one to the first `<PropertyGroup>`), publishes, then restores the file. |
| Tag | `<name>@<version>`, name being the directory name: `kit@0.1.0` (real). |
| What is published | `package` projects only: `dotnet pack <file> --nologo -c Release -o dist/<dir> -p:Version=<version>`, then `dotnet nuget push dist/<dir>/<PackageId>.<version>.nupkg --source <source> --skip-duplicate [--api-key <key>]`. `--skip-duplicate` makes a re-run safe. Libs are released by their tag. |
| Release assets of an app | With `[projects.<app>.release]` the plugin runs, for each of its `platforms` (Go's `os/arch`; default all six: `windows/amd64` `win-x64`, `windows/arm64` `win-arm64`, `linux/amd64` `linux-x64`, `linux/arm64` `linux-arm64`, `darwin/amd64` `osx-x64`, `darwin/arm64` `osx-arm64`), `dotnet publish <file> --nologo -c Release -r <rid> --self-contained true -p:PublishSingleFile=true -p:IncludeNativeLibrariesForSelfExtract=true -p:DebugType=None -p:Version=<version> -o dist/release/<app>/<rid>`, so the program carries the .NET runtime and runs where .NET is not installed. Each platform becomes `dist/release/<app>/<asset_name>` (a zip of what publish wrote, without `.pdb` files, plus the README and LICENSE files of the project's directory, else the workspace root, unless the publish has its own) and `checksums.txt` lists them (`sha256sum -c` format). .NET is managed code, so one machine publishes all six platforms: no native job and no linker are needed. The first publish per platform downloads that platform's runtime from NuGet (cached afterwards), so it needs network or a restored NuGet cache. The result feeds the package manager manifests, deb and rpm, npm, PyPI, NuGet tool and RubyGems packages like a Go app's. Each app of a workspace is its own asset set (`product` names it). The checks run in a dry run (`would attach dist/release/cli/cli_0.1.0_linux_arm64.zip, ...`); nothing is built. The project must be an app (`OutputType` `Exe`, or a Web or Worker SDK project), else `dotnet publish wrote no program named <name>: is the project an app (OutputType Exe)?`. A platform .NET has no runtime for: `"linux/mips" has no .NET runtime: release a .NET app for ...`. |
| Registry | nuget.org (`https://api.nuget.org/v3/index.json`) unless `NUGET_SOURCE` is set, or `[registries.*] nuget` names a feed for the project. `published` reads `https://api.nuget.org/v3-flatcontainer/<id>/index.json` for the newest stable version; for a private feed it answers "smith does not ask a private NuGet feed for its versions yet", which doctor shows as a warning. |
| Credentials | `NUGET_API_KEY` (required when a package goes to nuget.org). For an Azure Artifacts feed the key is the placeholder `az` and the credential provider signs the push; for GitHub Packages or a `token` feed, the token variable (`GITHUB_TOKEN`, or `token_env`). |
| Dry run | the plugin returns "dry run" and neither packs nor pushes. Real output: |

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  kit                  unreleased → 0.1.0  (1 releasing commit(s))

$ smith release --dry-run --json      # excerpt
  "tag": "kit@0.1.0",  "publish": "dry run",  "registry": "https://api.nuget.org/v3/index.json"
```

## CI

Real output of `smith ci generate` (GitHub):

```yaml
      - uses: actions/setup-dotnet@v4
        with:
          dotnet-version: '10.0.x'
      - name: Cache the nuget download store
        uses: actions/cache@v4
        with:
          path: ~/.nuget/packages
          key: smith-nuget-${{ runner.os }}-${{ hashFiles('**/*.csproj', '**/*.fsproj') }}
      ...
        env:
          NUGET_API_KEY: ${{ secrets.NUGET_API_KEY }}
```

| Contribution | What |
|---|---|
| Toolchain | the SDK channel of `global.json` (`major.minor`), else `10.0`: `actions/setup-dotnet@v4` with `<channel>.x`; `UseDotNet@2` on Azure; Microsoft's `dotnet-install.sh --channel` elsewhere; the `dotnet` dev container feature |
| Package store | `~/.nuget/packages` (`NUGET_PACKAGES`), keyed on every `.csproj` and `.fsproj` |
| Setup (`smith ci setup`) | `dotnet restore` in each project directory, and `dotnet restore <tests project>` when it has one. Real: `▶ dotnet restore web` and `▶ dotnet restore web tests` per project, `smith ci setup passed in 4s`. |
| Audit | `dotnet list package --vulnerable --include-transitive` once per project (run in the project's folder), reported and not blocking. `dotnet` only prints what it finds, so the step fails when the output says a package is vulnerable. The verdict is remembered per content of the project files (and tests project), `packages.lock.json`, `Directory.Packages.props` and `Directory.Build.props`, so it is skipped while they are unchanged. It needs the SDK and a feed that publishes vulnerability data (nuget.org does) and a restore first, which the setup phase does; a failed run is a warning. |
| Credentials, sign-in | `NUGET_API_KEY` (required; only when a package publishes to nuget.org). Azure Artifacts feed: `NuGetAuthenticate@1`, no secret. GitHub Packages: `GITHUB_TOKEN` and `packages: read` (`write` when publishing there). `token` feed: the variable named in `token_env`. |

Real check with `[registries.libs] auth = "azure-artifacts"` and `[workspace] registry = "libs"`: the Azure pipeline got `- task: NuGetAuthenticate@1` before the cache step, and `smith ci check` asked for the Feed Publisher role on the project's build service.

## Doctor, dependencies and removal

Doctor checks (`pipeline/ci_contribution_use_case.go`, `registries/check_nuget_config_use_case.go`):

| Check | Fails when |
|---|---|
| `the .NET SDK is installed` | `dotnet --version` fails; remedy: install the SDK of the CI channel |
| `.NET SDK is <channel>, as in CI` | `dotnet --version` is older than the channel CI installs (`10.0`, or the `global.json` SDK's major.minor); a newer SDK is a warning. Nothing is reported when `dotnet` is absent |
| `<app> releases for <platform>` | a platform in `[projects.<app>.release]` has no .NET runtime (anything but the six above); remedy: change `platforms` |
| `the .NET runtime for <rid> is downloaded for <app>` | warning only: the platform's runtime is not in the NuGet cache (`microsoft.netcore.app.runtime.<rid>` under `NUGET_PACKAGES` or `~/.nuget/packages`), so the release downloads it and needs network |
| `nuget.config lists the feeds smith.toml names` | only when `smith.toml` names a NuGet feed (or a smith-written `nuget.config` is stale). A `nuget.config` that carries smith's marker comment is rewritten by `smith doctor --fix`; one without the marker is yours, and the fix is shown but marked not safe. |

Real, with an Azure Artifacts feed named in `smith.toml`:

```
✖ nuget.config lists the feeds smith.toml names
    nuget.config does not match [registries] in smith.toml
    → run `smith doctor --fix`
! release tags are not behind the registry
    could not ask https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json about kit (smith does not ask a private NuGet feed for its versions yet), ...
```

`smith doctor --fix --yes` wrote a `nuget.config` with `<clear />` and the one feed (nuget.org is left out for Azure Artifacts, whose upstream sources serve it; for GitHub Packages and token feeds nuget.org is added beside it).

Dependencies: the plugin lists the `PackageReference` of every project and of its tests project, the references in the root `Directory.Build.props`, and the central `PackageVersion` entries of the root `Directory.Packages.props`, so `smith sync`, `smith up` and the one-version check in `smith doctor` cover .NET.

- A reference without a version takes the central one and declares nothing of its own; `VersionOverride` is left alone. Ranges (`[3.0,4.0)`), floating versions (`1.*`) and MSBuild properties (`$(SerilogVersion)`) are reported and never rewritten.
- `smith sync` brings every declaration of a package (a project, its tests project, the central file) to the highest version declared, changing only the version text: the attribute (`Version="1.0"`) or the child element (`<Version>1.0</Version>`), whichever the file uses. Package ids are compared without regard to case.
- `smith up` asks nuget.org for the newest stable release of each package, rewrites to it and runs `dotnet restore` in each project. A package nuget.org does not have (a private feed's) is skipped; one it cannot be asked about is reported.
- `smith install -w kit Newtonsoft.Json[@13.0.1]` runs `dotnet add <project> package`, `smith uninstall -w kit Newtonsoft.Json` runs `dotnet remove <project> package`, and a bare `smith install` runs `dotnet restore`.

Removal: the plugin cleans nothing outside the project directory. Real dry run: `Would remove core (libs/core)`, `the pipelines would be regenerated`.

## Requirements and limits

- The .NET SDK on PATH; generated projects need SDK 10 (they target `net10.0`). `smith add` itself needs no SDK (it only writes files), the first build does, plus network access to restore xunit.
- Because `Directory.Build.props` sets `TreatWarningsAsErrors`, a new warning (for instance a nullable warning) fails the build and `lint`.
- `build` is not cached and has no declared outputs; the test results are cached (`test`, `lint`).
- A project needs its own directory: the name comes from the directory, and a project file at the workspace root is ignored.
- Newer versions come from nuget.org only (private feeds are not asked), the vulnerability audit does not tell fixable findings from the rest, and there is no `format` or `typecheck` task. Release archives are self-contained single-file zips only (no framework-dependent or NativeAOT builds, no `.tar.gz`).
- NuGet package versions are not asked from a private feed.
- F#: discovered and built when present, never generated.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, .NET SDK 10.0.401 (the machine also has SDKs 3.1.426 and 7.0.306; no `global.json` was used, so 10.0.401 ran). Throw-away workspace `net-demo` from `smith new net-demo`. Run: `smith add lib|app|package ... --lang dotnet`, `smith projects`, `smith graph` (with a `ProjectReference`), `smith explain`, `smith run build|test|lint|pack|start`, `smith doctor` (clean; wrong `[toolchains] dotnet` pin; a private feed with a stale `nuget.config`), `smith doctor --fix --yes`, `smith ci generate`, `smith ci setup`, `smith ci audit`, `smith ci check`, `smith release --dry-run [--json]`, `smith install -w` (refusal), `smith remove core --dry-run --force`, `smith plugin list --json`.

Not verified:

- A real `dotnet nuget push` (needs a key and would publish); the pack-then-push sequence is read from `releasing/release_project_use_case.go` and its unit tests.
- Azure Artifacts, GitHub Packages and `token` feeds against a real feed, and `NuGetAuthenticate@1` on an Azure agent.
- F# (`--lang fsharp` output, `.fsproj` discovery), `global.json` pinning an SDK, `Directory.Packages.props` central package management.
- Linux and macOS.

## See also

- [The plugin catalogue](README.md)
- [vertical-slices](vertical-slices.md)
- [Releasing](../../../reference/releasing.md), [CI](../../../reference/ci.md), [Configuration: private registries](../../../reference/configuration.md#private-registries)
- [Writing a language plugin](../../creating/language-plugin.md)
