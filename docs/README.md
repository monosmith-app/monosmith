# MonoSmith documentation

## Guides

Start here to install smith and use it.

| | |
|---|---|
| [Installing](guides/installing.md) | every install channel, upgrading, checking a download |
| [Getting started](guides/getting-started.md) | a new workspace, adding projects, running tasks |
| [Converting from Nx](guides/converting.md) | `smith convert` for Nx and Turborepo workspaces and the old mnci |
| [Editors](guides/editors.md) | the `--json` contracts editors and tools use |

## Reference

What each part does and how to configure it.

| | |
|---|---|
| [Configuration](reference/configuration.md) | every `smith.toml` setting and environment variable |
| [Tasks, cache and affected](reference/tasks-and-cache.md) | how tasks run, the local and remote cache, `--affected` |
| [CI pipelines](reference/ci.md) | the generated GitHub, Azure and GitLab pipelines, their phases and secrets |
| [Releasing](reference/releasing.md) | versions from commits, tags, publishing to every registry |
| [Stacks](reference/stacks.md) | the plugins each project is built with |
| [Recipes](reference/recipes.md) | a team's choice of plugins and settings: the file, resolving, `smith recipe check`, the `smith recipe new` wizard, releasing recipes |
| [Conventions](reference/conventions.md) | workspace-wide conventions such as vertical slices |
| [Plugins](reference/plugins.md) | the built-in plugins, writing one, the protocol |

## Design

Why smith is built the way it is.

| | |
|---|---|
| [Architecture](design/architecture.md) | the core's slices, plugins as processes, how a run and a release flow |
| [Decisions](design/decisions.md) | what was decided, when and why, and what is still open |
| [Recipes and CI](design/recipes-and-ci.md) | proposal: team recipes of plugins, with choices, and CI as part of them |
| [The TUI and the IDE plugins](design/tui-and-ide.md) | proposal: a rich, mouse-driven interface with live results, and IDE plugins with the engine embedded |
| [Research](design/research/) | the research smith was planned from (historical) |

## Contributing

| | |
|---|---|
| [Working on MonoSmith](contributing/development.md) | the repository layout, checks, and how smith releases itself |
