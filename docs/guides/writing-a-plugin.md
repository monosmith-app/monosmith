# Writing a plugin

The guide to writing a plugin has moved and grown into a section of its own: one complete guide for
each kind of plugin, a quickstart, a protocol cheat sheet, and worked examples that are compiled and
run. Start at [Creating plugins](../plugins/creating/README.md), or go straight to the
[quickstart](../plugins/creating/quickstart.md).

A plugin is any program that talks to smith over its standard input and output; it can be written in
any language. smith's core knows no language, so everything that depends on a technology (finding
projects, their tasks, scaffolding, versions, publishing, CI needs, health checks) lives in a plugin.

## Where each part of the old guide went

| It used to be | It is now |
|---|---|
| What a plugin is for, roles and kinds | [What a plugin is and how plugins combine](../plugins/using/what-is-a-plugin.md) |
| The protocol in five minutes: framing, rules for stdout, conversation, error codes, timeouts | [The plugin protocol cheat sheet](../plugins/creating/protocol-cheat-sheet.md) |
| The methods: `initialize`, `discover`, `generate`, `version/write`, `publish`, `ci`, `doctor`, `remove`, dependencies | the cheat sheet for the shapes, and [Creating a language plugin](../plugins/creating/language-plugin.md) for how to implement them |
| Doctor fixes: files, steps and choices | [Creating a language plugin](../plugins/creating/language-plugin.md#doctor-fixes-files-steps-and-choices) |
| `contribute`, a plugin that adds tasks | [Creating a toolchain plugin](../plugins/creating/toolchain-plugin.md) and [Creating a structure plugin](../plugins/creating/structure-plugin.md) |
| Installing a plugin: `plugin.json` and the search path | [Installing, enabling and configuring plugins](../plugins/using/managing-plugins.md) and [Packaging and distributing a plugin](../plugins/creating/packaging-a-plugin.md) |
| Testing: `smith plugin check` | [Testing a plugin](../plugins/creating/testing-a-plugin.md) |
| The widget example in Go, Python and Node.js | [Writing a plugin in Go, Node.js or Python](../plugins/creating/plugin-sdks.md), and the [`go`](../examples/plugins/go), [`node`](../examples/plugins/node) and [`python`](../examples/plugins/python) examples, which are unchanged |
| Other languages | [Writing a plugin in Go, Node.js or Python](../plugins/creating/plugin-sdks.md#what-exists): Go and TypeScript have an SDK ([The TypeScript SDK](../plugins/creating/typescript-sdk.md)), Python and C# are next |
| Troubleshooting a plugin | [Troubleshooting plugins](../plugins/using/troubleshooting-plugins.md) |

## Pick a guide

| You want smith to | Read |
|---|---|
| understand a new project type | [Creating a language plugin](../plugins/creating/language-plugin.md) |
| run a tool in a slot such as `lint` or `test` | [Creating a toolchain plugin](../plugins/creating/toolchain-plugin.md) |
| enforce a team rule | [Creating a structure plugin](../plugins/creating/structure-plugin.md) |
| write a new project's starter files | [Creating a scaffold plugin](../plugins/creating/scaffold-plugin.md) |
| ship an app somewhere | [Creating a deploy plugin](../plugins/creating/deploy-plugin.md) |
| choose and configure plugins as data | [Creating a recipe and a stack](../plugins/creating/recipe-and-stack.md) |

Also see [Plugins (reference)](../reference/plugins.md), [the plugin system](../architecture/plugin-system.md)
and the [glossary](../reference/glossary.md).
