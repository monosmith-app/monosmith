# ADR 0077: smith offers to install the tools a stack needs

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
- A tool can be limited to a system (`systems`: `windows`, `darwin`, `linux`), so the Xcode Command Line Tools are not asked on Windows. The Tauri plugin declares Rust, the Xcode Command Line Tools, the Microsoft C++ Build Tools, WebView2 and the WebKitGTK libraries, limited to the choice `tauri`. Electron needs only Node. The C and C++ plugin declares a compiler per system. The Docker plugin declares Docker.
- A tool can be tied to files (`files`): it is needed only by a project whose folder holds one of them. The JVM plugin declares `gradle` (`build.gradle.kts`, `build.gradle`) and `mvn` (`pom.xml`), because its tasks run those programs. Gradle has no winget package and Ubuntu's apt package is too old for current builds, so on Windows and Debian the choice is the official page, or "I will install it myself". A language named by flag (`smith add --lang java`) does not bring a tool tied to files, since no project exists yet to hold the file.
- A tool can have one wanted version (`exact`, with the version in `suggested`): an installed version that differs, older or newer, is reported as "is 2.13.0, not the 2.15.0 that CI pins", and the skip option reads "Keep mine". The Go plugin declares `golangci-lint` this way, with the version of its CI template, installed with the `go` manager (`go install <module>@v{version}`). Its doctor check stays in the plugin, which names the file the PATH finds and the folder `go install` writes to.
- `smith add`, `smith new` and `smith convert` check the tools of the new project when they end (`toolinstall.Offer`): in a terminal they present the install choice of each tool that is missing or different; with `--yes` or without a terminal they install nothing and print the tools and the one `smith tools install <tools>` command. The editors read that notice and offer Install tools.
- Every install entry was checked against its package catalogue on the date in `checked` (the Fedora names against the Fedora package data, the .NET names on the Microsoft Learn pages). An entry that cannot be found is removed from the install options; the official page stays. Windows cannot tell whether the C++ workload is inside the Build Tools: smith checks only that the Visual Studio Installer is there.
- In the interactive shell, `tools install` opens a dialog (`apps/smith/tui/toolsdialog`): one question per missing tool with the exact command of each option and the marked default, a second question when the command needs administrator rights, the output of the install as it streams, and a result per tool. It reads the choices with `tools install --dry-run --json` and runs the answer with `--manager` (and `--with-admin` only after the second answer). A `sudo` command asks for a password, so it is handed over to the terminal with the answer given. A command with `--yes`, `--manager` or `--dry-run` keeps the terminal hand-over.
- `smith compose` and the composer screens list every tool of the chosen stack with its state and the line `smith tools install <tool>`, from the same data.
- On Windows, winget unpacks a portable package (the Dart SDK zip) into `%LOCALAPPDATA%\Microsoft\WinGet\Packages\<PackageId>_<Source>` and may put nothing on PATH. When no program on PATH answers, the detector looks there by the tool's winget package id (a version placeholder matches any version), at most three folders deep. A tool found there gets a different choice: let smith use it from that folder (the default), add the folder to the user PATH (after a confirmation; a PowerShell command that keeps `%VARIABLE%` entries), or show the folder and change nothing. A new install is not offered, since it would not help. The first two answers are install options with the managers `smith-path` and `user-path`, so `--manager`, the TUI dialog and the editors run them as they run any manager. "Use it from that folder" records the folder in `tool-folders.json` in smith's per-user folder (`SMITH_HOME`, else `~/.smith`), beside the plugins and trust decisions, because the folder belongs to the machine, not to the workspace. Each session adds the recorded folders at the end of smith's own PATH before it starts the plugins, so the checks, the plugins and the tasks see the same program. `.smith/bin` stays the folder for a pinned copy that must come first; a recorded folder never shadows a program that PATH already finds.
- winget's documented "already installed" exit codes (`0x8A15002B` no applicable upgrade, `0x8A150061`, `0x8A15010D`, `0x8A15010E`, `0x8A15004F`) are a success for an install: `smith tools install` and `smith doctor --fix` say the tool is already installed and go on.

## Sources

- Issue #887, `apps/smith/toolinstall`, `packages/plugin-sdk/toolcatalog`,
  `docs/manual/tools.md`.
