# ADR 0055: A committed launcher runs the pinned smith

- Status: accepted
- Date: 2026-11-01

## Context

Idea E3 of the design review: "old binary reads new config". A laptop with smith 1.x meets a
`smith.toml` written by 2.x and refuses it ([NewerSmithError]). CI already pins the version it
installs (`[ci] smith_version`, ADR for CI installs from registries: npm, pip or dotnet where the
workspace has that toolchain, else the generated `tools/smith/install.sh`). Laptops have no such pin.

## Decision

`smith bootstrap` writes two static files: `tools/smith/smith` (sh) and `tools/smith/smith.cmd`.
The launcher reads `[ci] smith_version` from `smith.toml` when it runs, uses the smith on `PATH` if
it already is that version, else a cached copy in `.smith/bin/<version>`, else installs one with the
generated `install.sh` or `npx monosmith@<version>`. No pin or `latest` means the smith on `PATH`.

It does not copy the install logic: the shell launcher calls `install.sh`, the one installer
smith already generates, and `npx`, the registry install CI already uses. Pinned versions older and
newer than the installed smith both work, because the pinned copy is installed beside it, never
over it. The files are static, so there is no template version to drift.

## Alternatives refused

- A launcher with the version written into it: it drifts from `smith.toml`.
- A launcher that downloads the binary itself: it duplicates `install.sh`.
- `smith ci generate` using the launcher in pipelines: CI already installs the pin with a package
  manager or the script, and a launcher would add a layer for no gain. Revisit if CI needs a
  second pinned tool.

## Cost

Two committed files per repository, which `smith bootstrap --check` keeps current. The Windows
launcher cannot install without `npx`. A pin that is a range is not supported.

## Smallest first slice (done)

The command, both launchers (tested by hand on Git Bash and cmd.exe: pin matches, pin older, pin
newer, cached copy, no pin, the launcher on `PATH`), the manual page. Not done: a `smith doctor`
note when the launcher is missing but a pin is set.

## Sources

- `apps/smith/bootstrap/`, `apps/smith/cipipeline/smith_install_policy.go`
