# Everything MonoSmith does, against every tool compared

One table with every capability MonoSmith (`smith`) has, set against the fourteen
other tools and tool families in the [comparison](README.md), as of **2026-10**,
smith 1.65.0 for the rows added on 2026-10-25 (plugins and flexibility, ready-made environments by language, the dependency graph) and 1.62.0 to 1.63.0 for the rest. Below it: what only smith does in this set, what other tools do that
smith does not yet do, and where the two are equal or behind.

This is the seller's view, and it says so. The author of this page is the smith
project, and the rows are the things smith does. A row count favours whoever chose the
rows; read the [limits](#limits-of-this-page) before you quote a score. The
[decision guide](choosing.md) is the buyer's view, with the cases where another tool
wins.

## How to read the table

Marks for the other tools are the [comparison's legend](README.md#legend), with one
addition:

| Mark | Meaning |
|---|---|
| ✅ | The tool's documentation says it has the feature |
| ◐ | Partly, experimental, or with a condition; the notes say which |
| ❔ | Not verified: the tool's documentation source was searched and neither states nor excludes the feature. Never counted as absent |
| ❌ | The documentation says it is not provided, describes a different mechanism for the same job, or the tool's documented scope is another single language |
| – | Not applicable (a tool converting from itself) |

Rules the table follows:

- **MonoSmith's column is held to a stricter standard.** ✅ means the feature is
  shipped and was run, is live on a public registry or release, or is covered by a test
  in the repository. ◐ means shipped but partial, or shipped and not verified end to
  end here. ❌ means not shipped. The last column says how each smith cell was checked
  (R, C, L, T, D or I, below) and, for a ◐ or ❌, which tracking issue or decision it
  belongs to. An issue number is a tracking issue, not a commitment.
- **Language rows** mark first-party, language-aware support: the tool finds the
  projects, knows their tasks. Running any command through a shell or a container
  is ◐ at most. A single-language tool (Lerna, Rush, Lage, Cargo, the JavaScript
  package managers) is ❌ in the other languages' rows, because its documentation
  states that scope.
- **Environment rows** ("ready-made environment", by language) are ✅ for a tool only where its
  documentation shows a built-in generator for a new project in that language and tasks that run it,
  with no rules for you to write; ◐ means the language is supported but you write BUILD files, rules,
  scripts or configuration, or add a plugin. The reasoning is on
  [Languages and environments](languages-and-environments.md).
- **Plugin rows** compare the extension surface; [Plugins and flexibility](plugins-and-flexibility.md)
  has a page of evidence per tool. **Graph rows** are explained in
  [The dependency graph](dependency-graph.md).
- **The `JS pkg mgrs` column** is npm, pnpm and Yarn workspaces together, marked by the
  best of the three; the [workspaces page](workspaces.md) says which.
- **CI rows**: ◐ means guides or recipes for running the tool in CI, with no generator.
- Cells copied from the [comparison tables](README.md) keep their sources on the tool
  pages. Cells new to this page are sourced in the "Added for the master table"
  section of each tool page; the plugin, ready-made-environment and graph rows added on
  2026-10-25 are sourced in the Sources of [plugins and flexibility](plugins-and-flexibility.md),
  [languages and environments](languages-and-environments.md) and [the dependency graph](dependency-graph.md).

| Code | How a smith cell was checked |
|---|---|
| R | run by the author of this page in a scratch workspace, 2026-10-25 |
| C | run by the authors of the [plugin catalogue](../../plugins/using/catalogue/README.md) on 1.62.0 |
| L | live public evidence: a package on its registry, or an asset on smith's own release |
| T | covered by a test in the repository (read, not re-run) |
| D | documented and read in the code; not exercised |
| I | an open issue only |

## The table

The table is generated from [`features.json`](features.json) by
[`build-everything-compared.mjs`](build-everything-compared.mjs); see
[Regenerating it](#regenerating-it).

<!-- everything-compared:table:begin -->

| Capability | **MonoSmith** | Nx | Turbo | Bazel | Pants | Lerna | Rush | moon | Buck2 | Gradle | Earthly | Lage | JS pkg mgrs | Cargo | Please | Smith: evidence, issue |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Languages and stacks** | | | | | | | | | | | | | | | | |
| JavaScript and TypeScript | **✅** | ✅ | ✅ | ◐ | ◐ | ✅ | ✅ | ✅ | ❔ | ❔ | ◐ | ✅ | ✅ | ❌ | ❔ | R |
| Go | **✅** | ◐ | ◐ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ◐ | ❌ | ❌ | ❌ | ✅ | R |
| Python | **✅** | ◐ | ◐ | ◐ | ✅ | ❌ | ❌ | ◐ | ✅ | ❔ | ◐ | ❌ | ◐ | ❌ | ✅ | C |
| Rust | **✅** | ◐ | ◐ | ◐ | ❔ | ❌ | ❌ | ✅ | ✅ | ❔ | ◐ | ❌ | ◐ | ✅ | ◐ | C |
| Java, Kotlin, Android | **◐** | ✅ | ❔ | ✅ | ✅ | ❌ | ❌ | ◐ | ✅ | ✅ | ◐ | ❌ | ❌ | ❌ | ✅ | C, not yet verified: Gradle run |
| .NET (C#, F#) | **✅** | ✅ | ◐ | ❔ | ❔ | ❌ | ❌ | ◐ | ❌ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | C |
| C and C++ | **✅** | ❔ | ❔ | ✅ | ❔ | ❌ | ❌ | ◐ | ✅ | ✅ | ◐ | ❌ | ❌ | ❌ | ✅ | C |
| Ruby | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | C |
| PHP | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | C |
| Dart and Flutter | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | C |
| Azure Bicep (infrastructure) | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | C |
| Front-end frameworks (React, Angular, Vue, Vite, Astro) | **✅** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| **Ready-made environments, by language** | | | | | | | | | | | | | | | | |
| Ready-made environment: JavaScript and TypeScript | **✅** | ✅ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ❔ | ❔ | ◐ | ◐ | ◐ | ❌ | ❔ | R |
| Ready-made environment: Python | **✅** | ◐ | ◐ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❔ | ❌ | ◐ | R |
| Ready-made environment: Go | **✅** | ◐ | ◐ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❌ | ❌ | ◐ | R |
| Ready-made environment: Rust | **✅** | ◐ | ◐ | ◐ | ❔ | ❌ | ❌ | ◐ | ◐ | ❔ | ◐ | ❌ | ❔ | ✅ | ◐ | R |
| Ready-made environment: Java, Kotlin and Android | **◐** | ◐ | ❔ | ◐ | ◐ | ❌ | ❌ | ◐ | ◐ | ✅ | ◐ | ❌ | ❌ | ❌ | ◐ | R, not yet verified end to end |
| Ready-made environment: .NET | **✅** | ◐ | ◐ | ❔ | ❔ | ❌ | ❌ | ◐ | ❌ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | R |
| Ready-made environment: Ruby | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | R |
| Ready-made environment: PHP | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ◐ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | R |
| Ready-made environment: C and C++ (CMake) | **✅** | ❔ | ❔ | ◐ | ❔ | ❌ | ❌ | ◐ | ◐ | ✅ | ◐ | ❌ | ❌ | ❌ | ◐ | R |
| Ready-made environment: Dart and Flutter | **◐** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | R, not yet verified end to end |
| Ready-made environment: Azure Bicep | **✅** | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ◐ | ❌ | ❌ | ❌ | ❔ | R |
| A new project lints, type-checks, tests and builds on its first run (a worked example is included) | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Writes a dev container with the toolchains CI installs | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| The generated pipeline installs each language's toolchain and caches its package store | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| **Project discovery and scaffolding** | | | | | | | | | | | | | | | | |
| Projects found from the manifests you already have | **✅** | ✅ | ◐ | ❌ | ◐ | ◐ | ◐ | ◐ | ❌ | ◐ | ❌ | ◐ | ✅ | ✅ | ❌ | R |
| Scaffold apps, libraries and packages | **✅** | ✅ | ✅ | ❔ | ◐ | ✅ | ◐ | ✅ | ◐ | ✅ | ❔ | ◐ | ◐ | ✅ | ◐ | R |
| Team recipes: a shared preset of plugins and settings, with questions at add time | **✅** | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Existing projects checked against their recipe, owned files upgraded | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Code layout checked inside projects (vertical slices) in nine languages | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| **The dependency graph** | | | | | | | | | | | | | | | | |
| Dependencies between projects read from source imports, in several languages | **✅** | ◐ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ◐ | T |
| Shared schemas (OpenAPI, AsyncAPI, protobuf, GraphQL) link producer and consumer projects across languages | **✅** | ❔ | ❔ | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Graph views: tree, JSON, DOT, HTML page, interactive, focus, affected | **✅** | ✅ | ✅ | ◐ | ❔ | ◐ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | R |
| Graph diff between two revisions (projects and imports added or removed) | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Rules for which projects may depend on which (boundaries, visibility) | **❌** | ✅ | ◐ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ◐ | ❔ | ✅ | I, #570 |
| Edges between projects of different languages | **◐** | ✅ | ❔ | ✅ | ◐ | ❌ | ❌ | ◐ | ✅ | ❔ | ◐ | ❌ | ❌ | ❌ | ✅ | R, by design, ADR 0028 |
| Asks what depends on a project or a file, transitively | **✅** | ✅ | ◐ | ✅ | ✅ | ❔ | ◐ | ◐ | ◐ | ❔ | ❔ | ◐ | ❔ | ◐ | ✅ | R |
| The graph as JSON for scripts | **✅** | ✅ | ◐ | ◐ | ✅ | ✅ | ◐ | ✅ | ◐ | ❔ | ❔ | ❔ | ◐ | ✅ | ✅ | R |
| A plugin can contribute graph edges for its own language | **✅** | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| **Tasks, graph and affected** | | | | | | | | | | | | | | | | |
| Tasks run in dependency order, in parallel | **✅** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | R |
| Only what changed since a git base (affected) | **✅** | ✅ | ✅ | ◐ | ✅ | ✅ | ✅ | ✅ | ◐ | ◐ | ◐ | ✅ | ✅ | ❔ | ✅ | R |
| Runs everything when it cannot be sure (no base, shallow clone, changed lockfile) | **✅** | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Affected selection by task, finer than a project | **❌** | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | D, #574 to decide; the unit is the project today |
| Dependency graph of single source files and targets (Bazel and Pants style) | **❌** | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | D, design choice: not planned |
| Compiler and runtime versions are part of the cache key | **✅** | ✅ | ◐ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| **Cache and execution** | | | | | | | | | | | | | | | | |
| Local content-addressed cache that restores outputs and replays logs | **✅** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ◐ | ✅ | ✅ | ✅ | ◐ | ✅ | ✅ | R |
| Remote cache you host yourself, no vendor service required | **✅** | ✅ | ✅ | ✅ | ✅ | ◐ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ◐ | ◐ | ✅ | T |
| Remote cache objects signed; a bad signature is a miss | **✅** | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | T |
| Explains a cache miss with one command that runs nothing | **✅** | ◐ | ◐ | ◐ | ❔ | ◐ | ❔ | ◐ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Commands to see, trim, clear and verify the local cache | **✅** | ◐ | ❔ | ◐ | ❔ | ◐ | ◐ | ◐ | ❔ | ◐ | ❔ | ❔ | ◐ | ◐ | ◐ | R |
| Local cache kept under a size limit automatically | **✅** | ✅ | ◐ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | R |
| Reports the time a cache hit saved | **✅** | ◐ | ✅ | ❔ | ✅ | ◐ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Speaks the Bazel Remote Execution API | **❌** | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | I, #345 |
| Remote or distributed task execution | **❌** | ◐ | ❔ | ✅ | ✅ | ◐ | ◐ | ◐ | ✅ | ❔ | ◐ | ◐ | ❔ | ❔ | ◐ | D, deliberate non-goal (execution-and-cache.md, section 9) |
| Hermetic or sandboxed execution | **❌** | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ◐ | ❔ | ✅ | ◐ | ❔ | ❔ | ✅ | D, deliberate non-goal; opt-in strict mode proposed in #563 |
| Pins language toolchains and downloads them for you | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ✅ | ❔ | ✅ | ❔ | ❔ | ◐ | ❔ | ❔ | R, deliberate (ADR 0020) |
| Watch mode: re-run tasks when files change | **❌** | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | D, deliberate non-goal (execution-and-cache.md, section 9) |
| **CI generation and verification** | | | | | | | | | | | | | | | | |
| Writes the CI pipeline files for you | **✅** | ✅ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | ❔ | ◐ | ◐ | ◐ | ◐ | ◐ | ◐ | R |
| GitHub Actions, Azure Pipelines and GitLab CI generated | **✅** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| CircleCI, Bitbucket Pipelines or Jenkins generated | **❌** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | I, #571 |
| CI fails when the committed pipeline is out of date, and a fix regenerates it | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Your own steps survive regeneration (named slots) | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Dependency vulnerability audit | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | D, #572 |
| CI fails when a task leaves a tracked file modified | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Release preview posted on the pull request | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Ready-made pipelines beyond CI (e2e, GitHub Pages, Azure Function) | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| **Release and publishing** | | | | | | | | | | | | | | | | |
| Versions from conventional commits | **✅** | ✅ | ❌ | ❔ | ◐ | ✅ | ❌ | ❌ | ❔ | ❌ | ❔ | ❔ | ❌ | ❌ | ❔ | L |
| Publishes to npm | **✅** | ✅ | ◐ | ❔ | ❔ | ✅ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ❌ | ❔ | L |
| Publishes to PyPI | **✅** | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| Publishes to crates.io | **◐** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | T, none needed; live publish unverified |
| Publishes to NuGet | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| Publishes to RubyGems | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| Publishes to pub.dev | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| Publishes to Maven repositories | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | D, live publish unverified |
| Releases Go modules and Packagist packages by tag | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | D, live publish unverified |
| Builds and pushes container images | **◐** | ◐ | ❔ | ❔ | ✅ | ❔ | ❔ | ❌ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | D, run unverified |
| Homebrew, Scoop and winget manifests from the release archives | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| .deb and .rpm packages attached to the release | **✅** | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| Per-platform app archives with checksums on a GitHub Release | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L, #573 |
| SBOM generation | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ◐ | ❔ | L, #348 |
| Signed provenance for what you publish | **◐** | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | D, #348 |
| Publishes without a long-lived key (trusted publishing) | **◐** | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | D, needs an issue |
| A failed release resumes: versions a registry already has are skipped, in every registry | **✅** | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Asks every registry whether it accepts the credential and the names before the first publish | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T, #347, #380 |
| Release journal of what was published where | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R, #347 |
| Publishes a VS Code extension to the Marketplace with the release | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| **Dependencies** | | | | | | | | | | | | | | | | |
| One version of each shared dependency across projects | **◐** | ◐ | ◐ | ◐ | ◐ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | R, #572 |
| Checks that a package declares what it imports | **◐** | ✅ | ❔ | ✅ | ◐ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | T, #572 |
| **Health and migration** | | | | | | | | | | | | | | | | |
| Read-only health check, each finding with a remedy | **✅** | ◐ | ◐ | ❔ | ◐ | ◐ | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | R |
| Applies fixes: diffs for file changes, questions for risky ones | **✅** | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Bring an existing repository under the tool | **✅** | ✅ | ✅ | ❔ | ✅ | ✅ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | C |
| A dry-run report of what adopting involves (what maps, what blocks), changing nothing | **✅** | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | C |
| Automated conversion from Nx | **✅** | – | ◐ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Automated conversion from Turborepo | **✅** | ◐ | – | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| Importer for Lerna, Rush, Bazel or Gradle workspaces | **❌** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | I, #553 |
| Config and generated files migrate when the tool upgrades | **✅** | ✅ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| **Plugins and extensibility** | | | | | | | | | | | | | | | | |
| A plugin model | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ◐ | ◐ | ✅ | ✅ | ❔ | ❔ | ◐ | ✅ | ✅ | R |
| Plugins are separate programs speaking a documented protocol | **✅** | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | R |
| An SDK for writing plugins | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ◐ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | L |
| A registry or catalogue of third-party plugins | **❌** | ✅ | ❔ | ◐ | ❔ | ❔ | ❔ | ◐ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | I, #164, #565 |
| Extensions are integrity-checked or sandboxed | **◐** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T, #565 |
| Named stacks: a per-project choice of toolchain plugins | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| Extensions can add support for a new language | **✅** | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ◐ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | R |
| An extension can be written in any language, not only the tool's own | **✅** | ❌ | ❔ | ❌ | ❌ | ❔ | ❌ | ◐ | ❌ | ❌ | ❔ | ❔ | ❌ | ✅ | ❌ | R |
| Each extension runs in a process of its own | **✅** | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | R |
| A trust step before an extension named by the repository can run, with pins for CI | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| A command that checks an extension keeps the protocol's promises | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| **Interfaces** | | | | | | | | | | | | | | | | |
| Full-screen terminal interface (dashboard, doctor, deps, release, settings) | **✅** | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | T |
| VS Code extension | **✅** | ✅ | ✅ | ✅ | ❌ | ✅ | ❔ | ◐ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ✅ | L |
| JetBrains IDE plugin | **❌** | ✅ | ❔ | ✅ | ❌ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | I, #303 |
| MCP server for AI assistants | **✅** | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | R |
| A documented protocol for editors to drive the engine | **✅** | ❔ | ◐ | ✅ | ◐ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ◐ | R |
| A hosted service (remote cache, analytics, CI features) | **❌** | ✅ | ✅ | ◐ | ❔ | ◐ | ❔ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | D, deliberate: no hosted service |
| **Operations** | | | | | | | | | | | | | | | | |
| Native Windows support | **✅** | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ◐ | ❔ | ❔ | ❔ | ❌ | L |
| **Score: rows marked ✅** (of 115) | **84** | 39 | 14 | 21 | 24 | 12 | 10 | 19 | 16 | 20 | 5 | 7 | 10 | 12 | 20 | ◐ 19, ❌ 12 |
| **Rows marked ❔** (not verified, not counted as absent) | – | 51 | 74 | 73 | 70 | 66 | 68 | 50 | 83 | 84 | 78 | 79 | 70 | 73 | 79 | |

<!-- everything-compared:table:end -->

### Notes on cells

Short notes where a mark needs its reason. Anything not listed is covered by the
comparison tables and the tool pages.

<!-- everything-compared:notes:begin -->

**Languages and stacks**

- *JavaScript and TypeScript.* Bazel: through separately maintained rules; Pants: beta; Earthly: any language, as containers (no language awareness).
- *Go.* Nx: named in its README, no Go page found; Turborepo: experimental. A tool whose documented scope is another single language is marked ❌.
- *Python.* Nx: a walkthrough adds Python with uv; Turborepo: experimental; Bazel: external rules_python; moon: partial; pnpm 12 can add `pypi:` packages (a package source, not Python projects).
- *Rust.* Nx: a guide publishes crates; Turborepo: experimental; Bazel: external rules_rust; pnpm 12 can add `crate:` packages; Please: rules in its plugin repository.
- *Java, Kotlin, Android.* MonoSmith: Maven (Java) was run; Gradle, Kotlin and Android were not. moon: only as a "system" task.
- *.NET (C#, F#).* Buck2: its language table lists C# as unavailable; Turborepo: a guide runs a .NET app through scripts.
- *Ruby.* moon: partial and experimental.
- *PHP.* moon: tier 1, project categorisation only.
- *Dart and Flutter.* Searched for ❔: "dart", "flutter", "pubspec".
- *Azure Bicep (infrastructure).* Searched for ❔: "bicep".
- *Front-end frameworks (React, Angular, Vue, Vite, Astro).* MonoSmith: stacks listed with smith stacks. Searched for ❔: framework names in each tool's guides were not searched beyond Nx's technologies section.

**Ready-made environments, by language**

- *Ready-made environment: JavaScript and TypeScript.* Nx: generators (`nx g`) create libraries and apps and plugins infer their tasks; the others run package.json scripts you write, or need rules.
- *Ready-made environment: Python.* Pants: `tailor` generates BUILD files for code that exists; it does not create a project.
- *Ready-made environment: Rust.* Cargo: `cargo new` and `cargo init`, with its own workspace and lockfile.
- *Ready-made environment: Java, Kotlin and Android.* MonoSmith: Maven (Java) ran on 1.65.0; Gradle, Kotlin and Android were not run here. Gradle: `gradle init` creates Java, Kotlin, Groovy and Scala applications and libraries.
- *Ready-made environment: C and C++ (CMake).* Gradle: `gradle init` creates C++ applications and libraries.
- *Ready-made environment: Dart and Flutter.* MonoSmith: `dart analyze` crashed on the test machine (Dart 3.12.0 on Windows, reproduced outside smith), so the lint task did not pass; scaffold, test and build did.
- *A new project lints, type-checks, tests and builds on its first run (a worked example is included).* MonoSmith: on 1.65.0, eleven languages in one workspace: test 14 of 14 tasks, build 10 of 10, lint 21 of 22 (the one failure is the Dart SDK crash noted on the Dart row). Searched for ❔: "new project", "first run", "generate" with the tool's scaffold commands in all fourteen sources; no page says what a generated project passes.
- *Writes a dev container with the toolchains CI installs.* Searched for ❔: "devcontainer", "dev container" in all fourteen sources (Pants, moon and Earthly mention dev containers as places to run the tool, not as something they generate).
- *The generated pipeline installs each language's toolchain and caches its package store.* MonoSmith: one GitHub workflow for eleven languages: eleven toolchain installs and seven package-store caches. Searched for ❔: "setup-node", "setup-go", "setup-python", "setup-java" in all fourteen sources: they appear in hand-written examples, not as generator output; Nx's `ci-workflow` generator exists but its output is not described.

**Project discovery and scaffolding**

- *Projects found from the manifests you already have.* ◐ means one root file or one file per project to write (turbo.json, lerna.json, rush.json, moon.yml, settings.gradle, lage.config.js); ❌ means per-target BUILD files or Earthfiles.
- *Scaffold apps, libraries and packages.* Pants: tailor writes BUILD files, not code.
- *Team recipes: a shared preset of plugins and settings, with questions at add time.* ◐: custom generators (Nx, Turborepo) or templates (moon); none documents levels, slots or versions.
- *Existing projects checked against their recipe, owned files upgraded.* Searched for ❔: "template" with "update", "upgrade", "drift" and "sync".
- *Code layout checked inside projects (vertical slices) in nine languages.* MonoSmith: Node through ESLint rules; Bicep has none. Searched for ❔: "folder", "file" with "naming", "layout", "structure" and "enforce" or "lint".

**The dependency graph**

- *Dependencies between projects read from source imports, in several languages.* Nx: imports are parsed for JavaScript and TypeScript only; its guide says that for other languages "Nx does not detect dependencies from source code imports". Please: the puku tool diagnoses missing imports. ❌: the tool's documented mechanism is declared dependencies.
- *Shared schemas (OpenAPI, AsyncAPI, protobuf, GraphQL) link producer and consumer projects across languages.* MonoSmith: #343 is still open although the feature shipped (ADR 0028). Bazel and Pants model protobuf through rules and code generation; Pants has an experimental OpenAPI backend.
- *Graph views: tree, JSON, DOT, HTML page, interactive, focus, affected.* Lerna: through the Nx project graph; Bazel: query output formats.
- *Graph diff between two revisions (projects and imports added or removed).* Searched for ❔: "graph" with "between", "since" and "compare".
- *Rules for which projects may depend on which (boundaries, visibility).* Turborepo: `turbo boundaries` is experimental; Yarn: constraints; moon: constraints and boundaries; Bazel, Buck2, Please, Pants: visibility.
- *Edges between projects of different languages.* MonoSmith: import edges stay inside one language; across languages, edges come from shared schemas you name in smith.toml (ADR 0028). Nx: plugins "connect cross-language boundaries too"; Buck2: rules "can combine languages"; Please: "Cross-language support is a big advantage"; moon: explicit `dependsOn`, implicit discovery per language; Earthly: targets of any language reference each other; Pants: code generators link languages. Searched for ❔: "cross-language", "multi-language" in Turborepo, Gradle.
- *Asks what depends on a project or a file, transitively.* Nx: "pass `--files` to any affected command"; Bazel: `rdeps`, `allrdeps`; Pants: `pants dependents --transitive`; Please: `reverseDeps`, `whatinputs`; Turborepo: a filter that includes dependents; Rush: `--impacted-by`; moon: `--dependents`; Lage: the scope includes dependents; Cargo: `cargo tree --invert`; Buck2: `rdeps` in cquery. Searched for ❔: "dependents", "reverse", "rdeps" in Lerna, Gradle, Earthly and the JavaScript package managers.
- *The graph as JSON for scripts.* Turborepo: `--graph` JSON output is deprecated in favour of `turbo query`; Bazel and Buck2: query output formats; Rush: `rush list --json` lists projects; the JavaScript package managers print dependency trees as JSON. Searched for ❔: "json" with "graph" in Gradle, Earthly and Lage.
- *A plugin can contribute graph edges for its own language.* Nx: a plugin exports `createDependencies`; Pants: "Add new forms of dependency inference". Searched for ❔: "dependency inference", "createDependencies", "dependencies" in extension pages of the others.

**Tasks, graph and affected**

- *Only what changed since a git base (affected).* ◐: reverse-dependency queries or up-to-date checks, no changed-files selection found (Bazel, Buck2, Gradle); Earthly's auto-skip needed Earthly Cloud. Cargo: no changed-since selection found.
- *Runs everything when it cannot be sure (no base, shallow clone, changed lockfile).* Turborepo documents that `--affected` falls back to running all tasks on a shallow clone. Nx, Turborepo and Pants otherwise document shallow clones as a CI setting. Searched for ❔: "fall back", "shallow", "fetch-depth", "no base" in the other thirteen.
- *Compiler and runtime versions are part of the cache key.* Nx and Lerna: runtime values can be hash inputs; Pants: documents not downloading an unfingerprinted Go toolchain.

**Cache and execution**

- *Local content-addressed cache that restores outputs and replays logs.* Buck2: artifacts in buck-out plus an in-memory cache; pnpm: experimental task cache.
- *Remote cache you host yourself, no vendor service required.* MonoSmith: HTTP, S3 and Azure Blob; not run against a live bucket. Lerna: build your own or use Nx Cloud; pnpm: experimental, from 12.12; Cargo: third-party sccache.
- *Remote cache objects signed; a bad signature is a miss.* Turborepo documents HMAC-SHA256 signing and treats failures as misses; pnpm signs results of its experimental pipeline cache.
- *Explains a cache miss with one command that runs nothing.* ◐ means a documented procedure: Nx (troubleshooting guide), Turborepo (--summarize and comparison), Bazel (--explain writes a log during a build), Gradle (a debugging chapter), moon (a debugging guide), Lerna (links to Nx Cloud).
- *Commands to see, trim, clear and verify the local cache.* ◐: a clear or purge command (nx reset, bazel clean, rush purge, moon clean, cargo clean, plz clean); npm has `cache verify`, pnpm `cache prune`.
- *Local cache kept under a size limit automatically.* Nx: maxCacheSize; Turborepo: eviction by age; Earthly: garbage collection of layers.
- *Reports the time a cache hit saved.* MonoSmith: time saved per run; no hit-rate report yet. Turborepo prints it; Pants logs cache statistics; Nx and Lerna through Nx Cloud; Gradle compares times in a guide.
- *Speaks the Bazel Remote Execution API.* Please: experimental, from v15.
- *Remote or distributed task execution.* ◐: through Nx Cloud (Nx, Lerna), experimental cobuilds (Rush), unfinished action distribution (moon), remote runners (Earthly), BuildXL (Lage), experimental (Please). Workspaces: npm ❌, pnpm and Yarn ❔.
- *Pins language toolchains and downloads them for you.* MonoSmith: pins checked by doctor and written to .tool-versions; never downloaded (ADR 0020). moon downloads tool versions (its tier 3); Gradle provisions JDK toolchains; Rush and pnpm manage their package manager's version.
- *Watch mode: re-run tasks when files change.* smith dev starts dev servers but does not rebuild.

**CI generation and verification**

- *Writes the CI pipeline files for you.* ◐ means guides or recipes, no generator found.
- *GitHub Actions, Azure Pipelines and GitLab CI generated.* MonoSmith: GitHub Actions run here; Azure and GitLab renderers read.
- *CircleCI, Bitbucket Pipelines or Jenkins generated.* Nx generates CircleCI and Bitbucket Pipelines; its documentation says the generator does not support Jenkins.
- *CI fails when the committed pipeline is out of date, and a fix regenerates it.* Searched for ❔: "out of date", "drift", "stale" next to "workflow", "pipeline" or "CI".
- *Your own steps survive regeneration (named slots).* Searched for ❔: no generator documents regeneration.
- *Dependency vulnerability audit.* MonoSmith: Node (blocking when fixable), Python, Go; other ecosystems have none. npm, pnpm and Yarn each document an audit command.
- *CI fails when a task leaves a tracked file modified.* Searched for ❔: "git diff --exit-code", "uncommitted changes", "dirty".
- *Release preview posted on the pull request.* moon's CI action posts the result of `moon ci`; Nx posts self-healing fixes; neither posts a release plan. Searched for ❔: "comment" next to "pull request".
- *Ready-made pipelines beyond CI (e2e, GitHub Pages, Azure Function).* Searched for ❔: not searched beyond the CI guides.

**Release and publishing**

- *Versions from conventional commits.* Turborepo: Changesets; Rush: change files; moon: "At this time, no"; Gradle and Cargo leave versions to you; Pants: version strings from git state.
- *Publishes to npm.* Turborepo: through Changesets. moon: its FAQ says it does not version or publish.
- *Publishes to PyPI.* Pants: experimental Python backend.
- *Publishes to crates.io.* MonoSmith: unit-tested; monosmith itself is not on crates.io.
- *Publishes to Maven repositories.* MonoSmith: GitHub Packages and private feeds; JitPack by tag; Central not verified.
- *Releases Go modules and Packagist packages by tag.* MonoSmith: tag-based; not run against Packagist here.
- *Builds and pushes container images.* MonoSmith: no Docker daemon was available to run it. Nx: experimental; Pants: docker publish; Earthly: builds and pushes images.
- *.deb and .rpm packages attached to the release.* Pants: an experimental backend that builds .deb files (packaging, not release).
- *Per-platform app archives with checksums on a GitHub Release.* MonoSmith: Go, Rust and .NET apps; other languages are tagged without assets. Searched for ❔: not searched beyond release and publish pages.
- *SBOM generation.* MonoSmith: CycloneDX on the release for npm and Go apps only. npm and pnpm have an `sbom` command; Cargo lists an unstable SBOM option.
- *Signed provenance for what you publish.* MonoSmith: GitHub Actions only: npm provenance and SLSA attestation of release files. Nx: a CI guide grants id-token for npm provenance; pnpm documents `publish --provenance`.
- *Publishes without a long-lived key (trusted publishing).* MonoSmith: RubyGems, pub.dev, PyPI and the Marketplace; not every registry. Nx: a CI guide grants id-token for npm provenance; npm documents `npm trust` (OIDC trust relationships with CI providers); pnpm can use the CI identity for dist-tags.
- *A failed release resumes: versions a registry already has are skipped, in every registry.* Nx: for npm, it looks the version up before publishing so that the same publish can run again; Lerna: its FAQ says a partly failed publish may leave the repository in an inconsistent state. Searched for ❔: "already published", "skip", "re-run" next to "publish" or "release".
- *Asks every registry whether it accepts the credential and the names before the first publish.* MonoSmith: npm, PyPI, NuGet, RubyGems. Searched for ❔: "preflight", "credential" and "before publish" were not searched in every tool; treat as not documented.
- *Release journal of what was published where.* MonoSmith: --status reads dist/release/<project>/journal.json. Searched for ❔: not searched beyond release pages.
- *Publishes a VS Code extension to the Marketplace with the release.* MonoSmith: MonoSmith's own extension is on the Marketplace. Searched for ❔: "vsce", "publish" next to "extension".

**Dependencies**

- *One version of each shared dependency across projects.* MonoSmith: Node (run) and Python (code read, not driven); Rust, .NET, Dart and Go are skipped. Rush: `rush check`; Yarn: constraints; pnpm: catalogs; Cargo: inherited dependencies; Gradle: platforms and catalogs; Bazel: one version per module.
- *Checks that a package declares what it imports.* MonoSmith: Node packages only. Nx: @nx/dependency-checks; Rush: `rush scan` finds phantom dependencies; Bazel: strict dependencies; npm documents catching phantom dependencies.

**Health and migration**

- *Read-only health check, each finding with a remedy.* ◐: nx report, turbo boundaries, tailor --check, lerna repair, rush check and scan, moon sync, npm and pnpm doctor, Yarn constraints.
- *Applies fixes: diffs for file changes, questions for risky ones.* MonoSmith: an edited workflow was regenerated by doctor --fix. Lerna repair and moon sync repair configuration; neither previews a diff or asks.
- *Bring an existing repository under the tool.* Rush documents adding projects one at a time.
- *A dry-run report of what adopting involves (what maps, what blocks), changing nothing.* Lerna's `init --dryRun` previews the file changes only.
- *Automated conversion from Nx.* Turborepo: a migration guide; moon: `moon ext migrate-nx`.
- *Automated conversion from Turborepo.* Nx has a guide for coming from Turborepo; moon: `moon ext migrate-turborepo`.
- *Importer for Lerna, Rush, Bazel or Gradle workspaces.* Searched for ❔: not searched; no tool is known to document one.
- *Config and generated files migrate when the tool upgrades.* Nx: `nx migrate`; Lerna: `lerna repair`.

**Plugins and extensibility**

- *Plugins are separate programs speaking a documented protocol.* Nx: "plugin isolation runs each plugin in a separate forked process", but a plugin is a Node package; moon: WASM plugins in any language that compiles to WASM, called experimental; Cargo: any program named cargo-<name> becomes a subcommand.
- *An SDK for writing plugins.* MonoSmith: Go (module packages/plugin-sdk, v0.34.0 on the Go proxy) TypeScript (@monosmith/plugin-sdk 0.2.0 on npm) and Python (monosmith-plugin-sdk 0.1.0 on PyPI); no C# SDK is published yet.
- *A registry or catalogue of third-party plugins.* Bazel: the Central Registry hosts modules; moon: a community plugin registry for its proto tool manager; Gradle: the Gradle Plugin Portal.
- *Extensions are integrity-checked or sandboxed.* MonoSmith: a plugin the workspace names must be trusted, runs with a minimal environment and is pinned by SHA-256 in CI (ADR 0036); it is not sandboxed. moon: WASM plugins "have access to a sandboxed file system"; see also the trust row.
- *Named stacks: a per-project choice of toolchain plugins.* Searched for ❔: "stack", "preset" were not searched; treat as not documented.
- *Extensions can add support for a new language.* moon: toolchain plugins, experimental; Bazel: "Support for a new language is typically done with a rule"; Pants: "Add support for new languages"; Gradle: language plugins; Please: "build rules for additional languages". Searched for ❔: "plugin", "extension", "extending" in Turborepo, Lerna, Rush, Earthly, Lage, the JavaScript package managers and Cargo (Rush plugins share configuration and add features; Cargo subcommands add commands, not languages).
- *An extension can be written in any language, not only the tool's own.* MonoSmith: a 102-line Python plugin with no SDK passes smith plugin check. Nx: "The plugin is a Node package"; Bazel and Buck2: Starlark; Pants: "typed Python 3 code"; Rush: an npm package; Gradle: Java, Groovy or Kotlin; Yarn plugins and pnpm hooks are JavaScript; Please: its build language; moon: "any language that compiles to WASM"; Cargo: any program named cargo-<name>. The programs an extension launches can be in any language; this row is about the extension itself. Searched for ❔: "plugin" in Turborepo, Lerna, Earthly and Lage.
- *Each extension runs in a process of its own.* Nx: "Plugin isolation runs each plugin in a separate forked process"; Cargo: a subcommand is an external executable. Searched for ❔: "isolat", "separate process", "in-process" in the extension pages of the other tools.
- *A trust step before an extension named by the repository can run, with pins for CI.* moon: its proto tool manager says a plugin "must be trusted with `proto trust` before the plugin is loaded"; Gradle: dependency verification covers the plugins a build consumes. Searched for ❔: "trust", "checksum", "sha256", "signature" near "plugin" in all fourteen documentation sources; the search was narrow, so the row is not claimed as unique.
- *A command that checks an extension keeps the protocol's promises.* Searched for ❔: "conformance", "validate plugin", "test your plugin" were searched in the extension pages; Gradle documents TestKit and Nx testing helpers for plugin authors, which are libraries, not a checking command.

**Interfaces**

- *Full-screen terminal interface (dashboard, doctor, deps, release, settings).* Nx and Turborepo document a terminal UI for task output.
- *VS Code extension.* Pants: "We have not yet developed tight IDE integrations"; moon: "early stages"; Earthly: an extension for Earthfiles.
- *A documented protocol for editors to drive the engine.* Turborepo: a language server; Bazel: the Build Event Protocol; Pants: BSP for IntelliJ; Gradle: the Tooling API; Please: a language server for BUILD files.
- *A hosted service (remote cache, analytics, CI features).* Earthly: its cloud was shut down in 2025; Bazel: commercial services are linked from its documentation.

**Operations**

- *Native Windows support.* MonoSmith: x64 and arm64 archives on the release; this page's runs were on Windows 11. Pants lists Windows 10 with WSL 2; Earthly calls native Windows experimental; Please: "not supported natively".

<!-- everything-compared:notes:end -->

## Only MonoSmith, in this set

<!-- rows:only-smith env-devcontainer, find-recdrift, find-layout, find-diff, ci-drift, ci-dirty, ci-comment, rel-nuget, rel-gems, rel-pub, rel-brew, rel-vsce -->
<!-- rows:weak env-first, env-ci, ci-slot, ci-tpl, plug-stack, plug-check -->
<!-- rows:nearly-only lang-rb, lang-php, lang-dart, lang-bicep, env-py, env-go, env-net, env-rb, env-php, env-bicep, find-recipe, find-schema, cache-why, cache-manage, rel-pypi, rel-linux, rel-resume, health-doctor, health-fix, health-dry, plug-oop, ui-tui, ops-win -->

These are the rows where smith is ✅ and every other tool is ❌, ❔ or not applicable.
The wording is deliberate: **no other tool in this comparison documents it, in the
sources searched.** It does not say no other tool can do it. No cell on this list is a
positively documented absence; each ❔ means the tool's documentation source was
searched for the feature's usual names (listed per row in the notes above) and found
nothing. The searches covered all fourteen tools' documentation sources as read on
2026-10-25. Plugins, third-party tools, commercial services and the tools' blogs were
not searched.

| Only smith documents | Evidence | What it is |
|---|---|---|
| A **drift guard on generated CI**: `smith doctor` and `smith ci generate --check` fail when the committed pipeline is out of date, and `doctor --fix` regenerates it | R: edited a generated workflow in a scratch workspace; both commands exited 1 and `doctor --fix` made `ci generate --check` pass | Nx is the only other tool that generates pipeline files, and its documentation describes a generator, not a check |
| **CI fails when a task leaves a tracked file modified** | T: `apps/smith/cipipeline/run_phase_use_case_test.go` and the golden-path test | Catches a build that rewrites committed files before a release refuses a dirty tree |
| A **release preview posted on the pull request**, updated by later runs | T: the `apps/smith/releasepreview` tests | moon's CI action posts `moon ci` results and Nx posts self-healing fixes; neither posts a release plan |
| A **graph diff between two revisions**: projects and imports added or removed | R: `smith graph --diff HEAD` listed the two projects added | `smith graph --diff <rev>` |
| **Existing projects checked against their recipe** and owned files upgraded | T | Recipes carry levels, slots and versions; `smith doctor` reports drift, `smith upgrade` rewrites the files smith owns |
| **Code layout checked inside projects** (vertical slices) in nine languages | T; the rules are in [Conventions](../../reference/conventions.md) | Role-named files and no junk-drawer folders, in each language's own naming |
| **Publishing to NuGet, RubyGems and pub.dev** from the same release | L: `monosmith` is live on nuget.org (1.2.0 to the current version), rubygems.org and pub.dev, 1.63.0 | Nx and Pants document other registries; none of the fourteen documents these three |
| **Homebrew, Scoop and winget manifests generated from the release archives** | L: `monosmith.rb`, `monosmith.json` and three winget files are assets of release `smith@1.63.0` | Pointing at the release's own archives with their SHA-256 |
| **Publishing a VS Code extension to the Marketplace** with the release | L: the MonoSmith extension is on the Visual Studio Marketplace, 0.2.0, updated 2026-10-24 | |
| A **dev container written from the workspace**, with the toolchains CI installs at the same versions | R: `smith new` and `smith add` wrote `.devcontainer/devcontainer.json` with one feature per toolchain for eleven languages (1.65.0) | Pants, moon and Earthly mention dev containers as places to run the tool; none documents generating one |

Six more rows have only smith at ✅ but were searched too little to claim: generated
pipelines whose hand-written steps survive regeneration (named slots), ready-made
pipelines beyond CI (e2e, GitHub Pages, Azure Function), per-project stacks, a command that
checks a plugin keeps the protocol's promises, a new project that passes lint, test and build on
its first run, and a generated pipeline that installs each language's toolchain. They are
in the table; this page does not call them unique.

### Nearly only smith

Here smith is ✅ and no other tool is better than ◐. The ◐ holders are named.

| Row | Who is ◐, and why |
|---|---|
| Ruby, PHP | moon: partial and experimental (Ruby); project categorisation only (PHP). Earthly runs anything in containers |
| Dart and Flutter; Azure Bicep | Earthly only, as containers. No other tool documents language-aware support |
| Shared schemas link projects across languages | Bazel and Pants, through protobuf rules and code generation; Pants has an experimental OpenAPI backend |
| Team recipes | Nx and Turborepo (custom generators), moon (templates) |
| Explain a cache miss with one command that runs nothing | Documented procedures in Nx, Turborepo, Bazel (`--explain`), Gradle, moon and Lerna (through Nx Cloud) |
| See, trim, clear and verify the cache | Clear or purge commands in Nx, Bazel, Rush, moon, Cargo, Please and the package managers |
| PyPI, `.deb` and `.rpm` | Pants, in experimental backends |
| A failed release resumes | Nx, for npm |
| Health check; applies fixes; adoption dry run | See the notes; none previews diffs, asks about risky fixes or reports what blocks adoption |
| Ready-made environment: Python, Go, .NET, Ruby, PHP, Bicep | Nobody is ✅; Nx, Bazel, Pants, moon, Buck2, Please and Earthly are ◐ where they support the language at all: you write BUILD files, rules, `moon.yml` or a plugin |
| Out-of-process plugins in any language | moon (WASM, experimental), Cargo (`cargo-*` programs), Nx (plugins run in forked processes, but are Node packages) |
| Terminal interface | Nx and Turborepo, for task output |
| Native Windows | Pants (WSL 2), Earthly (experimental native) |

### What "one tool for the whole chain" is worth

The composite claim, "scaffold, run, cache, generate CI and release, across many
languages, in one tool", does not survive as "only smith". Nx is ✅ for scaffolding,
cached tasks, CI file generation and release, and is the only other tool ✅ for CI
generation. The difference is breadth: in the language rows smith is ✅ for 11 of 12 and
◐ for the twelfth (Gradle, Kotlin and Android were not run); Nx is ✅ for 4. In the
ready-made-environment rows, which ask for a generator and tasks with no rules to write, smith is
✅ for nine of eleven languages and ◐ for two (JVM and Dart), and Nx is ✅ for one (JavaScript and
TypeScript), with Gradle ✅ for the JVM and C++ and Cargo for Rust. Nx publishes to npm, crates.io and Docker registries;
smith to the registries in the table. That is a difference of degree, and the table shows
it row by row.

## What others do that MonoSmith does not do yet

<!-- rows:not-yet lang-jvm, env-jvm, find-bounds, graph-xlang, task-granular, task-file, cache-reapi, cache-rexec, cache-hermetic, cache-toolchain, cache-watch, ci-more, ci-audit, rel-crates, rel-maven, rel-image, rel-sbom, rel-prov, rel-oidc, dep-sync, dep-decl, plug-reg, plug-sec, ui-jb, ui-cloud -->
<!-- rows:no-rival env-dart, rel-tag, rel-assets, rel-preflight, rel-journal, health-other -->

Every row where at least one tool is ✅ and smith is ❌ or ◐. "How much it matters" is
the author's judgement for a typical team, not a fact; "deliberate" means a written
decision, not a gap.

### Deliberate non-goals

These are recorded in
[Execution and cache, section 9](../../architecture/execution-and-cache.md) and
[Why MonoSmith](../why-monosmith.md). If you need them, choose another tool.

| Capability | Who has it | Matters | Why smith does not |
|---|---|---|---|
| Remote or distributed task execution | Bazel, Pants, Buck2; Nx and Lerna through Nx Cloud; experimental or partial in Rush, moon, Please, Lage, Earthly | **High** for very large repositories, low for most | Tasks run on one machine and only results are shared |
| Hermetic or sandboxed execution | Bazel, Pants, Please, Earthly; partly Buck2, Lage | **High** where cache correctness is paramount | A task's declared inputs are trusted, which lets smith run `npm`, `dotnet` and `cargo` unmodified; wrong `inputs` give wrong cache hits. An opt-in strict mode that reports files read but not declared is proposed in #563 |
| Dependency graph of single files and targets | Bazel, Pants, Buck2, Please | **High** at very large scale, low below it | The unit is the project, not the file |
| Affected selection by task, finer than a project | Nx, Bazel, Pants, Buck2, moon, Gradle, Please | Medium: fewer tasks run in a large project | A design choice (project-level affected, everything when unsure); #574 asks whether to change it |
| Watch mode | Nx, Turborepo, Rush, Gradle, Lage, Please | Low to medium | smith recomputes from git each run; `smith dev` starts dev servers |
| Downloading and pinning toolchains | moon, Gradle; Rush and pnpm for their package manager | Medium for onboarding | [ADR 0020](../../architecture/decisions/0020-toolchain-pins-checked-not-downloaded.md): smith pins and checks versions, and writes `.tool-versions` for mise and asdf |
| A hosted service (remote cache, analytics, CI features) | Nx Cloud, Vercel Remote Cache, Gradle's Develocity, moon's hosted option | Medium: zero-ops sharing and analytics | The remote cache is a bucket or server you run; a reference cache server is proposed in #563 |

### Gaps

| Capability | Who has it | Matters | Status |
|---|---|---|---|
| Rules for which projects may depend on which | Nx, Bazel, Pants, Buck2, moon, Please; Turborepo (experimental) and Yarn (constraints) in part | **Medium to high**: architecture governance as the repository grows | #570 |
| JetBrains IDE plugin | Nx, Lerna, Bazel, Gradle, Please | **Medium to high**: many JVM and .NET teams use JetBrains | #303 |
| Speaking the Bazel Remote Execution API for caching | Bazel, Pants, Buck2, moon; Please (experimental) | Medium: reuse of existing cache servers | #345 asks first which servers already work over HTTP |
| A plugin SDK for C# | Nx, Pants, Bazel, Buck2, Gradle ship SDKs in their own languages | Medium: writing a plugin is the route to a missing language | Go, TypeScript and Python are published; C# is next: #564 |
| Sandboxed plugins | moon (WASM plugins get a sandboxed file system) | Medium to high for teams that run plugins from outside | smith trusts, pins and limits the environment of a plugin but does not sandbox it: #565, [decision 0036](../../architecture/decisions/0036-external-plugins-are-trusted-and-get-a-minimal-environment.md) |
| Import edges between languages in the graph | Nx, Bazel, Buck2, Please | Medium for polyglot repositories | Edges across languages come from shared schemas you name; by design, [decision 0028](../../architecture/decisions/0028-graph-edges-from-shared-schemas.md) |
| Gradle, Kotlin and Android environments verified end to end | Gradle (`gradle init`) | Medium for JVM teams | Maven was run on 1.65.0; Gradle was not run (not installed); the `jvm` plugin supports both |
| A plugin registry or catalogue | Nx; Bazel for rule sets | Medium; grows with third-party plugins | #164, #565 |
| SBOM and provenance beyond npm and Go apps, and beyond GitHub Actions | npm and pnpm generate SBOMs; Nx and pnpm document npm provenance | **High** for supply-chain reviews | #348 |
| Trusted publishing (no long-lived key) for more registries | npm (`npm trust`) | Medium to high | RubyGems, pub.dev and PyPI done; no issue for the rest |
| CircleCI, Bitbucket Pipelines or Jenkins pipelines | Nx (CircleCI, Bitbucket Pipelines; not Jenkins) | Medium | #571 |
| One version of each shared dependency across projects, in every language | Rush, Gradle, Cargo, pnpm and Yarn, Bazel | Medium | `smith sync` covers Node, Python, Rust, Go, .NET, the JVM, Dart, Ruby and PHP. #572 |
| Checking that a package declares what it imports | Nx, Rush, Bazel, Pants | Medium | smith checks Node packages only (`deps-check`). #572 |
| Dependency vulnerability audit | npm, pnpm, Yarn | Medium to high | Node, Python, Go, Rust, .NET, Ruby and PHP; the JVM and Dart have no audit tool to run, see #572 |
| Container images, Maven, crates.io, Go tags and Packagist releases verified end to end | Pants and Earthly (images); Gradle (Maven); Nx and Cargo (crates.io) | Medium | Shipped and covered by tests; not run against the live registry or a Docker daemon for this page |

### Gaps no other tool fills either

Rows where smith is ◐ or ❌ and no tool is ✅. They are shortfalls in smith, not
advantages of a rival: per-platform release archives for apps outside Go (#573); Go-module and
Packagist release by tag verified live; a registry preflight (#347, #380) and a release
journal (#347) across every registry; an importer for Lerna, Rush, Bazel or Gradle
workspaces (#553 is the nearest); and the Dart lint task, which did not pass on the test machine because `dart analyze` itself crashed there.

## Where we are equal

On the core loop smith is level with the established JavaScript-first tools: tasks in
dependency order, affected-by-git-range, a local content-addressed cache and a remote
cache you host. It matches Turborepo on signed remote cache objects and Nx on a
conversion path, an MCP server and a VS Code extension. Rush and moon also ship MCP servers. On
supply-chain basics (SBOM generation, provenance) npm and pnpm are already there. The
equal ground is where the choice stops being about features.

## Where we are behind

The gaps are in scale, ecosystem and maturity. Bazel, Pants and Buck2 do remote
execution, hermetic builds and file-level graphs that smith chose not to. Nx has the
larger plugin ecosystem, a registry, JetBrains support, task-level affected, a hosted
service and a track record that a tool whose first commit is dated 2026-10-15 does not.
Boundary rules between projects are an open gap with several tools ahead. Dependency
tooling outside Node and Python is thin, and several release paths are shipped but not
yet proven against the live registry.

## Limits of this page

- **Counting rows favours whoever wrote the rows.** The Score rows count only ✅. Every
  row is something smith does or has a plan for; rows a rival would add (a hosted
  dashboard, file-level graphs, IDE breadth, years of production use) are in the table
  only where smith is also measured on them. The ❔ row is there so the score is not
  read as a count of absences: it is a count of cells not verified.
- **❔ is not absent.** The documentation sources were searched, with the keywords the
  notes list, and a feature described under another name can be missed. About two
  thirds of the cells for the other tools are ❔, because most rows are features a tool's
  own documentation never names; that is also why their Score rows are low.
- **Documentation only.** A mark is what the tool's own documentation says on 2026-10-25,
  not what it does after plugins, scripts or a vendor's service. Nothing was measured and no
  benchmark is quoted.
- **smith is young.** Its first commit is dated 2026-10-15, with one main contributor. The
  [evaluator FAQ](../faq-for-evaluators.md) answers the maturity question.
- **Cells that could not be verified here:** a live publish to crates.io, Maven Central,
  Packagist or a container registry; Gradle, Kotlin and Android runs of the `jvm`
  plugin; the Azure Pipelines and GitLab CI renderers; `smith adopt` outside a scratch
  repository; a remote cache against a live bucket; the terminal interface in an interactive
  session; build provenance attestation checked with `gh attestation verify`.
- **Dropped.** Candidates for "only smith" that did not survive checking: a trust step before a repository-named
  plugin runs (moon's proto tool manager documents one, and Gradle's dependency verification covers plugins);
  writing extensions in any language (moon and Cargo); one process per extension (Nx's plugin isolation); a
  graph built from manifests across ecosystems without per-target build files (Nx); signed remote
  cache objects (Turborepo documents HMAC-SHA256 signing, and pnpm signs its pipeline cache);
  an MCP server (Nx, Rush and moon have one); conversion from Nx or Turborepo (moon
  ships migration extensions for both); SBOM generation (npm and pnpm have an `sbom`
  command); a cache-management command (several tools clear caches); `why-miss` as a
  concept (five tools document ways to debug a miss).

## Regenerating it

The table and the cell notes come from [`features.json`](features.json). To change a
cell, edit the data and run:

```sh
node docs/product/comparison/build-everything-compared.mjs          # rewrite the generated blocks
node docs/product/comparison/build-everything-compared.mjs --check  # exit 1 when the page is stale
node docs/product/comparison/build-everything-compared.mjs --report # print the derived row lists
```

`--check` also compares the row lists named in this page's `rows:` comments (only smith,
nearly only, not yet, gaps with no rival) with the lists derived from the data, so a
changed cell cannot leave a claim behind. Each row has a short mark string for the
fourteen tools in the order of the table's columns (`Y` ✅, `P` ◐, `U` ❔, `N` ❌, `-` not
applicable), smith's mark, evidence code and issue, an optional note and the keywords
searched for its ❔ cells.

See also: [Monorepo tools compared](README.md), [Choosing a monorepo tool](choosing.md),
[Why MonoSmith](../why-monosmith.md).
