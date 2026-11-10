# ADR 0052: A project without a manifest is declared in smith.toml

- Status: accepted
- Date: 2026-11-01

## Context

smith finds a project by a manifest in its folder (`package.json`, `go.mod`, `pyproject.toml`). A
repository with one `go.mod` at the root and its programs and libraries as packages of that module
has no manifest per project, so smith found none
([#870](https://github.com/monosmith-app/monosmith/issues/870)). That layout is common in Go, and
the same need exists for any language whose manifest covers several projects. The owner decided:
single-module Go is supported in 2.0, a library inside one module cannot be released on its own,
and the mechanism is general for every language.

## Decision

A `[projects.<name>]` table that sets `root` and `language` declares a project with no manifest of
its own; `kind` (`app`, `lib` or `package`) is optional.

- The host passes the declared projects of a language to the plugin that owns it, in the `declared`
  field of `DiscoverParams` (an additive protocol change: a name, a root and a kind). The plugin
  describes each one as it describes a project it found, or answers with an error that says why it
  cannot. A declared project the plugin did not describe is an error, and so is a language that no
  plugin has.
- The core knows no language: what a declared folder means is the plugin's. The Go plugin reads the
  `go.mod` in the folder or above it, exports the module path plus the folder as the package path
  (so the graph's existing import resolution makes the edges), runs the usual tasks in the folder
  (Go commands work from a folder of a module), and adds the module's `go.mod` and `go.sum` to the
  task inputs. A folder with a `go.mod` of its own is a module project and is not declared.
- A module has one version. The Go plugin refuses the kind `package` for a declared project, and
  a lib of the module is released inside the apps that use it, as any lib is.
- Nothing is guessed at run time: the declarations are in `smith.toml`, reviewed like any config.
  `smith adopt` and `smith convert` infer them once and write them; a package the declared
  projects do not cover is reported by `smith doctor`.

Alternative not chosen: finding the projects at run time from `cmd/*` and `apps/*`. It needs no
configuration but guesses, and a new `main` package would change the build graph with nothing to
review.

## Consequences

- A repository with one root module builds, tests, caches and releases its apps with smith, without
  restructuring.
- A shared folder that no declared project holds is in no cache key until it is declared.
- Other plugins can answer `declared` the same way (a Python repository with several packages in
  one `pyproject`); none does yet.
- The TypeScript, Python and C# SDKs carry the new field in their generated types.

## Sources

- Issue [#870](https://github.com/monosmith-app/monosmith/issues/870)
- `apps/smith/projectgraph/declared_projects_mapper.go`, `libs/plugin-go/discovery/declared_projects_use_case.go`
