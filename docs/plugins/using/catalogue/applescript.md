# The applescript plugin

The applescript plugin makes directories of AppleScript sources a smith
workspace's projects. AppleScript has no package manager, no manifest and no
linter, so the plugin does three things only: it finds the projects, compiles
each one with `osacompile` into `dist/` (macOS only), and records a version in a
comment of the project's main script. A project is released by its tag.

## At a glance

| | |
|---|---|
| Plugin name | `applescript` (built in; runs as a separate process, `smith plugin serve applescript`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin) |
| `--lang` value | `applescript` |
| Manifest that makes a directory a project | `*.applescript` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `changeKind`, `channels` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app` and `lib`. `package` is refused: nothing publishes AppleScript |
| Tasks | `build` |
| Tags | `native` |

## What it detects

smith hands the plugin every `*.applescript` file in the files git lists. A
directory with one is a project; a script at the workspace root is skipped.
Compiled scripts (`.scpt`) and bundles (`.scptd`) are build outputs and are not
read.

- **Project name and export.** The directory's name (`apps/notes` is `notes`) and the
  directory.
- **Main script.** `main.applescript` if the directory has it, else the first
  `.applescript` file by name. The build compiles it, and its header holds the
  project's metadata. The other scripts of the directory are libraries for it
  and inputs of the build task.
- **Imports (graph edges).** None: AppleScript declares no dependencies smith
  could read, so an AppleScript project is a leaf of the graph.
- **Kind, stack, version.** Comments at the start of a line in the main script:
  `-- smithKind: app` (or `lib`), `-- smithStack: <name>`, `-- smithVersion: 1.2.0`.
  Without a kind, the top directory (`apps/`, `libs/`) decides, else `lib`.
- **Tag.** Every project is tagged `native` (see CI).

## Tasks

One task, cached, inputs `**/*.applescript`:

| Task | Command | Outputs |
|---|---|---|
| `build` | `sh -c 'mkdir -p "$1" && osacompile -o "$1/$2" "$3"' sh ../../dist/<dir> <name>.scpt <main script>` | `../../dist/<dir>` |

`osacompile` does not create the directory it writes into, so `sh` does (the
plugin names `sh` itself: the argv is not passed through a shell otherwise).
There is no `lint` and no `test`: no tool checks or runs AppleScript outside a Mac,
and the plugin invents none.

## Creating a project

```sh
smith add app notes --lang applescript
smith add lib dialogs --lang applescript
```

Each writes one `main.applescript` with the three header comments. An app has an
`on run` handler (`osascript dist/apps/notes/notes.scpt` runs it once built); a
lib has an `on greet(who)` handler. `smith add package x --lang applescript` is
refused, and so is `smith kind <p> package`.

## Settings

The plugin reads no `smith.toml` setting and no environment variable. Its keys
are the three header comments above.

## Releasing

| | |
|---|---|
| Version in the manifest | `-- smithVersion: <x.y.z>` in the main script |
| Version write | the comment is replaced in place, or added as the first line when absent |
| Tag | `<project>@<version>` |
| Publish | nothing is uploaded: "released by its tag" |
| Release channel | the plugin declares `channels`: nothing is uploaded, so no registry pointer can move, and the version is a comment that takes `2.0.0-beta.1` as it is |
| Credentials | none |

## CI

`osacompile` ships with macOS, so the plugin installs nothing. Because it can build
nowhere else, every project is tagged `native`: the generated pipelines leave it
out of the main job and build it in their `native` job, which by default runs on
Linux, Windows and macOS runners. Set

```toml
[ci]
native_runners = ["macos-latest"]
```

in `smith.toml` so that job runs on macOS only (the setting is the workspace's: it
applies to every native project). On GitLab, add the macOS job yourself.

## Doctor, dependencies and removal

**Doctor.** When the workspace has AppleScript projects: the warning `osacompile
is installed`, which fails on any system but macOS and names the remedy above.

**Dependencies.** None: `smith install`, `smith sync` and `smith up` do nothing.

**Remove.** No remove support; `smith remove` deletes the project's directory.

## Declared projects

AppleScript has no project file above a folder of scripts: the scripts themselves are the manifest, so a declared folder would have nothing to belong to. smith refuses a `[projects.<name>]` declaration for the language with "finds a project by its manifest only: give the folder a manifest of its own".

## Requirements and limits

- macOS with `osacompile` (it ships with the system) to build. `sh` is used too.
- Plain-text sources only. A `.scpt`-only directory is not a project.
- A directory is one project and one compiled script: the main one.
- No lint, no test, no dependency graph, no toolchain probe.

## Verified

Windows 11, smith built from this change. The machine is not a Mac: `osacompile`
was **not available**, so the build command was never run.

Run: the plugin's unit tests and the SDK conformance suite with `--generate`
(an app and a lib generated, discovered back, versions written); `smith new`,
`smith add app|lib --lang applescript`, `smith projects`, `smith explain`,
`smith doctor` (it warns about `osacompile`), `smith ci generate` (the `native`
job is present).

Not verified: `osacompile` and `osascript` on a generated project; the `native`
job on a macOS runner; that the generated scripts compile.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Native projects](../../../reference/ci.md#native-projects)
- [Writing a language plugin](../../creating/language-plugin.md)
