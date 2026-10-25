# The playwright plugin

The `playwright` plugin adds end-to-end testing to Node projects. It owns the `e2e` task: it runs
a Playwright suite, installs the browsers first in a task of its own, and checks the machine with
`smith doctor`. It is both a toolchain a stack can name and a convention you can switch on
workspace-wide with `[conventions.playwright]`.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `toolchain` |
| Provides (slots) | `e2e` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate`, `contribute`, `doctor` |
| Stack it publishes | `node-vitest-playwright`: scaffold `typescript`; toolchain `typescript`, `eslint`, `vitest`, `playwright` |
| Project kinds | any Node project with a Playwright suite |

## What it detects

A project has a suite when it has a `playwright.config.*` or declares `@playwright/test` (the test
runner flavour), or when it has its own `e2e` script that drives the `playwright` library from
another runner (the library flavour). The library may be declared by the project or, hoisted by npm
workspaces, by the root `package.json`. A project with no suite gets nothing.

## Tasks

| Task | Command | dependsOn | Cached | Outputs |
|---|---|---|---|---|
| `e2e` | `npx --no-install playwright test`, or the project's own `e2e` script when it has one; under `xvfb-run -a` on Linux when the display option is set | `e2e-browsers`, `^build`, and `build` when the project has one | yes | `dist/e2e/<project dir>` |
| `e2e-browsers` | `npx --no-install playwright install <browsers>`; in CI on Linux also `--with-deps` | none | no | none |

The contribution uses `replace`, so these tasks win over a script of the same name.

```text
a-pw:e2e — Run the end-to-end tests (Playwright runner, chromium)
  command    npm run e2e --silent
  cwd        apps/a-pw
  dependsOn  e2e-browsers (this project's e2e-browsers, first), ^build (build of the projects it imports, first), build (this project's build, first)
  inputs     **/*, {workspaceRoot}/package-lock.json
  outputs    ../../dist/e2e/apps/a-pw
  env        —
  cache      true

a-pw:e2e-browsers — Install the browsers the e2e task needs (chromium)
  command    npx --no-install playwright install chromium
  ...
  cache      false
```

(The `e2e` command printed here is `npm run e2e --silent` because the stack's scaffold wrote an
`e2e` script, `playwright test`, and the plugin keeps a project's own script.) Reports and traces
go to `dist/e2e/<project dir>`, beside the build and not in `dist/<project dir>`, which is replaced
whenever a build is restored from the cache.

## Creating a project

```sh
smith add app a-pw --stack node-vitest-playwright
```

```text
Wrote 12 file(s):
  .github/workflows/ci.yml
  apps/a-pw/e2e/smoke.e2e.ts
  apps/a-pw/package.json
  apps/a-pw/playwright.config.ts
  apps/a-pw/src/greeting/greet.use-case.spec.ts
  ...
```

Only a stack that names `playwright` gets a starter suite (`playwright.config.ts`, a smoke test in
`e2e/` named `*.e2e.ts` so Vitest and Jest never run it, an `e2e` script and `@playwright/test`).
As a workspace convention the plugin adds none: a project with no UI must not get a browser test.

## Settings

In a project's `package.json`:

```json
{ "smith": { "playwright": { "display": true, "browsers": ["chromium", "firefox"] } } }
```

| Key | Default | Means |
|---|---|---|
| `display` | `false` | run the suite under `xvfb-run -a` on Linux, for tests that open windows |
| `browsers` | `chromium` | what `e2e-browsers` installs |

## Doctor

| Check | Fails when |
|---|---|
| `Playwright is installed` | `node_modules` has no `playwright-core`. Remedy: `npm ci` |
| `a virtual display for e2e suites that open windows` | a suite sets `display` and `xvfb-run` is not on `PATH` (Linux). Remedy: `sudo apt-get install -y xvfb` |
| `a pipeline runs the e2e tasks` | `smith ci verify` leaves e2e out and no e2e pipeline is committed. Remedy: `smith pipeline e2e` |

## Running

Run suites one at a time: they usually share a fixture port and, under `xvfb-run`, a display.
`smith run e2e --parallel 1` does so; the generated `e2e` pipeline does too. The app under test is
started before the suite and stopped after by the task's `serve` setting; see
[Serving an app for a task](../../../reference/tasks-and-cache.md#serving-an-app-for-a-task).

## Releasing and CI

No `publish` and no `ci` method. The browsers can be cached by the pipeline under a key made of the
lockfile and the manifests; see [CI pipelines](../../../reference/ci.md).

## Verified

smith 1.62.0, Windows 11, Node v24.12.0. Scaffolding, `smith explain a-pw:e2e`, `explain
a-pw:e2e-browsers` and `smith doctor` (the pipeline finding) ran. **Not verified:** the browser
download and a real `smith run e2e` (they need the network and a browser); the `display` option
(Linux only).

## See also

- [vitest](vitest.md), [typescript](typescript.md), [eslint](eslint.md)
- [Plugins (reference)](../../../reference/plugins.md#playwright) for the notes this page grew from
