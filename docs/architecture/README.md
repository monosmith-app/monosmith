# Architecture manual

How MonoSmith is built, how it identifies and organises what it works on, and how each
major decision was taken. These pages are for people who change smith, write plugins or
integrations, or need to know why smith behaves as it does. Users who only want to run
commands need the [manual](../manual/README.md) and the [reference](../reference/README.md).

Every page states what it describes in the code (package paths and symbols), so a claim
can be checked. Where the project history does not record why something was decided, the
page says "rationale not recorded" and links the commit or pull request.

## Start here

| Page | Answers |
|---|---|
| [Architecture](architecture.md) | the repository, the core slices, the plugin boundary and the data flow of a run, a release and a CI phase, on one page |
| [MonoSmith system context and containers](system-context.md) | who and what smith talks to (C4 levels 1 and 2): editors, CI providers, registries, plugin processes, the caches |
| [The repository's own slice architecture](slice-architecture.md) | how `apps/smith`, `libs/*` and `packages/*` are organised, what enforces it and where it departs from its own rules |

## How smith identifies and organises a workspace

```mermaid
flowchart LR
  A[How a project is discovered and classified] --> B[The project graph and how edges are found]
  B --> C[The task graph and how tasks are ordered]
  C --> D[The cache key, hits and misses]
  B --> E[How affected projects are computed]
  E --> C
```

| Page | Answers |
|---|---|
| [How a project is discovered and classified](project-discovery.md) | which files make a project, how its language, kind and stack are decided, what `smith.toml` and per-project markers override |
| [The project graph and how edges are found](project-graph.md) | how imports, exports and shared schemas become edges, cycles, selection, blast radius |
| [The task graph and how tasks are ordered](task-graph.md) | where tasks come from, how `depends_on` and `^task` expand, scheduling, skipping and cancelling |
| [The cache key, hits and misses](cache.md) | what the key is made of, the store layout, the remote tier, why a task missed |
| [How affected projects are computed](affected.md) | the base, the changed files, owners and dependents, and the fallback to everything |
| [How tasks run, and how the cache decides](execution-and-cache.md) | the longer treatment of the three pages above, with the scheduler and process handling |

## Plugins, recipes, CI and release

| Page | Answers |
|---|---|
| [The plugin protocol: roles, handshake, capabilities and contributions](plugin-protocol.md) | the wire protocol at a glance and the order of calls when a project is assembled |
| [The plugin system](plugin-system.md) | why plugins are processes, how they are loaded, how they interact and fail |
| [Design: recipes, and CI as part of them](recipes-and-ci.md) | the design of recipes with levels and slots, with what is built and what is not |
| [How CI pipelines are generated](ci-generation.md) | the provider-neutral model, the phases, slots, the tracked-changes guard |
| [How a release is decided and published](release-design.md) | conventional commits to versions, publishing, tags, SBOM and provenance |
| [Releases and CI: how they are built](release-and-ci-internals.md) | the longer treatment of both |
| [How adopt and convert work](adopt-and-convert.md) | inventory, plan, scratch copy, depth levels |

## Front ends

| Page | Answers |
|---|---|
| [The architecture of the terminal interface](tui-architecture.md) | the shell, the keymap, pages, wizards and the runner |
| [The architecture of the VS Code extension](vscode-architecture.md) | how the extension finds smith, talks to `smith serve` and renders views |
| [Design: the TUI and the IDE plugins](tui-and-ide.md) | the original design, with an "as built" status box at the top |

## Decisions and background

| Page | Answers |
|---|---|
| [Decision records](decisions/README.md) | one page per decision: context, options, decision, consequences, status, date |
| [Technology: what smith is made of, and why](technology.md) | Go, the libraries, process boundaries, the security model, the test layers |
| [Research](research/01-current-mnci-inventory.md) | the three research notes smith was planned from (dated snapshots of mnci, Nx and prior art) |

The research notes are [the inventory of mnci](research/01-current-mnci-inventory.md),
[the Nx subsystems](research/02-nx-subsystems.md) and
[prior art and the stack](research/03-prior-art-and-stack.md).
