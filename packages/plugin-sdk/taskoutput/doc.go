// Package taskoutput reads the build task of a project and tells whether its
// command names an output folder inside the workspace's dist/. Go, Python and
// Dart builds take that folder as an option of the command, so a task written
// by hand without it writes beside the sources.
package taskoutput
