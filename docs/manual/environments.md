# smith environments

A web app built with Vite (React, Vue, Svelte or a plain page) reads its `VITE_` settings when it
is built, not when it runs. One build cannot serve your machine, dev, uat and production. smith
therefore builds a web app once per environment: each environment has its own `.env.<name>` file,
its own build task and its own output folder. `smith add` asks which environments the app has, and
`smith environments` lists, adds and removes them afterwards. See
[ADR 0081](../architecture/decisions/0081-a-web-app-builds-once-per-environment-and-overlap-fixes-are-presented-choices.md).

## When to use it

- You add a web app and want one build per environment (`smith add app web --with react --environments lcl,dev,uat,prd`).
- You need another environment later (`smith environments add web qa`), or drop one you do not use.
- You want to know which task builds which environment (`smith environments list`).

## What `smith add` writes for a web app

For the environments `lcl, dev, uat, prd` of an app `web`:

| What | Where | Why |
|---|---|---|
| `.env.lcl`, `.env.dev`, `.env.uat`, `.env.prd` | `apps/web/` | Vite loads `.env.<mode>` for `vite build --mode <name>`. Starter `VITE_API_URL=` and a header: values with the `VITE_` prefix end up in the browser bundle, so never put a secret here. |
| `vite.build.ts` | `apps/web/` | `build(mode)` returns the Vite build options of an environment. `lcl` keeps sourcemaps and does not minify; `dev` does not minify; `uat` and `prd` minify. |
| `vite.config.ts` | `apps/web/` | Calls `build(mode)`. |
| `.gitignore` | `apps/web/` | Keeps `.env.*.local` (your private overrides) out of git. |
| scripts and tasks | `apps/web/package.json` | `build` is the default environment; `build:lcl`, `build:dev`, `build:uat` are the others. Each runs `vite build --mode <name> --outDir <folder>`. |
| `environments`, `default_environment`, `output`, `env_files` | `smith.toml`, `[projects.web.settings.vite]` | What smith needs to add and remove environments later. |
| `environments` | `smith.toml`, `[workspace]` | The list the next web app starts from, when the workspace has none yet. |

The `lcl` environment also starts the dev server: `smith run start -p web` runs `vite --mode lcl`.

## The questions

Each answer is a presented choice with its default marked. In a terminal `smith add` asks them in
turn. In the terminal interface and in VS Code the add form asks the same questions. Without a
terminal, or with `--yes`, only `--environments` is required; the other answers take the default
and the plan shows them.

| Question | Flag | Default | What it costs |
|---|---|---|---|
| Which environments? Names are lowercase words, pre-filled and editable. | `--environments lcl,dev,uat,prd` (`none` for one build, as before) | `lcl, dev, uat, prd` | Each environment is one more build and one more `.env.<name>` file. |
| Which environment is `smith run build`? | `--default-environment prd` | `prd`, else the last | CI and the release build this one. The others are `build:<name>`. |
| Where do the builds go? | `--environment-output workspace` or `project` | `workspace`: `dist/apps/web/<name>` | `project` puts them in `apps/web/dist/<name>`. smith doctor accepts both. |
| Commit the `.env.<name>` files? | `--env-files commit` or `example` | `commit`: they hold no secret | `example` commits only `.env.example` and git ignores `.env.<name>`; each machine and CI job then makes its own files. |

`terser` is not the minifier. It needs the extra dependency `terser`, and `minify: true` uses the
minifier Vite ships with. To use terser, run `smith install -p web -D terser` and set
`minify: 'terser'` in `vite.build.ts`.

## Build

```sh
smith run build:uat -p web                 # one environment
smith run build build:lcl build:dev build:uat -p web   # all of them: build is the default one
```

`smith run build:uat` reads `apps/web/.env.uat` and writes `dist/apps/web/uat`. Each environment is
cached on its own, and no two tasks write the same folder, so all of them can run at once in one
run (see [run](run.md#files-and-environment)).

## Secrets

A `VITE_` value is written into the bundle: anyone who opens the site can read it. smith doctor
checks that each environment has its `.env.<name>` file, and warns when a `VITE_` name has `KEY`,
`SECRET` or `TOKEN` in it (a public key is fine, a private key is not). A secret the build needs goes
through `[ci] secrets` in `smith.toml`, never into a `VITE_` value.

## CI and releases

The pipelines smith generates build the default environment, the task `build`. A release asset is
made from the folder that task writes. To release another environment, name its task in the
project's release `build` in `smith.toml`. To build every environment in CI, add a step that runs
`smith run build build:lcl build:dev build:uat`.

## An app that exists

`smith environments add` works on an app that was made with environments. For an app that has one
build, move it by hand:

1. Copy `vite.build.ts` from a new app (`smith add app tmp --with react --environments lcl,dev,uat,prd`).
2. Make `vite.config.ts` call it: `defineConfig(({ mode }) => ({ ..., build: build(mode) }))`.
3. Write one script per environment: `vite build --mode <name> --outDir <folder>`, the default one as
   `build`.
4. Create `.env.<name>` for each environment.
5. Record them in `smith.toml`: `[projects.web.settings.vite]` with `environments`,
   `default_environment`, `output` and `env_files`.

## Synopsis

```sh
smith environments list [project]
smith environments add <project> <name>... [--dry-run] [--yes] [--commit]
smith environments remove <project> <name>... [--dry-run] [--yes] [--commit]
```

`add` and `remove` show a plan first: the files they create, the scripts they change and the line
in `smith.toml`. `remove` leaves the `.env.<name>` file, which holds your values. The default
environment cannot be removed.

## Subcommands

<!-- smith:begin environments -->
List, add or remove the environments a web app is built for

Usage: `smith environments <command>`

| Subcommand | Meaning |
|---|---|
| `smith environments add` | Add environments to a web app |
| `smith environments list` | List the environments of the web apps and the task that builds each |
| `smith environments remove` | Remove environments from a web app (its .env files stay) |

No flags.

<!-- smith:end environments -->

Each subcommand has a page: [`list`](environments-list.md), [`add`](environments-add.md),
[`remove`](environments-remove.md).
