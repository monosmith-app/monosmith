# scaffold-widget-service

A smith **scaffold plugin** in Go. `smith add app api --stack widget-service` writes a widget service:
`widget.json` (recording the stack), `service.json` with a port, a handler and a README. The port comes
from the plugin's settings in a recipe (`settings.widget-service.port`, default 8080). It publishes the
stack `widget-service`, which pairs it with the `widget-lint` toolchain.

Guide: [Creating a scaffold plugin](../../../plugins/creating/scaffold-plugin.md).

## Files

| File | What it is |
|---|---|
| `main.go` | `serve.Main(&scaffold{})` |
| `scaffold.go` | `Describe` (role, requirement, stack), `Generate`, the templates, the settings |
| `scaffold_test.go` | conformance, behaviour and golden-file tests |
| `testdata/golden/` | the reviewed output of the scaffold; `go test -update` rewrites it |
| `plugin.json` | how to install it |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget-service .
smith plugin check ./smith-plugin-widget-service
```

Do not add `--generate`: it asks for an app, a lib and a package, and a service is an app.

## Use

```toml
[plugins.external.widget-service]
command = ["/path/to/smith-plugin-widget-service"]
```

The stack names the `widget-lint` toolchain, so declare
[`toolchain-widget-lint`](../toolchain-widget-lint) as well. Then `smith add app api --stack widget-service`.
