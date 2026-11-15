// Package shiplink puts what a project ships into its shipped folder,
// dist/<project dir>, after a tool that keeps its output in its own work
// folder (dist/cargo, dist/dotnet, dist/gradle/<id>) builds it (ADR 0092).
//
// A task names those files in protocol.Task.Ship. Ship hard-links each one
// into the shipped folder, makes the link again when the tool replaced the
// file (a link to the old file keeps the old content), and copies a file it
// cannot link (another volume, a file system without hard links). It never
// makes a symbolic link. When Windows refuses to replace a running program,
// it keeps the old file and reports the program, so the build still passes.
package shiplink
