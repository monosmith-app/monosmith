# ADR 0048: `smith why`, `smith report` and `smith agent file` explain the workspace to people and agents

- Status: accepted
- Date: 2026-10-31

## Context

Smith 2.0 derives files, follows sibling projects and builds projects from choices (epic
[#819](https://github.com/monosmith-app/monosmith/issues/819)). A person or an agent that opens
such a workspace has three questions, and phase 6 ([#825](https://github.com/monosmith-app/monosmith/issues/825))
answers them:

1. Is this file mine? If it is not, what wrote it and how do I change it?
2. What is the state of the whole workspace, in a form I can attach to a pull request?
3. What are the rules, in the file that coding agents read at the start of a session?

Three commands already sit close to these questions. `smith explain` explains the resolved tasks of
a project. `smith why-miss` explains a cache miss. `smith derive --explain <file>` explains a derived
file only. `smith graph --format html` draws the graph only. `smith agent install` writes the
MonoSmith agent and the MCP registration, not the rules of the workspace.

## Options considered

**Extend `smith explain` with a file.** Rejected. `explain` takes a project and a task and prints
fields with their source. A file is a different subject, and a file path and a project name look
alike when a project is named after its folder.

**Extend `smith why-miss` or rename it.** Rejected. It is a released command with a clear meaning,
and the name `why` is short enough for the new one. The two names stay apart on purpose: `why <file>`
is about ownership, `why-miss <project>:<task>` is about the cache. Each help text names the other.

**Only `smith derive --explain`.** Rejected. It cannot say anything about a file smith does not
derive, and that is the larger share of files.

**Extend `smith graph --format html` into the report.** Rejected. The graph command shows one
thing and is cheap. The report opens the doctor, the derive plan and the release plan, which take
longer and need git history. The two share the drawing: `graphview.SVG` is the one function that
makes the inline SVG.

**A new top-level `smith agents` command.** Rejected. `smith agent` is the group for agents. `agent
file` is a subcommand next to `agent install`.

## Decision

### `smith why <file>`

`why` is a command of its own, in the `why` slice. It reads `smith.lock`, the derive plan (when
`[derive] enabled`), the project graph, the stacks of `smith.toml` and the first 2 KiB of the file.
It writes nothing. The answer has a kind:

| Kind | Meaning |
|---|---|
| `derived` | A renderer owns the file, a region or keys. The record gives the renderer, version, mode and owner. The derive plan gives the state. |
| `ejected` | The record names the owner `you`. |
| `generated` | A header or a marker says a smith command wrote it (`smith ci generate`, `smith agent file`, `smith agent install`). |
| `workspace` | `smith.toml`. |
| `project` | The file is inside a project and no record claims it. The owner is you. |
| `unknown` | No project contains it. Not an error: exit 0. |

The answer always says how to change the file. For a derived file: change the input, then
`smith derive`; or `smith derive --eject <file>`.

**A scaffolded file names the stack and its choices, with "if".** The host does not know which files
a choice writes. ADR 0045 keeps a choice's `files` inside the plugin on purpose, so the host stays
language-free. `why` therefore names the stack of the project and its choices and says that a choice
may have scaffolded the file once. A later step can add a scaffold record (path and choice) to
`smith.lock`. It is not done here, because it needs a protocol change and a writer in `smith add`.

`--json` prints the same answer for tools. The command is not in the MCP catalog yet (see
Consequences).

### `smith report`

`report` writes one HTML file, in the `report` slice. It has five sections: workspace map, health
(doctor), derived files (derive plan), stack and choices, and release plan.

- **Self-contained.** One file. No script, no external request, no web font. A test fails if the page
  contains `<script`, `<link`, `<img`, `src=`, `@import`, `url(` or a URL other than the SVG namespace.
- **Accessible.** Every status is a word: Pass, Warning or Fail. The colour only adds. A failed status
  has a solid border, a warning a dashed one, a pass a dotted one. Tables have a caption and header
  cells with a scope. The graph is an SVG with a title and a description, and the same facts are in a
  table. The page has a skip link, headings, a nav with a label, focus outlines and native
  `<details>` elements. The colours follow `prefers-color-scheme`.
- **Deterministic.** The page holds no time, no host name and no machine path. The same workspace
  gives the same page, so a golden file tests it.
- **The release plan is a preview.** It runs `release.MakePlan` on the current commit as if it were
  merged. It asks no registry and does not check the public API (ADR 0044). `smith release --dry-run`
  does the rest. A part that fails shows its error in its own section, and the page still renders.
- **Separation.** `Gather` reads the workspace into `Data`. `Render` turns `Data` into HTML. The
  golden test builds `Data` for a small workspace and needs no plugin.

### `smith agent file`

`agent file` writes `AGENTS.md` and a `CLAUDE.md`, in the `agentdoc` slice. It is a subcommand of
`smith agent`, next to `install`.

- **Sources.** The architectures in force come from the conventions of the projects. The file roles
  and the naming come from the vertical-slices convention and the languages of the projects. The done
  commands come from the tasks the projects have, then `smith doctor`, `smith derive --check` (when
  derived files are on) and `smith ci verify`. The derived and ejected files come from `smith.lock`.
  The choices come from the stack of each project.
- **Region markers.** The generated text sits between `smith:begin agent-instructions` and
  `smith:end agent-instructions` (HTML comments), the same form as the regions of ADR 0046. A hand
  written part outside them survives. A file with no markers keeps its text and gets the region at
  its end. A file with one marker is an error and is not touched. CRLF line ends stay.
- **`CLAUDE.md` is a pointer.** Its region holds `@AGENTS.md`, the import form of Claude Code, so the
  rules live in one file. `--no-claude-md` skips it.
- **Deterministic and checked.** The text depends only on the workspace. `--check` writes nothing and
  fails when a file is missing or out of date, so CI can run it.
- **Not `agent install`.** `install` writes the MonoSmith agent, a Claude Code subagent that drives
  smith, and the MCP registration. `file` writes the rules for any agent. The two do not overlap.

## Consequences

- Three new slices (`why`, `report`, `agentdoc`) and one new export, `graphview.SVG`. No protocol
  change, so the protocol stays at version 1.
- `why` and `derive --explain` overlap for derived files. `derive --explain` stays as it is.
- `smith why`, `smith report` and `smith agent file` have no MCP tool and no `smith serve` method yet.
  The catalog says "no tool yet". An agent that has the shell can run them.
- `smith doctor` does not yet warn when `AGENTS.md` is stale. Teams that want it run
  `smith agent file --check` in `smith ci verify` steps of their own.
- The `why` answer for a project file is honest but weak: it cannot say which choice wrote the file.
  The scaffold record named above closes the gap.
