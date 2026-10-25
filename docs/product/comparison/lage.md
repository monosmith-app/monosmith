# Lage compared with MonoSmith

Lage is a task runner for JavaScript monorepos from Microsoft. Its repository's
description gives that one line; the README points to a documentation site for the
rest.

## What Lage is best at

According to its documentation site:

- **An explicit task pipeline.** `lage.config.js` declares how tasks relate (`build:
  ["^build"]`, `test: ["build"]`), which `lage` combines with the package graph to
  schedule work in parallel. `npx lage init` writes a default configuration.
- **Caching.** It remembers completed work and skips what is not needed. The cache is
  layered: "first on disk, then on remote server". The remote-cache guide is written
  for Azure Blob Storage, and the configuration names the provider
  (`cacheStorageConfig.provider`, with `"azure-blob"` as the example). Uploading is
  off unless `LAGE_WRITE_REMOTE_CACHE` is set.
- **Selecting by change.** `--since <commit, tag or branch>` runs only packages that
  changed, and `--scope`, `--to` and dependency flags narrow a run.
- **Parallel execution** instead of sequential builds, with priorities.
- **Profiling.** A built-in profiler and build graph visualisations show bottlenecks.
- **Reporters** for GitHub Actions and Azure DevOps log formatting.
- **Distributed builds through BuildXL.** The documentation says "`lage` itself is a
  single-node build orchestrator" and describes a BuildXL integration for distribution.
- **Open source.** MIT.

## Where MonoSmith differs

| | Lage | MonoSmith |
|---|---|---|
| Scope | JavaScript monorepos | eleven languages |
| Task graph | `pipeline` in `lage.config.js`, written by you | inferred from imports and manifests, with `dependsOn` where you set it |
| Release, scaffolding, dependency alignment, health check | none found (searched the documentation source); `lage init` writes configuration only | releases from commits, `smith add`, `smith sync`, `smith doctor` |
| CI | log reporters for GitHub Actions and Azure DevOps; no pipeline generator found | generated pipelines for three providers |
| Editors, plugins | none found; task types are `npmScript` and `worker` | a VS Code extension, a terminal interface and an MCP server; JSON-RPC plugins |
| Repository status | MIT, not archived, last push 2026-10-24 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Lage instead

- You want a task runner for a JavaScript monorepo and nothing else.

## When MonoSmith is the better fit

- The repository has more than JavaScript, or you need CI and release handled too.

## Moving between Lage and MonoSmith

No automated path. A JavaScript monorepo using npm workspaces can be read by
`smith adopt --dry-run`. This was not tried.

## Sources

Retrieved 2026-10-25.

The documentation was read as Markdown source files from the repository's
`docs/docs` folder (paths below are relative to it), not through a summarising tool.

- Documentation site: <https://microsoft.github.io/lage/>
- Pipelines: `guides/pipeline.md` ("To define the task dependency graph, use the `pipeline` key in `lage.config.js`").
- Command line: `reference/cli.md` ("only runs packages that have changed since the given commit, tag, or branch"; the reporters table).
- Remote cache: `guides/remote-cache.md` ("first on disk, then on remote server") and `reference/config.md` (`cacheStorageConfig`).
- Distributed builds: `guides/buildxl.md` ("`lage` itself is a single-node build orchestrator").
- Quick start: `quick-start.md` (`npx lage init`).
- Repository (licence, description): <https://github.com/microsoft/lage>. GitHub API on 2026-10-25: licence MIT, archived false, last push 2026-10-24.

### Added for the master table

Retrieved 2026-10-25, from the same Markdown source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `reference/cli.md` (`--watch` "runs in watch mode").
