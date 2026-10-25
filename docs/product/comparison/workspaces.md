# npm, pnpm and Yarn workspaces compared with MonoSmith

Workspaces are the package managers' own support for several packages in one
repository. They are the baseline every JavaScript monorepo starts from, and
smith's Node plugin builds on npm's.

## What workspaces give you

- **npm:** `npm install` symlinks workspace packages into the root `node_modules`;
  `--workspace` (`-w`) targets one workspace and `--workspaces` runs a command in all
  of them, in the order of the root `package.json`; `npm install <pkg> -w <name>` adds
  a dependency to one workspace. The page read says nothing about caching or
  "affected" detection.
- **pnpm:** a `pnpm-workspace.yaml` defines the workspace; the `workspace:` protocol
  links a local package and refuses to resolve anything else, and becomes a regular
  range on publish; `--filter` selects packages; recursive runs are limited by
  `workspaceConcurrency`. Topological order is not guaranteed when workspace
  dependencies form cycles.
- **Yarn:** a `workspaces` array in the root `package.json`; the `workspace:` protocol;
  `yarn workspaces foreach` with `-p` for parallel, `-t` for topological order and
  `--since` for workspaces modified on the current branch.

None of the three, in the pages read, is described as having a task cache or a
release tool. That is not a claim that they have none; it is what the pages say.

## Where MonoSmith differs

| | Workspaces | MonoSmith |
|---|---|---|
| Languages | JavaScript and TypeScript | eleven languages |
| Caching, release, CI | not covered by the pages read | cache, releases from commits, generated CI |
| Relationship | the base | smith runs npm workspaces for Node and uses `package.json` scripts as tasks |

## When to stay with plain workspaces

- You have a handful of packages, a single language and short builds. A cache,
  a pipeline generator and a release tool are costs you may not need yet.
- Your team is committed to pnpm or Yarn. smith's Node plugin runs npm.

## When MonoSmith is the better fit

- The repository has grown several languages, long CI runs or a manual release.

## Moving to MonoSmith

An npm workspace is the easiest case: `smith adopt --dry-run` lists the projects smith
would see, the blockers and warnings, and `smith adopt --init` writes `smith.toml`
and the pipelines and commits. A pnpm or Yarn workspace first becomes an npm
workspace; `smith convert` does this for a Turborepo workspace, and a repository
with pnpm or Yarn and no Turborepo is not covered by a documented path (not tried).

## Sources

Retrieved 2026-10-25.

- npm workspaces: <https://docs.npmjs.com/cli/v11/using-npm/workspaces>
- pnpm workspaces: <https://pnpm.io/workspaces>
- Yarn workspaces: <https://yarnpkg.com/features/workspaces>
