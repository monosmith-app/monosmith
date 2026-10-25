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
| CI pipelines | Nx Cloud CI features; whether Nx generates pipeline files was not verified | smith generates GitHub Actions, Azure Pipelines and GitLab CI and `doctor` fails on drift |
| Implementation | built with Rust and extensible via TypeScript, per its README | one static Go binary |
| Editors | Nx Console for VS Code and JetBrains | a VS Code extension, a terminal interface and an MCP server; no JetBrains plugin was found in the repository |
| Maturity | not assessed here | release 1.62.0, ten days of repository history |

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

Retrieved 2026-10-25.

- Nx introduction: <https://nx.dev/docs/getting-started/intro>
- `nx release`: <https://nx.dev/docs/features/manage-releases>
- Remote cache: <https://nx.dev/docs/features/ci-features/remote-cache>
- Affected: <https://nx.dev/docs/features/ci-features/affected>
- Running tasks: <https://nx.dev/docs/features/run-tasks>
- Editor setup (Nx Console): <https://nx.dev/docs/getting-started/editor-setup>
- Pricing (free tier exists): <https://nx.dev/pricing>
- Repository and README (licence, `nx init`, plugins): <https://github.com/nrwl/nx>
