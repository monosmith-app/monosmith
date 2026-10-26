# The typescript plugin

The `typescript` plugin makes a Node project a TypeScript project. It is the scaffold and the
`build` and `typecheck` toolchain of every built-in Node stack (`node-ts`, `node-jest`,
`node-vitest`, `node-vitest-playwright`). Without it the [node plugin](node.md) scaffolds plain
JavaScript. The plugin writes files and scripts when a project is created; it adds no task of its
own afterwards (the node plugin turns the scripts into tasks).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold`, `toolchain` |
| Provides (slots) | `build`, `typecheck` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stack it publishes | `node-ts`: scaffold `typescript`, toolchain `typescript`, no lint and no tests |
| Project kinds | `app`, `lib`, `package` |

## What it does

When `smith add <kind> <name> --stack node-ts` (or a stack that names `typescript` as its
scaffold) creates a project, the plugin writes the TypeScript starter and the project's scripts.
Nothing it writes replaces a file that exists.

| Kind | Scripts it writes | Notes |
|---|---|---|
| `lib` | `typecheck` | a library is imported from its sources (`"exports": "./src/index.ts"`), so it has no `build` |
| `app` | `build` (`tsc -p tsconfig.build.json`), `typecheck`, `start` (`node ../../dist/apps/<name>/main.js`) | `main` points at the build output in `dist/` |
| `package` | `build`, `typecheck` | published as compiled JavaScript |

The node plugin reads the scripts and creates the tasks `build`, `typecheck` and `start`; see
[its Tasks section](node.md#tasks). `build` is cached and writes `dist/<project dir>`.

## Creating a project

```sh
smith add lib a-ts --stack node-ts
```

```text
Adding lib a-ts (stack node-ts) in libs/a-ts
Running npm install --no-audit --no-fund ...
Wrote 12 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .npmrc
  libs/a-ts/package.json
  libs/a-ts/src/greeting/greet.use-case.ts
  libs/a-ts/src/greeting/greeting.contract.ts
  libs/a-ts/src/greeting/index.ts
  libs/a-ts/src/index.ts
  libs/a-ts/tsconfig.build.json
  libs/a-ts/tsconfig.json
  package.json
  tsconfig.base.json
Import it from another Node project as '@add-demo/a-ts'.
Try: smith test -p a-ts
```

The starter is a `greeting` slice (a use case and a contract) in the vertical-slice layout. The
manifest records the stack: `"smith": { "kind": "lib", "stack": "node-ts" }`. The workspace gets a
`tsconfig.base.json`, which only extends `@monosmith/tsconfig/node.json` (target `es2023`, `module: nodenext`, `strict`,
`verbatimModuleSyntax`, `erasableSyntaxOnly`, `rewriteRelativeImportExtensions`: the options live in that package, so you
adapt them by setting a value in the base), and the root `package.json` gets the `typescript`, `@types/node` and
`@monosmith/tsconfig` dev dependencies. Each project's `tsconfig.json` extends the base and holds only `rootDir` and `include`;
its `build` script passes `--outDir` for `dist/<project dir>`, so no tsconfig names an output folder.

## Settings

None. The plugin reads no `smith.toml` table; the TypeScript configuration is the generated
`tsconfig` files, which you own.

## Toolchain

`typescript` and `@types/node` are installed with the workspace (`npm install` runs when the root
dev dependencies change). smith runs the compiler through the project's scripts, so the Node and
npm versions are the [node plugin's](node.md#toolchain).

## Releasing and CI

The plugin has no `publish` and no `ci` method. A TypeScript package releases through the node
plugin (`npm publish` of the build); see [node: Releasing](node.md#releasing).

## Requirements and limits

- Needs the `node` language plugin enabled.
- Does not add `lint` or `test`: `node-ts` has neither, and `smith doctor` says so
  (`a-ts has no lint task, so CI never checks it`). Use `node-jest` or `node-vitest` for those.
- Applying it to an existing project does nothing for files that exist: the plugin only adds.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0, npm 11.6.2. `smith add lib a-ts --stack node-ts` wrote
the tree above; `smith run typecheck -p a-ts` passed; `smith run build -p a-pw` (an app with the
`node-vitest-playwright` stack, which scaffolds with `typescript`) passed; `smith plugin check
typescript` passes.

## See also

- [The node plugin](node.md), [vitest](vitest.md), [jest](jest.md), [eslint](eslint.md)
- [Stacks](../../../reference/stacks.md): the built-in stacks
- [The catalogue](README.md)
