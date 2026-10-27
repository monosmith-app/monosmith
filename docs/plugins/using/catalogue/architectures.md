# The architecture plugins

`hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc` and `package-by-layer` are built-in plugins that keep a project's code to an architecture by
checking what each file imports. They are conventions: each is turned on by its own table in
`smith.toml`, and `smith architecture <name>` writes that table for you (and removes the table
of the architecture it replaces). The rules, and what the check can and cannot see in each
language, are in [conventions](../../../reference/conventions.md#architectures-that-check-imports).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Names | `hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc`, `package-by-layer` |
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
It also warns when the chosen architecture does not apply to the language of some of your
projects (it says which), since nothing checks those.

The languages an architecture applies to are what its plugin declares in the handshake (`languages`).
The built-in architectures declare every language in the table above; an architecture plugin
written for fewer languages declares only those, and smith asks it only about projects in them.
[`smith architecture`](../../../manual/architecture.md) shows them under each choice.

## Related

[`smith architecture`](../../../manual/architecture.md), [`smith slices check`](../../../manual/slices-check.md), [vertical-slices](vertical-slices.md).
