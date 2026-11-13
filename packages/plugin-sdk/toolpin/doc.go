// Package toolpin decides the toolchain version of a language in one order:
// the [toolchains] table of smith.toml, then the language's native pin file
// (go.work, rust-toolchain.toml, global.json, .python-version, .nvmrc, and so
// on), then smith's default. Every language plugin, the CI contribution, the
// dev container and doctor ask it, so one fact has one source (ADR 0020).
package toolpin
