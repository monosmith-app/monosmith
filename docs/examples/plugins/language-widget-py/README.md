# language-widget-py

The language plugin of [`language-widget`](../language-widget), in **Python on
[`monosmith-plugin-sdk`](../../../../packages/plugin-sdk-py)**. It teaches smith the widget project
type: a directory with a `widget.json` is a project, `uses` in it become graph edges, and each project
gets a cached `build` and `test` task. It also scaffolds widgets (`smith add lib core --lang widget`),
writes the release version into `widget.json` keeping its key order, and reports a doctor finding with
a fix.

For the same requests it gives the same bytes as the Go plugin and the [TypeScript
one](../language-widget-ts), which `packages/plugin-sdk-py/tests/integration` checks.

Guide: [Creating a language plugin](../../../plugins/creating/language-plugin.md); the SDK:
[The Python SDK](../../../plugins/creating/python-sdk.md).

## Files

| File | What it is |
|---|---|
| `plugin.py` | the entry point: serves the protocol when smith started the program (`SMITH_PLUGIN=1`), runs the command line tool otherwise |
| `widget.py` | the plugin: `describe`, `discover`, `generate`, `write_version`, `doctor` |
| `manifest.py` | reading and rewriting `widget.json` |
| `tool.py` | the command line the tasks call: `check`, `build <dir>` |
| `test_widget.py` | the conformance suite and behaviour tests, over an in-process connection (`python -m unittest`) |
| `workspace-test.sh` | an end-to-end test in a throwaway workspace |
| `plugin.cmd`, `plugin.json` | the Windows launcher, and how to install it on the search path |

Needs Python 3.10 or later with the SDK importable: `pip install monosmith-plugin-sdk`. Inside this
repository the workspace's `.venv` has the neighbouring package installed in editable mode.

## Test, check

```sh
python -m unittest
smith plugin check ./plugin.py --generate     # Linux, macOS: chmod +x plugin.py first (not run there)
PYTHON=/path/to/venv/bin/python sh workspace-test.sh
```

```text
PS> smith plugin check .\plugin.cmd --generate
```

## Use

```toml
[plugins.external.widget]
command = ["/path/to/venv/bin/python", "/path/to/plugin.py"]
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../../plugins/using/plugin-security.md).

```sh
smith add lib core --lang widget
smith add app shop --lang widget      # then add "uses": ["core"] to apps/shop/widget.json
smith graph
smith run build test
```
