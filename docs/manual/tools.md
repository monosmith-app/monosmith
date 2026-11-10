# smith tools

`smith tools` lists the programs the workspace needs on this machine (Python, the .NET SDK, Go,
a JDK, Docker, git) and says which are installed. [`smith tools install`](tools-install.md) offers
the ways to get what is missing.

A language plugin declares the tools its projects need. smith runs each tool's version command
(in parallel, with a 3 second limit in all), so the list is quick even when a program hangs. When
the workspace pins a version, smith judges the installed one against it:

| Tool | Pin files smith reads |
|---|---|
| Node.js | `.nvmrc`, `.node-version` |
| Go | `go.mod` (`toolchain`, else `go`) |
| Python | `.python-version`, `requires-python` in `pyproject.toml` |
| .NET SDK | `global.json` (`sdk.version`) |
| Rust | `rust-toolchain.toml`, `rust-toolchain` |
| Ruby | `.ruby-version` |

Gradle and Maven are listed only for a project that holds `build.gradle.kts`, `build.gradle` or
`pom.xml`, because its tasks run `gradle` or `mvn`. Windows has no package for Gradle: smith offers
the official page, and you put `gradle` on your PATH yourself.

## When to use it

- After `smith adopt` or on a new machine, to see what the workspace needs before the first build.
- When a task says a program "is not installed or not on PATH".
- `smith add`, `smith new` and `smith install` print the same warnings for the tools they need
  (with the command to run), and [`smith doctor`](doctor.md) reports a missing tool as a warning
  whose `--fix` offers the same choice.

## Synopsis

```sh
smith tools [--json]
smith tools install [tool...] [--manager <name>] [--yes] [--with-admin] [--dry-run]
```

## Arguments and flags

<!-- smith:begin tools -->
List the programs the workspace needs and which are installed

Usage: `smith tools <command> [flags]`

| Subcommand | Meaning |
|---|---|
| `smith tools install` | Install the programs the workspace needs, asking how for each |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the tools and their state as JSON |

<!-- smith:end tools -->

## Exit codes

0; 1 outside a workspace.

`smith add`, `smith new` and `smith convert` offer the same choices for the tools of the new project, right
after they make it (see [`smith add`](add.md#the-tools-the-project-needs)). Go projects declare golangci-lint at
the version CI pins. For such a tool the first and recommended choice installs the pinned version into
`.smith/bin` of the workspace (`GOBIN=.smith/bin go install ...@v<version>`): the lint task and doctor use it
before PATH, and your other copy stays. The plain `go install` and "Keep mine" stay as choices. A tool with
one wanted version is reported when yours is older or newer. `--manager go-managed` picks the first choice
in a script.

In the interactive shell, `tools install` opens a dialog instead of leaving for the terminal. It asks
how to get each missing tool and shows the exact command of the option under the cursor. When the
command needs administrator rights, it asks a second question: run it yourself, let smith run it,
or skip. The output of the install streams in the dialog, and each tool ends with a result. A
command that asks for a password (`sudo`) is handed over to the terminal, with your answer given.
Esc cancels, and nothing is installed.

## Files and environment

Reads the plugins' tool declarations, and the pin files named above at the workspace root and in
the folders of the projects that use the tool. It runs only version commands (`python3 --version`).
It changes nothing. Git is always listed; the GitHub CLI is listed when the workspace has a
`.github` folder. The tools for running a function app locally (Azure Functions Core Tools, the AWS
SAM CLI, and for a Tauri app Rust, the Xcode Command Line Tools, the Microsoft C++ Build Tools,
WebView2 and the WebKitGTK libraries) are listed only when a project's stack has that choice. A tool
that exists on one system only is listed on that system only.

## Examples

```sh
smith tools
```

```text
✓ Git: git 2.41.0
✓ the GitHub CLI: gh 2.94.0
× the Dart SDK: not installed
    If you skip it: smith install, build and test fail for the dart projects.
✓ Node.js: node 24.12.0
✓ Python: python3 3.13.14

To get what is missing: smith tools install
```
