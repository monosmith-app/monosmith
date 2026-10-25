# structure-layout-node

The structure plugin of [`structure-layout-go`](../structure-layout-go), written in **TypeScript on
[`@monosmith/plugin-sdk`](../../../../packages/plugin-sdk-ts)**. The SDK frames the messages, answers
`initialize`, derives the capabilities from the two methods the plugin has (`contribute` and `doctor`)
and keeps standard output for the protocol. It is both the plugin (when smith starts it,
`SMITH_PLUGIN=1`) and the checker the `layout-check` task runs (`node plugin.ts layout-check`).

The same plugin with the framing and dispatch written by hand, in about 40 lines of Node.js and no
SDK, is [`node`](../node) (and [`python`](../python) in Python). For the same requests this plugin and
the Go one answer with the same bytes, which `packages/plugin-sdk-ts/integration` checks.

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md); the SDK:
[The TypeScript SDK](../../../plugins/creating/typescript-sdk.md).

## Files

| File | What it is |
|---|---|
| `plugin.ts` | the whole plugin: `createLayoutPlugin()` (`describe`, `contribute`, `doctor`), the checker, and the entry point that serves it |
| `plugin.test.ts` | tests that serve the plugin in the test process with the SDK's `connectInProcess` and run its conformance checks (`node --test`) |
| `plugin.cmd` | a Windows launcher (`@echo off`, then `node "%~dp0plugin.ts" %*`) |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.ts`, which starts with `#!/usr/bin/env node`; use `./plugin.cmd` on Windows) |
| `package.json` | `"type": "module"`, the SDK as the one dependency, the test script |

Run with Node.js 24, which runs TypeScript files as they are (it only removes the types). Install the
SDK first outside this repository: `npm install`. Inside the repository it resolves to the neighbouring
package, once the workspace is installed.

## Test, check

```sh
node --test
smith plugin check ./plugin.ts        # Linux, macOS: chmod +x plugin.ts first (not run there)
```

```text
PS> smith plugin check .\plugin.cmd
```

## Use

```toml
[plugins.external.widget-layout]
command = ["node", "/path/to/plugin.ts"]

[conventions.widget-layout]
```
