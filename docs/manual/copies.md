# smith copies

`smith copies` counts the files that projects keep as exact copies of each other. It lists each file
at the root of three projects or more with the same content, the lines the copies cost, the lines they
would cost if each lived once at the workspace, and the projects that hold the file with other
content. It changes nothing.

## When to use it

- Before you move shared config to a base file (`tsconfig.base.json`, a shared ESLint config).
- After `smith convert`, to see how much of each project is copied config.

## Synopsis

```sh
smith copies [--json]
```

## Arguments and flags

<!-- smith:begin copies -->
Count the files that projects keep as exact copies of each other, and the lines they cost

Usage: `smith copies [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the result as JSON |

<!-- smith:end copies -->

## Example

```text
Copies: 2 files that 3 projects or more keep the same.

File                         Projects  Lines  Different
tsconfig.json                      22      9  -
tsconfig.build.json                22      4  -

These copies cost 286 lines. Kept once they would cost 13 (of 1254 lines of config at project roots).
```

Only files directly in a project's folder count. Blank lines are not counted. A file over 64 KB or
with binary content is skipped. The command does not remove anything: whether a tool still finds a
config after the copy is gone depends on the tool
([ADR 0059](../architecture/decisions/0059-measure-the-copies-before-removing-them.md)).
