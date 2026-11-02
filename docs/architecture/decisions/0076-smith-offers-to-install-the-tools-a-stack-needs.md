# ADR 0076: smith offers to install the tools a stack needs

- Status: accepted
- Date: 2026-11-02

## Context

A project of a language needs a program on the machine: Python, the .NET SDK, Go, a JDK, Docker.
When it was missing, smith only named the download page (`taskrun.MissingProgramMessage`), skipped
the language in `smith install`, or kept the new files of `smith add` and said which command to run
later. The person had to find the right package, the right command for the system and the right
version. Issue #887 asked for an offer to install. An install changes the machine, so it falls
under [ADR 0072](0072-smith-presents-every-decision-that-changes-something.md): it is a presented
choice, never an action in silence.

## Decision

- **The plugin declares facts.** `InitializeResult.tools` (a `ToolNeed` list) gives, for each
  program, the name, the reason, the consequence of its absence, how to detect it (program,
  version flag, the files that pin the version), the ways to get it per package manager, the
  official page, the source page and the date the entries were checked. The same way a plugin
  declares `choices`: no new method and no new capability. The Go plugins take the entries from
  `packages/plugin-sdk/toolcatalog`. The conformance suite checks the entries are well formed.
- **The core decides how to ask** (slice `apps/smith/toolinstall`). It detects the tools in
  parallel with a 3 second limit, reads the version range from the workspace pin when there is one
  (`.nvmrc`, the `toolchain` or `go` line of `go.mod`, `requires-python`, `global.json`,
  `rust-toolchain.toml`), detects the package managers of the system, and builds a plan.
- **The plan is a presented choice per missing tool.** The options are: install with a package
  manager (the exact command is shown), use the official page (the address), "I will install it
  myself", and skip it (with what then stops working). The marked default is the first install
  option that needs no administrator rights; when every option needs them, the default is to show
  the command and let the person run it.
- **Administrator rights are never taken in silence.** A command that needs them (`sudo` on Linux,
  `winget`, which may ask Windows for approval) is shown in full, and the person picks: run it
  myself, let smith run it (the system asks for the permission itself), or skip. `--yes` never
  picks for these; `--with-admin` names the choice explicitly in a script.
- **Nothing runs before the choice.** smith prints the command first, streams its output, detects
  the tool again and says when a new terminal is needed for the PATH. smith never downloads a
  script and runs it: the only commands are the ones the plugin's entries name.
- **Without a terminal smith refuses and names the flags** (`--yes`, `--manager`, `--dry-run`).
- **The surfaces:** `smith tools` lists the state, `smith tools install [tool...]` runs the flow,
  `smith doctor` reports a missing tool as a warning with the same fix choices (`doctor --fix`),
  `smith add`, `smith new` and `smith install` name the missing tools and the command, and
  `taskrun.MissingProgramMessage` points to `smith tools install`. The TUI and VS Code show the
  same options as a dialog with the command visible.

## Consequences

- A plugin author adds a tool by writing data. A new package manager is one entry in the core
  table.
- The catalogue needs upkeep: each entry carries its source and date so a reader can tell how old
  it is.
- An install can still fail for reasons smith cannot see (no network, no rights). The output of
  the manager is shown, and the page and the manual way stay on offer.
- The Tauri and Electron prerequisites and the compilers of C and C++ are not declared yet.

## Sources

- Issue #887, `apps/smith/toolinstall`, `packages/plugin-sdk/toolcatalog`,
  `docs/manual/tools.md`.
