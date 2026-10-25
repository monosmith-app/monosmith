# structure-layout-go

A smith **structure plugin** in Go. For widget projects it adds a `layout-check` task that fails on
folders named `util`, `utils`, `helper`, `helpers`, `common`, `shared` or `misc`, makes `lint` wait
for it, and warns in `smith doctor` about a project without a `README.md`. The same plugin is written
in [TypeScript](../structure-layout-node), on the TypeScript SDK, in [Python](../structure-layout-py), on the Python SDK,
and in [Python](../structure-layout-python), without an SDK.

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md).

## Files

| File | What it is |
|---|---|
| `main.go` | serves the protocol for smith; runs the check for the task |
| `plugin.go` | `Describe`, `Contribute`, `Doctor` |
| `check.go` | finding the junk drawers |
| `plugin_test.go` | conformance and behaviour tests |
| `plugin.json` | how to install it |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget-layout .
smith plugin check ./smith-plugin-widget-layout
```

## Use

```toml
[plugins.external.widget-layout]
command = ["/path/to/smith-plugin-widget-layout"]

[conventions.widget-layout]      # every widget project
exclude = ["legacy"]             # except these
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget-layout` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../../plugins/using/plugin-security.md).

or name it in a stack: `structure = ["widget-layout"]`.
