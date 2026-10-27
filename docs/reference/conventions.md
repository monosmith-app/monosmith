# Conventions

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). A convention is a plugin: see [plugins](plugins.md) and [writing a plugin](../guides/writing-a-plugin.md).

A language plugin knows how to build, test and release a kind of project. A
**convention** adds the rules a team holds on top: how code is laid out, which
checks run, what a pipeline adds. A convention applies to the projects of the
languages it supports, in a workspace that turns it on in `smith.toml`:

```toml
[conventions.vertical-slices]
# languages = ["go"]     # default: every language the convention supports
# exclude = ["legacy"]   # projects it leaves alone
```

You need not edit the file for this one: it is the workspace's **architecture**, and
[`smith architecture`](../manual/architecture.md) (or the Architecture section of the Settings
tab in [`smith tui`](../manual/tui.md)) lists the architectures smith knows, shows the one in
use and switches between them. Today they are `vertical-slices`, `hexagonal`, `clean`, `ddd`,
`modular-monolith`, `mvc`, `package-by-layer` and `none` (no table, no checks); choosing `none` removes the table, with the `languages` and `exclude` under it.

`smith new` turns on `vertical-slices`. A converted workspace does not get it
until you add the line, because an existing codebase rarely follows it yet.

A convention is any plugin applied workspace-wide. To give some projects a
plugin and not others, put it in their [stack](stacks.md) instead: a
structure plugin such as `vertical-slices` can be either.

| Convention | Languages | What it does |
|---|---|---|
| `vertical-slices` | Go, Python, Dart, Rust, Ruby, C++, .NET, JVM, PHP | a `slice-check` task, run before `lint` |
| `hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc`, `package-by-layer` | Go, Python, Node, JVM, .NET, PHP, Dart, Rust, Ruby, C++ | an `architecture-check` task, run before `lint`, and a `smith doctor` warning for a project with no folder named for a layer (see below) |
| `vertical-slices` | Node | `smith doctor` checks that the root ESLint config loads `@monosmith/eslint-config`, where Node's slice rules live (see below); nothing is written |

## How a convention combines with a project

1. The language plugin discovers each project and its tasks, and the
   project's [stack](stacks.md) adds its plugins' tasks.
2. Each configured convention adds its tasks to the projects it applies to. A
   task the project already has is kept as it is; a convention can also add a
   dependency to an existing task (`lint` depends on `slice-check`).
3. `smith.toml` overrides apply last, so they win over both.

`smith explain <project>:<task>` shows where each field came from
(`from convention vertical-slices`). A convention that answers `ci` or
`doctor` is asked about the projects it applies to, so its pipeline steps and
checks join the language plugins'. A convention that can `generate` adds its
part to every new project it applies to.

## vertical-slices

Code is organised by what it does: one folder per outcome (a slice), flat, its
files named for the role they play.

| Language | Checked by | Rules |
|---|---|---|
| Go | `slice-check` task (`smith slices check --lang go`) | a file is `<name>_<role>.go` (`_test`, and `_windows`-style platform suffixes allowed); only `doc.go` and `main.go` sit at a module's root |
| Python | `slice-check` task (`smith slices check --lang python`) | a module is `<name>_<role>.py`, a test `test_<name>.py`; only `__init__.py` and `__main__.py` sit at a package's top folder; packages are snake_case |
| Dart, Flutter | `slice-check` task (`smith slices check --lang dart`) | files and folders are lowercase_with_underscores (Effective Dart); a slice under `lib/src/` holds `<name>_<role>.dart` files and its barrel `<slice>.dart`; `lib/` holds the public libraries (and a Flutter app's `main.dart`), `bin/` the entry points; tests mirror `lib/` under `test/` as `<basename>_test.dart`, with test doubles in a `testing/` folder; generated `*.g.dart`-style files and the platform folders are not checked |
| Rust | `slice-check` task (`smith slices check --lang rust`) | modules are snake_case; `src/` holds `main.rs`, `lib.rs` and slice modules, each a folder of `<name>_<role>.rs` files whose API is its `mod.rs` (or the `<slice>.rs` beside the folder); `src/bin/` holds binaries; integration tests in `tests/` are `<name>_test.rs`, test doubles in a `testing/` folder |
| Ruby | `slice-check` task (`smith slices check --lang ruby`) | files and folders are snake_case; `lib/<gem>.rb` loads the gem, and `lib/<gem>/` holds `version.rb` and the slices, each a folder of `<name>_<role>.rb` files whose API is the barrel named after it (`orders/orders.rb`, or `orders.rb` beside the folder); specs and tests mirror `lib/` as `<basename>_spec.rb` or `<basename>_test.rb`, shared setup in `support/` |
| C, C++ | `slice-check` task (`smith slices check --lang cpp`) | both common spellings, snake_case (Google style) or PascalCase (LLVM, Qt), but not mixed in one folder; `src/` holds `main.cpp` and the slices, each a folder of `<name>_<role>` or `<Name><Role>` files; `include/<target>/` holds the public headers, a slice's API being the header named after it; tests in `tests/` are `<name>_test.cpp` or `<Name>Test.cpp`, test doubles in a `testing/` folder |
| .NET (C#, F#) | `slice-check` task (`smith slices check --lang dotnet`) | a file is named after its type in PascalCase and folders are PascalCase namespaces: the root holds `Program.cs`, a slice is a folder of `<Name><Role>.cs` files (`PlaceOrderUseCase.cs`) whose API is the file named after it (`Orders/Orders.cs`); tests, in `tests/` or a `*.Tests` project, are `<Name>Tests.cs`, test doubles in a `Testing/` folder; `bin/`, `obj/`, `Properties/` and generated `*.Designer.cs`-style files are not checked |
| JVM (Kotlin, Java, Android) | `slice-check` task (`smith slices check --lang jvm`) | a file is named after its class in PascalCase, with the role as a suffix (`PlaceOrderUseCase.kt`), and packages are lowercase; sources live in `src/<set>/<java or kotlin>/`, a slice being a package of `<Name><Role>` files, with entry points (`Main`, `<Name>Application`, an Android `<Name>Activity`) beside them; tests in `src/test/` are `<Name>Test`, test doubles in a `testing` package; `res/` and `resources/` are not checked |
| PHP | `slice-check` task (`smith slices check --lang php`) | PSR-1 and PSR-4: a file holds one class named like it, in PascalCase with the role as a suffix (`PlaceOrderUseCase.php`), and folders below `src/` and `tests/` are PascalCase namespaces; `src/` holds the slices (and a framework `Kernel.php`); tests mirror `src/` under `tests/` as `<Name>Test.php`, test doubles in a `Testing/` folder |
| Node | `@monosmith/eslint-config`'s vertical-slice rules, in `lint` (the `eslint` toolchain plugin of the `node-jest` and `node-vitest` stacks) | the same rules, plus slice boundaries and cycles |

The roles are `handler`, `use_case`, `algorithm`, `policy`, `model`,
`contract`, `mapper`, `validator`, `repository`, `client`, `store`,
`middleware`, `error`, `config` and `enum`; languages that name a file after its class spell them as a PascalCase suffix (`PlaceOrderUseCase`, `OrderTotalPolicy`). No folder is named `util(s)`,
`helper(s)`, `common`, `shared`, `manager`, `processor`, `data` or `misc`, in any case (`Helpers/` too).
`testdata`, `fixtures`, `vendor`, `dist`, `build` and dot-folders are not
checked.

A slice with more than 12 production files (tests, `doc.go`,
`__init__.py`, Dart barrels and Rust `mod.rs` files do not count) is reported for review, without failing: it
usually holds two outcomes and should become two slices. If it should not,
write the exception down in the project's README.

`slice-check` is cached on the project's source files, so it runs
again only when one of them changes, appears or disappears.

## Architectures that check imports

`hexagonal`, `clean`, `ddd`, `modular-monolith`, `mvc` and `package-by-layer` differ from `vertical-slices`:
they do not look at file names, they look at what each file imports. smith places every source
file in a **layer** by the folder names on its path, reads the imports of the project's own
code, and reports an import that goes where the architecture forbids. It runs as
`smith slices check --architecture <name>`, the project's `architecture-check` task, and fails
the project like `slice-check` does. Only one architecture is in force at a time: choosing
another with `smith architecture` removes the table of the first.

A layer is named by a folder anywhere on the path (`src/domain/...`, `internal/adapters/...`,
`com/acme/domain/...`). Names are compared without case, dashes or underscores, so `use-cases`,
`use_cases` and `UseCases` are the same folder. Test files and the folders `test`, `tests`,
`spec`, `e2e`, `testing`, `testdata`, `fixtures`, `vendor`, `node_modules`, `dist`, `build`,
`target`, `bin`, `obj` and dot-folders are not checked, so tests may cross layers. Code outside
every layer (a `main.go`, an `index.ts`, a composition root) may import anything.

If a project has source files but no folder named for a layer, the check prints a `review:`
line and passes: there is nothing to apply the rules to. `smith doctor` shows the same thing
as a warning.

### hexagonal

Ports and adapters. The **core** is `domain`, `application`, `core` or `ports`: business rules
and the interfaces (ports) they need, with no framework, database or transport. **Adapters**
are `adapters` or `infrastructure`: web handlers, databases, queues, clients, optionally under
`inbound/` and `outbound/` (or `driving/` and `driven/`).

| Rule | Reported as |
|---|---|
| the core never imports an adapter | `core code must not depend on adapters code` |
| an adapter never imports another adapter (`adapters/http` into `adapters/postgres`); they meet only through the core's ports | `adapters http must not depend on adapters postgres` |

An adapter may import the core. Code directly in `adapters/`, not in a sub-folder, belongs to
no adapter in particular.

### clean

The dependency rule: source code depends only inward. From the inside out the layers are

| Layer | Folders that mean it |
|---|---|
| entities | `entities`, `entity`, `enterprise` |
| use-cases | `use-cases`, `usecase`, `interactors` |
| interface-adapters | `interface-adapters`, `adapters`, `controllers`, `presenters`, `gateways` |
| frameworks | `frameworks`, `drivers`, `frameworks-and-drivers`, `infrastructure` |

A layer may import itself and the layers inside it, never one outside it
(`entities code must not depend on use-cases code`).

### ddd

Domain-driven design. Each **bounded context** is a folder (`sales/`, `billing/`, or under
`contexts/`) holding its own layers, from the inside out: `domain`, `application`,
`infrastructure` (or `infra`).

| Rule | Reported as |
|---|---|
| domain never imports application or infrastructure; application never imports infrastructure | `domain code must not depend on infrastructure code` |
| a context never imports a layer of another context; it may import what sits at that context's root (its public API file or package) | `reaches into the domain layer of the billing context` |

The context is the folder just above the layer folder. `src`, `main`, `java`, `kotlin`,
`internal`, `lib`, `app`, `pkg`, `cmd` and `contexts` do not count as one, so a project with a
single `domain/`, `application/`, `infrastructure/` is one context and only the layer rule
applies. A layout that puts the layer first (`domain/sales/`) is one context to the check.
Aggregates, value objects and the ubiquitous language are not checked; they are not visible in
imports.

### modular-monolith

One program split into **modules**, the folders directly inside a `modules/` folder
(`modules/orders/`, `src/modules/billing/`). A module's public face is what sits directly in
its folder, and its `api/` (or `public/`) folder; everything deeper is internal.

| Rule | Reported as |
|---|---|
| code outside a module, including `main` and other modules, imports only a module's public face | `reaches into the internals of the orders module` |

A module may import all of its own code. The check does not look inside a module, so it combines
with a layer naming of your own but does not enforce one. Dependencies between modules that
form a cycle are not reported.

### mvc

Models, views and controllers, each a folder: **model** (`models`, `model`), **view** (`views`,
`view`, `templates`) and **controller** (`controllers`, `controller`). From the inside out
they are model, view, controller.

| Rule | Reported as |
|---|---|
| the model imports neither views nor controllers | `model code must not depend on view code` |
| a view imports the model, never a controller | `view code must not depend on controller code` |

A controller may import models and views. A layer made of files and not folders (Django's
`models.py` and `views.py`, a single `controller.ts`) is not seen: the check works on folder
names.

### package-by-layer

The layout of many existing codebases, stated explicitly so it stops drifting: one folder per
technical layer, each layer calling only the ones below it. From the bottom: **model**
(`models`, `entities`), **repository** (`repositories`, `dao`, `persistence`), **service**
(`services`) and **controller** (`controllers`, `web`, `api`, `rest`).

| Rule | Reported as |
|---|---|
| a layer imports only itself and the layers below it | `service code must not depend on controller code` |

A layer may skip one: a controller may import a repository directly. Whether it should is a
team decision the check does not make.

### What is checked, and the limits

An import is read as text, so the check cannot see what is built at run time (reflection,
`import()` of a variable, dependency-injection containers), and it does not follow re-exports.
Per language, an import counts as the project's own code when:

| Language | The project's own imports |
|---|---|
| Go | under the module path of `go.mod` (read with Go's own parser) |
| Python | `from x import y` and `import x` where `x` is a top-level package or module of the project, and relative imports |
| Node (JavaScript, TypeScript) | relative paths and the `@/`, `~/`, `#/` and `src/` aliases, in `import`, `export from`, `import()` and `require()`; other `tsconfig` path aliases are not known |
| JVM (Java, Kotlin) | an `import` that starts like the file's own package (its first one or two names) |
| .NET, PHP | a `using` or `use` that starts like the file's own namespace (its first name or two) |
| Dart | `package:` imports of the project's own package, and relative paths |
| Rust | `use crate::`, `super::` and `self::` |
| Ruby | `require_relative` (a `require` through the load path is not followed) |
| C++ | quoted `#include "..."`, as written from the include root, or relative |

The layer of an import comes from the folder names in its path. Two packages of the project that
use a layer name in different places are the same layer to the check.
