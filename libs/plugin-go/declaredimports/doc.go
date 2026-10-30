// Package declaredimports checks that a Go module's go.mod matches what its
// code imports, the way `go mod tidy` would leave it: every imported package
// comes from a required module, and every direct requirement is imported.
// It uses `go mod tidy -diff` when the Go on the machine has it, and a
// read-only comparison of go.mod with the sources otherwise (and for a module
// that leans on go.work, which tidy ignores). Nothing here writes go.mod or
// go.sum.
package declaredimports
