# smith tools install

`smith tools install` offers the ways to get each tool that is missing or too old, and carries
out the one you pick. It never installs in silence: every answer is a presented choice
([ADR 0072](../architecture/decisions/0072-smith-presents-every-decision-that-changes-something.md),
[ADR 0077](../architecture/decisions/0077-smith-offers-to-install-the-tools-a-stack-needs.md)).

For each tool smith shows what it is for and the options, with the marked default:

```text
Python is not installed.
  Why smith needs it: the python projects install, build and test with Python and pip.
  Version wanted: 3.11 or newer (from pyproject.toml).

How do you want to get Python?
  1) Install with winget (default)  runs winget install --id Python.Python.3.13 -e; Windows may ask you to approve it as administrator
  2) Use the official page          go to https://www.python.org/downloads/ and install it there
  3) I will install it myself       smith shows the ways and goes on; run smith tools to check later
  4) Skip it                        go on without it; smith install, build and test fail for the python projects
```

- **Install with a package manager.** The managers smith knows are winget (Windows), brew (macOS and
  Linux), apt, dnf and pacman (Linux), and the languages' own installers when they are already on
  the machine: uv (Python), fnm (Node.js), rustup (Rust), npm (global tools). smith prints the exact
  command first, runs it with your terminal attached so the manager's own questions work, streams
  its output, and detects the tool again. When the tool is installed but not found yet, it says
  to open a new terminal, because the PATH of a running terminal does not update.
- **Use the official page.** smith prints the address.
- **I will install it myself.** smith prints every command that would work here and the page, and
  goes on.
- **Skip it.** smith says what then stops working. The project stays as it is.

## A tool that winget put off PATH

On Windows, winget unpacks some packages (a zip, such as the Dart SDK) into
`%LOCALAPPDATA%\Microsoft\WinGet\Packages\<PackageId>_<Source>\` and creates no link on PATH. The
program is then installed, but no terminal finds it. smith looks there for each tool that has a winget
way in the catalogue: it searches the package folders of that package id, at most three folders deep,
for `<program>.exe`. `smith tools` then says `not on PATH, found in <folder>`, and smith offers this
choice instead of a new install:

```text
the Dart SDK is installed in C:\Users\me\AppData\Local\Microsoft\WinGet\Packages\Google.DartSDK_Microsoft.Winget.Source_8wekyb3d8bbwe\dart-sdk\bin, but that folder is not on PATH.
  1) Let smith use it from that folder (recommended) (default)
  2) Add the folder to your user PATH
  3) Show me the folder and change nothing
  4) Skip it
```

- **Let smith use it from that folder.** smith records the folder in `tool-folders.json` in its
  per-user folder (`SMITH_HOME`, else `~/.smith`). Every smith command then adds the folder at the
  end of its own PATH before it starts the plugins, so the tool checks, the plugins and the tasks find
  the program. Your user and machine PATH do not change, and other programs do not see the tool. The
  same answer is `smith tools use <tool>`.
- **Add the folder to your user PATH.** smith asks you to confirm, then runs a PowerShell command
  that adds the folder to the `Path` of your Windows user (not the machine), keeps the entries that
  use `%VARIABLE%`, and tells Windows that the environment changed. Open a new terminal: the
  terminals that are open do not see the change.
- **Show me the folder and change nothing.** smith prints the folder and the PowerShell command that
  adds it to your PATH.

winget exits with a code that is not 0 when the package is already installed or has no newer
version (`0x8A15002B`, `0x8A150061`, `0x8A15010D`, `0x8A15010E`, `0x8A15004F`). smith treats these
codes as success: it says that the tool is already installed, looks for the folder, and offers the
choice above. `smith doctor --fix` does the same for a winget step.

## Administrator rights

A command that needs administrator rights (`sudo` on Linux, and winget, which may open the Windows
approval window) is never run in silence. smith shows the command and asks: run it yourself
(the default), let smith run it (the system asks for its own permission or password), or skip.
smith never downloads a script and runs it; the only commands are the ones in the plugins'
catalogue, which carries the source page and the check date of each entry.

## Without a terminal

smith refuses and names the flags. `--yes` takes the marked default of each choice, and never runs a
command that needs administrator rights (it prints the command instead). `--manager <name>`
picks that package manager for every tool. `--with-admin` names the choice to let smith run
commands that need administrator rights; the system still asks for its own permission. In a
script, `--yes`, `--manager` and `--with-admin` also add the flags that stop a manager from asking
questions (winget's agreements, `-y` for apt and dnf, `--noconfirm` for pacman). `--dry-run` prints
the plan and installs nothing.

## Synopsis

```sh
smith tools install [tool...] [--manager <name>] [--yes] [--with-admin] [--dry-run] [--json]
```

## Arguments and flags

<!-- smith:begin tools install -->
Install the programs the workspace needs, asking how for each

Usage: `smith tools install [tool...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `tool...` | no |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | print the plan and install nothing |
| `--json` | boolean | off | with --dry-run, print the plan as JSON |
| `--manager <manager>` | string |  | install every tool with this package manager (winget, brew, apt, dnf, pacman, uv, fnm, rustup, npm), or smith-path or user-path for a tool found off PATH |
| `--with-admin` | boolean | off | let smith run commands that need administrator rights (the system still asks for its own permission) |
| `-y, --yes` | boolean | off | take the marked default of each choice (never for commands that need administrator rights) |

<!-- smith:end tools install -->

## Exit codes

0 when every tool the workspace needs is installed afterwards, or only optional ones are missing;
1 when a tool the workspace needs is still missing, when there is no terminal and no flag answers,
and outside a workspace.

## Files and environment

Changes the machine, not the workspace: it runs the package manager you chose. It writes no file of
the workspace. The answer "Let smith use it from that folder" writes `~/.smith/tool-folders.json`
(under `SMITH_HOME` when set). It reads `LOCALAPPDATA` to find winget's package folders. `CI=1`
counts as no terminal.

## smith tools use

`smith tools use <tool>` finds the folder where winget unpacked the tool and records it, so smith uses
the program from there. It is the "Let smith use it from that folder" answer, as a command. It fails
when no package folder holds the program.

<!-- smith:begin tools use -->
Let smith use a tool from the folder winget put it in, without a change to PATH

Usage: `smith tools use <tool>`

| Argument | Required | Meaning |
|---|---|---|
| `tool` | yes |  |

No flags.

<!-- smith:end tools use -->

## Examples

```sh
smith tools install                     # every missing tool, one question each
smith tools install python              # only Python
smith tools install --dry-run           # what smith would offer, nothing installed
smith tools install --manager winget --with-admin go
```
