# smith commands

`smith commands` lists smith's own commands with their arguments and options. With
`--json` it prints one JSON document that editors and scripts read; the flag tables of this
manual are generated from it (see
[`docs/reference/cli-generate.mjs`](../reference/cli-generate.mjs)).

## When to use it

- To build a completion, an editor integration or documentation that cannot drift from the
  binary.
- To check which version of smith has which flag.

## Synopsis

```sh
smith commands [--json]
```

## Arguments and flags

<!-- smith:begin commands -->
List smith's commands with their arguments and options (--json for editors)

Usage: `smith commands [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the commands as one JSON document |

<!-- smith:end commands -->

Without `--json` it prints one line per command: the name, an internal group (`projects`,
`workspace`, `pipeline`, `inspect`, `dependencies`), and the description. The JSON is an array with one
object per top-level command: `name`, `aliases`, `group` (an internal grouping, not the
help groups), `description`, `arguments` (`name`, `description`, `required`, `variadic`),
`options` (`name`, `flags`, `description`, `takesValue` which is `none` or `required`,
`variadic`, `repeatable`, `negated`) and `subcommands` (the same shape, nested). Default
values and value types are not in it.

## Exit codes

0.

## Files and environment

None. It does not need a workspace.

## Examples

```sh
smith commands --json
```

prints, for `smith ci`:

```json
{
  "name": "ci",
  "aliases": [],
  "group": "pipeline",
  "description": "Run a CI phase, or regenerate the pipelines",
  "arguments": [],
  "options": [],
  "subcommands": [
    { "name": "generate", "options": [ { "name": "check", "flags": "--check", "takesValue": "none" } ] }
  ]
}
```

(abbreviated: the real document carries every field of every subcommand).

## Related

[`smith kinds`](kinds.md), [`smith schema`](schema.md), [CLI reference](../reference/cli.md).
