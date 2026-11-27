// Package versionhandoff lets the smith command keep the version a workspace
// pins when smith link points it at the program itself (ADR 0111). The npm,
// PyPI and NuGet launchers choose the version of a run (ADR 0101); a linked
// entry starts the Go program in the run folder with no launcher in front, so
// that program chooses instead, with the same rules: SMITH_VERSION, else the
// [workspace] smith pin of the nearest smith.toml, else the newest version in
// <smith home>/versions. When the choice is not the running version, it hands
// the run to that version's program, and downloads it first when a pinned
// version is missing.
package versionhandoff
