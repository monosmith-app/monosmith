# smith why

`smith why <file>` says who owns a file, what produced it, which rule or renderer is behind
it, and how to change it. It answers for a file smith derives, for a file a smith command
generates, for a file inside a project, and for a file smith does not know.

It is not [`smith why-miss`](why-miss.md), which says why a task ran again instead of coming
from the cache. `smith why` is about files. [`smith explain`](explain.md) is about the
settings of a project's tasks.

## A setting

Give `smith why` the path of a setting (`smith why workspace.commit_operations`) and it says the
value and where it comes from (`smith.toml` or the default), what the setting does, why the default
is what it is, and the two commands to change it and to go back. The reason is one line recorded
with the setting in the schema (`x-reason`); `smith config list` shows it after the description, and
`smith config list --json` and `smith why <setting> --json` carry it for editors. A setting with no
recorded reason says so. A name that is also an existing file is treated as the file.

## When to use it

- Before you edit a file you did not write: is it yours, or does smith rewrite it?
- After [`smith derive`](derive.md) reports drift: which renderer wrote the file, and from what?
- When an agent asks whether it may edit a file.

## Synopsis

```sh
smith why <file> [--json]
```

## Arguments and flags

<!-- smith:begin why -->
Say who owns a file, what produced it and how to change it (why-miss says why a task reran)

Usage: `smith why <file> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `file` | yes | the file to explain, relative to the current directory |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end why -->

The path is relative to the current directory and must lie inside the workspace.

## What it says

| Kind | Owner | What `why` names |
|---|---|---|
| derived | smith | the renderer and its version, the mode (`file`, `region` or `keys`), the project and its dependencies, the state from [`smith derive`](derive.md), and the two ways to change it |
| ejected | you | the renderer that wrote the file before you took it over |
| generated | smith | the command that writes it: `smith ci generate`, `smith agent file` or `smith agent install` |
| workspace | you | `smith.toml`: edit it, or use `smith config` |
| project | you | the project, its stack and its choices. A choice may have scaffolded the file once. No record says so, so `why` words it as "if" |
| unknown | nobody | a file that no project contains and no record lists |

To change a derived file, edit its input (a dependency, a fact or `smith.toml`), then run
`smith derive`. To keep your own version, run `smith derive --eject <file>`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | An answer was printed, also for a file smith does not know. |
| 1 | The path is outside the workspace, or `smith.lock` cannot be read. |

## Files and environment

Reads `smith.lock`, `smith.toml`, the project graph and the first lines of the file. With
`[derive] enabled = true` it also builds the derive plan, to show the state of each derived
file. It writes nothing.

## Examples

```sh
smith why packages/web/tsconfig.lib.json
```

```text
packages/web/tsconfig.lib.json
  Summary  smith owns the derived keys of this file. renderer node.tsconfig.references v1 writes it from the facts and the project graph.
  Owner    smith
  Project  web
  Because  the facts and the project graph of web (it depends on core, ui)
  Rule     renderer node.tsconfig.references v1
  Part     node.tsconfig.references v1 (keys, owner smith, current)
  Change   change the input (a dependency, a fact or smith.toml), then run smith derive
           to keep your own version, run smith derive --eject packages/web/tsconfig.lib.json
```

```sh
smith why notes/todo.txt
```

```text
notes/todo.txt
  Summary  smith does not know this file. No project contains it and no record lists it.
  Change   edit it as you like: smith does not read or write it
```

## Related

[`smith derive`](derive.md), [`smith explain`](explain.md), [`smith why-miss`](why-miss.md),
[Derived files](../guides/derived-files.md).
