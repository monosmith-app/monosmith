# The lua plugin

The lua plugin makes LuaRocks projects a smith workspace's projects. Every
directory with a `*.rockspec` is a project: a script or program (an app) or a
module others require (a lib). The plugin lints with luacheck, runs the
busted specs of `spec/`, installs the rock into `dist/` with `luarocks make`,
follows the rocks a rockspec depends on as dependencies, creates new Lua
projects, writes a release version into the rockspec, and tells CI to install
Lua, LuaRocks, luacheck and busted. It never uploads to luarocks.org: a project
is released by its tag.

## At a glance

| | |
|---|---|
| Plugin name | `lua` (built in; runs as a separate process, `smith plugin serve lua`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin) |
| `--lang` value | `lua` |
| Manifest that makes a directory a project | `*.rockspec` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `changeKind`, `channels`, `declared` |
| Not advertised | `remove`, `published`, `dependencies`, `declarations`, `latest` |
| Stacks it publishes | none |
| Project kinds `smith add` can create | `app` and `lib`. `package` is refused (see Releasing) |
| Tasks | `lint`, `test` (when `spec/` exists), `build` |

## What it detects

smith hands the plugin every `*.rockspec` in the files git lists. The plugin
reads the file with regular expressions and never runs it (a rockspec is Lua
code). One directory is one project: the first rockspec in it counts, and one at
the workspace root is skipped.

- **Project name and export.** The `package = "..."` assignment. A rockspec
  without one is not a project.
- **Imports (graph edges).** The names in `dependencies = { ... }`, without their
  version constraints and without `lua`. The host keeps the ones another
  workspace project exports.
- **Kind.** A `-- smithKind: app` comment (or `lib`, `package`); else the top
  directory (`apps/`, `libs/`, `packages/`); else `lib`.
- **Stack.** A `-- smithStack: <name>` comment. It is a comment because LuaRocks
  checks the fields of a rockspec.
- **Version.** `version = "1.2.0-1"` is `1.2.0`: LuaRocks' revision is dropped.

## Tasks

All run in the project's directory and are cached. Inputs: `**/*.lua`,
`*.rockspec`, `.luacheckrc`.

| Task | Command | Outputs | Appears |
|---|---|---|---|
| `lint` | `luacheck .` | none | every project |
| `test` | `busted` | none | when the project has a `spec/` directory |
| `build` | `luarocks make --tree ../../dist/<dir> <rockspec>` | `../../dist/<dir>` | every project |

## Creating a project

```sh
smith add app cli --lang lua
smith add lib strings --lang lua
```

Each writes `<name>-0.0.0-1.rockspec`, `src/<name>.lua` (a `greet` function),
`spec/<name>_spec.lua` (a busted test), and `.luacheckrc` (`std = "max"`, busted
globals in specs). An app also gets `bin/<name>.lua`, which its rockspec
installs. The rockspec's source URL is the placeholder `luarocks write_rockspec`
leaves for you to fill, and its license is `MIT`: change it if that is not yours.
`smith add package x --lang lua` is refused, and so is `smith kind <p> package`.

## Settings

The plugin reads no `smith.toml` setting and no environment variable.

| Where | Key | Effect |
|---|---|---|
| rockspec | `-- smithKind: ...` | `app`, `lib` or `package` |
| rockspec | `-- smithStack: ...` | the project's stack |
| rockspec | `package`, `version`, `dependencies` | the name, the version a release rewrites, the graph edges |
| project directory | `spec/` present | adds the `test` task |

## Releasing

| | |
|---|---|
| Version in the manifest | `version = "<x.y.z>-<revision>"` in the rockspec |
| Version write | the `version` string is replaced in place with the new version and revision `1`, or added under `package` when absent |
| Tag | `<project>@<version>` |
| Publish | nothing is uploaded: "released by its tag". Publishing to luarocks.org is not supported yet, so a project cannot be a `package` |
| Release channel | the plugin declares `channels`: nothing is uploaded, so no registry pointer can move. The rockspec gets `2.0.0-beta.1-1`, and LuaRocks orders the word `beta` below the release |
| Credentials | none |

LuaRocks expects a rockspec file to be named `<package>-<version>.rockspec`.
`smith release` rewrites the version inside the file and does not rename it
(a rename would break the release's restore of the files it touched), so
`smith doctor` warns `<project>'s rockspec is named for its version` and
`smith doctor --fix` renames it.

## CI

`smith ci generate` adds one toolchain step named `lua` on every provider: it
installs `lua5.4` and `luarocks` from the distribution's packages when
`luarocks` is missing (Debian or Ubuntu runners), then luacheck and busted with
`luarocks install`. There is no dev container feature, audit, package store or
credential.

## Doctor, dependencies and removal

**Doctor.** Run when the workspace has Lua projects: `LuaRocks is installed`
(`luarocks --version`), `luacheck is installed` (`luacheck --version`), and the
rockspec file name warning above.

**Dependencies.** Not supported: `smith install`, `smith sync` and `smith up` do
nothing for Lua. Edit `dependencies` in the rockspec.

**Remove.** No remove support; `smith remove` deletes the project's directory.

## Declared projects

A folder with no rockspec of its own, under a rockspec (the nearest one above it), can be declared with `[projects.<name>]`, a `root` and `language = "lua"`. The folder name is the module it exports, and the top-level modules its `.lua` files `require` are its imports, so a project that requires it depends on it. Its tasks are `lint` (`luacheck .`) and, when it has a `spec/` folder, `test` (`busted`). The rockspec above is an input of both. There is no `build`: LuaRocks installs a rock whole, from its rockspec. A declared project is an app or a lib, never a package. A missing folder, a folder with a rockspec of its own and a missing manifest above are refused with a message that names the `[projects.<name>]` table.

## Requirements and limits

- Lua with LuaRocks, luacheck and busted on the machine.
- Discovery is by regular expression: the assignments must start a line, and
  a `dependencies` table is read up to its first `}`.
- No toolchain probe, so no Lua version is part of a task's cache key.
- `luarocks make` fetches a rock's dependencies from the network.

## Verified

Windows 11, smith built from this change. Lua, LuaRocks, luacheck and busted were
**not installed** on that machine, so the commands above were never run against
the real tools.

Run: the plugin's unit tests and the SDK conformance suite with `--generate`
(an app and a lib are generated, discovered back, and their versions written);
`smith new`, `smith add app|lib --lang lua`, `smith projects`, `smith graph`,
`smith explain`, `smith doctor` (it reports the missing tools, with their remedy),
`smith ci generate` (the step above).

Not verified: `luacheck`, `busted` and `luarocks make` on a generated project;
whether LuaRocks accepts the generated rockspec (`rockspec_format = "3.0"`);
the CI install commands on a real runner; luarocks.org.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
