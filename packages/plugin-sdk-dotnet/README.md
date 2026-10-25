# MonoSmith.Plugin.Sdk

The C# SDK for writing [smith](https://github.com/monosmith-app/monosmith) plugins: the protocol's
types, the serve loop, a host to test a plugin with, the conformance checks of `smith plugin check`,
`[[ ]]` templating and JSON editing that keeps key order. No dependencies; .NET 8 or later (it runs on
every later runtime); nullable annotations and XML documentation throughout.

```sh
dotnet add package MonoSmith.Plugin.Sdk
```

A plugin is a class that implements `IPlugin` and one interface for each thing it does. The SDK frames
the JSON-RPC messages, answers `initialize`, derives the capabilities from the interfaces the class
implements and keeps standard output for the protocol:

```csharp
using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

return await PluginServer.RunAsync(new WidgetLint());

internal sealed class WidgetLint : IPlugin, IContributor
{
    public PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "widget-lint", Version = "0.1.0" },
        Kind = PluginKind.Convention,
        Roles = [Role.Structure],
        Languages = ["widget"],
    };

    public Task<ContributeResult> ContributeAsync(ContributeParams parameters, CancellationToken cancellation) =>
        Task.FromResult(Contribute.To(parameters.Projects, project => new Contribution
        {
            Tasks = new Dictionary<string, ProjectTask> { ["layout-check"] = ProjectTasks.Create(["dotnet", "check.dll"], cache: true) },
            DependsOn = new Dictionary<string, IReadOnlyList<string>> { ["lint"] = ["layout-check"] },
        }));
}
```

Test it in-process, with the real handshake and framing, and run the same checks as
`smith plugin check`:

```csharp
using MonoSmith.Plugin.Sdk.Testing;

await using var client = await PluginClient.ConnectInProcessAsync(new WidgetLint(), new HostOptions { WorkspaceRoot = workspace });
var report = await ConformanceChecks.CheckPluginAsync(client);
Assert.True(report.Ok, report.Failures());
```

The full guide and API reference:
[The C# SDK](https://github.com/monosmith-app/monosmith/blob/main/docs/plugins/creating/csharp-sdk.md).
Examples that use it:
[`structure-layout-cs`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/structure-layout-cs),
[`structure-scan-cs`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/structure-scan-cs) and
[`language-widget-cs`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/language-widget-cs).

## The protocol schema is the source of truth

The types in `Protocol/ProtocolTypesContract.cs`, the interfaces in `Serve/PluginContract.cs`, the method
table in `Serve/PluginMethodsContract.cs` and the typed calls in `Host/PluginClientCallsContract.cs` are
not written by hand. The Go types of the protocol (`packages/plugin-sdk/protocol`) are the source of truth;
their JSON Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../plugin-sdk/schema/protocol.schema.json), and the
Node script in `Generator` turns that into C#: a record for each message that reads and writes itself as
the Go SDK does, the error codes, the method names, the roles, one interface for each capability. After a
change to the protocol:

```sh
smith run schema -p plugin-sdk            # the Go side writes the schema
smith run generate -p plugin-sdk-dotnet   # this side writes the C#
```

A test (`tests/Integration/GeneratedProtocolTests.cs`) and `smith run generate-check -p plugin-sdk-dotnet`
fail when a generated file is stale. The generator itself fails when the schema has a method that no
interface in `Generator/plugin-api.config.mjs` answers, so a new method has to be given its place in the
plugin.

## What is where

| Folder | What it holds |
|---|---|
| `Json` | `JsonAny` and its kinds, a parser and a writer with the semantics of Go's `encoding/json` (key order, number format, escaping) |
| `Protocol` | the generated messages, the constants, `RpcError`, `ParamsError`, and how a message reads and writes itself |
| `Framing` | `Content-Length` framing over a byte stream |
| `Codec` | the JSON-RPC envelopes: requests, notifications, results and errors |
| `Cancellation` | `RequestCancelledError` and `CancellationToken.ThrowIfCancelled`, what a cancellable method calls |
| `Concurrency` | what a declared `concurrency` means: one at a time by default, a count, or unlimited |
| `Serve` | `PluginServer`: the loop, a bounded pool of workers, `cancel`, the plugin interfaces, capability derivation |
| `Stdio` | taking standard input and output for the protocol and leaving nothing else a way to them |
| `Host` | `PluginClient`: start a plugin process, connect over streams or in-process; for tests; it keeps to the plugin's concurrency and cancels a call with a `CancellationToken` |
| `Conformance` | `ConformanceChecks.CheckPluginAsync`: the checks of `smith plugin check` |
| `Templating`, `JsonDocuments` | ports of the Go SDK's `templating` and `jsondoc` |
| `Tasks`, `Scaffold`, `Ci` | helpers to build the answers of contribute, generate and ci |
| `BuildDir`, `PluginRoles`, `PluginRefs`, `Projects` | the protocol package's small functions: build directories, roles, plugin references, project kinds, manifests |

A slice is a folder of files named `<Name><Role>.cs`; the slices depend on one another in one direction
(`Json` and `Protocol` at the bottom, `Serve` above `Stdio`, `Framing`, `Codec` and `Concurrency`, `Host`
above `Serve`, `Conformance` above `Host`), and `smith slices check --lang dotnet .` lints the layout.
Public types of the SDK live in the namespace `MonoSmith.Plugin.Sdk`, the messages in
`MonoSmith.Plugin.Sdk.Protocol` and the test host and the conformance checks in
`MonoSmith.Plugin.Sdk.Testing`, whatever the folder: one `using` is all a plugin needs.

## Standard output

A plugin talks to smith on standard output, so a stray `Console.WriteLine` would corrupt the stream.
`PluginServer.RunAsync` takes private copies of the two standard handles for the protocol, and from then on
standard output is standard error and standard input is the null device: on Windows with `DuplicateHandle`
and `SetStdHandle`, elsewhere with `dup` and `dup2`. The copies are marked close-on-exec so that a child
process the plugin starts does not keep the stream open, except on macOS with ARM64, where `fcntl` cannot
be called as declared (its variable arguments travel differently there). This was checked on Windows; the
Linux and macOS code is exercised by the tests in CI (Linux).

## Develop

```sh
smith run lint typecheck test build generate-check -p plugin-sdk-dotnet
```

`lint` is `dotnet format --verify-no-changes` (style and analyzers; warnings are errors). The tests run on
`net10.0`; `dotnet test tests/MonoSmith.Plugin.Sdk.Tests.csproj -p:TestTargetFramework=net8.0` runs them on the oldest runtime the SDK
targets. The integration tests need `smith`, `go`, `node`, `dotnet` and, for the Python side, a Python with
`monosmith-plugin-sdk` installed; they skip themselves without them (in CI, where they exist, a missing one
fails the test). The package is released by smith from conventional commits, like the other packages of
this repository.
