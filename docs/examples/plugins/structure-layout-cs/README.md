# structure-layout-cs

The structure plugin of [`structure-layout-go`](../structure-layout-go), written in **C# on
[`MonoSmith.Plugin.Sdk`](../../../../packages/plugin-sdk-dotnet)**. The SDK frames the messages, answers
`initialize`, derives the capabilities from the two interfaces the plugin implements (`IContributor` and
`IDoctorContributor`) and keeps standard output for the protocol. It is both the plugin (when smith starts
it, `SMITH_PLUGIN=1`) and the checker the `layout-check` task runs (`dotnet StructureLayout.dll
layout-check`).

The same plugin with the framing and dispatch written by hand, no SDK, is
[`structure-layout-python`](../structure-layout-python) (and [`node`](../node) in Node.js). For the same
requests this plugin, the Go one, the TypeScript one and the Python one answer with the same bytes, which
`packages/plugin-sdk-dotnet/tests/Integration` checks.

Guide: [Creating a structure plugin](../../../plugins/creating/structure-plugin.md); the SDK:
[The C# SDK](../../../plugins/creating/csharp-sdk.md).

## Files

| File | What it is |
|---|---|
| `Program.cs` | the entry point: serves the plugin when smith started the program (`SMITH_PLUGIN=1`), runs the `layout-check` checker otherwise |
| `LayoutPlugin.cs` | the plugin: `LayoutPlugin` (`Describe`, `ContributeAsync`, `DoctorAsync`) and `FindJunkDrawers` |
| `tests/LayoutPluginTests.cs` | tests that serve the plugin in the test process with the SDK's `PluginClient.ConnectInProcessAsync` and run its conformance checks (`dotnet test`) |
| `plugin.json` | how to install it: the command is the program published as a single file |
| `Directory.Build.props` | empty on purpose, so the example does not inherit the repository's build settings |

Needs the .NET 8 SDK or later. The project refers to the SDK's project in this repository; in your own
project, drop that reference and `dotnet add package MonoSmith.Plugin.Sdk`.

## Test, check

```sh
dotnet test tests
dotnet build -c Release
smith plugin check "$(which dotnet)" bin/Release/net8.0/StructureLayout.dll
```

```text
PS> smith plugin check (Get-Command dotnet).Source .\bin\Release\net8.0\StructureLayout.dll
```

Give `smith plugin check` the **path** of the `dotnet` program: a bare `dotnet` is looked up as a plugin
name, and `dotnet` is the name of a built-in plugin.

## Use

```toml
[plugins.external.widget-layout]
command = ["dotnet", "/path/to/StructureLayout.dll"]

[conventions.widget-layout]
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run
`smith plugin trust widget-layout` (`--yes` in a script, `--pin` to also write the digest CI checks). To
have the whole plugin in the digest, publish it as one file (`dotnet publish -c Release
-p:PublishSingleFile=true -o publish`) and name that. It starts with a minimal environment; see [Plugin
security](../../../plugins/using/plugin-security.md).
