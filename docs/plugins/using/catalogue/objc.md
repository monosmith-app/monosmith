# The objc plugin

The objc plugin makes CocoaPods libraries a smith workspace's Objective-C
projects. Every directory with a `*.podspec` is a project: a library used inside
the workspace (a lib) or a pod others depend on (a package). The plugin checks
each with `pod lib lint` (a quick static pass and a full Xcode build), follows the
`s.dependency` lines between pods as dependencies, creates new pods, writes a
release version into the podspec, and tells CI to install CocoaPods. A release is
a git tag, which is where a pod's source is found (`s.source`'s `:tag`). It never
pushes to CocoaPods trunk.

It covers CocoaPods libraries only. Xcode projects (`.xcodeproj`), workspaces and app
targets are not discovered, and plain `.m`/`.h` sources without a podspec are not a
project: nothing there says what the project is, or how to build it.

## At a glance

| | |
|---|---|
| Plugin name | `objc` (built in; runs as a separate process, `smith plugin serve objc`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin) |
| `--lang` value | `objc` |
| Manifest that makes a directory a project | `*.podspec` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `changeKind`, `channels` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `lib` and `package`. `app` is refused (a podspec describes a library) |
| Tasks | `lint`, `build` |
| Tags | `native` |

## What it detects

smith hands the plugin every `*.podspec` in the files git lists. The plugin reads the
file with regular expressions and never runs it (a podspec is Ruby). One directory is
one project: the first podspec in it counts; a file with no `name` is not a pod; the
one at the workspace root is skipped.

- **Project name.** The directory's name (`libs/string-kit` is `string-kit`).
- **Export.** The pod's name (`s.name = 'StringKit'`), the name other pods depend on.
- **Imports (graph edges).** The pods in `s.dependency 'Name'` lines (also
  `add_dependency` and subspec dependencies); `Name/Sub` is reduced to `Name`. The
  host keeps the ones another workspace project exports.
- **Kind.** A `# smithKind: lib` comment (or `package`); else the top directory
  (`libs/`, `packages/`); else `lib`.
- **Stack.** A `# smithStack: <name>` comment.
- **Version.** The literal in `s.version = '1.2.0'`. A computed version
  (`File.read`) shows no version.
- **Tag.** Every project is tagged `native` (see CI).

## Tasks

Both run in the project's directory and are cached. Inputs: `*.podspec`, `**/*.h`,
`**/*.m`, `**/*.mm`, `**/*.c`. Neither writes into the workspace: CocoaPods works in a
temporary directory.

| Task | Command |
|---|---|
| `lint` | `pod lib lint --quick <podspec>`: checks the podspec without downloading or building |
| `build` | `pod lib lint <podspec>`: builds the pod with `xcodebuild` and lints it (macOS and Xcode) |

There is no `test` task: a pod's tests need a `test_spec` in the podspec that the
scaffold does not write.

## Creating a project

```sh
smith add lib string-kit --lang objc
smith add package core --lang objc
```

Each writes `<Module>.podspec` (the module is the project name in UpperCamelCase:
`StringKit`), and `<Module>/Classes/<Module>.h` and `.m` with a class method `greet:`. The
podspec has the header comments, version `0.0.0`, MIT as the license, and placeholder
homepage, source and author on `example.invalid`, which you replace before publishing.
`smith add app x --lang objc` is refused.

## Settings

The plugin reads no `smith.toml` setting and no environment variable.

| Where | Key | Effect |
|---|---|---|
| podspec | `# smithKind: ...` | `lib` or `package` |
| podspec | `# smithStack: ...` | the project's stack |
| podspec | `s.name`, `s.version`, `s.dependency` | the export, the version a release rewrites, the graph edges |

## Releasing

| | |
|---|---|
| Version in the manifest | the literal `s.version = '<x.y.z>'` |
| Version write | the string is replaced in place, keeping its quotes. A podspec with a computed version is left alone and the release reports no file |
| Tag | `<project>@<version>`; the podspec's `:tag => s.version.to_s` must name the tag you push, so set `tag_prefix` in `smith.toml` to match |
| Publish | nothing is uploaded: "released by its tag". `pod trunk push` is not run by smith |
| Release channel | the plugin declares `channels`: nothing is pushed to CocoaPods trunk, so no registry pointer can move. The podspec gets `2.0.0-beta.1`, which CocoaPods reads as a prerelease below `2.0.0` |
| Credentials | none |

## CI

`smith ci generate` adds one toolchain step named `cocoapods` on every provider:
`command -v pod >/dev/null || gem install cocoapods --no-document`. Because `pod lib lint` needs
Xcode, every project is tagged `native`: the pipelines leave it out of the main job and
build it in their `native` job, which by default runs on Linux, Windows and macOS
runners. Set

```toml
[ci]
native_runners = ["macos-latest"]
```

so that job runs on macOS only (the setting is the workspace's: it applies to every native
project). On GitLab, add the macOS job yourself.

## Doctor, dependencies and removal

**Doctor.** When the workspace has Objective-C projects, two warnings: `CocoaPods is
installed` (`pod` on PATH) and `Xcode is installed` (`xcodebuild` on PATH), each with its
remedy; both fail on any system but a Mac.

**Dependencies.** Not supported: `smith install`, `smith sync` and `smith up` do nothing.
Edit the podspec.

**Remove.** No remove support; `smith remove` deletes the project's directory.

## Declared projects

A pod is its podspec. No tool of the plugin builds or releases a folder of sources without one. smith refuses a `[projects.<name>]` declaration for the language with "finds a project by its manifest only: give the folder a manifest of its own".

## Requirements and limits

- A Mac with Xcode and CocoaPods for `build` (and `lint`, which needs `pod`).
- A podspec is Ruby: a name or version built by code is not seen.
- A directory with a `Package.swift` beside its podspec is a Swift package: the objc plugin leaves it to
  the swift plugin, so a CocoaPods pod that is also a Swift package is a swift project.
- No toolchain probe, so no CocoaPods or Xcode version is part of a task's cache key.

## Verified

Windows 11, smith built from this change. Ruby is present but CocoaPods and Xcode are **not**:
`pod lib lint` never ran.

Run: the plugin's unit tests and the SDK conformance suite with `--generate` (a lib and a
package generated, discovered back, versions written); `smith add lib|package --lang objc`,
the refusal of `smith add app`, `smith projects`, `smith explain`, `smith doctor` (the two
warnings), `smith ci generate` (the `cocoapods` step and the `native` job).

Not verified: `pod lib lint` (quick or full) on a generated pod, so whether CocoaPods accepts
the generated podspec and the placeholder URLs; `--quick`'s behavior beyond its documentation;
the native job on a macOS runner.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Native projects](../../../reference/ci.md#native-projects)
- [Writing a language plugin](../../creating/language-plugin.md)
