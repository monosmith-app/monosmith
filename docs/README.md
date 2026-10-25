# MonoSmith documentation

MonoSmith (`smith`) runs a monorepo in Node, Go, Python, Rust, .NET, Ruby, PHP,
C++, Dart, the JVM and Bicep from one command: it scaffolds projects, runs their
tasks in dependency order from a cache, generates the CI pipeline, and
releases from conventional commits.

New here? Read [Getting started](getting-started/first-workspace.md). It is a twenty
minute tour with real commands and real output. Stuck on a word? See the
[glossary](reference/glossary.md).

## I want to...

### Get going

| I want to | Read |
|---|---|
| install smith | [Installing](getting-started/installing.md) |
| try it end to end | [Getting started](getting-started/first-workspace.md) |
| bring an existing repository under smith | [Getting started, adopt](getting-started/first-workspace.md#or-adopt-an-existing-repository) |
| move from Nx, Turborepo or the old mnci | [Converting](getting-started/converting.md) |
| understand the words | [Glossary](reference/glossary.md) |
| get a quick answer | [FAQ](guides/faq.md) |

### Work every day

| I want to | Read |
|---|---|
| add, remove and connect projects | [Getting started, add projects](getting-started/first-workspace.md#3-add-projects) |
| run tasks, understand the cache, run only what changed | [Tasks, cache and affected](reference/tasks-and-cache.md) |
| share the cache with my team | [The remote cache](reference/tasks-and-cache.md#remote-cache) |
| run an end-to-end suite against a running app | [Serving an app for a task](reference/tasks-and-cache.md#serving-an-app-for-a-task) |
| change a setting in `smith.toml` | [Configuration](reference/configuration.md) |
| choose what a project is built with (React, Vue, Jest, ...) | [Stacks](reference/stacks.md) |
| look up a command or a flag | [CLI reference](reference/cli.md) |
| fix something that does not work | [Troubleshooting](guides/troubleshooting.md) |

### Ship

| I want to | Read |
|---|---|
| generate and customise CI | [CI pipelines](reference/ci.md) |
| release packages, binaries and container images | [Releasing](reference/releasing.md) |
| give every new project the same shape | [Recipes](reference/recipes.md) |
| keep code laid out by what it does | [Conventions](reference/conventions.md) |

### Extend and integrate

| I want to | Read |
|---|---|
| support a language smith lacks, or add a rule | [Writing a plugin](guides/writing-a-plugin.md) (working examples in Go, Python and Node.js in [`examples/plugins`](examples/plugins)) |
| look up a plugin role, the protocol or the search path | [Plugins](reference/plugins.md) |
| use smith from an editor or an AI assistant | [Editors](guides/editors.md) |

### Understand how smith works

| I want to | Read |
|---|---|
| see the big picture | [Architecture](architecture/architecture.md) |
| see why plugins are processes, and how they are loaded | [The plugin system](architecture/plugin-system.md) |
| see how a run and the cache work | [Execution and cache](architecture/execution-and-cache.md) |
| see how releases and CI are built | [Release and CI internals](architecture/release-and-ci-internals.md) |
| see why Go, and what else is in the stack | [Technology](architecture/technology.md) |
| read what was decided and why | [Decisions](architecture/decisions.md) |
| read the proposals | [Recipes and CI](architecture/recipes-and-ci.md), [The TUI and the IDE plugins](architecture/tui-and-ide.md) |
| read the research smith was planned from | [Research](architecture/research/) |
| work on smith itself | [Working on MonoSmith](contributing/development.md) |

## The pages

**Guides** (do something): [Installing](getting-started/installing.md),
[Getting started](getting-started/first-workspace.md),
[Converting from Nx](getting-started/converting.md), [Editors](guides/editors.md),
[Writing a plugin](guides/writing-a-plugin.md),
[Troubleshooting](guides/troubleshooting.md), [FAQ](guides/faq.md).

**Reference** (look something up): [CLI](reference/cli.md),
[Configuration](reference/configuration.md),
[Tasks, cache and affected](reference/tasks-and-cache.md),
[CI pipelines](reference/ci.md), [Releasing](reference/releasing.md),
[Stacks](reference/stacks.md), [Recipes](reference/recipes.md),
[Conventions](reference/conventions.md), [Plugins](reference/plugins.md),
[Serve and the event stream](reference/serve.md),
[Glossary](reference/glossary.md).

**Design** (why it is built this way): see above.

The screenshots in [`assets/`](assets) were captured from real runs; the
[`assets/render-terminal.mjs`](assets/render-terminal.mjs) script turns captured
terminal output into the images.
