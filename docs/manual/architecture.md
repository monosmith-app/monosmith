# smith architecture

`smith architecture` shows how the workspace's code is laid out, and changes it. An
architecture is a set of rules about where code goes and what files are called; smith checks
the code against the one you choose. These are known today:

| Name | What it asks of the code |
|---|---|
| `vertical-slices` | code grouped by what it does, one flat folder per outcome, files named for their role |
| `hexagonal` | ports and adapters: the core knows nothing of databases or the web, and adapters never import each other |
| `clean` | the dependency rule: frameworks, interface adapters, use cases, entities, and code depends only inward |
| `ddd` | domain-driven design: a folder per bounded context with its own domain, application and infrastructure |
| `modular-monolith` | one program split into modules, each used by the others only through its public API |
| `mvc` | models, views and controllers: the model knows neither, and a view never calls a controller |
| `package-by-layer` | one folder per technical layer (controller, service, repository, model), each calling only the ones below |
| `none` | smith checks no code layout |

The rules of
each, and what they can and cannot check in each language, are in
[conventions](../reference/conventions.md#architectures-that-check-imports).

## Which projects an architecture checks

An architecture applies to the languages its plugin declares, and checks only the projects
written in them. Every built-in one applies to all the languages smith reads imports for (Go,
Python, Node, the JVM, .NET, PHP, Dart, Rust, Ruby and C++); an architecture from a plugin may
name fewer. Run without a name, the command prints, under each choice, `applies to:` its
languages and `here:` which of your projects it would check and which it would skip because
their language is not covered. Choosing an architecture that would check none of your projects
prints a warning first (the Settings tab asks you to press Enter a second time), and
[`smith doctor`](doctor.md) keeps warning while a chosen architecture leaves projects out. If
leaving a project out is on purpose, name it in `exclude = [...]` under the architecture's table
in `smith.toml` and the warning stops.

## When to use it

- To see which architecture the workspace follows: run it without a name.
- To switch the checks on or off without editing `smith.toml` by hand.

The Settings tab of [`smith tui`](tui.md) has the same choice in its Architecture section.

## Synopsis

```sh
smith architecture [name]
```

## Arguments and flags

<!-- smith:begin architecture -->
Show the architecture of the workspace, or choose another

Usage: `smith architecture [name]`

| Argument | Required | Meaning |
|---|---|---|
| `name` | no | vertical-slices, hexagonal, clean, ddd, modular-monolith, mvc, package-by-layer or none; without it, the choices are listed and the one in use marked |

No flags.

<!-- smith:end architecture -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Listed, or the file was changed. |
| 1 | Outside a workspace, a name that is not an architecture, or an invalid resulting file. |

A choice that would check none of your projects prints `warning: ...` and is applied anyway.

## Files and environment

Rewrites `smith.toml`, changing only the lines that turn an architecture on: choosing
`vertical-slices` adds `[conventions.vertical-slices]` at the end (`hexagonal` adds
`[conventions.hexagonal]`, and so on), and choosing another one
removes that table together with the settings under it (`languages`, `exclude`). Comments and
every other line are kept. Choosing the architecture already in use changes nothing. See
[conventions](../reference/conventions.md) for what each one checks.

## Examples

```sh
smith architecture
```

```text
* vertical-slices  code grouped by what it does: one flat folder per outcome, files named for their role
                   applies to: cpp, dart, dotnet, go, jvm, node, php, python, ruby, rust
                   here: checks api (go), web (node); skips infra (bicep) (its language is not covered)
  hexagonal        ports and adapters: the core knows nothing of databases or the web, and adapters never import each other
                   applies to: cpp, dart, dotnet, go, jvm, node, php, python, ruby, rust
                   here: checks api (go), web (node); skips infra (bicep) (its language is not covered)
  (the other choices follow the same way)
  none             smith checks no code layout
```

```sh
smith architecture none
```

```text
architecture: none
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `no architecture called "onion"; the choices are vertical-slices, hexagonal, clean, ddd, modular-monolith, mvc, package-by-layer, none` | Use one of the listed names. |

## Related

[`smith config`](config.md), [`smith slices check`](slices-check.md), [conventions](../reference/conventions.md).
