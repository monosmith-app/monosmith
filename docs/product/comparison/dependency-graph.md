# The dependency graph: one graph for every language, from the files you already have

**MonoSmith builds the dependency graph of a multi-language monorepo from your manifests, your imports and your shared schemas. You write no BUILD files and no dependency list.** Then it uses that one graph for everything: the order tasks run in, the cache key, what a change affects, the release order, and the answer to "what breaks if I change this file?".

This page shows the graph at work, with real output, and sets it against the fourteen other tools in the [comparison](README.md). It is one of three pages on what smith does best; the others are [plugins and flexibility](plugins-and-flexibility.md) and [languages and environments](languages-and-environments.md).

## Why it matters to a monorepo owner

The graph is the monorepo tool's model of your repository. Everything that has to be right depends on it: if an edge is missing, a change ships unchecked; if it is spurious, CI runs for nothing. So the questions are who writes the graph, how fine it is, and what you can ask it.

- **Who writes it.** In a tool built on declared dependencies, you do, in every package, in the tool's language. In smith the language plugins read it from `package.json`, `go.mod`, `Cargo.toml`, `*.csproj` and the imports in your source.
- **How fine it is.** Files and targets are exact and expensive to maintain. Projects are coarser and free.
- **What you can ask it.** What a change reaches. What changed in the graph since last week. Who may depend on whom.

## The graph in one workspace

The workspace for these runs has 13 projects in 11 languages. Two are Node, two are Go, and a Go library owns an OpenAPI file that a Python package, a Node app and a .NET library consume. smith 1.65.0, 2026-10-25, a scratch workspace.

```text
$ smith graph
api  app · go · apps/api
  └─ core
billing  lib · dotnet · libs/billing
  └─ core
core  lib · go · libs/core
engine  lib · rust · libs/engine
infra  app · bicep · apps/infra
invoice  lib · php · libs/invoice
ledger  lib · jvm · libs/ledger
native  lib · cpp · libs/native
sdk  package · python · packages/sdk
  └─ core
toolkit  package · dart · packages/toolkit
ui  lib · node · libs/ui
web  app · node · apps/web
  ├─ core
  └─ ui
widgets  package · ruby · packages/widgets
```

No file declares `api → core` or `web → ui`. The Go plugin read the import `example.com/acme/libs/core/greeting`, the Node plugin read `import { greet } from '@acme/ui'`, and the match was made on what `core` and `ui` export. The edges from `billing`, `sdk` and `web` to `core` come from the schema: three lines in `smith.toml` (`consumes_schemas = ["libs/core/openapi.yaml"]`) and no import.

### Ask it what a change reaches

Add an endpoint to the OpenAPI file and ask what that touches. Go, Python, Node and .NET projects share no import with each other:

```text
$ smith affected
changes since main
api
billing
core
sdk
web

$ smith graph --blast-radius libs/core/openapi.yaml
libs/core/openapi.yaml starts at core
  api                  via import of core
  billing              via schema of core
  sdk                  via schema of core
  web                  via schema of core

$ smith graph --blast-radius ui
ui starts at ui
  web                  via import of ui
```

`smith run test --affected` runs just those five projects' tasks. When smith cannot be sure (no base branch, a shallow clone, a changed lockfile or `smith.toml`), it affects everything: too wide costs minutes, too narrow reports green for code nothing checked.

For a deployment platform, `smith affected --json` also lists, for each app, the directories to watch (the app's own and those of everything it depends on, transitively), which is what Railway or Render watch paths and GitOps path filters want:

```text
$ smith affected --json          # excerpt
  "watch": {
    "api": [ "apps/api", "libs/core" ],
    "web": [ "apps/web", "libs/core", "libs/ui" ]
  }
```

### See it

```text
$ smith graph --focus web           # web, what it uses and what uses it
core  lib · go · libs/core
ui  lib · node · libs/ui
web *  app · node · apps/web
  ├─ core
  └─ ui

$ smith graph --affected            # only the affected projects; * marks the ones that changed
$ smith graph --format dot | dot -Tsvg > graph.svg
$ smith graph --format html -o graph.html    # one self-contained page, no script
$ smith graph --format tui          # browse it in the terminal
$ smith graph --format json         # every project with its kind, language, dependencies and tasks
```

The HTML file for the 13 projects is 5 KB of inline SVG with no script and no external request, so you can attach it to a pull request or open it offline. `smith projects --json` and `smith graph --format json` give scripts the same facts.

### See how it changed

Add a library and import it from the app, then ask what the graph did since `main`:

```text
$ smith graph --diff main
project graph from main to the working tree
+ pricing (node lib, libs/pricing)
+ web → pricing
```

`--diff` accepts a branch, tag or commit and reports projects and imports added and removed, and kinds and stacks changed, as text or JSON. It is the review comment that says "this pull request adds a dependency from web to pricing".

### It orders the work, across languages

A task that depends on another project's build waits for it. Running web's build and test pulls in the Go library that web depends on through the schema, and runs it first:

```text
$ smith run build test -p web
core:build ▶ go build ./...
core:build ✔ 109ms
web:build ✔ 941ms
web:test ✔ 1.327s
```

`smith explain web:test` shows `dependsOn ^build (build of the projects it imports, first)`. The same graph sets the cache key (a task's key includes the sources of the projects it depends on), the order of a release (packages release in dependency order) and the affected set.

### It refuses a cycle, and names it

```text
$ smith graph
smith: projects import each other in a cycle: ui → web → ui
```

The command stops with that error until one import is removed, and the [architecture notes](../../architecture/project-discovery.md) say every other command does too, so a cycle cannot reach CI unnoticed.

## Compared with fourteen tools

**MonoSmith and Nx are the two tools in this set that build a multi-ecosystem graph from manifests and imports with no per-target build files; smith ships all eleven ecosystems built in, and Nx adds most of them as plugins. Only smith documents a graph diff between two revisions.** The graph is at project level, which is the trade: Bazel, Pants, Buck2 and Please model targets and files.

The marks follow the [legend](README.md#legend). These rows are also in the [master table](everything-compared.md).

| Row | **MonoSmith** | Nx | Turbo | Bazel | Pants | Lerna | Rush | moon | Buck2 | Gradle | Earthly | Lage | JS pkg mgrs | Cargo | Please |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| Projects found from the manifests you already have | **✅** | ✅ | ◐ | ❌ | ◐ | ◐ | ◐ | ◐ | ❌ | ◐ | ❌ | ◐ | ✅ | ✅ | ❌ |
| Edges read from source imports, in several languages | **✅** | ◐ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ◐ |
| Edges between projects of different languages | **◐** | ✅ | ❔ | ✅ | ◐ | ❌ | ❌ | ◐ | ✅ | ❔ | ◐ | ❌ | ❌ | ❌ | ✅ |
| Edges from shared schemas (OpenAPI, protobuf, GraphQL, AsyncAPI) | **✅** | ❔ | ❔ | ◐ | ◐ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| A plugin can contribute edges for its language | **✅** | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Graph of single files and targets | **❌** | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ |
| Views: tree, JSON, DOT, HTML, interactive, focus, affected | **✅** | ✅ | ✅ | ◐ | ❔ | ◐ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ |
| The graph as JSON | **✅** | ✅ | ◐ | ◐ | ✅ | ✅ | ◐ | ✅ | ◐ | ❔ | ❔ | ❔ | ◐ | ✅ | ✅ |
| Ask what depends on a project or file | **✅** | ✅ | ◐ | ✅ | ✅ | ❔ | ◐ | ◐ | ◐ | ❔ | ❔ | ◐ | ❔ | ◐ | ✅ |
| Graph diff between two revisions | **✅** | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Rules for who may depend on whom | **❌** | ✅ | ◐ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ◐ | ❔ | ✅ |

### How each tool builds and shows its graph

| Tool | Edges come from | Granularity | You look at it with |
|---|---|---|---|
| **MonoSmith** | language plugins read manifests, source imports and shared schemas; `smith.toml` overrides | project | `smith graph`: tree, JSON, DOT, HTML, terminal browser; `--focus`, `--affected`, `--diff`, `--blast-radius` |
| Nx | plugins read manifests; JS and TS imports are parsed; a plugin can do the same for its language | project; a separate task graph | `nx graph` (interactive, focus, which files created an edge, PNG), `--file=output.json`, `--graph` for tasks, `nx graph --affected` |
| Turborepo | "created by your package manager"; task relations in `turbo.json` | package, and a task graph | `--graph` to `svg`, `html`, `mermaid` or `dot`; `turbo devtools` in the browser; `turbo query` (GraphQL) |
| Bazel | declared in BUILD files; "declared dependencies" | target (and file inputs) | `bazel query`, `cquery`, `aquery`; `--output graph` for Graphviz; `rdeps`, `allpaths`, `somepath` |
| Pants | inferred: "will read your import statements and map those imports" | target, down to files | `pants dependencies`, `dependents`, `--format=json`, `peek`, `filedeps` |
| Lerna | the Nx project graph | package | `lerna ls --graph` (JSON adjacency list); `nx graph` |
| Rush | `package.json` dependencies of the projects listed in `rush.json` | project | `rush list --json`; selectors such as `--impacted-by` |
| moon | explicit `dependsOn` plus "implicit dependencies" found per language | project, and an action graph | `moon project-graph` (visual), `--dot`, `--json`, focus on a project and its dependents |
| Buck2 | declared in BUCK files | target | `buck2 uquery`, `cquery`, `aquery`; BXL scripts over the graph |
| Gradle | declared project dependencies; a task graph built before tasks run | project, and tasks | `gradlew :app:dependencies` (resolved dependencies); Build Scans with `--scan` |
| Earthly | targets that reference other targets | target | no graph view found |
| Lage | package dependencies; a target graph from `package.json` scripts | package | no graph view found; `--scope`, `--since` select |
| npm, pnpm, Yarn | the `package.json` files of the workspaces | package | `npm ls`, `pnpm list --recursive --json`, `yarn workspaces foreach --topological` |
| Cargo | `Cargo.toml` | crate | `cargo tree`, `cargo metadata` (JSON) |
| Please | declared in BUILD files | target | `plz query graph` (JSON), `deps`, `reverseDeps`, `somepath`, `changes`, `whatinputs` |

## What this costs you with other tools

- **Bazel, Buck2 and Please** want a BUILD or BUCK file per package, with `deps` listed by hand. That is what makes their graph exact, and it is also a file to write, review and keep in step with `go.mod` or `package.json` in every package.
- **Pants** removes most of that for supported languages with import inference, and still needs the BUILD files that `pants tailor` generates. Its graph is strongest for Python, Go and the JVM.
- **Nx** infers edges well, and parses source imports for JavaScript and TypeScript. For other languages its documentation says edges come from manifests, and a plugin can read imports. Each language is a plugin to install.
- **Turborepo, Lerna, Rush, Lage and the package managers** see what the JavaScript package manager sees. Your Go service and Python library are not in that graph.
- **moon** asks for a `moon.yml` per project and discovers implicit edges per language.
- **None of them** diffs the graph between two revisions in its documentation: you notice a new dependency in code review or you do not.

## Where another tool is the better choice

- **You need a file-level or target-level graph.** Bazel, Pants, Buck2 and Please model individual targets and files, so a change to one file selects only the targets that read it. smith's unit is the project: a change selects every project that owns the file and every project that depends on one (a deliberate choice, recorded under [non-goals](everything-compared.md#deliberate-non-goals)).
- **You need the graph to be provably complete.** In Bazel and Buck2 a build can only read what it declared, because it is sandboxed; the graph is the truth. smith's edges are as good as what the plugins read, and it trusts a task's declared inputs.
- **You need rules on who may depend on whom.** Nx (module boundaries), moon (constraints), Bazel, Buck2, Pants and Please (visibility) enforce them. smith has none yet ([#570](https://github.com/monosmith-app/monosmith/issues/570)).
- **You want to click through an interactive graph, or see which file created an edge.** Nx's `nx graph` and moon's visualizer do. smith's HTML view is static; its terminal view browses; `--blast-radius` says whether a step is an import or a schema, and not which file.
- **You want import edges between languages.** smith's import edges stay inside one language. Across languages the edges are shared schemas that you name in `smith.toml`. Nx, Bazel, Buck2 and Please document cross-language edges in a single graph.
- **Your language has no plugin.** An edge needs a language plugin that reports imports and exports, or a schema; you can [write one](plugins-and-flexibility.md).
- **You are inside one ecosystem with an excellent native graph.** `cargo tree` and `cargo metadata` for Rust, `gradle dependencies` for the JVM.

## Try it

```sh
smith adopt --dry-run                  # the projects and edges smith would see in your repository
smith graph --format tui               # browse the graph
smith graph --blast-radius <file>      # what a change to this file reaches
smith graph --diff main                # what this branch did to the graph
```

Nothing in `smith adopt --dry-run` changes the repository. Coming from Nx or Turborepo, `smith convert --dry-run` reads your workspace through its own project graph and reports what maps; see [Migration stories](../migration-stories.md). How edges are found, language by language, is documented in [The project graph and how edges are found](../../architecture/project-graph.md), and the schema edges in [decision 0028](../../architecture/decisions/0028-graph-edges-from-shared-schemas.md).

## Method and evidence

- **MonoSmith.** Every command and output on this page was run on 2026-10-25 with smith 1.65.0 in a throwaway workspace (not this repository's own), with two commits on `main` and a branch for the changes. `smith graph --format tui` was not run (it needs a terminal session). The cycle was created by importing `@acme/web` from the `ui` library and removed afterwards.
- **Other tools.** Read raw from each tool's own documentation source on 2026-10-25 and searched for "graph", "query", "dependents", "visibility", "cycle" and the tool's own vocabulary; a quoted phrase is the text relied on. ❔ means searched and not stated. A search of all fourteen sources for a graph comparison between revisions found nothing; that is a documentation search, and third-party tools (for example for Bazel) were not searched.
- **Not measured.** Nothing about graph size or speed was measured.

### Sources

Retrieved 2026-10-25.

- Nx (`nrwl/nx`, `astro-docs/src/content/docs`): `features/explore-graph.mdoc` ("Click on any dependency line to find which file(s) created the dependency."; `nx graph --focus`; `nx graph --file=output.json`); `features/multi-language-support.mdoc` ("they connect cross-language boundaries too"; "parses source files and creates dependencies from `import` statements"); `features/CI Features/affected.mdoc` ("pass `--files` to any affected command"; "`nx graph --affected` always shows projects"); `kb/add-language-support.mdoc` (`createDependencies`).
- Turborepo (`vercel/turborepo`, `apps/docs/content/docs`): `core-concepts/package-and-task-graph.mdx` ("created by your package manager"); `reference/run.mdx` ("`svg`, `html`, `mermaid`, or `dot`"); `reference/devtools.mdx` ("Visualize your monorepo's package graph in the browser."); `reference/query.mdx` ("Run GraphQL queries against your monorepo."); `reference/boundaries.mdx` (experimental).
- Bazel (`bazelbuild/bazel`, `docs`): `concepts/dependencies.mdx` ("declared dependency"); `query/guide.mdx` and `query/language.mdx` (`--output graph`, `allpaths`, `rdeps`, `allrdeps`, `somepath`); `concepts/visibility.mdx`.
- Pants (`pantsbuild/pants`, `docs/docs`): `using-pants/key-concepts/targets-and-build-files.mdx` ("Pants will read your import statements and map those imports back to your first-party code"); `using-pants/project-introspection.mdx` (`dependencies`, `dependents --transitive`, `--format=json`, `peek`, `filedeps`); `using-pants/validating-dependencies.mdx` ("Visibility rules are the mechanism by which to control who may depend on whom."); `writing-plugins/overview.mdx` ("Add new forms of dependency inference").
- Lerna (`lerna/lerna`): `libs/commands/list/README.md` ("Show dependency graph as a JSON-formatted adjacency list."); `website/docs/features/project-graph.md`.
- Rush (`microsoft/rushstack-websites`, `websites/rushjs.io/docs/pages`): `commands/rush_list.md` (`--json`); `commands/rush_build.md` (`--impacted-by`); `configs/rush_json.md` (`decoupledLocalDependencies`, the replacement name for `cyclicDependencyProjects`).
- moon (`moonrepo/moon`, `website/docs`): `commands/project-graph.mdx` ("generate and serve a visual graph of all configured projects"; `--dot`, `--json`, `--dependents`); `concepts/project.mdx` ("implicitly discovered by moon"); `config/workspace.mdx` (`constraints`).
- Buck2 (`facebook/buck2`, `docs`): `concepts/buck_query_language.md` (`buck2 uquery`, `cquery`, `aquery`); `concepts/target_pattern.md` ("specify all dependencies explicitly"); `bxl/index.md`; `concepts/visibility.md`.
- Gradle (`gradle/gradle`, `platforms/documentation/docs/src/docs/userguide`): `fundamentals/authoring-builds/build_lifecycle_intermediate.adoc` ("Gradle constructs the task graph by understanding the dependencies between tasks."); `fundamentals/running-builds/build_scans.adoc` (`--scan`); `fundamentals/running-builds/dependency_management_basics.adoc` (`:app:dependencies`).
- Lage (`microsoft/lage`, `docs/docs/reference/cli.md`): `--scope`, `--since`, "target graph". Earthly (`earthly/earthly`, `docs/caching/caching-in-earthfiles.md`): targets "inheriting from other targets, copying artifacts". Searched, not found: a graph view in Earthly and Lage.
- npm, pnpm, Yarn: `pnpm/pnpm.io` `docs/cli/list.md` (`--recursive`, `--json`, `--depth`); `npm/cli` `npm ls`; Yarn `workspaces foreach`.
- Cargo (`rust-lang/cargo`, `doc/book/src`): `commands/cargo-tree.md` ("Display a tree visualization of a dependency graph"; `--invert`); `reference/external-tools.md` (`cargo metadata` "in JSON").
- Please (`thought-machine/please`, `docs/commands.html`): `plz query` subcommands `graph` ("Prints a JSON representation of the build graph."), `deps`, `reverseDeps`, `somepath`, `changes`, `whatinputs`.
- MonoSmith: `docs/architecture/project-graph.md`, `docs/reference/tasks-and-cache.md`, `docs/manual/graph.md`, `docs/architecture/decisions/0028-*.md`, and the runs above.
