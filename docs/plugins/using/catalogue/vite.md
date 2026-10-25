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
| Stacks it publishes | `react`, `react-p5`, `vue`, `svelte`, `html`: scaffold `vite`; toolchain `vite`, `eslint`, `vitest` |
| Project kinds | `app` only |

## What it does

| Stack | Starter | Type check |
|---|---|---|
| `react` | a React app (`App` component, `main.tsx`) | `tsc` |
| `react-p5` | a React app drawing with p5 | `tsc` |
| `vue` | a Vue app | `vue-tsc` |
| `svelte` | a Svelte app | `svelte-check` |
| `html` | a page in TypeScript with no framework | `tsc` |

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
(each stack adds its own framework's). Asking for another kind is an error that says what to run:
`a react project is an app: smith add app <name> --stack react`. A stack name the plugin does not
scaffold fails with `the vite plugin scaffolds the [html react react-p5 svelte vue] stacks, not "x"`.

## Settings, releasing, CI

No settings; no `publish` or `ci` method (a built app is a static site, released by its tag).

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. `smith add app a-react --stack react`, then `smith run
build typecheck lint test -p a-react`: 4/4 tasks succeeded. **Not run:** the `react-p5`, `vue`,
`svelte` and `html` stacks (same code path, different starter templates), and `smith run start`
(a dev server).

## See also

- [angular](angular.md), [astro](astro.md): the other browser-app scaffolds
- [vitest](vitest.md), [eslint](eslint.md), [docker](docker.md) (serves `dist/<app dir>` with nginx)
