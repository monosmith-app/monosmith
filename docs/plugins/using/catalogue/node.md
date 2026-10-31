# The node plugin

The node plugin makes every `package.json` below the workspace root a smith project, in an npm workspace. It turns each project's `package.json` scripts into smith tasks (run through `npm run`, so `node_modules/.bin` is on `PATH`), works out the dependency graph from the source files and the manifest, and adds what scripts alone cannot give: a `prune` task that writes an app's deploy output, a `deps-check` task that stops a package from importing what it does not declare, and `package` for apps. It also installs Node in CI, audits the production dependency tree, and publishes packages to npm and VS Code extensions to the Marketplace. It is a built-in plugin: no installation needed. Build, lint, type-check and test scripts come from the toolchain add-ons of a stack (TypeScript, ESLint, Jest, Vitest, ...), which are separate plugins; this page says where the boundary is.

## At a glance

| | |
|---|---|
| Plugin name | `node` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold` (all three by default for a language plugin) |
| `--lang` values | `node`, `javascript`, `js` |
| Manifest that makes a directory a project | `package.json` (the one at the workspace root holds tooling only and is never a project) |
| Requires / conflicts / provides | none / none / no task slots |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest`. |
| Stacks it publishes | none itself. The add-ons publish them: `node-jest`, `node-vitest`, `node-ts`, `node-vitest-playwright`, `vscode-extension`, `react`, `vue`, `svelte`, `html`, `angular`, `docs-site`, `react-p5` (from `smith stacks --json`). `[default_stacks] node = "node-jest"` is what `smith new` writes. |
| Project kinds it can create | `app`, `lib`, `package` (`node-app`, `node-lib`, `node-package`) |

Toolchain add-ons that sit on top of this plugin: [typescript](typescript.md), [eslint](eslint.md), [jest](jest.md), [vitest](vitest.md), [playwright](playwright.md), [vite](vite.md), [angular](angular.md), [astro](astro.md), [vscode](vscode.md), and the cross-language [docker](docker.md) and [vertical-slices](vertical-slices.md).

## What it detects

(`libs/plugin-node/discovery/`)

| Field | How it is read |
|---|---|
| Project | each `package.json` in a subdirectory. Only npm is supported; the root `package.json` `workspaces` field must cover it (doctor checks). |
| Name | the last element of the directory (`packages/kit` is `kit`); the npm name is the export |
| Kind | `"smith": { "kind": "app" \| "lib" \| "package" }` in `package.json`; else the top directory (`apps/`, `libs/`, `packages/`); else `package` when the manifest is not `"private": true`, otherwise `lib` |
| Stack | `"smith": { "stack": "..." }` |
| Tags | `native` when `"smith": { "native": true }` (built on each OS) |
| Exports | the npm package `name`, when present |
| Version, description | `version`, `description` |
| Imports (graph edges) | the bare package names imported by every `.ts .tsx .mts .cts .js .jsx .mjs .cjs` file under the project (`import ... from`, `import()`, `require()`; comments and string contents are ignored; `.d.ts`, relative paths, `node:` built-ins, `#` subpath imports, and `node_modules`, `dist`, `coverage`, `build` and hidden folders are skipped), plus every name declared in `dependencies`, `devDependencies`, `peerDependencies` and `optionalDependencies`. A graph edge is made where a name equals another project's export. |
| Toolchain | `node <version>` from `node --version` |

Real output after `smith install -w web @ts-demo/kit`:

```
$ smith graph
kit  package · node · packages/kit
util  lib · node · libs/util
web  app · node · apps/web
  └─ kit
```

## Tasks

Tasks come from the project's `package.json`. This is the central rule: every key of `scripts` becomes a task named like the script, whose command is `npm run <script> --silent`, run in the project directory. Well-known names get caching and ordering; any other name is an uncached task with no dependencies.

| Script | Task description | dependsOn | Cached | Outputs |
|---|---|---|---|---|
| `build` | Compile into the workspace's `dist/<project dir>` | `^build` | yes | `dist/<project dir>` |
| `typecheck` | Type-check without emitting | `^build` | yes | none |
| `test` | Run unit tests | `^build` | yes | none |
| `lint` | Lint and check formatting | `^build` | yes | none |
| `start` | Run the app | `build`, only if the project has a `build` script | no | none |
| `dev` | Run the app in watch mode | none | no | none |
| any other (`codegen`, `e2e`, ...) | `npm script <name>` | none | no | none |

Cached tasks have the inputs `**/*`, `{workspaceRoot}/package-lock.json`, `{workspaceRoot}/tsconfig.base.json` and `{workspaceRoot}/eslint.config.mjs`. A dependency on a script the project does not have (`start` on `build` in plain JavaScript) is dropped; `^build` is always kept.

Tasks the plugin adds by itself, with no script behind them:

| Task | Command | Appears for | dependsOn | Cached |
|---|---|---|---|---|
| `package` | `npm pack --pack-destination {workspaceRoot}/dist/drop` for an app without a `build` script; `smith npm-pack <dir>` (packs the app's build directory, staged with its `package.json`) for an app that builds | an app, or any project whose `engines.vscode` is set (a VS Code extension), unless it has its own `package` script | `build` (when it builds) | no |
| `prune` | `smith npm-prune <dir>` | an app that has a `build` script, unless it has its own `prune` script | `build`, `^build` | no |
| `deps-check` | `smith npm-deps-check <dir>` | a `package`, unless it has its own `deps-check` script; the project's first of `lint`, `build`, `test` gains `deps-check` in its `dependsOn`, so CI verify runs it | none (the first of `lint`, `build`, `test` depends on it) | yes; inputs `**/*` and the root `package.json` |

`prune` writes the app's deploy output into its build directory (`dist/<app dir>`): a production `package.json`, a lockfile of the production tree at the versions the workspace installed, and the workspace packages the app uses, each as it would be published, in `workspace_modules/`, so the directory installs on its own with `npm ci --omit=dev` (a Docker image does exactly that, see [docker](docker.md)). `deps-check` fails when a file that ships imports a package that `dependencies`, `peerDependencies` (optional ones included) or `optionalDependencies` do not declare. Test, e2e, tool and config files (anything under `test/`, `tests/`, `__tests__/`, `testing/`, `e2e/`, `tools/`, or named `*.test.*`, `*.spec.*`, `*.config.*`) may also import `devDependencies`, the root's declarations and other workspace members' names.

What the toolchain add-ons contribute is the scripts themselves. In a workspace made with the default stack `node-jest`, `smith add` wrote these scripts (and so these tasks):

| Kind | Scripts in `package.json` (from the add-ons) | Tasks |
|---|---|---|
| lib | `typecheck`, `lint`, `test` (a lib is read from its sources, `"exports": "./src/index.ts"`, and has no `build`) | typecheck, lint, test |
| app | `build`, `typecheck`, `start`, `lint`, `test` | those, plus `package`, `prune` |
| package | `build`, `typecheck`, `lint`, `test` | those, plus `deps-check` |

Real output:

```
$ smith explain web
web has 7 tasks:
  build      npm run build --silent  (cached, after ^build)
  lint       npm run lint --silent  (cached, after ^build)
  package    smith npm-pack apps/web  (after build)
  prune      smith npm-prune apps/web  (after build, ^build)
  start      npm run start --silent  (after build)
  test       npm run test --silent  (cached, after ^build)
  typecheck  npm run typecheck --silent  (cached, after ^build)

$ smith explain web:build
web:build — Compile into the workspace's dist/<project dir>
  command    npm run build --silent
  cwd        apps/web
  dependsOn  ^build (build of the projects it imports, first)
  inputs     **/*, {workspaceRoot}/package-lock.json, {workspaceRoot}/tsconfig.base.json, {workspaceRoot}/eslint.config.mjs
  outputs    ../../dist/apps/web
  env        —
  cache      true

$ smith explain kit:lint
kit:lint — Lint and check formatting
  command    npm run lint --silent
  cwd        packages/kit
  dependsOn  ^build (build of the projects it imports, first), deps-check (this project's deps-check, first)
  ...
```

A plain-JavaScript workspace (`--stack none`) has only what the manifest has: the app has `start` (a script) and `package`; a package has `deps-check`; a lib has no task at all.

Real runs: `smith run build|typecheck|lint|test` succeeded; `smith run prune -p web` printed `pruned apps/web into dist/apps/web (workspace_modules: @ts-demo/kit)`, and the pruned `package.json` pointed `@ts-demo/kit` at `file:./workspace_modules/@ts-demo/kit`. A failing `deps-check`:

```
kit:deps-check │ src/extra.ts imports lodash, which package.json does not declare
kit:deps-check │ smith: packages/kit imports 1 package(s) its package.json does not declare where installing it installs them: add them to dependencies (peerDependencies when the installing app provides them)
kit:deps-check ✖ exited with code 1
```

## Creating a project

```
smith add app|lib|package <name> --lang node [--stack <stack>]
```

Without `--stack`, the workspace's default for node is used (`node-jest`, written by `smith new`); the stack's scaffold plugin (here [typescript](typescript.md)) writes the sources and the add-ons add their scripts and root dev dependencies ("Adding eslint", "Adding jest"). With `--stack none`, the node plugin's own scaffold writes plain JavaScript (CommonJS, nothing to compile). Both ways the plugin:

- writes the root `package.json` the first time (`name` = `<workspace>-workspace`, `private`, `workspaces` `["apps/*", "libs/*", "packages/*"]`) or, if one exists, adds only the `workspaces` and `devDependencies` that are missing, keeping key order and formatting;
- writes `.npmrc` once (`//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}`);
- names the package `[workspace] scope` + `/` + project name (`@js-demo/util`);
- names test files after what the workspace already uses (`*.spec.*` or `*.test.*`, the commoner one; `spec` for a new workspace);
- runs `npm install --no-audit --no-fund` at the root so the new package is linked into `node_modules`.

Real output of the plain-JavaScript flavour:

```
$ smith add lib util --lang node --stack none
Adding lib util (node) in libs/util
Running npm install --no-audit --no-fund ...
Wrote 9 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .npmrc
  azure-pipelines.yml
  libs/util/package.json
  libs/util/src/greeting/greet.use-case.js
  libs/util/src/greeting/index.js
  libs/util/src/index.js
  package.json
Import it from another Node project as '@js-demo/util'.
Try: smith test -p util
```

and of the default stack:

```
$ smith add lib util --lang node
Adding lib util (stack node-jest) in libs/util
Running npm install --no-audit --no-fund ...
Adding eslint
Adding jest
Wrote 14 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .npmrc
  eslint.config.mjs
  libs/util/package.json
  libs/util/src/greeting/greet.use-case.spec.ts
  libs/util/src/greeting/greet.use-case.ts
  libs/util/src/greeting/greeting.contract.ts
  libs/util/src/greeting/index.ts
  libs/util/src/index.ts
  libs/util/tsconfig.build.json
  libs/util/tsconfig.json
  package.json
  tsconfig.base.json
```

The plain-JavaScript starter in detail:

| Kind | `package.json` | Sources |
|---|---|---|
| app | `"private": true`, `"smith": { "kind": "app", "stack": ... }`, `"main": "src/main.js"`, script `start` = `node src/main.js` | `src/main.js` prints `Hello, world!`; `src/greeting/greet.use-case.js` and `index.js` (a worked slice with `greet(name)`) |
| lib | `"private": true`, kind `lib`, `"main": "src/index.js"` | `src/index.js` re-exports the greeting slice |
| package | `"license": "MIT"`, kind `package`, `"main": "src/index.js"`, `"files": ["src"]`, `"publishConfig": { "access": "public" }` | as lib, plus `README.md` |

The "Try: smith test -p util" hint is printed for plain JavaScript too, although that starter has no `test` script.

## Settings

The plugin has no `[projects.<p>.settings.node]` table. All configuration is in `package.json`, in the workspace files, and in `smith.toml`.

| Where | Key | Effect | Read in |
|---|---|---|---|
| `package.json` | `"smith": { "kind", "stack" }` | kind and stack | `discovery/package_manifest_contract.go` |
| `package.json` | `"smith": { "native": true }` | the project is built on each OS (tag `native`); `smith ci verify` and `pack` leave it to `smith ci native`; a Node app's release asset can be built only for the OS it runs on | same, `releasing/build_assets_use_case.go` |
| `package.json` | `private`, `name`, `version`, `description`, `scripts`, dependency sections | kind fallback, export, version, tasks, graph edges | `discovery/` |
| `package.json` | `engines.vscode` | marks a VS Code extension: gets `package`, is published to the Marketplace, not npm | `discovery/`, `releasing/publish_extension_use_case.go` |
| `package.json` | `main`, `module`, `types`, `typings`, `exports` | what doctor checks a built package has produced | `pipeline/check_entry_files_use_case.go` |
| root `package.json` | `workspaces` | which directories npm links; doctor checks every project is covered | `pipeline/check_workspace_use_case.go` |
| `package.json` build script / `tsconfig.build.json` / `tsconfig.json` | `--outDir` (wins) or `compilerOptions.outDir` | doctor checks it is inside `dist/` | `pipeline/check_build_dir_use_case.go` |
| `smith.toml` | `[workspace] scope` | npm scope of new packages | host, `scaffolding/generate_project_use_case.go` |
| `smith.toml` | `[toolchains] node = "24"` | version pin checked by doctor | host |
| `smith.toml` | `[ci] audit_policies = ["node"]`, `audit_warn`, `audit_allow_install_scripts` | lockfile policy in the audit | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[registries.<n>] npm`, `auth`, `token_env`; `[workspace] registry`; `[release] registry`; `[projects.<p>] registry` | private npm registry | `registries/`, `releasing/` |
| `smith.toml` | `[projects.<app>.release]` | release assets for an app | `releasing/build_assets_use_case.go` |
| `.npmrc` | `registry=`, tokens | read by npm; smith owns only the block between `# smith:registry` and `# smith:registry-end` | `registries/npmrc_block_mapper.go` |
| env | `NODE_AUTH_TOKEN` | npm token; read by npm through `.npmrc` | CI credential |
| env | `SMITH_NPM_ALREADY_AUTHENTICATED` | tells the release pre-flight that npm is logged in some other way | host |
| env | `VSCE_PAT`, `VSCE_AZURE_CREDENTIAL` | Marketplace token, or an Entra ID sign-in done by the pipeline | `releasing/publish_extension_use_case.go` |

## Toolchain

- `node` and `npm` must be on `PATH`. smith does not install Node. Only npm is supported: every command (`npm run`, `npm ci`, `npm install --workspace`, `npm publish`, `npm audit`) is npm's, and `smith convert` turns a pnpm or Yarn workspace into an npm one.
- The pin label is `node`, probed with `node --version`. Real output with `[toolchains] node = "22"` on a machine running 24.12.0:

```
✖ node matches the pinned version 22
    [toolchains] pins node 22, but this machine runs 24.12.0: tasks cached on another version are misses, and results can differ from CI
    → install node 22 (`mise use node@22`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] node
```

- The Node version CI installs is a constant of the plugin, `24` (`scaffolding/toolchain_versions_config.go`). The plugin's doctor compares your Node *major* with it (`Node.js is 24, as in CI`: newer warns, older fails); it does not read `.nvmrc` or `engines.node`.
- Output: `build` writes `dist/<project dir>` (set by each project's `outDir`, which doctor checks: `kit compiles into dist/packages/kit`); `package` writes `dist/drop/<name>-<version>.tgz` (a VS Code extension: `.vsix`); `prune` writes into `dist/<app dir>`; a release asset goes to `dist/release/<app>/`.

## Releasing

| Aspect | Behaviour |
|---|---|
| Version | the top-level `"version"` in `package.json`, plus the `^<version>` range of every workspace dependency (in `dependencies`, `devDependencies`, `optionalDependencies`, never `peerDependencies`) that is released in the same run. Only those string values are touched. The release writes them, publishes, and restores the manifests; a manifest with no top-level `version` fails (`package.json has no top-level "version"`). |
| Tag | `<name>@<version>` with the directory name, `kit@0.1.0` (real); override with `tag_prefix`. |
| Package (`kind = package`) | skips a version the registry has (`npm view <name>@<version> version`). A project that builds into `dist/` is published from there, staged: `Stage` copies it to a publishable `package.json` (paths into the build made relative, `files` and `scripts` removed, an `exports` of `./src/x.ts` turned into `./x.js` with `types ./x.d.ts`) beside the README and LICENSE, then `npm publish <dist dir> --access public`. A plain-JavaScript project is published from its own directory. A registry named in `smith.toml` is passed as `--registry`. |
| App | never sent to npm: "released by its tag". If its manifest is not private the message adds that to publish it, make it a package. With `[projects.<app>.release]`, the build directory is zipped into `dist/release/<app>/<product>_<version>_<os>_<arch>.zip` for this machine's platform only; asking for another platform fails with `a Node app's release assets are built on each OS ... mark it native`. The check runs in a dry run too (real, on Windows with `platforms = ["linux/amd64"]`: `Stopped: web: publishing: ...`; with `["windows/amd64"]` the dry run says `web: would attach dist/release/web/web_0.1.0_windows_amd64.zip`). |
| VS Code extension | a project with `engines.vscode` is packed with `vsce package` into `dist/drop/<name>-<version>.vsix` and published with `vsce publish --skip-duplicate`; with a native job's platform, `--target <win32-x64 ...>`. Needs `VSCE_PAT` (or `VSCE_AZURE_CREDENTIAL`). See [vscode](vscode.md). |
| Registry | `registry.npmjs.org` unless `.npmrc` or `[registries]` says otherwise; `published` runs `npm view <name> version`. A bare, unscoped name on the public registry belongs to whoever published first, so the version found there is not attributed to a project (`ownsName`); a scoped name or a named registry is. |
| Credentials | `NODE_AUTH_TOKEN` for npmjs.org; a private registry signs in as in [CI](#ci); `VSCE_PAT` for the Marketplace. |
| Dry run | publish reports "dry run" and runs nothing. Real: |

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  kit                  unreleased → 0.1.0  (1 releasing commit(s))

$ smith release --dry-run --json      # excerpt
  "tag": "kit@0.1.0",  "publish": "dry run",  "registry": "https://registry.npmjs.org/"
```

The release's SBOM lists the packages of the nearest `package-lock.json` (the project's or the workspace's), falling back to the ranges in `package.json`, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

Real output of `smith ci generate` (GitHub; Azure shows `UseNode@1` with `24.x`):

```yaml
      - uses: actions/setup-node@v7
        with:
          node-version: '24'
      - name: Cache the npm download store
        uses: actions/cache@v4
        with:
          path: ~/.npm
          key: smith-npm-${{ runner.os }}-${{ hashFiles('package-lock.json') }}
      ...
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NODE_AUTH_TOKEN }}
```

| Contribution | What |
|---|---|
| Toolchain | Node `24`: `actions/setup-node@v7`, `UseNode@1` (`24.x`), NodeSource elsewhere, the `node` dev container feature |
| Package store | `~/.npm` (`NPM_CONFIG_CACHE`), keyed on `package-lock.json` |
| Setup (`smith ci setup`) | `npm ci --no-audit --no-fund`, blocking. Real: `▶ Install Node dependencies`, `added 3 packages`, `smith ci setup passed in 1.5s`. |
| Audit | `smith npm-audit`, blocking, cached on `package-lock.json`: runs `npm audit --json` and `npm audit --omit=dev --json`; fails only for an advisory of moderate severity or higher in the production tree whose fix `npm audit fix --dry-run` can apply; dev-only advisories and unfixable ones are printed as notes. A broken audit service passes with the reason printed. Real: `npm audit - 4 advisory(ies), none actionable at moderate or above in production dependencies.` With `[ci] audit_policies = ["node"]` also `smith npm-policy [--allow-scripts=a,b]` over the lockfile: every registry package has an integrity hash, none comes from git or plain http, only the allow-listed packages run install scripts. The install scripts of the default stack's own tools are allowed without a setting (`@parcel/watcher` and `unrs-resolver`, brought by `jest` and `@monosmith/eslint-config` when the root `package.json` lists them); the generated step shows them in `--allow-scripts=`. Any other install script still needs `audit_allow_install_scripts`. |
| Credentials | `NODE_AUTH_TOKEN` (required; only when a package publishes to npmjs.org), `VSCE_PAT` (when an extension publishes). Private registries: Azure Artifacts gets `npmAuthenticate@0` on `.npmrc`, no secret; GitHub Packages gets `GITHUB_TOKEN` and `packages: read` (`write` when publishing there); a `token` registry gets the variable named in `token_env`. |

Real, with `[registries.libs] auth = "azure-artifacts"` and `[workspace] registry = "libs"`: the Azure pipeline got `- task: npmAuthenticate@0` with `workingFile: '.npmrc'`, and `smith doctor --fix` wrote this block into `.npmrc` (the rest of the file stays yours):

```
# smith:registry (written from [registries] in smith.toml; smith doctor --fix rewrites it)
registry=https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/
# smith:registry-end
```

## Doctor, dependencies and removal

Doctor checks (`pipeline/`; shown when the workspace has Node projects):

| Check | Fails when |
|---|---|
| `Node.js is <major>, as in CI` | a warning when your Node is newer than major `24`; a failure when it is older, or when `node` is not on PATH |
| `Node dependencies are installed` | no `node_modules` at the root; the fix runs `npm install --no-audit --no-fund` |
| `.npmrc points at the registries smith.toml names` | only with a registry named; the fix rewrites smith's block |
| `the scripts of <project> call files and tools that exist` | a script runs a file (`node x.js`, `--config x`) that does not exist, or starts a tool that is installed neither in a `node_modules/.bin` nor on PATH. Files under a build output are "not built yet", not missing. Judged once dependencies are installed; a failure when the broken script is one of `build`, `test`, `lint`, `typecheck`, `prebuild`, `postbuild`, `pretest`, `start`, `dev`, otherwise a warning. The fix asks whether to install the tools or remove the scripts. |
| `the packages the ESLint config imports are installed` | the root `eslint.config.*` imports something not in `node_modules`; the fix installs it |
| `npm workspaces include <dir>` | the root `workspaces` does not match the project (remedy also names `[workspace] exclude`) |
| `<project> compiles into dist/<dir>` | the `--outDir` of the build script, or the `outDir` of `tsconfig.build.json` / `tsconfig.json`, is inside the source tree |
| `<dir> declares the packages it imports` | a package (`kind = package`) fails `deps-check`; the fix declares the imports in `package.json` |
| `<project> builds every file its package.json names` | only after a build: `main`, `module`, `types`, `exports` entries inside `dist/` that the build did not produce. A `.d.ts` entry that nothing writes (a bundler such as esbuild emits none) gets a remedy that names the `tsc --emitDeclarationOnly` step to append to the `build` script, and a fix that appends it (`tsc -p tsconfig.build.json --emitDeclarationOnly --declaration --declarationDir <the types entry's directory>`; asked about, never `--yes` alone). A build script that already runs `tsc` or a dts option is told its declarations land elsewhere than `package.json` says. |

27 checks passed on the default-stack workspace. Findings in the plain-JavaScript workspace that come from the host and the vertical-slices convention rather than from the node plugin:

```
! web can be verified
    web has no lint task, so CI never checks it
    → add the task to the project (or in smith.toml under [projects.web.tasks])
✖ Node projects are checked for vertical slices (@monosmith/eslint-config)
    the workspace root has no eslint.config file, so the slice rules do not run on kit, util, web
```

Dependencies:

| Command | What it does |
|---|---|
| `smith install` | `npm install --no-audit --no-fund` at the root (every workspace) |
| `smith install -w <project> <pkg>...` | `npm install <pkg>... --workspace <dir> --no-audit --no-fund` (`-D` adds `--save-dev`) |
| `smith uninstall -w <project> <pkg>...` | `npm uninstall <pkg>... --workspace <dir>` |
| `smith up`, `smith sync` | the plugin lists every declaration of the root and each project (`dependencies`, `devDependencies`, `optionalDependencies`, `peerDependencies`), with installed versions from `node_modules`, and rewrites plain ranges (`^1.2.3`, `~1`, `>=2`); tags, URLs, git refs, `file:`, `workspace:`, `npm:` aliases and compound ranges are reported and left alone; peers are never rewritten. Latest versions come from `npm view`, so `.npmrc` applies; it also compares Node with the newest LTS. |

Real: `smith install -w web @ts-demo/kit` printed `✔ web: added @ts-demo/kit (apps/web/package.json, package-lock.json)`. `smith up --check`:

```
patch
  @monosmith/eslint-config     1.0.0 → 1.0.2  (root)
minor
  eslint                       10.12.0 → 10.13.0  (root)
major
  @types/node                  24.19.2 → 26.6.5  (root)
  typescript                   6.0.3 → 7.0.2  (root)

Tools smith pins (informational):
  Node.js (LTS)    24 ✔ (latest 24.22.1)
```

Removal: `smith remove <project>` deletes the directory, drops its entries from `package-lock.json`, and runs `npm install` so `node_modules` and the lockfile forget it (npm alone would keep it, marked extraneous). Real dry run: `Would remove util (libs/util)`, `changes package-lock.json`, `would run npm install to update the lockfile`.

## Requirements and limits

- `node` and `npm` on PATH; network for `npm install`. Node 24 is what CI uses and what the doctor expects.
- npm workspaces only. Yarn and pnpm are not supported by the plugin (commands are hard-wired to `npm`).
- The graph reads imports with a regular-expression scan of source files; dynamic specifiers (`require(name)`) are invisible.
- Every script becomes a task, but only the six well-known names are cached and ordered. Cache any other by overriding it in `smith.toml` (`[projects.<p>.tasks.<t>] cache = true`).
- A Node app's release assets cannot be cross-built; they are built per OS with `"native": true`.
- Plain-JavaScript projects (`--stack none`) have no `build`, `lint` or `test`; that is the toolchain add-ons' job.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, node v24.12.0, npm 11.6.2. Two throw-away workspaces: `js-demo` (`--stack none`, plain JavaScript) and `ts-demo` (default stack `node-jest`: TypeScript ~6.0.3, ESLint 10.12.0 with `@monosmith/eslint-config`, Jest 30). Run: `smith add app|lib|package ... --lang node` (both stacks), `smith projects`, `smith graph`, `smith explain`, `smith run build|typecheck|lint|test|deps-check|prune|package|start`, `smith ci pack`, `smith doctor` (clean; missing lint tasks; a wrong `[toolchains] node`; an undeclared import; a registry in `.npmrc`), `smith doctor --fix --yes` (`.npmrc`), `smith ci generate`, `smith ci setup`, `smith ci audit` (with and without the `node` policy), `smith ci check`, `smith release --dry-run [--json]` (including a Node app with release assets), `smith install -w`, `smith up --check`, `smith remove util --dry-run --force`, `smith plugin list --json`.

`smith run package` creates `dist/drop` before it runs, so an app without a `build` script packs on a fresh checkout as `smith ci pack` does.

Not verified:

- A real `npm publish` / staging of a package (`Stage`, the `exports` rewrite): needs a registry login and would publish; read from `releasing/stage_package_use_case.go`.
- `vsce` packaging and publishing of an extension, and `smith ci native` (see [vscode](vscode.md)).
- GitHub Packages and `token` registries; trusted publishing; GitLab output.
- pnpm and Yarn workspaces (not supported by design).
- Linux and macOS.

## See also

- [The plugin catalogue](README.md)
- Toolchains: [typescript](typescript.md), [eslint](eslint.md), [jest](jest.md), [vitest](vitest.md), [playwright](playwright.md), [vite](vite.md), [angular](angular.md), [astro](astro.md), [vscode](vscode.md)
- [docker](docker.md), [vertical-slices](vertical-slices.md)
- [Releasing](../../../reference/releasing.md), [Tasks and cache](../../../reference/tasks-and-cache.md), [CI](../../../reference/ci.md), [Configuration](../../../reference/configuration.md)
- [Writing a language plugin](../../creating/language-plugin.md)
