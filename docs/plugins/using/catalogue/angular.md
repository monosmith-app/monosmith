# The angular plugin

The `angular` plugin scaffolds an Angular app (a standalone root component, no zone.js) that the
Angular CLI serves and builds. It is the scaffold and the `build` and `typecheck` toolchain of the
`angular` stack, which also lints with [eslint](eslint.md) and tests with [vitest](vitest.md).

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold`, `toolchain` |
| Provides (slots) | `build`, `typecheck` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate` |
| Stack it publishes | `angular`: scaffold `angular`; toolchain `angular`, `eslint`, `vitest` |
| Project kinds | `app` |

## What it does

The scripts it writes become tasks through the [node plugin](node.md#tasks):

| Script | Command | Note |
|---|---|---|
| `build` | `ng build` | into `dist/<app dir>` |
| `typecheck` | `ngc -p tsconfig.app.json` | the Angular compiler type-checks the templates too |
| `start` | `ng serve` | |
| `lint` | `eslint .` | from the `eslint` plugin |
| `test` | `vitest run` | from the `vitest` plugin |

The root `package.json` gains `@angular/build`, `@angular/cli` and `@angular/compiler-cli`; the app
declares `@angular/common`, `compiler`, `core`, `platform-browser` and `rxjs`.

## Creating a project

```sh
smith add app a-ng --stack angular
```

```text
Wrote 14 file(s):
  apps/a-ng/.gitignore
  apps/a-ng/angular.json
  apps/a-ng/index.html
  apps/a-ng/package.json
  apps/a-ng/src/app/app.component.ts
  apps/a-ng/src/app/index.ts
  apps/a-ng/src/greeting/greet.use-case.spec.ts
  apps/a-ng/src/greeting/greet.use-case.ts
  apps/a-ng/src/greeting/greeting.contract.ts
  apps/a-ng/src/greeting/index.ts
  apps/a-ng/src/main.ts
  apps/a-ng/tsconfig.app.json
  apps/a-ng/tsconfig.json
  package.json
```

## Settings, releasing, CI

No settings; no `publish` or `ci` method. An Angular app is released by its tag.

## Requirements and limits

The Angular CLI refuses to run on an older Node than it supports. On Node 24.12.0 the build task
fails, and the message comes from the CLI, not from smith:

```text
a-ng:build │ The Angular CLI requires a minimum Node.js version of v22.22.3 or v24.15.0 or v26.0.0.
a-ng:build ✖ exited with code 3
```

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. Scaffolding and `smith run typecheck -p a-ng` (`ngc`)
passed. `smith run build -p a-ng` failed for the Node version reason above, so a successful
Angular build is **not verified** here.

## See also

- [vite](vite.md), [astro](astro.md), [vitest](vitest.md), [eslint](eslint.md)
