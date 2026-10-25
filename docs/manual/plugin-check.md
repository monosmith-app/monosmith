<!-- lore-master
parent: plugin.md
-->
# smith plugin check

`smith plugin check` runs the conformance suite against a plugin: it names itself, its roles
match what it implements, its stacks are complete, and every method it advertises answers.
It also checks how the plugin wants to be called: requests sent together are answered (one at a
time unless the plugin declares a `concurrency`), and a plugin that advertises `cancel` stops a
cancelled request within a bound and goes on serving.
With `--generate`, a plugin that generates projects makes an app, a lib and a package in a
temporary directory, which a language plugin must then discover as generated (this may
install dependencies).

A name is looked up the way the workspace would start it: `smith.toml`, then the built-in
plugins, then the plugin search path. A path runs that executable. Further arguments after
the plugin are passed to it.

## When to use it

While writing a plugin, and in the plugin's own tests. The `go-smith-plugin` stack scaffolds a
conformance test that does the same.

## Synopsis

```sh
smith plugin check <name | path to plugin> [args...] [--generate] [--json]
```

## Arguments and flags

<!-- smith:begin plugin check -->
Check that a plugin keeps the protocol's promises

Usage: `smith plugin check <name | path to plugin> [args...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `name \| path to plugin` | yes | a plugin by name, or the path of its executable (further arguments are passed to it) |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--generate` | boolean | off | also generate projects in a temporary directory and discover them |
| `--json` | boolean | off | print JSON |

<!-- smith:end plugin check -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every check passed. |
| 1 | A check failed, or the plugin could not be found or started. |

## Files and environment

Starts the plugin process with `SMITH_PLUGIN=1`. With `--generate`, writes into a temporary
directory and may run the plugin's package manager.

## Examples

A built-in plugin:

```sh
smith plugin check node
```

```text
node 1.62.0
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

The Node.js example plugin from the repository, by path:

```sh
smith plugin check docs/examples/plugins/node/plugin.cmd
```

```text
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `plugin "x" is not installed: no x/plugin.json in <dirs>; give its command in smith.toml, or install it there` | Install the plugin in one of the listed directories, or point `smith.toml` at it. |
| `requires at least 1 arg(s), only received 0` | Name a plugin or a path. |

## Related

[`smith plugin list`](plugin-list.md), [Writing a plugin](../guides/writing-a-plugin.md),
[Plugins](../reference/plugins.md).
