# The C# SDK

`MonoSmith.Plugin.Sdk` is the SDK for writing smith plugins in C#. It does what the [Go
SDK](plugin-sdks.md#go-with-the-sdk), the [TypeScript SDK](typescript-sdk.md) and the [Python
SDK](python-sdk.md) do: it frames the messages, answers `initialize`, derives the capabilities from the
interfaces your plugin implements, gives you a host to test it with and the conformance checks `smith
plugin check` runs, and ports the Go `templating` and `jsondoc` packages. It has no dependencies, uses
nullable annotations and has XML documentation on every public member, so the whole API shows in your
editor. This page walks through a plugin, then lists the whole API.

```sh
dotnet add package MonoSmith.Plugin.Sdk
```

```csharp
using MonoSmith.Plugin.Sdk;            // the plugin interfaces, the serve loop, the helpers
using MonoSmith.Plugin.Sdk.Protocol;   // the messages: Project, Task, DiscoverParams, Finding, ...
using MonoSmith.Plugin.Sdk.Testing;    // the test host and the conformance checks
```

It targets .NET 8, the oldest long-term-support release, and runs on every later runtime. The package is
built from [`packages/plugin-sdk-dotnet`](../../../packages/plugin-sdk-dotnet/README.md) in this
repository and released by smith like the other packages. The NuGet package id is
`MonoSmith.Plugin.Sdk`.

## Where the types come from

The Go types in `packages/plugin-sdk/protocol` are the source of truth of the protocol. Their JSON
Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json):
one definition per message, a methods table (each of the 20 methods with its params, its result and the
capability it needs), the error codes, the roles and the protocol version. The C# is generated from that
file by a Node script (`Generator/generate-protocol.handler.mjs`), so it cannot drift from the Go types: a
Go test fails when the schema is stale, a C# test fails when the generated files are, and the generator
fails when the schema has a method no plugin interface answers.

Every message is a `sealed record` that reads itself from JSON and writes itself the way the Go SDK
writes it. A member is `required` in C# exactly when it is a string, an enumeration or a message that the
Go SDK writes without `omitempty` (a `Finding` needs `Check`; `Ok` is a `bool` and false is its default).
Lists and maps start empty and are written only when they are not. An optional message
(`Finding.Fix`, `Project.Registries`) is a nullable property. An enumeration (`Role`, `ProjectKind`,
`PluginKind`) is a `readonly record struct` around a string with a static member for each value
(`Role.Structure`) and an implicit conversion from `string`: it is **not validated at run time**, so a
value the schema does not list, from a newer host, is kept as it came. Integers are `long`. The free-form
members of the protocol (`settings`, `params`, `result`) are `JsonAny`, the SDK's own JSON tree, so the
SDK needs no dependency on `System.Text.Json` and writes numbers and escapes as Go does.

The same file is the contract for a language without an SDK. [The protocol cheat
sheet](protocol-cheat-sheet.md) shows it as tables.

## A plugin, step by step

The structure plugin `widget-layout` of the [structure plugin guide](structure-plugin.md), in C#: it
adds a `layout-check` task to widget projects, makes their `lint` wait for it, and warns about a project
without a README. The finished program is
[`structure-layout-cs`](../../examples/plugins/structure-layout-cs).

### 1. Describe it

A plugin is a class. `IPlugin.Describe` is the one method every plugin has; it answers `initialize` with
the plugin's identity. The SDK adds the protocol version and the capabilities.

```csharp
using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

public sealed class LayoutPlugin : IPlugin
{
    private string _workspaceRoot = string.Empty;

    public PluginDescription Describe(InitializeParams parameters)
    {
        _workspaceRoot = parameters.WorkspaceRoot; // absolute; the paths in later requests are relative to it
        return new PluginDescription
        {
            Plugin = new Peer { Name = "widget-layout", Version = "0.1.0" },
            Kind = PluginKind.Convention,
            Roles = [Role.Structure],
            Languages = ["widget"], // a convention that names no language applies to every language
        };
    }
}
```

and the entry point, in `Program.cs`:

```csharp
return await PluginServer.RunAsync(new LayoutPlugin());
```

`RunAsync` runs the protocol on standard input and output until smith sends `shutdown` and returns the exit
code of the process: 0 after `shutdown`, 1 (after the reason is printed on standard error) when the stream
fails. Standard output carries only the protocol's frames, and the SDK enforces that at the level of the
handles: from the moment `RunAsync` starts, standard output and `Console.Out` are standard error, so a
`Console.WriteLine` anywhere in the program, a library's write or a child process that inherits standard
output reaches smith's log and cannot corrupt the stream. Standard input is the null device, so a child
process that reads standard input cannot take smith's requests. Log with `Console.Error` all the same.

### 2. Implement the interfaces the plugin needs

Each interface the class implements is a capability the SDK advertises, and each method it lacks is
answered with `-32601`, `method not supported`, as in Go. The methods return a `Task<TResult>` and take the
params and a `CancellationToken`. Their names end in `Async`; the protocol's `version/write` is
`WriteVersionAsync`.

```csharp
public sealed class LayoutPlugin : IPlugin, IContributor, IDoctorContributor
{
    // ... Describe as above

    public Task<ContributeResult> ContributeAsync(ContributeParams parameters, CancellationToken cancellation) =>
        Task.FromResult(Contribute.To(parameters.Projects, project => new Contribution
        {
            Tasks = new Dictionary<string, ProjectTask>
            {
                ["layout-check"] = ProjectTasks.Create(["dotnet", Checker, "layout-check"], cache: true),
            },
            DependsOn = new Dictionary<string, IReadOnlyList<string>> { ["lint"] = ["layout-check"] },
        }));

    public Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation) =>
        Task.FromResult(new DoctorResult
        {
            Findings =
            [
                .. parameters.Projects.Select(project => new Finding
                {
                    Check = $"{project.Name} has a README.md",
                    Ok = File.Exists(Path.Combine(_workspaceRoot, project.Root, "README.md")),
                }),
            ],
        });
}
```

The params and results are the generated records, so the compiler checks the members as you type. The
handshake of this plugin advertises exactly what is implemented, `{"contribute":true,"doctor":true}`, and
nothing else.

What happens to what you return and throw:

| You | The host sees |
|---|---|
| return a result | the result, written with the members in the order of the schema and every empty optional member left out, exactly as the Go SDK writes it |
| return `null` | `{}`: right for a result with no required member (`remove`), a result that reads as empty for the others |
| throw an exception | `-32000` (plugin failure) and its message, which smith shows the user: put the remedy in it |
| throw while the request is cancelled | `-32800` (request cancelled), see below |
| are sent params of the wrong JSON type | `-32602` (invalid params) before your method is called, and the message names the member: `params.manifests must be an array, not a string` |
| return something the host cannot be sent | `-32603` (internal error) |

Params reach your method shaped as the types promise: where the Go host writes a nil slice as `null`, you
get an empty list, a required member that is absent is its zero value, and members the SDK does not know
are ignored, so the protocol can grow without breaking you. An optional member is simply its zero value
when it was not sent: `parameters.Stack.Length == 0`.

### Stop work when the host cancels it

A method that can run for long, `PublishAsync` or `DoctorAsync` over a big tree, should stop when smith
gives up on it (Ctrl+C, a timeout). Say the plugin is `Cancellable`: the SDK then cancels the
`CancellationToken` of a request when the host sends the `cancel` notification for it (and when the
plugin shuts down), and advertises the `cancel` capability. A method that stops throws (the host sees the
request-cancelled error `-32800`, whatever the message) or returns the result it already has.

```csharp
public sealed class ScanPlugin : IPlugin, IDoctorContributor
{
    public bool Cancellable => true;

    public PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "widget-scan", Version = "0.1.0" },
        Kind = PluginKind.Convention,
        Roles = [Role.Structure],
        Languages = ["widget"],
        Concurrency = 4,
    };

    public async Task<DoctorResult> DoctorAsync(DoctorParams parameters, CancellationToken cancellation)
    {
        foreach (var project in parameters.Projects)
        {
            cancellation.ThrowIfCancelled("the scan was stopped");   // between units of work
            await ScanAsync(project, cancellation);                  // and pass the token on
        }

        return new DoctorResult { Findings = [] };
    }
}
```

`Cancellable` is a default interface member of `IPlugin` that is `false`; implement the property to say
`true`. The token is the BCL's `CancellationToken`, so everything that takes one (`Task.Delay`,
`HttpClient`, `Process.WaitForExitAsync`) stops with it, and `cancellation.ThrowIfCancelled(message)` is
the SDK's convenience: it throws `RequestCancelledError`. A method that throws any exception while its
request is cancelled is answered with `-32800` and `request cancelled: <message>`, so a method that lets
`OperationCanceledException` escape is fine too.

`Concurrency` in what `Describe` returns says how many requests the plugin can work on at the same time.
Leave it out for one at a time, in order, which is what a plugin that says nothing gets; a number above 1
runs that many requests together, started in the order they arrive, with the answers written in the order
they finish. `ConcurrencyPolicy.Unlimited` (-1) leaves the number to the host. A plugin that is not
`Cancellable` is never sent a cancel and its token never fires.
[`structure-scan-cs`](../../examples/plugins/structure-scan-cs) is the whole example; its Go, TypeScript and
Python twins answer the same requests with the same bytes.

#### Threads, tasks and state

The serve loop reads requests with asynchronous I/O and runs each one on the thread pool, through a
**bounded pool of workers**, as many as `Concurrency` allows, with one writer for the frames. Your methods
are `Task`-returning, so use `async` where they wait; a method that blocks or computes is fine too, because
it runs on a pool thread, not on the loop's. A plugin that declares a concurrency above 1 owns its locking:
state kept in the plugin object needs a lock or a concurrent collection.

### 3. Test it

`PluginClient.ConnectInProcessAsync` serves the plugin object in your test process over in-memory pipes,
with the real framing, handshake and dispatch, and gives you the host side. `ConformanceChecks.CheckPluginAsync`
runs the checks of `smith plugin check`.

```csharp
using MonoSmith.Plugin.Sdk.Testing;

await using var client = await PluginClient.ConnectInProcessAsync(new LayoutPlugin(), new HostOptions { WorkspaceRoot = workspace });
var report = await ConformanceChecks.CheckPluginAsync(client);
Assert.True(report.Ok, report.Failures());

var core = new Project { Name = "core", Root = "libs/core", Language = "widget", Kind = ProjectKind.Lib };
var result = await client.ContributeAsync(new ContributeParams { Projects = [core] });
Assert.Equal(["layout-check"], result.Projects["core"].DependsOn["lint"]);
```

The example's tests, run with `dotnet test tests` from its folder:

```text
Passed!  - Failed:     0, Passed:     6, Skipped:     0, Total:     6, Duration: 168 ms - StructureLayout.Tests.dll (net8.0)
```

Every method has a typed call on the client (`client.DiscoverAsync(params)`, `client.DoctorAsync(params)`,
...) that returns the decoded result. An answer that is an error throws an `RpcError` that has the `Code`.
A call that is not answered within its timeout (two minutes unless the `timeout` argument says otherwise)
kills the plugin, because its stream can no longer be trusted, and throws a `CallInterruptedError`. Calls
may be made from several tasks; the client never has more in flight than the plugin declared, as smith does.
To cancel one, pass a `CancellationToken` and cancel it:

```csharp
using var source = new CancellationTokenSource();
var call = client.DoctorAsync(new DoctorParams { Projects = projects }, source.Token);
source.CancelAfter(TimeSpan.FromMilliseconds(500));
var interrupted = await Assert.ThrowsAsync<CallInterruptedError>(() => call);
Assert.Equal(InterruptionReason.Cancelled, interrupted.Reason);
Assert.False(interrupted.Killed);
Assert.Equal(ErrorCodes.RequestCancelled, interrupted.Response!.Code);
```

To test the real process, with the command smith would use, start it instead:

```csharp
await using var client = await PluginClient.StartAsync(new StartOptions
{
    Command = ["dotnet", "bin/Release/net8.0/StructureLayout.dll"],
    WorkspaceRoot = workspace,
});
```

Do not make a test depend on how long something takes. The SDK's own tests wait on latches the test controls
(a `ManualResetEventSlim` the plugin's method waits for, a counter the plugin signals when it has started),
never on a pause: a Linux runner under load is slower than your machine.

### 4. Run it under smith

```text
PS> smith plugin check "C:\Program Files\dotnet\dotnet.exe" .\bin\Release\net8.0\StructureLayout.dll
widget-layout 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ contribute answers with no projects
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

The lines are the ones `structure-layout-go`, `structure-layout-node` and `structure-layout-py` print. A
bare `dotnet` would be looked up as a **name**, and `dotnet` is the name of a built-in language plugin, so
`smith plugin check dotnet ...` would check that plugin and not yours: give the path of the `dotnet` program
(`which dotnet`, `Get-Command dotnet`), then the assembly. Declare it in `smith.toml` the same way:

```toml
[plugins.external.widget-layout]
command = ["dotnet", "/path/to/StructureLayout.dll"]

[conventions.widget-layout]
```

Inside `smith.toml` the bare `dotnet` is the program on the search path, which is what you want; only the
`smith plugin check` argument needs the path. Give `command` the `dotnet` program and the assembly, not the
`StructureLayout` executable that is next to it: that executable is a stub that loads the assembly, and the
digest that `smith plugin trust --pin` records covers the program and the files named in the arguments, so
only `dotnet` plus the `.dll` puts the code of your plugin in the digest (a plugin published as a single
file has no stub and is covered whole, so it may be named by its path). The assemblies it loads from
the same folder, `MonoSmith.Plugin.Sdk.dll` included, are not covered, as a package a script loads is not
(see [what smith does not protect against](../using/plugin-security.md#what-smith-does-not-protect-against)).
The assembly must be able to find `MonoSmith.Plugin.Sdk.dll` next to it, which a build or a `dotnet
publish` puts there. smith starts an external plugin with a minimal
environment; see [Plugin security](../using/plugin-security.md).

### 5. A language plugin

[`language-widget-cs`](../../examples/plugins/language-widget-cs) is the language plugin of the [language
plugin guide](language-plugin.md): `DiscoverAsync`, `GenerateAsync`, `WriteVersionAsync` and `DoctorAsync`.
It uses the helpers below: `ProjectTasks.Build` for a cached `build` task that writes into `dist/`,
`Scaffold.Project` to write a new project from templates, and `EditableJsonDocument` to change a
`widget.json` without reordering it. Its conformance run, with `smith plugin check <dotnet> <dll> --generate`,
has the same checks as the Go, TypeScript and Python plugins': it generates an app, a lib and a package and
discovers them again.

For the same requests a C# plugin, a Python plugin, a TypeScript plugin and a Go plugin with the same logic
give **byte-identical answers**: the SDK writes the members in the order of the Go structs, leaves out what
`omitempty` leaves out, sorts the keys of maps by their UTF-8 bytes and escapes `<`, `>`, `&` and
U+2028/U+2029 as Go does, while writing every other character as it is. The SDK's integration tests check
it for `language-widget-cs`, `structure-layout-cs` and `structure-scan-cs` against their twins, cancel
included, and compare answers that may come out of order by their id.

## The plugin interfaces

`PluginServer` takes any `IPlugin`. The interfaces are generated from the schema, one for each capability;
a plugin implements as many as it needs.

| Interface | Methods | Protocol method | Capability | Go interface |
|---|---|---|---|---|
| `IPlugin` | `Describe`, `Cancellable` | `initialize` | always | `Plugin.Describe` |
| `IDiscoverer` | `DiscoverAsync` | `discover` | `discover` | `Discoverer` |
| `IGenerator` | `GenerateAsync` | `generate` | `generate` | `Generator` |
| `IVersioner` | `WriteVersionAsync` | `version/write` | `version` | `Versioner` |
| `IPublisher` | `PublishAsync` | `publish` | `publish` | `Publisher` |
| `IRegistryReader` | `PublishedAsync` | `published` | `published` | `RegistryReader` |
| `ICiContributor` | `CiAsync` | `ci` | `ci` | `CIContributor` |
| `IDoctorContributor` | `DoctorAsync` | `doctor` | `doctor` | `DoctorContributor` |
| `IRemover` | `RemoveAsync` | `remove` | `remove` | `Remover` |
| `IKindChanger` | `ChangeKindAsync` | `project/kind` | `changeKind` | `KindChanger` |
| `IComponentLister` | `ComponentsAsync` | `sbom/components` | `components` | `ComponentLister` |
| `IContributor` | `ContributeAsync` | `contribute` | `contribute` | `Contributor` |
| `IDependencyManager` | `InstallAsync` and `UninstallAsync` | `dependencies/install`, `dependencies/uninstall` | `dependencies` | `DependencyManager` |
| `IDeclarationEditor` | `DeclaredAsync` and `RewriteAsync` | `dependencies/declared`, `dependencies/rewrite` | `declarations` | `DeclarationEditor` |
| `ILatestReader` | `LatestAsync` | `dependencies/latest` | `latest` | `LatestReader` |
| `Cancellable => true` | | the `cancel` notification | `cancel` | `Cancellable` |

A pair of methods is one interface, so the half-implemented plugin that the dynamic SDKs have to warn
about cannot be written. `initialize` and `shutdown` are answered by the SDK. A request before `initialize`
is refused with `-32600`, a notification gets no answer, and an unreadable frame body is answered with
`-32700`. `shutdown` is answered after every request read before it has been answered with its result; when
the input ends instead (the host is gone), a cancellable plugin has what is in flight cancelled and what is
waiting refused with `-32800`, and any other finishes what it is doing.

## Reference

### Serving

| Name | What it is |
|---|---|
| `PluginServer.RunAsync(plugin)` | runs the plugin on standard input and output; returns the exit code; the whole body of an entry point |
| `PluginServer.ServeAsync(plugin, new ServeStreams(input, output, log), cancellation)` | the same loop on streams you give it; returns at `shutdown` or the end of the input |
| `ClaimStdioUseCase.Claim()` | takes standard input and output for the protocol and returns the two streams; `RunAsync` calls it |
| `IPlugin`, `IDiscoverer`, `IGenerator`, `IVersioner`, `IPublisher`, `IRegistryReader`, `ICiContributor`, `IDoctorContributor`, `IRemover`, `IKindChanger`, `IContributor`, `IDependencyManager`, `IDeclarationEditor`, `ILatestReader` | the shapes of the plugin |
| `CancellationExtensions.ThrowIfCancelled(token, message?)` | throws `RequestCancelledError` when the work was asked to stop |
| `RequestCancelledError` | what `ThrowIfCancelled` throws |

### Testing a plugin

| Name | What it is |
|---|---|
| `PluginClient.ConnectInProcessAsync(plugin, HostOptions)` | serves a plugin object in this process over in-memory pipes and connects to it |
| `PluginClient.StartAsync(StartOptions)` | starts a plugin process with `SMITH_PLUGIN=1`, performs the handshake, forwards its standard error (`Stderr = StderrMode.Ignore` drops it) |
| `PluginClient.ConnectAsync(PluginStreams, HostOptions)` | the handshake over streams you have |
| `PluginClient` | `Info` (the `initialize` result), one typed call per method (`DiscoverAsync(params, cancellation, timeout)`), `CallAsync<TParams, TResult>`, `CallRawAsync(method, paramsJson)`, `NotifyAsync`, `Supports(method)`, `PeakInFlight`, `CancelGrace`, `CloseAsync()`; `IAsyncDisposable` |
| `CallInterruptedError` | a call given up on because its token was cancelled or its time ran out: `Reason` (`Cancelled` or `Timeout`), `Killed`, and `Response` (the plugin's `-32800`, when it answered) |
| `ConformanceChecks.CheckPluginAsync(client, workspace?)` | the checks of `smith plugin check`; with `workspace` (the directory the plugin was started in) also `--generate` |
| `ConformanceReport` | `Plugin`, `Results`, `Ok`, `Failures()`, `ToJson()` (the shape of `smith plugin check --json`) |
| `RpcError` | an error answer, with its `Code` and `Message` |

`HostOptions` are `WorkspaceRoot` (required, absolute), `Host` (name and version, default `smith` `dev`),
`InitializeTimeout` (30 seconds) and `CancelGrace` (5 seconds: how long a plugin has to answer after it was
asked to stop); `StartOptions` adds `Command`, `Environment`, `WorkingDirectory` and `Stderr`. A client never
has more calls in flight than the plugin declared in `concurrency`, as smith does: with the default of one,
calls are serialised. Cancel a call and the plugin is sent `cancel` if it advertised it; the call ends with
a `CallInterruptedError`, or with the result if the plugin had finished. A plugin that does not answer within
the grace period is killed. `CheckPluginAsync` is a port of the Go `conformance.Check` with the same check
names; `smith plugin check` adds one more, "shuts down when asked", which is `CloseAsync` not throwing. To
run the real command from a test, start `smith plugin check` as `tests/Integration/SmithPluginCheckTests.cs`
does.

### Helpers for the answers

| Name | What it is |
|---|---|
| `ProjectTasks.Create(command, cwd, env, inputs, outputs, dependsOn, cache, description, verify, serve, serveUrl, serveTimeout)` | a `ProjectTask`: an argv, never a shell string; what you leave out is not written |
| `ProjectTasks.Build(root, command, inputs, description)` | a cached build task: `outputs` and the directory `command(outputDirectory)` is told to write into are `../../dist/<root>`, and it depends on `^build` |
| `Contribute.To(projects, contribution)` | the answer to `contribute` when each project gets the same kind of addition; a project for which `contribution` returns `null` is left out |
| `Scaffold.Project(workspaceRoot, params, files, data, notes)` | the answer to `generate` for a scaffold: renders `files` (target path to template) with `ScaffoldData.From(params)` (`Dir`, `Name`, `Kind`, `Stack`, `Workspace`, `Scope`, `ModulePrefix`) and `data`, never overwriting |
| `CiSteps.Action(uses, inputs)`, `CiSteps.Command(name, command, dir, blocking, cacheFiles)`, `CiSteps.Credential(env, description, required, setup, tasks)` | the pieces of a `ci` answer: a provider-native step with ordered inputs, a command a phase runs (blocking unless told not to), a secret the release needs |
| `BuildDirectory.ForProject(root)`, `FromProject(root)`, `Contains(path)`, `Name` | where a project builds: `dist/apps/web`, and `../../dist/apps/web` from inside the project |
| `PluginRoles.All`, `IsKnown`, `IsExclusive`, `IsAdditive`, `DefaultFor(kind)`, `Effective(info)` | the roles, as the schema lists them |
| `PluginRefs.RequiresLanguage(name)`, `ConflictsWith(name)`, `IsEmpty(ref)`, `Describe(ref)` | plugin references |
| `ProjectKinds.Resolve(declared, root, fallback)`, `ProjectKinds.DirectoryOf(kind)`, `ProjectPublishing.Publishes(project)`, `ManifestMatching.IsManifest(manifests, baseName)`, `ToolchainRecording.Set(result, probe)`, `StackPlugins.Of(stack)`, `AuditPolicies.IsEnabled(audit, policy)`, `AuditPolicies.Blocks(audit, policy)` | the small functions of the Go `protocol` and `projectkind` packages |

### Templating

`Templates.Write(root, files, data)` renders every template of `files` (target path to text) under `root`
with the `[[ ]]` delimiters and returns the paths it wrote. Target paths may contain actions
(`[[.Dir]]/Program.cs`). It checks every target before it writes anything: an existing file is a
`TemplateFileExistsError` (an `IOException`) and nothing is written, a missing field is a `TemplateError`
(the Go `missingkey=error`), and a target outside `root` is refused. `Templates.Render(text, data, name)`
renders one string. The data is a `JsonMap` (`new JsonMap { ["Name"] = "core" }`), and the values are the
SDK's JSON tree.

The language is the subset of Go's `text/template` that project templates use. Anything else is reported
with the name of the construct, not rendered wrongly.

| Supported | Not supported |
|---|---|
| `[[.Name]]`, `[[.A.B]]`, `[[$.Name]]`, `[[.]]` | variables (`[[$x := ...]]`), `define`, `template`, `block`, `break`, `continue` |
| `[[if]]`, `[[else if]]`, `[[else]]`, `[[with]]`, `[[range]]`, `[[end]]` | methods on values |
| `eq ne lt le gt ge not and or len index print printf` (`%s %d %v %q %t %%`), pipes, parentheses | other functions |
| string, raw string, number, `true`, `false`, `nil` | |
| `[[- x -]]` trim markers, `[[/* comments */]]` | |

Text such as `${{ secrets.TOKEN }}` in a GitHub workflow passes through untouched, as in Go.

### Editing JSON

`EditableJsonDocument` is the Go `jsondoc.Document`: a JSON object that keeps its key order and the text of
everything you do not touch (`1.50` stays `1.50`), so editing a `package.json` changes only what you edited.
`EditableJsonDocument.Parse(text)`, `Has`, `Keys`, `Count`, `Object(key, create)` (an object you can edit in
place), `GetString`, `GetValue`, `SetString`, `SetJson` (a `JsonMap` becomes an editable document, its
members in the order you added them), `Delete`, and `Render()`, which writes two-space indentation and a final
newline as npm does. The output is the same bytes as the Go SDK's for the same input. Write the result as
bytes with a `UTF8Encoding(false)`, so that no byte order mark is added.

### The JSON tree

`JsonAny` is the base of `JsonNull`, `JsonBool`, `JsonNumber`, `JsonText`, `JsonList` and `JsonMap`.
`JsonAny.Parse(text)` reads JSON the way Go does (a key that repeats keeps its first place and its last
value, a number written as an integer is an exact `long`, a lone surrogate escape is U+FFFD) and `Encode()`
writes it the way Go writes a `map[string]any` (compact, keys sorted by their UTF-8 bytes, `<`, `>`, `&`
escaped, the shortest round-tripping number). `JsonBuilder` is the writer the generated records use; a
plugin rarely needs it.

### Building blocks

For a tool that speaks the protocol itself: `Framing.Frame(body)` and `FrameReader` (the `Content-Length`
framing), `ProtocolInfo.Version` and `ProtocolInfo.MaxMessageBytes`, `ErrorCodes` (`ParseError`,
`InvalidRequest`, `MethodNotFound`, `InvalidParams`, `InternalError`, `PluginFailure`, `RequestCancelled`),
`ProtocolMethods` (every method name and the capability it needs), `ProtocolRoles`, `ConcurrencyPolicy`,
`IProtocolValue<T>` (what a message implements: `T.Read(json, path)` and `Write(writer)`) and `ParamsError`
(params of the wrong shape).

## What differs from the other SDKs

| | Go | TypeScript | Python | C# |
|---|---|---|---|---|
| a plugin | a struct with methods; `serve.Main(&plugin{})` | an object with members; `await serve(plugin)` | an object with methods; `serve(plugin)` | a class that implements interfaces; `await PluginServer.RunAsync(plugin)` |
| capabilities | the interfaces the struct satisfies | the members the object has | the methods the object has | the interfaces the class implements |
| messages | structs | interfaces | `TypedDict`s | records |
| method names | `WriteVersion` | `writeVersion` | `write_version` | `WriteVersionAsync` |
| concurrency model | a goroutine pool | a promise pool on one thread | a pool of worker threads | a bounded pool of tasks on the thread pool |
| cancel | `serve.Cancellable` (`WithContext(ctx)`) | `cancellable: true`, an `AbortSignal` | `cancellable = True`, a `CancelToken` | `Cancellable => true`, a `CancellationToken` |
| standard output | the program's job | `process.stdout` is redirected | file descriptors 0 and 1 are redirected | the standard handles are redirected |
| error codes | an error becomes `-32000` | the same, plus `-32602` and `-32603` | the same as TypeScript | the same as TypeScript |
| `null` in params | a nil slice or map | `[]` or `{}` for a required member, removed for an optional one | the same as TypeScript | the same as TypeScript |
| the host's timeouts | `time.Duration` | milliseconds | seconds | `TimeSpan` |
| `toolrun` | `toolrun.Run` | not ported | not ported | not ported: use `Process` and send its output to standard error |

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `malformed header "..."` in smith | something wrote to standard output before `RunAsync` took it over, for example a `Console.WriteLine` in a static initializer: remove it, or write to `Console.Error` |
| `smith plugin check dotnet ...` prints `dotnet dev` and not your plugin | a bare `dotnet` is looked up as a plugin name: give the full path of the `dotnet` program |
| `Could not load file or assembly 'MonoSmith.Plugin.Sdk'` | the assembly smith runs has no `MonoSmith.Plugin.Sdk.dll` beside it: run the plugin from the build or publish output |
| `You must install or update .NET to run this application` | the plugin targets a runtime that is not installed; add `<RollForward>LatestMajor</RollForward>` to the plugin's project, or install the runtime |
| `method not supported: x` for a method you wrote | the class has to implement the interface that carries the method (`IVersioner` for `WriteVersionAsync`), and a pair is one interface |
| `-32603` in a test | the host could not be sent what the method returned |
| `plugin stopped answering: EOF` in a test | the plugin process died: its standard error is on yours, unless you passed `Stderr = StderrMode.Ignore` |
| a test hangs | a plugin that never answers is killed after the call's timeout (two minutes); pass `timeout:` to fail sooner |
| the token never fires | the plugin is not `Cancellable`, or the host did not send `cancel`: only a call made with a token that is cancelled (or whose time runs out) sends it |
| a cancelled method keeps running | the method must look at the token (`ThrowIfCancelled`, or any API that takes it) between steps, and pass it on to what it calls |
| data corrupted when two requests run together | with `Concurrency` above 1 the methods run on several threads: guard state shared through the plugin object with a lock |
| a byte order mark in a file the plugin wrote | `File.WriteAllText(path, text, Encoding.UTF8)` adds one; use `new UTF8Encoding(false)` or write bytes |

## See also

- [Writing a plugin in Go, Node.js or Python](plugin-sdks.md): what exists for each language, and the protocol by hand
- [The TypeScript SDK](typescript-sdk.md) and [The Python SDK](python-sdk.md): the same API in TypeScript and Python
- [The plugin protocol cheat sheet](protocol-cheat-sheet.md) and [the architecture of the protocol](../../architecture/plugin-protocol.md)
- [Testing a plugin](testing-a-plugin.md)
- [`packages/plugin-sdk-dotnet`](../../../packages/plugin-sdk-dotnet/README.md), the package
