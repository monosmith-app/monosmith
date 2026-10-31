# Creating plugins

These pages are for the person who writes a plugin: what each kind of plugin must implement, a
complete worked example of each, and how to test, package and distribute the result. Every example
is a real program in [`docs/examples/plugins`](../../examples/plugins/README.md): it compiles (or
runs), its tests pass, `smith plugin check` accepts it, and the output pasted in the guides was
produced by running it.

Start with the [quickstart](quickstart.md), then read the guide for the kind of plugin you are
writing.

## Pick the guide

| You want smith to | Write a | Guide | Example |
|---|---|---|---|
| understand a new project type: find it, build its graph, version and publish it | language plugin | [Creating a language plugin](language-plugin.md) | [`language-widget`](../../examples/plugins/language-widget) |
| run a tool for projects of a language smith already knows (a linter, a test runner) | toolchain plugin | [Creating a toolchain plugin](toolchain-plugin.md) | [`toolchain-widget-lint`](../../examples/plugins/toolchain-widget-lint) |
| check or enforce your team's rules on layout and code | structure plugin | [Creating a structure plugin](structure-plugin.md) | [`structure-layout-go`](../../examples/plugins/structure-layout-go) |
| write the starter files of a new project your way | scaffold plugin | [Creating a scaffold plugin](scaffold-plugin.md) | [`scaffold-widget-service`](../../examples/plugins/scaffold-widget-service) |
| ship an app somewhere: pipeline steps, credentials, a release target | deploy plugin | [Creating a deploy plugin](deploy-plugin.md) | [`deploy-widget-bundle`](../../examples/plugins/deploy-widget-bundle) |
| let a team choose and configure plugins per project, with no code | recipe and stack | [Creating a recipe and a stack](recipe-and-stack.md) | [`recipe-widget-team`](../../examples/plugins/recipe-widget-team) |

A **recipe** is not a plugin: it is data that chooses and configures plugins. If a recipe can say
what you want, you do not need to write code.

One rule applies to every guide below: when you declare your plugin in a workspace, smith starts it
only after you [trust it](../using/managing-plugins.md#trusting-an-external-plugin) (the first command
that needs it asks in a terminal, or run `smith plugin trust <name>`), and it starts with a [minimal
environment](../using/plugin-security.md#the-environment-an-external-plugin-gets): a variable your
plugin reads must be named in `env` of its `plugin.json`
([how](packaging-a-plugin.md#declaring-the-environment-a-plugin-needs)). `smith plugin check` needs
no trust, so the commands in the guides that check a plugin work as written.

## The pages

| Page | Read it when |
|---|---|
| [Writing a plugin: quickstart](quickstart.md) | you want a working plugin in ten minutes |
| [The plugin protocol cheat sheet](protocol-cheat-sheet.md) | you need a method, a message or a field, in a table |
| [Creating a language plugin](language-plugin.md) | |
| [Creating a toolchain plugin](toolchain-plugin.md) | |
| [Creating a structure plugin](structure-plugin.md) | |
| [Creating a scaffold plugin](scaffold-plugin.md) | |
| [Creating a deploy plugin](deploy-plugin.md) | |
| [Creating a recipe and a stack](recipe-and-stack.md) | you want your team's choice of plugins as a file |
| [Writing a plugin in Go, Node.js or Python](plugin-sdks.md) | you choose a language: which have an SDK, and what you write yourself in Node.js and in Python without one |
| [The TypeScript SDK](typescript-sdk.md) | you write a plugin in TypeScript: a walk-through, then every function, type and helper of `@monosmith/plugin-sdk` |
| [The Python SDK](python-sdk.md) | you write a plugin in Python: a walk-through, then every function, type and helper of `monosmith-plugin-sdk` |
| [The C# SDK](csharp-sdk.md) | you write a plugin in C#: a walk-through, then every interface, type and helper of `MonoSmith.Plugin.Sdk` |
| [Sharing facts between plugins](shared-facts.md) | several tools must agree on a value (an alias, a source folder) and you want to publish or read it |
| [Testing a plugin](testing-a-plugin.md) | `smith plugin check`, unit tests, and a workspace test |
| [Packaging and distributing a plugin](packaging-a-plugin.md) | naming, `plugin.json`, binaries per platform, versions, `smith plugin pack` and how users install the result |

## The example family

The guides share one made-up project type so you can see the five kinds of plugin work together.
A **widget** is a directory with a `widget.json`; the plugins below teach smith about it, one level
each:

```mermaid
flowchart LR
  recipe["recipe-widget-team<br/>(choices)"] --> stack["stack widget-team-service"]
  stack --> L["language-widget<br/>discover, generate,<br/>version, doctor"]
  stack --> S["scaffold-widget-service<br/>generate"]
  stack --> T["toolchain-widget-lint<br/>provides lint"]
  stack --> R["structure-layout-go<br/>layout-check, doctor"]
  stack --> D["deploy-widget-bundle<br/>bundle task, publish, ci"]
```

Each plugin is a Go program on the SDK. The structure plugin is also written in
[TypeScript](../../examples/plugins/structure-layout-node), on the TypeScript SDK, and in
[Python](../../examples/plugins/structure-layout-py), on the Python SDK (and once more in
[Python without an SDK](../../examples/plugins/structure-layout-python)) and in
[C#](../../examples/plugins/structure-layout-cs), on the C# SDK; the language plugin
too, in [TypeScript](../../examples/plugins/language-widget-ts),
[Python](../../examples/plugins/language-widget-py) and [C#](../../examples/plugins/language-widget-cs). So the same
behaviour can be compared across languages: the Go, the TypeScript, the Python and the C# plugin answer the same
requests with the same bytes. The SDKs are Go, TypeScript, Python and C#
([#564](https://github.com/monosmith-app/monosmith/issues/564)).

## See also

- [Using plugins](../using/README.md)
- [The plugin system](../../architecture/plugin-system.md) and
  [the plugin protocol](../../architecture/plugin-protocol.md): the design
- [Plugins (reference)](../../reference/plugins.md)
