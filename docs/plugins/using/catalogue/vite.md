# The vite plugin

The `vite` plugin scaffolds browser apps in TypeScript that Vite serves and builds: React, React
with p5, Vue, Svelte, or a page with no framework. It is the scaffold and the `build` and
`typecheck` toolchain of five stacks, and each of them lints with [eslint](eslint.md) and tests
with [vitest](vitest.md).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold`, `toolchain` |
| Provides (slots) | `build`, `typecheck` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stacks it publishes | `react`, `react-p5`, `react-tauri`, `vue`, `svelte`, `html`: each a saved selection of choices; scaffold `vite`; toolchain `vite`, `eslint`, `vitest` (and `tauri` for `react-tauri`) |
| Choices it publishes | frameworks `react`, `vue`, `svelte`, `html`; library `p5`; styles `bootstrap`, `chakra` (requires `react`), `tailwind` |
| Project kinds | `app` only |

## What it does

The app's `tsconfig.json` extends the workspace's `tsconfig.browser.json` (written when absent), which extends
`tsconfig.base.json`; it adds only `types`, `jsx`, `paths` and `include`. See [the tsconfig chain](typescript.md#the-tsconfig-chain).

| Stack | Starter | Type check |
|---|---|---|
| `react` | a React app (`App` component, `main.tsx`) | `tsc` |
| `react-p5` | a React app drawing with p5 | `tsc` |
| `react-tauri` | the `react` app as a desktop app: the [tauri](tauri.md) plugin adds `src-tauri` | `tsc` |
| `vue` | a Vue app | `vue-tsc` |
| `svelte` | a Svelte app | `svelte-check` |
| `html` | a page in TypeScript with no framework | `tsc` |

## Choices

The frameworks and the styles are data files in the plugin's `choices/` folder, not code. A stack
is a saved selection of them, so you can mix a framework, a test runner and a style:

```sh
smith choices
smith add app web --with react,vitest,bootstrap
```

`bootstrap` and `tailwind` add their stylesheet to the entry file of any framework; `chakra` and `p5` require `react`. To add a
framework or a style, write one TOML file ([Writing choices](../../creating/choices.md)).

The scripts it writes become tasks through the [node plugin](node.md#tasks):

```json
"scripts": {
  "build": "vite build",
  "typecheck": "tsc -p tsconfig.json --noEmit",
  "start": "vite",
  "lint": "eslint .",
  "test": "vitest run"
}
```

`vite.config.ts` sends the build to the workspace's `dist/<app dir>` (`outDir: '../../dist/apps/<name>'`,
`emptyOutDir: true`), never to a `dist/` inside the app, so the task's declared output is where the
files are. Browser apps are released by their tag: they are never sent to a registry.

## Creating a project

```sh
smith add app a-react --stack react
```

```text
Adding app a-react (stack react) in apps/a-react
Adding eslint
Adding vitest
Wrote 12 file(s):
  apps/a-react/index.html
  apps/a-react/package.json
  apps/a-react/src/app/app.component.tsx
  apps/a-react/src/app/index.ts
  apps/a-react/src/greeting/greet.use-case.spec.ts
  apps/a-react/src/greeting/greet.use-case.ts
  apps/a-react/src/greeting/greeting.contract.ts
  apps/a-react/src/greeting/index.ts
  apps/a-react/src/main.tsx
  apps/a-react/tsconfig.json
  apps/a-react/vite.config.ts
  package.json
```

The root `package.json` gains `vite`, `@vitejs/plugin-react` and the React type packages
(each stack adds its own framework's). The `react` and `react-p5` stacks also add an npm
`overrides` entry that keeps `@babel/plugin-transform-runtime` on Babel 7 (next to any overrides the
workspace already has), so adding them to a workspace that holds Jest's Babel 7 still installs. Asking for another kind is an error that says what to run:
`a react project is an app: smith add app <name> --stack react`. A stack name the plugin does not
scaffold fails with `the vite plugin scaffolds the [html react react-p5 react-tauri svelte vue] stacks, not "x"`.

## One build per environment

Vite writes its `VITE_` values into the bundle when it builds, so smith builds a web app once per
environment ([ADR 0081](../../../architecture/decisions/0081-a-web-app-builds-once-per-environment-and-overlap-fixes-are-presented-choices.md)).
`smith add` asks which environments (default `lcl, dev, uat, prd`; `--environments lcl,dev,uat,prd`,
required without a terminal, `none` for one build) and writes for each a `.env.<name>` file, a script
`vite build --mode <name> --outDir <folder>` and, once, `vite.build.ts` with `build(mode)`. The default
environment is the task `build`; the others are `build:<name>`, each with its own output folder
(`dist/apps/<app>/<name>`, or `dist/<name>` in the app when asked). `smith environments` lists, adds and
removes them: see [Environments](../../../manual/environments.md).

## Settings, releasing, CI

Settings, in `[projects.<name>.settings.vite]` (written by `smith add`; they are what `smith environments`
reads): `environments` (the names), `default_environment`, `output` (`workspace` or `project`) and
`env_files` (`commit` or `example`). smith doctor checks that each environment has its `.env` file and warns
about a `VITE_` name with `KEY`, `SECRET` or `TOKEN` in it. The pipelines build the default environment. No `publish` or `ci` method (a built app is a static site, released by its tag). With `[derive] enabled = true` it answers `smith derive` with the renderer `vite.tsconfig`, the whole `tsconfig.json` of an app made from choices.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. `smith add app a-react --stack react`, then `smith run
build typecheck lint test -p a-react`: 4/4 tasks succeeded. **Not run:** the `react-p5`, `vue`,
`svelte` and `html` stacks (same code path, different starter templates), and `smith run start`
(a dev server).

## See also

- [angular](angular.md), [astro](astro.md): the other browser-app scaffolds
- [vitest](vitest.md), [eslint](eslint.md), [docker](docker.md) (serves `dist/<app dir>` with nginx)
