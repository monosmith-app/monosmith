<!-- lore-master
parent: config.md
-->
# smith config unset

`smith config unset <setting>` removes a setting from `smith.toml`, so it takes its
default again.

## Synopsis

```sh
smith config unset <setting>
```

## Arguments and flags

<!-- smith:begin config unset -->
Remove one setting, so it takes its default

Usage: `smith config unset <setting>`

| Argument | Required | Meaning |
|---|---|---|
| `setting` | yes | the setting to remove |

No flags.

<!-- smith:end config unset -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The setting was removed. |
| 1 | Unknown setting, or removing it would leave an invalid file (for example `workspace.name`, which is required). |

## Files and environment

Rewrites `smith.toml`.

## Examples

```sh
smith config unset ci.devcontainer
```

```text
✔ ci.devcontainer removed: it has its default
```

Removing the dev container setting means the default (`false`) applies, so the next
`smith ci generate --check` reports `.devcontainer/devcontainer.json is no longer configured`.

## Common errors

| Message | Cause and remedy |
|---|---|
| `refusing to write an invalid smith.toml: smith.toml: [workspace] name is required` | `workspace.name` cannot be removed. |

## Related

[`smith config set`](config-set.md), [`smith config list`](config-list.md).
