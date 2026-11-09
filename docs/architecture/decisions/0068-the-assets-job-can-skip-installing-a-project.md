# ADR 0068: The assets job can skip installing a project

- Status: accepted
- Date: 2026-11-01

## Context

The generated assets job runs `smith ci setup` on one runner per platform, then `smith ci assets`.
Setup installs the dependencies of every project. A project whose `[release] build` only zips files
needs none of them. A real report (a Python plugin pack on `windows-latest`) shows the cost: pip
could not resolve the project's dependencies on Windows, setup stopped, and no file was built,
although the build command never imports the package.

## Decision

- `[projects.<name>.release]` has `install` (boolean, default true). `install = false` means the
  build command needs none of this project's dependencies. It needs `build`; without it smith.toml
  fails to load with a sentence that says so.
- `smith ci setup` has the flag `--assets`. The generated assets job runs `smith ci setup --assets`.
  The `ci` job and the native job keep `smith ci setup`.
- With `--assets`, smith does not pass the projects with `install = false` to the plugins' CI
  contribution. A plugin that serves no other project adds no setup steps, toolchain included.
- Only the GitHub workflow has an assets job, so no other mapper changes.

## Consequences

- The knob is explicit and per project. smith does not guess from the build command whether it
  needs the dependencies.
- If the build command does need them, `install = false` is wrong and the build fails on a missing
  tool. The default keeps the old behaviour.
- A skipped plugin also skips its toolchain steps. A build command that needs a language runtime
  must keep `install` at its default.
