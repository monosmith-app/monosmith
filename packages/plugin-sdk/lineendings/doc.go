// Package lineendings keeps the line endings of the files smith writes in
// step with the repository: an edited file keeps its own ending, and a new
// file follows the git attributes, git's core.eol and core.autocrlf, then the
// files beside it.
package lineendings
