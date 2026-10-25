# Plugin examples

Working smith plugins and a recipe, small enough to read and real enough to run. They teach one
made-up project type, the **widget** (a directory with a `widget.json`), one level of the plugin
model at a time. Each one is explained step by step in
[Creating plugins](../../plugins/creating/README.md).

| Directory | Kind | Language | Guide |
|---|---|---|---|
| [`language-widget`](language-widget) | language plugin: discover, generate, version/write, doctor | Go, on the SDK | [Creating a language plugin](../../plugins/creating/language-plugin.md) |
| [`toolchain-widget-lint`](toolchain-widget-lint) | toolchain plugin: the `lint` slot, settings | Go | [Creating a toolchain plugin](../../plugins/creating/toolchain-plugin.md) |
| [`structure-layout-go`](structure-layout-go) | structure plugin: a layout check, doctor | Go | [Creating a structure plugin](../../plugins/creating/structure-plugin.md) |
| [`structure-layout-node`](structure-layout-node) | the same structure plugin | Node.js, no dependencies | [Writing a plugin in Go, Node.js or Python](../../plugins/creating/plugin-sdks.md) |
| [`structure-layout-python`](structure-layout-python) | the same structure plugin | Python, standard library | [Writing a plugin in Go, Node.js or Python](../../plugins/creating/plugin-sdks.md) |
| [`scaffold-widget-service`](scaffold-widget-service) | scaffold plugin: templates, settings, golden test | Go | [Creating a scaffold plugin](../../plugins/creating/scaffold-plugin.md) |
| [`deploy-widget-bundle`](deploy-widget-bundle) | deploy plugin: a task, `publish`, `ci`, doctor | Go | [Creating a deploy plugin](../../plugins/creating/deploy-plugin.md) |
| [`recipe-widget-team`](recipe-widget-team) | a recipe choosing the plugins above | TOML | [Creating a recipe and a stack](../../plugins/creating/recipe-and-stack.md) |
| [`go`](go), [`node`](node), [`python`](python) | the smallest language plugin (discovery only), three ways | Go, Node.js, Python | [Writing a plugin](../../guides/writing-a-plugin.md) |

## Building and checking

The Go examples have their own `go.mod`. Inside this repository they point at the neighbouring SDK
with a `replace` line; in your own module, drop the line and `go get` a released SDK. Because the
folders are not listed in the repository's `go.work`, build them with `GOWORK=off`:

```sh
cd language-widget
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget .
smith plugin check ./smith-plugin-widget --generate
```

(On Windows the binary is `smith-plugin-widget.exe`, and `smith plugin check` of a script needs a
`.cmd` launcher: `smith plugin check .\plugin.cmd` in PowerShell.) The Node.js example tests with
`node --test`, the Python one with `python -m unittest`.

Every example passes `smith plugin check`, except that `scaffold-widget-service` does not pass
`--generate`, which asks for an app, a lib and a package and the service is app-only. The commands, the
output and the workspace they were run in are pasted in the guides.

## Using them together

The widget plugins are designed to be declared side by side in one workspace; the recipe in
`recipe-widget-team` chooses them. A `smith.toml` that uses all of them:

```toml
[workspace]
name = "wd"
recipe = "recipes/widget-team.recipe.toml"

[plugins.external.widget]
command = ["/path/to/smith-plugin-widget"]

[plugins.external.widget-lint]
command = ["/path/to/smith-plugin-widget-lint"]

[plugins.external.widget-layout]
command = ["/path/to/smith-plugin-widget-layout"]

[plugins.external.widget-service]
command = ["/path/to/smith-plugin-widget-service"]

[plugins.external.widget-bundle]
command = ["/path/to/smith-plugin-widget-bundle"]
```

The tasks of the Go examples run the plugin's own executable, so they need no other tool; a real
plugin names the tool it wraps. A build output (`smith-plugin-*`) is ignored by each example's
`.gitignore`.
