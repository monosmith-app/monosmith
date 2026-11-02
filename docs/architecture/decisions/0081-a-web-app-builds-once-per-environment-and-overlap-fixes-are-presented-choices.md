# ADR 0081: A web app builds once per environment, and an output overlap is fixed by a presented choice

- Status: accepted
- Date: 2026-11-02

## Context

A React app built with Vite reads its settings (`VITE_*` variables) when it is built, not when it
runs. One build therefore cannot serve the local machine, dev, uat and production. The owner built
one build per environment by hand: a `vite.build.ts` with a function `build(mode)` that returns the
Vite build options of a mode, and one task per mode. smith then warned "no two tasks write the same
output directory" for tasks that were not wrong (`models:watch` is never run with `models:build`),
and for tasks that were wrong (`api:build` and `api:build-dev` clear each other). The text gave two
fixes in one long sentence and did not say what goes wrong in words a person can act on.

## Decision

### Part A: the doctor finding

- The finding text says which tasks overlap, which folder, and what goes wrong: smith keeps each
  task's output in its cache; restoring one task first clears its output folder, which deletes the
  files the other task wrote there; a task that runs clears its folder too.
- A task that runs for a long time (the name has the word `watch`, `serve` or `start`, or is `dev`)
  is not compared: it is never run in the same run as a build. A name such as `build-dev` is a build
  and is compared.
- The fix is a presented choice (`smith doctor --fix --choose N`):
  1. **Give each task its own folder** (suggested). smith shows the exact `outputs` edit for
     smith.toml and, for a task that runs an npm script, the script to change; the fix runs
     `smith explain` on the task so the command is on screen. smith does not edit the command: the
     command belongs to the tool or the plugin, and an `outputs` line that differs from what the
     command writes would make the cache store an empty folder.
  2. **These tasks never run together: tell smith.** `[workspace] alternative_outputs =
     [["api:build", "api:build-dev"]]`. A group lists tasks that are alternatives of each other.
     Doctor, `smith ci generate` and the assets run do not report an overlap inside one group.
     The cost: smith still clears the folder before each task and a cache restore still replaces it,
     so running two tasks of one group in one run leaves only the last one's files. smith does not
     stop that run.
- `shared_outputs` is not reused. It means that several tasks fill one folder on purpose with
  different files; smith then does not clear the folder at all, so a rebuild keeps stale files.
  Two alternative builds write the same file names and need the clearing.

### Part B: environment builds

- `[workspace] environments = ["lcl", "dev", "uat", "prd"]` is the list a new web app starts from.
  A project can keep its own list in `[projects.<name>.settings.vite] environments`. Names are
  lowercase words (letters, digits, `-`), pre-filled and editable.
- Adding a web app (a Vite scaffold: react, vue, svelte, html) asks the environments, the default
  environment (default `prd`, else the last), where the output goes and whether `.env.<env>` is
  committed. Every answer is a flag too: `--environments lcl,dev,uat,prd` (or `none` for a single
  build), `--default-environment`, `--environment-output workspace|project`, `--env-files
  commit|example`. Without a terminal `--environments` is required and the add stops naming it.
  The plan lists the files before anything is written. smith commits nothing unasked.
- Per environment the plugin writes `.env.<env>` (starter `VITE_` lines and a header that says the
  values end in the browser bundle and a secret never goes there), one `build(mode)` function in
  `vite.build.ts` and a `vite.config.ts` that calls it. `lcl` keeps sourcemaps and does not minify,
  `dev` does not minify, `uat` and `prd` minify with the minifier Vite ships with (`minify: true`). `terser` is not
  the default: it needs the extra dependency `terser`. `minify: 'esbuild'` is not used either: Vite 8
  does not bundle esbuild, so the build fails without it.
- The default environment is the task `build` (so lint, pack, assets and `^build` keep working); the
  other environments are tasks `build:<env>`. No task `build:<default>` exists, because two tasks
  would write one folder, which Part A reports. Build all: `smith run build build:lcl build:dev
  build:uat`. Each task runs `vite build --mode <env> --outDir <folder>`.
- The folder is in the script, not only in `vite.build.ts`: smith reads a task's `outputs` from the
  script's `--outDir`, so the cache and the build cannot disagree. The location is a choice:
  `dist/apps/<app>/<env>` in the workspace (default, as the owner did) or `apps/<app>/dist/<env>`
  in the project. The choice is recorded in `[projects.<name>.settings.vite] output`.
- `.env.<env>` is committed by default (it holds no secrets). With `--env-files example`, smith
  writes the files and an `.env.example`, and ignores `.env.<env>`. `.env.*.local` is always
  ignored. A real secret goes through `[ci] secrets`, never a `VITE_` value.
- Doctor checks that every environment has its `.env.<env>` file, and warns about a `VITE_` name
  with KEY, SECRET or TOKEN in it.
- CI and release assets build the default environment (`build`). A project whose release `build`
  names another task builds that one. The release asset is the folder the `build` task writes.
- `smith environments list|add|remove <project>` changes an existing project and shows a plan.

## Consequences

- A web app has several cached builds with no shared folder, so the Part A warning stays quiet.
- The default environment's folder is a sub-folder (`dist/apps/web/prd`), so code that assumed the
  build folder is `dist/apps/<app>` reads the folder of the `build` task instead.
- A user who wants a `terser` build adds the dependency and sets the minifier in `vite.build.ts`.
