# Research 01 — Current mnci inventory (parity source)

Source: the TypeScript mnci repo (`MoNecromanCi`), verified against `packages/cli/src/**`.
**Code wins over docs: `mnci-details.md` is badly stale** (see "Stale docs").

## Shape

Three layers a Go rewrite must treat separately:

1. Pure file and text generation (overlay, scaffolds, CI YAML) — Nx-independent.
2. Per-ecosystem dependency tooling — Nx-independent.
3. Orchestration that shells out to Nx (generators, run-many, affected, release, sync) — **what the Go tool replaces**.

## Commands (18, not 6)

| Command | Notes |
|---|---|
| `new [name]` | `--yes --scope --registry azure-artifacts\|npm\|none --organization --project --artifacts-feed --agent --variable-group --npm-auth pat\|build-identity --ci azure\|github\|both --test-runner jest\|vitest --nx-cloud --into --preset web-api` |
| `upgrade` | overlay flags (re-persisted) plus `--migrate`; reads defaults from the `mnci` block |
| `dev [projects...]` | `--all --dry-run`; runs `start` targets together |
| `doctor` | `--json`; read-only; exit 1 on a failed finding |
| `adopt` | report mode; steps `--tags --toolchain --nx --dependencies --kinds --kind dir=kind --overlay` |
| `ci <phase>` | `setup\|audit\|verify\|pack\|release\|native` |
| `lint` | `--affected --base`; one ESLint process for npm projects |
| `release` | `--dry-run --yes` |
| `pipeline [template]` | `e2e`, `package-zip`, `deploy-pages`, `deploy-azure-function` |
| `mcp` | stdio MCP server |
| `sync` | `--check --ecosystem` (11 ecosystems) |
| `up [packages...]` | `--check --yes --json --ecosystem --no-install` |
| `install`/`i`, `uninstall`/`un` | `-w <project>`; a package with no `-w` is refused |
| `remove <project>` | `--dry-run --yes --force` |
| `add [kind] [name]` | many kind-specific flags (`--empty --app --port --e2e --esm --publisher --sidecar --release --cgo --template --web --framework`) |
| `commands`, `kinds`, `projects`, `info` | `--json`; consumed by the VS Code extension |

## Project kinds (54)

Families: TS/JS (react, node, npm lib, function app), TS web (p5, html, angular, vue, svelte, docs-site),
IaC/ops (bicep, container), editor (vscode-extension), Python, Go, Flutter, C#, Rust, C++, JVM
(java/kotlin/android), PHP, Ruby. Layout: `apps/` (never released, packed to drop), `packages/` (publishable),
`python-packages/`, `libs/` (internal). Every kind scaffolds a `greeting` contract/use-case slice unless `--empty`.

## Overlay (owned files)

Writes: `nx.json` (merged), root `package.json`, `.npmrc`, `commitlint.config.mjs`, `.husky/commit-msg`,
`eslint.config.mnci.mjs` (always), `eslint.config.mjs` (write-once), `.devcontainer/devcontainer.json`,
`<workspace>.code-workspace`, `azure-pipelines.yml`, `.github/workflows/ci.yml`, `.github/dependabot.yml`,
`.gitignore` entries. Deletes retired formatter configs, `.vscode/`, per-project eslint configs, Nx agent
scaffolding (entry by entry, only when it holds nothing but Nx's entries). Write-once: agents, `.mcp.json`,
pipeline templates, pipeline slot contents.

## Doctor (27 checks)

ESLint root config and no per-project configs; no retired formatter; eslint plugin registered; eslint major;
`.npmrc` auth and scope routing; npm credential shape; no runtime deps in root; rollup source maps and
declaration specifiers; `versionActions` overrides; target files exist; TS project refs synced; Go plugin
registered; `go.work` entries exist; vscode sidecar dependency; native pipeline coverage; pipeline phases
present; `mnci.native` valid; `preserveMatchingDependencyRanges` false; no unused local registry; release tags
resolve under current names; tags not behind registry; React tsc output not Vite's folder.

## CI

Generated YAML calls `npx mnci ci <phase>` (no inline guards). Phases: **setup** (Python toolchain/projects,
`go work sync`, golangci-lint at a pinned, checksum-verified version, Flutter SDK cloned *outside* the workspace,
pub get, JVM/PHP/Ruby/C++ checks), **audit** (npm audit blocks moderate+ with a fix; pip-audit report-only),
**verify** (audit, `sync:check`, lint, typecheck/test/build; affected on PRs; verify everything when base is
unknown), **pack** (`dist/drop`, zip names are a contract), **release**, **native** (OS matrix). Azure and GitHub
renderers are kept byte-identical by an anti-drift test.

## Release

Independent tags `{projectName}@{version}`, conventional commits, tag-only (`git.commit:false`). Per-manifest
version writers: package.json, pyproject.toml, pubspec.yaml, .csproj, Cargo.toml, CMakeLists.txt, pom.xml,
build.gradle.kts, composer.json (`extra.mnci.version`), `version.rb`. go-app: git tag only. go-lib: tags
`<dir>/vX.Y.Z` (feat→minor, fix/perf→patch, breaking→major; 0.x maps breaking→minor, else→patch).
`ci release` sequence: refuse shallow clone, npm/PyPI preflights, count releasables, validate
`RELEASE_SPECIFIER`, publish credentials, forced versions via `version(<project>)[<version>]:` directives,
partial-release survivor tag push, go-lib tags, push tags. `mnci release` adds a blocker assessment and plan.

## Dependency model

Runtime deps live in the project that imports them, never the root. Ecosystems: npm, pip, pub, nuget, cargo,
cmake/vcpkg, maven, gradle, composer, bundler, go. `sync` converges ranges to the highest (peers excluded,
non-semver blocked). `up` reports patch/minor/major/non-semver grouped, with a tool-versions section.

## Invariants to keep

Overlay deterministic; upgrade writes only mnci-owned files; argv arrays never shell strings; peer ranges never
rewritten; Flutter SDK outside the workspace; build outputs are directories; both CI providers in lockstep;
drop zip basenames are a contract; tool versions pinned in one place; nothing installs `latest`.

## Pinned tools

Node 24, npm 12, ESLint ^10.8.0, golangci-lint 2.14.0, Go 1.27, Flutter 3.47.7, .NET 10.0.x, Rust 1.99,
Java 21, Gradle 9.8.1, Maven 3.10.0, PHP 8.4, Ruby 3.4, vcpkg 2026.07.29.

## What Nx provides that must be replaced

`create-nx-workspace`, generators, inferred targets (`@nx/eslint`, `@nx/js/typescript`, `@nx/vite`, …),
`run-many`/`run`, `affected`, project graph (+ tags, implicit deps, language import graphs), `nx sync`
(TS project references), `nx release`, cache (outputs must be directories), `nx migrate`, Nx Cloud.

## Stale documentation

`mnci-details.md` claims 6 commands / 19 kinds, single-line `node -e` CI guards, and a 10-file overlay. All
wrong. Test counts (437 unit / 151 e2e) unverified. Use the code, not the docs, as the parity target.

## Compatibility constraint

The VS Code extension consumes `commands|kinds|projects|info|doctor --json`; its manifest test snapshots
`mnci commands --json`. Keep these JSON contracts compatible.
