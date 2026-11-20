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
greeting slice, `src/functions/greeting.handler.ts` with its test, `README.md`. No script is copied
into the app: the tasks run smith's own commands (`smith func assemble`, `smith func package`) and the
tool of the host (ADR 0099). The app's `tsconfig.json` extends the
workspace's `tsconfig.base.json` (written when it is absent), like every TypeScript project. The workspace gets
`eslint` and `vitest` (the defaults); `jest` is refused.

| Host | Also writes | `start` runs |
|---|---|---|
| `azure-functions` | `host.json`, `local.settings.sample.json`, `.funcignore`, `.gitignore` (ignores `local.settings.json`); `main` is `dist/functions/*.js` | `func start --script-root <build folder>` (the build copies `local.settings.json` there when it exists) |
| `aws-lambda` | `template.yaml` (SAM), `events/greeting.json` | `sam local start-api` (HTTP) or `sam local invoke` with the sample event |
| `gcp-functions` | `src/index.ts`; `main` is `dist/index.js` | `functions-framework` in the build folder |

## Tasks

| Task | Command |
|---|---|
| `build` | `tsc -p tsconfig.build.json && smith func assemble`: the code goes to the workspace's `dist/apps/<name>/dist` and `package.json` (and `host.json`) next to it, which is what the host runs |
| `typecheck` | `tsc -p tsconfig.json --noEmit` |
| `lint`, `test` | `eslint .`, `vitest run` |
| `start` | builds, then runs the tool of the host (see above): no script |
| `package` | `smith func package` (Azure: `--with-dependencies`) builds, then zips `dist/apps/<name>/<name>.zip` and copies the zip to `dist/drop/<name>/<name>.zip`, which `smith ci pack` keeps beside the app settings; Azure's zip holds the production dependencies, Google's does not (the platform installs them). Before the zip, smith checks that the build holds compiled code: the `main` of the built `package.json` (a file or a glob such as `dist/src/functions/*.js`) must match a file, or, without `main`, `dist/` must hold a `.js` file. Otherwise `smith func package` fails with `the function app has no compiled code ...` and names the folder the tsconfig `outDir` must write into. A library of the workspace (smith install -p <app> workspace:<lib>) is built and shipped too: Node packs it into a tarball under the stage and points the manifest at it, Python copies its packages into the zip, .NET publishes the project reference. A Node library whose `main` leads to built JavaScript ships from its own folder, `dist` included (ADR 0102); one whose `main` names a TypeScript source needs a build in the workspace's `dist/<library dir>`. The install of Azure's dependencies runs outside the workspace, so smith copies the nearest `.npmrc` (project folder up to the workspace root) into the stage for the install and removes it before the zip: a scope mapped to a private feed keeps working, and a token line never ships |

## Doctor

`smith doctor` also finds the scripts that older smith copied into an app (`scripts/assemble.mjs`,
`package.mjs`, `start.mjs`). `smith doctor --fix` and `smith upgrade` remove a copy as smith wrote it and
point the `package.json` scripts at the commands above. A copy you edited stays until you choose: the
fix asks whether to keep it or to remove it.

`smith doctor` warns, never fails, when the tool of a host the workspace uses is missing: Azure
Functions Core Tools (`func`) or the AWS SAM CLI (`sam`), with the command that installs it. The
Functions Framework is a dependency of the app, so Google needs no extra tool.

## Verified

Every host and trigger (11 combinations) was scaffolded with `smith add` in a scratch workspace; each
passed `smith run lint typecheck test build package`, and `smith doctor` passed.

## See also

- [The plugin catalogue](README.md), [vite](vite.md), [eslint](eslint.md), [vitest](vitest.md)
- [smith add: Function apps](../../../manual/add.md#function-apps)

## Azure app settings per environment

An Azure function app can keep one `local.settings.<env>.json` per environment (for example `local.settings.uat.json`) beside `local.settings.json`. When the app packs (`smith ci pack`, or `smith func settings` on its own), smith turns each one into the app settings of that environment, on one line, in `dist/drop/<app>/env/<env>.json`:

```json
[{"name":"ApiUrl","value":"https://api.uat","slotSetting":false}]
```

`az functionapp config appsettings set --settings @dist/drop/<app>/env/uat.json` and the portal's advanced edit read that file. `local.settings.json` and the sample stay out. So do `AzureWebJobsStorage` and `APPLICATIONINSIGHTS_CONNECTION_STRING`, which Azure sets when it creates the app. A `ConnectionStrings` section is not app settings: smith says so and leaves it out. A setting stays with a deployment slot when smith.toml names it:

```toml
[projects.orders]
slot_settings = ["MODE"]
```

