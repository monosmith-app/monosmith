# Doctor findings and their fixes

`smith doctor` offers a fix for a finding in three ways. This page says which way each plugin
finding that names a tool takes, and why a few have none.

1. **The tools choice.** A finding named `<tool> is installed`, where the tool is one the plugin
   declares in the tools catalogue (`smith tools`), gets the same choice as `smith tools install`:
   install with a package manager (the exact command is shown), open the official page, install it
   yourself, or skip with the consequence. The terminal, the TUI and the VS Code Health view show it.
   These findings have it: Go, a JDK, the .NET SDK, the Dart SDK, Ruby, PHP, CMake, a C and C++
   compiler, Rust (cargo) and cargo-audit. Node.js and Python get the same choice from the
   tools list, which `smith doctor` adds when one is missing or older than the pin. A test of the
   built-in plugins (`TestAnInstallFindingOfAPluginNamesATool`) keeps each name equal to a
   declared tool.
2. **A structured fix.** A finding whose remedy is a smith command carries that command. A project
   that builds on macOS only (AppleScript, Objective-C) gets `smith config set ci.native_runners
   macos-latest` and `smith ci generate`, each with its plan.
3. **A written reason.** These findings name an official page and have no fix, because they are
   not in the catalogue and smith has no install command that is right on every system:

   | Finding | Why it has no fix |
   |---|---|
   | `Docker is running` | The tool is installed. The daemon is stopped, and starting Docker Desktop or the service is not a smith command. |
   | `Ninja`, `cppcheck`, `Maven`, `Gradle`, `LuaRocks`, `luacheck`, `Composer`, `Bundler`, `CocoaPods` | Helpers a project may skip. The package names differ by system and no catalogue entry is checked yet. The remedy names the official page. |
   | `the Android SDK`, `Flutter`, `Swift`, `Xcode`, `the Azure CLI`, `the Bicep CLI` | Large installers with their own license steps or a system limit (Xcode is macOS only). The remedy names the official page. |
   | `Playwright`, `the tauri CLI` | npm packages of the project: the fix is `smith install`, shown in the remedy, because it changes the project's dependencies. |

A tool of the second group joins the first when it gets a catalogue entry with a checked source and
an install command for each system: add it to `packages/plugin-sdk/toolcatalog`, list it in the
plugin's `Tools`, and add its check to the test above.
