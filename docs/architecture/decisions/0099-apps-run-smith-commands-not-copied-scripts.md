# ADR 0099: Function and Electron apps run smith's commands, not copied scripts

- Status: accepted
- Date: 2026-11-12

## Context

smith copied `scripts/assemble.mjs`, `package.mjs` and `start.mjs` into every Node function app
(Azure Functions, AWS Lambda, Google Cloud) and `assemble.mjs`, `start.mjs` and `package.mjs` into
every Electron app. A fix to a script never reached an app made earlier, and the copies failed the
lint of the app. The .NET function apps already ran smith's commands and the host's tool.
Owner decision D3 of the plugin audit: no project holds a copy of a smith script.

## Decision

- The shared logic is in the plugin SDK package `functionhost`: `AssembleFunctionApp`,
  `PackFunctionApp` (vendors the libraries of the workspace, installs the production dependencies
  when asked, writes the zip in Go) and `AssembleElectronApp`. One implementation serves Azure,
  AWS Lambda, Google Cloud and Electron.
- The host has two hidden commands that the tasks call: `smith func assemble`,
  `smith func package [--with-dependencies]` and `smith electron assemble`. They find the build
  folder from the folder they run in, so a moved project needs no edit.
- `start` runs the tool of the host (`func start --script-root`, `sam local`,
  `functions-framework`, `electron`) from the build folder. `package` of Electron runs
  `electron-forge package` there. No script stands between.
- The plugins write the three tasks into `package.json` from `functionhost` (`AssembleTask`,
  `StartTask`, `PackageTask`), so one text serves the scaffold and the migration.
- `smith doctor` reports a copy that is still in an app. Its fix removes a copy as smith wrote it
  (found by a fingerprint of the script without its filled-in strings) and points the scripts at
  the commands. `smith upgrade` applies that fix. A copy that was edited is kept until the person
  chooses, in a question of the fix.

## Consequences

- A fix to the packaging reaches every app with the next smith.
- The `bestzip` dependency of a function app is not used any more. An existing app keeps it in
  `package.json`: smith does not touch the lock file.
- A task that starts `func` or `sam` needs that tool on the machine. The doctor of the functions
  plugin checks it, and the script check no longer asks for an npm package of that name.
- Python function apps follow the same rule. Their tasks are inferred, so no `package.json` is
  edited: `build` runs `smith func stage --host <host> --python <venv python>`, `package` runs
  `smith func zip`, and `start` runs `func start` (with the venv's interpreter in
  `languageWorkers__python__defaultExecutablePath`), `sam local` (after the build) or the
  Functions Framework. The libraries of the workspace are found by a short script that the venv's
  python runs (`importlib.metadata`, the `direct_url.json` that pip writes); the copying and the
  zip are Go in `functionhost`. The doctor removes `package.py` and `start.py` by the same rules.
- `smith func assemble` copies `local.settings.json` into the build folder when the project has
  one (Azure `func start --script-root`). The build folder is `dist/`, which `.gitignore` covers.
