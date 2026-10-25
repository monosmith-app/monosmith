<!-- lore-master
parent: plugin.md
-->
# smith plugin set

`smith plugin set <name>...` enables exactly the named built-in plugins in `smith.toml` and
turns the others off.

## Synopsis

```sh
smith plugin set <name>...
```

## Arguments and flags

<!-- smith:begin plugin set -->
Enable exactly these built-in plugins in smith.toml

Usage: `smith plugin set <name>...`

| Argument | Required | Meaning |
|---|---|---|
| `name` | yes | built-in plugin names |

No flags.

<!-- smith:end plugin set -->

At least one name is required (`requires at least 1 arg(s), only received 0`).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | `smith.toml` was updated. |
| 1 | No name, or a name that is not a built-in plugin. |

## Files and environment

Rewrites the `plugins.enabled` line of `smith.toml`.

## Examples

```sh
smith plugin set go node python
```

```text
enabled plugins: go, node, python
```

```sh
smith config get plugins.enabled
```

```text
["go", "node", "python"]
```

## Related

[`smith plugin enable`](plugin-enable.md), [`smith plugin disable`](plugin-disable.md),
[`smith config`](config.md).
