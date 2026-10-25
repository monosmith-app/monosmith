<!-- lore-master
parent: plugin.md
-->
# smith plugin trust

`smith plugin trust <name>` reviews an external plugin the workspace uses and, once you say yes,
records that you trust it, so smith may start it. smith starts an external plugin only when it is
trusted, and a plugin is a program that runs with your privileges: read what it is before you
answer. `--pin` also writes the plugin's digest to `smith.toml`, which is what CI checks.

The decision is kept in your smith home (`$SMITH_HOME`, or `~/.smith`, in `trust.json`), never in
the repository, and is for this workspace, this plugin name, this command and these files. If the
command or a file changes, smith asks again.

## When to use it

- A command was refused with `the external plugin "x" is not trusted yet`.
- You updated a plugin on purpose, and smith says it `changed since you trusted it`.
- CI says the plugin `is not pinned`: run it with `--pin` on your machine and commit the result.

## Synopsis

```sh
smith plugin trust <name> [--yes] [--pin]
```

## Arguments and flags

<!-- smith:begin plugin trust -->
Review an external plugin and trust it, so smith may start it

Usage: `smith plugin trust <name> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `name` | yes | the external plugin, as `smith.toml` or `smith plugin list` names it |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--pin` | boolean | off | also write the plugin's sha256 to smith.toml, for CI |
| `-y, --yes` | boolean | off | trust without asking |

<!-- smith:end plugin trust -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The plugin is trusted (now, or already). |
| 1 | You declined, there was no terminal to ask in and no `--yes`, the plugin could not be found, or it is a built-in plugin. |

## Files and environment

Reads `smith.toml` and the plugin search path; writes `trust.json` in the smith home. With
`--pin`, changes only the `sha256` line of `[plugins.external.<name>]` in `smith.toml`, so comments
and every other setting stay.

## Examples

On your machine, in a terminal:

```sh
smith plugin trust widget
```

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

Trust it and pin it for CI, without asking:

```sh
smith plugin trust widget --yes --pin
```

```text
Trusted widget in ~/work/inst.
Pinned sha256 f6967352942e34f139256eafee1da615931c1f0e0888ddecead3d9a38c15a239 in smith.toml [plugins.external.widget], which is what CI checks.
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ refusing to trust a plugin without a terminal to ask in` | No one can answer. Read the plugin with `smith plugin list`, then run `smith plugin trust <name> --yes`. |
| `✖ <name> is a built-in plugin: it is the smith binary itself, so there is nothing to trust` | Only external plugins are trusted. |
| `✖ plugin "x" is not installed: no x/plugin.json in <dirs>; give its command in smith.toml, or install it there` | smith cannot find the plugin; check the name and `smith plugin list`. |
| `<name> was not trusted` | You answered no. |

## Related

[`smith plugin untrust`](plugin-untrust.md), [`smith plugin list`](plugin-list.md),
[`smith doctor`](doctor.md), [Plugin security](../plugins/using/plugin-security.md),
[Installing, enabling and configuring plugins](../plugins/using/managing-plugins.md#trusting-an-external-plugin).
