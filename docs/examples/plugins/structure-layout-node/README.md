# structure-layout-node

The structure plugin of [`structure-layout-go`](../structure-layout-go), written in **Node.js with no
dependencies and no SDK**: it frames the JSON-RPC messages itself. It is both the plugin (when smith
starts it, `SMITH_PLUGIN=1`) and the checker the `layout-check` task runs
(`node plugin.mjs layout-check`).

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md); how the framing
works: [Writing a plugin in Go, Node.js or Python](../../../plugins/creating/plugin-sdks.md).

## Files

| File | What it is |
|---|---|
| `plugin.mjs` | the whole plugin: framing, dispatch, `initialize`, `contribute`, `doctor`, the checker |
| `plugin.test.mjs` | tests that speak the protocol to the plugin as a child process (`node --test`) |
| `plugin.cmd` | a Windows launcher (`@echo off`, then `node "%~dp0plugin.mjs" %*`) |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.mjs`; use `./plugin.cmd` on Windows) |
| `package.json` | `"type": "module"` and the test script |

Run with Node.js 24; it uses ES modules and `node --test`.

## Test, check

```sh
node --test
smith plugin check ./plugin.mjs        # Linux, macOS: chmod +x plugin.mjs first (not run there)
```

```text
PS> smith plugin check .\plugin.cmd
```

## Use

```toml
[plugins.external.widget-layout]
command = ["node", "/path/to/plugin.mjs"]

[conventions.widget-layout]
```
