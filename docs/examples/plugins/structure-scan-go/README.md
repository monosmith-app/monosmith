# structure-scan-go

A smith **structure plugin** in Go whose `doctor` walks every file of a widget project and warns about
any over a mebibyte. It exists to show the two things a plugin does when a call can run for long:

- **It stops when smith cancels the request.** The plugin implements `serve.Cancellable`: `WithContext(ctx)`
  returns the plugin as it serves one request, `Serve` cancels that context when smith sends the `cancel`
  notification for the request (Ctrl+C, a timeout), the walk checks it before every file, and the plugin
  advertises the `cancel` capability. smith then gets the request-cancelled error (`-32800`) back in
  milliseconds instead of waiting for the walk, or killing the plugin.
- **It takes four requests at a time.** `Concurrency: 4` in what `Describe` returns says smith may have four
  requests in flight; `Serve` runs them on four workers. A plugin that says nothing is called one request at
  a time, which is the right default for a plugin that shares state.

The same plugin is written in [TypeScript](../structure-scan-node), on the TypeScript SDK, and in
[Python](../structure-scan-py), on the Python SDK; for the same requests, a cancel included, they answer with the
same bytes, which `packages/plugin-sdk-ts/integration` (Go and TypeScript) and
`packages/plugin-sdk-py/tests/integration` (Go, TypeScript and Python) check. The rule is in [the plugin protocol](../../../architecture/plugin-protocol.md#cancel-and-concurrency).

## Files

| File | What it is |
|---|---|
| `main.go` | serves the protocol for smith; `WIDGET_SCAN_DELAY_MS` adds a pause before each file, so a test can cancel a scan in the middle |
| `plugin.go` | `Describe` (with the concurrency), `WithContext`, `Doctor`, and the walk that watches the context |
| `plugin_test.go` | conformance, a cancelled scan, and four scans at once, over an in-process connection |
| `plugin.json` | how to install it |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget-scan .
smith plugin check ./smith-plugin-widget-scan
```

```text
widget-scan 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ declares a concurrency the protocol defines
  ✔ accepts 4 requests at the same time
  ✔ answers a cancelled request with its result or the cancelled error
  ✔ ignores a cancel for a request it does not know
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

## Use

```toml
[plugins.external.widget-scan]
command = ["/path/to/smith-plugin-widget-scan"]

[conventions.widget-scan]      # every widget project
```
