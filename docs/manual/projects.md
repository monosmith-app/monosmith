# smith projects

`smith projects` lists the workspace's projects: name, kind, language and folder. With
`--json` it prints the full record of each project: its `root` and `dir`, `kind`,
`language`, `ecosystem`, the `plugin` that serves it, the `stack`, `version`, `deps` (the
projects it imports), and its `tasks` and `targets`.

## When to use it

- To find a project's name for `-p`, `smith explain` or `smith remove`.
- In a script, as the machine-readable list of what smith sees.

## Synopsis

```sh
smith projects [--json]
```

## Arguments and flags

<!-- smith:begin projects -->
List the workspace's projects

Usage: `smith projects [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end projects -->

## Exit codes

0; 1 outside a workspace.

## Files and environment

Reads `smith.toml` and the manifests under the workspace.

## Examples

```sh
smith projects
```

```text
api                  app      go       apps/api
tools                package  python   packages/tools
ui                   lib      node     libs/ui
web                  app      node     apps/web
```

```sh
smith projects --json
```

```json
[
  {
    "name": "api",
    "root": "apps/api",
    "dir": "apps/api",
    "kind": "app",
    "language": "go",
    "ecosystem": "go",
    "plugin": "go",
    "deps": [],
    "tasks": ["build", "lint", "package", "slice-check", "start", "test"],
    "targets": ["build", "lint", "package", "slice-check", "start", "test"]
  }
]
```

(One project of four shown, arrays compacted.)

## Related

[`smith graph`](graph.md), [`smith explain`](explain.md), [`smith add`](add.md).
