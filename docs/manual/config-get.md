<!-- lore-master
parent: config.md
-->
# smith config get

`smith config get <setting>` prints one setting's value as it is written in
`smith.toml`. A setting that is not written has its default and is an error here:
read defaults with [`smith config list`](config-list.md).

## Synopsis

```sh
smith config get <setting>
```

## Arguments and flags

<!-- smith:begin config get -->
Print one setting's value as written in smith.toml

Usage: `smith config get <setting>`

| Argument | Required | Meaning |
|---|---|---|
| `setting` | yes | the setting, such as `ci.providers` |

No flags.

<!-- smith:end config get -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The setting is set; its value was printed. |
| 1 | The setting is not set, or is not a setting `smith config` knows. |

## Files and environment

Reads `smith.toml`.

## Examples

```sh
smith config get ci.providers
```

```text
["github"]
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `release.branch is not set; it has its default (see https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md)` | The key is absent from the file. `smith config list` shows its default. |

## Related

[`smith config list`](config-list.md), [`smith config set`](config-set.md).
