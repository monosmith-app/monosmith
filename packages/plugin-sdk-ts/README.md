# @monosmith/plugin-sdk

The TypeScript SDK for writing [smith](https://github.com/monosmith-app/monosmith) plugins: the
protocol's types, the serve loop, a host to test a plugin with, the conformance checks of
`smith plugin check`, `[[ ]]` templating and JSON editing that keeps key order. No runtime
dependencies; Node.js 22 or later (24 to run a plugin written in TypeScript as it is).

```sh
npm install @monosmith/plugin-sdk
```

A plugin is an object with `describe` and one method for each thing it does. The SDK frames the
JSON-RPC messages, answers `initialize`, derives the capabilities from the methods the object has and
keeps standard output for the protocol:

```ts
#!/usr/bin/env node
import { serve, task, contributeTo } from '@monosmith/plugin-sdk'

await serve({
  describe: () => ({ plugin: { name: 'widget-lint', version: '0.1.0' }, kind: 'convention', roles: ['structure'], languages: ['widget'] }),
  contribute: ({ projects }) => contributeTo(projects, () => ({
    tasks: { 'layout-check': task(['node', 'check.js'], { cache: true }) },
    dependsOn: { lint: ['layout-check'] },
  })),
})
```

Test it in-process, with the real handshake and framing, and run the same checks as
`smith plugin check`:

```ts
import { checkPlugin, connectInProcess } from '@monosmith/plugin-sdk'

const client = await connectInProcess(plugin, { workspaceRoot: dir })
const report = await checkPlugin(client)
if (!report.ok) throw new Error(report.failures())
await client.close()
```

The full guide and API reference: [The TypeScript SDK](../../docs/plugins/creating/typescript-sdk.md).
Examples that use it: [`structure-layout-node`](../../docs/examples/plugins/structure-layout-node) and
[`language-widget-ts`](../../docs/examples/plugins/language-widget-ts).

## The protocol schema is the source of truth

The types in `src/protocol` are not written by hand. The Go types of the protocol
(`packages/plugin-sdk/protocol`) are the source of truth; their JSON Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../plugin-sdk/schema/protocol.schema.json), and
`tools/generate-protocol` turns that into `protocol-types.contract.ts` (the interfaces, the methods
table, the error codes) and `protocol-schema.contract.ts` (the schema itself, which the wire codec and the
host's capability check read at run time). After a change to the protocol:

```sh
smith run schema -p plugin-sdk          # the Go side writes the schema
smith run generate -p plugin-sdk-ts     # this side writes the types
```

Two tests fail when either file is stale: the Go test `TestCommittedSchemaIsCurrent` and
`tools/generate-protocol/generate-protocol.use-case.spec.ts` (`smith run generate-check -p plugin-sdk-ts`
checks without writing). A test also fails when the schema has a method this SDK does not know how to
serve (`PLUGIN_MEMBERS`), so a new method has to be given its place in the plugin object.

## What is where

| Folder of `src` | What it holds |
|---|---|
| `protocol` | the generated types, the schema as data, `RpcError` |
| `codec` | the wire form of the messages: key order, `omitempty`, sorted maps and escaping as the Go SDK writes them, so the bytes are equal; params shaped as the types promise |
| `framing` | `Content-Length` framing over a byte stream |
| `serve` | `serve`, `serveStreams` (a bounded pool of workers, `cancel` and the plugin's `concurrency`), the plugin interfaces, capability derivation |
| `concurrency` | what a declared `concurrency` means: one at a time by default, a count, or unlimited |
| `host` | `startPlugin`, `connectPlugin`, `connectInProcess`, `PluginClient`: the host side, for tests; it keeps to the plugin's concurrency and cancels a call with an `AbortSignal` |
| `conformance` | `checkPlugin`: the checks of `smith plugin check` |
| `templating`, `jsondoc` | ports of the Go SDK's `templating` and `jsondoc` |
| `tasks`, `scaffold`, `ci` | helpers to build the answers of contribute, generate and ci |
| `build-dir`, `plugin-roles`, `plugin-refs`, `projects` | the protocol package's small functions: build directories, roles, plugin references, project kinds, manifests |

The source is organised by vertical slices (`smith` lints it); `integration` holds the tests that start
real processes: the examples under `smith plugin check` and in a throwaway workspace, and the
compatibility test that a TypeScript plugin and a Go plugin with the same logic answer with the same
bytes.

## Develop

```sh
smith run lint typecheck test build -p plugin-sdk-ts
```

The integration tests need `smith` and `go` on the path and skip themselves without them. The package
is released by smith from conventional commits, like the other packages of this repository.
