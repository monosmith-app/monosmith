# structure-layout-go

A smith **structure plugin** in Go. For widget projects it adds a `layout-check` task that fails on
folders named `util`, `utils`, `helper`, `helpers`, `common`, `shared` or `misc`, makes `lint` wait
for it, and warns in `smith doctor` about a project without a `README.md`. The same plugin is written
without an SDK in [Node.js](../structure-layout-node) and [Python](../structure-layout-python).

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

or name it in a stack: `structure = ["widget-layout"]`.
