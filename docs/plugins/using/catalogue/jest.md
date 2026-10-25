# The jest plugin

The `jest` plugin gives Node projects unit tests with Jest and `ts-jest`. It fills the `test`
slot, publishes the `node-jest` stack, and `node-jest` is the default stack of a workspace made
with `smith new`.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `toolchain` |
| Provides (slots) | `test` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stack it publishes | `node-jest`: scaffold `typescript`; toolchain `typescript`, `eslint`, `jest` |
| Project kinds | any Node project in a stack that lists it |

`jest` and [vitest](vitest.md) both provide `test`, so a stack cannot hold both:
`jest and vitest both provide the test task; keep one`.

## What it does

Like [eslint](eslint.md), the plugin acts at creation time and adds only what is absent:

| Where | What |
|---|---|
| root devDependencies | `jest` `^30.5.2`, `@jest/globals` `^30.5.2`, `ts-jest` `^29.4.15` |
| project scripts | `test`: `jest` |
| project `package.json` field | `jest`: test environment `node`, `roots` `<rootDir>/src`, a `moduleNameMapper` that maps a relative `.js` import to its `.ts` source, and a `ts-jest` transform that compiles to CommonJS |
| project files | `src/greeting/greet.use-case.spec.ts`, the starter's spec (only when the project has the starter's `greet.use-case.ts`), named `.test.ts` instead if the workspace's tests already use that suffix |

The `test` task comes from the script (cached, after `^build`, see [node: Tasks](node.md#tasks)).

## Creating a project

```sh
smith add lib a-jest --stack node-jest
```

```text
Adding lib a-jest (stack node-jest) in libs/a-jest
Adding eslint
Adding jest
Wrote 9 file(s):
  libs/a-jest/package.json
  libs/a-jest/src/greeting/greet.use-case.spec.ts
  libs/a-jest/src/greeting/greet.use-case.ts
  libs/a-jest/src/greeting/greeting.contract.ts
  libs/a-jest/src/greeting/index.ts
  libs/a-jest/src/index.ts
  libs/a-jest/tsconfig.build.json
  libs/a-jest/tsconfig.json
  package.json
```

```text
$ smith run typecheck lint test -p a-jest
a-jest:test │ Test Suites: 1 passed, 1 total
a-jest:test │ Tests:       1 passed, 1 total
...
✔ 3/3 tasks succeeded (0 from cache) in 4.804s
```

## Settings

None from smith. Jest is configured by the `jest` field of the project's `package.json`.

## Requirements and limits

The starter spec is written for the TypeScript starter. A plain JavaScript project (`--stack none`)
gets no Jest from this plugin.

## Releasing and CI

No `publish` and no `ci` method.

## Verified

smith 1.62.0, Windows 11, Node v24.12.0: the run above.

## See also

- [vitest](vitest.md), [eslint](eslint.md), [typescript](typescript.md), [playwright](playwright.md)
- [The node plugin](node.md)
