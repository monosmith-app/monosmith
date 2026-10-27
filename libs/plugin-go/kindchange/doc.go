// Package kindchange says what changing a Go module's kind edits: the
// "// smith:kind" comment in its go.mod, and, because a module's path ends in
// its directory, the module path itself, every import of it, the requires and
// replaces of other modules, and go.work.
package kindchange
