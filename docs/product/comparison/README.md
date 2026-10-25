# Monorepo tools compared

A feature comparison of MonoSmith (`smith`) with fourteen other tools or tool
families, as of **2026-10**. It is meant to help you shortlist, not to declare a
winner. The [decision guide](choosing.md) turns it into advice, and each tool has
its own page.

## Method

- **Sources.** Every statement about a competitor comes from that tool's official
  documentation or its repository, read between 2026-10-24 and 2026-10-25. Each
  tool page ends with its sources and retrieval date. Statements about smith come
  from its repository, its reference pages, `smith --help` and a run of smith 1.62.0
  in a scratch directory.
- **No measurements.** This page contains no benchmarks and no performance
  numbers. None were taken.
- **What a mark means.** A mark describes what the pages I read say, not what the
  tool can do after you add plugins, scripts or a vendor's service.
- **Unverified is not absent.** The pages were read through a fetch tool that
  summarises them, and a page that did not mention a feature leaves the cell as
  unverified. Many cells are unverified for that reason. Correct them through a
  pull request to this folder.
- **Versions move.** Nx, Turborepo, moon and others change quickly. Check the
  tool's current documentation before you decide.
- **The author is the smith project.** Read the page with that in mind. The
  [choosing guide](choosing.md) lists the cases where smith is the wrong tool.

### Legend

| Mark | Meaning |
|---|---|
| ✅ | The tool's documentation says it has the feature |
| ◐ | Partly, or with a condition; the note says which |
| ❔ | Not verified: the pages I read do not say either way |
| ❌ | Checked, and the feature is not provided |

## The tools

| Tool | Page | Kind of tool |
|---|---|---|
| MonoSmith | (this repository) | a monorepo tool: tasks, cache, CI, release |
| Nx | [nx](nx.md) | a build system with plugins; Nx Cloud is its commercial service |
| Turborepo | [turborepo](turborepo.md) | a task runner and cache for JavaScript and TypeScript |
| Bazel | [bazel](bazel.md) | a build and test tool with remote execution |
| Pants | [pants](pants.md) | a build system with dependency inference |
| Lerna | [lerna](lerna.md) | versioning and publishing for JavaScript packages |
| Rush | [rush](rush.md) | a build orchestrator for TypeScript monorepos |
| moon | [moon](moon.md) | a repository management and orchestration tool |
| Buck2 | [buck2](buck2.md) | a build system from Meta (pre-release) |
| Gradle | [gradle](gradle.md) | a build automation tool, with multi-project builds |
| Earthly | [earthly](earthly.md) | containerised builds (no longer actively maintained) |
| Lage | [lage](lage.md) | a task runner for JavaScript monorepos |
| npm, pnpm and Yarn workspaces | [workspaces](workspaces.md) | package-manager linking and script running |
| Cargo workspaces | [cargo](cargo.md) | Rust's package manager, for many crates |
| Please | [please](please.md) | a cross-language build system |

## Table 1: scope, licence and what you write

| Tool | Languages, as documented | Licence | What you write to get started |
|---|---|---|---|
| **MonoSmith** | Node (JS/TS), Python, Go, Rust, .NET, Ruby, PHP, C/C++ (CMake), Dart/Flutter, JVM, Bicep; others by plugin | MIT | `smith.toml` naming the workspace; projects and tasks are inferred from manifests |
| Nx | JS/TS first; plugins extend; README names Gradle, Maven, .NET and Go ecosystems | MIT (Nx Cloud is a commercial service with a free tier) | `npx nx init` reuses `package.json` scripts; plugins infer tasks |
| Turborepo | The npm ecosystem (npm, yarn, pnpm); no other language is mentioned | MIT | one `turbo.json` plus `package.json` scripts |
| Bazel | Java, C++, Android, iOS, Go and others, through rules | Apache-2.0 | BUILD files in Starlark |
| Pants | Python, Go, Java, Scala, Kotlin, Shell; plus generators such as Protobuf | Apache-2.0 | BUILD files; dependencies are inferred by static analysis |
| Lerna | JavaScript and TypeScript | MIT | ❔ |
| Rush | JavaScript and Node.js | MIT (rushstack monorepo) | ❔ |
| moon | Bun, Deno, Go, Node, Rust full; Python and Ruby partial; others limited | MIT | ❔ |
| Buck2 | Multiple, through rules in Starlark (C++, Python, Rust, Kotlin and others at Meta) | MIT or Apache-2.0 | BUCK files |
| Gradle | Java, Kotlin, Scala, Android, Groovy, C++, Swift | Apache-2.0 | `settings.gradle(.kts)` with `include(...)` and build scripts |
| Earthly | Any language, by running builds in containers | MPL-2.0 | an Earthfile |
| Lage | JavaScript monorepos | MIT | ❔ |
| npm / pnpm / Yarn workspaces | JavaScript and TypeScript | ❔ (not read) | a `workspaces` field (npm, Yarn) or `pnpm-workspace.yaml` |
| Cargo workspaces | Rust | MIT or Apache-2.0 | a `[workspace]` table in `Cargo.toml` |
| Please | "a number of popular languages" (Go, Python, Java named) | Apache-2.0 | BUILD files |

## Table 2: running tasks

| Tool | Task graph | Affected only | Local cache | Remote cache | Remote execution |
|---|---|---|---|---|---|
| **MonoSmith** | ✅ in dependency order, from imports | ✅ `--affected`; everything when unsure | ✅ content-addressed | ✅ HTTP, S3, Azure Blob, signed | ❌ tasks run on one machine |
| Nx | ✅ project and task graphs | ✅ `nx affected`, selected by task inputs | ✅ | ✅ Nx Cloud; a self-hosting guide exists | ◐ distributed task execution on Nx Cloud |
| Turborepo | ✅ task ordering and parallelism | ✅ `--affected`, `--filter` by git range | ✅ `.turbo/cache` | ✅ Vercel Remote Cache by default; self-hostable | ❔ |
| Bazel | ✅ dependency analysis | ❔ | ✅ | ✅ | ✅ |
| Pants | ✅ dependency inference | ✅ `--changed-since` with `--changed-dependents` | ✅ | ✅ REAPI servers and cache-only servers | ✅ |
| Lerna | ✅ dependency order | ✅ `--since` | ✅ | ✅ "can be shared with teammates and CI" | ❔ |
| Rush | ✅ dependency order | ◐ `rush rebuild --from` / `--to` select by dependency | ✅ | ✅ Azure Blob, Amazon S3, plugins | ❔ |
| moon | ✅ | ✅ `moon run --affected` | ✅ | ✅ | ❔ (README lists action distribution, not all features are supported) |
| Buck2 | ✅ single graph | ❔ | ❔ | ❔ | ✅ "remote execution first" |
| Gradle | ✅ projects built before their dependants | ❔ | ✅ build cache | ✅ HTTP build cache | ❔ |
| Earthly | ❔ | ❔ | ❔ | ❔ | ❔ |
| Lage | ❔ | ◐ "skips work based on your changes" | ✅ | ✅ your own cloud storage | ❔ |
| npm workspaces | ◐ runs in the order of the root `package.json` | ❔ | ❔ | ❔ | ❔ |
| pnpm workspaces | ◐ recursive runs; order is not guaranteed with cycles | ❔ | ❔ | ❔ | ❔ |
| Yarn workspaces | ✅ `foreach -t` topological | ✅ `--since` | ❔ | ❔ | ❔ |
| Cargo workspaces | ❔ | ❔ | ◐ shared `target/` directory | ❔ | ❔ |
| Please | ❔ | ❔ | ✅ caches by hash of each rule's inputs | ✅ shared remote caches | ❔ |

## Table 3: shipping and keeping it healthy

| Tool | CI pipeline generation | Release and versioning | Scaffolding | Dependency alignment | Health check | Editor or terminal UI | Plugin model |
|---|---|---|---|---|---|---|---|
| **MonoSmith** | ✅ GitHub Actions, Azure Pipelines, GitLab CI | ✅ conventional commits to 8 registries | ✅ `smith add`, recipes | ✅ `smith sync`, `smith up` | ✅ `smith doctor`, with fixes | ✅ terminal UI, VS Code extension, MCP | ✅ out-of-process JSON-RPC, any language |
| Nx | ❔ | ✅ `nx release`: npm, crates.io, Docker | ✅ generators | ❔ | ❔ | ✅ Nx Console (VS Code, JetBrains) | ✅ plugins |
| Turborepo | ◐ per-provider guides | ❔ | ✅ `turbo gen` | ❔ | ❔ | ❔ | ❔ |
| Bazel | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ extension language |
| Pants | ❔ | ◐ packagers listed; publishing not mentioned on the page read | ❔ | ❔ | ❔ | ❔ | ✅ plugin API |
| Lerna | ❔ | ✅ versioning and npm publishing | ❔ | ❔ | ❔ | ❔ | ❔ |
| Rush | ❔ | ✅ change files, bump, `npm publish` | ❔ | ◐ policies can enforce consistent versions | ◐ `rush check` exists; what it does was not read | ❔ | ✅ plugins (cache providers) |
| moon | ❔ | ❔ | ✅ code generation | ❔ | ❔ | ❔ | ❔ |
| Buck2 | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ user-implementable languages |
| Gradle | ❔ | ❔ | ❔ | ◐ version catalogs centralise versions | ❔ | ❔ | ❔ |
| Earthly | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Lage | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| Workspaces (npm, pnpm, Yarn) | ❔ | ◐ Yarn's docs show publishing all packages with `foreach` | ❔ | ◐ pnpm and Yarn `workspace:` ranges link local packages | ❔ | ❔ | ❔ |
| Cargo workspaces | ❔ | ❔ | ❔ | ✅ members inherit dependencies with `workspace = true` | ❔ | ❔ | ❔ |
| Please | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |

## Table 4: moving in, and what it costs

| Tool | Migration path in | Cost model |
|---|---|---|
| **MonoSmith** | `smith convert` from Nx, Turborepo or mnci; `smith adopt` for npm workspaces, Go modules, Python projects | free, open source; remote cache is a bucket or server you run |
| Nx | `npx nx init` in an existing npm, pnpm or Yarn workspace | open source core; Nx Cloud has a free tier and paid plans (see Nx's pricing page) |
| Turborepo | a documented guide for migrating from Nx; incremental adoption of `turbo.json` | open source; Vercel Remote Cache by default, or self-host |
| Others | ❔ | see each page |

Learning curve is not rated here: it is a judgement, not a documented fact. What
the documents do say is in the "What you write" column above. Bazel and Buck2 use
Starlark BUILD files, Turborepo needs one `turbo.json`, and smith needs a
`smith.toml` with a name in it.

## How to move between tools

Each tool page has a "Moving from it" section. The honest summary:

- **Automated into smith:** Nx, Turborepo and mnci, through `smith convert`
  ([Migration stories](../migration-stories.md)).
- **Documented by the tools themselves:** Nx adopts an existing npm, pnpm or Yarn
  workspace (`nx init`); Turborepo has a guide for migrating from Nx. No page read
  documents a path out to smith.
- **Everything else** is done by hand, by adopting the repository or by adding the
  projects one at a time.

## Pages

- [Choosing a monorepo tool](choosing.md): the decision guide.
- One page per tool: [Nx](nx.md), [Turborepo](turborepo.md), [Bazel](bazel.md),
  [Pants](pants.md), [Lerna](lerna.md), [Rush](rush.md), [moon](moon.md),
  [Buck2](buck2.md), [Gradle](gradle.md), [Earthly](earthly.md), [Lage](lage.md),
  [workspaces](workspaces.md), [Cargo](cargo.md), [Please](please.md).
