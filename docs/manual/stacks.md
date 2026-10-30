# smith stacks

`smith stacks` lists the stacks projects can be built with, and which projects use each. A
stack is a named set of plugins: a language plugin, and the toolchain, structure, scaffold
and deploy plugins on top. Plugins publish stacks; `smith.toml` declares more under
`[stacks.<name>]`. A stack is chosen with `smith add --stack`:

```sh
smith add app web --stack node-vitest
```

## When to use it

- To pick a `--stack` for `smith add`.
- To see which stack a workspace's default is (`default for node`) and which projects use
  which.

Unlike [`smith kinds`](kinds.md), it needs a workspace.

## Synopsis

```sh
smith stacks [--json]
```

## Arguments and flags

<!-- smith:begin stacks -->
List the stacks projects can be built with, and who uses each

Usage: `smith stacks [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end stacks -->

## Exit codes

0; 1 outside a workspace.

## Files and environment

Reads `smith.toml` (`[stacks]`, `[default_stacks]`) and the plugins.

## Examples

```sh
smith stacks
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
...
node-jest  (plugin jest)  default for node
  TypeScript on Node, linted by @monosmith/eslint-config, tested with Jest
  language   node
  toolchain  typescript, eslint, jest
  scaffold   typescript
  projects   ui, web
...
```

Eighteen stacks in this version: android, angular, docs-site, flutter, go-smith-plugin, html,
kotlin, node-jest, node-ts, node-vitest, node-vitest-playwright, node-vitest-tauri, react, react-p5,
react-tauri, svelte, vscode-extension, vue.

## Common errors

`there is no smith workspace here`: run it inside a workspace.

## Related

[`smith add`](add.md), [`smith kinds`](kinds.md), [Stacks](../reference/stacks.md).
