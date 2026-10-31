# The dart plugin

The dart plugin makes the packages of a pub workspace smith projects, whether
they are plain Dart or Flutter. A root `pubspec.yaml` lists the members under
`workspace:`; every member has its own `pubspec.yaml` and `resolution:
workspace`, and the whole workspace shares one `pubspec.lock`. The plugin finds
the members, gives each one `analyze`, `dart format` and `test`, builds apps
(a Flutter app for the web, a Dart app into a native executable bundle), creates
new packages and Flutter apps, writes release versions, publishes packages to
pub.dev or a private pub repository, and installs and removes pub dependencies.

## At a glance

| | |
|---|---|
| Plugin name | `dart` (built in; runs as a separate process, `smith plugin serve dart`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` values | `dart`, `flutter` (`--lang flutter` reaches this plugin but makes a plain Dart project; see Creating a project) |
| Manifest that makes a directory a candidate project | `pubspec.yaml` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Not advertised | none of the dependency capabilities |
| Stacks it publishes | `flutter`: "a Flutter app for the web, in the pub workspace" |
| Project kinds `smith add` can create | `app`, `lib`, `package` (`smith kinds` lists `dart-app`, `dart-lib`, `dart-package`); the `flutter` stack makes apps only |
| Tasks | `lint`, `format`, `test`, `build`, `start` |

Roles and capabilities are what `smith plugin list --json` prints; languages,
manifests, stacks and the absence of requires, conflicts and provides come from
the plugin's `Describe` method (`libs/plugin-dart/plugin/dart_plugin_handler.go`).
The `flutter` stack appears in `smith stacks` as `flutter  (plugin dart)`.

## What it detects

smith hands the plugin every `pubspec.yaml` in the files git lists. For each:

- **Root manifest.** A `pubspec.yaml` with a `workspace:` list only names its
  members, so it is not a project. The root at `.` is skipped too.
- **Project name.** The directory name (`libs/core` is `core`).
- **Exports.** The pubspec's `name` (`dart_demo_core`).
- **Imports (graph edges).** Every key of `dependencies` and `dev_dependencies`
  except the Flutter SDK itself (`flutter: sdk: flutter`). smith matches those
  names against other projects' exports.
- **Flutter or Dart.** A project is a Flutter project when `flutter` is
  declared with `sdk: flutter` in `dependencies` or `dev_dependencies`. Then its
  tasks run `flutter` instead of `dart`.
- **Kind.** `smith: kind:` in the pubspec (`app`, `lib` or `package`); else the
  top directory (`apps/`, `libs/`, `packages/`); else `lib`.
- **Stack.** `smith: stack:` in the pubspec.
- **Version and description.** `version` and `description` of the pubspec.
- **Publish target.** For a `package`, `publish_to` when it is set, else
  `pub.dev`. Other kinds have none.
- **Tests.** A `test/` directory.
- **Toolchain.** The first version in `dart --version`, as `dart 3.12.0`
  (Flutter carries Dart, so the label is `dart` for Flutter projects too).

A real run, after `packages/sdk` declared `dart_demo_core: ^0.0.0`:

```text
$ smith graph
core  lib · dart · libs/core
mobile  app · dart · apps/mobile
sdk  package · dart · packages/sdk
  └─ core
tool  app · dart · apps/tool
```

## Tasks

All tasks run in the member's directory. For a Flutter project, `lint`, `test`,
`build` and `start` use `flutter` with `--no-pub`: the workspace is resolved once
(`smith install` or CI setup), so parallel tasks never rewrite the shared
lockfile at once. `format` always uses `dart`.

Every task has these inputs (except `start`): `**/*`, `!.dart_tool/**`,
`!build/**`, `{workspaceRoot}/pubspec.yaml`, `{workspaceRoot}/pubspec.lock`.

| Task | Command (Dart / Flutter) | Outputs | dependsOn | Cached | Appears |
|---|---|---|---|---|---|
| `lint` | `dart analyze --fatal-infos` / `flutter analyze --fatal-infos --no-pub` | none | none | yes | every project |
| `format` | `dart format --output=none --set-exit-if-changed .` | none | none | yes | every project |
| `test` | `dart test` / `flutter test --no-pub` | none | none | yes | when `test/` exists |
| `build` | `dart build cli -t bin/main.dart -o {workspaceRoot}/dist/<dir>` / `flutter build web --no-pub --no-wasm-dry-run -o {workspaceRoot}/dist/<dir>` | `../../dist/<dir>` | none | yes | kind `app` only |
| `start` | `dart run bin/main.dart` / `flutter run --no-pub -d web-server` | none | none | no | kind `app` only |

Notes:

- A `lib` or `package` has no `build`: a Dart package ships as source.
- With the `vertical-slices` convention, each project also gets `slice-check`,
  and `lint` depends on it (the convention's, not the plugin's).
- `format` fails on any unformatted file; it never rewrites.
- `build` for a Dart app needs `bin/main.dart`; the Dart output is in
  `dist/<dir>/bundle`.

```text
$ smith explain tool          # a Dart app
tool has 6 tasks:
  build        dart build cli -t bin/main.dart -o {workspaceRoot}/dist/apps/tool  (cached)
  format       dart format --output=none --set-exit-if-changed .  (cached)
  lint         dart analyze --fatal-infos  (cached, after slice-check)
  slice-check  smith slices check --lang dart .  (cached)
  start        dart run bin/main.dart
  test         dart test  (cached)

$ smith explain mobile        # a Flutter app
mobile has 6 tasks:
  build        flutter build web --no-pub --no-wasm-dry-run -o {workspaceRoot}/dist/apps/mobile  (cached)
  format       dart format --output=none --set-exit-if-changed .  (cached)
  lint         flutter analyze --fatal-infos --no-pub  (cached, after slice-check)
  slice-check  smith slices check --lang dart .  (cached)
  start        flutter run --no-pub -d web-server
  test         flutter test --no-pub  (cached)
```

A real `build` output (the Dart app builds `dist/apps/tool/bundle/bin/main.exe`
on Windows; the Flutter app writes `index.html`, `main.dart.js`,
`flutter_bootstrap.js`, `assets/` and the rest into `dist/apps/mobile`):

```text
tool:build │ Generated: ...\dist\apps\tool\bundle\bin\main.exe
mobile:build │ Compiling lib\main.dart for the Web...                             20.7s
mobile:build │ √ Built ..\..\dist\apps\mobile
mobile:build ✔ 25.467s
```

`smith run start -p tool` printed `Hello, world!`. `flutter run -d web-server`
(the Flutter `start`) runs until stopped; not run by me.

## Creating a project

```sh
smith add lib core --lang dart
smith add app tool --lang dart
smith add package sdk --lang dart
smith add app mobile --stack flutter
```

Real output, trimmed of `pub get`'s package list:

```text
$ smith add lib core --lang dart
Adding lib core (dart) in libs/core
Resolving dependencies...
...
Wrote 9 file(s):
  .github/workflows/ci.yml
  .gitignore
  libs/core/analysis_options.yaml
  libs/core/lib/dart_demo_core.dart
  libs/core/lib/src/greeting/greet_use_case.dart
  libs/core/lib/src/greeting/greeting.dart
  libs/core/pubspec.yaml
  libs/core/test/src/greeting/greet_use_case_test.dart
  pubspec.yaml
Depend on it from another package with dart_demo_core: ^0.0.0 under dependencies (the workspace resolves it locally).
Try: smith test -p core

$ smith add app tool --lang dart
Wrote 8 file(s):
  apps/tool/analysis_options.yaml
  apps/tool/bin/main.dart
  apps/tool/lib/dart_demo_tool.dart
  apps/tool/lib/src/greeting/greet_use_case.dart
  apps/tool/lib/src/greeting/greeting.dart
  apps/tool/pubspec.yaml
  apps/tool/test/src/greeting/greet_use_case_test.dart
  pubspec.yaml

$ smith add app mobile --stack flutter
Wrote 18 file(s):
  .github/workflows/ci.yml
  .github/workflows/pub.yml
  apps/mobile/.gitignore
  apps/mobile/.metadata
  apps/mobile/analysis_options.yaml
  apps/mobile/lib/main.dart
  apps/mobile/lib/src/greeting/greet_use_case.dart
  apps/mobile/lib/src/greeting/greeting.dart
  apps/mobile/pubspec.yaml
  apps/mobile/test/src/greeting/greet_use_case_test.dart
  apps/mobile/web/favicon.png
  apps/mobile/web/icons/Icon-192.png
  apps/mobile/web/icons/Icon-512.png
  apps/mobile/web/icons/Icon-maskable-192.png
  apps/mobile/web/icons/Icon-maskable-512.png
  apps/mobile/web/index.html
  apps/mobile/web/manifest.json
  pubspec.yaml
```

What is written:

- **The workspace.** The first project creates the root `pubspec.yaml`:
  `name: <workspace>_workspace`, `publish_to: none`, `sdk: ^3.7.0`, and
  `workspace:` listing the member. Later projects add themselves to the list.
  `.dart_tool/` is added to `.gitignore`.
- **A package** (`libs/core/pubspec.yaml`): name `<workspace>_<project>` (lower
  case, underscores), `version: 0.0.0`, `publish_to: none` unless the kind is
  `package`, `resolution: workspace`, `lints ^6.0.0` and `test ^1.25.0` as dev
  dependencies, a `smith:` block with `kind` (and `stack`), and
  `analysis_options.yaml` including `package:lints/recommended.yaml`.
- **The worked example**, a greeting slice: `lib/src/greeting/greet_use_case.dart`
  (`String greet(String name) => 'Hello, $name!';`), a barrel
  `greeting.dart` that exports `greet`, `lib/<package>.dart` as the public
  API, and a mirrored test under `test/src/greeting/`.
- **A Dart app** adds `bin/main.dart`, which prints `greet('world')`.
- **A Flutter app** (stack `flutter`, kind `app`): `flutter create --no-pub
  --empty --platforms web` writes the web files; the plugin removes the
  `pubspec.yaml`, `analysis_options.yaml`, `README.md` and `lib/main.dart` that
  `flutter create` made and writes its own: a `MaterialApp` showing the greeting,
  `flutter_lints`, `flutter_test`, `uses-material-design: true`.
- The workspace is then resolved with `dart pub get`, or `flutter pub get` when
  any member is a Flutter package.

Limits seen:

- The `flutter` stack makes apps only: `smith add lib g --stack flutter` fails
  with `smith: the flutter stack makes apps; a Dart package or library is --lang dart`.
- `smith add lib f --lang flutter` printed `Adding lib f (dart)`: it made a
  plain Dart library, not a Flutter one. To get a Flutter library, add
  `flutter: sdk: flutter` to its pubspec yourself and its tasks switch to
  `flutter` (not run by me).
- Unlike most plugins, `pub get`'s output is always shown, not only on failure.

## Settings

The plugin has no `[projects.<p>.settings.dart]` keys and reads no environment
variable of its own beyond `PUB_TOKEN` (Releasing).

| Where | Key | Effect | Read in |
|---|---|---|---|
| `pubspec.yaml` | `smith: kind:` | `app`, `lib` or `package` | `discovery/pubspec_mapper.go` |
| `pubspec.yaml` | `smith: stack:` | the project's stack (`flutter` for generated Flutter apps) | same |
| `pubspec.yaml` | `dependencies: flutter: sdk: flutter` | makes the project a Flutter project: tasks use `flutter` | same |
| `pubspec.yaml` | `publish_to` | a pub repository URL for a package; default pub.dev | `discovery/publish_to_policy.go` |
| `pubspec.yaml` | `version` | the version the release starts from and rewrites | `releasing/pubspec_version_mapper.go` |
| `.flutter-version` (workspace root) | a Flutter release, for example `3.41.0` | the Flutter version CI installs; default the `stable` channel | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[toolchains] dart = "3.12"` | doctor pin | host |
| `smith.toml` | `[projects.<p>] registry = "none"`, `tag_prefix`, `[projects.<app>.release] pub`, `pub_tag` | release settings | host ([releasing](../../../reference/releasing.md)) |
| `smith.toml` | `[tasks.<task>]`, `[projects.<p>.tasks.<task>]` | override an inferred task | host |

The scaffold's constants (`sdk: ^3.7.0`, `lints ^6.0.0`, `test ^1.25.0`) are in
`scaffolding/package_templates_config.go` and not configurable. A `.dart-version`
file is not read: CI installs the Dart SDK from the `stable` channel.

## Toolchain

- **Probe.** `dart --version` (stderr or stdout), as `dart <version>`.
  `flutter` is probed only by doctor. The installed Dart version is part of every
  task's cache key.
- **Pin.** `[toolchains] dart = "3.0"` is checked by `smith doctor`. Real output
  with Dart 3.12.0:

```text
✖ dart matches the pinned version 3.0
    [toolchains] pins dart 3.0, but this machine runs 3.12.0: tasks cached on another version are misses, and results can differ from CI
    → install dart 3.0 (`mise use dart@3.0`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] dart
```

- **CI ignores `[toolchains]`.** Only `.flutter-version` changes the Flutter
  version in CI (with `3.41.0` in it the workflow said
  `flutter-version: '3.41.0'`); the Dart-only workflow always installs `stable`.
- **Needed on the machine.** `dart` (or `flutter`, which carries it), `flutter` for
  Flutter projects. Both SDKs must live outside the workspace.
- **Where output goes.** `dist/<app dir>` for an app's `build`. pub's own
  `.dart_tool/` and `pubspec.lock` are at the member and the workspace root.

## Releasing

| | |
|---|---|
| Version in the manifest | `version:` in the member's `pubspec.yaml` |
| Version write | the line is replaced in place; with none, `version: x` is added after `name:`. The file is restored after the release. A dry run lists `packages/sdk/pubspec.yaml` and writes nothing |
| Tag | `<project>@<version>` (override with `tag_prefix`); an app published to pub.dev also gets a pub tag (see below) |
| Publish: not a package | "released by its tag" |
| Publish: assets of an app | "Dart release assets are not built yet: the app is released by its tag" |
| Publish: dry run | nothing is run; the result names the registry (`https://pub.dev`, or the `publish_to` URL) and says "dry run" |
| Publish: pub.dev | not done by the release job. The result is "published to pub.dev by the pub workflow, on the package's pub tag": the release starts `.github/workflows/pub.yml` (generated by `smith ci generate`), which runs `smith release pub` and `dart pub publish` authenticated by GitHub's OIDC token. See [releasing: pub.dev](../../../reference/releasing.md#pubdev) |
| Publish: another repository (`publish_to: https://...`) | with `PUB_TOKEN` set, `dart pub token add <url> --env-var PUB_TOKEN`; then `dart pub publish --force` (`flutter pub publish --force` for a Flutter package) |
| Registry query | `GET <registry>/api/packages/<name>` for the newest version (`latest.version`); 404 means not published; other failures are "unreachable" warnings |
| Credentials | `PUB_TOKEN` only for a package whose `publish_to` is not pub.dev; none for pub.dev |
| Dry run | `smith release --dry-run` |

A real dry run for a package with the default target:

```text
$ smith release --dry-run
Dry run: nothing is published or tagged.
  sdk                  unreleased → 0.1.0  (1 releasing commit(s))
```

`"publish": "dry run"`, `"registry": "https://pub.dev"` and the file
`packages/sdk/pubspec.yaml` in the JSON report. With
`publish_to: https://pub.example.com` the report said
`"registry": "https://pub.example.com"` and the generated CI workflow gained
`PUB_TOKEN: ${{ secrets.PUB_TOKEN }}` on the release step; `smith ci check`
then listed `PUB_TOKEN` among the secrets to set up by hand.

A package whose pubspec says `publish_to: none` is never published: the plan
says `published by its tag`, the plugin asks no registry about it, and the CI
workflow does not ask for `PUB_TOKEN` (the same as `registry = "none"` under
`[projects.<name>]` in `smith.toml`).

pub.dev itself needs setup that smith cannot do: publish the first version by
hand, then enable automated publishing for the repository with the tag pattern;
see [releasing: pub.dev](../../../reference/releasing.md#pubdev).

The release's SBOM lists the packages of `pubspec.lock`, falling back to `pubspec.yaml`, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

| What | Where it comes from |
|---|---|
| Dart toolchain | GitHub: `dart-lang/setup-dart@v1` with `sdk: stable`. Elsewhere: downloads `dartsdk-linux-x64-release.zip` from `storage.googleapis.com/dart-archive/channels/stable/...` and prepends its `bin` to `PATH` |
| Flutter toolchain | replaces Dart's when any project runs `flutter`. GitHub: `subosito/flutter-action@v2` with `channel: 'stable'`, `cache: true`, and `flutter-version` from `.flutter-version`. Elsewhere: `git clone --depth 1 --branch <channel or version> https://github.com/flutter/flutter.git "$HOME/flutter"` |
| Store | `~/.pub-cache` (`PUB_CACHE`), keyed on `**/pubspec.lock` (GitHub `actions/cache@v4`, Azure `Cache@2`, GitLab `cache`) |
| Setup step | `Resolve the pub workspace`: `dart pub get` (`flutter pub get` when Flutter), blocking |
| Credentials | `PUB_TOKEN` for a package that publishes to a repository other than pub.dev |
| Permissions | with a pub.dev package or app the release job gets `actions: write` (it starts the pub workflow); real output: `contents: write`, `pull-requests: write`, `actions: write` |
| Extra workflow | `.github/workflows/pub.yml` (GitHub only): `workflow_dispatch` with `project` and `version`, `id-token: write`, then `smith release pub --project ... --version ...` |
| Dev container | none from this plugin |

The GitHub fragment of a workspace with a Flutter app, from a real run:

```yaml
      - uses: subosito/flutter-action@v2
        with:
          channel: 'stable'
          cache: true
      - name: Cache the pub download store
        uses: actions/cache@v4
        with:
          path: ~/.pub-cache
          key: smith-pub-${{ runner.os }}-${{ hashFiles('**/pubspec.lock') }}
```

Azure output (Flutter): a bash step cloning Flutter with
`--branch 3.41.0`, `##vso[task.prependpath]$HOME/flutter/bin`, a `Cache@2` step
and `PUB_CACHE` as a pipeline variable. GitLab was not generated for Dart by me.

## Doctor, dependencies and removal

**Doctor.** Run only when the workspace has Dart projects.

| Check | Fails when |
|---|---|
| `the Dart SDK is installed` | `dart --version` fails |
| `Flutter is installed` | (only if a project runs `flutter`) `flutter --version` fails |
| `the Dart and Flutter SDKs live outside the workspace` | `flutter/`, `dart-sdk/` or `.fvm/` exists at the workspace root |

All passed.

**Install and uninstall.**

| Command | Runs |
|---|---|
| `smith install` | `dart pub get` (`flutter pub get` when a member is Flutter) at the root. Printed `✔ dart: installed 4 project(s)` |
| `smith install -w core path` | `dart pub add path` in the member (`flutter pub add` for a Flutter project) |
| `smith install -w core -D mockito` | `dart pub add dev:mockito` |
| `smith uninstall -w core path mockito` | `dart pub remove path mockito` |

All four were run; each reported the member's `pubspec.yaml` and `pubspec.lock`
as changed. `smith install -w sdk dart_demo_core` added the dependency on a
workspace sibling and made it a graph edge (above).

**Versions: `smith sync` and `smith up`.** The plugin lists the `dependencies` and
`dev_dependencies` of every member's `pubspec.yaml` and of the workspace root's,
so `smith sync`, `smith up`, the Deps tab and the one-version check in `smith
doctor` cover Dart and Flutter.

- A dependency from a `path`, a `git` repository or the SDK (`sdk: flutter`)
  declares no version of its own and is left alone. A hosted entry written as a
  map (`hosted:` and `version:`) counts under its `version:`.
- `smith sync` brings every declaration of a package to the highest constraint
  declared, changing only the constraint (quotes, comments and other keys stay).
  `any` and ranges of two comparisons (`>=1.0.0 <2.0.0`) are reported and left
  alone.
- `smith up` asks pub.dev for the newest stable release of each package,
  rewrites to it keeping the operator (`^`), and runs `pub get` (`flutter pub get`
  for a Flutter workspace) so `pubspec.lock` follows. A package pub.dev does not
  have (a private server's) is skipped.
- Not offered: a vulnerability audit. pub has no audit command, so `smith ci audit`
  has nothing to run for Dart.

**Remove.** `smith remove tool --yes` deleted `apps/tool` and dropped it from the
root `pubspec.yaml`'s `workspace:` list (`changes pubspec.yaml`). The lock file
is not touched; `pub get` rewrites it on the next resolve.

## Requirements and limits

- The Dart SDK 3.7 or newer (`sdk: ^3.7.0` in new pubspecs). `dart build cli`
  is a recent command; I ran it on Dart 3.12.0 and cannot say from which version
  it exists.
- A Flutter app is built for the web only. There is no Android, iOS or desktop
  build task, and release assets for apps are not built.
- `format` fails rather than formats; run `dart format .` yourself.
- No `typecheck` or `package` task, and no `test` task for a member without
  `test/`.
- Pub workspaces need a single resolution: all members share one `pubspec.lock`.
  Adding the first Flutter app made pub move some shared packages to older
  versions (`< matcher 0.12.19 (was 0.12.20)` in the output).
- CI ignores `[toolchains]` and `.dart-version`.
- **Environment problem on this machine, not smith's.** With Dart 3.12.0 on
  Windows, `dart analyze` (so the `lint` task) exited with code 4 on every
  project: `FileSystemException: Cannot delete file, path =
  '...\AppData\Local\Dart\perf\<pid>' (OS Error: ..., errno = 1920)`, because
  `%LOCALAPPDATA%\Dart\perf` held about 170 reparse-point files that cannot be
  deleted. The same failure happens running `dart analyze` by hand. Running with a
  fresh `LOCALAPPDATA` (and `PUB_CACHE` pointing at the real cache) made every
  `lint` pass with `No issues found!`; all other outputs here were taken that way.

## Verified

Windows 11 Pro, smith 1.62.0, Dart 3.12.0 (stable), Flutter 3.44.0 (stable,
DevTools 2.57.0).

Run: `smith new`, `smith add lib|app|package --lang dart`,
`smith add app mobile --stack flutter` (a real `flutter create`),
`smith add` refusals (`--stack flutter` for a lib or package), `smith projects`,
`smith graph`, `smith explain`, `smith run lint format test build` for Dart
and Flutter projects, `smith run start -p tool`, `smith doctor` (plain, and with
a `[toolchains] dart` mismatch), `smith ci generate` for GitHub and Azure
(a workspace with a Flutter app), with `.flutter-version`, and with
`publish_to` set to a URL and to `none`, `smith ci check`, `smith release
--dry-run` and `--json`, `smith install` (all forms), `smith uninstall -w`,
`smith sync --check`, `smith up --json`, `smith remove`.

Not verified: a real publish to pub.dev or a private repository; the pub
workflow on GitHub; `flutter run -d web-server`; the Dart-only (no Flutter) CI
output of Azure and GitLab; a Flutter library; the live pub.dev query beyond
the doctor check passing; any Dart older than 3.12.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md) and [stacks](../../../reference/stacks.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Configuration](../../../reference/configuration.md)
- [Writing a language plugin](../../creating/language-plugin.md)
