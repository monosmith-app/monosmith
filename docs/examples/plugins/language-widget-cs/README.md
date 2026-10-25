# language-widget-cs

The language plugin of [`language-widget`](../language-widget), in **C# on
[`MonoSmith.Plugin.Sdk`](../../../../packages/plugin-sdk-dotnet)**. It teaches smith the widget project
type: a directory with a `widget.json` is a project, `uses` in it become graph edges, and each project
gets a cached `build` and `test` task. It also scaffolds widgets (`smith add lib core --lang widget`),
writes the release version into `widget.json` keeping its key order, and reports a doctor finding with
a fix.

For the same requests it gives the same bytes as the Go plugin, the [TypeScript
one](../language-widget-ts) and the [Python one](../language-widget-py), which
`packages/plugin-sdk-dotnet/tests/Integration` checks.

Guide: [Creating a language plugin](../../../plugins/creating/language-plugin.md); the SDK:
[The C# SDK](../../../plugins/creating/csharp-sdk.md).

## Files

| File | What it is |
|---|---|
| `Program.cs` | the entry point: serves the protocol when smith started the program (`SMITH_PLUGIN=1`), runs the command line tool otherwise |
| `WidgetPlugin.cs` | the plugin: `Describe`, `DiscoverAsync`, `GenerateAsync`, `WriteVersionAsync`, `DoctorAsync` |
| `WidgetManifest.cs` | reading, checking and rewriting `widget.json` |
| `WidgetTool.cs` | the command line the tasks call: `check`, `build <dir>` |
| `tests/WidgetPluginTests.cs` | the conformance suite and behaviour tests, over an in-process connection (`dotnet test`) |
| `workspace-test.sh` | an end-to-end test in a throwaway workspace |
| `plugin.json` | how to install it: the command is the program published as a single file |
| `Directory.Build.props` | empty on purpose, so the example does not inherit the repository's build settings |

Needs the .NET 8 SDK or later (the plugin targets `net8.0` and runs on a newer runtime too). The project
refers to the SDK's project in this repository; in your own project, drop that reference and
`dotnet add package MonoSmith.Plugin.Sdk`.

## Test, check

```sh
dotnet test tests
dotnet build -c Release
smith plugin check "$(which dotnet)" bin/Release/net8.0/LanguageWidget.dll --generate
sh workspace-test.sh
```

```text
PS> smith plugin check (Get-Command dotnet).Source .\bin\Release\net8.0\LanguageWidget.dll --generate
```

Give `smith plugin check` the **path** of the `dotnet` program: a bare `dotnet` is looked up as a plugin
name, and `dotnet` is the name of a built-in plugin.

## Use

```toml
[plugins.external.widget]
command = ["dotnet", "/path/to/LanguageWidget.dll"]
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run
`smith plugin trust widget` (`--yes` in a script, `--pin` to also write the digest CI checks). The digest
covers `dotnet` and the `.dll`, not the SDK's assembly next to it. To have the whole plugin in the digest,
publish it as one file and name that: `dotnet publish -c Release -p:PublishSingleFile=true -o publish`, then
`command = ["/path/to/publish/LanguageWidget"]`. It starts with a minimal environment; see [Plugin
security](../../../plugins/using/plugin-security.md).

```sh
smith add lib core --lang widget
smith add app shop --lang widget      # then add "uses": ["core"] to apps/shop/widget.json
smith graph
smith run build test
```
