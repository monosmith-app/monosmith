# CLI reference

Every `smith` command and flag, grouped by what you want to do. Each command
has one example whose output was copied from a real run in a scratch workspace
(a Go app and lib, two Node projects and a Python package, as in
[Getting started](../guides/getting-started.md)). Paths and timings will differ
on your machine, and where an output is long it is cut with `...` and said so.

> **Keeping this page current.** The usage line and flag table of each command
> sit between `<!-- smith:begin <command> -->` and `<!-- smith:end <command> -->`
> markers and are generated from `smith commands --json`. Everything else is
> written by hand. After changing a command, build smith and run, from the
> repository root:
>
> ```sh
> node docs/reference/cli-generate.mjs           # rewrite the blocks
> node docs/reference/cli-generate.mjs --check   # fail if the page is out of date
> ```
>
> The script names the binary with `SMITH` (default: `smith` on the PATH) and
> fails when a command has no block, so a new command cannot be forgotten.
> `smith <command> --help` always has the last word.

## Contents

- [Everywhere](#everywhere): flags and conventions every command shares
- [Create and change a workspace](#create-and-change-a-workspace): `new`, `adopt`, `convert`, `upgrade`, `add`, `remove`, `config`
- [Work on projects](#work-on-projects): `run`, `build`, `test`, `lint`, `typecheck`, `dev`, `install`, `uninstall`, `sync`, `up`
- [Look at the workspace](#look-at-the-workspace): `projects`, `graph`, `affected`, `explain`, `why-miss`, `doctor`, `info`, `kinds`, `stacks`, `slices`, `schema`, `commands`, `mcp`, `plugin`, `recipe`
- [Ship](#ship): `ci`, `pipeline`, `docker`, `release`
- [Shell](#shell): `completion`, `help`, and `smith` on its own

## Everywhere

```sh
$ smith --version
```

```text
smith version dev
```

(`dev` is a development build; a release prints its number.)

| Flag | Meaning |
|---|---|
| `-h`, `--help` | help for `smith` or for any command: `smith <command> --help` |
| `-v`, `--version` | print the version |

- **Workspace.** Almost every command works on the workspace it finds by looking
  for `smith.toml` in the current folder and then in each parent.
  `SMITH_WORKSPACE_ROOT` overrides that. Outside one, `smith` says so and
  suggests `smith new`.
- **`--json`.** Commands that report something print a stable JSON document
  with `--json`, for editors and scripts. See [Editors](../guides/editors.md).
- **Environment variables** are listed in [configuration](configuration.md#environment-variables).
- **Exit codes.** 0 on success; non-zero when a check fails, a task fails, or
  something blocks.
- **Project selection** (`-p`): a name, a glob over names or directories
  (`apps/*`), `kind:app`, `lang:go`, `tag:x`, and a leading `!` to exclude.
  Comma-separated or repeated.

## Create and change a workspace

### smith new

Make a workspace: `smith.toml`, `.gitignore`, `.gitattributes`, a README, the CI pipeline and a dev container, then `git init` on `main` and a first commit.

<!-- smith:begin new -->
Create a monorepo: configuration, git, CI pipelines and a first commit

Usage: `smith new <name> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--ci <ci>` | pipelines to generate: github, azure (default github) (repeatable) | `github` |
| `--dir <dir>` | directory to create the workspace in (default ./<name>) | `./<name>` |
| `--module-prefix <module-prefix>` | Go module prefix (default from the origin remote, else example.com/<name>) | from the `origin` remote, else `example.com/<name>` |
| `--recipe <recipe>` | the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md) |  |
| `--scope <scope>` | npm scope for packages (default @<name>) | `@<name>` |

<!-- smith:end new -->

With `--dir .` in an existing clone, files smith does not own are kept and the first commit holds only the files smith wrote. `--ci` also accepts `gitlab`, although the flag's help text lists only `github` and `azure`.

```sh
$ smith new demo --module-prefix github.com/acme/demo --scope @demo
```

```text
  wrote .gitignore
  wrote .gitattributes
  wrote smith.toml
  wrote README.md
  wrote .github/workflows/ci.yml
  wrote .devcontainer/devcontainer.json
  committed

Created demo. Next:
  cd demo
  smith add app web --lang node
```

### smith adopt

Report what bringing an existing repository (npm workspaces, Go modules, Python projects) under smith involves, and take the steps. Read-only unless you ask for a step; both steps need a clean git tree.

<!-- smith:begin adopt -->
Bring an existing repository under smith: report what it involves, then take the steps (start here)

Usage: `smith adopt [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--dry-run` | only report (this is the default; the flag makes it explicit) | off |
| `--init` | write smith.toml and the pipelines, and commit | off |
| `--install` | with --init, also run npm install for the Node projects | off |
| `--json` | print the report as JSON | off |
| `--tags` | create the baseline release tags the report asks for, locally | off |

<!-- smith:end adopt -->

`--tags` creates, locally, the release tags a project needs where its tags were made under another name (nothing is pushed). `--init` writes `smith.toml` and the pipelines and commits; a pipeline of yours at a path smith owns is kept beside it as `<file>.pre-smith`. An Nx workspace is for `smith convert`. See [Getting started](../guides/getting-started.md#or-adopt-an-existing-repository).

```sh
$ smith adopt
```

```text
C:\work\old3 would be workspace "old3" (scope @old3, CI github) with:
  lib                  node     package  packages/lib

✔ nothing blocks: smith adopt --init makes it a workspace
```

### smith convert

Turn an Nx workspace (or one made by the old Nx-based mnci) into a smith workspace, in place. Run it with `--dry-run` first. See [Converting from Nx](../guides/converting.md).

<!-- smith:begin convert -->
Convert an Nx, Turborepo or mnci workspace to smith (adopt sends you here when it recognises one)

Usage: `smith convert [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--allow-dirty` | convert even with uncommitted changes | off |
| `--dry-run` | report the plan without changing anything | off |
| `--json` | print the plan as JSON | off |
| `--skip-install` | do not run npm install afterwards | off |
| `--strict` | refuse to convert while anything needs manual work or is unsupported | off |
| `--verbose` | name every project a finding applies to, not just a few | off |
| `--yes` | apply without asking (a terminal asks before changing anything) | off |

<!-- smith:end convert -->

```sh
$ smith convert --dry-run
```

```text
smith: smith.toml already exists: this workspace is already a smith workspace
```

That is the real answer in a workspace that already has a `smith.toml`. This repository had no Nx workspace to convert, so a conversion report is not shown here; [Converting from Nx](../guides/converting.md) describes it.

### smith upgrade

After installing a newer smith, migrate `smith.toml` to the current schema (comments and layout kept), regenerate the files smith owns (the pipelines), and bring recipe-made projects up to date with their recipe.

<!-- smith:begin upgrade -->
Bring the workspace up to this smith and its recipes: migrate smith.toml, regenerate the files smith owns

Usage: `smith upgrade [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--check` | change nothing; fail when something would change | off |
| `--dry-run` | show the plan, with the change to smith.toml, and change nothing | off |
| `--json` | print the result as JSON | off |

<!-- smith:end upgrade -->

```sh
$ smith upgrade --dry-run
```

```text
The workspace is up to date.
```

### smith add

Add a project. The kind says what happens to it: `app` (deployed, never published, in `apps/`), `lib` (imported by other projects, never published, in `libs/`) or `package` (published on release, in `packages/`).

<!-- smith:begin add -->
Add an app, an internal lib or a publishable package

Usage: `smith add <kind> <name> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--choice <choice>` | answer a choice of the workspace's recipe, name=answer (repeatable); the rest are asked in a terminal (repeatable) |  |
| `-l, --lang <lang>` | language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, recipe (default: the stack's, or the workspace's only language) | the stack's, or the workspace's only language |
| `-s, --stack <stack>` | stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone) | the recipe, else `[default_stacks]` in `smith.toml` |
| `--verbose` | show the package manager's output as it runs (it is shown only if it fails otherwise) | off |

<!-- smith:end add -->

A combined kind from `smith kinds` names the language as well: `smith add node-app web` is `smith add app web --lang node`. When `smith.toml` names a recipe, the recipe creates the project and asks its choices; `--choice name=answer` answers one without asking, and `--stack` bypasses the recipe. The new project lints, type-checks, tests and builds straight away, and the CI pipelines are regenerated to install its toolchain.

```sh
$ smith add lib core --lang go
```

```text
go: creating new go.mod: module github.com/acme/demo/libs/core
  libs/core/greeting/greet_use_case.go
  libs/core/greeting/greet_use_case_test.go
  libs/core/greeting/greeting_contract.go
Import it from another Go project as github.com/acme/demo/libs/core/<slice>.
Try: smith test -p core
```

Only the last lines are shown; earlier lines list every file written.

### smith remove

Remove a project: delete its directory, let its plugin clean what it owns outside it (the `go.work` entry, the npm lockfile, the Python venv), drop its tables from `smith.toml` and regenerate the pipelines.

<!-- smith:begin remove -->
Remove a project: its directory, its traces in the workspace, its CI

Usage: `smith remove <project> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--dry-run` | show what would change; change nothing | off |
| `--force` | remove a project other projects still import | off |
| `--json` | print what changed as JSON | off |
| `-y, --yes` | do not ask for confirmation | off |

<!-- smith:end remove -->

A project other projects import is refused unless `--force`. Its release tags are kept. Without `--yes` it asks first.

```sh
$ smith remove web --dry-run
```

```text
Would remove web (apps/web)
  changes package-lock.json
  would run npm install to update the lockfile
  the pipelines would be regenerated
```

### smith config

Read and change the settings of `smith.toml`'s `[workspace]`, `[ci]`, `[release]`, `[cache]` and `[plugins]` tables. Values are checked against the schema, and only the setting's value changes in the file: comments and every other line stay as they are. Stacks, conventions, tasks and projects are edited in the file itself.

<!-- smith:begin config -->
Read and change settings of smith.toml

Usage: `smith config`

No flags.

<!-- smith:end config -->

| Subcommand | Does |
|---|---|
| `smith config list [--json]` | every setting, its value (or `(default)`) and what it does |
| `smith config get <setting>` | one setting's value as written in `smith.toml` |
| `smith config set <setting> <value...>` | change one setting; a list takes several values |
| `smith config unset <setting>` | remove a setting so it takes its default |

```sh
$ smith config set release.branch trunk
$ smith config get release.branch
$ smith config unset release.branch
$ smith config get release.branch
```

```text
✔ release.branch = "trunk"
"trunk"
✔ release.branch removed: it has its default
smith: release.branch is not set; it has its default (see https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md)
```

`smith config list` (the description column is left out here, and the list is cut after six settings):

```text
workspace.exclude         (default)
workspace.module_prefix   "example.com/acme"
workspace.name            "acme"
workspace.recipe          (default)
workspace.registry        (default)
workspace.scope           "@acme"
...
```

All settings are explained in [configuration](configuration.md).

## Work on projects

### smith run

Run tasks on projects, in dependency order, from the cache when nothing changed. `smith build`, `test`, `lint` and `typecheck` are `smith run` with the task name filled in and take the same flags.

<!-- smith:begin run -->
Run tasks on projects, in dependency order, from cache when nothing changed

Usage: `smith run <task> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | only projects affected by changes since --base | off |
| `--bail` | stop starting tasks after the first failure | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print the summary as JSON | off |
| `--json-events` | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else | off |
| `--no-cache` | run every task even if a cached result exists | off |
| `--parallel <parallel>` | tasks to run at once (default: CPU count) | the number of CPUs |
| `-p, --projects <projects>` | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude (repeatable) |  |

<!-- smith:end run -->

`--json-events` prints the run as [JSON lines](serve.md#the-event-stream), one event each, and nothing else (it cannot be combined with `--json`).

Tasks run in parallel, one per CPU by default. A failed task skips the tasks that depend on it; independent ones keep going unless `--bail`. A cache hit prints `✔ from cache`. See [Tasks, cache and affected](tasks-and-cache.md).

```sh
$ smith run test -p core
```

```text
core:test │ ?   	example.com/acme/libs/core	[no test files]
core:test │ ok  	example.com/acme/libs/core/greeting	0.449s
core:test ✔ from cache
✔ 1/1 tasks succeeded (1 from cache, 1.024s saved) in 8ms
```

```sh
$ smith run lint test --affected      # only what changed since main
$ smith build -p kind:package         # every publishable package
```

### smith build

Same as `smith run build`; a project name may follow instead of `-p`.

<!-- smith:begin build -->
Run the build task (same as `smith run build`)

Usage: `smith build [project] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | only projects affected by changes since --base | off |
| `--bail` | stop starting tasks after the first failure | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print the summary as JSON | off |
| `--json-events` | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else | off |
| `--no-cache` | run every task even if a cached result exists | off |
| `--parallel <parallel>` | tasks to run at once (default: CPU count) | the number of CPUs |
| `-p, --projects <projects>` | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude (repeatable) |  |

<!-- smith:end build -->

```sh
$ smith build -p core,api
```

```text
core:build ✔ from cache
api:build ✔ from cache
✔ 2/2 tasks succeeded (2 from cache, 1.064s saved) in 96ms
```

### smith test

Same as `smith run test`; a project name may follow instead of `-p`.

<!-- smith:begin test -->
Run the test task (same as `smith run test`)

Usage: `smith test [project] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | only projects affected by changes since --base | off |
| `--bail` | stop starting tasks after the first failure | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print the summary as JSON | off |
| `--json-events` | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else | off |
| `--no-cache` | run every task even if a cached result exists | off |
| `--parallel <parallel>` | tasks to run at once (default: CPU count) | the number of CPUs |
| `-p, --projects <projects>` | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude (repeatable) |  |

<!-- smith:end test -->

```sh
$ smith test ui
```

```text
ui:test ▶ npm run test --silent
ui:test │ Test Suites: 1 passed, 1 total
ui:test │ Tests:       1 passed, 1 total
ui:test │ Snapshots:   0 total
ui:test │ Time:        1.945 s
ui:test │ Ran all test suites.
ui:test ✔ 4.502s
✔ 1/1 tasks succeeded (0 from cache) in 4.534s
```

### smith lint

Same as `smith run lint`; a project name may follow instead of `-p`.

<!-- smith:begin lint -->
Run the lint task (same as `smith run lint`)

Usage: `smith lint [project] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | only projects affected by changes since --base | off |
| `--bail` | stop starting tasks after the first failure | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print the summary as JSON | off |
| `--json-events` | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else | off |
| `--no-cache` | run every task even if a cached result exists | off |
| `--parallel <parallel>` | tasks to run at once (default: CPU count) | the number of CPUs |
| `-p, --projects <projects>` | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude (repeatable) |  |

<!-- smith:end lint -->

```sh
$ smith lint -p ui
```

```text
ui:lint ▶ npm run lint --silent
ui:lint ✔ 4.164s
✔ 1/1 tasks succeeded (0 from cache) in 4.17s
```

### smith typecheck

Same as `smith run typecheck`; a project name may follow instead of `-p`.

<!-- smith:begin typecheck -->
Run the typecheck task (same as `smith run typecheck`)

Usage: `smith typecheck [project] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | only projects affected by changes since --base | off |
| `--bail` | stop starting tasks after the first failure | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print the summary as JSON | off |
| `--json-events` | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else | off |
| `--no-cache` | run every task even if a cached result exists | off |
| `--parallel <parallel>` | tasks to run at once (default: CPU count) | the number of CPUs |
| `-p, --projects <projects>` | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude (repeatable) |  |

<!-- smith:end typecheck -->

```sh
$ smith typecheck -p ui
```

```text
ui:typecheck ▶ npm run typecheck --silent
ui:typecheck ✔ 1.255s
✔ 1/1 tasks succeeded (0 from cache) in 1.259s
```

### smith dev

Start several projects together, such as a frontend and the API it calls. What they depend on is built first, then every `start` task runs at once with its output prefixed by the project's name; Ctrl+C stops them all.

<!-- smith:begin dev -->
Start several projects together, such as a frontend and the API it calls

Usage: `smith dev [projects...] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--all` | start every project that has a start task | off |
| `--dry-run` | print the equivalent smith run command and start nothing | off |

<!-- smith:end dev -->

Name the projects, pass `--all` for every project with a `start` task, or pick from a list in a terminal. It is `smith run start -p <projects>` with room for every one to run at once; `--dry-run` prints that command.

```sh
$ smith dev web api --dry-run
```

```text
→ smith run start -p web,api --parallel 6
```

### smith install

Add packages to one project with its own toolchain (`npm install --workspace`, the project's `pyproject.toml` and the workspace venv, `go get`), or install everything a fresh clone needs. Alias: `smith i`.

<!-- smith:begin install -->
Add packages to a project, or install the whole workspace (up updates, sync aligns versions)

Usage: `smith install [packages...] [flags]`

Alias: `smith i`

| Flag | Meaning | Default |
|---|---|---|
| `-D, --save-dev` | add as a development dependency (npm devDependencies, the Python dev extra, a Go tool) | off |
| `-w, --workspace <workspace>` | target project, by name or directory; repeat it for several (repeatable) |  |

<!-- smith:end install -->

Each project owns its dependencies, so a package needs a target: `-w <project>`. Without packages, `smith install` installs the whole workspace. Works for Node, Python, Go, Dart, PHP and Ruby projects; for the others use their own tool in the project's folder.

```sh
$ smith install -w ui zod  &&  smith uninstall -w ui zod
```

```text

up to date in 1s
✔ ui: added zod (libs/ui/package.json, package-lock.json)
---

up to date in 1s
✔ ui: removed zod (libs/ui/package.json, package-lock.json)
```

Only the last lines of each are shown (npm prints more).

### smith uninstall

Remove packages from one project; the counterpart of `smith install`. Alias: `smith un`.

<!-- smith:begin uninstall -->
Remove packages from a project (the opposite of install)

Usage: `smith uninstall <packages...> [flags]`

Alias: `smith un`

| Flag | Meaning | Default |
|---|---|---|
| `-w, --workspace <workspace>` | target project, by name or directory; repeat it for several (repeatable) |  |

<!-- smith:end uninstall -->

See the example under `smith install`.

### smith sync

Make every project declare a shared dependency at the same version, language by language. The target is the version installed, with the operator most declarations use; otherwise the highest version declared. Then it installs, so lockfiles and the venv follow.

<!-- smith:begin sync -->
Make every project declare a shared dependency at the same version (same as up --align)

Usage: `smith sync [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--check` | report only; fail when declarations disagree | off |
| `--json` | print the report as JSON | off |
| `--no-install` | rewrite the manifests without installing | off |

<!-- smith:end sync -->

Peer dependencies, the workspace's own packages and specs smith cannot safely rewrite (URLs, git refs, compound ranges) are left alone and reported. `smith sync --check` only reports and fails when declarations disagree: use it in CI.

```sh
$ smith sync --check
```

```text
✔ Every project agrees on every shared dependency version.
⊘ go: not compared; its plugin does not list declarations (a go.work builds every module with the highest version any of them requires, so differing requirements never ship two copies)
```

### smith up

List every declared dependency with a newer release, grouped as patch, minor and major with the projects that declare it, then update the ones you pick.

<!-- smith:begin up -->
Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)

Usage: `smith up [packages...] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--align` | bring dependencies declared at different versions to one (the named ones, or all); with --check, fail when any differ | off |
| `--check` | report only; never prompt, never write | off |
| `--json` | print the updates and tools as JSON; never write | off |
| `--no-install` | rewrite the manifests without installing | off |
| `-y, --yes` | take every update without prompting | off |

<!-- smith:end up -->

Each chosen declaration keeps its operator (`^`, `~`, `>=`) and moves to the latest version; then the workspace is installed again. Named packages update without a prompt, `--yes` takes everything, and without a terminal (or with `--check`) it only reports. A last section compares the tools smith pins with their newest releases.

```sh
$ smith up --check
```

```text
patch
  @monosmith/eslint-config     1.0.0 → 1.0.2  (root)
minor
  eslint                       10.12.0 → 10.13.0  (root)
major
  @types/node                  24.19.2 → 26.6.5  (root)
  typescript                   6.0.3 → 7.0.2  (root)

Tools smith pins (informational):
  Go               1.27 ✔ (latest 1.27.2)
  golangci-lint    2.14.0 → 2.15.0 available
  Node.js (LTS)    24 ✔ (latest 24.22.0)
```

Needs network access to the registries.

### smith deps

Work on dependencies across the workspace. See [One version per dependency](configuration.md#dependencies).

<!-- smith:begin deps -->
Group for dependency work across projects; deps up is the same as up

Usage: `smith deps`

No flags.

<!-- smith:end deps -->

| Subcommand | Does |
|---|---|
| `smith deps up [packages...] [--check] [--yes] [--json] [--no-install]` | the same as `smith up`: a named package is updated in every project that declares it, in one change |
| `smith deps up --align [packages...] [--check] [--json] [--no-install]` | bring a dependency that projects of one ecosystem declare at different versions to one (the installed version, else the highest declared), instead of updating. The packages named, or every fragmented one; the ones `[dependencies] allow` lists are left alone. With `--check` it only reports and fails when any differ |

```sh
$ smith deps up axios
$ smith deps up --align
```

## Look at the workspace

### smith projects

List the workspace's projects: name, kind, language and directory.

<!-- smith:begin projects -->
List the workspace's projects

Usage: `smith projects [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end projects -->

```sh
$ smith projects
```

```text
api                  app      go      apps/api
core                 lib      go      libs/core
sdk                  package  python  packages/sdk
ui                   lib      node    libs/ui
web                  app      node    apps/web
```

### smith graph

Show how projects depend on each other, as a tree (`text`), `json`, Graphviz `dot`, a static `html` page, or an interactive browser (`tui`). Every format shows the same projects; `--focus` and `--affected` narrow all of them the same way.

<!-- smith:begin graph -->
Show how projects depend on each other

Usage: `smith graph [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--affected` | keep the projects affected by changes since --base; mark the changed ones | off |
| `--base <base>` | base branch for --affected (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--blast-radius <blast-radius>` | list everything a change to a project or a file reaches, across languages and through shared schemas; text or json |  |
| `--diff <diff>` | show how the graph changed since a revision (a branch, tag or commit): projects and imports added and removed, kinds and stacks changed; text or json |  |
| `--focus <focus>` | keep these projects and everything connected to them (names, globs, tag:x) (repeatable) |  |
| `--format <format>` | text, json, dot, html or tui | `text` |
| `-o, --output <output>` | write to a file |  |

<!-- smith:end graph -->

`--diff <revision>` discovers the graph at another revision in a temporary git worktree and lists projects and imports added (`+`) and removed (`-`) and kinds, stacks, plugins and directories that changed (`~`). Projects that import each other in a cycle are reported with the path.

`--blast-radius <project|file>` lists everything a change to a project or a file reaches, across languages and through [shared schemas](tasks-and-cache.md#shared-schemas); `--format json` prints it as data.

```sh
$ smith graph
```

```text
api  app · go · apps/api
  └─ core
core  lib · go · libs/core
sdk  package · python · packages/sdk
ui  lib · node · libs/ui
web  app · node · apps/web
  └─ ui
```

```sh
$ smith graph --focus web
```

```text
ui  lib · node · libs/ui
web *  app · node · apps/web
  └─ ui
```

The `*` marks the project you focused on.

```sh
$ smith graph --diff HEAD~2
```

```text
project graph from HEAD~2 to the working tree
  no change
```

### smith affected

List the projects affected by changes since a base branch, one per line: the ones that own a changed file and everything that depends on them. When smith cannot be sure (no base, a shallow clone, a changed `smith.toml` or lockfile) every project is affected.

<!-- smith:begin affected -->
List the projects affected by changes since a base branch

Usage: `smith affected [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--base <base>` | base branch (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print JSON, with the directories to watch for each app | off |
| `--paths` | print each app and the directories a deployment platform should watch | off |

<!-- smith:end affected -->

`--paths` prints, for each app, the directories a deployment platform should watch (the app's own and those of everything it depends on): for Railway or Render watch paths and GitOps path filters. `--json` prints the whole answer with a `watch` entry per app. `--paths` and `--json` list every app, affected or not. The base is `--base`, else `SMITH_BASE`, else the pull request target, else the release branch.

```sh
$ smith affected
```

```text
no project changed since main
```

Nothing changed since `main` in this workspace. The watch paths, one app per line (the app, a tab, then its directories):

```sh
$ smith affected --paths
```

```text
api	apps/api libs/core
web	apps/web libs/ui
```

### smith explain

Show how a project's tasks resolved and where each field came from: the language plugin, a stack's plugin, a convention, or `smith.toml`.

<!-- smith:begin explain -->
Show how a project's tasks resolved, and where each setting came from (why-miss says why one reran)

Usage: `smith explain <project> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |
| `--recipe` | print only what the project's recipe decided, and where | off |
| `--task <task>` | explain only this task (same as <project>:<task>) |  |

<!-- smith:end explain -->

For a project made from a recipe it also lists what the recipe decided for it; `--recipe` prints only that. Without a task name it lists every task of the project; the output here is cut after the first fields.

```sh
$ smith explain web:test
```

```text
web:test — Run unit tests
  command    npm run test --silent
               from plugin node
  cwd        apps/web
               from plugin node
  dependsOn  ^build
               from plugin node
  inputs     **/*, {workspaceRoot}/package-lock.json, {workspaceRoot}/tsconfig.base.json, {workspaceRoot}/eslint.config.mjs
               from plugin node
  outputs    —
               from plugin node
  env        —
               from plugin node
  cache      true
```

### smith why-miss

Say why a task would run instead of coming from the cache, by comparing its key with its last cached run. It runs nothing.

<!-- smith:begin why-miss -->
Say why a task would run again instead of coming from the cache (a cache miss; explain shows its settings)

Usage: `smith why-miss <project> <task> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end why-miss -->

It names added, removed and changed input files, a changed command or environment variable, upstream projects whose sources changed, dependency tasks whose keys changed, a new smith or plugin version, and a new toolchain. The argument is `<project>:<task>`.

```sh
$ smith why-miss core:test
```

```text
core:test would come from the cache: nothing changed since its last cached run.
```

After a change to `ui`, `smith why-miss web:test` answers `web:test would run: 1 change(s) since its last cached run` and `upstream   ui changed`.

### smith doctor

Check the workspace's health: smith's own invariants and each plugin's checks. Every problem names its remedy.

<!-- smith:begin doctor -->
Check the workspace's health; every problem names its remedy

Usage: `smith doctor [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--fix` | show the fixes doctor can make as diffs, and apply the safe ones | off |
| `--json` | print the report as JSON | off |
| `--only <only>` | with --fix, fix only the named check, as doctor prints it (repeatable) (repeatable) |  |
| `-y, --yes` | with --fix, apply without asking | off |

<!-- smith:end doctor -->

With `--fix`, doctor shows the fixes it can make as unified diffs, then applies the safe ones once you confirm (or with `--yes`): a safe fix only restores what smith or a plugin owns, like `go.work` or a generated pipeline with its slots kept. Others are shown, never applied. `--only "go.work exists"` takes up just the fix of that check (the interactive shell's Doctor tab uses it to apply one fix). Without a terminal `--fix` only shows. The checks and their remedies are in [Troubleshooting](../guides/troubleshooting.md#what-smith-doctor-reports-and-what-to-do).

```sh
$ smith doctor
```

```text
✔ the venv's Python is 3.12, as in CI

24 passed, 0 failed, 0 warning(s)
```

The last lines of a healthy workspace: 24 checks, one line each above them. A failing one starts with `✖` and is followed by a `→` line with the fix.

### smith info

Show the installed smith, the newest released one, and this workspace's settings. `SMITH_NO_UPDATE_CHECK=1` stops it asking GitHub for the newest release.

<!-- smith:begin info -->
Show the installed smith, the newest released one, and this workspace's settings

Usage: `smith info [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the report as one JSON document | off |

<!-- smith:end info -->

```sh
$ smith info
```

```text
smith dev
latest release: 1.30.0
workspace: /work/acme
```

### smith kinds

List the project kinds `smith add` accepts: one per language and kind (`node-app`, `go-lib`, `python-package`, ...). Not to be confused with the three kinds `app`, `lib` and `package`.

<!-- smith:begin kinds -->
List the project kinds smith add accepts (--json for editors)

Usage: `smith kinds [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the kinds as one JSON document | off |

<!-- smith:end kinds -->

```sh
$ smith kinds
```

```text
bicep-app        Bicep: an app: deployed or run, never published. Built with the workspace's default stack, or --stack.
bicep-lib        Bicep: an internal library: imported by other projects, never published. Built with the workspace's default stack, or --stack.
bicep-package    Bicep: a package: published to its registry on release. Built with the workspace's default stack, or --stack.
cpp-app          Cpp: an app: deployed or run, never published. Built with the workspace's default stack, or --stack.
```

Cut after four lines; there is one per language and kind (33 in all).

### smith stacks

List the stacks projects can be built with and which projects use each. A stack is the set of plugins a project is built with. Plugins publish stacks and `smith.toml` can declare more under `[stacks.<name>]`. See [Stacks](stacks.md).

<!-- smith:begin stacks -->
List the stacks projects can be built with, and who uses each

Usage: `smith stacks [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end stacks -->

```sh
$ smith stacks
```

```text
android  (plugin jvm)
  an Android app in Kotlin, built with Gradle (AGP), linted with ktlint and Android Lint
  language   jvm

angular  (plugin angular)
  an Angular app in TypeScript, built with the Angular CLI, linted by eslint, tested with Vitest
  language   node
  toolchain  angular, eslint, vitest
  scaffold   angular

docs-site  (plugin astro)
  a documentation site with Astro Starlight, checked and built by Astro, linted by eslint
```

Cut after twelve lines.

### smith slices

Check the vertical-slice layout of projects.

<!-- smith:begin slices -->
Check the vertical-slice layout of projects

Usage: `smith slices`

No flags.

<!-- smith:end slices -->

| Subcommand | Does |
|---|---|
| `smith slices check <project directory>... [--lang <language>]` | check that files are named for their role and no folder is a junk drawer. `--lang` is one of cpp, dart, dotnet, go, jvm, php, python, ruby, rust (default `go`) |

The vertical-slices convention runs this as each project's `slice-check` task. See [Conventions](conventions.md).

```sh
$ smith slices check libs/core
```

```text
vertical-slice layout OK
```

### smith schema

Print the JSON Schema of `smith.toml`, for editors. `smith new` writes its address as the first line of `smith.toml`.

<!-- smith:begin schema -->
Print the JSON Schema of smith.toml, for editors

Usage: `smith schema`

No flags.

<!-- smith:end schema -->

```sh
$ smith schema | head -n 6
```

```text
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://raw.githubusercontent.com/monosmith-app/monosmith/main/apps/smith/workspace/smith.schema.json",
  "title": "smith.toml",
  "description": "The configuration of a smith workspace (schema 1). Every setting has a default; a key smith does not know is an error.",
  "type": "object",
```

### smith commands

List every command with its arguments and options. With `--json` it is the contract editors read; see [Editors](../guides/editors.md).

<!-- smith:begin commands -->
List smith's commands with their arguments and options (--json for editors)

Usage: `smith commands [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the commands as one JSON document | off |

<!-- smith:end commands -->

```sh
$ smith commands --json | head -c 300
```

```text
[
  {
    "name": "add",
```

### smith mcp

Serve the workspace to an AI assistant over the Model Context Protocol, on stdin and stdout. Each tool runs the smith command it stands for. Register `smith mcp` as the server's command, started in the workspace. The tools are listed in [Editors](../guides/editors.md#ai-assistants-mcp).

<!-- smith:begin mcp -->
Serve the workspace to an AI assistant over the Model Context Protocol (stdio)

Usage: `smith mcp`

No flags.

<!-- smith:end mcp -->

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

### smith serve

Serve the engine over JSON-RPC on stdin and stdout, one JSON message per line, in the workspace smith is started in. It opens no network port. A method per function, and a run's events as notifications. The protocol is in [Serve and the event stream](serve.md).

<!-- smith:begin serve -->
Serve the engine to an editor or a terminal interface over JSON-RPC (stdio)

Usage: `smith serve`

No flags.

<!-- smith:end serve -->

```sh
$ echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | smith serve
```

### smith plugin

Inspect plugins, or check one you wrote. (`smith plugin serve <name>` also exists: smith starts each built-in plugin as a process with it. You do not run it yourself.)

<!-- smith:begin plugin -->
Inspect plugins, or serve a built-in plugin

Usage: `smith plugin`

No flags.

<!-- smith:end plugin -->

| Subcommand | Does |
|---|---|
| `smith plugin list [--json]` | the built-in plugins and the ones installed on the plugin search path, with their levels, slots and requirements |
| `smith plugin enable <name>...` | enable built-in plugins in `smith.toml` |
| `smith plugin disable <name>...` | disable built-in plugins in `smith.toml` |
| `smith plugin set <name>...` | enable exactly these built-in plugins |
| `smith plugin check <name \| path> [args...] [--generate] [--json]` | run the conformance suite against a plugin: it names itself, its roles match what it implements, its stacks are complete, and every method it advertises answers. `--generate` also scaffolds an app, a lib and a package in a temporary directory |

```sh
$ smith plugin list | head -8
```

```text
PLUGIN           LEVELS                       SLOTS                REQUIRES
bicep            language,toolchain,scaffold  -                    -
cpp              language,toolchain,scaffold  -                    -
dart             language,toolchain,scaffold  -                    -
dotnet           language,toolchain,scaffold  -                    -
go               language,toolchain,scaffold  -                    -
jvm              language,toolchain,scaffold  -                    -
node             language,toolchain,scaffold  -                    -
...
```

```sh
$ smith plugin disable bicep
$ smith plugin enable bicep
```

```text
enabled plugins: cpp, dart, dotnet, go, jvm, node, php, python, recipe, ruby, rust
enabled plugins: bicep, cpp, dart, dotnet, go, jvm, node, php, python, recipe, ruby, rust
```

`smith plugin check` is explained, with three working plugins, in [Writing a plugin](../guides/writing-a-plugin.md#testing-smith-plugin-check).

### smith recipe

Check recipes: a team's choice of plugins and settings. See [Recipes](recipes.md).

<!-- smith:begin recipe -->
Check recipes: a team's choice of plugins and settings

Usage: `smith recipe`

No flags.

<!-- smith:end recipe -->

| Subcommand | Does |
|---|---|
| `smith recipe check <file-or-url> [--choice name=answer]... [--all] [--json]` | load a recipe and the recipes it extends, answer its choices, and check the result: one plugin per exclusive level and slot, requirements met, no conflicts. Prints the choices, the plugin for each level and where each value came from. Non-zero exit when the recipe is not valid. `--all` checks every combination of answers (up to 64) |
| `smith recipe new [path] [--force] [--from-answers <file>]` | write a recipe by answering questions, checked for every combination before it is written. Without a terminal, `--from-answers` reads the answers from a file |

```sh
$ smith recipe check team.recipe.toml --choice language=node
```

```text
Recipe team 1.0.0
  chain: team 1.0.0 (team.recipe.toml)

Choices
  language     node           (answered; options python, node) Language?

Plugins
  language     node                     team.recipe.toml:14 (recipe team)
  lint         eslint                   team.recipe.toml:15 (recipe team)
  test         vitest                   team.recipe.toml:16 (recipe team)

The recipe is valid.
```

## Ship

### smith ci

The generated pipelines call these phases, so CI runs exactly what you can run locally. See [CI pipelines](../reference/ci.md).

<!-- smith:begin ci -->
Run a CI phase, or regenerate the pipelines

Usage: `smith ci`

No flags.

<!-- smith:end ci -->

| Subcommand | Does |
|---|---|
| `smith ci setup` | install every toolchain dependency the workspace's plugins need |
| `smith ci audit` | check dependencies for known vulnerabilities |
| `smith ci verify` | lint, typecheck, test and build; only affected projects on a pull request |
| `smith ci pack` | run every project's `package` task into `dist/drop`, kept as the pipeline's "drop" artifact on pushes to the release branch |
| `smith ci native` | lint, test, build and package the projects tagged native (cgo) on this OS; the pipeline runs it on each OS and, on a release, attaches the build |
| `smith ci release` | version, publish and tag every package with releasable commits |
| `smith ci generate [--check]` | rewrite the pipelines from the workspace's projects; `--check` only reports drift |
| `smith ci check` | list what to set up by hand: secrets, and each private registry's permissions |

```sh
$ smith ci generate --check
$ smith ci check
```

```text
pipelines are up to date
To set up by hand before the pipelines run:
  - GitHub: add the repository secrets PYPI_TOKEN (Settings → Secrets and variables → Actions)
```

### smith pipeline

Add a ready-made pipeline file beside the generated CI. It starts the way every generated job does (toolchains, smith, `smith ci setup`), then does one thing. It is written once and yours from then on: `smith upgrade` never rewrites it, and an existing file is kept unless `--force`. With no template it lists them.

<!-- smith:begin pipeline -->
Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function

Usage: `smith pipeline [template] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--ci <ci>` | github, azure or both (default: the workspace's providers) | the workspace's providers |
| `--force` | replace a pipeline file that exists | off |
| `--project <project>` | the app a deploy template builds |  |

<!-- smith:end pipeline -->

```sh
$ smith pipeline
```

```text
e2e                                      github, azure  run every project's e2e task, one at a time, on a push to the release branch, on pull requests (GitHub) and on demand
package-zip                              github, azure  pack the projects (smith ci pack) and keep dist/drop as a build artifact, the input of a classic release pipeline
deploy-pages (--project <app>)           github         build an app and deploy its build output to GitHub Pages
deploy-azure-function (--project <app>)  github, azure  build a function app and deploy it to an Azure Function App
```

### smith docker

Docker image commands.

<!-- smith:begin docker -->
Docker image commands

Usage: `smith docker`

No flags.

<!-- smith:end docker -->

| Subcommand | Does |
|---|---|
| `smith docker description [--dry-run] [--strict] [-p <project>]...` | set the Docker Hub repository descriptions of each app whose registries include a `docker.io` entry: the full one from its `description_from` file, the short one from `short_description` or the project's description. Needs `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` (or the variables the registry entry names). A failure is a warning unless `--strict` |

Building and pushing images is part of `smith release` (see [Releasing](releasing.md#container-images)), not a command of its own.

```sh
$ smith docker description --dry-run
```

```text
no app has a description_from file or a short description for a docker.io registry
```

### smith release

Version, publish and tag every package with releasable commits since its last tag: `feat` is a minor bump, `fix` and `perf` a patch, a breaking change a major (a minor before 1.0). Packages release in dependency order; each is stamped, built, published and tagged, and manifests are restored afterwards (the tag is the record). A re-run after a failure resumes where it stopped.

<!-- smith:begin release -->
Version, publish and tag every package with releasable commits

Usage: `smith release [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--comment` | with --dry-run in a pull request build, post the preview on the pull request (GitHub, Azure Repos), updating smith's earlier comment | off |
| `--dry-run` | show what would be released; change nothing | off |
| `--format <format>` | how the result is printed: text, or markdown (what the CI job summary and --comment use) |  |
| `--json` | print the release report as JSON | off |
| `--no-push` | create tags locally only | off |
| `--notes` | also print each release's notes (what its GitHub Release says) | off |
| `--status` | show the release journal: what was published where (dist/release/<project>/journal.json) | off |
| `--verbose` | name every unchanged package | off |
| `-y, --yes` | do not ask for confirmation | off |

<!-- smith:end release -->

Without `--dry-run` it asks for confirmation (`--yes` skips the question). In CI, the pipeline runs `smith ci release`. `--comment` with `--dry-run` posts the preview on the pull request, updating smith's earlier comment. `--json` prints the report (below, cut). See [Releasing](releasing.md).

```sh
$ smith release --dry-run
```

```text
  sdk                  unreleased → 0.1.0  (2 releasing commit(s))
sdk: would release 0.1.0 (2 releasing commit(s))

### Release preview

| Package | Version | Why | Status |
|---|---|---|---|
| sdk | new → **0.1.0** | 2 releasing commit(s) | planned |
```

The line at the top is the preview for the terminal; the table is the markdown that `--comment` posts.

```sh
$ smith release --dry-run --json | head -c 700
```

```text
sdk: would release 0.1.0 (2 releasing commit(s))
{
  "dryRun": true,
  "started": "2026-10-19T07:06:13.8132743Z",
  "finished": "0001-01-01T00:00:00Z",
  "results": [
    {
      "project": "sdk",
      "plugin": "python",
      "next": "0.1.0",
      "level": "minor",
      "tag": "sdk@0.1.0",
```

## Shell

### smith completion

Generate the autocompletion script for `bash`, `zsh`, `fish` or `powershell`.

```sh
$ smith completion bash | head -n 3
```

```text
# bash completion V2 for smith                                -*- shell-script -*-

__smith_debug()
```

Load it the way your shell loads completions, for example `source <(smith completion bash)` in `~/.bashrc`, or `smith completion powershell | Out-String | Invoke-Expression` in a PowerShell profile. `smith completion <shell> --help` has the details.

### smith help

`smith help <command>` is `smith <command> --help`.

### smith (no arguments)

In a terminal, `smith` on its own opens the interactive shell: a menu, every command, the doctor, the release preview and the settings, as tabs. Each action ends as a smith command line, printed before it runs. When stdin or stdout is not a terminal, or `CI` is set, it prints its help instead. See [Getting started](../guides/getting-started.md#7-the-interactive-shell).
