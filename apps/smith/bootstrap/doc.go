// Package bootstrap writes the launcher a repository commits so that every
// laptop and CI agent runs the smith version the workspace pins in
// [ci] smith_version (ADR 0055). The launcher is two static files,
// tools/smith/smith (sh) and tools/smith/smith.cmd; they read the pin from
// smith.toml when they run, so changing the pin needs no regeneration.
package bootstrap
