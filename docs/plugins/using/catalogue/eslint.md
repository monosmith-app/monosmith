# The eslint plugin

The `eslint` plugin gives Node projects linting and formatting with ESLint and
`@monosmith/eslint-config` (TypeScript rules, JavaScript Standard Style and the vertical-slice
rules). It is the `lint` toolchain of the `node-jest`, `node-vitest`, `node-vitest-playwright`,
`react`, `vue`, `svelte`, `html`, `angular`, `docs-site` and `vscode-extension` stacks.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `toolchain` |
| Provides (slots) | `lint` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stacks it publishes | none |
| Project kinds | any Node project in a stack that lists it |

## What it does

The plugin has no `contribute` method. It acts when a project is created, and adds, only where
absent:

| Where | What |
|---|---|
| workspace root | `eslint.config.mjs`: `import monosmith from '@monosmith/eslint-config'` and `export default [ ...monosmith() ]`, with a comment that later blocks win. In a `react`, `react-p5` or `react-tauri` project it is `monosmith({ verticalSlices: { roles: ['hook', 'content', 'fixture', 'style', 'mock'] } })`, so the front-end slice roles are on from the first file |
| root `package.json` devDependencies | `@monosmith/eslint-config` `^1.0.0` and `eslint` `^10.12.0` |
| root `package.json` field | `overrides` that settle the `eslint` peer range of `eslint-plugin-jsx-a11y`, so the first install prints no peer-conflict warnings |
| project `package.json` scripts | `lint`: `eslint .` |

The `lint` task then comes from the script, through the [node plugin](node.md#tasks):

```text
a-vit:lint — Lint and check formatting
  command    npm run lint --silent
  cwd        libs/a-vit
  dependsOn  ^build (build of the projects it imports, first)
  inputs     **/*, {workspaceRoot}/package-lock.json, {workspaceRoot}/tsconfig.base.json, {workspaceRoot}/eslint.config.mjs
  outputs    —
  env        —
  cache      true
```

Changing `eslint.config.mjs` is a cache miss for every `lint`, because the file is an input.

## Creating a project

```sh
smith add lib a-vit --stack node-vitest
```

```text
Adding lib a-vit (stack node-vitest) in libs/a-vit
Running npm install --no-audit --no-fund ...
Adding eslint
Running npm install --no-audit --no-fund ...
Adding vitest
...
Wrote 10 file(s):
  eslint.config.mjs
  libs/a-vit/package.json
  ...
```

`Adding eslint` is this plugin; the files it wrote are `eslint.config.mjs` and the root and project
`package.json` edits.

## Settings

None. Configure ESLint in `eslint.config.mjs`, where you add your own blocks after the spread.

## Toolchain and requirements

ESLint and the config package are installed with the workspace. `smith doctor` checks that the
packages the ESLint config imports are installed.

## Releasing and CI

No `publish` and no `ci` method.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. `smith run lint -p a-vit` passed (9 s with a cold cache).

## See also

- [vertical-slices](vertical-slices.md): for Node, the slice rules live in this ESLint config
- [jest](jest.md), [vitest](vitest.md), [typescript](typescript.md)
- [The node plugin](node.md)
