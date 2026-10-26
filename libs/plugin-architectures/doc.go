// Package architectures holds the code-layout architectures smith can check
// besides vertical slices: hexagonal, clean, DDD, MVC, modular monolith and
// package-by-layer. Each is a table of layers and the dependencies they may
// have; the checker reads the imports of a project's sources, works out which
// layer each file and each import belongs to, and reports the imports the
// architecture forbids. Every one is a convention plugin
// ([conventions.<name>] in smith.toml) that adds an architecture-check task.
package architectures
