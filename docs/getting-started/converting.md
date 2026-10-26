# Converting from Nx (and from the old mnci)

`smith convert` turns an Nx workspace into a smith workspace in place. It also
converts a workspace made by the old, Nx-based mnci (4.x), whose settings it
reads too. Projects keep their directories and sources. What changes is how
they are described and run. A Turborepo workspace is converted too; see
[Converting from Turborepo](#converting-from-turborepo).

```sh
smith convert --dry-run    # the report; changes nothing
smith convert              # applies it
smith ci setup             # toolchains, the Python venv
smith run lint typecheck test build
git add -A && git commit -m "chore: convert to smith"
```

Convert needs a git repository with a clean working tree (or `--allow-dirty`),
so the whole conversion is one change you can review and undo.

| Flag | Meaning |
|---|---|
| `--dry-run` | report the plan; change nothing. Always run this first |
| `--yes` | apply without asking. In a terminal convert shows its verdict and asks `Apply? [y/N]`; scripts and CI are not asked |
| `--verbose` | name every project a finding applies to, not just a few |
| `--allow-dirty` | convert even with uncommitted changes (not with `--level full`) |
| `--strict` | refuse to convert while anything needs manual work or is unsupported |
| `--skip-install` | do not run `npm install` afterwards |
| `--json` | print the plan as JSON, for tools |
| `--level conservative\|full` | how far to go; see [Going further](#going-further---level-full). Default `conservative` |
| `--lint`, `--test`, `--build`, `--ci` | the choices behind the full level |
| `--skip-verify` | full level: do not build, test, lint and run doctor afterwards |

Every flag, its default and its errors are in [`smith convert`](../manual/convert.md).
If your repository is not an Nx or Turborepo workspace at all, you want
[`smith adopt`](adopting.md) instead; in a terminal, a bare `smith` in the folder opens a
wizard that does either (see [the adopt wizard](../manual/tui.md#the-adopt-wizard)).
Words used below (project, kind, task, stack) are defined in the
[glossary](../reference/glossary.md).

## How it reads the workspace

Convert asks the workspace's own Nx for its project graph (`nx graph --file`).
That gives every project with its targets **fully resolved**: targets that Nx
plugins infer (`@nx/js/typescript`, `@nx/eslint`, `@nx/jest`, `@nx/vite`, …),
`targetDefaults` and `namedInputs` applied, exactly as Nx would run them. It
does not re-implement Nx's inference.

It is tested with Nx 23. With another major the report says so.

## What maps to what

| Nx | smith |
|---|---|
| a project | a project named after its directory; its language comes from its manifest (`package.json`, `go.mod`, `pyproject.toml`) |
| `projectType: application` | kind `app` |
| a project `nx release` publishes (`release.projects`, or every non-private package when unset) | kind `package` |
| any other project | kind `lib` |
| a Node package outside the root `package.json` `"workspaces"`, known to Nx only by its `package.json` (no `project.json`, not released) | no project: its directory goes into `[workspace] exclude` (`<parent>/*` when every project there is like it), as npm leaves it out of the workspace |
| an app with a `Dockerfile` | the same app, and `[conventions.docker]`: an `image` task, and its release pushes the image |
| a kind its directory does not imply (`python-packages/x`, a published `libs/x`) | the kind written into the manifest (`"smith": {"kind": …}`, `[tool.smith] kind`, `// smith:kind`) |
| a Node target running in the project root | a `package.json` script, run by the Node plugin (npm runs scripts through a shell, so nothing is lost) |
| any other `nx:run-commands` target | an `[projects.<name>.tasks.<target>]` task in `smith.toml`: argv, cwd, inputs, outputs, `depends_on` and cache carried over |
| Go and Python executors (`@nx-go/nx-go`, the old mnci's Python executors) | the Go and Python plugins' own tasks; settings that differ (an extra `dependsOn`) are kept in `smith.toml`, and a `.golangci.yml` or `ruff.toml` you already have stays and wins over smith's default lint config |
| a `dependsOn` on another project's target (`"@acme/cli:build"`, `{projects, target}`) | `depends_on = ["cli:build"]`, with the project's smith name; one smith cannot name (a pattern or a tag) is reported |
| `@nx/esbuild:esbuild`, `@nx/js:tsc`, `@nx/jest:jest`, `@nx/vite:*`, `@nx/eslint:lint`, `@nx/js:node` | the command each stands for, as a script; a build's `assets` are copied after it by a `node -e` one-liner (Node 22 or later); `esbuildOptions` with a CLI flag (`outExtension`, `sourcemap`, `external`, `loader`, `define`, `target`, `minify`, …) become that flag, and any other is reported |
| a Node app's `prune` (`@nx/js:prune-lockfile`, `@nx/js:copy-workspace-modules`, `@nx/node:prune`) | the node plugin's `prune` task, which writes the same deploy output: a production `package.json` and lockfile and `workspace_modules/`; into the directory the Nx build wrote to (`--out`) when that is not smith's `dist/<app dir>`, so a Dockerfile written for Nx's output keeps working |
| `@nx/esbuild:esbuild` with `bundle: false` | a `node -e` one-liner that transpiles, as Nx does, the main file and every file the tsconfig includes into the same place under the output, with TypeScript reading the tsconfig and esbuild's API building; for Node CommonJS, Nx also writes a loader for the workspace libraries, so that build is reported for a hand edit |
| an Azure Pipelines definition | replaced by the one smith generates; the agent pool it named and the variable groups it linked are carried (`[ci] azure_pool`, `azure_variable_groups`), and the report says its own steps are not (see `git diff`: put them between the `smith:slot` markers). A private npm feed in `.npmrc` needs a `[registries]` entry |
| `{projectName}@{version}` tags | the same tags: a project whose Nx name differs from its directory gets `tag_prefix`, so the next release continues from the last tag. When a project has tags under both names (a rename left one stray old tag), the prefix of the newest release wins |
| `release.changelog…createRelease: "github"` | `[release] github_releases = true` |
| a root target | a root `package.json` script (smith runs projects, not the root) |
| root scripts calling `nx run-many`, `nx affected`, `nx run`, `nx graph`, `nx release --dry-run` | the `smith` command that does the same |

A command that only a shell can run (`&&`, pipes, redirects, `$VAR`) becomes a
script in a Node project. Elsewhere it is reported, never turned into
something that behaves differently.

## What it rewrites so Nx can go

| File | Becomes |
|---|---|
| `jest.preset.js` building on `@nx/jest/preset` | the `require('@nx/jest/preset')` line becomes `require('@monosmith/jest-preset')`, the published package with the same preset values; everything else in the file (your `setupFilesAfterEnv`, `transformIgnorePatterns`, ...) is kept, and `@monosmith/jest-preset` becomes a root devDependency. A file that is only a re-export becomes a one-line re-export of the package; one convert cannot read is replaced the same way and reported, since what it added is gone. The file stays because your projects' Jest configs name it as their `preset` |
| `@nx/jest/plugins/resolver` | nothing is written: `@monosmith/jest-preset` wires in a resolver that asks Jest's own resolution first, then TypeScript's, which finds a module that only has types (`jest.mock('vscode')` with just `@types/vscode` installed) and what a tsconfig `paths` entry names; a `./x.js` import of a TypeScript file resolves to `./x`. Older versions of convert wrote this as `jest.resolver.cjs` at the root: see below |
| a root `jest.config.ts` listing projects for Nx | removed; smith runs each project's tests |
| a project's `jest.config` using `@nx/react/plugins/jest` and `@nx/react/babel` | the transformer becomes `'@monosmith/jest-preset/asset-transform'` (an SVG import is a React component, any other asset is its file name), and the Babel presets `@nx/react/babel` stood for (`preset-env`, `preset-typescript`, `preset-react`) are written into the config. No helper file is written; the package is a root devDependency. Older versions wrote `jest.asset-transform.cjs` at the root: see below |
| `@nx/react/typings/*.d.ts` in a tsconfig's `types` | in a Vite project, `vite/client`, which declares the same CSS-module and image imports (an SVG import is its URL, with no `ReactComponent` export); elsewhere, reported for you to declare them in the project |
| an ESLint config using `@nx/eslint-plugin` presets | `@eslint/js` and `typescript-eslint`, which those presets wrap; the module-boundary and dependency-check rules read the Nx graph, so they are removed; every package's `deps-check` task does the dependency check instead (see tasks and cache) |
| a library built by `@nx/rollup`'s `withNx` (it needs the Nx graph) | `esbuild` for the bundle and `tsc` for the declarations, into the same files `package.json` already publishes; then `declaration-specifiers.mjs` (written at the root) gives each relative specifier in the declarations its file (`./hooks` → `./hooks.js` or `./hooks/index.js`), as the old mnci's rollup plugin did, so a `nodenext` consumer can follow them. A config with several entries (Rollup's `input`, or `additionalEntryPoints`) builds them all in one `esbuild --splitting` run, sharing code in `dist/chunks` as Rollup did; entries it cannot read are reported, never dropped. `main` counts as an entry beside `additionalEntryPoints`; a one-line config that builds through a shared base (`module.exports = require('../../rollup.base.cjs')()`) is read through it, and the base is removed once every project that used it is converted |

A workspace without Nx still needs what Nx's Jest plugins did, so convert adds
[`@monosmith/jest-preset`](../../packages/jest-preset/README.md) at `^0.1.0`
instead of leaving files at the root that a tidy teammate would delete. A
project that only needs the defaults says `preset: '@monosmith/jest-preset'`,
and a React project can say `preset: '@monosmith/jest-preset/react'`.

**Converted with an older smith?** Then `jest.resolver.cjs`, `jest.asset-transform.cjs`
and a `jest.preset.js` with the values written out are at your root. `smith doctor`
reports them ("Jest helper files of an older convert are replaced by
@monosmith/jest-preset") and `smith doctor --fix` replaces them: it points the Jest
configs at the package, adds the package to the root devDependencies, runs
`npm install` and deletes the two files. It asks first, and does nothing unless it
recognises every file as one convert wrote; if you changed the preset's values or a
config of yours still names a helper, doctor says which and leaves it to you.

Then it removes `nx.json`, `.nx/`, every `project.json` of a converted project,
the `nx` fields of `package.json` files, and the `nx`, `@nx/*`, `@nx-go/*` and
old-mnci dependencies and overrides. It adds what the new scripts need
(`esbuild`, `typescript-eslint`), drops Nx's entries from `package-lock.json`
(npm would otherwise keep `@nx/eslint-plugin`, an optional peer of
`@mnci/eslint-config`, and with it Nx), runs `npm install`, and checks that
`npm ci` accepts the lockfile.

## The old mnci

On top of the above, a workspace made by mnci 4.x gets:

- the workspace name, scope and CI providers from the `mnci` block of `nx.json`
- Go libraries (`type:go-lib`) as packages, keeping their `<dir>/vX.Y.Z` tags
- releasable Go apps (`release:go`) as apps with `[projects.<name>.release]`.
  The release builds their per-platform zips under the same names
  (`release.asset` in `project.json`, or `go-app-<app>-<os>-<arch>.zip`) and
  attaches them to the GitHub Release.
- the targets it generated for Nx release (`package-all`, `build-all`,
  `nx-release-publish`, and the `tools/*-version-actions.cjs` and
  `tools/go-app-release.cjs` hooks) removed, because smith does that itself
- `package` targets kept as `package` tasks, which `smith ci pack` runs into
  `dist/drop`; a Go or Python app's uses the plugin's own instead (the binary,
  the wheel). A kept task that declared the shared `dist/drop` as its output
  gets its own `dist/drop/<project dir>/` instead, in its command and its
  `outputs`, because a cache restore replaces a task's output directories and
  would delete another package's zip mid-write; `smith doctor` warns about
  any two tasks whose outputs overlap
- the per-project root scripts (`<project>:test`, …) removed in favour of
  `smith run <task> -p <project>`
- pipeline slots kept: `after-install` becomes `before-setup` and
  `before-release` becomes `after-verify`. Steps in a slot with no equivalent
  are reported.
- `requirements-dev.txt` completed with the tools the Python plugin runs

One behaviour changes, and the report says so: before 1.0, a `feat:` releases
a minor version (the old mnci released a patch).

## Converting from Turborepo

`smith convert` also converts a Turborepo workspace (it has a `turbo.json`).
Turborepo workspaces are usually pnpm or yarn, and smith's Node plugin runs
npm ([ADR 0007](../architecture/decisions/0007-npm-only-for-node.md)); smith has its own install command,
`smith install` (`smith i`), whatever the language, and for Node it runs npm.
So convert does not add pnpm or yarn: it turns the workspace into the npm
workspace `smith i` installs, and the report says exactly what changed or went.

```sh
smith convert --dry-run    # the report; nothing needs installing first
smith convert              # applies it, then runs npm install
smith i                    # from now on, instead of pnpm install or yarn
smith run build test
```

It reads `turbo.json` (comments allowed), the `turbo.json` some packages keep
beside their `package.json`, `pnpm-workspace.yaml` and the packages the
workspace lists, from the files alone. It is tested with Turborepo 2; with
another major the report says so.

| Turborepo | smith |
|---|---|
| a workspace package | a project named after its directory, as for Nx; `apps/` is an app, a package that is not `private` is a package, the rest are libs |
| `package.json` scripts | stay scripts; the Node plugin runs each as a task (`npm run <script>`) |
| `dependsOn: ["^build"]`, `["build"]`, `["web#build"]` | `depends_on` in `smith.toml`: `^build`, `build`, `web:build`. The Node plugin already makes `build`, `typecheck`, `test` and `lint` wait for `^build`, so a definition that says the same writes nothing, and one that says less (a `lint` with no `dependsOn`) writes an empty `depends_on`, as Turborepo ran it. A task waiting for its own `build` keeps `^build`, as `build` itself waits for it |
| `outputs` | `outputs`: the directory each glob starts in (`dist/**` is `dist`); a `!` exclusion cannot be stated, and the report says so |
| `inputs` | `inputs`, with `$TURBO_DEFAULT$` as `**/*` and `$TURBO_ROOT$/x` as `{workspaceRoot}/x`; the lockfile stays an input |
| `globalDependencies` | inputs of every cached task |
| `cache: false` | `cache = false` |
| `pkg#task` in the root, and a package's own `turbo.json` | applied to that package: the root entry replaces the general one, the package configuration is laid over it, `$TURBO_EXTENDS$` included |
| a task Turborepo caches that the Node plugin does not (`codegen`) | `cache = true` on it |
| root scripts and package scripts calling `turbo run`, `pnpm` or `yarn` | the `smith` or `npm` command that does the same: `turbo run build --filter=web` is `smith run build -p web --bail`, `pnpm install` is `smith install`, `pnpm -r run test` is `smith run test`, `pnpm build` is `npm run build`; a script that only enforced a package manager (`npx only-allow pnpm`) is removed |

What it changes to get npm workspaces, each in the report:

| From | To |
|---|---|
| `pnpm-workspace.yaml` `packages` (or yarn's object-form `workspaces`) | the root `package.json` `"workspaces"` list; `pnpm-workspace.yaml` is removed |
| `workspace:*`, `workspace:^`, `workspace:~`, `workspace:1.2.3` | `*`, `^<the package's version>`, `~<the package's version>`, `1.2.3`; npm links a workspace package by name when the range matches its version |
| `catalog:` and `catalog:name` | the range in `pnpm-workspace.yaml`'s `catalog` or `catalogs` |
| `pnpm.overrides`, pnpm 10's `overrides` in `pnpm-workspace.yaml`, yarn `resolutions` | `overrides`, for a plain package name; a selector (`a>b`, `foo@<2`, `**/foo`) is reported for you to write |
| `pnpm-lock.yaml`, `yarn.lock`, `bun.lock`, `.yarnrc.yml`, `.yarn/`, `.pnp.cjs` | removed. **Resolved versions are not carried over**: `npm install` resolves every range again and writes `package-lock.json`, so a dependency can move inside its range; review the diff |
| the `packageManager` field, `engines.pnpm` and `engines.yarn`, the `pnpm` field | removed (`pnpm.overrides` moved, any other key is reported as unsupported) |
| `turbo.json`, package `turbo.json` files, `.turbo/`, `turbo`, `turbo-ignore` and `@turbo/*` | removed |

What it cannot map is reported, never dropped silently:

| Turborepo | In the report |
|---|---|
| `env`, `globalEnv` | **Not supported**: Turborepo keys the cache on these variables' values, smith's task cache does not. Set `cache = false` on the task, or list the values in its `env` table |
| `passThroughEnv`, `globalPassThroughEnv` | dropped: smith runs tasks with the whole environment, there is no strict mode |
| `persistent: true` | **Behaves differently**: the task is left uncached and `smith run` keeps it running; `smith dev` starts several apps' `start` tasks together |
| `interactive`, `with` | **Not supported**: `smith run` does not forward stdin, and has no link between a persistent task and the tasks started beside it |
| `remoteCache` | **Not supported**: smith shares its cache through the bucket or URL set as `[cache] remote` (see [tasks and cache](../reference/tasks-and-cache.md#remote-cache)) |
| a build whose `outputs` are inside the package | **Needs you**: `smith doctor` fails a build that does not write to `dist/<project dir>`; point the tool there and update `main`, `types` and `exports` |
| `//#task` (a root task), a `dependsOn` smith cannot name | a root task stays a root script; **Needs you** for the rest |
| `ui`, `daemon`, `envMode`, `cacheDir`, `outputLogs` and the like | dropped: they only set how turbo runs |
| a script with a `--filter` by graph or path (`web...`, `./apps/*`), `--dry`, `--parallel`, arguments after `--` | **Needs you**: left as it is |
| ESLint configs using `eslint-plugin-turbo`, workflows and Dockerfiles calling `pnpm`, `yarn` or `turbo`, `.yarnrc.yml` registries, `.changeset/` | **Needs you** or **Behaves differently**, naming the file |

The sections, `--strict`, `--json` and `--dry-run` work as for Nx.

## Going further: `--level full`

A plain `smith convert` is the **conservative** level: it keeps what works,
maps what can be mapped, and lists the rest under "things to check". Most of
that list is old tooling that still runs: the shared ESLint config and the
plugins only it needed, a Rollup or SWC build nothing calls any more, a test
runner's configs. `--level full` makes smith the only tool: it converts, works
out which of those files and `devDependencies` nothing uses any more, and
deletes them, asking first.

```sh
smith convert --level full --dry-run     # the whole plan, as diffs; nothing changes
smith convert --level full               # in a terminal: asks, then converts
smith convert --level full --lint smith --test jest --build smith --ci smith --yes
```

It is for Node workspaces (npm packages, ESLint, Jest or Vitest, Rollup, SWC,
Webpack, Nx and mnci); other languages' leftovers are not looked for yet. It
needs a clean git tree, because `git restore --staged --worktree . && git clean -fd`
is the way back.

`smith adopt` sends a repository it recognises as Nx or Turborepo here, and
`smith adopt --level full` in one is `smith convert --level full`. The full level
for a repository with no such tool (`smith adopt --level full`) is not built
yet: it is tracked in issue #553, and the depth levels as a whole in #548. Other
languages' leftovers are not looked for yet.

In a terminal, a bare `smith` in the folder asks the same four questions in the
adopt wizard, converts on a scratch copy first to show what would be left, and
asks about each group of removals; see
[the adopt wizard](../manual/tui.md#the-adopt-wizard).

### The choices

Four decisions, each with a default taken from what the workspace uses. In a
terminal the ones you did not give as flags are asked as numbered questions
(the suggested option first; Enter takes it). Elsewhere (a pipe, CI) it
prints the plan and stops with a non-zero exit, naming the flags that apply it.

| Flag | Values | What it decides |
|---|---|---|
| `--lint` | `smith`, `keep` | `smith`: the root ESLint config uses `@monosmith/eslint-config` (with the vertical-slice rules) and the old lint packages and config files go. If the config imported `@mnci/eslint-config`, only the import and the call change and the rules you wrote beside it stay; a config written by hand is replaced by smith's two lines (the diff shows what is lost). Default: `smith` when the config imports a shared config smith replaces, else `keep`. |
| `--test` | `jest`, `vitest`, `keep` | The configs and packages of the runner **not** chosen go. A runner is never rewritten to another: if the other one's tests exist, its files stay and the report says why. Default: the runner the workspace uses. |
| `--build` | `smith`, `keep` | The configs and packages of the old build (Rollup, Webpack, SWC and Babel config files, their loaders and CLIs) go; the build is the one convert wrote, esbuild and `tsc`. Default: `smith`. |
| `--ci` | `smith`, `keep` | `keep` leaves your pipelines as they are instead of generating smith's; they keep calling the old tool, and `smith doctor` reports the pipelines as not matching until you run `smith ci generate`. Default: `smith`. |

`--yes` accepts every group below without asking; `--dry-run` prints the whole
plan and changes nothing (`--json` prints it as data); `--skip-install` leaves
the install, and `--skip-verify` the checks, out.

### What it removes, and when it keeps something

The removals come in groups, each with its reason, and each is asked on its own
(`y` applies it, anything else leaves it): lint ("remove 3 old lint packages and
1 other lint config file: you chose smith's ESLint"), the test runner not
chosen, the old build, and Nx and mnci leftovers. Every group is shown as the
diffs it makes (a `package.json` edit is a diff too) before you answer, and the
next group is planned from the files as they are after your answer, so a group
you decline keeps what it would have removed.

A file or a package is removed only when nothing that stays imports or runs it:
the scan reads every script, CI file, config and source file left after the
group (a comment on a line of its own does not count), and a package also stays
when a package that stays needs it as a peer dependency, or a project lists it as
a `peerDependency`. Whatever looks like the old tooling but stays is listed with
what keeps it ("kept rollup: used by packages/a/package.json"). When it cannot be
sure, it keeps.

Afterwards one `npm install` refreshes `package-lock.json` for everything taken
out and put in, and the workspace is checked the way a teammate would: it builds,
tests and lints, and `smith doctor` finds nothing wrong. The step that fails is
named, with the end of its output.

## The report

Every finding has a severity:

| | |
|---|---|
| **Blocking** | stops the conversion (two projects that would get the same name) |
| **Needs you** | something convert could not do; it says what to do instead |
| **Not supported yet** | a project convert does not map: it maps Node, Go and Python projects only, so a Flutter/Dart, .NET, Rust, JVM, PHP, Ruby, C++ or Bicep project is reported here even though smith has a plugin for its language. Also a `Dockerfile` with no manifest beside it, pnpm or yarn workspaces in an Nx workspace (a Turborepo one is converted to npm workspaces), releasing an app that is not Go, and publishing a VS Code extension. These projects keep their files and `project.json`; bringing them under smith is by hand for now (their language's workspace file, see [plugins](../reference/plugins.md)). |
| **Behaves differently** | a change you should know about |
| **Converted** / **Dropped** | what moved where, and the Nx-only (or Turborepo, pnpm and yarn) plumbing that went |

`--strict` refuses to convert while anything needs you or is unsupported.
`--json` prints the plan for tools.
