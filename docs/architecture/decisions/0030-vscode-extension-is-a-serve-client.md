# ADR 0030: The VS Code extension is a client of `smith serve`

- Status: accepted
- Date: 2026-10-23

## Context

The old mnci had a VS Code extension built on the `--json` contracts. MonoSmith needed
first-party IDE integration
([#303](https://github.com/monosmith-app/monosmith/issues/303)).

## Options considered

The issue proposed a language server (`smith lsp`) so every editor gets the same features,
plus the extension, and a JetBrains plugin later. The first extension shipped without the
language server.

## Decision

The extension, created with `smith add package vscode --stack vscode-extension`, runs the
installed smith ([ADR 0025](0025-ide-plugins-run-the-installed-smith.md)) over `smith serve`
and carries no engine. It uses only documented contracts, with no parsing of human output.
Its first version has a Projects view, task runs with live output and cancel, and a status
bar. Later versions added health, dependencies and release views, Problems panel entries,
a task provider, a tree filter and an Explain page. It is released as an app by the vscode
plugin, one `.vsix` per platform when it is native.

## Consequences

The language server and the JetBrains plugin are not built. In a workspace that is not
trusted, the extension does not run the workspace's own smith.

## Sources

- Pull request [#426](https://github.com/monosmith-app/monosmith/pull/426); pull requests
  [#430](https://github.com/monosmith-app/monosmith/pull/430),
  [#431](https://github.com/monosmith-app/monosmith/pull/431),
  [#540](https://github.com/monosmith-app/monosmith/pull/540); epic
  [#523](https://github.com/monosmith-app/monosmith/issues/523)
- [VS Code extension architecture](../vscode-architecture.md),
  [The MonoSmith VS Code extension](../../manual/vscode.md)
