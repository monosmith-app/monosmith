# The bundle plugin

The bundle plugin makes a directory with no language a smith project: one with an
`app.json` and a `flow.json` and no other manifest. Such an app is assembled from
other projects' builds (a sound or video flow, say) rather than compiled from
sources of its own. It has its own version, read from `app.json`, its own release
tag, and edges in the task graph to the projects it is built from.

It runs nothing by itself. It has no tasks until `smith.toml` gives it some, and it
never publishes to a registry: a release is a git tag.

## At a glance

| | |
|---|---|
| Plugin name | `bundle` (built in; runs as a separate process, `smith plugin serve bundle`) |
| Kind | language |
| `--lang` value | `bundle` (`smith add app <name> --lang bundle`) |
| Manifest that makes a directory a project | `flow.json`, with an `app.json` beside it |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `channels` |
| Project kinds | `app` |
| Tasks | none |

## What it detects

smith hands the plugin every `flow.json` in the files git lists. A directory is a
bundle when it also has an `app.json`, and none of `package.json`, `Cargo.toml`,
`go.mod`, `pyproject.toml`, `pom.xml` or `pubspec.yaml` (then a language plugin owns
it). The workspace root is skipped. An `app.json` that is not valid JSON is an error
naming the file.

- **Project name.** The directory's name (`apps/music-fixer` is `music-fixer`).
- **Version.** `app.json`'s top-level `"version"`. A release writes the new version
  back there, leaving the rest of the file as it was.
- **Stack.** `app.json`'s `"stack"`, if any.
- **Export.** `app.json`'s `"name"`, when it differs from the directory's.
- **Dependencies (graph edges).** `app.json`'s `"dependsOn"`: a list of identifiers
  other projects export (a Rust crate name, an npm package name, a Go module path).
  The host keeps the ones another workspace project exports.

```json
{ "name": "music-fixer", "version": "1.2.0", "dependsOn": ["videoplugins-music-sync"] }
```

## Tasks, tag prefix and release

Declare what the bundle runs in `smith.toml`. `^build` runs the build of every
project it depends on first:

```toml
[projects.music-fixer]
tag_prefix = "music-fixer-v"     # tags music-fixer-v1.2.0; the default is music-fixer@

[projects.music-fixer.tasks.build]
command = ["node", "assemble.mjs"]
depends_on = ["^build"]
outputs = ["dist"]

[projects.music-fixer.release]   # makes the app releasable
platforms = ["linux/amd64"]
```

`smith release` then versions the bundle on its own conventional commits (a change in
a project it depends on counts), writes the version into `app.json` and tags it.
See [releasing](../../../reference/releasing.md).

## Release channels

A release channel stamps a prerelease version. The plugin declares `channels`: nothing is uploaded, so no registry pointer can move, and the `version` string of `app.json` takes `2.0.0-beta.1` as it is.

## Declared projects

A bundle is an `app.json` and the files beside it; no tool builds or checks a folder of a bundle on its own. smith refuses a `[projects.<name>]` declaration for the language with "finds a project by its manifest only: give the folder a manifest of its own".

## Not covered

A bundle is an app only (`smith add lib` is refused), and the plugin adds nothing to
CI or `smith doctor`: install what the bundle's own tasks need in your pipeline.

## See also

- [Plugin catalogue](README.md)
- [Plugins reference](../../../reference/plugins.md)
- [Configuration reference](../../../reference/configuration.md)
