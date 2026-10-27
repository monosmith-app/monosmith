<!-- lore-master
parent: config.md
-->
# smith config set

`smith config set <setting> <value>...` changes one setting. A list setting takes several
values; any other takes exactly one. The value is checked against the schema before
anything is written, and the file is not touched if the result would be invalid.

## Synopsis

```sh
smith config set <setting> <value>...
```

## Arguments and flags

<!-- smith:begin config set -->
Change one setting (a list takes several values, or one comma-separated)

Usage: `smith config set <setting> <value...>`

| Argument | Required | Meaning |
|---|---|---|
| `setting` | yes | the setting, such as `ci.providers` |
| `value...` | yes | the new value; a list setting takes several |

No flags.

<!-- smith:end config set -->

Values for booleans are `true` and `false`; integers are plain numbers; lists are the
values separated by spaces (`smith config set ci.providers github azure`).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The file was changed. |
| 1 | Unknown setting, a value the schema rejects, a wrong number of values, or an invalid resulting file. |

## Files and environment

Rewrites `smith.toml`, changing only that setting's line (or adding it to its table).
Comments and everything else are kept.

## Examples

```sh
smith config set ci.devcontainer false
```

```text
✔ ci.devcontainer = false
```

```sh
smith config set ci.providers github azure
```

```text
✔ ci.providers = ["github", "azure"]
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `nonsense is not a setting smith config edits Run smith config list to see them (stacks, tasks and projects are edited in smith.toml).` | Check the name against `smith config list`. |
| `ci.providers takes github or azure or gitlab or circleci, not "gitlab2"` | Use one of the listed values. |
| `release.branch takes one value` | A single-value setting got none or several. |

## Related

[`smith config unset`](config-unset.md), [`smith ci generate`](ci-generate.md) (after changing
`ci.*`), [`smith upgrade`](upgrade.md).
