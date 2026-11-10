# The vba plugin

The vba plugin makes directories of VBA modules, exported from Office documents
as text, a smith workspace's projects. VBA has no command-line compiler, linter,
test runner or package manager, so the plugin has no tasks and no registry: it
gives the modules a place in the graph, a kind, a version and a release by tag,
and scaffolds new ones. It is the plugin to use when macros are kept in git as
`.bas`, `.cls` and `.frm` files and you want them versioned and released with
the rest of the workspace.

## At a glance

| | |
|---|---|
| Plugin name | `vba` (built in; runs as a separate process, `smith plugin serve vba`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin) |
| `--lang` value | `vba` |
| Manifests that make a directory a project | `*.bas`, `*.cls`, `*.frm` (only files that start like an exported VBA module) |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `changeKind`, `channels` |
| Not advertised | `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app` and `lib`. `package` is refused: nothing publishes VBA |
| Tasks | none |

## What it detects

smith hands the plugin every `.bas`, `.cls` and `.frm` file in the files git
lists. A directory is a project when one of them is a VBA module, which the
plugin tells by the `Attribute VB_Name` line every exported module starts with.
A LaTeX `.cls` file or a plain `.bas` source has none and is left out. Modules
at the workspace root are skipped. The binary files next to forms (`.frx`) and the
Office documents themselves are not read.

- **Project name and export.** The directory's name and the directory.
- **Main module.** `main.bas` if it is a VBA module, else the first `.bas` by
  name, else the first `.cls`, else the first `.frm`. Its header holds the
  project's metadata.
- **Imports (graph edges).** None: a VBA module refers to the Office host's
  libraries, not to files.
- **Kind, stack, version.** Comments at the start of a line in the main module,
  under its `Attribute` line: `' smithKind: app` (or `lib`), `' smithStack: <name>`,
  `' smithVersion: 1.2.0`. Without a kind, the top directory (`apps/`, `libs/`)
  decides, else `lib`.

## Tasks

None. A VBA project has no `build`, `lint` or `test` task, so `smith test` says
`no selected project has a [test] task` and `smith doctor` warns `<project> can be
verified` for each (the host's own check: CI never runs anything for it). Tools
that import modules into an Office document or run them need Office and a
desktop session; the plugin does not wrap any.

## Creating a project

```sh
smith add app sales-report --lang vba
smith add lib strings --lang vba
```

Each writes one `main.bas`, with CRLF line endings, as the VBA editor exports it:
`Attribute VB_Name` (the project name made into an identifier: `sales_report`),
the three header comments, `Option Explicit`, and a `Main` procedure (an app) or a
`Greet` function (a lib). Import it into the document with File > Import File in
the VBA editor, and export the modules back to keep them in git.
`smith add package x --lang vba` is refused, and so is `smith kind <p> package`.

## Settings

The plugin reads no `smith.toml` setting and no environment variable. Its keys are
the three header comments above.

## Releasing

| | |
|---|---|
| Version in the manifest | `' smithVersion: <x.y.z>` in the main module |
| Version write | the comment is replaced in place, or added under the `Attribute VB_Name` line (as the first line when there is none), in the file's own line ending |
| Tag | `<project>@<version>` |
| Publish | nothing is uploaded: "released by its tag" |
| Release channel | the plugin declares `channels`: nothing is uploaded, so no registry pointer can move, and the version is a comment that takes `2.0.0-beta.1` as it is |
| Credentials | none |

## CI, doctor, dependencies and removal

The plugin adds nothing to CI and has no doctor checks. `smith install`,
`smith sync` and `smith up` do nothing for VBA, and there is no remove support:
`smith remove` deletes the project's directory.

## Declared projects

VBA has no project file above a folder of exported modules, and no command-line tool: the modules are the manifest, so a declared folder would have nothing to belong to. smith refuses a `[projects.<name>]` declaration for the language with "finds a project by its manifest only: give the folder a manifest of its own".

## Requirements and limits

- Nothing on the machine: the plugin only reads and edits text.
- Exports must be text modules. A module exported by a tool that does not write
  `Attribute VB_Name` is not recognised.
- Modules with the same extensions in other languages (Visual Basic 6 shares the
  format) are read as VBA.
- Only the main module is read and versioned; the others are plain files.
- No tasks, no dependency graph, no toolchain probe.

## Verified

Windows 11, smith built from this change. Office was not involved: no VBA was
imported or run.

Run: the plugin's unit tests (discovery with CRLF files, a LaTeX `.cls` left out,
version and kind edits keeping CRLF) and the SDK conformance suite with
`--generate`; `smith new`, `smith add app|lib --lang vba`, `smith projects`,
`smith explain`, `smith doctor`, `smith kind strings app`, `smith test`.

Not verified: that the generated module imports into Office and runs, and
that Office accepts the header comments under the `Attribute` line.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Releasing](../../../reference/releasing.md)
- [Writing a language plugin](../../creating/language-plugin.md)
