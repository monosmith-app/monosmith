# smith schema

`smith schema` prints the JSON Schema of `smith.toml` (draft 07). Editors that read TOML
schemas, such as Even Better TOML in VS Code, pick it up from the `#:schema` line that
`smith new` writes at the top of `smith.toml`; the schema is also served at
`https://raw.githubusercontent.com/monosmith-app/monosmith/main/apps/smith/workspace/smith.schema.json`.

The schema says `additionalProperties: false` at the top and in each table, so a key smith
does not know is an error, and `[workspace] name` is the only required key.

## When to use it

- To validate `smith.toml` in a tool that is not an editor plugin.
- To read the type and default of a setting; [`smith config list`](config-list.md) shows the
  same descriptions with the current values.

## Synopsis

```sh
smith schema
```

## Arguments and flags

<!-- smith:begin schema -->
Print the JSON Schema of smith.toml, for editors

Usage: `smith schema`

No flags.

<!-- smith:end schema -->

## Exit codes

0.

## Files and environment

None; it needs no workspace.

## Examples

```sh
smith schema > smith.schema.json
```

The beginning of the output:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://raw.githubusercontent.com/monosmith-app/monosmith/main/apps/smith/workspace/smith.schema.json",
  "title": "smith.toml",
  "description": "The configuration of a smith workspace (schema 1). Every setting has a default; a key smith does not know is an error.",
  "type": "object",
  "additionalProperties": false,
  "required": ["workspace"],
  "properties": {
    "schema": { "type": "integer", "const": 1, "description": "The smith.toml schema version; smith upgrade migrates older files." },
    ...
```

## Related

[`smith config`](config.md), [Configuration](../reference/configuration.md), [Editors](../guides/editors.md).
