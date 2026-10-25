# Rush compared with MonoSmith

Rush is described on its site as a build orchestrator for large-scale TypeScript
monorepos. It is part of the Rush Stack projects, hosted under the Microsoft
organization on GitHub.

## What Rush is best at

- **One install, linked projects.** It installs all dependencies in one step, links
  projects to each other and builds them in dependency order. PNPM, NPM and Yarn are
  supported package managers.
- **Incremental builds.** `rush rebuild --to <project>` builds a project's upstream
  dependencies and `--from <project>` rebuilds the affected downstream projects.
- **Build cache providers.** Built-in `local-only`, `azure-blob-storage` and
  `amazon-s3`, and custom providers as Rush plugins.
- **Change files and publishing.** Rush can require a major, minor or patch change
  entry for each affected project in a pull request, bump versions, run `npm publish`
  and aggregate the entries into `CHANGELOG.md`.
- **Policies.** Enterprise policies can require review of new libraries and enforce
  consistent dependency versions across projects.
- **Open source.** The Rush Stack monorepo is MIT licensed.

## Where MonoSmith differs

| | Rush | MonoSmith |
|---|---|---|
| Languages | JavaScript and Node.js | eleven languages built in |
| Package manager | PNPM, NPM or Yarn | npm for Node |
| Version source | change files written per pull request | conventional commits |
| Remote cache | Azure Blob, S3, plugins | HTTP, S3, Azure Blob |
| Dependency consistency | policies | `smith sync`, and `smith doctor` checks one version per production dependency |
| CI generation | not covered by the pages read | generated pipelines |

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

- Welcome: <https://rushjs.io/pages/intro/welcome/>
- Build cache: <https://rushjs.io/pages/maintainer/build_cache/>
- Rush Stack repository (licence, description): <https://github.com/microsoft/rushstack> and <https://raw.githubusercontent.com/microsoft/rushstack/main/LICENSE>
