# The smith manual

One page for every `smith` command and subcommand, for the terminal interface and for the
VS Code extension. Each command page says what the command is for, when to use it and when
not to, its synopsis, every argument and flag (type, default, effect, how flags interact),
exit codes, the files it reads and writes, the environment variables it uses, worked
examples whose output was captured from a real run of smith 1.62.0, the common errors with
their remedies, and related commands.

The flag tables are generated from `smith commands --json` by
[`docs/reference/cli-generate.mjs`](../reference/cli-generate.mjs), so they cannot drift from the
binary; everything else is written by hand. For a one-page list of every flag, see the
[CLI reference](../reference/cli.md). `smith <command> --help` always has the last word.

Examples use a throwaway workspace named `acme` (a Go app and lib, two Node projects and a
Python package). Paths in output are shortened.

## Conventions every command shares

- **Workspace.** Almost every command works on the workspace it finds by looking for
  `smith.toml` in the current folder and then in each parent folder. `SMITH_WORKSPACE_ROOT`
  names another folder. Outside a workspace, a command that needs one fails with:

  ```text
  ✖ there is no smith workspace here (no smith.toml in this folder or above it)
    → smith new <name> creates one
    → in an existing repository, smith adopt --dry-run brings it in
  ```

  Commands that do not need one: `new`, `adopt`, `convert`, `kinds`, `schema`, `commands`,
  `info`, `mcp`, `plugin list`, `plugin check`,
  `completion`, `help`.
- **Exit codes.** Every command exits 0 on success and 1 on any failure: a failed check or
  task, a refusal, an invalid flag or a blocker. There are no other codes. Pages list what
  makes a command exit 1.
- **Errors.** Two shapes: `smith: <message>` for a plain error, and a block that starts with
  `✖` followed by one or more `→` lines naming the remedy.
- **`--json`.** Commands that report something print a stable JSON document with `--json`, for
  editors and scripts.
- **Missing arguments.** `smith new` and `smith add` ask for a missing argument in a terminal
  (and when `CI` is unset); everywhere else a missing argument is an error that gives an
  example.
- **Colour and update checks.** `NO_COLOR` turns colour off. `SMITH_NO_UPDATE_CHECK` stops
  `smith info` from asking the registry for the newest release.
- **Help.** `smith --help` lists the commands in four groups; `smith <command> --help` is
  the command's own help. `smith help <command>` is the same.
- **Bare `smith`.** In a terminal, with `CI` unset, `smith` with no arguments opens
  [the terminal interface](tui.md).

## Create

| Command | Does |
|---|---|
| [`smith new`](new.md) | Create a monorepo: configuration, git, CI pipelines and a first commit. |
| [`smith adopt`](adopt.md) | Bring an existing repository under smith: report what it involves; `--init` takes the steps. |
| [`smith convert`](convert.md) | Convert an Nx, Turborepo or mnci workspace to smith. |
| [`smith add`](add.md) | Add an app, an internal lib or a publishable package. |
| [`smith remove`](remove.md) | Remove a project: its directory, its traces in the workspace, its CI. |
| [`smith config`](config.md) | Read and change settings of `smith.toml`: [`list`](config-list.md), [`get`](config-get.md), [`set`](config-set.md), [`unset`](config-unset.md). |
| [`smith upgrade`](upgrade.md) | Bring the workspace up to this smith and its recipes. |

## Work

| Command | Does |
|---|---|
| [`smith run`](run.md) | Run tasks on projects, in dependency order, from cache when nothing changed. |
| [`smith build`](build.md), [`smith test`](test.md), [`smith lint`](lint.md), [`smith typecheck`](typecheck.md) | `smith run` with the task fixed. |
| [`smith dev`](dev.md) | Start several projects together. |
| [`smith install`](install.md), [`smith uninstall`](uninstall.md) | Add or remove packages in a project, or install the whole workspace. |
| [`smith up`](up.md) | Show dependencies with newer releases and update the ones you pick. |
| [`smith sync`](sync.md) | Make every project declare a shared dependency at the same version. |
| [`smith deps`](deps.md) | Group for dependency work; [`smith deps up`](deps-up.md) is `smith up`. |

## Ship

| Command | Does |
|---|---|
| [`smith ci`](ci.md) | Run a CI phase, or regenerate the pipelines: [`ci generate`](ci-generate.md), [`ci check`](ci-check.md). |
| [`smith pipeline`](pipeline.md) | Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function. |
| [`smith docker`](docker.md) | Docker image commands: [`docker description`](docker-description.md). |
| [`smith release`](release.md) | Version, publish and tag every package with releasable commits: [`release pub`](release-pub.md). |

## Inspect

| Command | Does |
|---|---|
| [`smith projects`](projects.md) | List the workspace's projects. |
| [`smith graph`](graph.md) | Show how projects depend on each other. |
| [`smith affected`](affected.md) | List the projects affected by changes since a base branch. |
| [`smith explain`](explain.md) | Show how a project's tasks resolved, and where each setting came from. |
| [`smith why-miss`](why-miss.md) | Say why a task would run again instead of coming from the cache. |
| [`smith doctor`](doctor.md) | Check the workspace's health; every problem names its remedy. |
| [`smith info`](info.md) | The installed smith, the newest released one, and this workspace's settings. |
| [`smith kinds`](kinds.md), [`smith stacks`](stacks.md) | What `smith add` accepts. |
| [`smith plugin`](plugin.md) | Inspect plugins: [`list`](plugin-list.md), [`check`](plugin-check.md), [`enable`](plugin-enable.md), [`disable`](plugin-disable.md), [`set`](plugin-set.md). |
| [`smith recipe`](recipe.md) | Check recipes: [`check`](recipe-check.md), [`new`](recipe-new.md). |
| [`smith slices`](slices.md) | Check the vertical-slice layout: [`slices check`](slices-check.md). |
| [`smith schema`](schema.md) | Print the JSON Schema of `smith.toml`. |
| [`smith commands`](commands.md) | List smith's commands with their arguments and options. |
| [`smith serve`](serve.md) | Serve the engine to an editor over JSON-RPC (stdio). |
| [`smith mcp`](mcp.md) | Serve the workspace to an AI assistant over the Model Context Protocol (stdio). |

## Shell, interface and editor

| Page | About |
|---|---|
| [`smith completion`](completion.md) | The autocompletion script for bash, fish, PowerShell and zsh. |
| [The smith terminal interface](tui.md) | The full-screen shell a bare `smith` opens: views, every key, the wizards. |
| [The MonoSmith VS Code extension](vscode.md) | Every command, view, setting, task field and key. |

## Not documented here

`smith` has a few internal commands that its own tasks call (for example `npm-pack`,
`npm-prune` and `npm-deps-check`, which `smith explain` shows as task commands); they are hidden from
`smith --help` and are not part of the interface.
