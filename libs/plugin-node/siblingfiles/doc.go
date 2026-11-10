// Package siblingfiles derives the config files that the sibling projects of a
// workspace share (ADR 0050). When at least three projects of one kind, and at
// least two thirds of those that have the file, hold a file at the same relative
// path in the same shape, the shape is the workspace's convention: smith derives
// it for each project that has the file. A file is "in the shape" when it equals
// the majority form after the project name, the package name and the project's
// internal references are substituted. A file that differs in anything else is
// the project's own: it is never proposed and never overwritten, and it is
// reported as diverging. The package reads files and writes none.
package siblingfiles
