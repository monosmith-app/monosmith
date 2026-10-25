# Writing a plugin in Go, Node.js or Python

A plugin is any program that speaks the protocol on its standard input and output, so it can be
written in any language. What differs is how much of the protocol you write yourself. This page says
exactly what exists, what each choice costs, and shows the same plugin in Go, TypeScript and Python.

## What exists

| Language | Support | You write |
|---|---|---|
| **Go** | the plugin SDK, `github.com/monosmith-app/monosmith/packages/plugin-sdk`: typed messages, the serving loop, the host side for tests, the conformance suite, templating and JSON helpers | the plugin's methods |
| **TypeScript** (and JavaScript) | the plugin SDK, [`@monosmith/plugin-sdk`](typescript-sdk.md) on npm: the same, with the types generated from the protocol's JSON Schema. Examples: [`structure-layout-node`](../../examples/plugins/structure-layout-node) (a structure plugin) and [`language-widget-ts`](../../examples/plugins/language-widget-ts) (a language plugin) | the plugin's methods |
| **Node.js** without the SDK | working example that frames the messages by hand: [`node`](../../examples/plugins/node) (a language plugin) | the framing loop and the dispatch, about 40 lines, plus the methods |
| **Python** | no SDK yet. [`python`](../../examples/plugins/python) and [`structure-layout-python`](../../examples/plugins/structure-layout-python), standard library only | the same |
| anything else (C# next) | no SDK, no example. The protocol is the contract: [`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json) is its JSON Schema, generated from the Go types, with a methods table, the error codes and the roles | everything. Not tried here |

The SDKs for the other languages are tracked in [#564](https://github.com/monosmith-app/monosmith/issues/564):
TypeScript first, then Python, then C#. All of them are generated from the same JSON Schema, so a
protocol change reaches every language. Do not look for a `pip` package: it does not exist yet.

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
method. An error a method returns becomes a `-32000` response with your message. `serve.Main` also
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

## Without an SDK: Node.js and Python, by hand

There are three things to write, and they are the same in every language. (In TypeScript the SDK
writes them for you; this is what it does, and what a language without an SDK has to do.)

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
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

The same six lines for `python\plugin.cmd` and for the Go example. See
[Testing a plugin](testing-a-plugin.md#checking-an-interpreted-plugin) for how `smith plugin check`
decides whether its argument is a path or a name.

## The same plugin three ways

`widget-layout` (a [structure plugin](structure-plugin.md)):

| | Go | TypeScript | Python |
|---|---|---|---|
| source | `plugin.go`, `check.go`, `main.go` (146 lines) | `plugin.ts`, `layout.ts` (102 lines) | `plugin.py` (153 lines) |
| framing and dispatch | `serve.Main` | `serve` | written in the file |
| capabilities | derived from the methods | derived from the members | written in the `initialize` answer |
| tests | `go test` with `conformance.Check` and `host.Connect` | `node --test` with `checkPlugin` and `connectInProcess` | `python -m unittest`, speaking the protocol through a child process |
| launcher on Windows | the `.exe` | `plugin.cmd` | `plugin.cmd` |
| dependencies | the SDK | `@monosmith/plugin-sdk` | none |

All three give the same seven lines from `smith plugin check`, the same tasks and the same
`smith doctor` finding, and the Go and the TypeScript plugin answer the same requests with the same
bytes. The Node.js plugin that frames the messages by hand is the smaller
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
