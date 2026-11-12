# ADR 0098: One namespace names every new project, and one package builds the names

- Status: accepted
- Date: 2026-11-12
- Amends: [ADR 0062](0062-one-word-for-one-meaning-a-vocabulary-audit.md)

## Context

The plugin audit (job 6) found three inputs and nine private sanitizers behind one fact, the default
name of a new project. The same workspace gave `@acme/core` (npm, from `[workspace] scope`), `acme-core`
(Cargo, Python), `acme_core` (Dart, C++, Ruby), `Acme.Core` (.NET), `acme/core` (Composer) and
`example.com/acme/libs/core` (Go, from `[workspace] module_prefix`). Dart and C++ had no skip-prefix
rule, so project `acme-core` in workspace `acme` gave `acme_acme_core`. Pull request #1070 added an
interim rule (`WorkspaceInfo.PackagePrefix`: the scope without `@`, else the name). The surprise stayed:
`name` and `scope` both meant "the word before the project name".

## Decision

1. **One input: `[workspace] namespace`.** Its default is the workspace `name`. Every language plugin
   puts it before the project name.
2. **`scope` and `module_prefix` derive from it** unless set. The npm scope is `@<namespace>` (kebab-case)
   and the Go fallback module prefix is `example.com/<namespace>`. A person who sets `scope` or
   `module_prefix` keeps that value for npm and Go. Other languages never read them.
3. **`--package-name` wins over everything** (ADR 0084 flow, #1068).
4. **One package builds every name:** `packages/plugin-sdk/projectnaming`. It holds the namespace
   rule, one skip-prefix rule (a project name that already starts with `<namespace>-` or `<namespace>_`
   is not prefixed twice, so `acme-core` stays `acme-core` in every language) and one sanitizer per
   target syntax (`Kebab`, `Snake`, `Pascal`, `Flat`). The nine private sanitizers and the interim
   `PackagePrefix` are gone.
5. **Swift, Objective-C, VBA and Lua have no prefix.** Their name is an identifier written in code
   (a module, a class, an `import`, a `require`), and no registry needs a unique prefix: SwiftPM
   and CocoaPods identify a package by its repository, not by a global name, and a Lua rock is local
   until published. A prefix would only lengthen what the person types. They still build the name
   through `projectnaming.Name`, so one package owns the sanitizer, and a prefix is one line if a
   registry needs it later.
6. **A conformance test keeps it so.** `scaffold_naming_validator_test.go` in `projectnaming` fails for a
   language plugin whose scaffolding does not import the package, or that defines a private sanitizer.
7. **Existing `smith.toml` files migrate with `smith upgrade`** (a presented change, never silent). A
   workspace whose scope differs from its name gets `namespace` set to that scope, so the names of new
   projects do not change.
8. **`smith doctor` warns** when a project name does not follow the namespace. The rename is a fix the
   person confirms.

## Consequences

- A workspace with no `scope` now gets `@<namespace>/name` for new npm projects, not a bare name.
- The protocol `WorkspaceInfo` gains `namespace` (schema and the three SDKs regenerated).
- Names of new projects can change for a workspace that has a scope until its `smith.toml` is migrated.
  No compatibility shim is kept (beta).
