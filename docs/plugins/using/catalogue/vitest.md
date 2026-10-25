# The vitest plugin

The `vitest` plugin gives Node projects unit tests with Vitest. It fills the `test` slot and
publishes the `node-vitest` stack. The browser-app stacks (`react`, `vue`, `svelte`, `html`,
`angular`) and the `vscode-extension` stack test with it as well.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `toolchain` |
| Provides (slots) | `test` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stack it publishes | `node-vitest`: scaffold `typescript`; toolchain `typescript`, `eslint`, `vitest` |
| Project kinds | any Node project in a stack that lists it |

It conflicts, through the slot, with [jest](jest.md): a stack cannot name both.

## What it does

At creation time, and only where absent:

| Where | What |
|---|---|
| root devDependencies | `vitest` `^5.0.5` |
| project scripts | `test`: `vitest run` |
| project files | `src/greeting/greet.use-case.spec.ts`, the starter's spec, importing from `vitest` |

```json
"scripts": {
  "typecheck": "tsc -p tsconfig.json --noEmit",
  "lint": "eslint .",
  "test": "vitest run"
}
```

The `test` task comes from the script, through the [node plugin](node.md#tasks): cached, after
`^build`.

## Creating a project

```sh
smith add lib a-vit --stack node-vitest
smith run typecheck lint test -p a-vit
```

```text
Wrote 10 file(s):
  eslint.config.mjs
  libs/a-vit/package.json
  libs/a-vit/src/greeting/greet.use-case.spec.ts
  ...
a-vit:test │       Tests  1 passed (1)
a-vit:typecheck ✔ 1.03s
a-vit:test ✔ 1.054s
a-vit:lint ✔ 9.227s
✔ 3/3 tasks succeeded (0 from cache) in 9.234s
```

## Settings

None from smith. Vitest reads its own configuration (`vitest.config.*` or `vite.config.*`).

## Releasing and CI

No `publish` and no `ci` method.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0: the run above, and the React stack's `smith run test -p
a-react` (4/4 tasks).

## See also

- [jest](jest.md), [eslint](eslint.md), [typescript](typescript.md), [playwright](playwright.md)
- [The node plugin](node.md)
