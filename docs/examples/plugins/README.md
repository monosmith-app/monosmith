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
| [`structure-layout-node`](structure-layout-node) | the same structure plugin | TypeScript, on the SDK | [The TypeScript SDK](../../plugins/creating/typescript-sdk.md) |
| [`structure-scan-go`](structure-scan-go) | a structure plugin whose `doctor` can be cancelled and runs four at a time | Go, on the SDK | [The plugin protocol](../../architecture/plugin-protocol.md#cancel-and-concurrency) |
| [`structure-scan-node`](structure-scan-node) | the same plugin | TypeScript, on the SDK | [The TypeScript SDK](../../plugins/creating/typescript-sdk.md) |
| [`structure-layout-py`](structure-layout-py) | the same structure plugin | Python, on the SDK | [The Python SDK](../../plugins/creating/python-sdk.md) |
| [`structure-scan-py`](structure-scan-py) | the same cancellable, concurrent plugin as `structure-scan-go` | Python, on the SDK | [The Python SDK](../../plugins/creating/python-sdk.md#stop-work-when-the-host-cancels-it) |
| [`structure-layout-python`](structure-layout-python) | the same structure plugin | Python, standard library, no SDK | [Writing a plugin in Go, Node.js or Python](../../plugins/creating/plugin-sdks.md) |
| [`language-widget-ts`](language-widget-ts) | the same language plugin as `language-widget` | TypeScript, on the SDK | [The TypeScript SDK](../../plugins/creating/typescript-sdk.md) |
| [`language-widget-py`](language-widget-py) | the same language plugin as `language-widget` | Python, on the SDK | [The Python SDK](../../plugins/creating/python-sdk.md) |
| [`structure-layout-cs`](structure-layout-cs) | the same structure plugin | C#, on the SDK | [The C# SDK](../../plugins/creating/csharp-sdk.md) |
| [`structure-scan-cs`](structure-scan-cs) | the same cancellable, concurrent plugin as `structure-scan-go` | C#, on the SDK | [The C# SDK](../../plugins/creating/csharp-sdk.md#stop-work-when-the-host-cancels-it) |
| [`language-widget-cs`](language-widget-cs) | the same language plugin as `language-widget` | C#, on the SDK | [The C# SDK](../../plugins/creating/csharp-sdk.md) |
| [`scaffold-widget-service`](scaffold-widget-service) | scaffold plugin: templates, settings, golden test | Go | [Creating a scaffold plugin](../../plugins/creating/scaffold-plugin.md) |
| [`deploy-widget-bundle`](deploy-widget-bundle) | deploy plugin: a task, `publish`, `ci`, doctor | Go | [Creating a deploy plugin](../../plugins/creating/deploy-plugin.md) |
| [`recipe-widget-team`](recipe-widget-team) | a recipe choosing the plugins above | TOML | [Creating a recipe and a stack](../../plugins/creating/recipe-and-stack.md) |
| [`go`](go), [`node`](node), [`python`](python) | the smallest language plugin (discovery only), three ways; `node` and `python` write the framing by hand, with no SDK | Go, Node.js, Python | [Writing a plugin](../../guides/writing-a-plugin.md) |

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
`.cmd` launcher: `smith plugin check .\plugin.cmd` in PowerShell.) The TypeScript examples
(`structure-layout-node`, `language-widget-ts`) run as they are on Node.js 24, which only removes the
types, and test with `node --test`; they import `@monosmith/plugin-sdk`, which inside this repository is
the neighbouring `packages/plugin-sdk-ts` once the workspace is installed (`npm install` at the root),
and which in your own project you `npm install`. The Python examples on the SDK (`language-widget-py`,
`structure-layout-py`, `structure-scan-py`) test with `python -m unittest` and import
`monosmith_plugin_sdk`, which inside this repository is the editable install of `packages/plugin-sdk-py` in
the workspace's `.venv` (`smith ci setup`), and which in your own project you `pip install monosmith-plugin-sdk`;
check one with the interpreter that has the SDK: `smith plugin check <path of that python> plugin.py`. The
Python example without an SDK tests with `python -m unittest` too. The C# examples (`language-widget-cs`,
`structure-layout-cs`, `structure-scan-cs`) are projects of their own with an empty `Directory.Build.props`, so they do not
inherit the repository's build settings; each has a `tests` project (`dotnet test tests`) and refers to the neighbouring
`packages/plugin-sdk-dotnet` inside this repository, which in your own project you replace with `dotnet add package
MonoSmith.Plugin.Sdk`. Check one with the path of the `dotnet` program and then the assembly (`smith plugin check "$(which
dotnet)" bin/Release/net8.0/LanguageWidget.dll --generate`): a bare `dotnet` would be looked up as the name of the built-in
plugin.
`packages/plugin-sdk-ts/integration` runs both TypeScript examples under `smith plugin check`, in a
throwaway workspace, and against their Go counterparts; `packages/plugin-sdk-py/tests/integration` does the
same for the Python examples and compares Go, TypeScript and Python; `packages/plugin-sdk-dotnet/tests/Integration` does it for
the C# examples and compares all four.

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

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../plugins/using/plugin-security.md).

The tasks of the Go examples run the plugin's own executable, so they need no other tool; a real
plugin names the tool it wraps. A build output (`smith-plugin-*`) is ignored by each example's
`.gitignore`.
