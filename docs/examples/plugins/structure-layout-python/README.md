# structure-layout-python

The structure plugin of [`structure-layout-go`](../structure-layout-go), written in **Python with the
standard library only** and no SDK: it frames the JSON-RPC messages itself. It is both the plugin (when
smith starts it, `SMITH_PLUGIN=1`) and the checker the `layout-check` task runs
(`python plugin.py layout-check`).

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md); how the framing
works: [Writing a plugin in Go, Node.js or Python](../../../plugins/creating/plugin-sdks.md).

## Files

| File | What it is |
|---|---|
| `plugin.py` | the whole plugin: framing, dispatch, `initialize`, `contribute`, `doctor`, the checker |
| `test_plugin.py` | tests that speak the protocol to the plugin as a child process (`python -m unittest`) |
| `plugin.cmd` | a Windows launcher (`@echo off`, then `python "%~dp0plugin.py" %*`) |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.py`; use `./plugin.cmd` on Windows) |

Needs Python 3 (it was run with 3.13).

## Test, check

```sh
python -m unittest
smith plugin check ./plugin.py         # Linux, macOS: chmod +x plugin.py first (not run there)
```

```text
PS> smith plugin check .\plugin.cmd
```

## Use

```toml
[plugins.external.widget-layout]
command = ["python", "/path/to/plugin.py"]

[conventions.widget-layout]
```
