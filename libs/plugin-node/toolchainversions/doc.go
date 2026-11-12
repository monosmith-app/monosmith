// Package toolchainversions is the one source of the Node toolchain ranges
// that every stack of the Node plugins writes into a workspace: TypeScript,
// @types/node, esbuild, Vitest and Jest. A stack never writes one of these
// ranges itself, so two stacks in one workspace cannot ask for two versions.
package toolchainversions
