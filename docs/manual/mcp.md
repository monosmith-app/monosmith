# smith mcp

`smith mcp` serves the workspace to an AI assistant over the Model Context Protocol, on
standard input and output. Each tool runs the smith command it stands for, with the same
validation, so an assistant sees what you would. Register `smith mcp` as the server's
command, started in the workspace.

## When to use it

- To let an assistant list projects, run the doctor, preview a release, run tasks, add
  projects and the like through typed tools rather than a shell.

For an editor, use [`smith serve`](serve.md) (JSON-RPC with run events); the
[VS Code extension](vscode.md) does.

## Synopsis

```sh
smith mcp
```

A client configuration, as a `.mcp.json` entry:

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

## Arguments and flags

<!-- smith:begin mcp -->
Serve the workspace to an AI assistant over the Model Context Protocol (stdio)

Usage: `smith mcp <command>`

| Subcommand | Meaning |
|---|---|
| `smith mcp install` | Install the smith MCP server for Claude Code and VS Code |

No flags.

<!-- smith:end mcp -->

The server opens no network port and ends when standard input closes. Logging goes to
standard error.

## Registering it

`smith mcp install` writes the entry for Claude Code (`.mcp.json`) and VS Code
(`.vscode/mcp.json`), merging into an existing file; `smith agent install` also adds the
Claude Code agent that knows these tools. See [the agent guide](../guides/agent.md).

## Tools

Read-only tools (annotated `readOnlyHint`): `smith_projects`, `smith_graph` (input `focus`),
`smith_affected` (`base`), `smith_kinds`, `smith_stacks`, `smith_choices`, `smith_commands`, `smith_info`,
`smith_doctor`, `smith_release_preview`, `smith_updates`, `smith_explain`,
`smith_recipe_check`, `smith_plugin_list`, `smith_plugin_check`, `smith_doctor_fixes`,
`smith_upgrade_preview`, `smith_change_kind_preview` (input `project`, `kind`).

Tools that change the workspace: `smith_run`, `smith_add_project`, `smith_remove_project`,
`smith_install`, `smith_uninstall`, `smith_doctor_fix`, `smith_upgrade`, `smith_recipe_new`,
`smith_sync`, `smith_pipeline`, `smith_change_kind` (input `project`, `kind`; it never overrules a refusal, only a person does, with `smith kind --force`).

Each tool's inputs are in its `inputSchema` (`tools/list`). A value is one argument and a
value that looks like an option is refused. The tool descriptions and the initialize
`instructions` tell the assistant to start with `smith_projects` and `smith_doctor`.

## Exit codes

0 when the client closes the stream (checked: started outside a workspace with standard
input closed, it exited 0).

## Files and environment

Reads and writes whatever the commands behind the tools read and write. Same environment as
the commands.

## Examples

The handshake and the first tools (client lines abbreviated):

```text
> {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05",...}}
< {"jsonrpc":"2.0","id":1,"result":{"capabilities":{"tools":{}},"instructions":"smith runs this monorepo: ... Start with smith_projects and smith_doctor; ...","protocolVersion":"2024-11-05","serverInfo":{"name":"smith","version":"1.62.0"}}}
> {"jsonrpc":"2.0","id":2,"method":"tools/list"}
< {"jsonrpc":"2.0","id":2,"result":{"tools":[{"annotations":{"destructiveHint":false,"openWorldHint":false,"readOnlyHint":true},"description":"List the workspace's projects as JSON: name, directory, kind, language, stack, dependencies, tasks.","inputSchema":{...},"name":"smith_projects"}, ...
```

## Common errors

An assistant gets the command's own error text as the tool result. For editor setup and a
step-by-step registration, see [Editors](../guides/editors.md).

## Related

[`smith serve`](serve.md), [Editors](../guides/editors.md), [`smith commands`](commands.md).
