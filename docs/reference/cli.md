# CLI reference

Every `smith` command with its usage line, aliases and a table of its flags, grouped as
`smith --help` groups them. For what a command is for, when to use it, its exit codes, files,
environment variables, worked examples and errors, follow the link to its page in
[the manual](../manual/README.md).

> **Keeping this page current.** The usage line and flag table of each command sit between
> `<!-- smith:begin <command> -->` and `<!-- smith:end <command> -->` markers and are
> generated from `smith commands --json`, as are the fuller tables in the manual. Everything
> else is written by hand. After changing a command, build smith and run, from the repository
> root:
>
> ```sh
> node docs/reference/cli-generate.mjs           # rewrite the blocks
> node docs/reference/cli-generate.mjs --check   # fail if a page is out of date
> ```
>
> The script names the binary with `SMITH` (default: `smith` on the PATH) and fails when a
> command has no block, here or in the manual, so a new command cannot be forgotten.
> `smith <command> --help` always has the last word.

## Everywhere

```sh
smith --version
```

```text
smith version 1.62.0
```

| Flag | Meaning |
|---|---|
| `-h`, `--help` | help for `smith` or for any command: `smith <command> --help` |
| `-v`, `--version` | print the version |

- **Workspace.** Almost every command works on the workspace it finds by looking for
  `smith.toml` in the current folder and then in each parent. `SMITH_WORKSPACE_ROOT`
  overrides that. Outside one, `smith` says so and suggests `smith new`.
- **`--json`.** Commands that report something print a stable JSON document with `--json`, for
  editors and scripts. See [Editors](../guides/editors.md).
- **Environment variables** are listed in [configuration](configuration.md#environment-variables).
- **Exit codes.** 0 on success; 1 on any failure (a failed check or task, a refusal, an invalid
  flag).
- **Project selection** (`-p`): a name, a glob over names or directories (`apps/*`),
  `kind:app`, `lang:go`, `tag:x`, and a leading `!` to exclude. Comma-separated or repeated.
- **Bare `smith`** in a terminal opens [the terminal interface](../manual/tui.md); `smith
  completion <shell>` prints a completion script ([manual](../manual/completion.md)).

## Create

### smith new

[Manual page](../manual/new.md)

<!-- smith:begin new -->
Create a monorepo: configuration, git, CI pipelines and a first commit

Usage: `smith new <name> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--architecture <architecture>` | the code layout smith checks: vertical-slices, hexagonal, clean, ddd, modular-monolith, mvc, package-by-layer, none (default vertical-slices) | `vertical-slices` |
| `--cache-dir <cache-dir>` | the task cache folder inside the workspace (default .smith/cache) | `.smith/cache` |
| `--cache-max-mb <cache-max-mb>` | the largest the task cache may grow, in MB; 0 for no limit (default 2048) | `2048` |
| `--ci <ci>` | pipelines to generate: github, azure, gitlab, circleci (default github) (repeatable) | `github` |
| `--default-stack <default-stack>` | the stack Node projects get, as node=<stack>: node-jest (default), node-vitest or node-ts (repeatable) | `node=node-jest` |
| `--dir <dir>` | directory to create the workspace in (default ./<name>) | `./<name>` |
| `--github-releases` | create a GitHub Release for every release tag | off |
| `--language <language>` | the main language: node, python, go, jvm, dotnet, rust, or later (default later) | `later` |
| `--module-prefix <module-prefix>` | Go module prefix (default from the origin remote, else example.com/<name>) | from the `origin` remote, else `example.com/<name>` |
| `--recipe <recipe>` | the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md) |  |
| `--release-branch <release-branch>` | the branch releases run from (default main) | `main` |
| `--release-remote <release-remote>` | the git remote that receives release tags (default origin) | `origin` |
| `--scope <scope>` | npm scope for packages (default @<name>) | `@<name>` |
| `-y, --yes` | take the default for every question that has no flag | off |

<!-- smith:end new -->

### smith adopt

[Manual page](../manual/adopt.md)

<!-- smith:begin adopt -->
Bring an existing repository under smith: report what it involves; --init takes the steps (start here)

Usage: `smith adopt [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--dry-run` | only report (this is the default; the flag makes it explicit) | off |
| `--eslint` | with --init, write the two-line root eslint.config.mjs and add @monosmith/eslint-config to the root devDependencies | off |
| `--init` | write smith.toml and the pipelines, and commit | off |
| `--install` | with --init, also run npm install for the Node projects | off |
| `--json` | print the report as JSON | off |
| `--level <level>` | how far to go: conservative (keep the build and lint setup) or full (smith convert's, for Nx and Turborepo workspaces) | `conservative` |
| `--tags` | create the baseline release tags the report asks for, locally | off |

<!-- smith:end adopt -->

### smith convert

[Manual page](../manual/convert.md)

<!-- smith:begin convert -->
Convert an Nx, Turborepo or mnci workspace to smith (adopt sends you here when it recognises one)

Usage: `smith convert [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--allow-dirty` | convert even with uncommitted changes (not with --level full) | off |
| `--build <build>` | full level: build with smith's stack build (smith) or keep the current build setup (keep) | `smith` |
| `--ci <ci>` | full level: the pipeline smith generates (smith) or keep the current one (keep) | `smith` |
| `--dry-run` | report the plan without changing anything | off |
| `--json` | print the plan as JSON | off |
| `--level <level>` | how far to go: conservative (keep what works, list the rest) or full (make smith the only tool, deleting the old tooling after asking) | `conservative` |
| `--lint <lint>` | full level: lint with smith's ESLint config (smith) or keep the current setup (keep) | `smith` when the ESLint config imports a shared config smith replaces, else `keep` |
| `--skip-install` | do not run npm install afterwards | off |
| `--skip-verify` | full level: install, but do not build, test, lint and run doctor afterwards | off |
| `--strict` | refuse to convert while anything needs manual work or is unsupported | off |
| `--test <test>` | full level: the test runner to keep, jest or vitest (the other's configs and packages go), or keep | the runner the workspace uses, else `keep` |
| `--verbose` | name every project a finding applies to, not just a few | off |
| `--yes` | apply without asking, and install an Nx workspace's packages first (a terminal asks before changing anything) | off |

<!-- smith:end convert -->

### smith add

[Manual page](../manual/add.md)

<!-- smith:begin add -->
Add a project: an app, an internal lib or a publishable package (to add a dependency, use smith install)

Usage: `smith add <kind> <name> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--choice <choice>` | answer a choice of the workspace's recipe, name=answer (repeatable); the rest are asked in a terminal (repeatable) |  |
| `-l, --lang <lang>` | language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, recipe (default: the stack's, or the workspace's only language) | the stack's, or the workspace's only language |
| `-s, --stack <stack>` | stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone) | the recipe, else `[default_stacks]` in `smith.toml` |
| `--verbose` | show the package manager's output as it runs (it is shown only if it fails otherwise) | off |

<!-- smith:end add -->

### smith remove

[Manual page](../manual/remove.md)

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

### smith kind

[Manual page](../manual/kind.md)

<!-- smith:begin kind -->
Make a project an app, an internal lib or a published package

Usage: `smith kind <project> <kind> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--dry-run` | list every file edit and the move; change nothing | off |
| `--force` | go ahead although other projects import it or it was already released | off |
| `--json` | print the plan and what was done as JSON | off |

<!-- smith:end kind -->

### smith config

[Manual page](../manual/config.md)

<!-- smith:begin config -->
Read and change settings of smith.toml

Usage: `smith config <command>`

| Subcommand | Meaning |
|---|---|
| `smith config get` | Print one setting's value as written in smith.toml |
| `smith config list` | List the settings, their values and what they do |
| `smith config set` | Change one setting (a list takes several values, or one comma-separated) |
| `smith config unset` | Remove one setting, so it takes its default |

No flags.

<!-- smith:end config -->

### smith architecture

[Manual page](../manual/architecture.md)

<!-- smith:begin architecture -->
Show the architecture of the workspace, or choose another

Usage: `smith architecture [name]`

No flags.

<!-- smith:end architecture -->

### smith upgrade

[Manual page](../manual/upgrade.md)

<!-- smith:begin upgrade -->
Bring the workspace up to this smith and its recipes: migrate smith.toml, regenerate the files smith owns

Usage: `smith upgrade [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--check` | change nothing; fail when something would change | off |
| `--dry-run` | show the plan, with the change to smith.toml, and change nothing | off |
| `--json` | print the result as JSON | off |

<!-- smith:end upgrade -->

## Work

### smith run

[Manual page](../manual/run.md)

<!-- smith:begin run -->
Run tasks on projects, in dependency order, from cache when nothing changed

Usage: `smith run <task>... [flags]`

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

### smith build

[Manual page](../manual/build.md)

<!-- smith:begin build -->
Run the build task (same as `smith run build`)

Usage: `smith build [project]... [flags]`

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

### smith test

[Manual page](../manual/test.md)

<!-- smith:begin test -->
Run the test task (same as `smith run test`)

Usage: `smith test [project]... [flags]`

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

### smith lint

[Manual page](../manual/lint.md)

<!-- smith:begin lint -->
Run the lint task (same as `smith run lint`)

Usage: `smith lint [project]... [flags]`

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

### smith typecheck

[Manual page](../manual/typecheck.md)

<!-- smith:begin typecheck -->
Run the typecheck task (same as `smith run typecheck`)

Usage: `smith typecheck [project]... [flags]`

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

### smith eject

[Manual page](../manual/eject.md)

<!-- smith:begin eject -->
Copy smith's default configuration of a tool into the repository, to own and edit it

Usage: `smith eject [tool...] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--all` | eject every default that applies | off |
| `--force` | replace a file that exists already | off |

<!-- smith:end eject -->

### smith dev

[Manual page](../manual/dev.md)

<!-- smith:begin dev -->
Start several projects together, such as a frontend and the API it calls

Usage: `smith dev [projects...] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--all` | start every project that has a start task | off |
| `--dry-run` | print the equivalent smith run command and start nothing | off |

<!-- smith:end dev -->

### smith install

[Manual page](../manual/install.md)

<!-- smith:begin install -->
Add a package to a project (-w), or with none install everything the workspace declares (up updates, sync aligns versions)

Usage: `smith install [packages...] [flags]`

Alias: `smith i`

| Flag | Meaning | Default |
|---|---|---|
| `-D, --save-dev` | add as a development dependency (npm devDependencies, the Python dev extra, a Go tool) | off |
| `--verbose` | show the package manager's own output and warnings (otherwise warnings fold into one line) | off |
| `-w, --workspace <workspace>` | target project, by name or directory; repeat it for several (repeatable) |  |

<!-- smith:end install -->

### smith uninstall

[Manual page](../manual/uninstall.md)

<!-- smith:begin uninstall -->
Remove packages from a project (the opposite of install)

Usage: `smith uninstall <packages...> [flags]`

Alias: `smith un`

| Flag | Meaning | Default |
|---|---|---|
| `-w, --workspace <workspace>` | target project, by name or directory; repeat it for several (repeatable) |  |

<!-- smith:end uninstall -->

### smith sync

[Manual page](../manual/sync.md)

<!-- smith:begin sync -->
Make every project declare a shared dependency at the same version (same as up --align)

Usage: `smith sync [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--check` | report only; fail when declarations disagree | off |
| `--json` | print the report as JSON | off |
| `--no-install` | rewrite the manifests without installing | off |

<!-- smith:end sync -->

### smith up

[Manual page](../manual/up.md)

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

### smith deps

[Manual page](../manual/deps.md)

<!-- smith:begin deps -->
Group for dependency work across projects; deps up is the same as up

Usage: `smith deps <command>`

| Subcommand | Meaning |
|---|---|
| `smith deps up` | Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync) |

No flags.

<!-- smith:end deps -->

## Ship

### smith ci

[Manual page](../manual/ci.md)

<!-- smith:begin ci -->
Run a CI phase, or regenerate the pipelines

Usage: `smith ci <command>`

| Subcommand | Meaning |
|---|---|
| `smith ci audit` | Run the audit phase |
| `smith ci check` | List what to set up by hand for the pipelines: secrets, and each private registry's permissions |
| `smith ci generate` | Rewrite the CI pipelines from the workspace's projects (--check only reports drift) |
| `smith ci native` | Run the native phase |
| `smith ci pack` | Run the pack phase |
| `smith ci release` | Run the release phase |
| `smith ci setup` | Run the setup phase |
| `smith ci verify` | Run the verify phase |

No flags.

<!-- smith:end ci -->

### smith pipeline

[Manual page](../manual/pipeline.md)

<!-- smith:begin pipeline -->
Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function

Usage: `smith pipeline [template] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--ci <ci>` | github, azure or both (default: the workspace's providers) | the workspace's providers |
| `--force` | replace a pipeline file that exists | off |
| `--project <project>` | the app a deploy template builds |  |

<!-- smith:end pipeline -->

### smith docker

[Manual page](../manual/docker.md)

<!-- smith:begin docker -->
Docker image commands

Usage: `smith docker <command>`

| Subcommand | Meaning |
|---|---|
| `smith docker description` | Set the Docker Hub descriptions of the apps' images |

No flags.

<!-- smith:end docker -->

### smith release

[Manual page](../manual/release.md)

<!-- smith:begin release -->
Version, publish and tag every package with releasable commits

Usage: `smith release <command> [flags]`

| Subcommand | Meaning |
|---|---|
| `smith release pub` | Publish a released app to pub.dev from its GitHub Release (the generated pub workflow runs it) |

| Flag | Meaning | Default |
|---|---|---|
| `--comment` | with --dry-run in a pull request build, post the preview on the pull request (GitHub, Azure Repos), updating smith's earlier comment | off |
| `--dry-run` | show what would be released; change nothing | off |
| `--format <format>` | how the result is printed: text, or markdown (what the CI job summary and --comment use) | `text` |
| `--json` | print the release report as JSON | off |
| `--no-push` | create tags locally only | off |
| `--notes` | also print each release's notes (what its GitHub Release says) | off |
| `--status` | show the release journal: what was published where (dist/release/<project>/journal.json) | off |
| `--verbose` | name every unchanged package | off |
| `-y, --yes` | do not ask for confirmation | off |

<!-- smith:end release -->

## Inspect

### smith projects

[Manual page](../manual/projects.md)

<!-- smith:begin projects -->
List the workspace's projects

Usage: `smith projects [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end projects -->

### smith graph

[Manual page](../manual/graph.md)

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

### smith affected

[Manual page](../manual/affected.md)

<!-- smith:begin affected -->
List the projects affected by changes since a base branch

Usage: `smith affected [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--base <base>` | base branch (default: the PR target, else the release branch) | the pull request target, else the release branch |
| `--json` | print JSON, with the directories to watch for each app | off |
| `--paths` | print each app and the directories a deployment platform should watch | off |

<!-- smith:end affected -->

### smith explain

[Manual page](../manual/explain.md)

<!-- smith:begin explain -->
Show how a project's tasks resolved, and where each setting came from (why-miss says why one reran)

Usage: `smith explain <project>[:<task>] [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |
| `--recipe` | print only what the project's recipe decided, and where | off |
| `--task <task>` | explain only this task (same as <project>:<task>) |  |

<!-- smith:end explain -->

### smith why-miss

[Manual page](../manual/why-miss.md)

<!-- smith:begin why-miss -->
Say why a task would run again instead of coming from the cache (a cache miss; explain shows its settings)

Usage: `smith why-miss <project>:<task> [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end why-miss -->

### smith cache

[Manual page](../manual/cache.md)

<!-- smith:begin cache -->
Look after the local task cache: status, prune, clear, verify

Usage: `smith cache <command>`

| Subcommand | Meaning |
|---|---|
| `smith cache clear` | Delete the local cache (the remote cache is not touched) |
| `smith cache prune` | Trim the local cache to its limit now, removing the least recently used entries |
| `smith cache status` | Show the local cache: where it is, how big, how full, and the largest entries |
| `smith cache verify` | Check the local cache for damage; --fix removes what is broken |

No flags.

<!-- smith:end cache -->

### smith doctor

[Manual page](../manual/doctor.md)

<!-- smith:begin doctor -->
Check the workspace's health; every problem names its remedy

Usage: `smith doctor [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--choose <choose>` | with --fix and --only, answer the fix's questions by option number, in order (1,2) |  |
| `--fix` | show the fixes doctor can make as diffs, and apply the safe ones | off |
| `--json` | print the report as JSON | off |
| `--only <only>` | with --fix, fix only the named check, as doctor prints it (repeatable) (repeatable) |  |
| `--verbose` | list every passed check, not one line for each check that repeats per project | off |
| `-y, --yes` | with --fix, apply without asking | off |

<!-- smith:end doctor -->

### smith info

[Manual page](../manual/info.md)

<!-- smith:begin info -->
Show the installed smith, the newest released one, and this workspace's settings

Usage: `smith info [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the report as one JSON document | off |

<!-- smith:end info -->

### smith kinds

[Manual page](../manual/kinds.md)

<!-- smith:begin kinds -->
List the project kinds smith add accepts (--json for editors)

Usage: `smith kinds [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the kinds as one JSON document | off |

<!-- smith:end kinds -->

### smith stacks

[Manual page](../manual/stacks.md)

<!-- smith:begin stacks -->
List the stacks projects can be built with, and who uses each

Usage: `smith stacks [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print JSON | off |

<!-- smith:end stacks -->

### smith plugin

[Manual page](../manual/plugin.md)

<!-- smith:begin plugin -->
Inspect plugins, or serve a built-in plugin

Usage: `smith plugin <command>`

| Subcommand | Meaning |
|---|---|
| `smith plugin check` | Check that a plugin keeps the protocol's promises |
| `smith plugin disable` | Disable built-in plugins in smith.toml |
| `smith plugin enable` | Enable built-in plugins in smith.toml |
| `smith plugin install` | Install a packaged plugin, checked against its sha256 |
| `smith plugin list` | List the built-in plugins, the external ones the workspace uses (trusted or not) and the ones installed on the plugin search path |
| `smith plugin pack` | Pack a plugin directory into an archive `smith plugin install` accepts |
| `smith plugin set` | Enable exactly these built-in plugins in smith.toml |
| `smith plugin trust` | Review an external plugin and trust it, so smith may start it |
| `smith plugin untrust` | Forget that you trust an external plugin |

No flags.

<!-- smith:end plugin -->

### smith recipe

[Manual page](../manual/recipe.md)

<!-- smith:begin recipe -->
Check recipes: a team's choice of plugins and settings

Usage: `smith recipe <command>`

| Subcommand | Meaning |
|---|---|
| `smith recipe check` | Resolve and validate a recipe, showing what it would use and where each value came from |
| `smith recipe new` | Write a recipe by answering questions |

No flags.

<!-- smith:end recipe -->

### smith boundaries

[Manual page](../manual/boundaries.md)

<!-- smith:begin boundaries -->
Check which projects depend on which against your rules

Usage: `smith boundaries [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the violations as JSON | off |

<!-- smith:end boundaries -->

### smith slices

[Manual page](../manual/slices.md)

<!-- smith:begin slices -->
Check the vertical-slice layout of projects

Usage: `smith slices <command>`

| Subcommand | Meaning |
|---|---|
| `smith slices check` | Check that files are role-named and no folder is a junk drawer |

No flags.

<!-- smith:end slices -->

### smith schema

[Manual page](../manual/schema.md)

<!-- smith:begin schema -->
Print the JSON Schema of smith.toml, for editors

Usage: `smith schema`

No flags.

<!-- smith:end schema -->

### smith commands

[Manual page](../manual/commands.md)

<!-- smith:begin commands -->
List smith's commands with their arguments and options (--json for editors)

Usage: `smith commands [flags]`

| Flag | Meaning | Default |
|---|---|---|
| `--json` | print the commands as one JSON document | off |

<!-- smith:end commands -->

### smith serve

[Manual page](../manual/serve.md)

<!-- smith:begin serve -->
Serve the engine to an editor or a terminal interface over JSON-RPC (stdio)

Usage: `smith serve`

No flags.

<!-- smith:end serve -->

### smith mcp

[Manual page](../manual/mcp.md)

<!-- smith:begin mcp -->
Serve the workspace to an AI assistant over the Model Context Protocol (stdio)

Usage: `smith mcp`

No flags.

<!-- smith:end mcp -->
