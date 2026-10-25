# monosmith-plugin-sdk

The Python SDK for writing [smith](https://github.com/monosmith-app/monosmith) plugins: the
protocol's types, the serve loop, a host to test a plugin with, the conformance checks of
`smith plugin check`, `[[ ]]` templating and JSON editing that keeps key order. No runtime
dependencies; Python 3.10 or later; fully type-annotated.

```sh
pip install monosmith-plugin-sdk
```

A plugin is an object with `describe` and one method for each thing it does. The SDK frames the
JSON-RPC messages, answers `initialize`, derives the capabilities from the methods the object has and
keeps standard output for the protocol:

```python
from monosmith_plugin_sdk import contribute_to, serve, task


class WidgetLint:
    def describe(self, params):
        return {
            "plugin": {"name": "widget-lint", "version": "0.1.0"},
            "kind": "convention",
            "roles": ["structure"],
            "languages": ["widget"],
        }

    def contribute(self, params):
        return contribute_to(
            params["projects"],
            lambda project: {
                "tasks": {"layout-check": task(["python", "check.py"], cache=True)},
                "dependsOn": {"lint": ["layout-check"]},
            },
        )


if __name__ == "__main__":
    serve(WidgetLint())
```

Test it in-process, with the real handshake and framing, and run the same checks as
`smith plugin check`:

```python
from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process

with connect_in_process(WidgetLint(), HostOptions(workspace_root=str(tmp_path))) as client:
    report = check_plugin(client)
    assert report.ok, report.failures()
```

The full guide and API reference:
[The Python SDK](https://github.com/monosmith-app/monosmith/blob/main/docs/plugins/creating/python-sdk.md).
Examples that use it:
[`structure-layout-py`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/structure-layout-py),
[`structure-scan-py`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/structure-scan-py) and
[`language-widget-py`](https://github.com/monosmith-app/monosmith/tree/main/docs/examples/plugins/language-widget-py).

## The protocol schema is the source of truth

The types in `src/monosmith_plugin_sdk/protocol` are not written by hand. The Go types of the protocol
(`packages/plugin-sdk/protocol`) are the source of truth; their JSON Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../plugin-sdk/schema/protocol.schema.json), and
`tools/generate_protocol` turns that into `protocol_types_contract.py` (a `TypedDict` per message, the
methods table, the error codes) and `protocol_schema_contract.py` (the schema itself, which the wire codec
and the host's capability check read at run time). After a change to the protocol:

```sh
smith run schema -p plugin-sdk          # the Go side writes the schema
smith run generate -p plugin-sdk-py     # this side writes the types
```

A test (`tests/generate_protocol`) and `smith run generate-check -p plugin-sdk-py` fail when either file
is stale. A test also fails when the schema has a method this SDK does not know how to serve
(`PLUGIN_MEMBERS`), so a new method has to be given its place in the plugin object.

## What is where

| Folder of `src/monosmith_plugin_sdk` | What it holds |
|---|---|
| `protocol` | the generated types, the schema as data, `RpcError` |
| `codec` | the wire form of the messages: key order, `omitempty`, sorted maps and escaping as the Go SDK writes them, so the bytes are equal; params shaped as the types promise |
| `framing` | `Content-Length` framing over a byte stream |
| `cancellation` | `CancelToken`, what a cancellable method watches |
| `serve` | `serve`, `serve_streams` (a bounded pool of worker threads, `cancel` and the plugin's `concurrency`), the plugin protocols, capability derivation, standard input and output claimed for the protocol |
| `concurrency` | what a declared `concurrency` means: one at a time by default, a count, or unlimited |
| `host` | `start_plugin`, `connect_plugin`, `connect_in_process`, `PluginClient`: the host side, for tests; it keeps to the plugin's concurrency and cancels a call with a `CancelToken` |
| `conformance` | `check_plugin`: the checks of `smith plugin check` |
| `templating`, `jsondoc` | ports of the Go SDK's `templating` and `jsondoc` |
| `tasks`, `scaffold`, `ci` | helpers to build the answers of contribute, generate and ci |
| `build_dir`, `plugin_roles`, `plugin_refs`, `projects` | the protocol package's small functions: build directories, roles, plugin references, project kinds, manifests |

The source is organised by vertical slices (`smith slices check` lints it); `tests/integration` holds the
tests that start real processes: the examples under `smith plugin check` and in a throwaway workspace, and
the compatibility test that a Python plugin, a TypeScript plugin and a Go plugin with the same logic answer
with the same bytes.

## Develop

```sh
smith run lint typecheck test build generate-check -p plugin-sdk-py
```

The integration tests need `smith`, `go` and `node` on the path and skip themselves without them (in CI,
where they exist, a missing one fails the test). The package is released by smith from conventional
commits, like the other packages of this repository.
