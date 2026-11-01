# Reference

Facts to look up: the configuration file, the concepts, the conventions. Commands and
their flags are in the [manual](../manual/README.md); the pages here say what the
settings and concepts mean.

| Page | Holds |
|---|---|
| [Configuration: `smith.toml`](configuration.md) | every setting of the workspace file, its default, and the environment variables smith reads |
| [Glossary](glossary.md) | the words smith uses: workspace, project, kind, task, stack, recipe, slot, and the rest |
| [Conventions](conventions.md) | the vertical-slice layout smith generates and checks, and how to turn the check on or off |
| [Feature parity](feature-parity.md) | which feature lives in the CLI, the TUI, VS Code and JetBrains |
| [Derived files](derived-files.md) | the ownership modes, `smith.lock`, the states of a derived file and the renderers |
| [Stacks](stacks.md) | what a project can be built with (React, Vue, Jest, ...) and how a stack is chosen |
| [Recipes](recipes.md) | a team's choice of plugins and settings, the recipe file format and the answers file |
| [Tasks, cache and affected](tasks-and-cache.md) | how tasks are defined and overridden, the cache and the remote cache, `smith dev` and serving an app for a task |
| [CI pipelines](ci.md) | the generated pipelines, their phases, slots, providers and caches |
| [Releasing](releasing.md) | versions, registries, tags, binaries, container images and the Marketplace |
| [Plugins](plugins.md) | the built-in plugins, roles, the protocol and the search path |
| [Serve and the event stream](serve.md) | `smith serve`'s JSON-RPC methods and the run events |
| [CLI reference](cli.md) | every command with its flags in one generated table (the [manual](../manual/README.md) has a page per command) |

How these pieces work inside is in the [architecture manual](../architecture/README.md).
