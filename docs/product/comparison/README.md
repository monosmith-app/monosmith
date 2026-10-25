# Monorepo tools compared

A feature comparison of MonoSmith (`smith`) with fourteen other tools or tool
families, as of **2026-10**. It is meant to help you shortlist, not to declare a
winner. The [decision guide](choosing.md) turns it into advice, and each tool has
its own page.

## Method

- **Sources, read raw.** Every statement about a competitor comes from that tool's
  own repository and documentation, read on 2026-10-25. The documentation
  sources (Markdown, MDX, Markdoc, AsciiDoc or HTML files) were downloaded from each
  tool's repository and searched with the tool's own vocabulary, not read through
  a summarising fetch tool. The first revision of this comparison used a summarising
  tool; it dropped details, and this revision corrects it (see the end of this
  page). Licence, archive status and last push date come from the GitHub API on
  2026-10-25. Each tool page ends with its sources, the retrieval date and the short
  phrase each non-obvious mark relies on. Statements about smith come from its
  repository, its reference pages, `smith --help` and runs of smith 1.62.0 in scratch
  directories.
- **No measurements.** This page contains no benchmarks and no performance
  numbers. None were taken.
- **What a mark means.** A mark describes what the tool's own documentation says,
  not what the tool can do after you add plugins, scripts or a vendor's service.
- **What ❔ means now.** The documentation source was searched for the feature
  and its usual names, and it neither states the feature nor excludes it. The cell
  says what was searched. ❔ is not "absent", and it is not "unknown to me": it is
  "not stated in the sources searched". A ❌ is used only where a documentation
  page says the feature is not provided, or documents a different mechanism
  for the same job and a full-text search found nothing; the cell or the page
  quotes it.
- **CI generation is read narrowly.** ✅ means the tool writes pipeline files for
  you. ◐ means the documentation gives guides or recipes for running the tool in
  CI, but no generator was found.
- **Versions move.** Nx, Turborepo, pnpm, moon and others change quickly, and
  several features below are marked experimental or beta by their authors. Check
  the tool's current documentation before you decide.
- **The author is the smith project.** Read the page with that in mind. The
  [choosing guide](choosing.md) lists the cases where smith is the wrong tool.

### Legend

| Mark | Meaning |
|---|---|
| ✅ | The tool's documentation says it has the feature |
| ◐ | Partly, experimental, or with a condition; the note says which |
| ❔ | Not verified: searched in the tool's documentation source, and it neither states nor excludes the feature |
| ❌ | Checked, and the feature is not provided; the tool page quotes why |

### Status of each tool's repository

From the GitHub API on 2026-10-25. "Pushed" is the date of the last push to the
default branch.

| Tool | Licence file | Archived | Pushed | Note |
|---|---|---|---|---|
| Nx | MIT | no | 2026-10-23 | |
| Turborepo | MIT | no | 2026-10-23 | |
| Bazel | Apache-2.0 | no | 2026-10-25 | |
| Pants | Apache-2.0 | no | 2026-10-25 | |
| Lerna | MIT | no | 2026-10-25 | stewardship moved to Nx |
| Rush (Rush Stack) | MIT, stated in the repository's LICENSE | no | 2026-10-23 | the API reports no single licence for the monorepo |
| moon | MIT | no | 2026-10-24 | |
| Buck2 | MIT and Apache-2.0 (two licence files) | no | 2026-10-25 | the README says it has no stable release tag; releases on GitHub are flagged pre-release |
| Gradle | Apache-2.0 | no | 2026-10-25 | |
| Earthly | MPL-2.0 | no | 2026-10-23 | the README says active maintenance has ended; a fork, EarthBuild (MPL-2.0), was pushed on 2026-10-25 |
| Lage | MIT | no | 2026-10-24 | |
| npm CLI | Artistic-2.0 | no | 2026-10-23 | |
| pnpm | MIT | no | 2026-10-25 | |
| Yarn (berry) | BSD-2-Clause | no | 2026-10-23 | |
| Cargo | MIT or Apache-2.0 (two licence files) | no | 2026-10-25 | |
| Please | Apache-2.0 | no | 2026-10-24 | |
| MonoSmith | MIT | no | 2026-10-25 | first commit 2026-10-15 |

## The tools

| Tool | Page | Kind of tool |
|---|---|---|
| MonoSmith | (this repository) | a monorepo tool: tasks, cache, CI, release |
| Nx | [nx](nx.md) | a build system with plugins; Nx Cloud is its commercial service |
| Turborepo | [turborepo](turborepo.md) | a task runner and cache for JavaScript and TypeScript, with experimental Rust, Python and Go support |
| Bazel | [bazel](bazel.md) | a build and test tool with remote execution |
| Pants | [pants](pants.md) | a build system with dependency inference |
| Lerna | [lerna](lerna.md) | versioning and publishing for JavaScript packages |
| Rush | [rush](rush.md) | a build orchestrator for TypeScript monorepos |
| moon | [moon](moon.md) | a repository management and orchestration tool |
| Buck2 | [buck2](buck2.md) | a build system from Meta (pre-release) |
| Gradle | [gradle](gradle.md) | a build automation tool, with multi-project builds |
| Earthly | [earthly](earthly.md) | containerised builds (no longer actively maintained) |
| Lage | [lage](lage.md) | a task runner for JavaScript monorepos |
| npm, pnpm and Yarn workspaces | [workspaces](workspaces.md) | package-manager linking and script running; pnpm 12 adds a task runner |
| Cargo workspaces | [cargo](cargo.md) | Rust's package manager, for many crates |
| Please | [please](please.md) | a cross-language build system |

## Table 1: scope, licence and what you write

| Tool | Languages, as documented | Licence | What you write to get started |
|---|---|---|---|
| **MonoSmith** | Node (JS/TS), Python, Go, Rust, .NET, Ruby, PHP, C/C++ (CMake), Dart/Flutter, JVM, Bicep; others by plugin | MIT | `smith.toml` naming the workspace; projects and tasks are inferred from manifests |
| Nx | JS/TS first; plugins extend; README names Gradle, Maven, .NET and Go ecosystems | MIT (Nx Cloud is a commercial service with a free Hobby plan) | `npx nx init` reuses `package.json` scripts; plugins infer tasks |
| Turborepo | JavaScript and TypeScript; other toolchains run through `package.json` scripts (a guide shows .NET); Cargo, uv and Go workspaces are documented as experimental | MIT | one `turbo.json` plus `package.json` scripts |
| Bazel | Java, C++, Android, iOS, Go and others, through rules | Apache-2.0 | BUILD files in Starlark |
| Pants | Python, Go, Java, Scala, Kotlin, Shell and Docker; JavaScript in beta; code generators such as Protobuf | Apache-2.0 | BUILD files, which `pants tailor` generates; dependencies are inferred by static analysis |
| Lerna | JavaScript and TypeScript | MIT | `lerna init`; configuration in `lerna.json` and `nx.json` |
| Rush | JavaScript and Node.js | MIT (Rush Stack monorepo) | `rush init` writes config templates; projects are listed in `rush.json` |
| moon | Bun, Deno, Go, Node, Rust full; Python and Ruby partial; others limited | MIT | `moon init` creates a `.moon` folder with `.moon/workspace.*`; projects get a `moon.yml` |
| Buck2 | Prelude rules for C/C++, Erlang, Go, Haskell, Java, Kotlin, OCaml, Python and Rust; C#, Objective-C and Swift listed as unavailable | MIT and Apache-2.0 | BUCK files |
| Gradle | Java, Kotlin, Scala, Android, Groovy, C++, Swift | Apache-2.0 | `settings.gradle(.kts)` with `include(...)` and build scripts |
| Earthly | Any language, by running builds in containers | MPL-2.0 | an Earthfile |
| Lage | JavaScript monorepos | MIT | `npx lage init` writes `lage.config.js` with a `pipeline` |
| npm / pnpm / Yarn workspaces | JavaScript and TypeScript | npm CLI: Artistic-2.0; pnpm: MIT; Yarn: BSD-2-Clause | a `workspaces` field (npm, Yarn) or `pnpm-workspace.yaml` |
| Cargo workspaces | Rust | MIT or Apache-2.0 | a `[workspace]` table in `Cargo.toml` |
| Please | "a number of popular languages" (Go, Python, Java named) | Apache-2.0 | BUILD files; `plz init` writes `.plzconfig` and a `pleasew` wrapper |

## Table 2: running tasks

| Tool | Task graph | Affected only | Local cache | Remote cache | Remote execution |
|---|---|---|---|---|---|
| **MonoSmith** | ✅ in dependency order, from imports and shared schemas | ✅ `--affected`; everything when unsure | ✅ content-addressed | ✅ HTTP, S3, Azure Blob, signed | ❌ tasks run on one machine |
| Nx | ✅ project and task graphs | ✅ `nx affected`, selected by task inputs | ✅ | ✅ Nx Cloud; a self-hosting guide exists | ◐ distributed task execution on Nx Cloud |
| Turborepo | ✅ task ordering and parallelism | ✅ `--affected`, `--filter` by git range | ✅ `.turbo/cache` | ✅ Vercel Remote Cache by default; self-hostable | ❔ no mention in the documentation (searched "remote execution", "distributed") |
| Bazel | ✅ dependency analysis | ◐ `bazel query` has `rdeps`; no changed-files selection found | ✅ disk cache and action cache | ✅ | ✅ |
| Pants | ✅ dependency inference | ✅ `--changed-since` with `--changed-dependents` | ✅ | ✅ REAPI servers and cache-only servers | ✅ |
| Lerna | ✅ dependency order | ✅ `--since` | ✅ | ✅ "can be shared with teammates and CI" | ◐ distributed task execution through Nx Cloud |
| Rush | ✅ dependency order | ✅ incremental builds rebuild changed projects and their dependents, by file hashes against the last build (not a git range) | ✅ | ✅ Azure Blob, Amazon S3, plugins | ◐ "cobuilds" (experimental) share work between machines through the build cache |
| moon | ✅ | ✅ `moon run --affected` | ✅ | ✅ a self-hosted Bazel-REAPI-compatible server, or a hosted service | ◐ README lists "action distribution" and says not all features are supported; no documentation page found |
| Buck2 | ✅ single graph | ◐ `cquery`'s `rdeps` selects dependents; no changed-files command documented | ◐ artifacts in `buck-out`; the daemon keeps an in-memory cache | ✅ remote-execution action cache and CAS | ✅ "remote execution first" |
| Gradle | ✅ projects built before their dependants | ◐ up-to-date checks skip unchanged tasks; no git-range selection found | ✅ build cache | ✅ HTTP build cache | ❔ no mention in the user guide (searched "remote execution", "distributed build") |
| Earthly | ✅ targets form a build graph and run in parallel | ◐ "auto-skip" needs Earthly Cloud, which was shut down in 2025 | ✅ layer caching, cache mounts | ✅ `--remote-cache` (registry) | ◐ remote runners (BuildKit); the hosted Satellites were shut down |
| Lage | ✅ `pipeline` in `lage.config.js` | ✅ `--since` | ✅ | ✅ Azure Blob and your own storage | ◐ "a single-node build orchestrator"; a BuildXL integration distributes builds |
| npm workspaces | ◐ runs in the order of the root `package.json` | ❌ workspaces are selected by name or path only | ❌ `npm cache` holds downloaded packages | ❌ | ❌ |
| pnpm workspaces | ✅ `tasks.dependsOn` in `pnpm-workspace.yaml` (pnpm 11 and 12 documentation) | ✅ `--filter "[origin/main]"` | ◐ `pnpm pipeline` caches tasks that declare `outputs` (experimental, v12.4 and newer) | ◐ a Turborepo-API server or `pnpr`, from v12.12 (experimental) | ❔ no mention (searched) |
| Yarn workspaces | ✅ `foreach -t` topological | ✅ `--since` | ❌ the documented cache is for installed packages | ❌ | ❔ no mention (searched) |
| Cargo workspaces | ✅ builds follow the dependency graph | ❔ no changed-since selection found | ✅ the target directory; dep-info files record what to rebuild | ◐ third-party `sccache` is documented | ❔ no mention (searched) |
| Please | ✅ a build graph, with `plz query` | ✅ `plz query changes` against a revision | ✅ directory cache, on by default | ✅ HTTP cache | ◐ experimental, from v15 |

## Table 3: shipping and keeping it healthy

| Tool | CI pipeline generation | Release and versioning | Scaffolding | Dependency alignment | Health check | Editor or terminal UI | Plugin model |
|---|---|---|---|---|---|---|---|
| **MonoSmith** | ✅ GitHub Actions, Azure Pipelines, GitLab CI | ✅ conventional commits to npm, PyPI, crates.io, NuGet, RubyGems, pub.dev and Maven repositories; Go modules and Packagist by tag | ✅ `smith add`, recipes | ✅ `smith sync`, `smith up` | ✅ `smith doctor`, with fixes | ✅ terminal UI, VS Code extension, MCP | ✅ out-of-process JSON-RPC over stdio: "any executable that speaks the protocol"; SDKs for Go and TypeScript |
| Nx | ✅ `nx g ci-workflow` for GitHub, GitLab, CircleCI, Azure and Bitbucket Pipelines | ✅ `nx release`: npm, crates.io, Docker | ✅ generators | ◐ docs recommend a single version policy; `@nx/dependency-checks` lint rule | ◐ `nx report`; a Conformance plugin is an Nx Enterprise feature; no doctor command found | ✅ Nx Console (VS Code, JetBrains) | ✅ plugins |
| Turborepo | ◐ per-provider guides, no generator found | ◐ no turbo command; the guide uses Changesets | ✅ `turbo gen` | ◐ docs point to syncpack, manypkg, sherif or the package manager | ◐ `turbo boundaries` (experimental) checks workspace violations | ✅ Turborepo LSP (VS Code extension) and a terminal UI (`"ui": "tui"`) | ❔ none found (searched "plugin") |
| Bazel | ◐ the FAQ says CI integrates by exit code; no generator | ❔ no release command found | ❔ none found | ◐ Minimal Version Selection picks one version per module | ❔ none found | ◐ IntelliJ plugin (FAQ) | ✅ extension language |
| Pants | ◐ a CI page, an example workflow and an action; no generator | ◐ `pants publish` for Python distributions, Docker, Helm and Kubernetes (Python publish is experimental); version tags from git | ◐ `pants tailor` generates BUILD files, not project code | ◐ one lockfile per resolve | ◐ `pants tailor --check`; visibility rules validate dependencies | ❌ "We have not yet developed tight IDE integrations" | ✅ plugin API |
| Lerna | ◐ CI workflow examples for distributed execution; no generator found | ✅ `lerna version` (including `--conventional-commits`) and `lerna publish` | ✅ `lerna init`, `lerna create` | ❔ none found | ◐ `lerna repair` updates configuration after an upgrade; no repository check found | ✅ Nx Console (VS Code, JetBrains) | ❔ none found |
| Rush | ◐ a CI page and a bootstrap script; no generator | ✅ change files, version policies, `rush publish` | ◐ `rush init` writes config; adding a project is documented as manual steps | ✅ `rush check` | ◐ `rush check` and `rush scan` (phantom dependencies) | ◐ an MCP server; no editor extension found | ◐ plugins, which the plugin pages title "experimental" (cache providers are one use) |
| moon | ◐ `moon ci` runs affected tasks; no pipeline files generated | ❌ the FAQ says moon does not version or publish packages | ✅ code generation | ❔ none found (searched) | ◐ `moon sync projects` repairs configuration "to help achieve a healthy repository state"; no read-only check found | ✅ VS Code extension ("early stages") and an MCP server | ◐ WASM plugins and extensions, which the pages call "still experimental" |
| Buck2 | ❔ none found | ❔ none found | ◐ `buck2 init` creates a project | ❔ none found | ❔ none found | ◐ the Starlark library gives LSP and DAP bindings; no editor extension documented | ✅ rules are external to the binary, in Starlark |
| Gradle | ◐ recipes for six CI systems; no generator | ◐ `maven-publish` and other publishing plugins; versioning is yours | ✅ `gradle init` | ✅ platforms and version alignment; version catalogs | ❔ none found | ✅ IDE integrations (Android Studio, IntelliJ IDEA, Eclipse Buildship) | ✅ plugins |
| Earthly | ◐ CI integration guides | ❔ none found | ❔ none found | ❔ none found | ❔ none found | ◐ a VS Code extension for Earthfiles exists (its build is listed as an example) | ❔ none found |
| Lage | ◐ log reporters for GitHub Actions and Azure DevOps; no generator | ❔ none found | ◐ `npx lage init` writes the config only | ❔ none found | ❔ none found | ❔ none found | ❔ task types (`npmScript`, `worker`) only; no plugin API found |
| Workspaces (npm, pnpm, Yarn) | ◐ pnpm: `pnpm ci` is a clean install for CI; none generates pipelines | ◐ npm: `publish --workspaces`; pnpm: `pnpm version`, `pnpm change` (changesets format), `publish -r`; Yarn: an experimental release workflow | ◐ npm: `npm init -w`; pnpm and Yarn: none for workspaces found | ✅ pnpm catalogs; Yarn constraints and catalogs; npm: none found | ◐ `npm doctor` and `pnpm doctor` check the environment; Yarn constraints check the repository | ❔ none found | ◐ Yarn plugins; pnpm hooks (`.pnpmfile.mjs`); none found for npm |
| Cargo workspaces | ◐ the Cargo book has CI samples for GitHub Actions, GitLab CI, sr.ht and CircleCI | ◐ `cargo publish --workspace` publishes all members; versions are yours | ✅ `cargo new`, `cargo init` | ✅ members inherit dependencies with `workspace = true` | ❔ none found | ❔ none found | ✅ third-party `cargo-*` subcommands |
| Please | ◐ a codelab for GitHub Actions; no generator | ❔ none found | ◐ `plz init` writes `.plzconfig` and a wrapper | ❔ none found | ❔ none found | ✅ extensions for VS Code, IntelliJ, Neovim and a BUILD-file language server | ✅ plugins |

## Table 4: moving in, and what it costs

| Tool | Migration path in | Cost model |
|---|---|---|
| **MonoSmith** | `smith convert` from Nx, Turborepo or mnci; `smith adopt` (its help text names npm workspaces, Go modules and Python projects; a dry run in scratch directories also listed a Cargo workspace and Gradle modules) | free, open source; remote cache is a bucket or server you run |
| Nx | `npx nx init` in an existing npm, pnpm or Yarn workspace | open source core; Nx Cloud has a free Hobby plan and paid plans; self-hosted caching is documented as free |
| Turborepo | a documented guide for migrating from Nx; incremental adoption of `turbo.json` | open source; Vercel Remote Cache by default, or self-host |
| moon | `moon ext migrate-nx` and `moon ext migrate-turborepo` migrate those repositories | open source; the remote cache is a self-hosted REAPI-compatible server or a hosted service |
| Others | Pants has an incremental-adoption page; Rush documents adding projects one at a time; no migration guide found for the rest (see each page) | see each page |

Learning curve is not rated here: it is a judgement, not a documented fact. What
the documents do say is in the "What you write" column above. Bazel and Buck2 use
Starlark BUILD files, Turborepo needs one `turbo.json`, and smith needs a
`smith.toml` with a name in it.

## How to move between tools

Each tool page has a "Moving from it" section. The honest summary:

- **Automated into smith:** Nx, Turborepo and mnci, through `smith convert`
  ([Migration stories](../migration-stories.md)).
- **Documented by the tools themselves:** Nx adopts an existing npm, pnpm or Yarn
  workspace (`nx init`); Turborepo has a guide for migrating from Nx; moon ships
  extensions for migrating from Nx and Turborepo. No page read documents a path
  out to smith.
- **Everything else** is done by hand, by adopting the repository or by adding the
  projects one at a time.

## Pages

- Three deep dives on what smith does best, each with its own table of the fourteen tools:
  [plugins and flexibility](plugins-and-flexibility.md) (how each tool is extended, in what
  language, how safely), [languages and ready-made environments](languages-and-environments.md)
  (a scaffold, a workspace, tasks, CI, dev container and release for each language) and
  [the dependency graph](dependency-graph.md) (how it is built, what you can ask it, what it
  drives).
- [Everything MonoSmith does, against every tool compared](everything-compared.md): one
  table of every smith capability against the fourteen tools, what only smith documents,
  and what other tools do that smith does not yet do.
- [Choosing a monorepo tool](choosing.md): the decision guide.
- One page per tool: [Nx](nx.md), [Turborepo](turborepo.md), [Bazel](bazel.md),
  [Pants](pants.md), [Lerna](lerna.md), [Rush](rush.md), [moon](moon.md),
  [Buck2](buck2.md), [Gradle](gradle.md), [Earthly](earthly.md), [Lage](lage.md),
  [workspaces](workspaces.md), [Cargo](cargo.md), [Please](please.md).

## What changed in this revision

This revision (2026-10-25) replaces the first pass, which read the tools'
documentation through a summarising fetch tool.

**Cells.** The first revision's tables held 119 cells marked ❔. After reading the
raw sources, 85 of them hold a checked fact (22 ✅, 49 ◐, 8 ❌ and 6 that became
plain text, such as a licence or what you write) and 34 remain ❔. Per tool, ❔
cells before and after:

| Tool | Before | After |
|---|---|---|
| Nx | 3 | 0 |
| Turborepo | 6 | 2 |
| Bazel | 7 | 3 |
| Pants | 5 | 0 |
| Lerna | 8 | 2 |
| Rush | 5 | 0 |
| moon | 8 | 1 |
| Buck2 | 9 | 4 |
| Gradle | 8 | 2 |
| Earthly | 12 | 5 |
| Lage | 10 | 5 |
| npm, pnpm and Yarn workspaces | 17 | 3 |
| Cargo workspaces | 10 | 4 |
| Please | 10 | 3 |
| "Others" row of Table 4 | 1 | 0 (the row now names what was found) |

**Marks that moved.** Two ✅ were lowered, both for the tool's own experimental label: Rush's plugin
model (the pages on creating and using plugins are titled "experimental") and moon's
plugin model ("our WASM plugin implementations are still experimental"), each now ◐. Seven ◐ became ✅ once the raw page said
more: Rush's affected builds and `rush check`, Lage's `--since`, pnpm's task graph
(`tasks.dependsOn`), Cargo's build cache, Gradle's version alignment and the
workspaces' alignment cell (pnpm catalogs and Yarn constraints).

**Statements corrected.**

- Turborepo: the first revision said no language other than the npm ecosystem was
  mentioned. There is a multi-language guide (other toolchains through
  `package.json` scripts) and experimental Cargo, uv and Go workspace support.
- Pants: JavaScript is documented in beta.
- Please: the first revision said the page "does not say Windows is unsupported".
  Its FAQ says Windows is not supported natively.
- Rush: `rush check` ensures all dependencies are of the same version.
- Lerna: distributed execution through Nx Cloud and conventional-commit versions are
  documented.
- pnpm: the current documentation (pnpm 11 and 12) describes task dependencies, an
  experimental task cache, a shared cache from v12.12.0, `pnpm change`, `pnpm version`
  and `pnpm doctor`. The workspaces page no longer says pnpm has no task cache or
  release tool. These features are new and several are marked experimental.
- Buck2: the licence is both MIT and Apache-2.0 (the GitHub API reports one). The
  README still says there is no stable release tag; GitHub lists dated tags from
  2026-09-15 to 2026-10-15, each flagged pre-release.
- Earthly: the notice is confirmed, and the dates are now given: latest release
  2025-07-16, and the last push (2026-10-23) is the commit that made the notice
  prominent.
- Nx: pipeline files are generated (`nx g ci-workflow`), for five providers.
- moon: the FAQ says moon does not version or publish packages; migration
  extensions exist for Nx and Turborepo.
- Turborepo's environment-variable hashing: a task-level `env` also exists.

**Checks on smith's own cells.** Each smith cell was checked against the reference
pages, `smith --help` and runs of smith 1.62.0. All held except these, now fixed:

- "8 registries" is not a figure the documentation states. The cells now list what
  `docs/reference/releasing.md` and the plugin sources show: npm, PyPI, crates.io,
  NuGet, RubyGems, pub.dev and Maven repositories, with Go modules and Packagist
  released by tag.
- "From imports" in the task-graph cell now also names shared schemas, which the
  reference describes as graph edges.
- Table 4's `smith adopt` cell said npm workspaces, Go modules and Python projects,
  which is what `smith adopt --help` says. A scratch run also listed a Cargo workspace,
  Gradle modules and a Lerna repository. The cell now says the help text names three
  and that the others were seen once, in a dry run.
- "Any language" for plugins rests on the protocol ("A plugin is any executable that speaks
the protocol on stdin/stdout"); the only SDK in the repository is for Go, so the cell now
says so.
- The VS Code extension in the editor cell is published: the Visual Studio
  Marketplace lists MonoSmith 0.2.0 (updated 2026-10-24), and Open VSX had no entry on
  2026-10-25. The first revision did not check this.
- "`doctor` fails on drift" was tried: after editing a generated workflow,
  `smith doctor` and `smith ci generate --check` both exit non-zero.

**Still ❔, and why.** The 34 cells remaining are features whose names were searched
in the tool's documentation source and not found, where the documentation also does
not say the feature is absent: Turborepo's plugin model and remote execution, Bazel's
release, scaffolding and health checks, Buck2's CI, release, dependency alignment and
health check, Lerna's dependency alignment and plugins, Gradle's remote execution and
health check, Earthly's release, scaffolding, dependency alignment, health check and
plugins, Lage's release, dependency alignment, health check, editor and plugins,
Cargo's affected, remote execution, health check and editor, Please's release,
dependency alignment and health check, moon's dependency alignment, and a few
package-manager cells. A search of documentation can miss a feature described under
another name; the pages for these tools list what was searched.

**Limits of this revision.**

- Tool sites were not opened in a browser. The published URLs on the tool pages were
  derived from the source files' paths and not each fetched; the quoted source path is
  the reference. Nx Cloud's pricing page and the other sites' non-repository pages
  (blogs, marketing pages) were not read.
- A phrase "none found" means "none found in the documentation source on 2026-10-25".
  Commercial services, third-party plugins and community tools were not searched.
- Nothing here was measured. Experimental and beta labels are the tools' own.
