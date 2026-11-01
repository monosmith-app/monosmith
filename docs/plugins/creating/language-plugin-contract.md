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

## How to check

`smith plugin check <plugin>` runs the conformance suite. For the rules above it checks that:

- a plugin with `declared` refuses a declared folder that does not exist and invents no project for
  it, and a plugin without it ignores a declaration;
- a plugin with `channels` also has `publish`, and one with `declared` also has `discover`.

`smith plugin list` shows, for each language plugin, whether it describes declared projects
(`DECLARED`) and publishes channels (`CHANNELS`).

The built-in plugins declare these through the Go SDK. The TypeScript, Python and C# SDKs carry
the capabilities in their generated types, and have no way yet for a plugin to declare them.

See also [the language plugin guide](language-plugin.md), [the protocol cheat sheet](protocol-cheat-sheet.md) and
[ADR 0053](../../architecture/decisions/0053-every-language-plugin-follows-one-contract.md).
