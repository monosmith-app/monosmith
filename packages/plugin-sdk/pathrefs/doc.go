// Package pathrefs rewrites the references a manifest or workspace file makes
// to a project's directory when the directory moves: a path dependency, a
// workspace member, a go.work entry, a module path. Language plugins use it to
// answer what changing a project's kind edits.
package pathrefs
