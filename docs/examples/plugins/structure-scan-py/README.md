# structure-scan-py

The structure plugin of [`structure-scan-go`](../structure-scan-go), written in **Python on
[`monosmith-plugin-sdk`](../../../../packages/plugin-sdk-py)**. Its `doctor` walks every file of a widget project
and warns about any over a mebibyte, and it shows the two things a plugin does when a call can run for long:

- **It stops when smith cancels the request.** The plugin says `cancellable = True`, so the SDK passes every method
  that takes one a `CancelToken` as its second argument, sets it when smith sends the `cancel` notification for the
  request (Ctrl+C, a timeout), and advertises the `cancel` capability. The walk waits between files with
  `cancel.wait(delay)`, a sleep that ends the moment the token is cancelled, so it ends within one file's time.
- **It takes four requests at a time.** `"concurrency": 4` in what `describe` returns says smith may have four
  requests in flight, and the SDK runs them together on four worker threads. A plugin that says nothing is called one
  request at a time.

For the same requests, a cancel included, this plugin, the Go one and the TypeScript one answer with the same bytes,
which `packages/plugin-sdk-py/tests/integration` checks. The rule is in
[the plugin protocol](../../../architecture/plugin-protocol.md#cancel-and-concurrency); the SDK in
[The Python SDK](../../../plugins/creating/python-sdk.md#stop-work-when-the-host-cancels-it).

## Files

| File | What it is |
|---|---|
| `scan.py` | the plugin: `ScanPlugin(delay)` (`describe` with the concurrency, `doctor`) and the walk that watches the token |
| `plugin.py` | the entry point that serves it; `WIDGET_SCAN_DELAY_MS` adds a pause before each file, so a test can keep a scan running until it cancels it |
| `test_scan.py` | tests that serve the plugin in the test process with `connect_in_process`: conformance, a cancelled scan, four scans at once; they wait on events, never on a pause (`python -m unittest`) |
| `plugin.cmd` | a Windows launcher |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.py`; use `./plugin.cmd` on Windows) |

Needs Python 3.10 or later with the SDK importable: `pip install monosmith-plugin-sdk`. Inside this
repository the workspace's `.venv` has the neighbouring package installed in editable mode.

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
[plugins.external.widget-scan]
command = ["/path/to/venv/bin/python", "/path/to/plugin.py"]

[conventions.widget-scan]
```
