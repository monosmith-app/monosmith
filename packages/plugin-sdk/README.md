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
| `jsondoc` | edit JSON such as `package.json` in place, keeping key order and CRLF line endings |
| `lineendings` | keep the line ending of an edited file; pick one for a new file from `.gitattributes`, git settings and sibling files |
| `projectnaming` | the default name of a new project (package, crate, gem, module) from the workspace namespace, one skip-prefix rule and one sanitizer per syntax (ADR 0098) |
| `toolpin` | the one place a language's toolchain version is decided: `[toolchains]` in smith.toml, then the native pin file, then smith's default |
| `protocolschema` | the JSON Schema of the protocol, built from the types (see above); `schemagen` is its command |
| `projectkind` | a project's kind when its manifest declares none, and the directory each kind lives in |
| `registry` | `registry.GetJSON`: read a JSON document from a package registry with a 10-second timeout and a status check |
| `discovery` | `discovery.WithDeclared`: add the projects `smith.toml` declares to a discover result |
| `toolrun` | run toolchain commands with their output on stderr, keeping stdout for the protocol; `Run` and `Quiet` colour a tool's output when the host hands colour on, `Output` and `Capture` always ask for plain text |
| `termcolour` | smith's one colour decision (`Wanted`, `ToolWanted`), the environment that asks a tool for colour or plain text, the filter that lets only SGR codes through, and smith paint for tools without colour |
| `buildrefs` | find the files that name a build folder, so a doctor fix that moves the build can refuse with the file named |
| `components` | what every plugin's SBOM listing needs: find a lockfile or manifest up to the workspace, and write package URLs |
| `fileedits` | collect a plugin's edits to workspace files in memory, so several rewrites of one file add up; `RetargetDependents` points the other projects' files at a moved directory |
| `functionhost` | the local tool each function host needs (Azure Functions Core Tools, AWS SAM CLI, Functions Framework) and its doctor finding |
| `gitrevision` | read the files of a git revision through one long-lived git process |
| `npmpropagation` | wait until the packages of one npm release can be fetched, before a dist-tag moves |
| `pathrefs` | rewrite the references to a project's directory (path dependency, workspace member, `go.work` entry) when it moves |
| `releaseassets` | what every plugin does the same way for release archives: the file name pattern, the zip, the checksum lines |
| `taskoutput` | tell whether a build task's command names an output folder inside the workspace's `dist/` |
| `tomltext` | edit single keys of a TOML file as text, keeping its comments, order and spacing |
| `toolcatalog` | the data of the programs the language plugins need (git, Python, the .NET SDK, Docker) and how to get each |
| `tooldefaults` | find whether the repository brought its own linter configuration, and write the plugin's embedded default under `.smith/defaults/` |
| `xmltext` | edit single elements of an XML file (a `.csproj`, a `pom.xml`) as text, keeping its comments, order and spacing |

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

## Written exceptions

| Path | Rule waived | Constraint | Owner | Temporary |
|---|---|---|---|---|
| `protocol` (30 files) | the 12-file review trigger for a slice | The package is the wire contract: one outcome, the JSON-RPC messages. `protocolschema` reflects over the package, and the SDKs of other languages generate their types from it, so splitting it would change the Go import path of every plugin and the schema. | the smith maintainers | no |

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
