<!-- lore-master
parent: plugin.md
-->
# smith plugin enable

`smith plugin enable <name>...` turns built-in language plugins on in `smith.toml`
(`[plugins] enabled`). Until `enabled` is written, all built-in plugins are on; once it is
written, only those listed are loaded.

## Synopsis

```sh
smith plugin enable <name>...
```

## Arguments and flags

<!-- smith:begin plugin enable -->
Enable built-in plugins in smith.toml

Usage: `smith plugin enable <name>...`

| Argument | Required | Meaning |
|---|---|---|
| `name...` | yes | built-in plugin names |

No flags.

<!-- smith:end plugin enable -->

Names are the built-in language plugins: bicep, cpp, dart, dotnet, go, jvm, node, php, python,
recipe, ruby, rust.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | `smith.toml` was updated. |
| 1 | A name is not a built-in plugin. |

## Files and environment

Rewrites the `plugins.enabled` line of `smith.toml`.

## Examples

After `smith plugin disable go`, `smith plugin enable go` puts it back:

```sh
smith plugin enable go
```

```text
enabled plugins: bicep, cpp, dart, dotnet, go, jvm, node, php, python, recipe, ruby, rust
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `no built-in plugin "nonexistent" (built-in: bicep, cpp, dart, dotnet, go, jvm, node, php, python, recipe, ruby, rust)` | Pick from the list. Toolchain plugins such as `docker` are not switched this way. |

## Related

[`smith plugin disable`](plugin-disable.md), [`smith plugin set`](plugin-set.md),
[`smith config unset plugins.enabled`](config-unset.md) (back to all of them).
