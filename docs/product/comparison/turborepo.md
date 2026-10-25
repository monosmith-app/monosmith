# Turborepo compared with MonoSmith

Turborepo is a build system for JavaScript and TypeScript monorepos, written in
Rust, that schedules `package.json` scripts and caches their results. It is small
by design: one `turbo.json` on top of the package manager's workspaces.

## What Turborepo is best at

- **Very little to adopt.** It uses your existing `package.json` scripts and
  declared dependencies, plus a single `turbo.json`, and can be adopted
  incrementally.
- **Caching with a clear model.** Results are stored in `.turbo/cache`; a hit
  restores the declared `outputs` and replays the logs. The hash covers the global
  inputs (root `turbo.json`, lockfile, `globalDependencies`, `globalEnv` values) and
  the package's own files. Remote caching is by default Vercel Remote Cache, and a
  different provider can be self-hosted.
- **Git-aware selection.** `turbo run --affected` filters to packages changed on the
  current branch (by default `--filter=...[main...HEAD]`), and `--filter` accepts
  git ranges.
- **Other toolchains.** A multi-language guide runs other toolchains through
  `package.json` scripts (its example combines JavaScript with a .NET console app
  and a C# library). Native workspace discovery for Rust (Cargo), Python (uv) and Go
  (`go.work`) is documented, each marked experimental by its own page.
- **Editor and terminal.** A Turborepo language server (a VS Code extension, usable
  from other editors that speak LSP) and a terminal UI selected with `"ui": "tui"`.
- **Generators.** `turbo gen workspace` adds an app or package, optionally copying an
  existing one, and custom generators use Plop.
- **Open source.** MIT.

## Where MonoSmith differs

| | Turborepo | MonoSmith |
|---|---|---|
| Languages | JavaScript and TypeScript natively; other toolchains through `package.json` scripts; Cargo, uv and Go workspaces are experimental | eleven languages built in, plus plugins |
| Package manager | npm, Yarn or pnpm | npm for Node (a decision recorded in the design docs) |
| Environment variables in the cache key | `globalEnv` values are in every task's hash, and a task's `env` lists the variables it depends on, with wildcards | a task's declared `env` table |
| Release | no turbo command; the publishing guide uses Changesets ("We recommend `changesets`" for versioning, changelogs and publishing) | conventional commits to npm, PyPI, crates.io, NuGet, RubyGems, pub.dev and Maven repositories, Go modules and Packagist by tag |
| CI | guides per provider (Vercel, GitHub Actions, CircleCI, GitLab CI, Travis CI, Buildkite) | generated pipelines for GitHub Actions, Azure Pipelines and GitLab CI |
| Dependency alignment | the documentation points to `syncpack`, `manypkg` and `sherif`, or to the package manager | `smith sync`; `smith doctor` checks one version per production dependency |
| Health check | `turbo boundaries` (experimental) checks workspace dependency violations; `turbo scan` is documented as removed | `smith doctor`, with `--fix` |
| Remote execution | no mention in the documentation (searched "remote execution" and "distributed") | none |
| Repository status | MIT, not archived, last push 2026-10-23 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Turborepo instead

- Your repository is JavaScript or TypeScript only, you use pnpm or Yarn, and all you
  want is an orderly, cached script runner that adds as little as possible.
- You deploy on Vercel and want its remote cache by default.
- You rely on Turborepo's environment-variable hashing (`globalEnv`, wildcards). smith's cache hashes a task's declared `env` table, and its reference documents nothing like `globalEnv` or wildcards.

## When MonoSmith is the better fit

- The repository has other languages beside TypeScript.
- You want CI generation and releases from the same tool.

## Moving from Turborepo to MonoSmith

`smith convert` reads `turbo.json`, package-level `turbo.json` files and
`pnpm-workspace.yaml` from the files alone; nothing needs to be installed first. It
converts `dependsOn`, `outputs`, `inputs` and `cache: false`, turns pnpm or Yarn
workspaces into npm workspaces and rewrites `turbo run`, `pnpm` and `yarn` calls in
scripts. Resolved dependency versions are not carried over: review the lockfile
diff. `env`, `globalEnv`, `interactive`, `with` and `remoteCache` are reported as
not supported. See [Migration stories](../migration-stories.md).

## Moving from MonoSmith to Turborepo

No documented path. Turborepo's own guides cover migrating from Nx
(not from smith). Because smith's Node projects sit in an npm workspace with
`package.json` scripts, adding a `turbo.json` by hand is the natural start. This was
not tried.

## Sources

Retrieved 2026-10-25.

The documentation was read as MDX source files from the repository's
`apps/docs/content/docs` folder, not through a summarising tool.

- Introduction: <https://turborepo.dev/docs>
- Caching: <https://turborepo.dev/docs/crafting-your-repository/caching>
- `turbo run` reference (`--affected`, `--filter`): <https://turborepo.dev/docs/reference/run>
- Multi-language support: <https://turborepo.dev/docs/guides/multi-language> ("Turborepo can run other language toolchains through scripts"); experimental native support: <https://turborepo.dev/docs/guides/tools/rust>, <https://turborepo.dev/docs/guides/tools/python> and <https://turborepo.dev/docs/guides/tools/go> ("Native Go workspace support is experimental").
- Generating code: <https://turborepo.dev/docs/guides/generating-code>
- CI vendors: <https://turborepo.dev/docs/guides/ci-vendors>
- Publishing: <https://turborepo.dev/docs/guides/publishing-libraries> and <https://turborepo.dev/docs/core-concepts/internal-packages> ("We recommend changesets for managing versioning, changelogs, and the publishing process").
- Dependency alignment: <https://turborepo.dev/docs/crafting-your-repository/managing-dependencies> ("Tools like syncpack, manypkg, and sherif can be used for this specific purpose").
- Boundaries: <https://turborepo.dev/docs/reference/boundaries> ("the experimental turbo boundaries command"); removed command: <https://turborepo.dev/docs/support-policy> ("This command has been removed").
- Editor integration: <https://turborepo.dev/docs/getting-started/editor-integration> ("Turborepo LSP"); terminal UI: <https://turborepo.dev/docs/reference/configuration> ("Select a terminal UI for the repository").
- Repository: <https://github.com/vercel/turborepo>. GitHub API on 2026-10-25: licence MIT, archived false, last push 2026-10-23.

### Added for the master table

Retrieved 2026-10-25, from the same MDX source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `core-concepts/remote-caching.mdx` (artifacts signed with `HMAC-SHA256`; artifacts that fail to verify are treated as a cache miss).
- `crafting-your-repository/constructing-ci.mdx` ("`--affected` ... falls back to running all tasks" on shallow clones).
- `crafting-your-repository/caching.mdx` and `reference/run.mdx` (`--summarize`, `--dry`, `--graph`); `core-concepts/package-and-task-graph.mdx`; `crafting-your-repository/developing-applications.mdx` (terminal UI, watch mode); `reference/configuration.mdx` (cache eviction by age).
