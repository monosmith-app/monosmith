# toolchain-widget-lint

A smith **toolchain plugin** in Go. It fills the `lint` slot of widget projects (see
[`language-widget`](../language-widget)): the task fails when `widget.json` leaves a required field
empty. The required fields are a per-project setting. It publishes the stack `widget-linted` and, when
a project is created with it, writes a `WIDGET-LINT.md`.

Guide: [Creating a toolchain plugin](../../../plugins/creating/toolchain-plugin.md).

## Files

| File | What it is |
|---|---|
| `main.go` | serves the protocol for smith; runs the linter for the task |
| `plugin.go` | `Describe` (roles, `provides`, `requires`, stacks), `Contribute`, `Generate` |
| `linter.go` | the lint command line |
| `plugin_test.go` | conformance, settings and generate tests |
| `plugin.json` | how to install it |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget-lint .
smith plugin check ./smith-plugin-widget-lint
```

## Use

In a workspace that has the `widget` plugin:

```toml
[plugins.external.widget-lint]
command = ["/path/to/smith-plugin-widget-lint"]

# optional: the fields every widget of this project must set
[projects.extra.settings.widget-lint]
require = ["description", "license", "homepage"]
```

```sh
smith add lib extra --stack widget-linted
smith run lint
```
