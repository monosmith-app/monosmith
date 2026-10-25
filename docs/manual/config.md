# smith config

`smith config` reads and changes the settings of `smith.toml`'s `[workspace]`, `[ci]`,
`[release]`, `[cache]` and `[plugins]` tables, and the per-language default stack
(`default_stacks.node`). Values are checked against the file's JSON Schema, and only the
setting's value changes: comments and every other line stay as they are. Stacks,
conventions, tasks and projects are edited in the file itself.

## When to use it

- You want to see every setting, its current value and what it does (`config list`).
- You want to change one setting without opening the file and with the value checked.

Do not use it for `[stacks.*]`, `[conventions.*]`, `[tasks.*]`, `[projects.*]` or
`[dependencies]` allow lists: edit `smith.toml`, with the schema in your editor
([`smith schema`](schema.md)). The full list of keys is in
[Configuration](../reference/configuration.md).

## Synopsis

```sh
smith config list [--json]
smith config get <setting>
smith config set <setting> <value>...
smith config unset <setting>
```

## Subcommands

<!-- smith:begin config -->
Read and change settings of smith.toml

Usage: `smith config <command>`

| Subcommand | Meaning |
|---|---|
| `smith config get` | Print one setting's value as written in smith.toml |
| `smith config list` | List the settings, their values and what they do |
| `smith config set` | Change one setting (a list takes several values) |
| `smith config unset` | Remove one setting, so it takes its default |

No flags.

<!-- smith:end config -->

Each subcommand has a page: [`smith config list`](config-list.md),
[`smith config get`](config-get.md), [`smith config set`](config-set.md),
[`smith config unset`](config-unset.md).

A setting is named by its table and key (`ci.providers`, `release.branch`,
`cache.max_size_mb`). A default stack is named `default_stacks.<language>`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The command succeeded. |
| 1 | The setting does not exist or is not set (`get`), the value fails the schema, or the edit would leave an invalid `smith.toml`. |

## Files and environment

Reads and, for `set` and `unset`, rewrites `smith.toml` in place. `SMITH_WORKSPACE_ROOT`
selects the workspace (see [the manual's conventions](README.md#conventions-every-command-shares)).

## Examples

```sh
smith config set ci.providers github azure
```

```text
✔ ci.providers = ["github", "azure"]
```

Run [`smith ci generate`](ci-generate.md) afterwards to write the pipelines the new
provider list asks for.

## Related

[`smith schema`](schema.md), [`smith upgrade`](upgrade.md), [`smith info`](info.md),
[Configuration](../reference/configuration.md).
