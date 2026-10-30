# The swift plugin

The swift plugin makes Swift packages a smith workspace's projects. Every
directory with a `Package.swift` is a project: a command-line program (an app),
a library used inside the workspace (a lib) or a library others depend on from
your repository (a package). The plugin builds and tests with SwiftPM, follows
local `.package(path:)` dependencies between projects, creates new packages, writes
a release version into `Package.swift`, and tells CI to install Swift. A release
is a git tag, which is how SwiftPM publishes: there is no registry to push to.

It covers SwiftPM packages only. Xcode projects (`.xcodeproj`, `.xcworkspace`) and
Apple-platform app targets are not discovered.

## At a glance

| | |
|---|---|
| Plugin name | `swift` (built in; runs as a separate process, `smith plugin serve swift`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin) |
| `--lang` value | `swift` |
| Manifest that makes a directory a project | `Package.swift` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `changeKind` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app`, `lib` and `package` |
| Tasks | `build`, `test` |

## What it detects

smith hands the plugin every `Package.swift` in the files git lists; the one at the
workspace root is skipped. The plugin reads the manifest with regular expressions
and never runs it (it is Swift code).

- **Project name and export.** The directory's name and the directory.
- **Imports (graph edges).** Every `.package(path: "...")` (with or without a
  `name:` first), resolved to a directory. It is an import when it is outside the
  project's own directory and inside the workspace. Remote
  `.package(url:)` dependencies are ignored.
- **Kind.** A `// smithKind: app` comment (or `lib`, `package`); else the top
  directory (`apps/`, `libs/`, `packages/`); else `app` when the manifest declares
  an `.executableTarget(` or `.executable(`, and `lib` when it does not.
- **Stack and version.** `// smithStack: <name>` and `// smithVersion: <x.y.z>`
  comments. SwiftPM has no version field in the manifest: consumers use the git tag.

## Tasks

All run in the project's directory and are cached. Inputs: `Package.swift`,
`Package.resolved`, `**/*.swift`.

| Task | Command | Outputs |
|---|---|---|
| `build` | `swift build --scratch-path ../../dist/<dir>` | `../../dist/<dir>` |
| `test` | `swift test` | none (it works in `.build`, which the scaffold ignores) |

There is no `lint`. The Swift 6 toolchain has `swift format`, but its verdict
depends on a style each team chooses and the plugin does not guess one.

## Creating a project

```sh
smith add app sales-report --lang swift
smith add lib strings --lang swift
smith add package core --lang swift
```

Each writes `Package.swift` (tools version 5.9, the three header comments under it),
a source target, an XCTest target and a `.gitignore` for `.build/` and `.swiftpm/`.
The module is the project name in UpperCamelCase (`SalesReport`). An app's target is an
executable whose `main.swift` prints a greeting, with a test of it. A lib's or package's
is a library with a public `greet` function.

## Settings

The plugin reads no `smith.toml` setting and no environment variable.

| Where | Key | Effect |
|---|---|---|
| `Package.swift` | `// smithKind: ...` | `app`, `lib` or `package` |
| `Package.swift` | `// smithStack: ...` | the project's stack |
| `Package.swift` | `// smithVersion: ...` | the version a release starts from and rewrites |
| `Package.swift` | `.package(path: ...)` | a graph edge |

## Releasing

| | |
|---|---|
| Version in the manifest | `// smithVersion: <x.y.z>` in `Package.swift` |
| Version write | the comment is replaced in place, or added under the `swift-tools-version` line, which must stay first |
| Tag | `<project>@<version>` by default |
| Publish | nothing is uploaded: "released by its tag" |
| Credentials | none |

SwiftPM resolves semantic-version tags (`1.2.3` or `v1.2.3`), not `name@1.2.3`. For a
package that others depend on by URL, and when it is the only package released from
the repository, set `tag_prefix = "v"` under `[projects.<name>]` in `smith.toml`. With
several packages in one repository, SwiftPM cannot tell their tags apart: that is
SwiftPM's limit, not smith's.

## CI

`smith ci generate` adds one toolchain step named `swift`. On GitHub it is
`swift-actions/setup-swift@v2` with `swift-version: '6.0'`. On the other providers the step
only checks that `swift` is on PATH and fails with where to get it: swift.org
publishes tarballs per distribution and patch release, which a generic step cannot
name. There is no dev container feature, audit, package store or credential.

## Doctor, dependencies and removal

**Doctor.** When the workspace has Swift projects: `Swift is installed`
(`swift --version`).

**Dependencies.** Not supported: `smith install`, `smith sync` and `smith up` do
nothing for Swift. Edit `Package.swift`.

**Remove.** No remove support; `smith remove` deletes the project's directory.

**Changing the kind.** `smith kind` edits the `smithKind` comment and, when the folder
moves, the `.package(path:)` entries of its own manifest that leave it and those other
projects make on it.

## Requirements and limits

- The Swift toolchain (SwiftPM and XCTest) on the machine.
- Discovery is by regular expression: a `.package(path:)` written across several lines
  in an unusual way, or a path built from a variable, is not seen.
- No toolchain probe, so no Swift version is part of a task's cache key.
- Swift packages whose tests need an Apple platform (UIKit, SwiftUI previews) need a macOS
  runner; the generated pipeline runs on Linux.

## Verified

Windows 11, smith built from this change. Swift was **not installed** on that machine, so
`swift build` and `swift test` were never run, on generated or existing packages.

Run: the plugin's unit tests and the SDK conformance suite with `--generate` (an app, a lib
and a package generated, discovered back, versions written); `smith add app|lib|package
--lang swift`, `smith projects`, `smith explain`, `smith doctor` (it reports the missing
Swift, with its remedy), `smith ci generate` (the step above).

Not verified: that the generated packages build and their tests pass (including `@testable
import` of an executable target); the CI step on a real runner; `swift-actions/setup-swift`'s
input name.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
