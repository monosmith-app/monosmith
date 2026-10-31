# smith agent

`smith agent` starts Claude Code with the MonoSmith agent (`claude --agent smith`) in the
workspace. The agent knows what smith can do and works through [`smith mcp`](mcp.md): it
reads the projects, the doctor and dry runs on its own and asks before anything that writes,
installs, tags or publishes. [`smith agent install`](agent-install.md) puts the agent in a
workspace.

## When to use it

- To ask for help with a workspace in plain words: "why does this project not build",
  "add a Go lib for the parser", "what would the next release do".

Not when you already work in an editor with the agent: use it there.

## Synopsis

```sh
smith agent
smith agent install [--claude] [--vscode] [--yes]
smith agent file [--check] [--no-claude-md]
```

## Arguments and flags

<!-- smith:begin agent -->
Start the MonoSmith agent in Claude Code; agent install adds it to a workspace

Usage: `smith agent <command>`

| Subcommand | Meaning |
|---|---|
| `smith agent file` | Write AGENTS.md (and a CLAUDE.md that points to it) from the workspace's rules |
| `smith agent install` | Install the MonoSmith agent and the smith MCP server in this workspace |

No flags.

<!-- smith:end agent -->

## Exit codes

0 when Claude Code ends normally. 1 when the workspace has no agent file yet, when `claude`
is not on the PATH, or when Claude Code stops with an error.

## Files and environment

Reads `.claude/agents/smith.md` to see that the agent is installed. Runs `claude --agent
smith` with the workspace root as its folder.

## Common errors

```text
✖ this workspace has no MonoSmith agent yet
  → smith agent install
```

```text
✖ Claude Code is not installed (no claude command on the PATH)
  → install Claude Code from https://claude.com/claude-code, then run smith agent again
  → in VS Code, the MonoSmith extension and .vscode/mcp.json give the same tools to its agent mode
```

## Related

[The MonoSmith agent and the MCP server](../guides/agent.md), [`smith agent install`](agent-install.md), [`smith agent file`](agent-file.md),
[`smith mcp`](mcp.md).
