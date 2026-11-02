# ADR 0071: smith install links a library of the workspace

- Status: accepted
- Date: 2026-11-02

## Context

Nothing let an app depend on a library of its own workspace. After `libs/core-py` and
`libs/core-cs` existed, the up and down columns of the Projects view stayed at 0 0, the Install
dialog did not list the workspace projects, and a library name given to `smith install` went to
the language registry (pip looked `core-py` up on PyPI). The owner's rules for this change:
no special flag, never choose silently for a user who read no manual, and every language
behaves the same ([ADR 0053](0053-every-language-plugin-follows-one-contract.md)).

## Decision

`smith install -w <project> <name>...` reads each name. A name is a library of the workspace
when it is the name of a project, its folder, or an identifier the project exports (an npm
package name, a Python distribution name, a module path). Any other name goes to the registry, as
before.

- `workspace:<name>` (the prefix pnpm and yarn use) always means the project. `--registry`
  always means the registry, for every name of the command.
- A plain name that matches a project is checked against the registry with the existing
  `dependencies/latest` method (8 seconds, a registry that does not answer is "unknown", the
  plugin is never killed). No such package: the project is linked and one line says so. A
  package exists, or the registry could not be asked: smith asks which is meant and explains
  both answers; without a terminal, or with `--yes`, it refuses and prints both fixes. smith
  never chooses silently.
- A new plugin capability `link` adds the methods `dependencies/link` and `dependencies/unlink`.
  A link request has a plan form (`plan`): the plugin writes nothing and answers the manifest,
  the line it would write and, where the ecosystem has a real choice, the options of the version
  range with their meaning. The host shows the plan before it asks the plugin to write.
- The host refuses, before any plugin is called: a project linked to itself, an app as the
  library (only a lib or a package can be depended upon), a library of another language, a
  cycle (the cycle is named), and a name that is no project (with near names). A plugin without
  `link` is refused with one sentence that names the manifest to edit by hand.
- A failed link puts the manifest and lock files back, with the guard of a failed install.
- `smith uninstall -w <project> <library>` removes the link, and says when nothing was linked.
- A dependency a manifest declares is a graph edge before any file imports it: the Go plugin now
  reports the modules `go.mod` requires among the project's imports (Node, Python, Rust and .NET
  already did).
- The Python CI setup and `smith install` install projects in dependency order, so a library is
  in the venv when pip meets the project that lists it, and pip never asks PyPI.

Node, Python, Go, Rust and .NET implement `link`. The other language plugins do not declare it.

## Consequences

- A user links a library with the command they already know, in every language that has `link`.
- The TUI Install dialog and a new VS Code command ask first "a library of this workspace or a
  package from the registry"; the library list holds only valid libraries.
- The TypeScript, Python and C# SDKs carry `link` in their generated types and members, and
  can implement it.
- Not done: a range choice for languages other than Node, the JVM, Ruby, PHP, Dart and the
  other plugins, and the JetBrains plugin (it has no dialog to add a dependency).

## Sources

- `apps/smith/internallink`, `docs/manual/install.md`, `docs/plugins/creating/language-plugin-contract.md`
