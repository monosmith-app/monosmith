# ADR 0072: Function apps are choices of the language plugins

- Status: accepted
- Date: 2026-11-02

## Context

The owner wants `smith add app api --lang python --with azure-functions`, and the same for
Node and C#, for three hosts: Azure Functions, AWS Lambda and Google Cloud Functions (now
"Cloud Run functions"). That is nine templates ([#898](https://github.com/monosmith-app/monosmith/issues/898)).
Today the Add wizard finds nothing for `function` or `azure`, and a generated Python or C# app
has no `function_app.py`, `host.json` or Functions worker.

Questions to settle: where the templates live, how the person answers the questions a function
app has (trigger, runtime version), and how a wrong combination fails.

## Decision

### Where it lives: choices on the `framework` axis, in the language plugins

Rejected: a new plugin for function hosts. It would write a manifest for three languages and so
copy what the `node`, `python` and `dotnet` plugins already do (the venv, `package.json` scripts,
the csproj, discovery). The `vite` plugin works because it calls the node scaffolding; a
cross-language plugin would call three.

Chosen: the plugin that writes the project for a language publishes three choices (ADR 0045) on
the `framework` axis: `azure-functions`, `aws-lambda` and `gcp-functions`, with the role
`scaffold`. For Python and C# that is the language plugin. For Node it is a new toolchain plugin,
`functions`, beside `vite`: the `node` plugin writes plain JavaScript and the TypeScript
starters already live in toolchain plugins. The stack that `smith add --with` saves names that
plugin as its scaffold, so the host calls its `generate` with `choices`. The `framework` axis gives the conflicts for free:
`azure-functions` with `react` or with `aws-lambda` stops with "both on the framework axis: keep
one", before any write. Node declares `needs = ["lint", "test"]` like `react`: eslint and Vitest
join.

### Core change that is unavoidable: a choice name is unique per language

Three languages offer the name `azure-functions`. Until now a name was unique among all plugins.
Now the key is language and name. `Solve` takes the language (`--lang`); with none it infers it
and, when a name is offered for several languages, stops with "azure-functions is offered for
node, python and dotnet: add --lang". A saved stack in `smith.toml` gets a `language` line so it
expands the same way later. `smith choices` lists the language of every row.

### Trigger and runtime are axes, and nothing is chosen silently

`trigger` has `http-trigger`, `timer-trigger`, `queue-trigger` and `blob-trigger`. `runtime` has
the versions each platform supports. Each host `needs` both axes. A choice gets a new member,
`suggested`: the choice an asker pre-selects, which the solver never adds on its own (`default`
does, and stays for lint and test).

- Terminal: `smith add` shows the options with the suggested one marked, says what each means in
  plain words, lists the files it will write, and asks.
- No terminal: it refuses before any write: `azure-functions needs a choice on the trigger axis:
  add one of http-trigger (suggested), timer-trigger, queue-trigger, blob-trigger to --with`.
  No new flags: `--with azure-functions,http-trigger,node24` is the whole answer, and any future
  axis works the same.
- TUI and VS Code: the choices step lists the axes; a radio step per needed axis appears after
  a host is ticked, with the suggested option pre-selected and the file list in the confirmation.

Where a host has no such event, the choice conflicts with it and says why in the error:
`timer-trigger` with `gcp-functions` (on Google a schedule is a Cloud Scheduler job that calls an
HTTP or Pub/Sub function). `queue-trigger` means Azure Storage Queue, AWS SQS and Google Pub/Sub;
`blob-trigger` means an Azure blob, an AWS S3 event and a Cloud Storage event.

### Runtimes (checked on 2026-11-02, see Sources)

| | Azure Functions v4 | AWS Lambda | Cloud Run functions |
|---|---|---|---|
| Node.js | 22, 24 | 22, 24 | 22, 24 |
| Python | 3.12, 3.13, 3.14 (also 3.11) | 3.12, 3.13, 3.14 | 3.12, 3.13, 3.14 |
| .NET | 8, 9, 10 (isolated worker) | 8, 10 (9: container only) | 10 (8 is deprecated on 2026-11-10) |

Offered: `node24` (suggested, the Active LTS), `node22`; `python313` (suggested), `python312`,
`python314`; `dotnet10` (suggested, the current LTS), `dotnet8`. Python has no LTS: 3.13 is the
suggested one because all three platforms support it and it has the widest wheel coverage; 3.14
is GA everywhere but newest. `dotnet8` conflicts with `gcp-functions`. Node 20 and .NET 9 are not
offered (Node 20 left every platform in 2026; .NET 9 is not on AWS and Google). The runtime files
are data; a new version is a data change.

### Tasks, tools, pipelines

Tasks follow the language contract: `build`, `test`, `lint`, `start`, `package`. `start` is
`func start`, `sam local start-api` (or `sam local invoke` for a trigger that is not HTTP) or the
functions-framework; `package` writes a zip (Node, Python) or the publish folder (C#) under the
workspace `dist/`. The build goes to the workspace's `dist/` like every build (`smith doctor`
refuses a build inside the source tree). A host wants the code and its manifest side by side, so
Node copies `package.json` and `host.json` next to the compiled code and runs the host there; Python
stages the deployed files in `dist/apps/<name>/app` (`build`) and zips them (`package`). Each template unit-tests the handler without any cloud
tool. `smith doctor` warns (never fails) when the host's local tool (`func`, `sam`, the
functions-framework) is missing and says how to get it; the install-the-tools offer (#887) will
hook there.

AWS and Google get templates only: no CI and no deploy pipeline in v2. Azure keeps
`smith pipeline deploy-azure-function`, unchanged; the Azure docs and next-step text name it.

### No new flags

The brief named `--trigger` and `--runtime`. They are not added: the trigger and the runtime are
axes, so the answer is part of `--with` (`--with azure-functions,http-trigger,node24`), the
refusal names exactly those words, and a future axis (a region, a plan) needs no new flag.

### Shared words

The tool of each host and the doctor finding are in `packages/plugin-sdk/functionhost`, so the
three languages say the same thing.

## Consequences

- No new plugin, no new release unit. Each language plugin grows one `functions` slice.
- Small core change: language-keyed names, `suggested`, a typed "needs an axis" error, the
  terminal asker. Protocol stays at version 1 (the new member is optional).
- Every template was scaffolded and run through `smith run lint typecheck test build package` with
  the real toolchains (Node 24, Python 3.12, .NET 10). The cloud tools (`func`, `sam`) were not
  installed, so `start` was checked by import and by serving the function with the Functions
  Framework, not with `func` or `sam`. `package.mjs` installs production dependencies with `npm
  install` in a staging folder, which fails for a dependency that is a library of the workspace.

## Sources

- Azure Functions language support: https://learn.microsoft.com/en-us/azure/azure-functions/supported-languages
- Azure Functions runtime versions: https://learn.microsoft.com/en-us/azure/azure-functions/functions-versions
- AWS Lambda runtimes: https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html
- Cloud Run functions runtime support: https://docs.cloud.google.com/functions/docs/runtime-support
- Issue #898 and ADR 0045 (choices as data).
