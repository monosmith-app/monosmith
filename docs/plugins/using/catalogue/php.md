# The php plugin

The php plugin makes the Composer packages of a workspace smith projects. Every
package lives in its own directory with its own `composer.json`; the workspace
root has one `composer.json` that requires them all from their directories, so
the whole workspace shares one `vendor/` and one `composer.lock`. The plugin
finds the packages, gives each one PHPStan (lint), PHPUnit (test) and, for an
app, a `start` task, creates new packages, writes release versions into their
manifests, installs and removes Composer dependencies, and tells the CI
pipelines which PHP to install.

## At a glance

| | |
|---|---|
| Plugin name | `php` (built in; runs as a separate process, `smith plugin serve php`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` value | `php` |
| Manifest that makes a directory a candidate project | `composer.json` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`, `dependencies` |
| Not advertised | `declarations` and `latest`: `smith sync` and `smith up` skip PHP |
| Stacks it publishes | none. The `php` language alone is the stack; a project records no stack unless `extra.smith.stack` names one |
| Project kinds `smith add` can create | `app`, `lib`, `package` (`smith kinds` lists `php-app`, `php-lib`, `php-package`) |
| Tasks | `lint`, `test`, `start`; no `build`, `typecheck`, `format` or `package` |

The roles and capabilities above are what `smith plugin list --json` prints for
`php`; `languages`, `manifests` and the absence of requires, conflicts and
provides come from the plugin's `Describe` method
(`libs/plugin-php/plugin/php_plugin_handler.go`).

## What it detects

smith hands the plugin every `composer.json` it found in the workspace. It
finds them in the files git lists (tracked and untracked, not ignored), so the
`vendor/` directory the scaffold adds to `.gitignore` is never searched
([project discovery](../../../architecture/project-discovery.md)). For each one:

- **Root manifest.** The `composer.json` at the workspace root only gathers the
  packages, so it is not a project (`discovery.Discover` skips a manifest whose
  directory is `.`).
- **Project name.** The name of the directory, not the Composer name:
  `packages/sdk` is the project `sdk`.
- **Exports.** The Composer `name` (`php-demo/sdk`). If `name` is empty, the
  directory name is used.
- **Imports (graph edges).** Every key of `require` and `require-dev`, except
  platform requirements: `php` and anything without a `/` (`ext-json`). smith
  matches those names against other projects' exports, so a package that
  requires `php-demo/util` depends on the `util` project, and `util`'s tasks
  run first when a task depends on `^` projects.
- **Kind.** `extra.smith.kind` when it is `app`, `lib` or `package`; otherwise
  the directory it sits under (`apps/` is an app, `libs/` a lib, `packages/` a
  package); otherwise `lib`.
- **Version.** The `version` field of the package's `composer.json`
  (`0.0.0` in a new package).
- **Description.** The manifest's `description`.
- **Stack.** `extra.smith.stack`, empty when absent.
- **Toolchain.** The first version number printed by `php --version`, reported
  as `php 8.4.25`. When PHP is missing, the toolchain is left out of the
  project.
- **Tags.** The plugin sets none.
- **Tests.** A package has tests if it has a `tests/` directory.
- **Executable.** The first entry of the manifest's `bin` list, used by `start`.

A real run, with three packages and `packages/sdk` requiring `php-demo/util`:

```text
$ smith projects
cli                  app      php      apps/cli
sdk                  package  php      packages/sdk
util                 lib      php      libs/util

$ smith graph
cli  app · php · apps/cli
sdk  package · php · packages/sdk
  └─ util
util  lib · php · libs/util
```

`"ext-json": "*"` in the same `require` block produced no edge.

## Tasks

The plugin contributes three tasks. All of them run in the package's directory
and use the workspace's `vendor/`.

| Task | Command | Inputs | Outputs | dependsOn (from the plugin) | Cached | Appears |
|---|---|---|---|---|---|---|
| `lint` | `php {workspaceRoot}/vendor/bin/phpstan analyse --no-progress --memory-limit=1G --level 8 --autoload-file {workspaceRoot}/vendor/autoload.php src tests` | `**/*`, `!vendor/**`, `{workspaceRoot}/composer.json`, `{workspaceRoot}/composer.lock` | none | none | yes | every package; `tests` is analysed only when `tests/` exists |
| `test` | `php {workspaceRoot}/vendor/bin/phpunit --bootstrap {workspaceRoot}/vendor/autoload.php tests` | same as `lint` | none | none | yes | only when the package has a `tests/` directory |
| `start` | `php <bin>` (for example `php bin/cli`) | every file of the project | none | none | no | only when the kind is `app` and the manifest's `bin` is not empty |

PHPStan runs at level 8 over `src` and `tests`. There is nothing to build in
PHP, so a PHP project has no `build` task; `smith run build` in the demo
workspace printed `no selected project has a [build] task`.

If the workspace enables the `vertical-slices` convention (the default
`smith.toml` of `smith new` does), every PHP project also gets a `slice-check`
task and `lint` depends on it. That task is the convention's, not the php
plugin's.

`smith explain` for a lib and an app, from a real workspace:

```text
$ smith explain util
util has 3 tasks:
  lint         php {workspaceRoot}/vendor/bin/phpstan analyse --no-progress --memory-limit=1G …  (cached, after slice-check)
  slice-check  smith slices check --lang php .  (cached)
  test         php {workspaceRoot}/vendor/bin/phpunit --bootstrap {workspaceRoot}/vendor/autol…  (cached)

$ smith explain cli:start
cli:start — Run the app
  command    php bin/cli
  cwd        apps/cli
  dependsOn  —
  inputs     **/* (every file of the project)
  outputs    —
  env        —
  cache      false
```

Real runs of the tasks:

```text
$ smith run test
...
util:test │ OK (1 test, 1 assertion)
cli:test  │ OK (1 test, 1 assertion)
sdk:test  │ OK (1 test, 1 assertion)
✔ 3/3 tasks succeeded (0 from cache) in 211ms

$ smith run start -p cli
cli:start ▶ php bin/cli
cli:start │ Hello, world!
```

## Creating a project

```sh
smith add lib util --lang php
smith add app cli --lang php
smith add package sdk --lang php
```

The first package of a workspace also writes the root `composer.json`. The real
output for the first one (the composer output in the middle is trimmed):

```text
$ smith add lib util --lang php
Adding lib util (php) in libs/util
./composer.json has been updated
Loading composer repositories with package information
Updating dependencies
...
Wrote 8 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .gitignore
  composer.json
  composer.lock
  libs/util/composer.json
  libs/util/src/Greeting/GreetUseCase.php
  libs/util/tests/Greeting/GreetUseCaseTest.php
Require it from another package with "php-demo/util": "*" (the workspace resolves it from its directory).
Try: smith test -p util
```

An app additionally gets `bin/<name>`; the second package onward does not
rewrite the root manifest's tools:

```text
$ smith add app cli --lang php
...
Wrote 6 file(s):
  apps/cli/bin/cli
  apps/cli/composer.json
  apps/cli/src/Greeting/GreetUseCase.php
  apps/cli/tests/Greeting/GreetUseCaseTest.php
  composer.json
  composer.lock
```

What is generated:

- **A package** (`libs/util/composer.json`): Composer name `<workspace>/<name>`
  (the workspace name and the project name, lower-cased, runs of other
  characters turned into one dash), `type` `library` (`project` for an app),
  `version` `0.0.0`, `require` `php >=8.2`, a PSR-4 autoload of
  `<Workspace>\<Name>\` to `src/`, and `extra.smith.kind` (plus `stack` when one
  was given).
- **A worked example slice**, `src/Greeting/GreetUseCase.php` (a final class
  with a static `greet(string $name): string` that returns `Hello, {$name}!`)
  and `tests/Greeting/GreetUseCaseTest.php`, a PHPUnit test of it. A namespace
  and folder per outcome, a class per role, tests mirrored under `tests/`, as
  in the vertical-slice layout.
- **For an app**, `bin/<name>`, which requires the Composer autoloader and
  prints `Hello, world!`, and a `"bin": ["bin/<name>"]` entry in the manifest.
- **The root `composer.json`**, written once by the first package: name
  `<workspace>/workspace`, `type` `project`, `require` `php >=8.2`,
  `require-dev` `phpunit/phpunit ^11.5` and `phpstan/phpstan ^2.1`,
  `minimum-stability` `dev` with `prefer-stable`. Each package is then added
  with `composer config repositories.<dir-dashed> path <dir>` and
  `composer require <name>:@dev --no-update`, followed by `composer update`.
  The root manifest ends up like this:

```json
"require": {
    "php": ">=8.2",
    "php-demo/util": "@dev",
    "php-demo/cli": "@dev",
    "php-demo/sdk": "@dev"
},
...
"repositories": [{
    "name": "packages-sdk",
    "type": "path",
    "url": "packages/sdk"
},{
    "name": "apps-cli",
    "type": "path",
    "url": "apps/cli"
},{
    "name": "libs-util",
    "type": "path",
    "url": "libs/util"
}]
```

- **`.gitignore`** gets a `# PHP` block with `/vendor/` unless `vendor` is
  already ignored.
- **The CI files** are regenerated, so the new workspace installs PHP.

`smith add` fails if the target directory already exists. It needs `composer`
on `PATH`. The `kind` can be `app`, `lib` or `package`; no stacks are involved.
Composer's output is shown only when it fails or with `--verbose`; otherwise
smith prints one `Running composer ...` line per command (the plugin runs
Composer through `toolrun.Quiet`).

## Settings

The plugin has no `[projects.<p>.settings.php]` keys and reads no environment
variables of its own. What can be configured:

| Where | Key | Effect | Read in |
|---|---|---|---|
| `composer.json` | `extra.smith.kind` | `app`, `lib` or `package`; overrides the directory convention | `discovery/composer_manifest_mapper.go` |
| `composer.json` | `extra.smith.stack` | records the project's stack | same |
| `composer.json` | `bin` | the first entry is what `start` runs; an app without it has no `start` | `discovery/discover_packages_use_case.go` |
| `composer.json` | `version` | the version the release starts from and rewrites | `releasing/manifest_version_mapper.go` |
| `.php-version` (workspace root) | first line, for example `8.3` | the PHP version CI and the dev container install; default `8.4` | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[toolchains] php = "8.4"` | a pin that `smith doctor` checks against the installed PHP (see Toolchain) | host, not the plugin |
| `smith.toml` | `[projects.<name>] tag_prefix`, `split_repo`, `registry` | release settings every project has; see Releasing | host |
| `smith.toml` | `[tasks.lint]`, `[projects.<name>.tasks.test]` | override any task the plugin infers | host ([tasks and cache](../../../reference/tasks-and-cache.md)) |

The scaffold constants are not configurable: PHP `>=8.2`, PHPUnit `^11.5`,
PHPStan `^2.1` (`scaffolding/package_templates_config.go`). Edit the root
`composer.json` afterwards to change them.

## Toolchain

- **Probe.** The project's toolchain is `php <version>`, from the first version
  number in the output of `php --version`. The answer is cached per plugin
  process. The installed version is part of every task's cache key, so a
  machine on another PHP never reuses a cached result.
- **Pin.** `[toolchains] php = "8.4"` (a prefix on dot boundaries) makes
  `smith doctor` fail when the installed PHP differs. A real run with the pin
  set to `8.3` on a machine running 8.4.25:

```text
✖ php matches the pinned version 8.3
    [toolchains] pins php 8.3, but this machine runs 8.4.25: tasks cached on another version are misses, and results can differ from CI
    → install php 8.3 (`mise use php@8.3`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] php
```

- **The pin does not reach CI.** With `[toolchains] php = "8.3"` and no
  `.php-version`, the generated GitHub workflow still said
  `php-version: '8.4'`. Only `.php-version` changes the CI version (with
  `8.3` in it the workflow said `'8.3'`). Keep the two in step by hand.
- **What must be installed.** `php` and `composer` on `PATH`.
- **Where output goes.** Nothing is built, so nothing is written under `dist/`.
  Composer writes `vendor/` and `composer.lock` at the workspace root.

See [configuration: toolchains](../../../reference/configuration.md#toolchains).

## Releasing

| | |
|---|---|
| Version in the manifest | the `version` field of the package's `composer.json` |
| Version write | at release time `version` is replaced in place (formatting kept); if there is no `version` field, one is added after the `name` line. The manifest is restored afterwards: the tag is the record. A dry run reports the file and writes nothing |
| Tag | `<project>@<version>`, for example `sdk@0.1.0`; change it with `[projects.<name>] tag_prefix` |
| Publish | nothing is uploaded. For a package the result is "Packagist publishes from the repository's tags" (registry `packagist.org`); for other kinds "released by its tag" |
| Registry query | `Published` asks `https://repo.packagist.org/p2/<vendor>/<name>.json` for the newest stable version (versions with `-`, `@` or a `dev` prefix are skipped, a leading `v` is dropped). A 404 means "not published"; a network failure or other status means "unreachable", not an error |
| Credentials | none |
| Private registry | not supported by the plugin |
| Dry run | `smith release --dry-run` plans without tagging; it does not call Packagist |

The tag `sdk@0.1.0` is not one Composer reads as a version. For Packagist to
see versions, mirror the package to a repository of its own: set
`split_repo` and each release pushes the package's directory there, tagged
`v<version>`. A real dry run:

```text
$ smith release --dry-run
Dry run: nothing is published or tagged.
  sdk                  unreleased → 0.1.0  (1 releasing commit(s))
sdk: would mirror packages/sdk to acme/sdk-php@v0.1.0
```

The JSON report names the file it would stamp (`packages/sdk/composer.json`)
and the registry (`packagist.org`). `SPLIT_REPO_TOKEN`, the mirror's token, is
described in [releasing: split repositories](../../../reference/releasing.md#split-repositories).
You register the mirror with Packagist yourself once.

## CI

`smith ci generate` adds, when the workspace has at least one PHP project:

| Provider | What is added |
|---|---|
| GitHub Actions | `shivammathur/setup-php@v2` with `php-version` and `tools: composer:v2`; `actions/cache@v4` for `~/.cache/composer`, keyed on `**/composer.lock` |
| Azure Pipelines | a bash step installing the distribution's PHP (`php-cli php-xml php-mbstring php-curl php-zip unzip`) and Composer's installer; a `Cache@2` step; `COMPOSER_CACHE_DIR` as a pipeline variable |
| GitLab CI | the same install commands in `before_script`; a `cache` on `**/composer.lock`; `COMPOSER_CACHE_DIR` |
| Dev container | the `ghcr.io/devcontainers/features/php:1` feature with `installComposer` and the version |

The PHP version is the first line of `.php-version`, else `8.4`
(`PHPVersion` in `pipeline/ci_contribution_use_case.go`). The setup step
(`smith ci setup`) runs `composer install --no-interaction --no-progress` and
stops the job if it fails. The plugin adds no audit step (`smith ci audit`
said "nothing to audit") and needs no secret or permission: Packagist reads
tags.

The GitHub fragment from a real run:

```yaml
      - uses: shivammathur/setup-php@v2
        with:
          php-version: '8.4'
          tools: 'composer:v2'
      - name: Cache the composer download store
        uses: actions/cache@v4
        with:
          path: ~/.cache/composer
          key: smith-composer-${{ runner.os }}-${{ hashFiles('**/composer.lock') }}
```

See [CI pipelines](../../../reference/ci.md).

## Doctor, dependencies and removal

**Doctor.** Two checks, run only when the workspace has PHP projects:
`PHP is installed` (`php --version`) and `Composer is installed`
(`composer --version`). Each failure names the remedy (install PHP at the
`.php-version` version, install Composer). Both passed in my run. On my
machine `composer --version` also printed a Composer warning about the PHP
build to stderr; the check only looks at the exit status.

**Install and uninstall.**

| Command | What the plugin runs |
|---|---|
| `smith install` (no package) | `composer install --no-interaction --no-progress` at the root; the file it reports is `composer.lock` |
| `smith install -w util monolog/monolog` | `composer require monolog/monolog --no-update --no-interaction` in the package's directory (`--dev` for `-D`), then `composer update` at the root |
| `smith uninstall -w util monolog/monolog` | `composer remove ... --no-update` in the package, then `composer update` at the root |

Real run of the last two: both changed only `libs/util/composer.json` and
`composer.lock`, and printed
`✔ util: added monolog/monolog (libs/util/composer.json, composer.lock)` and
`✔ util: removed monolog/monolog (...)`.

**Not supported.** `smith sync` and `smith up` skip PHP: `smith sync --check`
printed `⊘ php: not compared; its plugin does not list declarations`, and
`smith up --json` found no updates.

**Remove.** `smith remove <project>` deletes the directory, then the plugin
runs `composer remove <name> --no-update` and
`composer config --unset repositories.<dir-dashed>` at the root, then
`composer update` so `composer.lock` stops naming the package too (`--dry-run`
only lists `composer.json` and `composer.lock`).

## Requirements and limits

- PHP and Composer on `PATH`. The scaffold needs PHP 8.2 or newer.
- No `build`, `typecheck`, `format` or `package` task, and no tests task for a
  package without `tests/`.
- The php plugin does not upload anything; Packagist needs a mirror repository
  (`split_repo`) to read versions.
- The `[toolchains] php` pin is checked by doctor but not used by CI generation
  (see Toolchain).
- The root `composer.json` that `smith add` writes sets
  `minimum-stability: dev`, so workspace packages resolve at `@dev`.

## Verified

Windows 11 Pro, smith 1.62.0, PHP 8.4.25 (cli, ZTS, Visual C++ 2022 x64),
Composer 2.10.3, PHPUnit 11.5.57, PHPStan 2.3.2 (installed by the scaffold).

Run: `smith new`, `smith add lib|app|package --lang php`, `smith projects`,
`smith graph`, `smith explain` (project and task), `smith run lint`,
`smith run test`, `smith run start -p cli`, `smith doctor` (with and without a
mismatching pin), `smith ci generate` for GitHub, Azure and GitLab,
`smith ci setup`, `smith ci audit`, `smith install -w`, `smith uninstall -w`,
`smith sync --check`, `smith up --json`, `smith remove`,
`smith release --dry-run` and `--json`, with and without `split_repo`.

Not verified: a real (non-dry) release; the live Packagist query (`Published`)
beyond reading the code and its unit test; the Azure and GitLab pipelines
running on a real agent; the `ghcr.io/devcontainers/features/php` feature;
`smith install` with no package.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Configuration](../../../reference/configuration.md)
- [Writing a language plugin](../../creating/language-plugin.md)
- [The ruby plugin](ruby.md), [the python plugin](python.md): other languages with one shared resolution
