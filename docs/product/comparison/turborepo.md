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
- **Generators.** `turbo gen workspace` adds an app or package, optionally copying an
  existing one, and custom generators use Plop.
- **Open source.** MIT.

## Where MonoSmith differs

| | Turborepo | MonoSmith |
|---|---|---|
| Languages | the npm ecosystem; no other language is mentioned in the docs read | eleven languages built in, plus plugins |
| Package manager | npm, Yarn or pnpm | npm for Node (a decision recorded in the design docs) |
| Environment variables in the cache key | `globalEnv` values are in the global hash (task-level `env` was not read) | a task's declared `env` table |
| Release | not covered by the pages read | conventional commits to eight registries |
| CI | guides per provider (Vercel, GitHub Actions, CircleCI, GitLab CI, Travis CI, Buildkite) | generated pipelines for GitHub Actions, Azure Pipelines and GitLab CI |
| Health check, dependency alignment | not covered by the pages read | `smith doctor`, `smith sync` |
| Maturity | not assessed here | release 1.62.0, ten days of repository history |

## When to choose Turborepo instead

- Your repository is JavaScript or TypeScript only, you use pnpm or Yarn, and all you
  want is an orderly, cached script runner that adds as little as possible.
- You deploy on Vercel and want its remote cache by default.
- You rely on Turborepo's environment-variable hashing, which smith's cache does not reproduce.

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

- Introduction: <https://turborepo.dev/docs>
- Caching: <https://turborepo.dev/docs/crafting-your-repository/caching>
- `turbo run` reference (`--affected`, `--filter`): <https://turborepo.dev/docs/reference/run>
- Generating code: <https://turborepo.dev/docs/guides/generating-code>
- CI vendors: <https://turborepo.dev/docs/guides/ci-vendors>
- Repository (licence): <https://github.com/vercel/turborepo>
