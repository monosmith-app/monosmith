# What every language plugin does the same

A user should not have to learn a language's quirks to use smith. The same `smith.toml` settings
mean the same thing in every language, and where a language cannot do something, it says so in the
same words. This page is the contract; `smith plugin check` tests the parts a test can reach.

## The rules

1. **A project is found by its manifest, or declared.** A plugin describes each manifest it owns
   (`discover`). A project that has no manifest of its own, such as a program that is a package of
   the module above it, is declared in `smith.toml` with `[projects.<name>]`, a `root` and a
   `language`. A plugin that can describe such a folder says so with the `declared` capability and
   describes it, or answers with an error that names the folder and says what is wrong. A folder
   that does not exist is refused, never invented. A plugin that cannot describe a declared
   project does not declare the capability: the host refuses the declaration for its language with
   one message ("finds a project by its manifest only: give the folder a manifest of its own"), and
   `discover` ignores the declaration.
2. **An app is never published to a registry.** The core skips the plugin's registry publish for an
   app that asks for no plugin-built files: an app is released by its tag, its release files and
   its deploy plugins. To publish a program to a registry it is a package.
3. **A package of a release channel never takes the place of the stable version.** A plugin declares
   `channels` once it has checked that a prerelease version of its registry is ordered below the
   release (or published under the channel's tag), and that its version file spells the prerelease
   the way the registry does (`2.0.0-beta.1`, `2.0.0b1`, `2.0.0.pre.beta.1`). Without the
   capability a package of the plugin is refused on a channel branch.
4. **The same task names for the same kinds.** An app has `build`, `test` and `lint`, and `start`
   when it can be run; a lib has `build`, `test` and `lint`; a package adds what releasing needs. A
   task a language has no tool for is left out, not stubbed.
5. **The same words for the same failure.** A problem names the file or folder, says what is wrong
   and says the fix.

6. **A project can depend on a library of the same workspace.** `smith install -w <project>
   <library>` links it ([ADR 0071](../../architecture/decisions/0071-smith-install-links-a-library-of-the-workspace.md)).
   A plugin that can write the dependency in its ecosystem's own form has the `link` capability:
   `dependencies/link` and `dependencies/unlink`, with a plan form (`plan: true`) that writes
   nothing and answers the manifest, the line and the real choices of the range, each with its
   meaning. The host has checked, before it calls the plugin, that the library is not the project,
   not an app, of the same language, and makes no cycle. The plugin refuses a project whose
   manifest is not there, naming the file, and never invents one. A plugin without `link` is
   refused by the host with one sentence: "the X plugin does not link an internal library yet: add
   the dependency in <manifest> by hand". The graph shows the link as soon as the manifest has it,
   so a plugin reports the dependencies its manifest declares among the project's `imports`.

   | Language | `link` | What it writes |
   |---|---|---|
   | Node | yes | the package name in `dependencies` (a range the workspace follows, or `*` / `^x.y.z` by choice), then `npm install` |
   | Python | yes | the distribution name in `[project] dependencies`; the library goes into the shared venv first |
   | Go | yes | `require <module> v0.0.0` in `go.mod`, the library in `go.work` (a `replace` without one) |
   | Rust | yes | a path dependency, by `cargo add --path` |
   | .NET | yes | a `ProjectReference`, by `dotnet add reference` |
   | JVM, Dart, Ruby, PHP, Swift, Lua, C++ and the others | no | the user edits the manifest by hand |

7. **A language plugin declares the tools it needs.** `InitializeResult.tools` lists each program a
   project of the plugin needs on the machine (Python, the .NET SDK, Docker). Each entry (`ToolNeed`)
   gives the name and the plain-words title, why smith needs it, what stops working without it
   (`consequence`), how to detect it (`program`, `alternatives`, `versionArgs`), the files where a
   workspace pins the version (`pins`: `.nvmrc`, `go.mod`, `pyproject.toml` `requires-python`,
   `global.json`, `rust-toolchain.toml`), the ways to get it (`install`: a package manager and a
   package, for `winget`, `brew`, `brew-cask`, `apt`, `dnf`, `pacman`, `npm` and the language's own
   `uv`, `fnm`, `rustup`), the official page (`download`), and the page the entries were checked
   against with the date (`source`, `checked`). The plugin writes facts only. The host builds the
   exact command, presents the choice and runs nothing until the person picks
   ([ADR 0076](../../architecture/decisions/0076-smith-offers-to-install-the-tools-a-stack-needs.md)).
   The Go plugins take their entries from `packages/plugin-sdk/toolcatalog`; a plugin in another
   language writes the same fields.

## How to check

`smith plugin check <plugin>` runs the conformance suite. For the rules above it checks that:

- a plugin with `declared` refuses a declared folder that does not exist and invents no project for
  it, and a plugin without it ignores a declaration;
- a plugin with `channels` also has `publish`, and one with `declared` or `link` also has `discover`;
- a plugin with `link` refuses a link in a folder that does not exist, with an error;
- every declared tool has a name, a title, a reason, a consequence, a program to detect, known package
  managers, an official page, a source and a date, and no tool is declared twice.

`smith plugin list` shows, for each language plugin, whether it describes declared projects
(`DECLARED`), publishes channels (`CHANNELS`) and links an internal library (`LINK`).

The built-in plugins declare these through the Go SDK. The TypeScript, Python and C# SDKs carry
the capabilities in their generated types, and have no way yet for a plugin to declare them.

See also [the language plugin guide](language-plugin.md), [the protocol cheat sheet](protocol-cheat-sheet.md) and
[ADR 0053](../../architecture/decisions/0053-every-language-plugin-follows-one-contract.md).
