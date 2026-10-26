# The architecture plugins

`hexagonal`, `clean`, `ddd` and `modular-monolith` are built-in plugins that keep a project's code to an architecture by
checking what each file imports. They are conventions: each is turned on by its own table in
`smith.toml`, and `smith architecture <name>` writes that table for you (and removes the table
of the architecture it replaces). The rules, and what the check can and cannot see in each
language, are in [conventions](../../../reference/conventions.md#architectures-that-check-imports).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Names | `hexagonal`, `clean`, `ddd`, `modular-monolith` |
| Applies to | Go, Python, Node, JVM, .NET, PHP, Dart, Rust, Ruby and C++ projects |
| Capabilities | `contribute`, `doctor` |
| Stacks they publish | none; turn one on with `[conventions.hexagonal]`, `[conventions.clean]` and so on |
| Project kinds | any |

## What it does

For every project of a language it reads, the plugin contributes an `architecture-check` task
and makes the project's `lint` (when it has one) wait for it:

```text
tool:architecture-check — checks the hexagonal architecture: the core never imports an adapter; one adapter never imports another
  command    smith slices check --architecture hexagonal --lang go .
  inputs     **/*.go
```

The task is cached on the project's source files, like `slice-check`. `smith doctor` adds a
warning for a project that has no folder named for a layer, since the rules then check nothing.

## Related

[`smith architecture`](../../../manual/architecture.md), [`smith slices check`](../../../manual/slices-check.md), [vertical-slices](vertical-slices.md).
