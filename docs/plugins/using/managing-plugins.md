# Installing, enabling and configuring plugins

This page is every way to change which plugins a workspace uses and how they are set up: the
`smith plugin` subcommands, the `[plugins]` tables of `smith.toml`, installing an external plugin,
and the settings a plugin reads. The model behind it is in
[What a plugin is and how plugins combine](what-is-a-plugin.md).

## The `smith plugin` commands

| Command | Does |
|---|---|
| `smith plugin list [--json]` | lists the built-in plugins, the external plugins the workspace uses (trusted or not, and what each may see) and the plugins installed on the search path |
| `smith plugin check <name or path> [--generate] [--env NAME] [--json]` | runs the conformance suite against a plugin, which needs no trust |
| `smith plugin search [term]` | lists the plugins of the [marketplace](marketplace.md) that match the term |
| `smith plugin install <name[@version]|path|url|github:owner/repo@tag>` | installs a packaged plugin after checking it against its sha256 (a name is looked up in the marketplace); `--pin` also pins the digest in `smith.toml` |
| `smith plugin pack [dir]` | packs a plugin directory you wrote into an archive for `install` |
| `smith plugin trust <name> [--yes] [--pin]` | reviews an external plugin and records that you trust it, so smith may start it |
| `smith plugin untrust <name>` | forgets that you trust an external plugin |
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
tauri            toolchain                    tauri-build          language plugin node
typescript       scaffold,toolchain           build,typecheck      language plugin node
vertical-slices  structure                    -                    -
vite             scaffold,toolchain           build,typecheck      language plugin node
vitest           toolchain                    test                 language plugin node
vscode           scaffold,toolchain,deploy    build,typecheck      language plugin node

External plugins this workspace uses (smith starts them only when trusted):
  PLUGIN         FROM               TRUSTED  PINNED  MAY SEE BEYOND THE BASELINE
  widget         .smith/plugins     no       no      GITHUB_TOKEN, NPM_TOKEN

Installed (search path: ~/work/inst/.smith/plugins, ~/.smith/plugins):
  widget         0.1.0      ~/work/inst/.smith/plugins/widget
```

| Column | Means |
|---|---|
| `PLUGIN` | the plugin's name, as it is written in `smith.toml` and in stacks |
| `LEVELS` | the roles it can be chosen for. A language plugin that declares none shows the default: language, toolchain and scaffold |
| `SLOTS` | the tasks a toolchain owns; two plugins of one project may not share a slot |
| `REQUIRES` | plugins that must be in the same project (`language plugin node`); `not <ref>` marks a conflict |

The external plugins section lists the plugins `smith.toml` declares under `[plugins.external]`,
which are the ones the workspace needs:

| Column | Means |
|---|---|
| `FROM` | where the plugin comes from: `smith.toml` (its `command`), `.smith/plugins`, `SMITH_PLUGIN_PATH` or `$SMITH_HOME/plugins` |
| `TRUSTED` | whether smith would start it now. On your machine, whether you have trusted this command and file in this workspace; in CI, whether `smith.toml` pins the file's digest and the file matches |
| `PINNED` | whether `smith.toml` gives the plugin a `sha256` |
| `MAY SEE BEYOND THE BASELINE` | the environment variables it declares and you granted, by name (never value); `-` when it gets the baseline only |

A plugin installed on the search path but not declared in `smith.toml` is not used by the workspace,
is never started, and needs no trust. The capabilities, which `--json` adds, are not shown for it:

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

(abridged: the real output is indented one value per line). `source` is `built-in`, or where an
installed plugin was found; installed plugins show `name`, `version`, `source` and the `env` their
manifest declares, since listing does not start them. An external plugin the workspace uses also
has `"used": true`, its `command`, the `env` it may see, `trusted` and `pinned`. A `plugin.json` without a `name` or `command` makes the
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
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ discover answers with no manifests
  ✔ ci answers with no projects
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

The first argument is a **path** when a file by that name exists, otherwise a **name**, looked up
as a workspace would: `smith.toml`, then the built-in plugins, then the search path. Checking is for
the person writing the plugin, so it does not need the plugin to be trusted. It does use the
[minimal environment](plugin-security.md#the-environment-an-external-plugin-gets) of an external
plugin: `--env NAME` (repeatable; `--env 'AWS_*'` for a prefix) grants a variable for the run, and a
plugin found by name also gets the variables its `plugin.json` and `smith.toml` declare. On Windows a
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
| `env` | the environment variables the plugin may see beyond the baseline: names, or prefixes ending in `*` (`["GITHUB_TOKEN", "AWS_*"]`). Adds to what the plugin's own `plugin.json` declares |
| `sha256` | the SHA-256 of the plugin's file (the script, when the command runs an interpreter on one). In CI the plugin starts only when this matches. `smith plugin trust <name> --pin` writes it |
| `version` | the version the plugin must report: `"1.2.3"` (exactly), `"^1.2"` (same major, at least 1.2), `"~1.2"` (same major and minor, at least 1.2), `">=1.2"`; empty or `"*"` is any |

smith checks that the plugin answers to the name it is declared under, and that its version meets
the constraint, as soon as it starts it:

```text
smith: smith.toml [plugins.external.widget] wants version ^2, but the plugin is 0.1.0
smith: smith.toml [plugins.external.other]: the plugin calls itself something else; declare it under its own name
```

A declared plugin with the name of a built-in plugin replaces the built-in one. An external plugin
starts only once you [trust it](#trusting-an-external-plugin), and with a minimal environment: see
[Plugin security](plugin-security.md).

### Install a packaged plugin

A plugin that someone packed with [`smith plugin pack`](../../manual/plugin-pack.md) installs from
the archive, from a URL or from a GitHub release:

```sh
smith plugin install ./widget-0.1.0.smith-plugin.zip --sha256 db7fb4c0...2204d --pin
smith plugin install https://example.com/widget-0.1.0.smith-plugin.zip --sha256 db7fb4c0...2204d --pin
smith plugin install github:acme/smith-plugin-widget@v0.1.0 --pin
```

```text
Installed widget 0.1.0 in ~/work/inst/.smith/plugins/widget (2 files, from ./widget-0.1.0.smith-plugin.zip).
  archive sha256 db7fb4c0...2204d matches the digest you gave, or the marketplace or release published.
  every file matches checksums.sha256.
  the plugin was not started.
Pinned sha256 9e36a074...e7087 in smith.toml [plugins.external.widget], which is what CI checks.
smith will ask before it first starts widget (or run `smith plugin trust widget`).
```

What to rely on:

- Get the digest from the author through a channel other than the download (their release notes, the
  `.sha256` file on a GitHub release, a message) and pass it with `--sha256`. A digest that does not
  match stops the install and writes nothing, and an archive with no digest to be checked against is
  refused. The digest proves the archive is the one the author published, not who the author is.
- smith checks every file against the archive's `checksums.sha256`, refuses paths that leave the
  plugin's folder, links and special files, and stops at 512 MiB per download and per file.
- Installing **never starts the plugin**, and does not trust it either: smith asks the first time
  the plugin is used, unless you pass `--trust`. In a workspace the plugin goes to `.smith/plugins/<name>`,
  and with `--pin` its digest is also pinned in `smith.toml` (CI needs the pin); with `--global`, or
  outside a workspace, it goes to `$SMITH_HOME/plugins`, where `--pin` is an error. An installed plugin
  is replaced only with `--force`.

Every flag, check and message is in the [`smith plugin install`](../../manual/plugin-install.md) page.

### Install it by hand on the search path

An installed plugin is a directory holding a `plugin.json`:

```json
{
  "name": "widget",
  "version": "0.1.0",
  "description": "projects described by a widget.json",
  "command": ["./smith-plugin-widget{exe}"],
  "env": ["NPM_TOKEN"]
}
```

| Field | Meaning |
|---|---|
| `name`, `command` | required |
| `version`, `description` | shown by `smith plugin list` |
| `env` | the environment variables the plugin needs beyond the baseline every external plugin gets: names, or prefixes ending in `*`. Shown by `smith plugin list` and in the prompt that asks you to trust the plugin. A manifest cannot ask for the variables smith uses for its own cache and releases (`SMITH_CACHE_KEY` and the like); only `smith.toml` can grant those |

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
  ✔ declares a concurrency the protocol defines
  ✔ names itself
  ...
  ✔ shuts down when asked
```

Then name it, with or without a constraint, and use it:

```toml
[plugins.external.widget]
version = "^0.1"
```

The first command that needs the plugin asks you to trust it (see
[Trusting an external plugin](#trusting-an-external-plugin)); after that:

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

Installing a plugin by name from a registry is not decided yet
([#164](https://github.com/monosmith-app/monosmith/issues/164)): today a plugin arrives by
[`smith plugin install`](../../manual/plugin-install.md) from an archive, a URL or a GitHub release, or
by any other way that puts a directory with a `plugin.json` on the search path. See
[Packaging and distributing a plugin](../creating/packaging-a-plugin.md).

A project's cache keys include the version of every plugin that shapes it, so upgrading a plugin
re-runs its tasks.

## Trusting an external plugin

smith starts an external plugin only after you have trusted it. The first time a command in a
terminal needs one, smith shows it and asks:

```text
The external plugin "widget" wants to run on this machine.
  from      .smith/plugins (~/work/inst/.smith/plugins/widget)
  command   ~/work/inst/.smith/plugins/widget/smith-plugin-widget.exe
  runs      ~/work/inst/.smith/plugins/widget/smith-plugin-widget.exe
            sha256 4332d20df523297dd2a70e1956740782b439aeee8a8cef5f604f08f741efd8f5
  sees      the baseline environment plus GITHUB_TOKEN, NPM_TOKEN
A trusted plugin runs with your privileges and is not sandboxed. Read its code before you trust it.
Trust "widget"? [y/N]
```

Answer `y` and the command carries on. The decision is kept in your smith home (`$SMITH_HOME`, or
`~/.smith`, in `trust.json`), never in the repository, and it is for **this workspace, this plugin
name, this command and these files**. If the command changes, or the file it runs changes (you
updated the plugin, or someone replaced it), smith asks again. For an interpreter, such as
`node plugin.mjs`, the interpreter and the script are both covered.

Where there is nobody to ask (`smith serve`, the shell, `smith mcp`, an editor, a script, or a
command with its input redirected), smith never asks; it refuses and says what to run:

```sh
smith projects
```

```text
✖ the external plugin "widget" is not trusted yet, so smith did not start it
  → run `smith plugin trust widget` to review it and trust it
  → run `smith plugin list` to see every external plugin and what it may see
```

| Command | Does |
|---|---|
| `smith plugin trust <name>` | shows the plugin as above, asks, and records the decision |
| `smith plugin trust <name> --yes` | records it without asking (for a script; read the plugin first) |
| `smith plugin trust <name> --pin` | also writes the plugin's digest to `smith.toml`, for CI |
| `smith plugin untrust <name>` | forgets the decision; smith asks again next time |

```sh
smith plugin trust widget --yes --pin
```

```text
Trusted widget in ~/work/inst.
Pinned sha256 f6967352942e34f139256eafee1da615931c1f0e0888ddecead3d9a38c15a239 in smith.toml [plugins.external.widget], which is what CI checks.
```

```toml
[plugins.external.widget]
version = "^0.1"
env = ["GITHUB_TOKEN"]
sha256 = "f6967352942e34f139256eafee1da615931c1f0e0888ddecead3d9a38c15a239"
```

### In CI

In CI (the `CI` variable, or `GITHUB_ACTIONS` or `TF_BUILD`, is set) nobody is there to ask, and your
decision on your own machine does not exist. An external plugin starts only when `smith.toml` pins
its `sha256` and the file still has that digest. On your machine a pin on its own does not make a
plugin trusted. An unpinned plugin stops the job with the remedy:

```text
✖ the external plugin "widget" is not pinned, so smith did not start it: in CI an external plugin runs only when smith.toml pins its sha256
  → run `smith plugin trust widget --pin` on your machine and commit the sha256 it writes to smith.toml
```

When you update a plugin on purpose, run `smith plugin trust <name> --pin` again and commit the new
`sha256`. [`smith doctor`](../../manual/doctor.md) lists every external plugin with its source,
whether it is trusted and pinned and what it may see, and fails on one smith would refuse to start.

### What a plugin gets to see

An external plugin starts with a minimal environment: the baseline (`PATH`, `HOME`, the temporary
directory, locale, terminal and proxy variables, and on Windows the variables a process needs to
run) plus the variables it declares in its `plugin.json` and the ones you grant it with `env` in
`smith.toml`. Your other variables, such as `NPM_TOKEN`, cloud credentials and smith's own cache key,
do not reach it unless named. The built-in plugins are part of smith and keep the whole environment.
The full list and the limits of this protection are on [Plugin security](plugin-security.md).

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
- [Plugin security: what smith protects and what it does not](plugin-security.md)
- [`smith plugin`](../../manual/plugin.md) and its subcommands in the manual
- [Configuration](../../reference/configuration.md): every `smith.toml` table
- [Packaging and distributing a plugin](../creating/packaging-a-plugin.md)
