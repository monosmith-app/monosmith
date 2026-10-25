# Nx compared with MonoSmith

Nx is a build system for JavaScript and TypeScript monorepos with a plugin
system. It is the tool smith's predecessor mnci was built on, and the one smith
converts from most completely. This page is neutral by intent; the tables in
[Monorepo tools compared](README.md) hold the side-by-side view.

## What Nx is best at

- **Task-level affected.** `nx affected` selects each task whose declared inputs
  include a changed file, and tasks that read the outputs of a selected task. It
  works at the task level, not the project level.
- **A hosted service around the cache.** Nx Cloud provides remote caching,
  distributed task execution, end-to-end test splitting and a "self-healing CI"
  feature, in which an agent proposes fixes for CI failures. Nx Cloud has a free
  tier and paid plans.
- **Release management.** `nx release` versions projects, writes changelogs and
  publishes to npm, crates.io and Docker registries. It can be driven by
  conventional commits, version plan files or manual versions.
- **Plugins and generators.** Plugins infer tasks, configure cache inputs and
  outputs and scaffold code. Nx Console puts generators and the project and task
  graph in VS Code and JetBrains IDEs.
- **Incremental adoption.** `npx nx init` in an existing npm, pnpm or Yarn
  workspace reuses the `package.json` scripts, caches their outputs and runs only
  affected tasks.
- **Open source.** The repository is MIT licensed.

## Where MonoSmith differs

| | Nx | MonoSmith |
|---|---|---|
| Centre of gravity | JS and TS; plugins reach other ecosystems (the README names Gradle, Maven, .NET and Go) | Eleven languages built in as plugins; the core knows none of them |
| Affected | task-level, from declared inputs | project-level: the projects owning a changed file and everything depending on them; everything when smith cannot be sure |
| Remote cache | Nx Cloud, or a self-hosted server built to Nx's OpenAPI specification | a server, S3 bucket or Azure Blob container you run, objects signed with HMAC-SHA256; no hosted service |
| Distributed execution | Nx Cloud | none: tasks run on one machine and only results are shared |
| CI pipelines | `nx g @nx/workspace:ci-workflow --ci=<provider>` generates a starting workflow for GitHub, GitLab, CircleCI, Azure Pipelines and Bitbucket Pipelines (the documentation says the generator does not support Jenkins); Nx Cloud adds CI features | smith generates GitHub Actions, Azure Pipelines and GitLab CI and `doctor` fails on drift (tried: an edited generated workflow made `smith doctor` and `smith ci generate --check` exit 1) |
| Dependency alignment | the documentation recommends a single version policy and points to catalogs, `syncpack` or the `@nx/dependency-checks` lint rule, which compares a project's `package.json` with what it imports | `smith sync` aligns a shared dependency, and `smith doctor` checks one version per production dependency |
| Health check | `nx report` prints the versions of Nx and its plugins; a Conformance plugin that enforces workspace rules is described as an Nx Enterprise feature; no doctor command was found | `smith doctor`, with `--fix` |
| Implementation | built with Rust and extensible via TypeScript, per its README | one static Go binary |
| Editors | Nx Console for VS Code and JetBrains, "built and maintained by the Nx team" | a VS Code extension, a terminal interface and an MCP server; no JetBrains plugin was found in the repository |
| Repository status | MIT, not archived, last push 2026-10-23 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Nx instead

- Your repository is mostly JavaScript or TypeScript and you want the largest plugin
  ecosystem for it.
- You want **distributed task execution**, test splitting or Nx Cloud's CI
  features. smith has none of these.
- You use JetBrains IDEs and want the editor integration.
- You want task-level affected selection.
- You need a tool with a long track record. smith's repository history is ten days old.

## When MonoSmith is the better fit

- The repository has several languages and you want the same commands, cache and
  pipeline for all of them.
- You want release and CI generation in the same tool, for registries beyond npm and
  crates.io (PyPI, NuGet, RubyGems, pub.dev, Maven, Go module tags).
- You do not want a hosted service in the loop for caching.

## Moving from Nx to MonoSmith

`smith convert` reads the workspace through Nx's own project graph, maps
projects, kinds, targets, release tags and CI, removes `nx.json`, `.nx/`,
`project.json` files and the `nx` dependencies, and reports what it could not map.
Start with `npm ci` and `smith convert --dry-run`. Details and limits are in
[Migration stories](../migration-stories.md): it maps Node, Go and Python projects
only and is tested with Nx 23.

## Moving from MonoSmith to Nx

No automated path is documented by either project. Nx documents `nx init` for an
existing npm, pnpm or Yarn workspace; smith's Node projects live in an npm
workspace, so that is the likely starting point. This was not tried.

## Sources

Retrieved 2026-10-25. The documentation was read as Markdoc source files from the
repository's `astro-docs/src/content/docs` folder (the published pages are the
same files), not through a summarising tool.

- Nx introduction: <https://nx.dev/docs/getting-started/intro>
- `nx release`: <https://nx.dev/docs/features/manage-releases>, which says publishing goes to a registry "such as npm for TypeScript/JavaScript libraries, crates.io for Rust, or Docker registries".
- Remote cache: <https://nx.dev/docs/features/ci-features/remote-cache>, which links a guide to "Run your own cache server against the OpenAPI specification".
- Affected: <https://nx.dev/docs/features/ci-features/affected>
- Running tasks: <https://nx.dev/docs/features/run-tasks>
- CI workflow generator: <https://nx.dev/docs/getting-started/setup-ci> ("nx g @nx/workspace:ci-workflow --ci=<provider>") and <https://nx.dev/docs/kb/ci-workflow-examples> ("The CI workflow generator doesn't support Jenkins").
- Dependency alignment: <https://nx.dev/docs/getting-started/tutorials/managing-dependencies> (single version policy; "the `@nx/dependency-checks` ESLint rule to catch version mismatches") and <https://nx.dev/docs/kb/dependency-checks>.
- Conformance plugin: <https://nx.dev/docs/reference/conformance/overview> ("allows Nx Enterprise users to write and apply rules").
- Editor setup (Nx Console): <https://nx.dev/docs/getting-started/editor-setup> ("built and maintained by the Nx team").
- Nx Cloud plans: <https://nx.dev/docs/features/ci-features> ("starts on the free Hobby plan"); self-hosted caching: <https://nx.dev/docs/enterprise/activate-license> ("Self-hosted caching is now free for everyone"). Nx's pricing page itself was not read.
- Repository and README (licence, `nx init`, plugins): <https://github.com/nrwl/nx>. GitHub API on 2026-10-25: licence MIT, archived false, last push 2026-10-23.

### Added for the master table

Retrieved 2026-10-25, from the same Markdoc source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `technologies/` (`dotnet`, `java`, `node`, `react`, `angular`, `vue`): the technologies Nx documents; `kb/nx-vs-bazel.mdoc` ("JS/TS native, plugins for Gradle, Maven, .NET").
- `features/multi-language-support.mdoc` ("parses source files and creates dependencies from `import` statements") and `getting-started/Tutorials/managing-dependencies.mdoc` ("For non-JS languages, Nx does not detect dependencies from source code imports").
- `core-features/code-organization/enforce-module-boundaries` (boundary rules).
- `features/enhance-AI.mdoc` and `getting-started/ai-setup.mdoc` (the Nx MCP server); `features/enhance-AI.mdoc` links the plugin registry (`/docs/plugin-registry`).
- `kb/release-npm-packages.mdoc` ("run the same publish action multiple times without any negative side-effects"); `guides/Nx Release/publish-in-ci-cd.mdoc` (`id-token: write` for provenance; Docker images marked experimental); `kb/publish-rust-crates.mdoc`.
- `reference/environment-variables.mdoc` (`NX_MAX_CACHE_SIZE`); `getting-started/Tutorials/caching.mdoc` (`nx reset`; runtime values in the hash); `features/cache-task-results.mdoc` and `getting-started/Tutorials/understanding-your-workspace.mdoc` (debugging cache misses).
- `features/CI Features/affected.mdoc` (task-based affected; `fetch-depth: 0` is required); `kb/workspace-watching.mdoc` (`nx watch`); `features/automate-updating-dependencies.mdoc` (`nx migrate`); `kb/from-turborepo.mdoc`.
