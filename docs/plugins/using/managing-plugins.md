# Installing, enabling and configuring plugins

This page is every way to change which plugins a workspace uses and how they are set up: the
`smith plugin` subcommands, the `[plugins]` tables of `smith.toml`, installing an external plugin,
and the settings a plugin reads. The model behind it is in
[What a plugin is and how plugins combine](what-is-a-plugin.md).

## The `smith plugin` commands

| Command | Does |
|---|---|
| `smith plugin list [--json]` | lists the built-in plugins and the plugins installed on the search path |
| `smith plugin check <name or path> [--generate] [--json]` | runs the conformance suite against a plugin |
| `smith plugin enable <name>...` | turns built-in language plugins on in `smith.toml` |
| `smith plugin disable <name>...` | turns built-in language plugins off in `smith.toml` |
| `smith plugin set <name>...` | enables exactly these built-in language plugins |

`smith plugin serve <name>` also exists but is hidden: it is what smith runs to start a built-in
plugin, not a command to use.

### List

```sh
smith plugin list
```

```text
PLUGIN           LEVELS                       SLOTS                REQUIRES
bicep            language,toolchain,scaffold  -                    -
cpp              language,toolchain,scaffold  -                    -
...
rust             language,toolchain,scaffold  -                    -
angular          scaffold,toolchain           build,typecheck      language plugin node
astro            scaffold,toolchain           build                language plugin node
docker           deploy                       -                    -
eslint           toolchain                    lint                 language plugin node
go-smith-plugin  scaffold                     -                    language plugin go
jest             toolchain                    test                 language plugin node
playwright       toolchain                    e2e                  language plugin node
typescript       scaffold,toolchain           build,typecheck      language plugin node
vertical-slices  structure                    -                    -
vite             scaffold,toolchain           build,typecheck      language plugin node
vitest           toolchain                    test                 language plugin node
vscode           scaffold,toolchain,deploy    build,typecheck      language plugin node

Installed (search path: ~/work/inst/.smith/plugins, ~/.smith/plugins):
  widget         0.1.0      ~/work/inst/.smith/plugins/widget
```

| Column | Means |
|---|---|
| `PLUGIN` | the plugin's name, as it is written in `smith.toml` and in stacks |
| `LEVELS` | the roles it can be chosen for. A language plugin that declares none shows the default: language, toolchain and scaffold |
| `SLOTS` | the tasks a toolchain owns; two plugins of one project may not share a slot |
| `REQUIRES` | plugins that must be in the same project (`language plugin node`); `not <ref>` marks a conflict |

Things the list does **not** show: a plugin declared in `smith.toml` with a `command` (it is started
from there, not found on the search path), and the capabilities, which `--json` adds:

```sh
smith plugin list --json
```

```json
{
  "name": "playwright",
  "source": "built-in",
  "roles": ["toolchain"],
  "provides": ["e2e"],
  "requires": [{ "level": "language", "name": "node" }],
  "capabilities": { "generate": true, "doctor": true, "contribute": true }
}
```

(abridged: the real output is indented one value per line). `source` is `built-in`, or the
directory an installed plugin was found in; installed plugins show `name`, `version` and `source`
only, since listing does not start them. A `plugin.json` without a `name` or `command` makes the
whole listing fail: `<dir>\plugin.json: name and command are required`.

### Check

`smith plugin check <name | path>` starts a plugin and runs the conformance suite. See
[Testing a plugin](../creating/testing-a-plugin.md) for what each check means.

```sh
smith plugin check go
```

```text
go 1.62.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ plays a known role: toolchain
  ✔ plays a known role: scaffold
  ✔ a language plugin discovers projects
  ✔ a scaffold generates projects
  ✔ discover answers with no manifests
  ✔ ci answers with no projects
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

The first argument is a **path** when a file by that name exists, otherwise a **name**, looked up
as a workspace would: `smith.toml`, then the built-in plugins, then the search path. On Windows a
script is not an executable: check a `.cmd` launcher by its path with a leading `.\`, and use
backslashes (the Git Bash shell eats them: use PowerShell or `cmd`).

### Enable, disable and set

Until `[plugins] enabled` is written in `smith.toml`, every built-in language plugin loads. Writing
it limits the workspace to the names listed. These commands edit that line:

```sh
smith plugin disable go python
```

```text
enabled plugins: bicep, cpp, dart, dotnet, jvm, node, php, recipe, ruby, rust
```

```toml
[plugins]
enabled = ["bicep", "cpp", "dart", "dotnet", "jvm", "node", "php", "recipe", "ruby", "rust"]
```

```sh
smith add lib x --lang go
```

```text
smith: no plugin handles language "go"; available: bicep, cpp, dart, dotnet, jvm, node, php, recipe, ruby, rust
```

| Command | Result |
|---|---|
| `smith plugin enable go` | `enabled plugins: bicep, cpp, dart, dotnet, go, jvm, node, php, recipe, ruby, rust` |
| `smith plugin set node go` | `enabled plugins: go, node` (only these) |
| `smith config unset plugins.enabled` | back to the default: all of them |

| Error | Cause |
|---|---|
| `no built-in plugin "nosuch" (built-in: bicep, cpp, dart, ...)` | the name is not a built-in **language** plugin. The add-ons (`typescript`, `eslint`, `docker`, ...) always load and cannot be switched this way |
| `at least one plugin must stay enabled` | `disable` would leave none |
| `smith.toml enables unknown plugin "nope"; built-in plugins are [...]` | a hand-written `enabled` names a plugin that does not exist |

Disabling a language plugin does not touch existing projects of that language; smith simply stops
finding them, and a stack whose language plugin is not loaded is an error:
`language plugin "x" is not loaded`.

## External plugins

An external plugin is any executable that speaks the protocol. There are two ways to tell smith
about one, and they combine.

### Declare it in `smith.toml`

```toml
[plugins.external.widget]
command = ["C:/tools/smith-plugin-widget.exe"]   # any executable; on PATH, or a path; arguments allowed
version = "^0.1"                                  # optional constraint
```

| Key | Meaning |
|---|---|
| `command` | the argv that starts the plugin. Empty or absent: look the plugin up on the search path by name |
| `version` | the version the plugin must report: `"1.2.3"` (exactly), `"^1.2"` (same major, at least 1.2), `"~1.2"` (same major and minor, at least 1.2), `">=1.2"`; empty or `"*"` is any |

smith checks that the plugin answers to the name it is declared under, and that its version meets
the constraint, as soon as it starts it:

```text
smith: smith.toml [plugins.external.widget] wants version ^2, but the plugin is 0.1.0
smith: smith.toml [plugins.external.other]: the plugin calls itself something else; declare it under its own name
```

A declared plugin with the name of a built-in plugin replaces the built-in one.

### Install it on the search path

An installed plugin is a directory holding a `plugin.json`:

```json
{
  "name": "widget",
  "version": "0.1.0",
  "description": "projects described by a widget.json",
  "command": ["./smith-plugin-widget{exe}"]
}
```

| Field | Meaning |
|---|---|
| `name`, `command` | required |
| `version`, `description` | shown by `smith plugin list` |

In `command`, `{os}`, `{arch}` and `{exe}` (`.exe` on Windows, empty elsewhere) are filled in, so
one directory can carry a binary per platform (`./bin/smith-plugin-widget-{os}-{arch}{exe}`). Only
the first element is made relative to the plugin's directory, and only when it starts with `./` or
`../`; later elements are passed as they are, to a process whose working directory is wherever you
ran `smith`.

Put the directory in one of these, the first match for a name winning:

| Order | Place |
|---|---|
| 1 | `<workspace>/.smith/plugins` |
| 2 | each directory of `SMITH_PLUGIN_PATH` (`;` separated on Windows, `:` elsewhere) |
| 3 | `$SMITH_HOME/plugins` (`SMITH_HOME` defaults to `~/.smith`) |

```sh
mkdir -p .smith/plugins/widget
cp plugin.json smith-plugin-widget.exe .smith/plugins/widget/
smith plugin check widget
```

```text
widget 0.1.0
  ✔ names itself
  ...
  ✔ shuts down when asked
```

Then name it, with or without a constraint, and use it:

```toml
[plugins.external.widget]
version = "^0.1"
```

```text
$ smith add lib hello --lang widget
Adding lib hello (widget) in libs/hello
Wrote 2 file(s):
  libs/hello/README.md
  libs/hello/widget.json
$ smith projects
hello                lib      widget   libs/hello
```

A plugin that is not on the search path and has no command is an error that lists the places it
looked: `plugin "nosuch" is not installed: no nosuch/plugin.json in <dirs>; give its command in
smith.toml, or install it there`.

How plugins will be published and installed by name from a registry is not decided yet
([#164](https://github.com/monosmith-app/monosmith/issues/164)): any channel only has to put a
directory with a `plugin.json` on the search path. See
[Packaging and distributing a plugin](../creating/packaging-a-plugin.md).

A project's cache keys include the version of every plugin that shapes it, so upgrading a plugin
re-runs its tasks.

## Choosing plugins for a project

| Where | What it sets |
|---|---|
| `smith add <kind> <name> --lang <language>` | the language plugin of a new project |
| `smith add <kind> <name> --stack <stack>` | the whole stack of a new project (`--stack none`: the language plugin alone) |
| `[stacks.<name>]` | declares a stack: `language`, `toolchain`, `structure`, `scaffold`, `deploy` |
| `[default_stacks]` | the stack of a language's projects that record none (`node = "node-jest"`) |
| `[projects.<name>] stack = "..."` | replaces the stack a project's manifest records |
| `[conventions.<name>]` | applies a plugin to every project of the languages it supports; `languages` narrows it, `exclude` leaves projects out |
| `[workspace] recipe = "..."` | creates projects from a recipe |

`smith stacks` lists the stacks, where each comes from, and what is wrong with one that cannot be
used. See [Stacks](../../reference/stacks.md), [Conventions](../../reference/conventions.md) and
[Recipes](../../reference/recipes.md).

## Per-plugin settings

A plugin reads its own settings from the projects it receives (`Project.Settings`, by plugin
name). Which tables of `smith.toml` become those settings depends on the plugin:

| Table | Becomes settings of | Used by |
|---|---|---|
| `[projects.<project>.settings.<plugin>]` | that plugin, for that project | any plugin in the project's stack, or applied as a convention. A recipe writes it when it creates the project |
| `[settings.<plugin>]` in a recipe | the same table, per project made from the recipe | the plugin when it generates the project, and later |
| `[conventions.docker]` and `[projects.<name>.docker]` | the `docker` plugin's settings: `registry`, `registries`, `push`, `verify`, `context`, ... | only `docker`; see its [page](catalogue/docker.md) |

`[conventions.<name>]` accepts a fixed set of keys (`languages`, `exclude` and the docker ones).
Any other key is rejected when smith loads the file:

```text
smith: smith.toml: 71| [conventions.widget-bundle]
72| target = "dist/deploy"
  | ~~~~~~ unknown field
```

So an external plugin is configured per project with `[projects.<project>.settings.<plugin>]`:

```toml
[projects.shop.settings.widget-bundle]
target = "dist/deploy"
```

Which keys a plugin reads is documented on its page: the built-in ones are in the
[catalogue](catalogue/README.md). Of the built-in plugins, only `docker` reads settings today; the
others take their configuration from the project's manifest.

## See also

- [Troubleshooting plugins](troubleshooting-plugins.md)
- [`smith plugin`](../../manual/plugin.md) and its subcommands in the manual
- [Configuration](../../reference/configuration.md): every `smith.toml` table
- [Packaging and distributing a plugin](../creating/packaging-a-plugin.md)
