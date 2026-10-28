# The MonoSmith agent and the MCP server

A smith workspace can carry an agent that knows what smith can do and works through it.
It is two things: a Claude Code agent file, and the registration of `smith mcp` so Claude
Code and VS Code can call smith's tools instead of guessing command names.

## What is installed

| File | For | What it is |
|---|---|---|
| `.claude/agents/smith.md` | Claude Code | The `smith` agent: what smith is, its habits, what it may do alone and what it must ask for, and the list of tools |
| `.mcp.json` | Claude Code | Registers `smith mcp` as a project MCP server |
| `.vscode/mcp.json` | VS Code | Registers `smith mcp` for agent mode (`servers`, with `"type": "stdio"`) |

`smith new` writes all three. In an existing repository, `smith adopt` and `smith convert`
tell you about `smith agent install`, and `smith doctor` warns until it is done. The adopt and
convert wizard in the terminal interface offers it as its last step, before the result: it shows
the command it will run and installs on Enter. In the interface, "Install the MonoSmith agent"
is also in the menus of a workspace and in the command palette (`smith agent install`).

The MCP entry is the same in both files apart from the shape each client wants:

```json
{ "mcpServers": { "smith": { "command": "smith", "args": ["mcp"] } } }
```

```json
{ "servers": { "smith": { "type": "stdio", "command": "smith", "args": ["mcp"] } } }
```

The command is `smith` because every way of installing smith globally puts it on the PATH
(the install script, `npm i -g monosmith`, `dotnet tool`, `uv`, `pipx`, `pip`). When the
smith that writes the file runs from the workspace's own `node_modules` (monosmith as a
devDependency), it is not on the PATH, so the entry is `npx smith mcp` instead. Both forms
count as current for `smith doctor`.

## Install

```sh
smith agent install        # the agent file and the MCP registration
smith mcp install          # the MCP registration alone
```

Both show each file as a diff and ask before writing; `--yes` writes without asking. Without
`--claude` or `--vscode` they write for the clients the workspace or the machine has (a
`.claude` or `.vscode` folder, a `claude` or `code` command) and for both when they find
neither. They are safe to run again: when the files are current nothing changes.

An existing `.mcp.json` or `.vscode/mcp.json` is merged: your other servers, other keys and
their order stay. A file with comments cannot be merged without losing them, so smith says
so and leaves it; add the entry above by hand.

### Edits are yours

The agent file carries a marker with a hash of its text. A file whose hash still matches is
smith's, and `smith upgrade` rewrites it when smith's text changes. Edit it, and the hash no
longer matches: smith never overwrites it. `smith agent install` shows the difference and
stops; `smith doctor` accepts it. The same goes for the `smith` entry of an MCP file. To take
smith's version again, delete the file (or the entry) and run the command.

To keep a workspace without the files, set this in `smith.toml`; `smith doctor` and
`smith upgrade` then leave them alone:

```toml
[agent]
enabled = false
```

## Use it

In Claude Code, run `smith agent` (it runs `claude --agent smith` in the workspace), or start
`claude` and pick the `smith` agent. The first time, Claude Code asks whether to trust the
`smith` server from `.mcp.json`. Without the `claude` command, `smith agent` says how to
install it.

In VS Code, open `.vscode/mcp.json`, choose Start above the `smith` server, then use the
tools in agent mode. The MonoSmith extension offers the same install as a command.

## What it does alone and what it asks

It reads on its own: the projects, `smith info`, the doctor and its list of fixes, the
project graph, what a change affects, how a task resolved, a release preview, an upgrade
preview, and the dry runs of `adopt` and `convert`.

It asks first for anything that writes, installs, tags, publishes or rewrites many files:
`smith new`, adding or removing a project, changing a kind, installing a package, `smith
upgrade`, `smith adopt --init`, `smith convert`, `smith release` and applying doctor fixes.
Running lint, typecheck and tests after a change is routine and needs no question.

The tool list in the agent file comes from the same catalog the MCP server serves, so it
cannot name a tool that does not exist; a test fails when they differ.

## The habits it carries, and why

These came from real sessions, recorded on
[issue 407](https://github.com/monosmith-app/monosmith/issues/407).

- **Start with the projects and the doctor.** A finding names its remedy.
- **Use smith for anything it owns, then read the diff.** A project made by hand misses the
  marker, the CI key and the slice layout. After a smith command, look for versions going
  backwards and a changed `go` line.
- **Compare `smith --version` with the newest release before a command that writes.** An old
  smith regenerates files from old templates and downgrades them.
- **A dependency belongs to a project, never the root.** In a converted Node project, install
  with `smith i`, which uses npm.
- **Run the project's own tasks and `smith slices check` before pushing.** CI runs them.
- **Errors come as the `problem` shape.** `✖ problem`, then one `→ way out` per line. The
  agent reports it and follows a `→` line.
- **Releases come from conventional commits.** `version(<project>)[x.y.z]: message` forces a
  version. It never creates or pushes tags for you; a naming conflict is a rule smith should
  learn.
- **Check which version a checkout holds before blaming a registry** when a publish fails.
- **A branch behind main fails tests main already fixed.**

## See also

[`smith agent`](../manual/agent.md), [`smith agent install`](../manual/agent-install.md),
[`smith mcp install`](../manual/mcp-install.md), [`smith mcp`](../manual/mcp.md),
[Editors](editors.md).
