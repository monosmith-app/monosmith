# Rush compared with MonoSmith

Rush is described on its site as a build orchestrator for large-scale TypeScript
monorepos. It is part of the Rush Stack projects, hosted under the Microsoft
organization on GitHub.

## What Rush is best at

- **One install, linked projects.** It installs all dependencies in one step, links
  projects to each other and builds them in dependency order. PNPM, NPM and Yarn are
  supported package managers.
- **Incremental builds.** `rush build` rebuilds the projects whose files changed since
  their last successful build, "as well as any projects that directly or indirectly
  depend on a changed project"; `--impacted-by` and `--changed-projects-only` narrow
  the selection, and `rush rebuild --to <project>` and `--from <project>` select by
  dependency. Selection is by file hashes against the last build, not by a version
  control range.
- **Build cache providers.** Built-in `local-only`, `azure-blob-storage` and
  `amazon-s3`, and custom providers as Rush plugins.
- **Change files and publishing.** Rush can require a major, minor or patch change
  entry for each affected project in a pull request, bump versions, run `npm publish`
  and aggregate the entries into `CHANGELOG.md`.
- **Policies and checks.** Enterprise policies can require review of new libraries.
  `rush check` "ensures that all dependencies are of the same version throughout the
  repository", and `rush scan` reports "phantom dependencies" a project imports without
  declaring.
- **Cobuilds (experimental).** Several CI machines run the same pipeline and share work
  through the build cache.
- **An MCP server** (`@rushstack/mcp-server`) for coding assistants.
- **Open source.** The Rush Stack monorepo is MIT licensed.

## Where MonoSmith differs

| | Rush | MonoSmith |
|---|---|---|
| Languages | JavaScript and Node.js | eleven languages built in |
| Package manager | PNPM, NPM or Yarn | npm for Node |
| Version source | change files written per pull request | conventional commits |
| Remote cache | Azure Blob, S3, plugins | HTTP, S3, Azure Blob |
| Dependency consistency | `rush check` fails on different versions of one dependency | `smith sync`, and `smith doctor` checks one version per production dependency |
| CI generation | a CI page with the commands to put in your script, and a generated bootstrap script (`install-run-rush.js`); no pipeline generator found | generated pipelines |
| Remote execution | experimental cobuilds share work through the cache | none: tasks run on one machine |
| Plugins | Rush plugins (for example cache providers); the pages on creating and using them are titled experimental | out-of-process JSON-RPC plugins in any language |
| Scaffolding | `rush init` writes configuration; adding a project is documented as copying files and editing `rush.json` | `smith add`, recipes |
| Repository status | MIT (stated in the monorepo LICENSE), not archived, last push 2026-10-23 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Rush instead

- You work in a large TypeScript monorepo, use PNPM, and want an explicit change-file
  workflow and enterprise policies.

## When MonoSmith is the better fit

- The repository has languages beyond JavaScript, or you want versions derived from
  commit messages instead of change files.

## Moving between Rush and MonoSmith

smith has no Rush importer. `smith adopt --dry-run` reads npm workspaces; how a
Rush repository's projects map to smith's npm workspaces was not checked.

## Sources

Retrieved 2026-10-25.

The documentation was read as Markdown source files from the `rushstack-websites`
repository (`websites/rushjs.io/docs/pages`), not through a summarising tool.

- Welcome: <https://rushjs.io/pages/intro/welcome/>
- Build cache: <https://rushjs.io/pages/maintainer/build_cache/>
- Incremental builds: <https://rushjs.io/pages/advanced/incremental_builds/>, and `rush build`: <https://rushjs.io/pages/commands/rush_build/>
- `rush check`: <https://rushjs.io/pages/commands/rush_check/> ("ensures that all dependencies are of the same version throughout the repository"); `rush scan`: <https://rushjs.io/pages/commands/rush_scan/>
- Cobuilds: <https://rushjs.io/pages/maintainer/cobuilds/> ("a lightweight solution for distributing work across multiple machines"; the page title says experimental)
- Enabling CI builds: <https://rushjs.io/pages/maintainer/enabling_ci_builds/>; adding a project: <https://rushjs.io/pages/maintainer/add_to_repo/>; `rush init`: <https://rushjs.io/pages/commands/rush_init/>
- MCP server: <https://rushjs.io/pages/ai/rush_mcp/>
- Rush Stack repository: <https://github.com/microsoft/rushstack>. Its LICENSE file says "The projects in this monorepo are licensed under the MIT license". GitHub API on 2026-10-25: archived false, last push 2026-10-23.
