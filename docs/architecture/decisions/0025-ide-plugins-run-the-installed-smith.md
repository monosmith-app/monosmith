# ADR 0025: IDE plugins run the installed smith

- Status: accepted (replaced an earlier plan to embed the engine)
- Date: 2026-10-18

## Context

The product owner asked for IDE plugins that offer the same functions as the terminal
interface and use the smith the team has installed, the way the ESLint and Jest extensions
use the workspace's own tools, so the team controls the version.

## Options considered

- Embed the binary in each extension package (platform packages for VS Code, bundled
  binaries for JetBrains): the earlier plan. Not the default, because it would let the IDE
  run a different version from the rest of the team. It remains possible later as a fallback
  for a machine with no smith.
- A WebAssembly build of the engine: not the plan, because the engine runs processes and
  reads the file system, which a browser sandbox does not allow.
- Run the installed smith: chosen.

## Decision

The plugins do not carry the engine. They find smith in this order: a setting
(`monosmith.path` in the VS Code extension), the workspace's own install
(`node_modules/.bin/smith` or `.smith/bin/smith`), then `smith` on the `PATH`. If none is
found, the plugin shows what is missing and the install commands. `smith serve` reports its
version and protocol version so the plugin can tell what the smith it found can do.
Updating smith is the team's decision, made in the repository.

## Consequences

The IDE, the terminal and CI run the same smith, so a task gives the same result, cache key
and release preview everywhere. See [ADR 0030](0030-vscode-extension-is-a-serve-client.md)
for the extension and [ADR 0026](0026-event-stream-and-smith-serve.md) for the protocol.

## Sources

- Pull request [#359](https://github.com/monosmith-app/monosmith/pull/359) (the design
  proposal), commit `dc84b52`
- [The TUI and the IDE plugins](../tui-and-ide.md#the-plugins-use-the-installed-smith)
