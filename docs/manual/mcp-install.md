<!-- lore-master
parent: mcp.md
-->
# smith mcp install

`smith mcp install` registers `smith mcp` where the clients read it: `.mcp.json` for Claude
Code and `.vscode/mcp.json` for VS Code. It merges into an existing file and keeps your other
servers, shows the change as a diff and asks before writing.

## When to use it

- To give an assistant smith's tools without the agent file.

[`smith agent install`](agent-install.md) does this and also writes the Claude Code agent.

## Synopsis

```sh
smith mcp install [--claude] [--vscode] [--yes]
```

## Arguments and flags

<!-- smith:begin mcp install -->
Install the smith MCP server for Claude Code and VS Code

Usage: `smith mcp install [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--claude` | boolean | off | for Claude Code: .mcp.json |
| `--vscode` | boolean | off | for VS Code: .vscode/mcp.json |
| `-y, --yes` | boolean | off | write without asking |

<!-- smith:end mcp install -->

## Exit codes, files and errors

The same as [`smith agent install`](agent-install.md), without the agent file. The entry
written is `smith mcp` (`npx smith mcp` when smith runs from the workspace's `node_modules`).
For VS Code the entry also has `"type": "stdio"`, as its format asks.

## Related

[The MonoSmith agent and the MCP server](../guides/agent.md), [`smith mcp`](mcp.md).
