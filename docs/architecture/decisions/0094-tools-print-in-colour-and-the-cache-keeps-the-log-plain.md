# ADR 0094: Tools print in colour for a person, and the cache keeps the log plain

- Status: accepted, implemented
- Date: 2026-11-11

## Context

smith runs tools (npm, jest, cargo, pytest, dotnet, go test) with their stdout on a pipe, so
every tool printed plain text, in a terminal and in the TUI alike. Errors and successes were
hard to see. The owner asked for "tool colour + smith paint": tools colour their own output, and
smith colours the result lines of a tool that cannot.

Several readers must not get colour codes: `--json`, `--json-events`, the MCP server, the editor
engine, every parser (failure hint, CI tail, test counts, annotations, job summary, a plugin's
`Output` and `Capture`), and the cache, whose entries can come from another machine.

## Decision

1. One decision. `termcolour.Wanted` (moved from `cipipeline.colourWanted`) decides for smith's
   own text. `termcolour.ToolWanted` adds `SMITH_TOOL_COLOR`, which smith sets for the processes
   whose stdout is a pipe it owns: its plugins and the child smith of the TUI.
2. A run that streams to a person who sees colour sets the colour variables in each task's
   environment (`FORCE_COLOR=1`, `npm_config_color=always`, `CARGO_TERM_COLOR=always`,
   `PY_COLORS=1`, `PYTHON_COLORS=1`, `MYPY_FORCE_COLOR=1`,
   `DOTNET_SYSTEM_CONSOLE_ALLOW_ANSI_COLOR_REDIRECTION=1`) and clears `NO_COLOR`. A plain run
   sets `NO_COLOR=1` and the same variables off. They are not the task's declared env, so they
   never join the cache key. `CLICOLOR_FORCE=1` is not set: Go's terminal libraries (termenv,
   colorprofile, lipgloss) obey it, and smith's own TUI tests failed in CI when it was set.
3. Only SGR codes pass to a colour destination; every other escape sequence is removed. A plain
   destination gets no escape code at all.
4. smith paint wraps a recognised result line in one palette colour when the line has no colour
   yet, and never changes its characters.
5. **The cache stores the log plain.** The local and the remote cache keep the log without escape
   codes. A replay to a plain destination is plain; a replay to a colour destination gets smith
   paint, but not the tool's own colour. An old entry that holds codes is filtered like live
   output.
6. `--json-events` keeps `line` plain and adds `styled` only for a run handed tool colour, which
   the TUI is.

## Options considered

- Store the log as the tool printed it and strip on replay: rejected. An entry from an older
  smith or another machine may hold cursor codes; the parsers that read the log (failure hint)
  would each need to strip; a plain log is smaller and the same on every machine.
- Colour every `--json-events` line: rejected. The stream is read by programs (the editor), and
  plain must stay plain.
- Turn on `FORCE_COLOR` for the whole TUI child: rejected. The child's own text and every
  capture the TUI parses would then carry codes.

## Consequences

- A cache hit shows less colour than the run that stored it.
- A tool that ignores every variable (tsc, golangci-lint, phpunit, Gradle) prints plain, and only
  smith paint helps, for the lines it recognises.
