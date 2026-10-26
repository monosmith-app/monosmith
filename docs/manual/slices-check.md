<!-- lore-master
parent: slices.md
-->
# smith slices check

`smith slices check <project directory>...` checks Go, Python, Dart, Rust, Ruby, C++, .NET,
JVM or PHP projects against the vertical-slice rules:

- a file is named `<name>_<role>` (Go and the other snake_case languages), where the role is
  one of `handler`, `use_case`, `algorithm`, `policy`, `model`, `contract`, `mapper`,
  `validator`, `repository`, `client`, `store`, `middleware`, `error`, `config`, `enum`;
- a module's or package's root holds only its entry files;
- no folder is named `util`, `helper`, `common`, `shared`, `manager`, `processor`, `data` or
  `misc`;
- a slice with more than 12 production files is reported for review.

The vertical-slices convention (`[conventions.vertical-slices]` in `smith.toml`) runs this
as each project's `slice-check` task, so [`smith ci verify`](ci.md#verify) catches a violation.

With `--architecture <name>` the command checks another [architecture](architecture.md)
instead: it reads the imports of the project's sources and reports the ones the architecture
forbids, for Go, Python, Node, JVM, .NET, PHP, Dart, Rust, Ruby and C++ projects. The rules
and their limits are in [conventions](../reference/conventions.md#architectures-that-check-imports);
the convention of the architecture runs it as each project's `architecture-check` task.

```sh
smith slices check --architecture hexagonal --lang go apps/api
```

```text
domain/order.go - imports example.com/shop/adapters/sql: core code must not depend on adapters code (the core never imports an adapter; one adapter never imports another)
smith: 1 import(s) break the hexagonal architecture; an exception is written in the project README
```

## When to use it

Directly, to check a folder that is not yet a project, or to see violations without the
rest of `verify`. Normally use `smith run slice-check`.

## Synopsis

```sh
smith slices check <project directory>... [--lang <language>]
```

## Arguments and flags

<!-- smith:begin slices check -->
Check that files are role-named and no folder is a junk drawer

Usage: `smith slices check <project directory>... [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project directory...` | yes | directories of projects of one language |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--lang <lang>` | language name | `go` | language of the projects: cpp, dart, dotnet, go, jvm, php, python, ruby, rust; with --architecture also node |
| `--architecture <name>` | architecture name | none | check this architecture's import rules instead: hexagonal, clean, ddd, modular-monolith |

<!-- smith:end slices check -->

`--lang` names the language of **all** the directories given; the default is `go`. Pass the
language of the project, or the file-name rule of the wrong language is applied. The
`slice-check` task a plugin defines for a project passes it for you
(`smith slices check --lang python .`).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | `vertical-slice layout OK`. |
| 1 | Violations were found, or a directory cannot be read. |

## Files and environment

Reads the file tree under each directory. Writes nothing. An exception to a rule is written
in the project's README (see [Conventions](../reference/conventions.md)).

## Examples

```sh
smith slices check apps/api
```

```text
vertical-slice layout OK
```

With a badly named file and a junk-drawer folder:

```text
greeting/helpers.go - a Go file is <snake>_<role>.go (roles: handler, use_case, algorithm, policy, model, contract, mapper, validator, repository, client, store, middleware, error, config, enum)
util/ - "util" is a junk drawer; name the folder for the outcome it serves
smith: 2 file(s) or folder(s) break the vertical-slice layout; an exception is written in the project README
```

The exit code is 1.

## Common errors

| Message | Cause and remedy |
|---|---|
| `open apps\nope: The system cannot find the file specified.` | The directory does not exist. |
| `<file> - a <language> file is <name>_<role>...` | Rename the file for its role, or move its code. |
| `"util" is a junk drawer; name the folder for the outcome it serves` | Rename the folder after what it does. |

## Related

[`smith slices`](slices.md), [`smith lint`](lint.md), [Conventions](../reference/conventions.md).
