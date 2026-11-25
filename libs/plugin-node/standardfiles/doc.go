// Package standardfiles gives smith's standard version of a Node project's
// configuration files (protocol method project/standard), for a converted
// workspace to take: the source condition a library's package.json "exports"
// and the root tsconfig.base.json need for editors and tsc to read the
// library's TypeScript sources, and the project's tsconfig.json and
// tsconfig.build.json where the project adds nothing the standard lacks. It
// writes nothing: the host compares, asks and writes.
package standardfiles
