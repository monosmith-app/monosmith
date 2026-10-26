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
use and switches between them. Today they are `vertical-slices` and `none` (no table, no
checks); choosing `none` removes the table, with the `languages` and `exclude` under it.

`smith new` turns on `vertical-slices`. A converted workspace does not get it
until you add the line, because an existing codebase rarely follows it yet.

A convention is any plugin applied workspace-wide. To give some projects a
plugin and not others, put it in their [stack](stacks.md) instead: a
structure plugin such as `vertical-slices` can be either.

| Convention | Languages | What it does |
|---|---|---|
| `vertical-slices` | Go, Python, Dart, Rust, Ruby, C++, .NET, JVM, PHP | a `slice-check` task, run before `lint` |
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
