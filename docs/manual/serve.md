# smith serve

`smith serve` serves the engine over JSON-RPC 2.0 on standard input and output, one JSON
message per line, in the workspace it is started in. It opens no network port. It is how the
[VS Code extension](vscode.md) and the [terminal interface](tui.md) talk to smith: a method
for each function, and a run's events arriving as `event` notifications while it goes.

The full protocol, with every event and method, is in
[Serve and the event stream](../reference/serve.md).

## When to use it

- Writing an editor or tool integration that wants run events and cancellation.

For a script that needs one answer, use the command with `--json`. For an AI assistant, use
[`smith mcp`](mcp.md).

## Synopsis

```sh
smith serve
```

## Arguments and flags

<!-- smith:begin serve -->
Serve the engine to an editor or a terminal interface over JSON-RPC (stdio)

Usage: `smith serve`

No flags.

<!-- smith:end serve -->

It ends when standard input closes or the client sends `shutdown`; anything still running
is cancelled and its processes stopped. smith's own logging goes to standard error.

## Methods

`initialize` answers the methods the running smith supports. In 1.62.0:

`add`, `affected`, `cancel`, `config.get`, `config.list`, `config.set`, `config.unset`,
`doctor`, `doctor.fix`, `doctor.fixes`, `explain`, `graph`, `info`, `initialize`, `install`,
`kinds`, `ping`, `plugins`, `projects`, `recipe.check`, `recipe.describe`, `recipe.list`,
`recipe.roles`, `release`, `release.preview`, `remove`, `run`, `shutdown`, `sync`,
`uninstall`, `updates`, `upgrade`, `upgrade.preview`.

Errors use the JSON-RPC codes: `-32700` parse error, `-32601` method not found, `-32602`
invalid params, `-32000` a function that failed (the message says why).

## Exit codes

0 when it ends because the client closed the stream or sent `shutdown`.

## Files and environment

Reads and writes what the methods' commands do, and uses the same environment variables.

## Examples

```text
> {"jsonrpc":"2.0","id":1,"method":"initialize"}
< {"jsonrpc":"2.0","id":1,"result":{"name":"smith","version":"1.62.0","protocolVersion":1,"eventVersion":1,"workspace":".","methods":["add","affected",...]}}
> {"jsonrpc":"2.0","id":2,"method":"ping"}
< {"jsonrpc":"2.0","id":2,"result":{}}
> {"jsonrpc":"2.0","id":4,"method":"nope"}
< {"jsonrpc":"2.0","id":4,"error":{"code":-32601,"message":"method not found: nope"}}
```

## Related

[Serve and the event stream](../reference/serve.md), [`smith run`](run.md) (`--json-events`),
[`smith mcp`](mcp.md), [Editors](../guides/editors.md).
