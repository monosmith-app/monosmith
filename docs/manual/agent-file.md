<!-- lore-master
parent: agent.md
-->
# smith agent file

`smith agent file` writes `AGENTS.md`, and a `CLAUDE.md` that points to it, from the rules of
the workspace. Coding agents (Claude Code, Codex, Cursor and others) read these files at the
start of a session, so they follow the rules without being told in each prompt.

The file states:

- where code goes: the architectures in force and, for vertical slices, the one-folder-per-outcome rule
- how files are named: the roles, and the spelling of each language in the workspace
- how to add a project: `smith add ... --with ...`
- which commands mean done: `smith lint`, `smith test`, `smith build`, `smith doctor`,
  `smith derive --check` and `smith ci verify`, for the tasks the workspace has
- which derived files not to edit by hand, from `smith.lock`
- the stack and the choices of each project

## When to use it

- When you start using a coding agent in the workspace.
- After you change the architecture, add a project or derive a new file. Run it again, or let CI
  run `smith agent file --check`.

## Synopsis

```sh
smith agent file [--check] [--no-claude-md]
```

## Arguments and flags

<!-- smith:begin agent file -->
Write AGENTS.md (and a CLAUDE.md that points to it) from the workspace's rules

Usage: `smith agent file [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | write nothing; fail when a file is missing or out of date |
| `--no-claude-md` | boolean | off | do not write CLAUDE.md |

<!-- smith:end agent file -->

## Region markers

The generated text sits between two HTML comment lines. The first holds the words
`smith:begin agent-instructions`, and the last holds `smith:end agent-instructions`. A new
file shows both, with the generated text between them.

Write your own text anywhere else in the file. A new run keeps it. A file with no markers keeps
its text, and the region is added at its end. A file with only one marker is an error: smith
cannot tell which lines it owns, so it writes nothing.

`CLAUDE.md` gets a small region that imports `AGENTS.md` (`@AGENTS.md`).

The text is the same for the same workspace. A file with CRLF line ends keeps them.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The files were written, or `--check` found them current. |
| 1 | `--check` found a file missing or out of date, a file has broken markers, or `smith.lock` cannot be read. |

## Files and environment

Reads `smith.toml`, `smith.lock` and the project graph. Writes `AGENTS.md` and `CLAUDE.md` at the
workspace root.

## Examples

```sh
smith agent file
smith agent file --check
```

## Related

[`smith agent install`](agent-install.md), [`smith why`](why.md),
[The MonoSmith agent and the MCP server](../guides/agent.md).
