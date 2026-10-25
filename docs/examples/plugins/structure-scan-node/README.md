# structure-scan-node

The structure plugin of [`structure-scan-go`](../structure-scan-go), written in **TypeScript on
[`@monosmith/plugin-sdk`](../../../../packages/plugin-sdk-ts)**. Its `doctor` walks every file of a widget project
and warns about any over a mebibyte, and it shows the two things a plugin does when a call can run for long:

- **It stops when smith cancels the request.** The plugin says `cancellable: true`, so the SDK passes every method an
  `AbortSignal` as its second argument, aborts it when smith sends the `cancel` notification for the request
  (Ctrl+C, a timeout), and advertises the `cancel` capability. The walk waits between files with
  `timers/promises` and the signal, so it ends within one file's time.
- **It takes four requests at a time.** `concurrency: 4` in what `describe` returns says smith may have four requests
  in flight, and the SDK runs them together. A plugin that says nothing is called one request at a time.

For the same requests, a cancel included, this plugin and the Go one answer with the same bytes, which
`packages/plugin-sdk-ts/integration` checks; the Python twin is [`structure-scan-py`](../structure-scan-py), compared with
both by `packages/plugin-sdk-py/tests/integration`. The rule is in
[the plugin protocol](../../../architecture/plugin-protocol.md#cancel-and-concurrency); the SDK in
[The TypeScript SDK](../../../plugins/creating/typescript-sdk.md#stop-work-when-the-host-cancels-it).

## Files

| File | What it is |
|---|---|
| `scan.ts` | the plugin: `createScanPlugin(delayMs)` (`describe` with the concurrency, `doctor`) and the walk that watches the signal |
| `plugin.ts` | the entry point that serves it; `WIDGET_SCAN_DELAY_MS` adds a pause before each file, so a test can cancel a scan in the middle |
| `scan.test.ts` | tests that serve the plugin in the test process with `connectInProcess`: conformance, a cancelled scan, four scans at once (`node --test`) |
| `plugin.cmd` | a Windows launcher |
| `plugin.json` | how to install it (the Unix launcher is `./plugin.ts`; use `./plugin.cmd` on Windows) |
| `package.json` | `"type": "module"`, the SDK as the one dependency, the test script |

Run with Node.js 24, which runs TypeScript files as they are. Inside this repository the SDK resolves to the
neighbouring package once the workspace is installed (`npm install` at the root); in your own project, `npm install`.

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
[plugins.external.widget-scan]
command = ["node", "/path/to/plugin.ts"]

[conventions.widget-scan]
```
