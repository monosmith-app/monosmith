# The functions plugin

The `functions` plugin scaffolds a serverless function app in TypeScript for Azure Functions
(Node v4 model), AWS Lambda or Google Cloud (Cloud Run) functions. It publishes the choices behind
`smith add app api --lang node --with azure-functions,http-trigger,node24`. The python and dotnet
plugins publish the same choices for their languages; this page is the Node side. See
[Function apps](../../../manual/add.md#function-apps) for what a person sees, and
[ADR 0074](../../../architecture/decisions/0074-function-apps-are-choices-of-the-language-plugins.md)
for why.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold` |
| Provides (slots) | none: the tasks are the scripts of the app's `package.json` |
| Requires | language plugin `node` |
| Applies to | `node` projects |
| Capabilities | `generate`, `doctor` |
| Stacks it publishes | none: `--with` saves the selection as a stack |
| Choices it publishes | frameworks `azure-functions`, `aws-lambda`, `gcp-functions`; triggers `http-trigger`, `timer-trigger` (not with Google), `queue-trigger`, `blob-trigger`; runtimes `node24`, `node22` |
| Project kinds | `app` only |

## What it writes

Every app: `package.json` (ES module, `engines.node`), `tsconfig.json`, `tsconfig.build.json`, the
greeting slice, `src/functions/greeting.handler.ts` with its test, `scripts/assemble.mjs`,
`scripts/package.mjs`, `scripts/start.mjs`, `README.md`. The workspace gets
`eslint` and `vitest` (the defaults); `jest` is refused.

| Host | Also writes | `start` runs |
|---|---|---|
| `azure-functions` | `host.json`, `local.settings.sample.json`, `.funcignore`, `.gitignore` (ignores `local.settings.json`); `main` is `dist/functions/*.js` | `func start` in the build folder |
| `aws-lambda` | `template.yaml` (SAM), `events/greeting.json` | `sam local start-api` (HTTP) or `sam local invoke` with the sample event |
| `gcp-functions` | `src/index.ts`; `main` is `dist/index.js` | `functions-framework` in the build folder |

## Tasks

| Task | Command |
|---|---|
| `build` | `tsc -p tsconfig.build.json && node scripts/assemble.mjs`: the code goes to the workspace's `dist/apps/<name>/dist` and `package.json` (and `host.json`) next to it, which is what the host runs |
| `typecheck` | `tsc -p tsconfig.json --noEmit` |
| `lint`, `test` | `eslint .`, `vitest run` |
| `start` | builds, then runs the function locally (see above) |
| `package` | builds, then zips `dist/apps/<name>/<name>.zip`; Azure's zip holds the production dependencies, Google's does not (the platform installs them) |

## Doctor

`smith doctor` warns, never fails, when the tool of a host the workspace uses is missing: Azure
Functions Core Tools (`func`) or the AWS SAM CLI (`sam`), with the command that installs it. The
Functions Framework is a dependency of the app, so Google needs no extra tool.

## Verified

Every host and trigger (11 combinations) was scaffolded with `smith add` in a scratch workspace; each
passed `smith run lint typecheck test build package`, and `smith doctor` passed.

## See also

- [The plugin catalogue](README.md), [vite](vite.md), [eslint](eslint.md), [vitest](vitest.md)
- [smith add: Function apps](../../../manual/add.md#function-apps)
