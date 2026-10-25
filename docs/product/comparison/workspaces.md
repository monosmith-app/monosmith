# npm, pnpm and Yarn workspaces compared with MonoSmith

Workspaces are the package managers' own support for several packages in one
repository. They are the baseline every JavaScript monorepo starts from, and
smith's Node plugin builds on npm's.

## What workspaces give you

- **npm** (licence: Artistic-2.0 for the CLI): `npm install` symlinks workspace
  packages into the root `node_modules`; `--workspace` (`-w`) targets one workspace and
  `--workspaces` runs a command in all of them, in the order of the root
  `package.json`; `npm install <pkg> -w <name>` adds a dependency to one workspace;
  `npm init -w ./packages/a` creates a workspace's folders and boilerplate;
  `npm publish --workspaces` publishes them; `npm doctor` checks the npm environment.
  The documented cache is the package cache (`npm cache` "manipulates packages
  cache"), and no selection of changed workspaces was found.
- **pnpm** (MIT): a `pnpm-workspace.yaml` defines the workspace; the `workspace:`
  protocol links a local package and refuses to resolve anything else, and becomes a
  regular range on publish; `--filter` selects packages by name, relation or change
  (`--filter "[origin/main]"` selects packages changed since a commit or branch);
  catalogs define version ranges once for the whole workspace. The current
  documentation (pnpm 11 and 12) goes further than the other two package managers:
  `tasks` in `pnpm-workspace.yaml` declare task dependencies (`dependsOn`), and
  `pnpm pipeline` (marked experimental, added in v12.4.0) runs "the affected projects'
  task graph, with cached results restored instead of re-run" for tasks that declare
  `outputs`. From v12.12.0 a task's result can be shared through a server that speaks
  the Turborepo Remote Cache API, or a `pnpr` server. `pnpm change` (v11.13.0) records
  a change intent in the Changesets format, `pnpm version` bumps versions, and
  `pnpm doctor` (v11.14.0) "runs diagnostics on the pnpm installation and the
  environment it runs in". Hooks live in `.pnpmfile.mjs` or `.pnpmfile.cjs`.
- **Yarn** (BSD-2-Clause): a `workspaces` array in the root `package.json`; the
  `workspace:` protocol; `yarn workspaces foreach` with `-p` for parallel, `-t` for
  topological order and `--since` for workspaces modified on the current branch;
  constraints that can require "the same version of dependencies across packages"
  (checked with `yarn constraints`) and catalogs (a plugin included by default from Yarn 4.10.0); an experimental release workflow for versions
  across a monorepo; plugins. The documented cache is the install cache.

The pnpm items above are the documentation's account of pnpm 12; they are new, and
several are marked experimental by their own pages. Check them against the version
you would run.

## Where MonoSmith differs

| | Workspaces | MonoSmith |
|---|---|---|
| Languages | JavaScript and TypeScript | eleven languages |
| Task cache | npm and Yarn: none documented; pnpm 12: experimental, for tasks that declare `outputs` | content-addressed, all languages |
| Release | npm: publish; Yarn: an experimental release workflow; pnpm: version bumps and a change-intent command | releases from commits to npm, PyPI, crates.io, NuGet, RubyGems, pub.dev and Maven repositories, Go modules and Packagist by tag |
| CI | none generates pipelines; `pnpm ci` is a clean install for CI | generated pipelines for three providers |
| Relationship | the base | smith runs npm workspaces for Node and uses `package.json` scripts as tasks |
| Repository status | npm CLI: Artistic-2.0, pushed 2026-10-23; pnpm: MIT, pushed 2026-10-25; Yarn: BSD-2-Clause, pushed 2026-10-23; none archived (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

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

The documentation was read as Markdown or MDX source files from each tool's
repository, not through a summarising tool. Paths are relative to the documentation
folder of that repository (`npm/cli: docs/lib/content`, `pnpm/pnpm.io: docs`,
`yarnpkg/berry: packages/docusaurus/docs`).

- npm workspaces: <https://docs.npmjs.com/cli/v11/using-npm/workspaces> (`using-npm/workspaces.md`: "Commands will be run in each workspace in the order they appear in your `package.json`"); `commands/npm-init.md` (`npm init -w`); `commands/npm-publish.md` (`--workspaces`); `commands/npm-doctor.md` ("Check the health of your npm environment"); `commands/npm-cache.md` ("Manipulates packages cache").
- pnpm workspaces: <https://pnpm.io/workspaces>; `filtering.md` ("Selects all the packages changed since the specified commit/branch"); `workspace-task-orchestration.md` (`tasks`, `dependsOn`); `cli/pipeline.md` ("experimental"; "with cached results restored instead of re-run"; "Added in: v12.12.0" for the shared cache); `catalogs.md`; `cli/change.md`; `cli/version.md`; `cli/doctor.md`; `pnpmfile.md`; `cli/ci.md`. pnpm's GitHub releases on 2026-10-25: v12.15.0 (2026-10-21) and v11.28.8 (2026-10-22), neither flagged pre-release.
- Yarn workspaces: <https://yarnpkg.com/features/workspaces>; `features/constraints.mdx` ("Ensuring the same version of dependencies across packages"); `features/release-workflow.mdx` (flagged "Experimental"); `features/caching.mdx` (the install cache); `features/extensibility.mdx` (plugins); `features/catalogs.mdx` ("The catalog plugin is included by default starting from Yarn 4.10.0").
- Licences: the `LICENSE` files of `npm/cli` ("Licensed on the terms of The Artistic License 2.0"), `pnpm/pnpm` (MIT) and `yarnpkg/berry` (`LICENSE.md`, BSD 2-Clause), and the GitHub API on 2026-10-25.
