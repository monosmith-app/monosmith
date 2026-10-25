# structure-scan-cs

The structure plugin of [`structure-scan-go`](../structure-scan-go), written in **C# on
[`MonoSmith.Plugin.Sdk`](../../../../packages/plugin-sdk-dotnet)**. Its `doctor` walks every file of a
widget project and warns about any over a mebibyte, and it shows the two things a plugin does when a call
can run for long:

- **It stops when smith cancels the request.** The plugin says `Cancellable => true`, so the SDK cancels the
  `CancellationToken` of a request when smith sends the `cancel` notification for it (Ctrl+C, a timeout),
  and advertises the `cancel` capability. The walk waits between files with `Task.Delay(delay, cancellation)`
  and then `cancellation.ThrowIfCancelled(...)`, so it ends within one file's time.
- **It takes four requests at a time.** `Concurrency = 4` in what `Describe` returns says smith may have
  four requests in flight, and the SDK runs them together. A plugin that says nothing is called one request
  at a time.

For the same requests, a cancel included, this plugin, the Go one, the TypeScript one and the Python one
answer with the same bytes, which `packages/plugin-sdk-dotnet/tests/Integration` checks. The rule is in
[the plugin protocol](../../../architecture/plugin-protocol.md#cancel-and-concurrency); the SDK in
[The C# SDK](../../../plugins/creating/csharp-sdk.md#stop-work-when-the-host-cancels-it).

## Files

| File | What it is |
|---|---|
| `ScanPlugin.cs` | the plugin: `ScanPlugin(delay)` (`Describe` with the concurrency, `DoctorAsync`) and the walk that watches the token |
| `Program.cs` | the entry point that serves it; `WIDGET_SCAN_DELAY_MS` adds a pause before each file, so a test can keep a scan running until it cancels it |
| `tests/ScanPluginTests.cs` | tests that serve the plugin in the test process with `ConnectInProcessAsync`: conformance, a cancelled scan, four scans at once; they wait on latches, never on a pause (`dotnet test`) |
| `plugin.json` | how to install it: the command is the program published as a single file |
| `Directory.Build.props` | empty on purpose, so the example does not inherit the repository's build settings |

Needs the .NET 8 SDK or later. The project refers to the SDK's project in this repository; in your own
project, drop that reference and `dotnet add package MonoSmith.Plugin.Sdk`.

## Test, check

```sh
dotnet test tests
dotnet build -c Release
smith plugin check "$(which dotnet)" bin/Release/net8.0/StructureScan.dll
```

```text
PS> smith plugin check (Get-Command dotnet).Source .\bin\Release\net8.0\StructureScan.dll
```

Give `smith plugin check` the **path** of the `dotnet` program: a bare `dotnet` is looked up as a plugin
name, and `dotnet` is the name of a built-in plugin.

## Use

```toml
[plugins.external.widget-scan]
command = ["dotnet", "/path/to/StructureScan.dll"]

[conventions.widget-scan]
```
