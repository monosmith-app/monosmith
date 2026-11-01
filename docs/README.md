# MonoSmith documentation

MonoSmith (`smith`) runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP,
C++, Dart, the JVM and Bicep from one command: it scaffolds projects, runs their
tasks in dependency order from a cache, generates the CI pipeline, and
releases from conventional commits.

Evaluating smith for your team? Start with [Why MonoSmith](product/README.md).

New here? Read [Getting started](getting-started/README.md). Stuck on a word? See the
[glossary](reference/glossary.md).

## Who reads what

| You are | Start with |
|---|---|
| a new user | [Getting started](getting-started/README.md): [install](getting-started/installing.md), then [your first workspace](getting-started/first-workspace.md), [adopt a repository](getting-started/adopting.md) or [convert from Nx or Turborepo](getting-started/converting.md) |
| an operator or a daily user | the [manual](manual/README.md), one page per command, then the [reference](reference/README.md) and the [guides](guides/README.md) |
| using or writing a plugin, or building an integration | [Plugins: using and creating them](plugins/README.md), [Plugins (reference)](reference/plugins.md), [Editors](guides/editors.md) |
| a contributor | the [architecture manual](architecture/README.md), the [decision records](architecture/decisions/README.md) and [Contributing](contributing/README.md) |

## The sections

| Section | Holds |
|---|---|
| [Getting started](getting-started/README.md) | installing; a first workspace; adopting an existing repository; converting from Nx, Turborepo or mnci |
| [Manual](manual/README.md) | every command and subcommand with its flags, exit codes, files and examples, grouped as Create, Work, Ship and Inspect; plus [the terminal interface](manual/tui.md) and [the VS Code extension](manual/vscode.md) |
| [Plugins](plugins/README.md) | what plugins are and how they combine; installing, enabling, configuring and troubleshooting them; [a page for each of the 37 built-in plugins](plugins/using/catalogue/README.md); a complete guide for each type of plugin you can write, with [working examples](examples/plugins/README.md) |
| [Guides](guides/README.md) | [editors and assistants](guides/editors.md), [the MonoSmith agent](guides/agent.md), [derived files](guides/derived-files.md), [know your files and brief your agents](guides/know-your-files-and-brief-your-agents.md), [troubleshooting](guides/troubleshooting.md), the [FAQ](guides/faq.md), [writing a plugin](guides/writing-a-plugin.md) (a pointer to the plugin section) |
| [Reference](reference/README.md) | [configuration](reference/configuration.md), [glossary](reference/glossary.md), [conventions](reference/conventions.md), [derived files](reference/derived-files.md), [stacks](reference/stacks.md), [recipes](reference/recipes.md), [tasks, cache and affected](reference/tasks-and-cache.md), [CI](reference/ci.md), [releasing](reference/releasing.md), [plugins](reference/plugins.md), [serve](reference/serve.md), [CLI](reference/cli.md) |
| [Architecture](architecture/README.md) | how smith is built, identifies and organises a workspace, and why: overview, context, discovery, graphs, cache, CI, release, plugins, front ends, [decision records](architecture/decisions/README.md) |
| [Contributing](contributing/README.md) | [working on smith](contributing/development.md) and the [documentation audit](contributing/documentation-audit.md) |

## I want to...

| I want to | Read |
|---|---|
| install smith | [Installing](getting-started/installing.md) |
| try it end to end | [Your first workspace](getting-started/first-workspace.md) |
| bring an existing repository under smith | [Adopting an existing repository](getting-started/adopting.md) |
| move from Nx, Turborepo or the old mnci | [Converting](getting-started/converting.md) |
| add, remove and connect projects | [`smith add`](manual/add.md), [`smith remove`](manual/remove.md), [your first workspace](getting-started/first-workspace.md#3-add-projects) |
| run tasks, understand the cache, run only what changed | [Tasks, cache and affected](reference/tasks-and-cache.md), [`smith run`](manual/run.md) |
| share the cache with my team | [The remote cache](reference/tasks-and-cache.md#remote-cache) |
| run an end-to-end suite against a running app | [Serving an app for a task](reference/tasks-and-cache.md#serving-an-app-for-a-task) |
| change a setting in `smith.toml` | [Configuration](reference/configuration.md), [`smith config`](manual/config.md) |
| choose what a project is built with | [Stacks](reference/stacks.md) |
| give every new project the same shape | [Recipes](reference/recipes.md) |
| see what a built-in plugin does for my projects | [The built-in plugin catalogue](plugins/using/catalogue/README.md) |
| install, switch off or configure a plugin | [Installing, enabling and configuring plugins](plugins/using/managing-plugins.md) |
| write a plugin | [Creating plugins](plugins/creating/README.md), [quickstart](plugins/creating/quickstart.md) |
| keep code laid out by what it does | [Conventions](reference/conventions.md) |
| build a project from a framework, a test runner and a style | [`smith add --with`](manual/add.md#building-from-choices), [`smith choices`](manual/choices.md) |
| write the config that repeats in every project from one place | [Derived files](guides/derived-files.md), [`smith derive`](manual/derive.md) |
| see who owns a file, or attach a report of the workspace to a pull request | [`smith why`](manual/why.md), [`smith report`](manual/report.md) |
| brief a coding agent on the rules of the workspace | [`smith agent file`](manual/agent-file.md), [The MonoSmith agent](guides/agent.md) |
| revert the last `smith add`, `remove`, `upgrade`, `derive` or `sync` | [`smith undo`](manual/undo.md) |
| release prerelease versions from a branch such as `next` | [Release channels](reference/releasing.md#release-channels) |
| generate and customise CI | [CI pipelines](reference/ci.md) |
| release packages, binaries and container images | [Releasing](reference/releasing.md) |
| use smith from the keyboard | [The terminal interface](manual/tui.md) |
| use smith in VS Code or from an AI assistant | [The VS Code extension](manual/vscode.md), [Editors](guides/editors.md) |
| fix something that does not work | [Troubleshooting](guides/troubleshooting.md), [`smith doctor`](manual/doctor.md) |
| look up a command or a flag | [The manual](manual/README.md), [CLI reference](reference/cli.md) |
| see why smith is built this way | [Architecture](architecture/README.md), [Decision records](architecture/decisions/README.md) |

The screenshots in [`assets/`](assets) were captured from real runs; the
[`assets/render-terminal.mjs`](assets/render-terminal.mjs) script turns captured
terminal output into the images.
