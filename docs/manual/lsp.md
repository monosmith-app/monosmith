# smith lsp

`smith lsp` serves `smith.toml` to an editor over the Language Server Protocol, on standard
input and output (`Content-Length` framing). It reads the schema that `smith schema` prints,
so completion and hover follow the same keys that smith accepts. It opens no network port,
runs no task and changes no file.

## When to use it

- To get diagnostics, completion and hover for `smith.toml` in any editor that can start a
  language server.

For the engine's functions and live run results, use [`smith serve`](serve.md).

## Synopsis

```sh
smith lsp
```

Register it as the language server of `smith.toml`, with `smith` as the command and `lsp` as
its only argument. The server answers for documents named `smith.toml`; it ignores every
other document.

## Arguments and flags

<!-- smith:begin lsp -->
Serve smith.toml to an editor over the Language Server Protocol (stdio)

Usage: `smith lsp`

No flags.

<!-- smith:end lsp -->

## What it does

| LSP feature | Behaviour |
|---|---|
| `initialize` | Reports full-text sync, completion (triggered by `[` and `.`) and hover. |
| `textDocument/didOpen`, `didChange` | Publishes diagnostics: TOML syntax errors, a key smith does not know (underlined where it is), a value of the wrong type, and the values smith itself rejects (placed on the first line when they have no position). |
| `textDocument/didClose` | Clears the diagnostics. |
| `textDocument/completion` | After `[`: table names. On a key line: the keys of the current table. After `=`: the allowed values of an enumerated key, or `true` and `false`. |
| `textDocument/hover` | On a table header or a key: its path, type, description and allowed values. |
| `shutdown`, `exit` | Ends the server. |

Not yet supported: recipe files, diagnostics from `smith doctor` with code actions for safe
fixes, completion and hover that use the installed plugins, and tracking of tables inside
multi-line values. Progress is on issue
[#303](https://github.com/monosmith-app/monosmith/issues/303).

## Exit codes

0 when the client closes the stream or sends `shutdown` and then `exit`. 1 when the client
sends `exit` without `shutdown`, or the stream is malformed.

## Files and environment

Reads nothing from disk: it works on the text the editor sends. Logging goes to standard
error.
