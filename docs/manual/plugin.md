# smith plugin

`smith plugin` inspects the plugins smith loads and chooses which built-in language plugins
are on. A plugin is a separate process smith starts and talks to over standard input and
output; the built-in ones ship inside the smith binary and are started by it, and
installed ones live on a search path. The protocol is described in
[Plugins](../reference/plugins.md) and
[Writing a plugin](../guides/writing-a-plugin.md).

## When to use it

- To see which plugins exist and what level (language, toolchain, scaffold, structure,
  deploy) and which task slots each fills: [`smith plugin list`](plugin-list.md).
- To switch built-in language plugins on or off for the workspace:
  [`enable`](plugin-enable.md), [`disable`](plugin-disable.md), [`set`](plugin-set.md).
- To verify a plugin you wrote keeps the protocol's promises:
  [`smith plugin check`](plugin-check.md).
- To package a plugin you wrote, or install one someone packaged (an archive, a URL or a GitHub
  release) with its digests checked: [`pack`](plugin-pack.md), [`install`](plugin-install.md).
- To trust an external plugin so smith may start it, or take the trust back:
  [`trust`](plugin-trust.md), [`untrust`](plugin-untrust.md).

## Subcommands

<!-- smith:begin plugin -->
Inspect plugins, or serve a built-in plugin

Usage: `smith plugin <command>`

| Subcommand | Meaning |
|---|---|
| `smith plugin check` | Check that a plugin keeps the protocol's promises |
| `smith plugin disable` | Disable built-in plugins in smith.toml |
| `smith plugin enable` | Enable built-in plugins in smith.toml |
| `smith plugin install` | Install a packaged plugin, checked against its sha256 |
| `smith plugin list` | List the built-in plugins, the external ones the workspace uses (trusted or not) and the ones installed on the plugin search path |
| `smith plugin pack` | Pack a plugin directory into an archive `smith plugin install` accepts |
| `smith plugin set` | Enable exactly these built-in plugins in smith.toml |
| `smith plugin trust` | Review an external plugin and trust it, so smith may start it |
| `smith plugin untrust` | Forget that you trust an external plugin |

No flags.

<!-- smith:end plugin -->

## Where installed plugins are found

In order, the first match for a name winning:

1. `<workspace>/.smith/plugins`
2. each directory of `SMITH_PLUGIN_PATH` (separated like `PATH`)
3. `$SMITH_HOME/plugins`, where `SMITH_HOME` defaults to `~/.smith`

Each plugin is a directory directly under one of them, holding a `plugin.json`. Plugins
declared in `smith.toml` with a command are started from there. A plugin process is started
with `SMITH_PLUGIN=1` in its environment.

## External plugins: trust and environment

A built-in plugin is the smith binary. An external plugin is another program, so smith starts it
only after you have trusted it (`smith plugin trust <name>`; in CI, when `smith.toml` pins its
`sha256`), and gives it a minimal environment plus the variables it declares. The plugins the
workspace needs are the ones `smith.toml` declares under `[plugins.external]`; one that is only
installed on the search path is never started. Trust is not a sandbox: a trusted plugin runs with
your privileges. See [Plugin security](../plugins/using/plugin-security.md).

## Exit codes

0 for the help; see the subcommands.

## Files and environment

`smith.toml` (`[plugins] enabled`, `[plugins.external.<name>]`), the search path above, and
`trust.json` in the smith home, where trust decisions are kept. Variables: `SMITH_PLUGIN_PATH`,
`SMITH_HOME`.

## Related

[Plugins](../reference/plugins.md), [Writing a plugin](../guides/writing-a-plugin.md),
[`smith stacks`](stacks.md), [`smith recipe`](recipe.md).
