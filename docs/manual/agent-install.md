<!-- lore-master
parent: agent.md
-->
# smith agent install

`smith agent install` adds the MonoSmith agent to the workspace: `.claude/agents/smith.md`
for Claude Code, and the registration of `smith mcp` in `.mcp.json` (Claude Code) and
`.vscode/mcp.json` (VS Code). It shows each file as a diff and asks before writing.

## When to use it

- In a workspace made before smith had the agent, or made with `--no-agent`.
- After `smith doctor` warns that the agent files are missing.

For the MCP registration alone, use [`smith mcp install`](mcp-install.md).

## Synopsis

```sh
smith agent install [--claude] [--vscode] [--yes]
```

## Arguments and flags

<!-- smith:begin agent install -->
Install the MonoSmith agent and the smith MCP server in this workspace

Usage: `smith agent install [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--claude` | boolean | off | for Claude Code: .mcp.json and .claude/agents/smith.md |
| `--vscode` | boolean | off | for VS Code: .vscode/mcp.json |
| `-y, --yes` | boolean | off | write without asking |

<!-- smith:end agent install -->

Without `--claude` or `--vscode` it writes for the clients the workspace or this machine has
(a `.claude` or `.vscode` folder, a `claude` or `code` command), and for both when it finds
neither.

## Exit codes

0 when the files are written, already current, or you declined at the question. 1 when a file
was edited (nothing is written), when an MCP file has comments smith cannot merge around,
when there is no terminal and no `--yes`, or outside a workspace.

## Files and environment

Writes `.claude/agents/smith.md`, `.mcp.json` and `.vscode/mcp.json`. An existing MCP file is
merged: other servers, other keys and their order stay. The agent file carries a marker with
the hash of its text; a file that no longer matches is yours and is never overwritten.
`[agent] enabled = false` in `smith.toml` says the files are absent on purpose.

## Examples

```text
$ smith agent install
● Add the smith server to .mcp.json
--- a/.mcp.json
+++ b/.mcp.json
...
● Add .vscode/mcp.json
...
Write 2 files? [y/N] y
✓ Wrote .mcp.json
✓ Wrote .vscode/mcp.json
→ Claude Code: run `smith agent`, or start claude and pick the smith agent. It asks to trust the smith server in .mcp.json the first time.
→ VS Code: open .vscode/mcp.json and choose Start on the smith server, then use it in agent mode.
```

## Common errors

```text
✖ .vscode/mcp.json was changed after smith wrote it, so smith left everything as it is
  → keep your version: nothing to do (smith doctor accepts it)
  → take smith's: delete .vscode/mcp.json (or its smith entry in an MCP file) and run the command again
```

```text
✖ nothing was written: there is no terminal to ask on
  → run it again with --yes to write the files above
```

## Related

[The MonoSmith agent and the MCP server](../guides/agent.md), [`smith agent`](agent.md),
[`smith mcp install`](mcp-install.md), [`smith doctor`](doctor.md), [`smith upgrade`](upgrade.md).
