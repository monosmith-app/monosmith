# plugin-sdk

The Go SDK for writing smith plugins: the protocol's messages, its framing,
the plugin runtime, and the host side used to test a plugin.

```sh
go get github.com/monosmith-app/monosmith/packages/plugin-sdk@latest
```

Released as the Go module tag `packages/plugin-sdk/v<version>`. Before 1.0 a
minor version may change the Go API; the wire protocol is versioned on its own
(protocol version 1), and a breaking change to it raises that integer.

## Packages

| Package | What it is for |
|---|---|
| `protocol` | the JSON-RPC 2.0 messages between smith and a plugin: every method's params and results, project kinds, tasks, stacks, CI contributions |
| `serve` | run a plugin: `serve.Main(plugin)` answers the host on stdin/stdout, deriving the capabilities from the interfaces the plugin implements |
| `transport` | Language Server Protocol framing (`Content-Length`) over a byte stream |
| `host` | the host side: start a plugin, handshake, call with timeouts, shut down; `host.Connect` drives a plugin in-process for tests |
| `conformance` | `conformance.Check`: the checks `smith plugin check` runs, for a plugin's own tests |
| `templating` | render a new project's files from `[[ ]]` templates, never overwriting a file |
| `jsondoc` | edit JSON such as `package.json` in place, keeping key order |
| `protocolschema` | the JSON Schema of the protocol, built from the types (see above); `schemagen` is its command |
| `projectkind` | a project's kind when its manifest declares none, and the directory each kind lives in |
| `toolrun` | run toolchain commands with their output on stderr, keeping stdout for the protocol |

## The protocol schema

The Go types in `protocol` are the source of truth of the wire protocol. Their JSON Schema (draft
2020-12) is committed as [`schema/protocol.schema.json`](schema/protocol.schema.json): one definition
per message, a methods table (`x-methods`: each of the 19 methods with its params and result
definitions and the capability a plugin advertises to be asked it; `cancel` is a notification, without a result), the error codes (`x-errorCodes`),
the roles (`x-roles`), the protocol version (`x-protocolVersion`) and the framing (`x-transport`).
The SDKs of other languages generate their types from it (see
[`packages/plugin-sdk-ts`](../plugin-sdk-ts/README.md), [`packages/plugin-sdk-py`](../plugin-sdk-py/README.md) and [`packages/plugin-sdk-dotnet`](../plugin-sdk-dotnet/README.md)), and a plugin in a language without an SDK can
read it as the contract. A member is required when the Go field has no `omitempty`; the field and
type doc comments are the descriptions.

It is written by `protocolschema` (the slice that reflects over the types and reads the package's
source for comments and enumerations) and the `schemagen` command. After any change to `protocol`:

```sh
cd packages/plugin-sdk
go run ./schemagen                  # or: smith run schema -p plugin-sdk
smith run generate -p plugin-sdk-ts # the TypeScript types follow the schema
smith run generate -p plugin-sdk-py # and the Python types
smith run generate -p plugin-sdk-dotnet # and the C#
```

`go run ./schemagen -check` writes nothing and fails when the file is stale. A Go test
(`TestCommittedSchemaIsCurrent`) fails the same way, so a protocol change cannot be merged without its
schema; and `schemagen` refuses to run when the protocol has a method, an error code or a capability
that `protocolschema`'s tables (`method_table_config.go`, `error_code_table_config.go`) do not describe
yet, so a new method has to be given its params, result and capability on purpose.

## A plugin

```go
package main

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

type kiwi struct{}

func (k *kiwi) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "kiwi", Version: "0.1.0"}, Kind: protocol.KindLanguage,
		Languages: []string{"kiwi"}, Manifests: []string{"kiwi.toml"},
	}, nil
}

func main() { serve.Main(&kiwi{}) }
```

`smith add app my-plugin --stack go-smith-plugin` scaffolds a working plugin
with a conformance test. The protocol, the plugin roles and how a workspace
declares a plugin are in [docs/reference/plugins.md](../../docs/reference/plugins.md);
`apps/smith/goldenpath/testdata/kiwiplugin` is a complete language plugin.
