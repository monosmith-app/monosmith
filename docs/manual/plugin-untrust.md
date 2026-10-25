<!-- lore-master
parent: plugin.md
-->
# smith plugin untrust

`smith plugin untrust <name>` forgets that you trust an external plugin in this workspace. The
next command that needs the plugin asks again, or is refused where there is no one to ask. A
`sha256` pinned in `smith.toml` is not touched.

## When to use it

When you no longer want a plugin to run: you stopped using it, or you doubt it.

## Synopsis

```sh
smith plugin untrust <name>
```

## Arguments and flags

<!-- smith:begin plugin untrust -->
Forget that you trust an external plugin

Usage: `smith plugin untrust <name>`

| Argument | Required | Meaning |
|---|---|---|
| `name` | yes | the external plugin, as `smith plugin list` names it |

No flags.

<!-- smith:end plugin untrust -->

## Exit codes

0 on success, including when the plugin was not trusted. 1 when the workspace or the smith home
cannot be read.

## Files and environment

Edits `trust.json` in the smith home (`$SMITH_HOME`, or `~/.smith`).

## Examples

```sh
smith plugin untrust widget
```

```text
No longer trusting widget in ~/work/inst.
```

Run again:

```text
widget was not trusted in ~/work/inst.
```

## Related

[`smith plugin trust`](plugin-trust.md), [`smith plugin list`](plugin-list.md),
[Plugin security](../plugins/using/plugin-security.md).
