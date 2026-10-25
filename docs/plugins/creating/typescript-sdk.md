# The TypeScript SDK

`@monosmith/plugin-sdk` is the SDK for writing smith plugins in TypeScript or JavaScript. It does what
the [Go SDK](plugin-sdks.md#go-with-the-sdk) does: it frames the messages, answers `initialize`,
derives the capabilities from the methods your plugin has, gives you a host to test it with and the
conformance checks `smith plugin check` runs, and ports the Go `templating` and `jsondoc` packages.
It has no runtime dependencies. This page walks through a plugin, then lists the whole API.

```sh
npm install @monosmith/plugin-sdk
```

It needs Node.js 22 or later; Node.js 24 runs a `.ts` file as it is (it only removes the types), so a
plugin needs no build step. The package is built from
[`packages/plugin-sdk-ts`](../../../packages/plugin-sdk-ts/README.md) in this repository and released
by smith like the other packages.

## Where the types come from

The Go types in `packages/plugin-sdk/protocol` are the source of truth of the protocol. Their JSON
Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json):
one definition per message, a methods table (each of the 17 methods with its params, its result and the
capability it needs), the error codes, the roles and the protocol version. The TypeScript types are
generated from that file, so they cannot drift from the Go ones: a Go test fails when the schema is
stale, a TypeScript test fails when the generated types are, and a test fails when the schema has a
method the SDK cannot serve. A member is optional in a type exactly when the Go field has
`omitempty`.

The same file is the contract for a language without an SDK. [The protocol cheat
sheet](protocol-cheat-sheet.md) shows it as tables.

## A plugin, step by step

The structure plugin `widget-layout` of the [structure plugin guide](structure-plugin.md), in
TypeScript: it adds a `layout-check` task to widget projects, makes their `lint` wait for it, and warns
about a project without a README. The finished program is
[`structure-layout-node`](../../examples/plugins/structure-layout-node).

### 1. Describe it

A plugin is an object. `describe` is the one method every plugin has; it answers `initialize` with the
plugin's identity. The SDK adds the protocol version and the capabilities.

```ts
import { serve } from '@monosmith/plugin-sdk'
import type { SmithPlugin } from '@monosmith/plugin-sdk'

let workspaceRoot = ''

const base: SmithPlugin = {
  describe: (params) => {
    workspaceRoot = params.workspaceRoot // absolute; the paths in later requests are relative to it

    return {
      plugin: { name: 'widget-layout', version: '0.1.0' },
      kind: 'convention',
      roles: ['structure'],
      languages: ['widget'], // a convention that names no language applies to no project
    }
  },
}

await serve(base)
```

`serve` runs the protocol on standard input and output and ends the process when smith sends
`shutdown`. A `console.log` anywhere in the program, or in a library, cannot corrupt the stream: from the
moment `serve` starts, whatever is written to `process.stdout` goes to standard error, where smith shows
it to the user. Log with `console.error` all the same.

### 2. Add the methods the plugin has

Each method the plugin object has is a capability the SDK advertises, and each it lacks is answered with
`-32601`, `method not supported`, as in Go. Methods may be `async`.

```ts
import { existsSync } from 'node:fs'
import { join } from 'node:path'
import { contributeTo, serve, task } from '@monosmith/plugin-sdk'

const plugin: SmithPlugin = {
  ...base,

  contribute: ({ projects }) => contributeTo(projects, () => ({
    tasks: { 'layout-check': task(['node', checker, 'layout-check'], { cache: true, description: 'fail on folders named for nothing' }) },
    dependsOn: { lint: ['layout-check'] },
  })),

  doctor: ({ projects }) => ({
    findings: projects.map(project => ({
      check: `${project.name} has a README.md`,
      ok: existsSync(join(workspaceRoot, ...project.root.split('/'), 'README.md')),
    })),
  }),
}

await serve(plugin)
```

The params and results are typed from the schema, so `project.root`, `task`'s options and the shape
of a finding are checked as you type. The handshake of this plugin advertises
`{ "doctor": true, "contribute": true }` and nothing else, because nothing else is implemented.

What happens to what you return and throw:

| You | The host sees |
|---|---|
| return a result | the result, written with the members in the order of the schema and every empty optional member left out, exactly as the Go SDK writes it |
| return a result with a required member missing, or of the wrong type | `-32603` (internal error) and a message that names the member: `doctor: result.findings[0].check is required` |
| throw an `Error` | `-32000` (plugin failure) and its message, which smith shows the user: put the remedy in it |
| are sent params of the wrong JSON type | `-32602` (invalid params) before your method is called |

Params reach your method shaped as the types promise: where the Go host writes a nil slice as `null`,
you get `[]`, a required member that is absent is its zero value, and members the SDK does not know are
kept.

### Stop work when the host cancels it

A method that can run for long, `publish` or `doctor` over a big tree, should stop when smith gives up
on it (Ctrl+C, a timeout). Say the plugin is `cancellable`: the SDK then passes every method an
`AbortSignal` as its second argument, aborts it when the host sends the `cancel` notification for that
request, and advertises the `cancel` capability. A method that stops throws (the host sees the
request-cancelled error `-32800`, whatever the message) or returns the result it already has.

```ts
const plugin: SmithPlugin = {
  cancellable: true,
  describe:    () => ({ plugin: { name: 'widget-scan', version: '0.1.0' }, kind: 'convention', roles: ['structure'], languages: ['widget'], concurrency: 4 }),
  doctor:      async ({ projects }, signal) => {
    for (const project of projects) {
      signal.throwIfAborted() // check between units of work, and pass the signal to what you call (fetch, timers.promises, child processes)
      // ... scan the project
    }

    return { findings: [] }
  },
}
```

`concurrency` in what `describe` returns says how many requests the plugin can work on at the same time.
Leave it out for one at a time, in order, which is what a plugin that says nothing gets; the SDK then runs
the requests of a number above 1 together, in the order they arrive, and writes the answers in the order
they finish. `-1` leaves the number to the host. A plugin without `cancellable` is never sent a cancel and
its signal never aborts. [`structure-scan-node`](../../examples/plugins/structure-scan-node) is the whole
example, and the Go twin [`structure-scan-go`](../../examples/plugins/structure-scan-go) answers the same
requests with the same bytes.

### 3. Test it

`connectInProcess` serves the plugin object in your test process over pipes, with the real framing,
handshake and dispatch, and gives you the host side. `checkPlugin` runs the checks of `smith plugin
check`.

```ts
import { checkPlugin, connectInProcess } from '@monosmith/plugin-sdk'

const client = await connectInProcess(createLayoutPlugin(), { workspaceRoot: workspace })

const report = await checkPlugin(client)
assert.ok(report.ok, report.failures())

const result = await client.call('contribute', { projects: [{ name: 'core', root: 'libs/core', language: 'widget', kind: 'lib' }] })
assert.deepEqual(result.projects.core?.dependsOn?.lint, ['layout-check'])

await client.close()
```

The example's tests, run with `node --test` (Node.js 24):

```text
✔ the plugin conforms (5.9366ms)
✔ the handshake advertises exactly what is implemented (1.8395ms)
✔ lint waits for the layout check (1.5453ms)
✔ an unknown method is -32601 (1.5749ms)
✔ doctor warns about a missing README (2.0646ms)
✔ findJunkDrawers skips build output (2.474ms)
```

`client.call` is typed by method name: `client.call('discover', { manifests: [] })` returns a
`DiscoverResult`. An answer that is an error rejects with an `RpcError` that has the `code`. A call that
is not answered within its timeout (two minutes unless `{ timeoutMs }` says otherwise) kills the
plugin process, because its stream can no longer be trusted, and rejects.

To test the real process, with the launcher smith would use, start it instead:

```ts
const client = await startPlugin({ command: [process.execPath, 'plugin.ts'], workspaceRoot: workspace })
```

### 4. Run it under smith

```text
PS> smith plugin check .\plugin.cmd
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

`plugin.cmd` is the two-line launcher of [Starting an interpreted
plugin](plugin-sdks.md#starting-an-interpreted-plugin) (`node "%~dp0plugin.ts" %*`). On Linux and macOS
the script starts with `#!/usr/bin/env node` and is made executable: `smith plugin check ./plugin.ts`
(not run there). To check a script without a launcher, give `smith plugin check` the absolute path of
`node` and then the script, as the SDK's own integration tests do. Declare it in `smith.toml`:

```toml
[plugins.external.widget-layout]
command = ["node", "/path/to/plugin.ts"]

[conventions.widget-layout]
```

### 5. A language plugin

[`language-widget-ts`](../../examples/plugins/language-widget-ts) is the language plugin of the [language
plugin guide](language-plugin.md): `discover`, `generate`, `writeVersion` and `doctor`. It uses the
helpers below: `buildTask` for a cached `build` task that writes into `dist/`, `scaffoldProject` to
write a new project from templates, and `JsonDocument` to change a `widget.json` without reordering it.
Its conformance run, with `smith plugin check .\plugin.cmd --generate`, has the same 27 checks as the
Go plugin's: it generates an app, a lib and a package and discovers them again.

For the same requests a TypeScript plugin and a Go plugin with the same logic give **byte-identical
answers**: the SDK writes the members in the order of the Go structs, leaves out what `omitempty`
leaves out, sorts the keys of maps and escapes `<`, `>` and `&` as Go does. The SDK's integration tests
check it for `language-widget-ts` against `language-widget` and `structure-layout-node` against
`structure-layout-go`, which is how a protocol change shows up in every language.

## The plugin interfaces

`serve` takes a `SmithPlugin`: a `Plugin` (`describe`) with any of these, each `(params) => result |
Promise<result>`. The names follow the Go interfaces.

| Member | Protocol method | Capability | Go interface |
|---|---|---|---|
| `describe` | `initialize` | always | `Plugin.Describe` |
| `discover` | `discover` | `discover` | `Discoverer` |
| `generate` | `generate` | `generate` | `Generator` |
| `writeVersion` | `version/write` | `version` | `Versioner` |
| `publish` | `publish` | `publish` | `Publisher` |
| `published` | `published` | `published` | `RegistryReader` |
| `ci` | `ci` | `ci` | `CIContributor` |
| `doctor` | `doctor` | `doctor` | `DoctorContributor` |
| `remove` | `remove` | `remove` | `Remover` |
| `contribute` | `contribute` | `contribute` | `Contributor` |
| `install` and `uninstall` | `dependencies/install`, `dependencies/uninstall` | `dependencies` | `DependencyManager` |
| `declared` and `rewrite` | `dependencies/declared`, `dependencies/rewrite` | `declarations` | `DeclarationEditor` |
| `latest` | `dependencies/latest` | `latest` | `LatestReader` |

A pair is advertised only when the plugin has both members, as in Go, where the interface needs both
methods; with one of them the SDK warns on standard error and does not call it. `initialize` and
`shutdown` are answered by the SDK. A request before `initialize` is refused with `-32600`, a
notification gets no answer, and an unreadable frame body is answered with `-32700`.

## Reference

Everything is exported from `@monosmith/plugin-sdk`. The messages (`Project`, `Task`, `DiscoverParams`,
`Finding`, and the rest) are exported as types with the descriptions of the Go doc comments.

### Serving

| Name | What it is |
|---|---|
| `serve(plugin)` | runs the plugin on standard input and output and exits; the whole body of an entry point |
| `serveStreams(plugin, { input, output, log? })` | the same loop on streams you give it; resolves at `shutdown` or the end of the input |
| `capabilitiesOf(plugin)` | the `Capabilities` the plugin would advertise |
| `implementsMethod(plugin, method)` | whether the plugin has the members a protocol method needs |
| `Plugin`, `SmithPlugin`, `Discoverer`, `Generator`, `Versioner`, `Publisher`, `RegistryReader`, `CIContributor`, `DoctorContributor`, `Remover`, `Contributor`, `DependencyManager`, `DeclarationEditor`, `LatestReader`, `Cancellable`, `PluginDescription`, `Awaitable` | the types of the plugin object |

### Testing a plugin

| Name | What it is |
|---|---|
| `connectInProcess(plugin, options)` | serves a plugin object in this process over pipes and connects to it |
| `startPlugin({ command, workspaceRoot, env?, cwd?, stderr?, ... })` | starts a plugin process with `SMITH_PLUGIN=1`, performs the handshake, forwards its standard error |
| `connectPlugin({ input, output }, options)` | the handshake over streams you have |
| `PluginClient` | `info` (the `initialize` result), `call(method, params, { timeoutMs?, signal? })`, `notify(method, params)`, `supports(method)`, `peakInFlight`, `close()` |
| `CallInterruptedError` | a call given up on because its `signal` aborted or its time ran out: `reason` (`cancelled` or `timeout`), `killed`, and `response` (the plugin's `-32800`, when it answered) |
| `checkPlugin(client, { workspace? })` | the checks of `smith plugin check`; with `workspace` (the directory the plugin was started in) also `--generate` |
| `ConformanceReport` | `plugin`, `results`, `ok`, `failures()`; `JSON.stringify(report)` is the shape of `smith plugin check --json` |
| `RpcError` | an error answer, with its `code` |

`options` are `workspaceRoot` (required, absolute), `host` (name and version, default `smith` `dev`) and
`initializeTimeoutMs` (30 seconds) and `cancelGraceMs` (5 seconds: how long a plugin has to answer after it was asked to
stop). A client never has more calls in flight than the plugin declared in `concurrency`, as smith does: with
the default of one, calls are serialised. Abort the `signal` of a call and the plugin is sent `cancel` if it
advertised it; the call ends with a `CallInterruptedError`, or with the result if the plugin had finished.
A plugin that does not answer within the grace period is killed.
`checkPlugin` is a port of the Go `conformance.Check` with the same check names;
`smith plugin check` adds one more, "shuts down when asked", which is `client.close()` not rejecting. To
run the real command from a test, spawn `smith plugin check` as
`integration/smith-plugin-check.integration.spec.ts` does.

### Helpers for the answers

| Name | What it is |
|---|---|
| `task(command, options?)` | a `Task`: an argv, never a shell string |
| `buildTask(root, command, options?)` | a cached build task: `outputs` and the directory `command(outputDir)` is told to write into are `../../dist/<root>`, and it depends on `^build` |
| `contributeTo(projects, contribution)` | the answer to `contribute` when each project gets the same kind of addition; a project for which `contribution` returns `undefined` is left out |
| `scaffoldProject({ workspaceRoot, params, files, data?, notes? })` | the answer to `generate` for a scaffold: renders `files` (target path to template) with `scaffoldData(params)` (`Dir`, `Name`, `Kind`, `Stack`, `Workspace`, `Scope`, `ModulePrefix`) and `data`, never overwriting |
| `actionStep(uses, inputs?)`, `ciStep(name, command, options?)`, `credential(env, description, options?)` | the pieces of a `ci` answer: a provider-native step with ordered inputs, a command a phase runs (blocking unless told not to), a secret the release needs |
| `projectBuildDir(root)`, `buildDirFromProject(root)`, `inBuildDir(path)`, `BUILD_DIR` | where a project builds: `dist/apps/web`, and `../../dist/apps/web` from inside the project |
| `ROLES`, `isKnownRole`, `isExclusiveRole`, `isAdditiveRole`, `defaultRoles(kind)`, `effectiveRoles(info)` | the roles, as the schema lists them |
| `requiresLanguage(name)`, `conflictsWith(name)`, `isEmptyRef(ref)`, `describeRef(ref)` | plugin references |
| `resolveProjectKind(declared, root, fallback)`, `projectKindDirectory(kind)`, `publishes(project)`, `isManifest(manifests, base)`, `setToolchain(result, probe)`, `stackPlugins(stack)`, `auditEnabled(audit, policy)`, `auditBlocks(audit, policy)` | the small functions of the Go `protocol` and `projectkind` packages |

### Templating

`writeTemplates(root, files, data?)` renders every template of `files` (target path to text) under `root`
with the `[[ ]]` delimiters and returns the paths it wrote. Target paths may contain actions
(`[[.Dir]]/main.ts`). It checks every target before it writes anything: an existing file is a
`FileExistsError` and nothing is written, a missing field is a `TemplateError` (the Go
`missingkey=error`), and a target outside `root` is refused. `renderTemplate(text, data, name?)` renders
one string.

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

`JsonDocument` is the Go `jsondoc.Document`: a JSON object that keeps its key order and the text of
everything you do not touch (`1.50` stays `1.50`), so editing a `package.json` changes only what you
edited. `JsonDocument.parse(text)`, `has`, `keys()`, `size`, `object(key, create?)` (an object you can
edit in place), `getString`, `getValue`, `setString`, `setJson` (a plain object becomes an editable
document), `delete`, and `render()`, which writes two-space indentation and a final newline as npm does.
The output is the same bytes as the Go SDK's for the same input.

### Building blocks

For a tool that speaks the protocol itself: `encodeParams`, `encodeResult`, `decodeParams`,
`decodeResult` (the wire form of a message by method name), `FrameReader` and `frameMessage` (the
`Content-Length` framing), `protocolSchema` (the schema as data), `PROTOCOL_VERSION`, `ERROR_CODES` and
`MAX_MESSAGE_BYTES`.

## What differs from the Go SDK

| | Go | TypeScript |
|---|---|---|
| a plugin | a struct with methods; `serve.Main(&plugin{})` | an object with members; `await serve(plugin)` |
| async | methods block | members may return a promise; requests are answered one at a time, in order, unless `describe` declares a `concurrency` |
| cancel | `serve.Cancellable` (`WithContext(ctx)`) | `cancellable: true`, and an `AbortSignal` as the second argument of every member |
| concurrency | `Concurrency` in what `Describe` returns | `concurrency` in what `describe` returns |
| error codes | an error becomes `-32000` | the same; the SDK also uses `-32602` for params of the wrong type and `-32603` for a result it cannot write |
| a result with a required member missing | writes the zero value | `-32603`, naming the member |
| `null` in params | a nil slice or map | `[]` or `{}` for a required member, removed for an optional one |
| `toolrun` (run a tool with its output on standard error) | `toolrun.Run` | not ported: use `child_process` with `stdio: ['ignore', process.stderr, process.stderr]` |
| `smith plugin check` in a test | `conformance.Check` | `checkPlugin`, or spawn `smith plugin check` |

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `malformed header "..."` in smith | output on standard output outside the SDK, for example a child process with `stdio: 'inherit'`: give it `stdio: ['ignore', process.stderr, process.stderr]` |
| `ERR_UNSUPPORTED_NODE_MODULES_TYPE_STRIPPING` | Node.js does not strip types in a file below `node_modules`: the SDK is published as JavaScript, so this is your own package; build it, or keep the plugin outside `node_modules` |
| `method not supported: x` for a method you wrote | a pair needs both members (`install` and `uninstall`; `declared` and `rewrite`), and the member names are the ones in the table above (`writeVersion`, not `version`) |
| `-32603` with `result.x is required` | a required member of the result is missing: the type says which |
| `plugin stopped answering: EOF` in a test | the plugin process died: its standard error is on yours, unless you passed `stderr: 'ignore'` |
| a test hangs | a plugin that never answers is killed after the call's timeout (two minutes); pass `{ timeoutMs }` to fail sooner |
| the signal never aborts | the plugin is not `cancellable`, or the host did not send `cancel`: only a call made with a `signal` that aborts (or whose time runs out) sends it |

## See also

- [Writing a plugin in Go, Node.js or Python](plugin-sdks.md): what exists for each language, and the protocol by hand
- [The plugin protocol cheat sheet](protocol-cheat-sheet.md) and [the architecture of the protocol](../../architecture/plugin-protocol.md)
- [Testing a plugin](testing-a-plugin.md)
- [`packages/plugin-sdk-ts`](../../../packages/plugin-sdk-ts/README.md), the package
