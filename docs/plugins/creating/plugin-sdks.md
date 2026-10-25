# Writing a plugin in Go, Node.js or Python

A plugin is any program that speaks the protocol on its standard input and output, so it can be
written in any language. What differs is how much of the protocol you write yourself. This page says
exactly what exists, what each choice costs, and shows the same plugin in Go, TypeScript, Python and C#.

## What exists

| Language | Support | You write |
|---|---|---|
| **Go** | the plugin SDK, `github.com/monosmith-app/monosmith/packages/plugin-sdk`: typed messages, the serving loop, the host side for tests, the conformance suite, templating and JSON helpers | the plugin's methods |
| **TypeScript** (and JavaScript) | the plugin SDK, [`@monosmith/plugin-sdk`](typescript-sdk.md) on npm: the same, with the types generated from the protocol's JSON Schema. Examples: [`structure-layout-node`](../../examples/plugins/structure-layout-node) (a structure plugin) and [`language-widget-ts`](../../examples/plugins/language-widget-ts) (a language plugin) | the plugin's methods |
| **Node.js** without the SDK | working example that frames the messages by hand: [`node`](../../examples/plugins/node) (a language plugin) | the framing loop and the dispatch, about 40 lines, plus the methods |
| **Python** | the plugin SDK, [`monosmith-plugin-sdk`](python-sdk.md) on PyPI (import `monosmith_plugin_sdk`, Python 3.10 or later, no dependencies): the same, with the types generated from the protocol's JSON Schema. Examples: [`structure-layout-py`](../../examples/plugins/structure-layout-py) (a structure plugin), [`language-widget-py`](../../examples/plugins/language-widget-py) (a language plugin) and [`structure-scan-py`](../../examples/plugins/structure-scan-py) (cancel and concurrency) | the plugin's methods |
| **Python** without the SDK | working examples that frame the messages by hand, standard library only: [`python`](../../examples/plugins/python) and [`structure-layout-python`](../../examples/plugins/structure-layout-python) | the framing loop and the dispatch, plus the methods |
| **C#** | the plugin SDK, [`MonoSmith.Plugin.Sdk`](csharp-sdk.md) on NuGet (.NET 8 or later, no dependencies): the same, with the types generated from the protocol's JSON Schema. Examples: [`structure-layout-cs`](../../examples/plugins/structure-layout-cs) (a structure plugin), [`language-widget-cs`](../../examples/plugins/language-widget-cs) (a language plugin) and [`structure-scan-cs`](../../examples/plugins/structure-scan-cs) (cancel and concurrency) | the plugin's methods |
| anything else | no SDK, no example. The protocol is the contract: [`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json) is its JSON Schema, generated from the Go types, with a methods table, the error codes and the roles | everything. Not tried here |

The SDKs for the other languages are tracked in [#564](https://github.com/monosmith-app/monosmith/issues/564):
TypeScript first, then Python, then C#, all done. All of them are generated from the same JSON Schema, so a
protocol change reaches every language.

## Go, with the SDK

The SDK is a Go module with no dependencies of its own:

```sh
go get github.com/monosmith-app/monosmith/packages/plugin-sdk@latest
```

Released as the Go module tag `packages/plugin-sdk/v<version>`. Before 1.0 a minor version may change
the Go API; the wire protocol is versioned on its own (`protocol.Version`, today 1), and a breaking
change to it raises that integer. The scaffold [`go-smith-plugin`](../using/catalogue/go-smith-plugin.md)
pins a released version in the module it writes. The examples in this repository `replace` it with
the neighbouring source so they always build against the code in the tree; in your own module,
drop the `replace` line.

| Package | What it is for |
|---|---|
| `protocol` | the messages: every method's params and results, project kinds, tasks, stacks, CI contributions |
| `serve` | `serve.Main(plugin)` answers the host on stdin and stdout, and derives the capabilities from the interfaces the plugin implements |
| `host` | the host side: `host.Start` a process or `host.Connect` over pipes, handshake, calls with timeouts. For tests |
| `conformance` | `conformance.Check(client, opts)`: the checks `smith plugin check` runs, for your own tests |
| `templating` | `templating.Write`: render `[[ ]]` templates into files, never overwriting |
| `jsondoc` | edit JSON such as `package.json` in place, keeping key order |
| `projectkind` | a project's kind when its manifest declares none, and the directory each kind lives in |
| `toolrun` | run toolchain commands with their output on stderr, keeping stdout for the protocol |
| `transport` | `Content-Length` framing over a byte stream |

A plugin is a type with `Describe`, plus one method for each capability it wants:

```go
type widget struct{ root string }

func (w *widget) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) { /* ... */ }
func (w *widget) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error)     { /* ... */ }

func main() { serve.Main(&widget{}) }
```

The interfaces in `serve` are `Discoverer`, `Generator`, `Versioner`, `Publisher`, `RegistryReader`,
`CIContributor`, `DoctorContributor`, `Remover`, `DependencyManager`, `DeclarationEditor`,
`LatestReader` and `Contributor`; [the cheat sheet](protocol-cheat-sheet.md#methods) maps each to its
method. An error a method returns becomes a `-32000` response with your message. Two optional features are
opt-in: a `Concurrency` in what `Describe` returns (the host then sends up to that many requests at once, and
`Serve` runs them on a bounded pool), and `serve.Cancellable`, whose `WithContext(ctx)` returns the plugin as it
serves one request, with handlers that stop when `ctx` ends (the plugin then advertises `cancel`). A plugin
without either is called one request at a time and never sent a cancel;
[`structure-scan-go`](../../examples/plugins/structure-scan-go) shows both. `serve.Main` also
answers `shutdown`, rejects calls before `initialize` and answers an unknown method with `-32601`.

## TypeScript, with the SDK

```sh
npm install @monosmith/plugin-sdk
```

A plugin is an object with `describe` and one member for each method it has. The SDK does everything the
Go `serve.Main` does, and derives the capabilities from the members:

```ts
#!/usr/bin/env node
import { serve } from '@monosmith/plugin-sdk'

await serve({
  describe: () => ({ plugin: { name: 'kiwi', version: '0.1.0' }, kind: 'language', languages: ['kiwi'], manifests: ['kiwi.toml'] }),
  discover: ({ manifests }) => ({ projects: [] }),
})
```

The types are generated from the protocol's JSON Schema, a test host and the conformance checks are
included, and for the same requests the plugin answers with the same bytes as the Go one. [The
TypeScript SDK](typescript-sdk.md) has the walk-through and the whole API.

## Python, with the SDK

```sh
pip install monosmith-plugin-sdk
```

A plugin is an object with `describe` and one method for each thing it does; the messages are plain
dicts, typed by `TypedDict`s generated from the protocol's JSON Schema. The SDK does everything the Go
`serve.Main` does and derives the capabilities from the methods (`write_version` is `version/write`):

```python
from monosmith_plugin_sdk import serve


class Kiwi:
    def describe(self, params):
        return {"plugin": {"name": "kiwi", "version": "0.1.0"}, "kind": "language",
                "languages": ["kiwi"], "manifests": ["kiwi.toml"]}

    def discover(self, params):
        return {"projects": []}


if __name__ == "__main__":
    serve(Kiwi())
```

`serve` keeps standard output for the protocol at the file-descriptor level: a `print`, a library or a child
process that writes to standard output reaches standard error instead, and standard input is the null
device. Requests run on a bounded pool of worker threads; a plugin that sets `cancellable = True` has a
`CancelToken` passed as an optional second parameter. A test host, the conformance checks, templating and
key-order-keeping JSON editing are included, and for the same requests the plugin answers with the same
bytes as the Go and TypeScript ones. [The Python SDK](python-sdk.md) has the walk-through and the whole API.

## C#, with the SDK

```sh
dotnet add package MonoSmith.Plugin.Sdk
```

A plugin is a class that implements `IPlugin` and one interface for each thing it does; the messages are
records, generated from the protocol's JSON Schema. The SDK does everything the Go `serve.Main` does and derives
the capabilities from the interfaces (`WriteVersionAsync` of `IVersioner` is `version/write`):

```csharp
using MonoSmith.Plugin.Sdk;
using MonoSmith.Plugin.Sdk.Protocol;

return await PluginServer.RunAsync(new Kiwi());

internal sealed class Kiwi : IPlugin, IDiscoverer
{
    public PluginDescription Describe(InitializeParams parameters) => new()
    {
        Plugin = new Peer { Name = "kiwi", Version = "0.1.0" },
        Kind = PluginKind.Language,
        Languages = ["kiwi"],
        Manifests = ["kiwi.toml"],
    };

    public Task<DiscoverResult> DiscoverAsync(DiscoverParams parameters, CancellationToken cancellation) =>
        Task.FromResult(new DiscoverResult { Projects = [] });
}
```

`RunAsync` keeps standard output for the protocol at the level of the handles: a `Console.WriteLine`, a library
or a child process that writes to standard output reaches standard error instead, and standard input is the null
device. Requests run on a bounded pool of workers; a plugin that implements `Cancellable => true` has the
`CancellationToken` of a request cancelled when the host cancels it. A test host, the conformance checks,
templating and key-order-keeping JSON editing are included, and for the same requests the plugin answers with the
same bytes as the Go, TypeScript and Python ones. [The C# SDK](csharp-sdk.md) has the walk-through and the whole
API. Run the plugin as `["dotnet", "Kiwi.dll"]`, and give `smith plugin check` the path of `dotnet`, because a
bare `dotnet` is looked up as a plugin name.

## Without an SDK: Node.js and Python, by hand

There are three things to write, and they are the same in every language. (In Go, TypeScript and Python
the SDK writes them for you; this is what it does, and what a language without an SDK has to do.)

### 1. Read and write frames

A frame is `Content-Length: <bytes>`, a blank line (`\r\n\r\n`), then that many **bytes** of JSON.
Count bytes, not characters, and flush after every message.

```python
def read_message(stream):
    """Return the next message as a dict, or None when smith closed the pipe."""
    length = None
    while True:
        line = stream.readline()
        if not line:
            return None
        line = line.decode("ascii").strip()
        if line == "":
            break
        name, _, value = line.partition(":")
        if name.lower() == "content-length":
            length = int(value.strip())
    if length is None:
        return None
    return json.loads(stream.read(length))


def write_message(stream, message):
    body = json.dumps(message).encode("utf-8")
    stream.write(b"Content-Length: %d\r\n\r\n" % len(body))
    stream.write(body)
    stream.flush()
```

Use the **binary** streams (`sys.stdin.buffer`, `sys.stdout.buffer`) so Windows does not translate
newlines. In Node.js, messages arrive in arbitrary chunks, so buffer the bytes and cut them at the
frame boundaries:

```js
function send (message) {
  const body = Buffer.from(JSON.stringify(message), 'utf8')
  process.stdout.write(Buffer.concat([Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, 'ascii'), body]))
}

let buffer = Buffer.alloc(0)
process.stdin.on('data', (chunk) => {
  buffer = Buffer.concat([buffer, chunk])
  for (;;) {
    const end = buffer.indexOf('\r\n\r\n')
    if (end < 0) return
    const match = /content-length:\s*(\d+)/i.exec(buffer.subarray(0, end).toString('ascii'))
    if (!match) throw new Error('message without Content-Length')
    const length = Number(match[1])
    if (buffer.length < end + 4 + length) return
    const body = buffer.subarray(end + 4, end + 4 + length).toString('utf8')
    buffer = buffer.subarray(end + 4 + length)
    handle(JSON.parse(body))
  }
})
```

### 2. Dispatch on `method`

Answer every request that has an `id` with a response that has the same `id`: a `result` or an
`error`. A request without an `id` is a notification: do not answer it.

```python
HANDLERS = {"initialize": initialize, "contribute": contribute, "doctor": doctor, "shutdown": lambda params: {}}

request = read_message(stdin)
handler = HANDLERS.get(request["method"])
response = {"jsonrpc": "2.0", "id": request["id"]}
if handler is None:
    response["error"] = {"code": -32601, "message": "method not supported: " + request["method"]}
else:
    try:
        response["result"] = handler(request.get("params") or {})
    except Exception as e:  # -32000: a failure the user must act on
        response["error"] = {"code": -32000, "message": str(e)}
write_message(stdout, response)
if request["method"] == "shutdown":
    return
```

### 3. Declare exactly what you implement

`initialize` returns the identity and the capabilities. Nothing derives them for you here, so keep
them true: smith calls only the methods you advertise, and `smith plugin check` calls each one.

```json
{
  "protocolVersion": 1,
  "plugin": { "name": "widget-layout", "version": "0.1.0" },
  "kind": "convention",
  "roles": ["structure"],
  "languages": ["widget"],
  "capabilities": { "contribute": true, "doctor": true }
}
```

Check that `1` is in `protocolVersions` and raise an error if it is not.

### Rules for stdout

- **Only protocol frames go to stdout.** One stray `print` or `console.log` corrupts the stream and
  the call fails with `malformed header "..."`. Log to **stderr**: smith shows it to the user.
- smith sets `SMITH_PLUGIN=1`. The examples use it to be both the plugin and the command line tool the
  plugin's tasks run.
- Paths in messages are workspace-relative with forward slashes (`manifest.split("/")`,
  `replaceAll('\\', '/')`); only `workspaceRoot` is absolute.
- Exit after answering `shutdown`; exit when stdin closes.

## Starting an interpreted plugin

smith starts a plugin with an argv. A script is not an executable on Windows.

| Where | How |
|---|---|
| Linux, macOS | make the script executable with a shebang (`#!/usr/bin/env python3`) and use `./plugin.py` as the command. Not run on Linux or macOS for this page |
| Windows | a two-line `.cmd` launcher next to the script: `@echo off` then `python "%~dp0plugin.py" %*` |
| any | in `smith.toml`, `command = ["python", "tools/widget/plugin.py"]` starts the interpreter directly |

In an installed `plugin.json` only the first element of `command` is made relative to the plugin's
directory, so `["python", "plugin.py"]` does not work from there; use the launcher.

```text
> smith plugin check .\node\plugin.cmd
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

The same six lines for `python\plugin.cmd` and for the Go example. See
[Testing a plugin](testing-a-plugin.md#checking-an-interpreted-plugin) for how `smith plugin check`
decides whether its argument is a path or a name.

### Optional: concurrency and cancel

A plugin written as above handles one request at a time and never hears of cancel, which is what smith
expects of a plugin that says nothing. Two members of the `initialize` answer change that, and a loop that
reads the next frame while a handler runs is what they need:

- `"concurrency": 4` in the result says up to four requests may be in flight; smith never sends more.
  Answers may then be written in any order, each with the `id` of its request. `-1` leaves the number to
  smith, `0` or absent is one.
- `"cancel": true` in `capabilities` says the plugin stops a request when it reads the notification
  `{"jsonrpc":"2.0","method":"cancel","params":{"id":2}}` (no `id`, no answer to it). It then answers request 2
  with its result, if it had finished, or with the error `{"code":-32800,"message":"..."}`. A cancel for an id it does not know
  is ignored. smith waits five seconds for that answer and then kills the process.

[The protocol cheat sheet](protocol-cheat-sheet.md#a-cancelled-request) has the traffic of a cancelled
request, and the [architecture page](../../architecture/plugin-protocol.md#cancel-and-concurrency) the whole rule.

## The same plugin five ways

`widget-layout` (a [structure plugin](structure-plugin.md)):

| | Go | TypeScript | Python, SDK | C#, SDK | Python, by hand |
|---|---|---|---|---|---|
| source | `plugin.go`, `check.go`, `main.go` (146 lines) | `plugin.ts`, `layout.ts` (102 lines) | `plugin.py`, `layout.py` | `Program.cs`, `LayoutPlugin.cs` | `plugin.py` (153 lines) |
| framing and dispatch | `serve.Main` | `serve` | `serve` | `PluginServer.RunAsync` | written in the file |
| capabilities | derived from the methods | derived from the members | derived from the methods | derived from the interfaces | written in the `initialize` answer |
| tests | `go test` with `conformance.Check` and `host.Connect` | `node --test` with `checkPlugin` and `connectInProcess` | `python -m unittest` with `check_plugin` and `connect_in_process` | `dotnet test` with `ConformanceChecks.CheckPluginAsync` and `PluginClient.ConnectInProcessAsync` | `python -m unittest`, speaking the protocol through a child process |
| launcher on Windows | the `.exe` | `plugin.cmd` | `plugin.cmd` | `dotnet LayoutPlugin.dll`, or the program published as one file | `plugin.cmd` |
| dependencies | the SDK | `@monosmith/plugin-sdk` | `monosmith-plugin-sdk` | `MonoSmith.Plugin.Sdk` | none |

All of them give the same lines from `smith plugin check`, the same tasks and the same
`smith doctor` finding, and the Go, TypeScript, Python-on-SDK and C# plugins answer the same requests with the
same bytes. The Node.js plugin that frames the messages by hand is the smaller
[`node`](../../examples/plugins/node), a language plugin without tests.

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `plugin stopped answering: malformed header "hello ..."` | output to stdout other than frames |
| `plugin did not answer within 30s` | the loop never answers `initialize`: waiting for a newline-terminated line instead of a frame, or reading text-mode stdin |
| `plugin stopped answering: EOF` | the process died; read its stderr above |
| garbled non-ASCII text, or a hung call | the length counts characters, not bytes |
| `%1 is not a valid Win32 application` | a script is not an executable on Windows; use a `.cmd` launcher |
| a method is called that is not implemented, `method not supported: x` | the capability is advertised but not implemented |
| the `Content-Length` header has different case | the examples match `content-length` case-insensitively; do the same |

## See also

- [The plugin protocol cheat sheet](protocol-cheat-sheet.md), [Testing a plugin](testing-a-plugin.md)
- [`packages/plugin-sdk`](../../../packages/plugin-sdk/README.md) for the SDK's own README
- [The plugin protocol](../../architecture/plugin-protocol.md)
