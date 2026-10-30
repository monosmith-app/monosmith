# MonoSmith for JetBrains IDEs

A first slice of the JetBrains plugin of issue #303, after the VS Code extension.

## What it does

- A **MonoSmith** tool window (right side) lists the projects of the workspace and their tasks.
  The workspace is the nearest folder up from the IDE project that has a `smith.toml`.
- **Double-click a task** to run `smith run <task> -p <project>` in a Run window. The first line
  is the command line, so you see what is about to run.
- The projects come from the documented `smith projects --json` contract, never from human output.
- smith is the **installed** one ([ADR 0025](../../docs/architecture/decisions/0025-ide-plugins-run-the-installed-smith.md)):
  `node_modules/.bin/smith`, then `.smith/bin/smith`, then `smith` on the `PATH`.

## Not built yet

A `smith.path` setting, the protocol range check and install hints, `smith serve` for live results,
affected projects, doctor, the release preview, new project from a recipe, and the language
server (`smith lsp`, not built in the engine either).

## Build and test

```
smith run build test lint -p jetbrains
```

Needs a JDK 21 and Gradle (smith's CI installs both). The first build downloads IntelliJ IDEA
Community 2025.1 (large, cached in `.intellijPlatform/`). `gradle :apps:jetbrains:buildPlugin`
writes the installable zip to `dist/gradle/monosmith-jetbrains/distributions/`.
`gradle :apps:jetbrains:runIde` starts a sandbox IDE to try it.

## Publishing

Nothing publishes this plugin to the JetBrains Marketplace; it needs a Marketplace token and
signing certificate, and deciding to publish is the owner's. The project is a Kotlin `app` of
the workspace, so `smith release` does not release it.

## Layout

Slices under `src/main/kotlin/dev/monosmith/jetbrains/`, named `<Name><Role>` as `smith slices check`
requires for Kotlin: `workspace` (find the root, list the projects, the contract and its JSON
mapper), `smithcommand` (which smith, the command lines), `taskrun` (run a task), `toolwindow`
(the IDE entry point, a handler). The pure parts have unit tests; the IDE parts are compiled
against the platform but not tested in a running IDE.
