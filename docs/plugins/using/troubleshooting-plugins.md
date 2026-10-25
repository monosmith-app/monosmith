# Troubleshooting plugins

What to do when a plugin does not start, does not answer, is not found, clashes with another plugin
or is the wrong version. Every message here is the real one: either captured from a run on
Windows with smith 1.62.0, or, where marked "from the code", read from the source because it is
hard to provoke. For the failures of smith in general, see
[Troubleshooting](../../guides/troubleshooting.md).

## Start with the conformance check

`smith plugin check` starts a plugin the way a workspace would and reports exactly which promise it
breaks. Run it first, on the same plugin you declared:

```sh
smith plugin check widget                       # a name: smith.toml, built-in, then the search path
smith plugin check ./bin/smith-plugin-widget    # a path to an executable
smith plugin check widget --json                # the report as JSON
```

A failed check prints why under the check's name, and the command exits non-zero with
`smith: the plugin does not conform`. The checks are explained in
[Testing a plugin](../creating/testing-a-plugin.md#what-smith-plugin-check-verifies).

## The plugin does not start

All of these come from `initialize`, the first call, which has **30 seconds**. They are printed as
`smith: the plugin did not start and initialize: initializing plugin: initialize: <cause>` by
`smith plugin check`, and prefixed with `plugin <name>:` when a workspace command starts it.

| You see | Cause and remedy |
|---|---|
| `initialize: plugin did not answer within 30s` | the process started and never wrote a reply. It is waiting for input it does not get, stuck in start-up work, or not reading stdin at all. smith kills it. Answer `initialize` before doing anything slow |
| `initialize: plugin stopped answering: EOF` | the process exited before replying. Read the lines printed above the message: a plugin's stderr is shown to you. With a `.cmd` launcher on Windows, a forward-slash path is the usual cause |
| `initialize: plugin stopped answering: malformed header "hello from my plugin"` | something other than a protocol frame reached **stdout**. Only `Content-Length: N`, a blank line and N bytes of JSON may go there. Send logs, banners and `print` output to stderr |
| `starting plugin [.\plugin.py]: fork/exec .\plugin.py: %1 is not a valid Win32 application.` | Windows cannot run a script. Point at a `.cmd` launcher (`@echo off` then `python "%~dp0plugin.py" %*`) or declare it in `smith.toml` with the interpreter as the first element |
| `starting plugin [plugin.cmd]: exec: "plugin.cmd": executable file not found in %PATH%` | a bare file name is looked up on `PATH`. Write `.\plugin.cmd` |
| `starting plugin [.]: exec: ".": executable file not found in %PATH%` | the path is a directory. Point at the executable |
| `plugin chose protocol version 2, host speaks 1` (`plugin <name> chose protocol version <n>, host speaks <n>`, from the code) | the plugin answered `initialize` with a `protocolVersion` smith does not speak. smith lists `[1]` in `protocolVersions`; choose one of them |
| `no common protocol version: host speaks [2], plugin speaks 1` (from the code) | the other way round, from a plugin built on the Go SDK: smith is newer than the SDK in the plugin and lists no version the plugin speaks. Upgrade the plugin |

A plugin can print as much as it likes to stderr: smith shows it, and it is the first place to look.

To see what the plugin says without smith, run it the way smith does, with `SMITH_PLUGIN=1` set,
and type a framed `initialize` (or use `smith plugin check`, which does the same).

## The plugin stops answering later

| You see | Cause and remedy |
|---|---|
| `discover: plugin did not answer within 1m0s` (or `contribute`, `ci`, `doctor`) | the call took longer than its limit. If the plugin advertised `cancel`, smith asks it to stop and gives it five seconds to answer; otherwise, or when it does not answer, smith kills it, because its stream can no longer be trusted (and fails every other call in flight to it). Limits: `initialize` 30 s; `discover`, `contribute`, `ci`, `doctor`, `published` and the native `version/write` 1 min; `generate`, `remove` and the dependency methods 20 min; `publish` 30 min; `shutdown` 5 s; any call made without a limit 2 min. Do slow work in tasks, which smith runs, not in the answers |
| `<method>: cancelled; the plugin did not stop within 5s and was killed` | you pressed Ctrl+C during a call and the plugin did not answer in five seconds. A plugin that advertises `cancel` must stop and answer within that time; one that does not advertise it is killed after the same wait |
| `<method>: plugin stopped answering: EOF` | the plugin crashed or exited in the middle of a call; read its stderr above |
| `<method>: plugin is not accepting requests: ...` | the plugin had already exited when smith wrote the request |
| `<method>: plugin stopped answering: malformed response: ...` (from the code) | the plugin wrote JSON that is not a JSON-RPC response |
| `plugin did not exit after shutdown; killed` | it ignored `shutdown`. Reply `{}` and exit |
| every call fails after the first | something wrote to stdout that was not a frame, after a good start. See above |
| `method not supported: publish` (code `-32601`) | smith called a method the plugin did not implement. smith only calls advertised methods, so the plugin advertises a capability it does not implement (with the Go SDK this cannot happen: capabilities come from the interfaces implemented) |
| `initialize must come first` (code `-32600`) | the plugin served a request before the handshake |

A failure the user must act on is returned by the plugin as error code `-32000` with the remedy in
`message`; smith shows it as the plugin's answer. For example, a deploy plugin that finds nothing to
deploy: `deploying with widget-bundle: no build output at <path>: run smith run build -p api first`.

## The plugin is not found

| You see | Cause and remedy |
|---|---|
| `plugin "nosuch" is not installed: no nosuch/plugin.json in <dirs>; give its command in smith.toml, or install it there` | the name is not on the search path. The message lists the directories looked in: `<workspace>/.smith/plugins`, each directory of `SMITH_PLUGIN_PATH`, `$SMITH_HOME/plugins`. Prefixed with `smith.toml [plugins.external.nosuch]:` when it comes from a declaration |
| `<dir>\plugin.json: name and command are required` | the file lacks a `name` or a `command`. It also stops `smith plugin list`, so fix or remove it |
| `no built-in plugin "typescript" (built-in: bicep, cpp, ...)` | `smith plugin enable` and `disable` change the 12 language plugins only; the add-ons are always loaded |
| `smith.toml enables unknown plugin "nope"; built-in plugins are [bicep cpp ...]` | a hand-written `[plugins] enabled` names a plugin that does not exist |
| `no plugin handles language "go"; available: bicep, cpp, ...` | the language plugin is not enabled, or `--lang` is not one of its languages. `smith plugin list` shows the names, `smith plugin enable go` turns it on |

## Wrong name or wrong version

| You see | Cause and remedy |
|---|---|
| `smith.toml [plugins.external.other]: the plugin calls itself something else; declare it under its own name` | the name in `smith.toml` must be the name the plugin returns in `initialize` (`plugin.name`) |
| `smith.toml [plugins.external.widget] wants version ^2, but the plugin is 0.1.0` | the `version` constraint of the declaration is not met. Install the right version, or change the constraint. Constraints: `1.2.3`, `^1.2`, `~1.2`, `>=1.2`, `*` |
| `smith.toml [plugins.external.widget]: version constraint "x": not a version` / `plugin version "dev": not a version` (from the code) | the constraint or the version the plugin reports is not `major[.minor[.patch]]` (a leading `v` and a `-pre` or `+build` suffix are ignored) |

## Stacks and conventions

A stack is validated against the loaded plugins; a recipe goes through the same rules. Each message
names both sides.

| You see | Cause and remedy |
|---|---|
| `smith.toml [stacks.bad]: jest and vitest both provide the test task; keep one` | two toolchain plugins fill one slot. Keep one |
| `smith.toml [stacks.bad]: jest applies to [node], not to widget projects` | the plugin's `languages` do not include the stack's language |
| `smith.toml [stacks.bad]: eslint is listed as structure, but it plays [toolchain]` | the plugin was listed at a level it does not declare in `roles` |
| `smith.toml [stacks.bad]: plugin "nope" is not loaded` | the plugin is neither built-in nor declared under `[plugins.external]` |
| `smith.toml [stacks.bad]: language plugin "nope" is not loaded` | the stack's `language` is not loaded |
| `smith.toml configures unknown convention "nope"; built-in ones are [angular astro docker ...], and an external one is declared under [plugins.external.nope]` | a `[conventions.nope]` table names a plugin that is neither an add-on nor declared |
| `smith.toml [conventions.x]: x is not a convention plugin` (from the code) | only a plugin of kind `convention` can be applied workspace-wide |
| `smith.toml [conventions.x]: it does not apply to "go"; it supports [...]` (from the code) | `languages` names a language the convention does not support |
| `no stack named "nope"; known stacks: ... (smith stacks lists them)` (from the code) | `--stack`, a manifest or `[default_stacks]` names a stack that does not exist |
| `stack x (plugin y) cannot be used: <problem>` (from the code) | a stack published by a plugin is invalid with the plugins loaded; the problem is one of the rows above |
| `project p is discovered by the node plugin, but stack s is for go projects` (from the code) | a project records a stack of another language |
| `<name> requires a language plugin node, and none is chosen`, `<a> conflicts with <b> (plugin <b>); keep one` (from the code) | a `requires` or `conflicts` of a plugin was not satisfied |

## A convention or a toolchain changes nothing

| Symptom | Cause and remedy |
|---|---|
| a structure plugin adds no task | it names no `languages`: a convention that names none applies to no project, and nothing says so. Set `languages` in its `initialize` answer |
| `[conventions.x]` is configured but no task appears | the plugin advertises no `contribute` capability, or none of its languages match a project. `smith explain <project>:<task>` says which plugin each field came from |
| a toolchain's `generate` ran but its `lint` is missing | a toolchain that only implements `generate` writes scripts and files; the language plugin turns the scripts into tasks. The built-in `eslint`, `jest` and `vitest` work this way (see [the eslint page](catalogue/eslint.md)). To add a task directly, implement `contribute` |
| a `dependsOn` is ignored | `dependsOn` in a contribution applies only to a task the project has after earlier plugins contributed |
| `smith doctor` says `p has no lint task, so CI never checks it` | no plugin in the stack provides one; add a toolchain that does |
| `smith plugin check --generate` fails `generates a lib` for a scaffold that only writes apps | `--generate` asks for an app, a lib and a package and every one must succeed. Run `smith plugin check` without it for an app-only scaffold |

## Language plugin checks that fail

| Check | Cause |
|---|---|
| `a scaffold generates projects` fails | a language plugin that declares no `roles` is also a scaffold by default. Declare `"roles": ["language"]`, or implement `generate` |
| `a language plugin discovers projects` fails | it needs `capabilities.discover` and at least one entry in `manifests` |
| `a toolchain names the slots it fills` fails | a convention-kind plugin with the toolchain role must list `provides` |
| `only a toolchain provides tasks` fails | `provides` is set without the `toolchain` role |
| `stack x is complete` fails | a stack needs a unique `name` and a `language` |
| `the generated lib records its stack` fails | `generate` must write `params.stack` into the manifest, so `discover` reports it back |

## See also

- [Installing, enabling and configuring plugins](managing-plugins.md)
- [What a plugin is and how plugins combine](what-is-a-plugin.md)
- [Testing a plugin](../creating/testing-a-plugin.md)
- [The plugin system](../../architecture/plugin-system.md): lifecycle and failure modes
