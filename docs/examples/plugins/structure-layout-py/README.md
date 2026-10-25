# structure-layout-py

The structure plugin of [`structure-layout-go`](../structure-layout-go), written in **Python on
[`monosmith-plugin-sdk`](../../../../packages/plugin-sdk-py)**. The SDK frames the messages, answers
`initialize`, derives the capabilities from the two methods the plugin has (`contribute` and `doctor`)
and keeps standard output for the protocol. It is both the plugin (when smith starts it,
`SMITH_PLUGIN=1`) and the checker the `layout-check` task runs (`python plugin.py layout-check`).

The same plugin with the framing and dispatch written by hand, no SDK, is
[`structure-layout-python`](../structure-layout-python) (and [`node`](../node) in Node.js). For the same
requests this plugin, the Go one and the TypeScript one answer with the same bytes, which
`packages/plugin-sdk-py/tests/integration` checks.

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md); the SDK:
[The Python SDK](../../../plugins/creating/python-sdk.md).

## Files

| File | What it is |
|---|---|
| `plugin.py` | the entry point: serves the plugin when smith started the program (`SMITH_PLUGIN=1`), runs the `layout-check` checker otherwise |
| `layout.py` | the plugin: `LayoutPlugin` (`describe`, `contribute`, `doctor`) and `find_junk_drawers` |
| `test_layout.py` | tests that serve the plugin in the test process with the SDK's `connect_in_process` and run its conformance checks (`python -m unittest`) |
| `plugin.cmd` | a Windows launcher (`@echo off`, then `python "%~dp0plugin.py" %*`) |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.py`, which starts with `#!/usr/bin/env python3`; use `./plugin.cmd` on Windows) |

Needs Python 3.10 or later with the SDK importable: `pip install monosmith-plugin-sdk`. Inside this
repository the workspace's `.venv` has the neighbouring package installed in editable mode (`smith ci
setup`).

## Test, check

```sh
python -m unittest
smith plugin check ./plugin.py        # Linux, macOS: chmod +x plugin.py first (not run there)
```

```text
PS> smith plugin check .\plugin.cmd
```

## Use

```toml
[plugins.external.widget-layout]
command = ["/path/to/venv/bin/python", "/path/to/plugin.py"]

[conventions.widget-layout]
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget-layout` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../../plugins/using/plugin-security.md).
