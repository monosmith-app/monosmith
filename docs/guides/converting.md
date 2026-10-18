# Converting from Nx (and from the old mnci)

`smith convert` turns an Nx workspace into a smith workspace in place. It also
converts a workspace made by the old, Nx-based mnci (4.x), whose settings it
reads too. Projects keep their directories and sources. What changes is how
they are described and run.

```sh
npm ci                    # Nx must be installed: it describes the projects
smith convert --dry-run    # the report; changes nothing
smith convert              # applies it
smith ci setup             # toolchains, the Python venv
smith run lint typecheck test build
git add -A && git commit -m "chore: convert to smith"
```

Convert needs a git repository with a clean working tree (or `--allow-dirty`),
so the whole conversion is one change you can review and undo.

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
| Go and Python executors (`@nx-go/nx-go`, the old mnci's Python executors) | the Go and Python plugins' own tasks; settings that differ (golangci-lint instead of `go vet`, an extra `dependsOn`) are kept in `smith.toml` |
| a `dependsOn` on another project's target (`"@acme/cli:build"`, `{projects, target}`) | `depends_on = ["cli:build"]`, with the project's smith name; one smith cannot name (a pattern or a tag) is reported |
| `@nx/esbuild:esbuild`, `@nx/js:tsc`, `@nx/jest:jest`, `@nx/vite:*`, `@nx/eslint:lint`, `@nx/js:node` | the command each stands for, as a script; a build's `assets` are copied after it by a `node -e` one-liner (Node 22 or later); `esbuildOptions` with a CLI flag (`outExtension`, `sourcemap`, `external`, `loader`, `define`, `target`, `minify`, …) become that flag, and any other is reported |
| a Node app's `prune` (`@nx/js:prune-lockfile`, `@nx/js:copy-workspace-modules`, `@nx/node:prune`) | the node plugin's `prune` task, which writes the same deploy output: a production `package.json` and lockfile and `workspace_modules/`; into the directory the Nx build wrote to (`--out`) when that is not smith's `dist/<app dir>`, so a Dockerfile written for Nx's output keeps working |
| `@nx/esbuild:esbuild` with `bundle: false` | a `node -e` one-liner that transpiles, as Nx does, the main file and every file the tsconfig includes into the same place under the output, with TypeScript reading the tsconfig and esbuild's API building; for Node CommonJS, Nx also writes a loader for the workspace libraries, so that build is reported for a hand edit |
| `{projectName}@{version}` tags | the same tags: a project whose Nx name differs from its directory gets `tag_prefix`, so the next release continues from the last tag |
| `release.changelog…createRelease: "github"` | `[release] github_releases = true` |
| a root target | a root `package.json` script (smith runs projects, not the root) |
| root scripts calling `nx run-many`, `nx affected`, `nx run`, `nx graph`, `nx release --dry-run` | the `smith` command that does the same |

A command that only a shell can run (`&&`, pipes, redirects, `$VAR`) becomes a
script in a Node project. Elsewhere it is reported, never turned into
something that behaves differently.

## What it rewrites so Nx can go

| File | Becomes |
|---|---|
| `jest.preset.js` re-exporting `@nx/jest/preset` | the same preset values written out, with a mapper for `./x.js` imports of TypeScript files (the only thing Nx's resolver was needed for) |
| a root `jest.config.ts` listing projects for Nx | removed; smith runs each project's tests |
| a project's `jest.config` using `@nx/react/plugins/jest` and `@nx/react/babel` | the transformer written out as `jest.asset-transform.cjs` at the root, and the Babel presets `@nx/react/babel` stood for (`preset-env`, `preset-typescript`, `preset-react`) |
| `@nx/react/typings/*.d.ts` in a tsconfig's `types` | in a Vite project, `vite/client`, which declares the same CSS-module and image imports (an SVG import is its URL, with no `ReactComponent` export); elsewhere, reported for you to declare them in the project |
| an ESLint config using `@nx/eslint-plugin` presets | `@eslint/js` and `typescript-eslint`, which those presets wrap; the module-boundary and dependency-check rules read the Nx graph, so they are removed; every package's `deps-check` task does the dependency check instead (see tasks and cache) |
| a library built by `@nx/rollup`'s `withNx` (it needs the Nx graph) | `esbuild` for the bundle and `tsc` for the declarations, into the same files `package.json` already publishes; then `declaration-specifiers.mjs` (written at the root) gives each relative specifier in the declarations its file (`./hooks` → `./hooks.js` or `./hooks/index.js`), as the old mnci's rollup plugin did, so a `nodenext` consumer can follow them. A config with several entries (Rollup's `input`, or `additionalEntryPoints`) builds them all in one `esbuild --splitting` run, sharing code in `dist/chunks` as Rollup did; entries it cannot read are reported, never dropped |

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

## The report

Every finding has a severity:

| | |
|---|---|
| **Blocking** | stops the conversion (two projects that would get the same name) |
| **Needs you** | something convert could not do; it says what to do instead |
| **Not supported yet** | a project convert does not map: it maps Node, Go and Python projects only, so a Flutter/Dart, .NET, Rust, JVM, PHP, Ruby, C++ or Bicep project is reported here even though smith has a plugin for its language. Also a `Dockerfile` with no manifest beside it, pnpm or yarn workspaces, releasing an app that is not Go, and publishing a VS Code extension. These projects keep their files and `project.json`; bringing them under smith is by hand for now (their language's workspace file, see [plugins](../reference/plugins.md)). |
| **Behaves differently** | a change you should know about |
| **Converted** / **Dropped** | what moved where, and the Nx-only plumbing that went |

`--strict` refuses to convert while anything needs you or is unsupported.
`--json` prints the plan for tools.
