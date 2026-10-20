// Package toolchainpin lets a workspace pin the runtimes its projects build
// with ([toolchains] in smith.toml): it compares a pin with the version a
// plugin probed, and writes the pins to .tool-versions, which mise and asdf
// read, so a developer's manager installs them. smith does not download
// runtimes itself (docs/design/decisions.md).
package toolchainpin
