# Packaging and distributing a plugin

A plugin is distributed as a directory with a `plugin.json` that smith finds on the plugin search
path, or as a command written in `smith.toml`. This page covers the file, the naming, one binary per
platform, versions, and the ways to get a plugin onto other machines. How plugins will one day be
published and installed by name from a registry is not decided
([#164](https://github.com/monosmith-app/monosmith/issues/164)): today, anything that puts a directory
with a `plugin.json` on the search path is an install channel.

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
  "command": ["./bin/smith-plugin-widget-{os}-{arch}{exe}"]
}
```

| Field | Rule |
|---|---|
| `name` | required; the plugin's name. **It must be the `plugin.name` the program returns in `initialize`**, or smith refuses it: `the plugin calls itself something else; declare it under its own name` |
| `command` | required; an argv. `{os}` and `{arch}` are Go's `GOOS` and `GOARCH` (`windows`, `linux`, `darwin`, `amd64`, `arm64`, ...); `{exe}` is `.exe` on Windows and empty elsewhere. **Only the first element** is made relative to the plugin's directory, and only when it starts with `./` or `../`; later elements are passed as they are, to a process started in the directory where you ran smith |
| `version` | shown by `smith plugin list`. Keep it equal to the version the program reports: that is the one `version` constraints in `smith.toml` and recipes are checked against, and it goes into task cache keys |
| `description` | shown by `smith plugin list` |

A `plugin.json` without a `name` or a `command` fails with `<dir>\plugin.json: name and command are
required`, and stops `smith plugin list` altogether.

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
because the second element is not made relative. The `plugin.json` files in the Node.js and Python
examples show the Unix launcher; a Windows install uses the `.cmd`.

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

The **protocol** has its own version (`protocol.Version`, 1). A plugin on the Go SDK is built against
a version of the SDK; if smith ever stops speaking the protocol version the plugin was built for, the
handshake fails and you rebuild with a newer SDK. Additions to the protocol are optional fields and
new capabilities, which old plugins ignore, so an old plugin keeps working with a newer smith.

## Ways to distribute

| Way | How | Trade-off |
|---|---|---|
| in the repository | commit `.smith/plugins/<name>/` (plugin.json and the binaries), or the plugin's source as a project of the workspace and a `command` in `smith.toml` that builds and runs it (`["go", "run", "./apps/my-plugin"]`, what the `go-smith-plugin` scaffold writes) | nothing to install; binaries in git are heavy, and `go run` needs Go and is slow to start |
| a release asset | attach a zip of the directory to a GitHub Release (smith can attach an app's build to a release; see [Releasing](../../reference/releasing.md)) and unpack it into `$SMITH_HOME/plugins` | one download per platform; you write the unpack step |
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

## See also

- [Installing, enabling and configuring plugins](../using/managing-plugins.md)
- [Troubleshooting plugins](../using/troubleshooting-plugins.md)
- [Testing a plugin](testing-a-plugin.md)
