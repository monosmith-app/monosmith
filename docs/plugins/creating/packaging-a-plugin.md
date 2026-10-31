# Packaging and distributing a plugin

A plugin is distributed as a directory with a `plugin.json` that smith finds on the plugin search
path, or as a command written in `smith.toml`. This page covers the file, the naming, one binary per
platform, versions, packing the plugin with `smith plugin pack`, and the ways to get it onto other
machines. Installing a plugin by name from a registry is not decided
([#164](https://github.com/monosmith-app/monosmith/issues/164)): today a plugin is installed with
`smith plugin install` from an archive, a URL or a GitHub release, and anything else that puts a
directory with a `plugin.json` on the search path is an install channel too.

## Two ways to be found

| | `plugin.json` on the search path | `command` in `smith.toml` |
|---|---|---|
| declared by | the plugin's author, in the plugin's directory | the workspace, per plugin |
| smith.toml | `[plugins.external.<name>]` (optionally with `version`), or nothing for a convention you do not configure | `[plugins.external.<name>] command = [...]` |
| best for | a plugin installed on a machine or shared by a team | a plugin that lives in the workspace's own repository |
| per-platform binaries | yes: `{os}`, `{arch}`, `{exe}` | write the command for the machine |

## `plugin.json`

```json
{
  "name": "widget",
  "version": "0.1.0",
  "description": "projects described by a widget.json",
  "command": ["./bin/smith-plugin-widget-{os}-{arch}{exe}"],
  "env": ["NPM_TOKEN"]
}
```

| Field | Rule |
|---|---|
| `name` | required; the plugin's name. **It must be the `plugin.name` the program returns in `initialize`**, or smith refuses it: `the plugin calls itself something else; declare it under its own name` |
| `command` | required; an argv. `{os}` and `{arch}` are Go's `GOOS` and `GOARCH` (`windows`, `linux`, `darwin`, `amd64`, `arm64`, ...); `{exe}` is `.exe` on Windows and empty elsewhere. **Only the first element** is made relative to the plugin's directory, and only when it starts with `./` or `../`; later elements are passed as they are, to a process started in the directory where you ran smith |
| `version` | shown by `smith plugin list`. Keep it equal to the version the program reports: that is the one `version` constraints in `smith.toml` and recipes are checked against, and it goes into task cache keys |
| `description` | shown by `smith plugin list` |
| `env` | optional; the environment variables the plugin needs, as names or prefixes ending in `*`. See [Declaring the environment a plugin needs](#declaring-the-environment-a-plugin-needs) |

A `plugin.json` without a `name` or a `command` fails with `<dir>\plugin.json: name and command are
required`, and stops `smith plugin list` altogether.

### Declaring the environment a plugin needs

smith starts an external plugin with a small environment, not the shell's: `PATH`, the home and
temporary directories, locale, terminal and proxy variables, `SMITH_PLUGIN`, and on Windows the
variables a process needs to run. The full list is in [Plugin
security](../using/plugin-security.md#the-environment-an-external-plugin-gets). Anything else your plugin
reads has to be declared, or it is not there:

```json
{
  "name": "widget",
  "version": "0.1.0",
  "command": ["./bin/smith-plugin-widget-{os}-{arch}{exe}"],
  "env": ["NPM_TOKEN", "AWS_*"]
}
```

- Each entry is a variable name, or a prefix ending in `*` (`AWS_*` is every variable that starts with
  `AWS_`). A variable that is not set is not passed.
- Declare what the plugin needs and no more. The declaration is shown to the person who is asked to
  trust the plugin, and by `smith plugin list`; a plugin that asks for everything will not be trusted.
- A manifest cannot ask for the variables smith uses for its own cache and releases (`SMITH_CACHE_KEY`,
  `SMITH_RELEASE_*` and the like). The workspace owner can grant those, by name, in `smith.toml`.
- The workspace owner can add to the list for one workspace with `env = [...]` under
  `[plugins.external.<name>]`; the two lists are joined.
- A plugin declared with a `command` in `smith.toml` has no `plugin.json`; its variables come from
  the `env` key there.
- Test with the environment your users will have: `smith plugin check ./bin/smith-plugin-widget --env
  NPM_TOKEN` runs the check with the baseline plus the variables you name, and needs no trust. Without
  `--env` your plugin sees the baseline only, which is what a user who declared nothing gets.

Users must also trust the plugin before smith starts it: on first use in a terminal they are shown
its command, the SHA-256 of its file and the variables above, and asked. A new build of your plugin
is a new file, so they are asked again; CI runs pin the digest with `smith plugin trust <name> --pin`.
See [Installing, enabling and configuring plugins](../using/managing-plugins.md#trusting-an-external-plugin).

### Where it goes

The directory sits **directly** under a search directory, in this order, the first match for a name
winning:

1. `<workspace>/.smith/plugins`, committed to the repository or provisioned by a setup script;
2. each directory of `SMITH_PLUGIN_PATH` (`;` on Windows, `:` elsewhere);
3. `$SMITH_HOME/plugins` (`SMITH_HOME` defaults to `~/.smith`).

The lookup is by the `name` inside `plugin.json`: the directory may be called anything, and the plugin
`widget` in a directory called `anything` is found. A directory one level deeper (`group/widget`) is
not:

```text
$ SMITH_PLUGIN_PATH=.../search smith plugin list
Installed (search path: .../search, C:\Users\you\.smith\plugins):
  widget         0.1.0      .../search/anything
$ SMITH_PLUGIN_PATH=.../search2 smith plugin check widget
smith: plugin "widget" is not installed: no widget/plugin.json in .smith\plugins, .../search2, C:\Users\you\.smith\plugins; give its command in smith.toml, or install it there
```

## Naming

smith puts few rules on names; these keep them workable.

| Thing | Rule |
|---|---|
| plugin name | lowercase words joined by dashes (`widget-lint`). It is a key in `smith.toml` (`[plugins.external.widget-lint]`, `[conventions.widget-lint]`, `[projects.x.settings.widget-lint]`) and an entry in stacks and recipes. Use the same name in `plugin.json`, in `initialize` and in the docs |
| a built-in name | an external plugin with the name of a built-in one **replaces** it |
| executable | `smith-plugin-<name>` by convention, plus `-<os>-<arch>` for a directory that carries several, plus `.exe` on Windows |
| language names | the `languages` you declare are what `--lang` accepts and what other plugins' `languages` and `requires` match. Do not take a built-in plugin's (`node`, `go`, ...) unless you mean to replace it |
| stacks | unique across the workspace; a stack in `smith.toml` wins over a published one of the same name |

## One binary per platform

For a Go plugin, build one binary per platform into the directory `bin/`:

```sh
GOOS=linux   GOARCH=amd64 go build -o bin/smith-plugin-widget-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o bin/smith-plugin-widget-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o bin/smith-plugin-widget-windows-amd64.exe .
```

and let `plugin.json` pick with `./bin/smith-plugin-widget-{os}-{arch}{exe}`. The Windows build and the
placeholder expansion were run for this page (the directory above: `smith plugin check widget` passes
with only `smith-plugin-widget-windows-amd64.exe` present); the cross builds were not.

For an interpreted plugin (Node.js, Python) ship the script and a launcher, and make the launcher the
first element of `command`: `./plugin.py` (executable, with a shebang) on Linux and macOS,
`./plugin.cmd` on Windows. `["python", "plugin.py"]` does not work from an installed directory,
because the second element is not made relative. The `plugin.json` files in the Node.js, TypeScript and
Python examples show the Unix launcher; a Windows install uses the `.cmd`. A TypeScript plugin runs
as it is on Node.js 24 (`./plugin.ts` starting with `#!/usr/bin/env node`), but not from a folder
called `node_modules`, where Node.js refuses to strip types: install it elsewhere, or build it to
JavaScript first. Ship its dependency, `@monosmith/plugin-sdk`, with it (`npm install` in the plugin's
directory). A Python plugin on the SDK needs `monosmith-plugin-sdk` importable by the interpreter that
runs it: `pip install monosmith-plugin-sdk` into the environment the launcher's `python` belongs to (a
virtual environment, whose interpreter the launcher names).

A C# plugin is a program that needs the .NET runtime. Publish it as one file, so that `command` is a single
executable whose digest covers the whole plugin: `dotnet publish -c Release -p:PublishSingleFile=true -o publish`, then
`"command": ["./LanguageWidget"]` in `plugin.json` (the `.exe` on Windows, which the `{exe}` placeholder gives). A
plugin that is the assembly run by `dotnet` (`["dotnet", "Plugin.dll"]` in `smith.toml`) cannot be a `plugin.json`
for the same reason `["python", "plugin.py"]` cannot: the second element is not made relative; and the digest
`smith plugin trust --pin` writes covers `dotnet` and the `.dll`, not the SDK's assembly next to it.

## Versions

smith checks the version a plugin reports in `initialize`:

| Where | What is checked |
|---|---|
| `[plugins.external.<name>] version = "^0.1"` | the plugin's version meets the constraint: `1.2.3` (exactly), `^1.2` (same major, at least 1.2), `~1.2` (same major and minor, at least 1.2), `>=1.2`, or empty/`*` (any) |
| a recipe's `use.lint = "widget-lint@^1.2"` | the same constraint, when the recipe is checked and used |
| task cache keys | the version of every plugin that shapes a project is part of the key: **upgrading a plugin re-runs its tasks**, so change the version when behaviour changes |

A version is `major[.minor[.patch]]`; a leading `v` and a `-pre` or `+build` suffix are ignored; anything
else fails with `plugin version "x": not a version`. Use semantic versioning: a breaking change to
what the plugin does is a new major, so users who wrote `^1` do not pick it up unasked.

The **protocol** has its own version (`protocol.Version`, 1; `PROTOCOL_VERSION` in the TypeScript and Python SDKs, `ProtocolInfo.Version` in the C# one,
and `x-protocolVersion` in the JSON Schema). A plugin on an SDK is built against a version of the SDK; if smith ever stops speaking the protocol version the plugin was built for, the
handshake fails and you rebuild with a newer SDK. Additions to the protocol are optional fields and
new capabilities, which old plugins ignore, so an old plugin keeps working with a newer smith.

## Pack it

`smith plugin pack` turns the plugin directory (with the binaries already built into it) into a
single archive that `smith plugin install` verifies and unpacks:

```sh
smith plugin pack ./widget --out dist
```

```text
Packed widget 0.1.0: 3 files, 14852081 bytes.
  archive  dist/widget-0.1.0.smith-plugin.zip
  sha256   db7fb4c07eeba779d2dd7dc313873cd180f8dbe133f8e4bb76c9c6956d22204d  (also in widget-0.1.0.smith-plugin.zip.sha256)
  program  bin/smith-plugin-widget-windows-amd64.exe
The plugin was not started. Publish the archive and its .sha256 file together; users install it with `smith plugin install`.
```

- `plugin.json` needs a `name` and a `version`; they name the archive, `<name>-<version>.smith-plugin.zip`.
- Everything in the directory goes in (except `.git`, earlier archives, an old `checksums.sha256`
  and what a `.smithpluginignore` file at the directory's root lists, one pattern per line such as
  `src/` or `*.log`), stored uncompressed in name order with a fixed timestamp, so packing the same
  directory twice gives the same bytes. Symbolic links are refused. List sources, secrets and build
  leftovers in `.smithpluginignore`, or keep them out of the directory.
- Publish the archive's `.sha256` with it: `smith plugin install` refuses an archive that has no digest
  to be checked against.
- The files that the first element of `command` names for any platform are marked executable, so
  a package built on Windows installs runnable on Linux. If the `command` names a `./` program and no
  file matches, packing fails: build for at least one platform first.
- smith writes `checksums.sha256` into the archive (`<sha256>  <path>` for every file) and the archive's own
  digest to `<archive>.sha256` beside it. `pack` does not run the plugin: run `smith plugin check` before
  you pack.
- Pack one archive per release with a binary for every platform you support. A user whose platform is
  missing gets `the plugin has no program for this platform` and nothing is installed.

To publish, attach the archive and its `.sha256` file to a GitHub Release. Users then run:

```sh
smith plugin install github:<owner>/<repo>@<tag>
```

and smith compares the archive with the digest the release publishes. Tell users the digest through
another channel too (release notes, your site) so they can pass `--sha256` themselves. A GitHub
release must have exactly one asset ending in `.smith-plugin.zip`; with several, users name one with
`github:<owner>/<repo>@<tag>#<asset>`. Assets of private repositories are not supported.

[`smith plugin pack`](../../manual/plugin-pack.md) and [`smith plugin
install`](../../manual/plugin-install.md) list every check, flag and message. Installing never starts
your plugin, and does not make users trust it: they are still asked the first time smith needs it.

## Ways to distribute

| Way | How | Trade-off |
|---|---|---|
| in the repository | commit `.smith/plugins/<name>/` (plugin.json and the binaries), or the plugin's source as a project of the workspace and a `command` in `smith.toml` that builds and runs it (`["go", "run", "./apps/my-plugin"]`, what the `go-smith-plugin` scaffold writes) | nothing to install; binaries in git are heavy, and `go run` needs Go and is slow to start |
| a packed archive | `smith plugin pack`, attach the archive and its `.sha256` to a GitHub Release (or host it anywhere over https); users run `smith plugin install github:owner/repo@tag` or give the URL and `--sha256` | one archive carries every platform; digests are checked and the pin written for CI; no signature yet |
| a package manager | publish the directory as an npm, pip or other package whose install step puts it under `$SMITH_HOME/plugins` or a directory on `SMITH_PLUGIN_PATH` | reuses the registry; the install step is yours |
| a setup script / dev container | copy or build into `.smith/plugins` in the machine setup | fits a team |

Whatever the channel, finish with the check smith itself runs:

```sh
smith plugin list            # is it found, and which version?
smith plugin check <name>    # does it start and conform?
```

## Releasing the plugin itself

A plugin written in Go with the `go-smith-plugin` scaffold is an app project; its source can live in
a workspace and be released like any other app. A plugin that is a library other plugins import (a
shared package of rules) is a normal package of its language. Nothing about being a plugin changes
how it is released.

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `plugin "x" is not installed: no x/plugin.json in ...` | the directory is not directly under a search directory, or `plugin.json` is missing; the message lists every directory looked in |
| `<dir>\plugin.json: name and command are required` | an incomplete manifest; it breaks `smith plugin list` for everyone on that path |
| `smith.toml [plugins.external.x]: the plugin calls itself something else; declare it under its own name` | the program's `plugin.name` differs from `plugin.json`'s `name` / the `smith.toml` key |
| `smith.toml [plugins.external.x] wants version ^2, but the plugin is 0.1.0` | the installed plugin is older (or newer) than the constraint; the `version` in `plugin.json` does not matter here, the program's does |
| `starting plugin [...]: fork/exec ...: %1 is not a valid Win32 application.` | the first element of `command` is a script; use a `.cmd` launcher |
| the plugin works from the repository and not when installed | a later element of `command` is a relative path; only the first is resolved against the plugin's directory |
| a built-in behaviour changed after installing | an external plugin with a built-in's name replaces it |
| the plugin cannot find a token or a cloud credential it used to read | it is not in `env` of the `plugin.json`: external plugins start with a minimal environment |
| `✖ the external plugin "x" is not trusted yet, so smith did not start it` | the workspace has not trusted this command and file yet: `smith plugin trust x` |

## See also

- [Installing, enabling and configuring plugins](../using/managing-plugins.md)
- [Listing a packed plugin in the marketplace](../using/marketplace.md#publish-a-plugin)
- [Troubleshooting plugins](../using/troubleshooting-plugins.md)
- [Testing a plugin](testing-a-plugin.md)
