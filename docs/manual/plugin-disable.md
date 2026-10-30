<!-- lore-master
parent: plugin.md
-->
# smith plugin disable

`smith plugin disable <name>...` turns built-in language plugins off in `smith.toml`. If
`[plugins] enabled` was not written, it is written with every built-in plugin except the
named ones.

## Synopsis

```sh
smith plugin disable <name>...
```

## Arguments and flags

<!-- smith:begin plugin disable -->
Disable built-in plugins in smith.toml

Usage: `smith plugin disable <name>...`

| Argument | Required | Meaning |
|---|---|---|
| `name...` | yes | built-in plugin names |

No flags.

<!-- smith:end plugin disable -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | `smith.toml` was updated. |
| 1 | A name is not a built-in plugin. |

## Files and environment

Rewrites the `plugins.enabled` line of `smith.toml`. Plugins left out of the list are not
loaded (the effect on discovery was not exercised for this page).

## Examples

```sh
smith plugin disable go
```

```text
enabled plugins: bicep, cpp, dart, dotnet, jvm, node, php, python, recipe, ruby, rust
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `no built-in plugin "docker" (built-in: bicep, cpp, ...)` | Only the fifteen language plugins can be disabled. |

## Related

[`smith plugin enable`](plugin-enable.md), [`smith plugin set`](plugin-set.md),
[`smith plugin list`](plugin-list.md).
