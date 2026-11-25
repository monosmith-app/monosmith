# ADR 0110: A plugin gives the standard configuration of a converted project

- Status: accepted
- Date: 2026-11-25

## Context

`smith adopt --init` and `smith convert` keep a project's configuration files as the repository had
them. That is safe, but it has a cost. In a converted Node workspace, a library's package.json still
points `main` and `types` at its build, so Ctrl+click on an import opens the built `.d.ts`, not the
TypeScript source. A new library created with `smith add` then copies that shape from its siblings,
so the old shape spreads. A workspace that smith creates exports a library's sources instead.

`smith adopt --init` already asks what happens to a pipeline the repository has: replace it and keep
a copy as `<file>.pre-smith` (the default), or keep it. Project configuration needs the same choice,
for every language, and each language knows its own files.

## Decision

1. The plugin protocol gets the method `project/standard` and the capability `standard`. The host
   sends the plugin's projects (`StandardParams{projects}`). The plugin answers the standard version
   of each project's configuration files (`StandardFileContent`): `project` (empty for a file of the
   workspace), `path`, `mode`, the content or the keys, and `reason`, one sentence on what the
   standard gives. The plugin writes nothing and compares nothing: the host compares with the files
   on disk.
2. Two modes. `file`: the plugin gives the whole file; the host writes it and keeps the file it
   replaces as `<path>.pre-smith` (`-2` and up when a copy is already there, so none is lost).
   `keys`: the plugin gives whole values of top-level keys of a JSON or TOML file; the host sets
   them in place and keeps every other key. A manifest (package.json, pyproject.toml, Cargo.toml)
   holds the dependencies, the name and the version, so it is never replaced or copied whole. Key
   values travel as raw JSON, so the members of an object keep the plugin's order: the conditions of
   a package.json `exports` are read in order.
3. A plugin offers a file only where the standard cannot change how the project builds. A file that
   carries a choice of the project (a browser project's `jsx`, `paths`, `strict: false`, a build file
   that holds the dependencies) is left out, and the project keeps it.
4. `smith adopt --init` asks "Replace them, keep a copy" (the default) or "Keep mine"
   (`--project-config replace|keep`; without a terminal, replace and say so). `smith convert` takes
   `--project-config`, replace by default. The step runs after the workspace is written.
5. `smith doctor` asks every plugin with the capability. Each project the plugin has a standard for
   gets the check "<folder> uses smith's standard configuration", and the workspace's own files the
   check "the workspace uses smith's standard configuration". A difference is a warning, since the
   project works. Its fix writes the copy and the standard; it is never applied by `--yes` alone.
   The remedy names the `[[doctor.accept]]` entry that keeps the project's own files.
6. Node, first: a library whose entry points lead to its build gets the `source` condition first in
   `exports["."]` (other conditions and subpaths keep their order; a library without `exports` gets
   one from `main` and `types`, and `"./*": "./*"` so a deep import still resolves). The root
   `tsconfig.base.json` gets `"customConditions": ["source"]`, but only when it resolves modules the
   node16, nodenext or bundler way; otherwise no library gets the condition, since nothing would
   read it. A project's `tsconfig.json` and `tsconfig.build.json` take the typescript stack's shape
   when their own options are ones the standard has too (rootDir src, Node's types, emit options,
   `outDir` when the build passes `--outDir`). An Nx layout (`tsconfig.lib.json`) keeps its files.

## Consequences

- In a converted workspace, Ctrl+click on a library import opens its source, and tsc reads the
  source with no build first. Node, bundlers and npm ignore the `source` condition, so what runs and
  what is published still read the build.
- New libraries copy the standard shape from their siblings once the old ones take it.
- Python, Dart, JVM, Rust and PHP give their standard files through the same method in later
  releases. A `.csproj`, `CMakeLists.txt`, `.gemspec` or `build.gradle.kts` holds the build and its
  dependencies, so no plugin offers to replace it.
